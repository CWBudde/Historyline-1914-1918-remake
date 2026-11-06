import { expect, it } from "vitest";
import { compound, level, play, at, phaseAction } from "@historyline/engine";
import { Full } from "../../../engine/src/testing/tables.ts";
import { playData } from "../../../engine/src/testing/play.ts";
import { ContentsController } from "./controller.ts";

function fixture(
  options: { carrier?: boolean; units?: string[]; builds?: string[] } = {},
) {
  const lv: level.Level = {
    mapName: "contents fixture",
    mapNo: 0,
    player: "Central Powers",
    extra: {
      credits: new Map(),
      builds: options.builds ?? ["CP_infantry", "CP_construction"],
      backdrop: "",
      season: "summer",
      seasons: new Map(),
    },
    lines: Array.from({ length: 17 }, (_, line) =>
      Array.from({ length: 8 - (line % 2) }, (): level.Hex => ({
        terrain: "grass",
        infra: { ids: [], list: false },
        unit: "",
        content: null,
      })),
    ),
  };
  const own = at(compound.compounds, 2),
    enemy = at(compound.compounds, 1);
  const north = { line: 4, col: 2 },
    enemyNorth = { line: 10, col: 5 };
  for (const [cp, n] of [
    [own, north],
    [enemy, enemyNorth],
  ] as const)
    for (const piece of Object.keys(cp.pieces)) {
      const c = compound.at(cp, n, piece);
      at(at(lv.lines, c.line), c.col).terrain = cp.prefix + piece;
    }
  const entrance = compound.at(own, north, own.entrance),
    enemyEntrance = compound.at(enemy, enemyNorth, enemy.entrance);
  at(at(lv.lines, entrance.line), entrance.col).content = {
    owner: 1,
    units: options.units ?? ["CP_infantry", "CP_construction"],
  };
  at(at(lv.lines, enemyEntrance.line), enemyEntrance.col).content = {
    owner: 2,
    units: ["EC_infantry"],
  };
  lv.extra?.credits.set(level.coordKey(entrance), 7);
  if (options.carrier)
    at(at(lv.lines, north.line), north.col).unit = "CP_supplyCar";
  at(at(lv.lines, 2), 2).unit = "CP_infantry";
  at(at(lv.lines, 12), 2).unit = "EC_supplyCar";
  at(at(lv.lines, 14), 2).unit = "EC_infantry";
  const session = play.newSession(lv, playData, { profile: Full });
  session.game.energy[0] = 500;
  return { session, entrance, north, enemyEntrance };
}
function action(s: play.Session) {
  if (play.phase(s, 0) !== phaseAction) {
    play.confirm(s, 0);
    play.confirm(s, 1);
    play.resolve(s);
  }
}
function opened(
  s: play.Session,
  c: { line: number; col: number },
  blocked = () => false,
) {
  const controller = new ContentsController(s, 0, blocked);
  expect(controller.open(c)).toBe(true);
  return controller;
}

it("normalizes building parts and shows native stored income separately from account energy", () => {
  const { session: s, north, entrance } = fixture();
  const c = opened(s, north);
  expect(c.box?.at).toEqual(entrance);
  expect(c.box?.hold).toBe(82);
  expect(c.box?.slot).toBe(0);
  expect(c.name()).toBe("factory");
  expect(c.resources()).toEqual({ stored: 7, total: 500 });
  expect(c.slots()).toHaveLength(7);
  expect(c.unit()?.id).toBe("CP_infantry");
  expect(c.options().map((o) => [o.type, o.cost, o.reason])).toEqual([
    [18, 35, "repairs and builds are given in the action phase"],
    [16, 105, "repairs and builds are given in the action phase"],
  ]);
});
it("refuses enemy buildings and filled enemy carriers without exposing their contents", () => {
  const { session: s, enemyEntrance } = fixture();
  const c = new ContentsController(s, 0);
  expect(c.open(enemyEntrance)).toBe(false);
  expect(c.box).toBeNull();
  expect(c.slots()).toEqual([]);
  action(s);
  const infantry = play.view(s, 1).units.get("14,2");
  if (infantry === undefined) throw new Error("missing enemy infantry");
  play.move(s, 1, infantry.index, { line: 12, col: 2 });
  expect(play.contents(s, 1, { line: 12, col: 2 })?.slots.length).toBe(1);
  expect(c.open({ line: 12, col: 2 })).toBe(false);
  expect(c.box).toBeNull();
});
it("prioritizes the own carrier over the building part under it", () => {
  const { session: s, north } = fixture({ carrier: true });
  const c = new ContentsController(s, 0);
  expect(c.open(north)).toBe(false);
  expect(c.message).toBe("nothing stored here");
  const infantry = play.view(s, 0).units.get("2,2");
  if (infantry === undefined) throw new Error("missing infantry");
  play.move(s, 0, infantry.index, north);
  expect(c.open(north)).toBe(true);
  expect(c.box?.at).toEqual(north);
  expect(c.box?.hold).toBeLessThan(70);
  expect(c.name()).toBe(play.view(s, 0).units.get(level.coordKey(north))?.name);
  expect(c.unit()?.index).toBe(infantry.index);
  expect(c.resources()).toEqual({ stored: 0, total: 500 });
  expect(c.options()).toEqual([]);
  action(s);
  expect(c.actions().repair).toBe("no such building");
});
it("opens an empty factory with production but refuses an empty non-producing container", () => {
  const a = fixture({ units: [] });
  expect(new ContentsController(a.session, 0).open(a.entrance)).toBe(true);
  const b = fixture({ units: [], builds: [] });
  const c = new ContentsController(b.session, 0);
  expect(c.open(b.entrance)).toBe(false);
  expect(c.message).toBe("nothing stored here");
  expect(c.open({ line: 999, col: 0 })).toBe(false);
});
it("retains all seven holes and reserved slots and selects the first actual unit", () => {
  const { session: s, entrance } = fixture();
  const h = s.game.holds.get(82);
  if (h === undefined) throw new Error("missing hold");
  const infantry = at(h.slots, 0);
  h.slots[0] = 255;
  h.slots[4] = infantry;
  h.slots[2] = 240;
  const c = opened(s, entrance);
  expect(c.box?.slot).toBe(1);
  expect(c.slots().map((slot) => slot.unit?.id ?? null)).toEqual([
    null,
    "CP_construction",
    null,
    null,
    "CP_infantry",
    null,
    null,
  ]);
  c.select(4);
  expect(c.unit()?.index).toBe(infantry);
  c.select(2);
  expect(c.unit()).toBeNull();
  expect(c.actions().leave).toBe("empty slot");
  expect(() => {
    c.select(7);
  }).toThrow("no such slot");
  expect(c.box?.slot).toBe(2);
});
it("builds through engine prices, selects the new slot and keeps production open on a refused build", () => {
  const { session: s, entrance } = fixture();
  const c = opened(s, entrance);
  c.showProduction();
  const before = play.save(s);
  expect(() => c.build(18)).toThrow("action phase");
  expect(play.save(s)).toEqual(before);
  expect(c.box?.producing).toBe(true);
  action(s);
  expect(c.options().map((o) => o.reason)).toEqual([null, null]);
  const energy = play.energy(s, 0),
    events = c.build(18);
  expect(events[0]?.kind).toBe(play.Built);
  expect(play.energy(s, 0)).toBe(energy - 35);
  expect(c.box?.slot).toBe(2);
  expect(c.box?.producing).toBe(false);
  expect(c.unit()?.index).toBe(events[0]?.unit);
  expect(c.message).toBe(`Infantry built - ${energy - 35} energy left`);
  const after = play.save(s);
  expect(() => c.build(999)).toThrow("does not build");
  expect(play.save(s)).toEqual(after);
});
it("repairs damage using native costs and refuses another order without spending again", () => {
  const { session: s, entrance } = fixture();
  action(s);
  const c = opened(s, entrance),
    u = s.game.units.get(c.unit()?.index ?? -1);
  if (u === undefined) throw new Error("missing unit");
  u.size = 1;
  u.exp = 3;
  const before = play.energy(s, 0),
    events = c.repair();
  expect(events[0]?.kind).toBe(play.Repaired);
  expect(c.unit()?.size).toBe(6);
  expect(c.unit()?.exp).toBe(3);
  expect(play.energy(s, 0)).toBe(before - 15);
  expect(c.actions().repair).toBe("the unit has its order");
  const saved = play.save(s);
  expect(() => c.repair()).toThrow("has its order");
  expect(play.save(s)).toEqual(saved);
});
it("allows full-strength builder repairs to restore charges for the native builder cost", () => {
  const { session: s, entrance } = fixture();
  action(s);
  const c = opened(s, entrance);
  c.select(1);
  const u = s.game.units.get(c.unit()?.index ?? -1);
  if (u === undefined) throw new Error("missing builder");
  u.hold = 0;
  u.exp = 4;
  expect(c.actions().repair).toBeNull();
  const before = play.energy(s, 0);
  c.repair();
  expect(c.unit()?.size).toBe(6);
  expect(s.game.units.get(c.unit()?.index ?? -1)?.hold).toBe(2);
  expect(c.unit()?.exp).toBe(4);
  expect(play.energy(s, 0)).toBe(before - 50);
  expect(c.box?.slot).toBe(1);
});
it("does not spend energy when building or repairing cannot be afforded", () => {
  const { session: s, entrance } = fixture();
  action(s);
  s.game.energy[0] = 14;
  const c = opened(s, entrance);
  expect(c.options().every((o) => o.reason === "not enough energy")).toBe(true);
  expect(c.actions().produce).toBe("not enough energy");
  expect(c.actions().repair).toBe("not enough energy");
  const before = play.save(s);
  expect(() => c.build(18)).toThrow("Not enough energy");
  expect(() => c.repair()).toThrow("Not enough energy");
  expect(play.save(s)).toEqual(before);
});
it("keeps a full factory's slots stable and refuses further production without spending", () => {
  const { session: s, entrance } = fixture();
  action(s);
  const c = opened(s, entrance);
  for (let i = 0; i < 5; i++) c.build(18);
  expect(c.slots().filter((slot) => slot.unit !== null)).toHaveLength(7);
  expect(c.actions().produce).toBe("no free slot");
  const before = play.save(s);
  expect(() => c.build(18)).toThrow("No free slot");
  expect(play.save(s)).toEqual(before);
});
it("prepares and cancels exit aiming without issuing an order or changing the board", () => {
  const { session: s, entrance } = fixture();
  const c = opened(s, entrance);
  const before = play.save(s),
    unit = c.unit();
  const aim = c.prepareExit();
  expect(aim.hold).toBe(82);
  expect(aim.slot).toBe(0);
  expect(aim.unit.index).toBe(unit?.index);
  expect(aim.from).toEqual(entrance);
  expect(aim.destinations.size).toBeGreaterThan(0);
  expect(c.box).toBeNull();
  expect(play.save(s)).toEqual(before);
  expect(play.version(s)).toBe(0);
  expect(s.log).toEqual([]);
  c.close(true);
  expect(play.save(s)).toEqual(before);
});
it("selects the actual production slot without collapsing reserved empty positions", () => {
  const { session: s, entrance } = fixture();
  action(s);
  const hold = s.game.holds.get(82);
  if (hold === undefined) throw new Error("missing hold");
  hold.slots[2] = 240;
  const c = opened(s, entrance),
    events = c.build(18);
  expect(c.box?.slot).toBe(3);
  expect(c.slots()[2]?.unit).toBeNull();
  expect(c.slots()[3]?.unit?.index).toBe(events[0]?.unit);
});
it("refuses exit aiming when native terrain costs leave no destinations", () => {
  const { session: s, entrance } = fixture(),
    c = opened(s, entrance);
  const tables = s.game.profile.tables?.move,
    hold = s.game.holds.get(82);
  if (tables === null || tables === undefined || hold === undefined)
    throw new Error("missing native tables/hold");
  const mask = at(tables.types, 18).mask,
    water = tables.terrains.findIndex((t) => (t.mask & mask) === 0);
  if (water < 0) throw new Error("no impassable native terrain");
  for (const map of s.game.maps)
    for (let i = 0; i < map.length; i += 2)
      if (i !== hold.cells[0]) map[i] = water;
  expect(play.exitDestinations(s, 0, 82, 0).size).toBe(0);
  expect(c.actions().leave).toBe("nowhere to go");
  const before = play.save(s);
  expect(() => c.prepareExit()).toThrow("nowhere to go");
  expect(c.box).not.toBeNull();
  expect(play.save(s)).toEqual(before);
});
it("keeps contents open when an exit is refused in the action phase or after an order", () => {
  const a = fixture();
  const c = opened(a.session, a.entrance);
  action(a.session);
  expect(c.actions().leave).toBe("leaves in the movement phase");
  const before = play.save(a.session);
  expect(() => c.prepareExit()).toThrow("leaves in the movement phase");
  expect(c.box).not.toBeNull();
  expect(play.save(a.session)).toEqual(before);
  const b = fixture(),
    d = opened(b.session, b.entrance),
    u = b.session.game.units.get(d.unit()?.index ?? -1);
  if (u === undefined) throw new Error("missing unit");
  u.flags4 |= 0x200;
  expect(d.actions().leave).toBe("the unit has its order");
  expect(() => d.prepareExit()).toThrow("has its order");
  expect(d.box).not.toBeNull();
});
it.each(["busy", "ready", "ended"] as const)(
  "blocks mutations while %s but preserves read-only inspection and cancel",
  (state) => {
    const { session: s, entrance } = fixture();
    action(s);
    let busy = state === "busy";
    if (state === "ready") play.confirm(s, 0);
    if (state === "ended") play.surrender(s, 1);
    const c = opened(s, entrance, () => busy);
    const why =
      state === "busy"
        ? "the orders are being carried out"
        : state === "ready"
          ? "orders confirmed: F1 opens them again"
          : "the battle is over";
    const before = play.save(s);
    expect(c.actions()).toEqual({ leave: why, repair: why, produce: why });
    expect(c.options().every((o) => o.reason === why)).toBe(true);
    expect(() => c.build(18)).toThrow(why);
    expect(() => c.repair()).toThrow(why);
    expect(() => c.prepareExit()).toThrow(why);
    expect(play.save(s)).toEqual(before);
    c.select(1);
    c.showProduction();
    expect(c.box?.producing).toBe(true);
    c.close();
    expect(c.box?.producing).toBe(false);
    expect(c.box).not.toBeNull();
    c.close();
    expect(c.box).toBeNull();
    expect(c.open(entrance)).toBe(true);
    c.showProduction();
    c.close(true);
    expect(c.box).toBeNull();
    busy = false;
  },
);

it("uses native slot neighbours and production picks without moving or spending", () => {
  const { session, entrance } = fixture(),
    c = new ContentsController(session, 0);
  action(session);
  c.open(entrance);
  const mark = play.mark(session),
    energy = play.energy(session, 0);
  c.step("up");
  expect(c.box?.slot).toBe(1);
  c.step("up");
  expect(c.box?.slot).toBe(1);
  c.step("down");
  expect(c.box?.slot).toBe(0);
  c.step("left");
  expect(c.box?.slot).toBe(0);
  c.step("right");
  expect(c.box?.slot).toBe(1);
  c.showProduction();
  expect(c.box?.pick).toBe(c.options().findIndex((o) => o.reason === null));
  c.step("down");
  expect(c.box?.pick).toBe(1);
  c.step("right");
  expect(c.box?.pick).toBe(1);
  c.step("up");
  expect(c.box?.pick).toBe(0);
  expect(play.mark(session)).toBe(mark);
  expect(play.energy(session, 0)).toBe(energy);
});
