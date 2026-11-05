// Twin of internal/dosrules/save_test.go and save_golden_test.go: the saves Go writes (testdata/saves) load in
// the port and re-save byte for byte, the port's own games save to the same bytes, a loaded game plays on as the
// saved one, and saves that do not fit together are refused with Go's messages. testing/goLoadCases.ts holds
// LoadGame's results for odd inputs, recorded from Go.

import { describe, expect, it } from "vitest";
import { season } from "./board.ts";
import { holdFlag40 } from "./container.ts";
import { type Game, confirm, evBattle, resolve } from "./cycle.ts";
import { GoError, at } from "./goruntime.ts";
import { loadGame, save, saveSchema } from "./save.ts";
import { score } from "./victory.ts";
import { listDir, readBytes } from "./testing/files.ts";
import { goLoadCases } from "./testing/goLoadCases.ts";
import { containerScene, midRound, reload } from "./testing/scenes.ts";
import {
  Builds,
  Demo,
  Full,
  Profiles,
  cInfantry,
  cSupply,
} from "./testing/tables.ts";

const savesDir = "internal/dosrules/testdata/saves";
const utf8 = (s: string): Uint8Array => new TextEncoder().encode(s);
const text = (b: Uint8Array): string =>
  new TextDecoder("utf-8", { fatal: true }).decode(b);

/** bytesEqual reports whether a and b hold the same bytes. */
function bytesEqual(a: Uint8Array, b: Uint8Array): boolean {
  return a.length === b.length && a.every((v, i) => v === b[i]);
}

/** loadError is the message of the GoError loadGame throws on data, "" where it loads. */
function loadError(data: Uint8Array): string {
  try {
    loadGame(data, Builds);
    return "";
  } catch (e) {
    expect(e).toBeInstanceOf(GoError);
    return e instanceof Error ? e.message : String(e);
  }
}

/** edit is TestLoadGameRefuses' edit: the save as generic JSON, changed by f and written again. */
function edit(
  b: Uint8Array,
  f: (m: Record<string, unknown>) => void,
): Uint8Array {
  const m = JSON.parse(text(b)) as Record<string, unknown>;
  f(m);
  return utf8(JSON.stringify(m));
}

const obj = (v: unknown): Record<string, unknown> =>
  v as Record<string, unknown>;
const arr = (v: unknown): unknown[] => v as unknown[];

const goldens = listDir(savesDir)
  .filter((f) => f.endsWith(".json"))
  .sort();

describe("the saves Go writes", () => {
  it("are six", () => {
    expect(goldens).toStrictEqual([
      "demo-ended.json",
      "demo-midround.json",
      "demo-resolved.json",
      "full-midround.json",
      "full-resolved.json",
      "full-winter-containers.json",
    ]);
  });
  it.each(goldens)("%s loads and re-saves byte for byte", (name) => {
    const want = readBytes(`${savesDir}/${name}`);
    const got = save(loadGame(want, Builds));
    expect(text(got)).toBe(text(want));
    expect(bytesEqual(got, want)).toBe(true);
  });
  it("load from a string too", () => {
    const want = readBytes(`${savesDir}/demo-ended.json`);
    expect(bytesEqual(save(loadGame(text(want), Builds)), want)).toBe(true);
  });
});

// TestSaveGolden: the port's own games write the bytes Go wrote.
describe("TestSaveGolden", () => {
  const cases: Record<string, () => Game> = {
    "full-winter-containers": () => containerScene(season(Full, true)),
    "demo-ended": () => {
      const g = containerScene(Demo);
      g.buildings = new Map([
        [holdFlag40, { credits: 7, builds: [cSupply, cInfantry] }],
      ]);
      g.energy = [120, 35];
      g.players[1].computer = true;
      g.ended = { over: true, winner: 0, won: ["", ""], score: score(g) };
      return g;
    },
  };
  for (const [build, p] of Profiles) {
    cases[`${build}-midround`] = () => midRound(p);
    cases[`${build}-resolved`] = () => {
      const g = midRound(p);
      confirm(g, 0);
      resolve(g);
      return g;
    };
  }
  it("has a case per golden file", () => {
    expect(
      Object.keys(cases)
        .map((k) => `${k}.json`)
        .sort(),
    ).toStrictEqual(goldens);
  });
  it.each(Object.entries(cases))("%s", (name, scene) => {
    const want = readBytes(`${savesDir}/${name}.json`);
    const got = save(scene());
    expect(text(got)).toBe(text(want));
    expect(bytesEqual(got, want)).toBe(true);
    expect(loadError(want), `${name} does not load`).toBe("");
  });
});

// TestSaveResume saves a round with an unconfirmed side and a confirmed side with a queued attack: the loaded
// game equals the saved one, and both run the rest of the round to the same events and state.
describe("TestSaveResume", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = midRound(p);
    const [p0, p1] = g.players;
    expect(
      [p0.ready, p1.ready, p0.orders.length, p1.orders.length],
      "set-up",
    ).toStrictEqual([false, true, 1, 1]);
    const h = reload(g);
    expect(h, "loaded game differs").toStrictEqual(g);

    const evs = [g, h].map((x) => {
      confirm(x, 0);
      return resolve(x);
    });
    const [e0, e1] = evs;
    expect(e0?.length).toBeGreaterThan(0);
    expect(e0?.[0]?.kind).toBe(evBattle);
    expect(e1, "events differ").toStrictEqual(e0);
    expect(h, "state after the round differs").toStrictEqual(g);

    // A save after the phase change reloads equal too.
    expect(reload(g), "reloaded after the round differs").toStrictEqual(g);
  });
});

// TestLoadGameRefuses checks that a save that does not fit together is refused rather than loaded, with Go's
// messages.
describe("TestLoadGameRefuses", () => {
  const b = save(midRound(Full));
  it.each([
    [
      "another schema",
      (m: Record<string, unknown>) => {
        m["schema"] = 1;
      },
      "dosrules: load: schema 1, want 3",
    ],
    [
      "an unknown build",
      (m: Record<string, unknown>) => {
        m["build"] = "00000000";
      },
      'dosrules: load: unknown tables "00000000"',
    ],
    [
      "a truncated map",
      (m: Record<string, unknown>) => {
        const mp = arr(m["maps"]);
        mp[1] = arr(mp[1]).slice(0, 10);
      },
      "dosrules: load: map 1 has 10 entries for 81 cells",
    ],
    [
      "an unknown type",
      (m: Record<string, unknown>) => {
        obj(obj(m["units"])["0"])["type"] = 99;
      },
      "dosrules: load: unit 0 of type 99",
    ],
    [
      "a missing occupant",
      (m: Record<string, unknown>) => {
        delete obj(m["units"])["1"];
      },
      "dosrules: load: map 0 shows unit 1 at 31",
    ],
    [
      "an unknown order",
      (m: Record<string, unknown>) => {
        obj(arr(m["players"])[1])["orders"] = [77];
      },
      "dosrules: load: side 1 orders unit 77",
    ],
    [
      "a phase pair",
      (m: Record<string, unknown>) => {
        obj(arr(m["players"])[0])["phase"] = 2;
      },
      "dosrules: load: dosrules: phases 2/2",
    ],
  ] as const)("%s", (_what, f, msg) => {
    expect(loadError(edit(b, f))).toBe(msg);
  });
});

describe("Save", () => {
  it("writes schema 3", () => {
    expect(saveSchema).toBe(3);
  });
  it("refuses a profile without tables", () => {
    const g = midRound(Full);
    g.profile = { ...g.profile, tables: null };
    expect(() => save(g)).toThrow(new GoError("dosrules: save: no tables"));
  });
});

// LoadGame against Go's results on odd inputs: the same error message, or the same bytes saved again.
describe("LoadGame as Go's", () => {
  it("has the recorded cases", () => {
    expect(goLoadCases).toHaveLength(166);
  });
  it.each(goLoadCases.map((c, i) => [i, c] as const))("case %i", (_i, c) => {
    const data =
      c.hex === undefined
        ? utf8(c.in ?? "")
        : new Uint8Array(
            (c.hex.match(/../g) ?? []).map((h) => Number.parseInt(h, 16)),
          );
    if (c.err !== undefined) {
      expect(loadError(data)).toBe(c.err);
      return;
    }
    const got = save(loadGame(data, Builds));
    expect(text(got)).toBe(c.out);
    expect(bytesEqual(got, utf8(c.out ?? ""))).toBe(true);
  });
  it("refuses an integer past the port's range that Go's int64 holds", () => {
    expect(loadError(utf8('{"turn":9007199254740993}'))).toMatch(
      /^dosrules: load: json: number 9007199254740993 is outside the port's integer range/,
    );
  });
});

describe("loadGame", () => {
  it("keeps the units' map in key order and copies nothing it was given", () => {
    const b = readBytes(`${savesDir}/full-winter-containers.json`);
    const g = loadGame(b, Builds);
    expect([...g.units.keys()]).toStrictEqual([0, 1, 5, 6, 7, 10]);
    expect([...g.holds.keys()]).toStrictEqual([3, 70, 71]);
    expect(at(g.maps, 0)).not.toBe(at(g.maps, 1));
  });
});
