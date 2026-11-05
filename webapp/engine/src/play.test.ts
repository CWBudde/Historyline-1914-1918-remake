import { describe, expect, it } from "vitest";
import * as play from "./play.ts";
import { Full, Builds } from "./testing/tables.ts";
import { playData, shippedLevel } from "./testing/play.ts";
import { parse } from "./level.ts";
import { readBytes } from "./testing/files.ts";
import { attempt, errorsIs } from "./goruntime.ts";

describe("battle sessions", () => {
  it("preserves replay events and the original refusal when a later command fails", () => {
    const s = play.newSession(shippedLevel("1914-08"), playData, {
      profile: Full,
    });
    let caught: unknown;
    try {
      play.replay(s, [
        { kind: "confirm", side: 1, cell: 0 },
        { kind: "move", side: 0, unit: 999, cell: 0 },
      ]);
    } catch (e) {
      caught = e;
    }
    expect(caught).toBeInstanceOf(play.ReplayError);
    expect(errorsIs(caught, play.OrderError)).toBe(true);
    if (caught instanceof play.ReplayError)
      expect(caught.events.map((e) => e.kind)).toEqual([play.Confirmed]);
  });
  it("uses the session's construction and trench messages", () => {
    expect(play.messageText(0x6e)).toBe("Trench work ordered.");
    expect(play.messageText(0x46)).toBe("Depot construction ordered.");
    expect(play.messageText(0x47)).toBe("No more depots can be built.");
    expect(play.messageText(0x4b)).toBe("There is no room for a depot here.");
    expect(play.messageText(0x4c)).toBe(
      "The unit has nothing left to build with.",
    );
  });
  it("retains Go's route snapshot when adding a moving ship's rear animation", () => {
    const s = play.newSession(shippedLevel("1915-08"), playData, {
      profile: Full,
    });
    const ship = [...play.view(s, 0).units.values()].find(
      (u) =>
        u.side === 0 &&
        u.rear !== null &&
        (attempt(() => play.destinations(s, 0, u.index))[0]?.size ?? 0) > 0,
    );
    if (ship === undefined) throw new Error("no movable ship in fixture");
    const to = play.coords(play.destinations(s, 0, ship.index))[0];
    if (to === undefined) throw new Error("no destination");
    expect(play.move(s, 0, ship.index, to)[0]?.rear.length).toBeGreaterThan(0);
    play.confirm(s, 0);
    play.confirm(s, 1);
    const reveal = play
      .resolve(s)
      .find((e) => e.kind === play.Revealed && e.unit === ship.index);
    expect(reveal?.rear).toEqual([]);
  });
  const start = () =>
    play.newSession(shippedLevel("1914-08"), playData, { profile: Full });
  it("keeps each side's own map until phase change and reveals surviving routes", () => {
    const s = start();
    const unit = [...play.view(s, 0).units.values()].find(
      (u) =>
        u.side === 0 &&
        !u.held &&
        attempt(() => play.destinations(s, 0, u.index))[0]?.size,
    );
    if (unit === undefined) throw new Error("no mobile unit");
    const to = play.coords(play.destinations(s, 0, unit.index))[0];
    if (to === undefined) throw new Error("no destination");
    const before = new Map(
      [...play.view(s, 1).units].map(([key, u]) => [key, u.index]),
    );
    const route = play.route(s, 0, unit.index, to);
    expect(route[0]).toEqual(unit.at);
    expect(route.at(-1)).toEqual(to);
    expect(
      play.costs(s, 0, unit.index).get(`${to.line},${to.col}`),
    ).toBeGreaterThan(0);
    const events = play.move(s, 0, unit.index, to);
    expect(events[0]?.kind).toBe(play.Moved);
    expect(events[0]?.path[0]).toEqual(unit.at);
    expect(events[0]?.path).toEqual(route);
    expect(
      new Map([...play.view(s, 1).units].map(([key, u]) => [key, u.index])),
    ).toEqual(before);
    play.confirm(s, 0);
    play.confirm(s, 1);
    const resolved = play.resolve(s);
    expect(
      resolved.some((e) => e.kind === play.Revealed && e.unit === unit.index),
    ).toBe(true);
    expect(play.phase(s, 0)).toBe(2);
    expect(play.view(s, 0).units).toEqual(play.view(s, 1).units);
  });
  it("clones for off-thread planning and replays a complete phase independently", () => {
    const s = start();
    const c = play.clone(s, Builds);
    const mark = play.mark(c);
    play.confirm(c, 0);
    play.confirm(c, 1);
    const evs = play.resolve(c);
    expect(play.phase(s, 0)).toBe(1);
    const commands = play.commandsSince(c, mark);
    const got = play.replay(s, commands);
    expect(got.slice(-evs.length)).toEqual(evs);
    expect(play.save(s)).toEqual(play.save(c));
    expect(play.commandsSince(s, -1)).toEqual([]);
  });
  it("resumes byte-identically and rejects a save from another map", () => {
    const s = start();
    play.confirm(s, 0);
    const loaded = play.load(s.level, playData, play.save(s), Builds);
    expect(play.save(loaded)).toEqual(play.save(s));
    expect(() =>
      play.load(shippedLevel("1914-10"), playData, play.save(s), Builds),
    ).toThrow();
    expect(() =>
      play.replay(s, [{ kind: "move", side: 0, unit: 999, cell: 0 }]),
    ).toThrow();
  });
  it("surrenders once and logs only confirmations actually withdrawn", () => {
    const s = start();
    expect(play.unconfirm(s, 0)).toEqual([]);
    expect(play.mark(s)).toBe(0);
    play.confirm(s, 0);
    play.unconfirm(s, 0);
    expect(play.mark(s)).toBe(2);
    expect(play.surrender(s, 0)[0]?.end?.winner).toBe(1);
    expect(play.surrender(s, 0)).toEqual([]);
  });
  it("returns contents and building footprints at their entrances", () => {
    const s = start();
    for (const [key, building] of play.view(s, 0).buildings) {
      const at = play.coordFromKey(key);
      expect(play.entrance(s, at)).toEqual(at);
      expect(play.footprint(s, at)[0]).toEqual(at);
      const content = play.contentSlots(s, play.sideOf(building.owner), at);
      if (building.owner === 1) expect(content?.slots).toHaveLength(7);
    }
  });
  it("gives type details, scores and correct domains from the rule tables", () => {
    const s = start();
    expect(play.typeInfo(s, 0, 18)?.domain).toBe(play.Infantry);
    expect(play.typeInfo(s, 0, 18)?.id).toBe("CP_infantry");
    expect(play.typeInfo(s, 0, 999)).toBeNull();
    expect(play.stats(s).level).toBe("PULSE");
    expect(play.stats(s).map.units).toBe(s.game.units.size);
  });
  it("accepts the legacy fixture as a session", () => {
    const s = play.newSession(
      parse(readBytes("internal/level/testdata/pulse.json")),
      playData,
      { profile: Full },
    );
    expect(play.view(s, 0).units.size).toBeGreaterThan(0);
  });
});
