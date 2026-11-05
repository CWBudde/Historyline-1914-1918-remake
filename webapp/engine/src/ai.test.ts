import { describe, expect, it } from "vitest";
import * as ai from "./ai.ts";
import * as play from "./play.ts";
import { type Level, type Hex } from "./level.ts";
import { type Coord, distance, around } from "./hexgrid.ts";
import { cell } from "./bind.ts";
import { Full, Builds } from "./testing/tables.ts";
import { playData, shippedLevel } from "./testing/play.ts";
import { at } from "./goruntime.ts";

function grass(units: readonly [Coord, string][]): play.Session {
  const lv: Level = {
    mapName: "grass",
    mapNo: 0,
    player: "Central Powers",
    extra: null,
    lines: Array.from({ length: 21 }, (_, l) =>
      Array.from({ length: 8 - (l % 2) }, (): Hex => ({
        terrain: "grass",
        infra: { ids: [], list: false },
        unit: "",
        content: null,
      })),
    ),
  };
  for (const [c, id] of units) at(at(lv.lines, c.line), c.col).unit = id;
  return play.newSession(lv, playData, {
    profile: Full,
    computer: [true, true],
  });
}
describe("original-profile computer", () => {
  it("returns the events already planned when a human move limit refuses the next attack", () => {
    const s = grass([
      [{ line: 4, col: 3 }, "EC_infantry"],
      [{ line: 4, col: 4 }, "EC_infantry"],
      [{ line: 5, col: 3 }, "CP_supplyCar"],
      [{ line: 5, col: 4 }, "CP_supplyCar"],
    ]);
    s.game.players[1].computer = false;
    s.game.limit = 1;
    let caught: unknown;
    try {
      ai.plan({}, s, 1);
    } catch (e) {
      caught = e;
    }
    expect(caught).toBeInstanceOf(ai.PlanError);
    if (caught instanceof ai.PlanError)
      expect(caught.events.map((e) => e.kind)).toEqual([play.Ordered]);
    expect(play.ready(s, 1)).toBe(false);
  });
  it("reports progress without changing a plan, validates sides and leaves confirmed sides alone", () => {
    const a = play.newSession(shippedLevel("1914-08"), playData, {
        profile: Full,
      }),
      b = play.clone(a, Builds);
    const progress: [number, number][] = [];
    expect(
      ai.plan(
        {
          progress: (done, total) => {
            progress.push([done, total]);
          },
        },
        a,
        0,
      ),
    ).toEqual(ai.plan({}, b, 0));
    expect(progress[0]?.[0]).toBe(0);
    expect(progress.at(-1)?.[0]).toBe(progress.at(-1)?.[1]);
    expect(progress.length).toBeGreaterThan(1);
    expect(ai.plan({}, a, 0)).toEqual([]);
    expect(() => ai.plan({}, a, 2)).toThrow("dosai: side 2");
    play.surrender(a, 1);
    expect(ai.plan({}, a, 1)).toEqual([]);
  });
  it("plans both phases on independent clones while the other side gives orders", () => {
    const a = play.newSession(shippedLevel("1914-08"), playData, {
        profile: Full,
        computer: [true, true],
      }),
      b = play.clone(a, Builds);
    for (let n = 0; n < 8 && play.ended(a) === null; n++) {
      const want = ai.plan({}, a, 1),
        human = ai.plan({}, a, 0),
        c = play.clone(b, Builds);
      const cmds = ai.commands({}, c, 1);
      expect(ai.plan({}, b, 0)).toEqual(human);
      expect(play.replay(b, cmds)).toEqual(want);
      expect(cmds.at(-1)?.kind).toBe("confirm");
      expect(play.save(b)).toEqual(play.save(a));
      expect(play.resolve(b)).toEqual(play.resolve(a));
    }
  });
  it("PlayOut gives the same state as explicit planning and allows a passing side", () => {
    const a = play.newSession(shippedLevel("1914-08"), playData, {
        profile: Full,
      }),
      b = play.clone(a, Builds);
    ai.playOut(a, [{}, null], 4);
    for (let n = 0; n < 4 && play.ended(b) === null; n++) {
      ai.plan({}, b, 0);
      play.confirm(b, 1);
      play.resolve(b);
    }
    expect(play.save(a)).toEqual(play.save(b));
  });
  it("digs towards a nearby enemy, resolves work and never fills another sapper's trench", () => {
    const from = { line: 4, col: 3 },
      enemy = { line: 12, col: 3 };
    const s = grass([
      [from, "EC_sappers"],
      [enemy, "CP_infantry"],
    ]);
    const tasked = ai.plan({}, s, 1).filter((e) => e.kind === play.Tasked);
    expect(tasked).toHaveLength(1);
    expect(distance(from, at(tasked, 0).to)).toBe(1);
    expect(distance(at(tasked, 0).to, enemy)).toBeLessThan(
      distance(from, enemy),
    );
    play.confirm(s, 0);
    expect(play.resolve(s).find((e) => e.kind === play.Dug)?.text).toBe(
      "Trench dug.",
    );
    const two = grass([
      [from, "EC_sappers"],
      [{ line: 4, col: 4 }, "EC_sappers"],
      [{ line: 10, col: 4 }, "CP_infantry"],
    ]);
    expect(
      ai.plan({}, two, 1).filter((e) => e.kind === play.Tasked),
    ).toHaveLength(2);
    play.confirm(two, 0);
    for (const ev of play.resolve(two).filter((e) => e.kind === play.Dug))
      expect(ev.text).toBe("Trench dug.");
  });
  it("keeps distant sappers idle and does not fill surrounding trenches", () => {
    const far = grass([
      [{ line: 1, col: 3 }, "EC_sappers"],
      [{ line: 19, col: 3 }, "CP_infantry"],
    ]);
    expect(ai.plan({}, far, 1).some((e) => e.kind === play.Tasked)).toBe(false);
    const from = { line: 6, col: 3 },
      s = grass([
        [from, "EC_sappers"],
        [{ line: 12, col: 3 }, "CP_infantry"],
      ]);
    for (const n of around(from)) {
      const c = cell(s.board, n);
      if (c !== null) for (const m of s.game.maps) m[2 * c] = 0xa7;
    }
    expect(ai.plan({}, s, 1).some((e) => e.kind === play.Tasked)).toBe(false);
  });
  it("computes reverse entry costs through the hex graph, leaving closed cells absent", () => {
    expect(
      ai.dijkstra(
        new Map([
          ["0,0", 4],
          ["2,0", 7],
          ["4,0", 9],
        ]),
        new Map([["0,0", 3]]),
      ),
    ).toEqual(
      new Map([
        ["0,0", 3],
        ["2,0", 7],
        ["4,0", 14],
      ]),
    );
  });
});
