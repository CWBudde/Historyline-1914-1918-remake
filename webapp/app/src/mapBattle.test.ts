import { expect, it } from "vitest";
import { ai, hexgrid, level, play } from "@historyline/engine";
import { Full, Builds, Extended } from "../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../engine/src/testing/play.ts";
import {
  MapBattle,
  actionPrompt,
  actionTargets,
  effectStatus,
  startHex,
} from "./mapBattle.ts";

function start(month = "1914-08") {
  return play.newSession(shippedLevel(month), playData, {
    profile: Full,
    computer: [false, true],
  });
}
function mobile(session: play.Session, side: number) {
  for (const u of play.view(session, side).units.values()) {
    if (u.side !== side || u.head !== null) continue;
    try {
      const to = play.coords(play.destinations(session, side, u.index))[0];
      if (to !== undefined) return { u, to };
    } catch {
      /* Units without a legal movement order are skipped. */
    }
  }
  throw new Error("fixture has no movable unit");
}

it("selects own map units, follows legal routes and blocks orders during movement", () => {
  const s = start(),
    battle = new MapBattle(s, 0),
    { u, to } = mobile(s, 0);
  const original = structuredClone(s.level);
  const enemy = [...battle.view().units.values()].find((v) => v.side === 1);
  if (enemy === undefined) throw new Error("missing fixture enemy");
  battle.click(enemy.at);
  expect(battle.selected).toBeNull();
  battle.click(u.at);
  expect(battle.selected).toBe(u.index);
  expect(battle.destinations().has(level.coordKey(to))).toBe(true);
  expect(battle.route(to).at(-1)).toEqual(to);
  battle.click(to);
  expect(battle.selected).toBeNull();
  expect(battle.playback.moving()?.unit.index).toBe(u.index);
  expect(play.unit(s, u.index).at).toEqual(to);
  const mark = play.mark(s);
  battle.click(enemy.at);
  expect(play.mark(s)).toBe(mark);
  battle.tick(100_000);
  expect(battle.busy()).toBe(false);
  battle.click(to);
  expect(battle.selected).toBeNull();
  expect(battle.status).not.toBe("");
  expect(s.level).toEqual(original);
});

it("selects a two-cell ship's head when its rear is clicked", () => {
  const s = start("1915-08"),
    battle = new MapBattle(s, 0);
  const rear = [...battle.view().units.values()].find(
    (u) => u.side === 0 && u.head !== null,
  );
  if (rear?.head === undefined || rear.head === null)
    throw new Error("missing ship");
  const head = battle.view().units.get(level.coordKey(rear.head));
  battle.click(rear.at);
  expect(battle.selected).toBe(head?.index);
  battle.cancel();
  expect(battle.selected).toBeNull();
  expect(battle.targets().size).toBe(0);
});

it("shows the viewing side's map while enemy moves remain unrevealed", () => {
  const s = start();
  play.confirm(s, 0);
  play.confirm(s, 1);
  play.resolve(s);
  const battle = new MapBattle(s, 0),
    { u, to } = mobile(s, 1);
  play.move(s, 1, u.index, to);
  const seen = [...battle.view().units.values()].find(
    (v) => v.index === u.index,
  );
  expect(seen?.at).toEqual(u.at);
  expect(seen?.moved).toBe(true);
  expect(seen?.target).toBeNull();
  battle.click(to);
  expect(battle.selected).toBeNull();
});

it("confirms the human, replays worker commands and resolves a complete phase", async () => {
  const s = start(),
    expected = play.clone(s, Builds);
  play.confirm(expected, 0);
  ai.plan({}, expected, 1);
  const expectedBefore = [play.view(expected, 0), play.view(expected, 1)];
  play.resolve(expected);
  let release: () => void = () => {};
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  const battle = new MapBattle(s, 0, {
    commands: async (session, side) => {
      const snapshot = play.clone(session, Builds);
      await gate;
      return ai.commands({}, snapshot, side);
    },
  });
  let observed: readonly play.View[] | null = null;
  battle.onResolution = (before) => {
    observed = before;
  };
  const pending = battle.phase();
  expect(battle.busy()).toBe(true);
  expect(play.ready(s, 0)).toBe(true);
  const mark = play.mark(s);
  battle.click({ line: 0, col: 0 });
  await battle.phase();
  expect(play.mark(s)).toBe(mark);
  release();
  await pending;
  expect(play.save(s)).toEqual(play.save(expected));
  expect(observed).toEqual(expectedBefore);
  battle.tick(100_000);
  expect(battle.view()).toEqual(play.view(s, 0));
});

it("allows retry after a worker failure and blocks confirmed or ended inputs", async () => {
  const s = start(),
    battle = new MapBattle(s, 0, {
      commands: () => Promise.reject(new Error("worker unavailable")),
    });
  await battle.phase();
  expect(battle.status).toContain("worker unavailable");
  expect(play.ready(s, 0)).toBe(false);
  expect(battle.busy()).toBe(false);
  const own = mobile(s, 0).u.at;
  play.confirm(s, 0);
  battle.click(own);
  expect(battle.selected).toBeNull();
  play.surrender(s, 1);
  const mark = play.mark(s);
  await battle.phase();
  expect(play.mark(s)).toBe(mark);
});

it("exposes legal action targets and queues a selected attack", () => {
  const s = start();
  for (let n = 0; n < 60; n++) {
    for (const u of play.view(s, 0).units.values()) {
      if (u.side !== 0 || play.sapper(s, u.index) || play.builder(s, u.index))
        continue;
      const target = play.coords(play.targets(s, 0, u.index))[0];
      if (target === undefined) continue;
      const battle = new MapBattle(s, 0);
      battle.click(u.at);
      expect(battle.targets().has(level.coordKey(target))).toBe(true);
      battle.click(target);
      expect(battle.orders().some((o) => o.unit === u.index)).toBe(true);
      expect(battle.playback.directions.has(u.index)).toBe(true);
      expect(battle.selected).toBeNull();
      return;
    }
    ai.plan({}, s, 0);
    ai.plan({}, s, 1);
    play.resolve(s);
  }
  throw new Error("fixture never offered an attack");
}, 30_000);

it("words a resolution's gas for each side", () => {
  const ev = (fields: Partial<play.Event>): play.Event => play.newEvent(fields);
  const evs = [
    ev({ kind: play.Released, side: 1, text: "Gas released." }),
    ev({ kind: play.Gassed, side: 0, text: "Infantry lose 2 to gas." }),
  ];
  expect(effectStatus(evs, 1)).toBe("Gas released. Infantry lose 2 to gas.");
  expect(effectStatus(evs, 0)).toBe(
    "The enemy released gas. Infantry lose 2 to gas.",
  );
  expect(effectStatus([ev({ kind: play.Battle })], 0)).toBeNull();
});

it("words a resolution's work units, the enemy's tunnels hidden", () => {
  const ev = (fields: Partial<play.Event>): play.Event => play.newEvent(fields);
  const evs = [
    ev({ kind: play.Wired, side: 0, text: "Wire laid." }),
    ev({ kind: play.Tunnelled, side: 0, text: "Tunnelling begun." }),
    ev({ kind: play.TunnelLost, side: 0, text: "The tunnel was abandoned." }),
    ev({ kind: play.Blown, side: 0, text: "A mine went up." }),
    ev({ kind: play.Mined, side: 1, text: "Infantry lose 4 to a mine." }),
    ev({ kind: play.Healed, side: 1, text: "Infantry regain 1." }),
  ];
  expect(effectStatus(evs, 0)).toBe(
    "Wire laid. Tunnelling begun. The tunnel was abandoned. A mine went up. Infantry lose 4 to a mine. Infantry regain 1.",
  );
  expect(effectStatus(evs, 1)).toBe(
    "The enemy laid wire. A mine went up. Infantry lose 4 to a mine. Infantry regain 1.",
  );
});

it("prompts the wiring party, the tunnellers and the field ambulance", () => {
  const enemy = hexgrid.around({ line: 6, col: 3 })[3] ?? { line: 0, col: 0 };
  const units: [hexgrid.Coord, string][] = [
    [{ line: 4, col: 3 }, "CP_wiring"],
    [{ line: 6, col: 3 }, "CP_tunnellers"],
    [{ line: 6, col: 5 }, "CP_ambulance"],
    [enemy, "EC_infantry"],
  ];
  const lv: level.Level = {
    mapName: "grass",
    mapNo: 0,
    player: "Central Powers",
    extra: null,
    lines: Array.from({ length: 11 }, (_, l) =>
      Array.from({ length: 8 - (l % 2) }, (): level.Hex => ({
        terrain: "grass",
        infra: { ids: [], list: false },
        unit: "",
        content: null,
      })),
    ),
  };
  for (const [c, id] of units) {
    const h = lv.lines[c.line]?.[c.col];
    if (h !== undefined) h.unit = id;
  }
  const s = play.newSession(lv, playData, {
    profile: Full,
    extended: Extended,
  });
  while (play.phase(s, 0) !== 2) {
    play.confirm(s, 0);
    play.confirm(s, 1);
    play.resolve(s);
  }
  const v = play.view(s, 0),
    at = (c: hexgrid.Coord): play.UnitView => {
      const u = v.units.get(level.coordKey(c));
      if (u === undefined) throw new Error("no unit");
      return u;
    },
    wiring = at({ line: 4, col: 3 }),
    tun = at({ line: 6, col: 3 }),
    amb = at({ line: 6, col: 5 }),
    prompt = (u: play.UnitView) =>
      actionPrompt(s, 0, u, actionTargets(s, 0, u.index).size);
  expect(prompt(wiring)).toBe("Wiring Party: pick where to lay wire (2 left)");
  expect(prompt(tun)).toBe(
    "Tunnellers: pick a hex to mine (it blows in 2 turns if they stay)",
  );
  expect(prompt(amb)).toBe(
    "Field Ambulance: heals the foot units next to it each turn",
  );
  expect(actionTargets(s, 0, amb.index).size).toBe(0);
  expect(actionTargets(s, 0, tun.index).has(level.coordKey(enemy))).toBe(true);
  play.attack(s, 0, tun.index, enemy);
  play.confirm(s, 0);
  play.confirm(s, 1);
  play.resolve(s);
  expect(play.tunnels(s, 0)).toHaveLength(1);
  expect(play.tunnels(s, 1)).toStrictEqual([]);
  expect(prompt(tun)).toMatch(
    new RegExp(
      `^Tunnellers: digging under ${String(enemy.line)},${String(enemy.col)} \\(\\d/2 turns\\)$`,
    ),
  );
  expect(actionTargets(s, 0, tun.index).size).toBe(0);
});

it("starts each side's cursor on its own HQ, as the desktop does", () => {
  const session = start();
  for (const side of [0, 1]) {
    const battle = new MapBattle(session, side),
      view = play.view(session, side),
      hq = [...view.buildings].find(([, b]) => b.hq && b.owner === side + 1);
    expect(hq).toBeDefined();
    expect(level.coordKey(battle.cursor)).toBe(hq?.[0]);
  }
  // Without an HQ: the first own unit in line and column order, else the middle.
  const units = new Map([
    [1, { side: 0, at: { line: 5, col: 2 } }],
    [2, { side: 0, at: { line: 3, col: 7 } }],
    [3, { side: 1, at: { line: 1, col: 1 } }],
  ]);
  const view = { buildings: new Map(), units } as unknown as play.View,
    map = {
      lines: Array.from({ length: 9 }, () => Array<null>(11).fill(null)),
    };
  expect(startHex(view, 0, map)).toEqual({ line: 3, col: 7 });
  expect(startHex(view, 1, map)).toEqual({ line: 1, col: 1 });
  expect(
    startHex(
      { buildings: new Map(), units: new Map() } as unknown as play.View,
      0,
      map,
    ),
  ).toEqual({ line: 4, col: 5 });
});
