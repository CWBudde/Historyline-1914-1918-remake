import { expect, it } from "vitest";
import { ai, play, phaseMovement } from "@historyline/engine";
import { Full, Builds } from "../../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../../engine/src/testing/play.ts";
import { LocalMatch } from "./match.ts";

function session() {
  return play.newSession(shippedLevel("1914-08", "2p"), playData, {
    profile: Full,
    computer: [false, false],
  });
}
function move(match: LocalMatch, side: 0 | 1) {
  const b = match.sides[side];
  for (const unit of b.view().units.values()) {
    if (unit.side !== side || unit.head !== null) continue;
    let to;
    try {
      to = play.coords(play.destinations(b.session, side, unit.index))[0];
    } catch {
      // Keep looking for a unit that can move in this side's current mode.
      continue;
    }
    if (to === undefined) continue;
    b.click(unit.at);
    b.click(to);
    expect(play.unit(b.session, unit.index).at).toEqual(to);
    b.tick(100_000);
    return { unit, to };
  }
  throw Error(`No movement order for side ${side}`);
}

it("confirms only one human and leaves the other in their own mode", async () => {
  const s = session(),
    match = new LocalMatch(s, "simultaneous");
  const before = play.phase(s, 0);
  await match.sides[1].phase();
  expect(play.ready(s, 1)).toBe(true);
  expect(play.ready(s, 0)).toBe(false);
  expect(play.phase(s, 0)).toBe(before);
  expect(match.sides[0].session).toBe(match.sides[1].session);
  expect(match.sides[0].contents).not.toBe(match.sides[1].contents);
  expect(match.sides[0].playback).not.toBe(match.sides[1].playback);
  expect(play.phase(s, 0)).toBe(phaseMovement);
  const otherVersion = match.sides[1].version;
  move(match, 0);
  expect(match.sides[1].version).toBeGreaterThan(otherVersion);
  expect(play.ready(s, 1)).toBe(true);
});

it("resolves once after near-simultaneous confirmation and refreshes both fair views", async () => {
  const s = session(),
    expected = play.clone(s, Builds),
    match = new LocalMatch(s, "simultaneous");
  play.confirm(expected, 0);
  play.confirm(expected, 1);
  play.resolve(expected);
  await Promise.all(match.sides.map((b) => b.phase()));
  expect(play.save(s)).toEqual(play.save(expected));
  for (const b of match.sides) {
    b.tick(100_000);
    expect(b.view()).toEqual(play.view(s, b.side));
    expect(play.ready(s, b.side)).toBe(false);
  }
});

it("keeps independent selections, cursors and overlays while the peer refreshes", () => {
  const match = new LocalMatch(session(), "simultaneous"),
    [a, b] = match.sides;
  a.panel = "game";
  a.setCursor({ line: 4, col: 3 });
  b.setCursor({ line: 7, col: 5 });
  b.panel = "map";
  expect(a.cursor).toEqual({ line: 4, col: 3 });
  expect(b.cursor).toEqual({ line: 7, col: 5 });
  expect(a.panel).toBe("game");
  expect(b.panel).toBe("map");
});

it("restores a confirmed human without confirming or changing the peer", async () => {
  const s = session();
  play.confirm(s, 0);
  const bytes = play.save(s),
    restored = play.load(s.level, playData, bytes, Builds),
    match = new LocalMatch(restored, "simultaneous");
  expect(play.save(restored)).toEqual(bytes);
  await match.sides[1].phase();
  expect(play.phase(restored, 0)).not.toBe(play.phase(s, 0));
});

it("blocks both humans during a hot-seat handover, then enables only its recipient", async () => {
  const s = session(),
    match = new LocalMatch(s, "hotseat", 1);
  expect(match.handover).toBe(1);
  const bytes = play.save(s);
  await Promise.all(match.sides.map((b) => b.phase()));
  expect(play.save(s)).toEqual(bytes);
  match.takeOver();
  expect(match.handover).toBeNull();
  expect(match.activeSide).toBe(1);
  await match.sides[0].phase();
  expect(play.save(s)).toEqual(bytes);
  await match.sides[1].phase();
  expect(match.handover).toBe(0);
  expect(match.saveSide()).toBe(0);
  expect(play.ready(s, 1)).toBe(true);
  expect(play.ready(s, 0)).toBe(false);
  const after = play.save(s);
  await match.sides[0].phase();
  expect(play.save(s)).toEqual(after);
  match.takeOver();
  move(match, 0);
  await match.sides[0].phase();
  expect(match.handover).toBe(1);
  expect(play.ready(s, 0)).toBe(false);
  expect(play.ready(s, 1)).toBe(false);
});

it("keeps a handover locked across pause/resume and ignores disposed input", async () => {
  const match = new LocalMatch(session(), "hotseat");
  match.setPaused(true);
  match.takeOver();
  const bytes = play.save(match.session);
  await match.sides[0].phase();
  expect(play.save(match.session)).toEqual(bytes);
  match.setPaused(false);
  expect(match.handover).toBe(0);
  match.takeOver();
  await match.sides[0].phase();
  expect(play.ready(match.session, 0)).toBe(true);
  match.dispose();
  match.takeOver();
  const ended = play.save(match.session);
  await Promise.all(match.sides.map((b) => b.phase()));
  expect(play.save(match.session)).toEqual(ended);
});

it("rejects non-local sessions and invalid active sides", () => {
  const s = session();
  s.game.players[1].computer = true;
  expect(() => new LocalMatch(s, "simultaneous")).toThrow(/human/i);
  expect(() => new LocalMatch(session(), "hotseat", 2)).toThrow(/side/i);
});

it("blocks contents mutations when a local panel is paused or disposed", () => {
  const lv = shippedLevel("1914-10", "2p");
  for (const line of lv.lines)
    for (const hex of line)
      if (hex.terrain.startsWith("factory") && hex.content !== null) {
        hex.content.owner = 1;
        hex.content.units = [];
      }
  const s = play.newSession(lv, playData, { profile: Full });
  play.confirm(s, 0);
  play.confirm(s, 1);
  play.resolve(s);
  s.game.energy[0] = 500;
  const match = new LocalMatch(s, "simultaneous"),
    b = match.sides[0],
    factory = [...b.view().buildings.entries()].find(
      ([, v]) => v.kind === play.Factory && v.owner === 1,
    );
  if (factory === undefined) throw Error("Missing fixture factory");
  b.openContents(play.coordFromKey(factory[0]));
  const option = b.contents.options().find((v) => v.reason === null);
  if (option === undefined) throw Error("Missing fixture build option");
  const before = play.save(s);
  match.setPaused(true);
  b.produce(option.type);
  expect(play.save(s)).toEqual(before);
  match.setPaused(false);
  b.produce(option.type);
  expect(play.save(s)).not.toEqual(before);
  match.dispose();
  const disposed = play.save(s);
  b.produce(option.type);
  expect(play.save(s)).toEqual(disposed);
});

it("reopens one human's confirmation without changing the peer's orders or mode", async () => {
  const match = new LocalMatch(session(), "simultaneous"),
    [a, b] = match.sides;
  await b.phase();
  move(match, 0);
  const orders = a.orders(),
    mode = play.phase(match.session, 0);
  await b.requestPhase();
  expect(play.ready(match.session, 1)).toBe(false);
  expect(a.orders()).toEqual(orders);
  expect(play.phase(match.session, 0)).toBe(mode);
  await a.phase();
  expect(play.ready(match.session, 1)).toBe(false);
  expect(play.phase(match.session, 0)).toBe(mode);
});

it("plays a real queued engagement independently in both fair panels", async () => {
  const s = session();
  for (let n = 0; n < 60; n++) {
    for (const u of play.view(s, 0).units.values()) {
      if (u.side !== 0 || play.sapper(s, u.index) || play.builder(s, u.index))
        continue;
      const target = play.coords(play.targets(s, 0, u.index))[0];
      if (target === undefined) continue;
      const match = new LocalMatch(s, "simultaneous"),
        [a, b] = match.sides;
      for (const side of match.sides) side.playback.setScenes(true);
      a.click(u.at);
      a.click(target);
      expect(a.orders().some((event) => event.unit === u.index)).toBe(true);
      const expected = play.clone(s, Builds);
      play.confirm(expected, 0);
      play.confirm(expected, 1);
      const events = play.resolve(expected),
        battle = events.find((e) => e.kind === play.Battle)?.battle;
      expect(battle).toBeTruthy();
      await Promise.all(match.sides.map((side) => side.phase()));
      expect(play.save(s)).toEqual(play.save(expected));
      expect(a.playback.scene()?.battle).toEqual(battle);
      expect(b.playback.scene()?.battle).toEqual(battle);
      a.tick(100_000);
      expect(a.playback.scene()).toBeNull();
      expect(b.playback.scene()).not.toBeNull();
      expect(a.view()).toEqual(play.view(s, 0));
      b.tick(100_000);
      expect(b.view()).toEqual(play.view(s, 1));
      return;
    }
    ai.plan({}, s, 0);
    ai.plan({}, s, 1);
    play.resolve(s);
  }
  throw Error("No fixture engagement");
}, 30_000);

it("can pause from a synchronous change callback without recursively changing orders", async () => {
  const match = new LocalMatch(session(), "simultaneous");
  let calls = 0;
  match.onChange = () => {
    calls++;
    match.setPaused(true);
  };
  await match.sides[0].phase();
  expect(calls).toBeGreaterThan(0);
  expect(calls).toBeLessThan(10);
  expect(play.ready(match.session, 0)).toBe(true);
  expect(play.ready(match.session, 1)).toBe(false);
  const before = play.save(match.session);
  await match.sides[1].phase();
  expect(play.save(match.session)).toEqual(before);
});
