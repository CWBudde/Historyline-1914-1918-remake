import { expect, it } from "vitest";
import { play, level, phaseMovement } from "@historyline/engine";
import { Full, Builds } from "../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../engine/src/testing/play.ts";
import { MapBattle } from "./mapBattle.ts";

function start() {
  return play.newSession(shippedLevel("1914-08"), playData, {
    profile: Full,
    computer: [false, true],
  });
}

it("opens a cancellable ready prompt, withdraws ready while the worker thinks and resolves once", async () => {
  const s = start();
  let release: () => void = () => {},
    calls = 0,
    notices = 0;
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  const b = new MapBattle(s, 0, {
    commands: async (session, side, progress) => {
      calls++;
      progress?.(1, 3);
      const snapshot = play.clone(session, Builds),
        mark = play.mark(snapshot);
      await gate;
      play.confirm(snapshot, side);
      return play.commandsSince(snapshot, mark);
    },
  });
  b.onChange = () => {
    notices++;
  };
  const mark = play.mark(s);
  await b.requestPhase();
  expect(b.phasePrompt).toBe(true);
  expect(play.mark(s)).toBe(mark);
  b.cancel();
  expect(b.phasePrompt).toBe(false);
  await b.requestPhase();
  const pending = b.requestPhase();
  expect(play.ready(s, 0)).toBe(true);
  expect(b.progress).toEqual({ done: 1, total: 3 });
  await b.requestPhase();
  expect(play.ready(s, 0)).toBe(false);
  release();
  await pending;
  expect(b.busy()).toBe(false);
  expect(calls).toBe(1);
  await b.requestPhase();
  await b.requestPhase();
  expect(play.phase(s, 0)).not.toBe(phaseMovement);
  expect(calls).toBe(1);
  expect(notices).toBeGreaterThan(3);
});

it("keeps info and native action placement fair, and cancels aiming before opening a ring", () => {
  const b = new MapBattle(start(), 0),
    own = [...b.view().units.values()].find((u) => u.side === 0),
    enemy = [...b.view().units.values()].find((u) => u.side === 1);
  if (!own || !enemy) throw Error("missing units");
  b.showInfo(enemy.at);
  expect(b.panel).toBe("orders");
  expect(b.infoUnit).toBeNull();
  expect(b.status).toMatch(/own/i);
  b.showInfo(own.at);
  expect(b.panel).toBe("unit");
  expect(b.infoUnit?.index).toBe(own.index);
  b.cancel();
  expect(b.panel).toBe("orders");
  expect(b.actions(own.at)).toContain("info");
  b.click(own.at);
  expect(b.actions(own.at)).toEqual([]);
  b.run("cancel", own.at);
  expect(b.selected).toBeNull();
});

it("opens an owned building through any part and aims its held unit's exit", () => {
  const lv = shippedLevel("1914-08");
  let entrance: { line: number; col: number } | null = null;
  for (const [line, row] of lv.lines.entries())
    for (const [col, h] of row.entries())
      if (entrance === null && h.content?.owner === 1) {
        entrance = { line, col };
        h.content.units = ["CP_infantry"];
        h.unit = "";
      }
  if (entrance === null) throw Error("missing building");
  const s = play.newSession(lv, playData, { profile: Full }),
    b = new MapBattle(s, 0);
  const part =
    play
      .footprint(s, entrance)
      .find(
        (c) =>
          level.coordKey(c) !== level.coordKey(entrance) &&
          !b.view().units.has(level.coordKey(c)),
      ) ?? entrance;
  b.click(part);
  expect(b.contents.box?.at).toEqual(entrance);
  const unit = b.contents.unit();
  if (unit === null) throw Error("missing held unit");
  b.run("exit", entrance);
  expect(b.contents.box).toBeNull();
  expect(b.exiting?.unit.index).toBe(unit.index);
  const to = play.coords(b.destinations())[0];
  if (to === undefined) throw Error("no exit destination");
  expect(b.route(to).at(-1)).toEqual(to);
  b.click(to);
  expect(play.unit(s, unit.index).held).toBe(false);
  expect(b.exiting).toBeNull();
  expect(b.playback.moving()?.unit.index).toBe(unit.index);
});

it("routes keyboard cursor and fire to open contents instead of the map", () => {
  const lv = shippedLevel("1914-08");
  let entrance: { line: number; col: number } | null = null;
  for (const [line, row] of lv.lines.entries())
    for (const [col, h] of row.entries())
      if (entrance === null && h.content?.owner === 1) {
        entrance = { line, col };
        h.content.units = ["CP_infantry", "CP_infantry"];
        h.unit = "";
      }
  if (entrance === null) throw Error("missing building");
  const b = new MapBattle(play.newSession(lv, playData, { profile: Full }), 0);
  b.openContents(entrance);
  const cursor = { ...b.cursor };
  b.stepCursor("right");
  expect(b.contents.box?.slot).toBe(1);
  expect(b.cursor).toEqual(cursor);
  b.fire();
  expect(b.contents.box).toBeNull();
  expect(b.exiting?.slot).toBe(1);
});

it("gives each unit its own orders on fixed slots: absent where they do not apply, a reason where not possible", () => {
  const b = new MapBattle(start(), 0),
    units = [...b.view().units.values()],
    enemy = units.find((u) => u.side === 1),
    bunker = units.find(
      (u) => u.side === 0 && play.typeInfo(b.session, 0, u.type)?.speed === 0,
    ),
    mover = units.find(
      (u) =>
        u.side === 0 &&
        !u.carrier &&
        (play.typeInfo(b.session, 0, u.type)?.speed ?? 0) > 0 &&
        play.destinations(b.session, 0, u.index).size > 0,
    );
  if (!enemy || !mover) throw Error("missing units");
  // An enemy's cell and a never-moving unit's cell (in the movement mode) have no menu.
  expect(b.menu(enemy.at)).toEqual([]);
  if (bunker !== undefined && play.contents(b.session, 0, bunker.at) === null)
    expect(b.menu(bunker.at)).toEqual([]);
  // A unit that can move: Move up, ready; the strip's orders are never on the menu.
  const items = b.menu(mover.at);
  expect(items[0]).toEqual({ slot: "n", action: "move", reason: null });
  expect(items.map((i) => i.action)).not.toContain("mode");
  expect(items.map((i) => i.action)).not.toContain("map");
  expect(items.map((i) => i.action)).not.toContain("info");
  // Once it has moved, Move stays on its slot with the reason it is not possible.
  const to = [...play.destinations(b.session, 0, mover.index)][0];
  if (to === undefined) throw Error("no destination");
  const [line = 0, col = 0] = to.split(",").map(Number);
  play.move(b.session, 0, mover.index, { line, col });
  const moved = [...b.view().units.values()].find(
    (u) => u.index === mover.index,
  );
  if (moved === undefined) throw Error("lost the unit");
  const after = b.menu(moved.at).find((i) => i.slot === "n");
  expect(after?.action).toBe("move");
  expect(after?.reason).not.toBeNull();
  // A prompt hides the menu.
  b.phasePrompt = true;
  expect(b.menu(moved.at)).toEqual([]);
});
