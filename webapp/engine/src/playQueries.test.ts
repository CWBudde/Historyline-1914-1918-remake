import { describe, expect, it } from "vitest";
import * as play from "./play.ts";
import { type Level, type Hex, coordKey } from "./level.ts";
import { type Coord, around } from "./hexgrid.ts";
import { at, attempt } from "./goruntime.ts";
import { Full, Builds } from "./testing/tables.ts";
import { playData, shippedLevel } from "./testing/play.ts";

function grass(units: readonly [Coord, string][]): Level {
  const lv: Level = {
    mapName: "fixture",
    mapNo: 0,
    player: "Central Powers",
    extra: null,
    lines: Array.from({ length: 11 }, (_, line) =>
      Array.from({ length: 6 - (line % 2) }, (): Hex => ({
        terrain: "grass",
        infra: { ids: [], list: false },
        unit: "",
        content: null,
      })),
    ),
  };
  for (const [c, id] of units) at(at(lv.lines, c.line), c.col).unit = id;
  return lv;
}
function bridge(id: string, deck = "path25"): play.Session {
  const lv = grass([[{ line: 2, col: 2 }, id]]);
  for (let line = 1; line < 11; line += 2)
    at(at(lv.lines, line), 2).infra = { ids: ["stream36"], list: true };
  at(at(lv.lines, 3), 2).infra = {
    ids: ["stream36", "rBridge25", deck],
    list: true,
  };
  return play.newSession(lv, playData, { profile: Full });
}
function phase(s: play.Session, side: number): void {
  while (play.phase(s, side) !== 2) {
    play.confirm(s, 0);
    play.confirm(s, 1);
    play.resolve(s);
  }
}
function resolve(s: play.Session) {
  play.confirm(s, 0);
  play.confirm(s, 1);
  return play.resolve(s);
}

describe("session queries and terrain work", () => {
  it("shows queued attacks with Go's zero raw event and their target hex", () => {
    const s = play.newSession(
      grass([
        [{ line: 4, col: 3 }, "CP_infantry"],
        [{ line: 5, col: 3 }, "EC_infantry"],
      ]),
      playData,
      { profile: Full },
    );
    const enemy = [...play.view(s, 1).units.values()].find((u) => u.side === 1);
    if (enemy === undefined) throw new Error("no enemy");
    const target = play.coords(play.targets(s, 1, enemy.index))[0];
    if (target === undefined) throw new Error("no target");
    play.attack(s, 1, enemy.index, target);
    const order = play.orders(s, 1)[0];
    expect(order?.to).toEqual(target);
    expect(order?.raw.kind).toBe("");
    expect(order?.text).toBe("");
  });
  it.each([
    ["CP_infantry", true, true],
    ["CP_eliteInf", true, true],
    ["CP_sappers", true, true],
    ["CP_antiTank", true, false],
    ["CP_staticAA", true, false],
    ["CP_cavalry", true, true],
    ["CP_lArtillery", true, false],
    ["CP_ft17", false, false],
    ["CP_armCar", false, false],
    ["CP_markI", false, false],
    ["CP_mArtillery", false, false],
    ["CP_supplyCar", false, false],
    ["CP_construction", false, false],
  ] as const)(
    "%s follows the Go light bridge restrictions",
    (id, onBridge, east) => {
      const s = bridge(id),
        dst = play.destinations(s, 0, 0);
      expect(dst.has("3,2")).toBe(onBridge);
      if (east) expect(dst.has("2,3")).toBe(true);
      if (!onBridge) expect(dst.has("2,3")).toBe(false);
    },
  );
  it("keeps roads open to tanks and restores temporary light-bridge terrain after refusals", () => {
    const road = bridge("CP_ft17", "road25");
    expect(play.destinations(road, 0, 0).has("3,2")).toBe(true);
    const s = bridge("CP_ft17"),
      before = s.game.maps.map((m) => m.filter((_, i) => i % 2 === 0));
    expect(
      play
        .move(s, 0, 0, { line: 2, col: 3 })
        .some((e) => e.kind === play.Moved),
    ).toBe(false);
    expect(s.game.maps.map((m) => m.filter((_, i) => i % 2 === 0))).toEqual(
      before,
    );
  });
  it("previews, digs and fills trenches, retaining their look over a save", () => {
    const from = { line: 4, col: 3 },
      to = at(around(from), 2);
    const s = play.newSession(
      grass([
        [from, "CP_sappers"],
        [{ line: 8, col: 3 }, "EC_infantry"],
      ]),
      playData,
      { profile: Full },
    );
    const u = play.view(s, 0).units.get(coordKey(from))?.index ?? -1;
    phase(s, 0);
    const preview = play.work(s, 0, u, to);
    expect(preview.hexes).toHaveLength(3);
    expect(preview.fill).toBe(false);
    expect(play.attack(s, 0, u, to)[0]?.kind).toBe(play.Tasked);
    expect(play.orders(s, 0)).toEqual([]);
    const dug = resolve(s).find((e) => e.kind === play.Dug);
    expect(dug?.hexes).toEqual(preview.hexes);
    expect(
      play
        .reworked(s)
        .get(coordKey(to))
        ?.infra.ids.some((id) => id.startsWith("trench")),
    ).toBe(true);
    const loaded = play.load(s.level, playData, play.save(s), Builds);
    expect(play.reworked(loaded)).toEqual(play.reworked(s));
    phase(s, 0);
    expect(play.work(s, 0, u, to).fill).toBe(true);
    play.attack(s, 0, u, to);
    expect(resolve(s).find((e) => e.kind === play.Dug)?.text).toBe(
      "Trench filled.",
    );
    expect(play.reworked(s).get(coordKey(to))?.infra.ids).toEqual([]);
  });
  it("builds a depot in two halves and supplies its dynamic footprint after reload", () => {
    const from = { line: 4, col: 3 },
      s = play.newSession(
        grass([
          [from, "CP_construction"],
          [{ line: 9, col: 3 }, "EC_infantry"],
        ]),
        playData,
        { profile: Full },
      );
    const u = play.view(s, 0).units.get(coordKey(from))?.index ?? -1;
    phase(s, 0);
    play.canConstruct(s, 0, u);
    const sites = play.coords(play.targets(s, 0, u));
    expect(sites).toHaveLength(10);
    const to = at(sites, 0);
    expect(play.attack(s, 0, u, to)[0]?.text).toBe(
      "Depot construction ordered.",
    );
    const half = resolve(s).find((e) => e.kind === play.Constructed);
    expect(half?.hexes).toHaveLength(4);
    expect(play.reworked(s).get(coordKey(to))?.content).toBeNull();
    expect(play.view(s, 0).buildings.has(coordKey(to))).toBe(false);
    phase(s, 0);
    play.attack(s, 0, u, to);
    expect(resolve(s).find((e) => e.kind === play.Constructed)?.text).toBe(
      "Depot built.",
    );
    const loaded = play.load(s.level, playData, play.save(s), Builds);
    expect(play.footprint(loaded, to)).toHaveLength(4);
    for (const c of play.footprint(loaded, to))
      expect(play.entrance(loaded, c)).toEqual(to);
    expect(play.reworked(loaded).get(coordKey(to))?.content?.owner).toBe(1);
  });
  it("offers only the viewing side's factory types and accounts for production and repair", () => {
    const s = play.newSession(shippedLevel("1914-10"), playData, {
      profile: Full,
    });
    const bd = [...play.view(s, 0).buildings.values()].find(
      (b) => b.kind === play.Factory && b.owner === 1,
    );
    if (bd === undefined) throw new Error("no factory");
    const options = play.buildable(s, 0, bd.hold);
    expect(options.length).toBeGreaterThan(0);
    expect(options.every((o) => o.id.startsWith("CP_"))).toBe(true);
    expect(play.buildable(s, 2, bd.hold)).toEqual([]);
    expect(
      attempt(() => {
        play.canBuild(s, 0, bd.hold, at(options, 0).type);
      })[1],
    ).toBeInstanceOf(play.OrderError);
    s.game.energy[0] = 1000;
    phase(s, 0);
    const option = at(options, 0),
      before = play.energy(s, 0);
    play.canBuild(s, 0, bd.hold, option.type);
    const built = play.build(s, 0, bd.hold, option.type);
    expect(built[0]?.kind).toBe(play.Built);
    expect(play.energy(s, 0)).toBe(before - option.cost);
    const entrance = [...play.view(s, 0).buildings].find(
      ([, b]) => b.hold === bd.hold,
    )?.[0];
    if (entrance === undefined) throw new Error("no entrance");
    const held = play.contentSlots(s, 0, play.coordFromKey(entrance));
    expect(held?.slots).toHaveLength(7);
    const slot = held?.slots.find(
      (v) => v.unit?.index === built[0]?.unit,
    )?.slot;
    if (slot === undefined) throw new Error("new unit was not held");
    resolve(s);
    resolve(s);
    const u = s.game.units.get(built[0]?.unit ?? -1);
    if (u === undefined) throw new Error("missing unit");
    u.size = 1;
    const energy = play.energy(s, 0);
    play.canRepair(s, 0, bd.hold, slot);
    play.repair(s, 0, bd.hold, slot);
    expect(play.energy(s, 0)).toBe(energy - 15);
    expect(play.unit(s, built[0]?.unit ?? -1).size).toBe(
      play.typeInfo(s, 0, option.type)?.group,
    );
  });
});
