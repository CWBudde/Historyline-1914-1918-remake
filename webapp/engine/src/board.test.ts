// Twin of internal/dosrules/board_test.go: the typed API is checked against the recorded fixtures: the records
// it decodes must be the recorded bytes, and its routes, refusals and occupants what the record-level outcome
// holds. Go's seen counts run over both builds; the twin asserts them exactly.

import { describe, expect, it } from "vitest";
import {
  ErrNoMovement,
  type Live,
  type Moved,
  type Phased,
  decodeHold,
  decodeLive,
  destinations,
  exit,
  liveCell,
  liveRecord,
  move,
  profileChangePhase,
  route,
  season,
  winter,
  zeroLive,
} from "./board.ts";
import { type Hold, holdRecord } from "./container.ts";
import { attempt, errorsIs } from "./goruntime.ts";
import {
  type FixtureRecord,
  type StepExpected,
  containerMovement,
  fixtureMovement,
  movementQuery,
  visibilityPhaseChange,
  type VisibilityExpected,
} from "./motionFixtures.ts";
import { type Movement, liveLen, newMovement, pieceAt } from "./move.ts";
import { newProfile } from "./profile.ts";
import { load } from "./testing/fixtures.ts";
import {
  decodeContainerFixture,
  decodeMovementFixture,
  decodeStepFixture,
  decodeVisibilityFixture,
} from "./testing/motionFixtures.ts";
import { Demo, Full, Profiles, moveTables } from "./testing/tables.ts";

// record is a recorded record's bytes.
function record(r: FixtureRecord): Uint8Array {
  return Uint8Array.from(r.bytes ?? []);
}

function bytesEqual(a: Uint8Array, b: Uint8Array): boolean {
  return a.length === b.length && a.every((v, i) => v === b[i]);
}

function hex(b: Uint8Array): string {
  return Array.from(b, (v) => v.toString(16).padStart(2, "0")).join(" ");
}

// Seen counts the behaviours a test sees and collects the failures it finds (Go's t.Errorf).
class Seen {
  readonly counts = new Map<string, number>();
  readonly errors: string[] = [];
  count(name: string): void {
    this.counts.set(name, (this.counts.get(name) ?? 0) + 1);
  }
  // need is board_test.go need: no failure, and every named behaviour seen exactly as often as Go sees it.
  need(want: Readonly<Record<string, number>>): void {
    expect(this.errors).toStrictEqual([]);
    expect(Object.fromEntries(this.counts)).toStrictEqual(want);
  }
}

// TestLiveRoundTrip reads and writes back every recorded live and container record.
describe("live round trip", () => {
  it("reads and writes back every recorded live and container record", () => {
    let lives = 0;
    let holds = 0;
    for (const [build] of Profiles) {
      const live: FixtureRecord[] = [];
      const hold: FixtureRecord[] = [];
      for (const f of load(build, "steps.json", decodeStepFixture)) {
        live.push(...(f.expected.live ?? []));
      }
      for (const f of load(build, "containers.json", decodeContainerFixture)) {
        live.push(...(f.expected.live ?? []));
        hold.push(...(f.expected.holds ?? []));
      }
      for (const f of load(build, "visibility.json", decodeVisibilityFixture)) {
        live.push(...(f.expected.live ?? []));
        hold.push(...(f.expected.holds ?? []));
      }
      for (const r of live) {
        const b = record(r);
        const back = liveRecord(decodeLive(b));
        expect(
          bytesEqual(back, b),
          `${build}: live record ${r.index} ${hex(b)} reads back as ${hex(back)}`,
        ).toBe(true);
        lives++;
      }
      for (const r of hold) {
        const b = record(r);
        const back = holdRecord(decodeHold(b));
        expect(
          bytesEqual(back, b),
          `${build}: container record ${r.index} ${hex(b)} reads back as ${hex(back)}`,
        ).toBe(true);
        holds++;
      }
    }
    expect([lives, holds]).toStrictEqual([7064, 582]);
  });
});

describe("live fields", () => {
  it("names the bytes of the record", () => {
    const b = Uint8Array.from({ length: liveLen }, (_, i) => i + 1);
    const u = decodeLive(b);
    expect(u.size).toBe(3);
    expect(u.flags4).toBe(0x0605);
    expect(u.type).toBe(9);
    expect(u.hold).toBe(11);
    expect(u.pos).toStrictEqual([0x0d0c, 0x0f0e]);
    expect(u.facing).toStrictEqual([16, 17]);
    expect(u.target).toBe(0x1312);
    expect(u.action[6]).toBe(26);
    const c: Live = { ...zeroLive(), pos: [2 * 57 + 1, 2 * 12 + 1] };
    expect([liveCell(c, 0), liveCell(c, 1)]).toStrictEqual([57, 12]);
  });
});

describe("no movement tables", () => {
  it("Move without tables gives ErrNoMovement", () => {
    const p = newProfile({
      name: "dos-0.1",
      tables: { build: "", types: [], terms: [], move: null },
    });
    const [, err] = attempt(() =>
      move(p, newMovement({}), { goal: 0, retarget: 0 }),
    );
    expect(errorsIs(err, ErrNoMovement), String(err)).toBe(true);
  });
});

// The port's own check of Season and Winter (query_test.go reads them through a game, L5): winter takes the
// build's winter table, the summer profile is left as it is, and a build without one has no movement then.
describe("season", () => {
  it("gives the winter terrain records", () => {
    expect(season(Full, false)).toBe(Full);
    const w = season(Full, true);
    expect(winter(w)).toBe(true);
    expect(winter(Full)).toBe(false);
    expect(season(w, true)).toBe(w);
    expect(w.tables?.move?.terrains).toBe(moveTables(Full).winter);
    expect(moveTables(Full).terrains).not.toBe(moveTables(Full).winter);
    expect(w.tables?.types).toBe(Full.tables?.types);
    const d = season(Demo, true);
    expect(winter(d)).toBe(true);
    const [, err] = attempt(() => destinations(d, newMovement({})));
    expect(errorsIs(err, ErrNoMovement), String(err)).toBe(true);
  });
});

// TestTypedRoutes asks every recorded movement query for its destinations and routes through the typed API.
describe("typed routes", () => {
  it("gives every recorded destination list and route", () => {
    const seen = new Seen();
    for (const [build, p0] of Profiles) {
      for (const f of load(build, "movements.json", decodeMovementFixture)) {
        const p = season(p0, f.winter ?? false);
        const g = movementQuery(f);
        const x = f.expected.destinations;
        if (x !== null) {
          const [got, err] = attempt(() => destinations(p, g));
          if (err !== null || !sliceEqual(got, x)) {
            seen.errors.push(
              `${f.id}: destinations ${String(got)} (${String(err)}), recorded ${String(x)}`,
            );
          }
          seen.count("destinations");
        }
        for (const [k, goal] of f.goals.entries()) {
          const [got, err] = attempt(() => route(p, g, goal));
          const want = f.expected.paths[k]?.slice().reverse() ?? null;
          if (err !== null || !sliceEqual(got, want)) {
            seen.errors.push(
              `${f.id}: route to ${goal} is ${String(got)} (${String(err)}), recorded ${String(want)}`,
            );
          }
          if (
            want !== null &&
            got !== null &&
            got.length > 0 &&
            got[0] === pieceAt(g, g.unit).cell &&
            got[got.length - 1] === goal
          ) {
            seen.count("route from the unit to the goal");
          }
        }
      }
    }
    seen.need({
      destinations: 1513,
      "route from the unit to the goal": 16564,
    });
    // A reach and a path search per goal, some 17000 of them: seconds, not Vitest's default 5.
  }, 60_000);
});

// sliceEqual is Go's slices.Equal: nil equals an empty slice.
function sliceEqual(
  a: readonly number[] | null,
  b: readonly number[] | null,
): boolean {
  const x = a ?? [];
  const y = b ?? [];
  return x.length === y.length && x.every((v, i) => v === y[i]);
}

function mapEqual(
  a: ReadonlyMap<number, number>,
  b: ReadonlyMap<number, number>,
): boolean {
  return a.size === b.size && [...a].every(([k, v]) => b.get(k) === v);
}

// movedAgrees compares a typed outcome with the recorded one; it returns the difference, or null.
function movedAgrees(
  g: Movement,
  m: Moved,
  x: StepExpected,
  seen: Seen,
): string | null {
  const refused = (x.order & 1) === 0 || (x.start & 1) === 0;
  if (m.refused !== refused) {
    return `refused ${m.refused}, recorded order ${x.order} start ${x.start}`;
  }
  if (m.twoCell !== ((x.order & 8) !== 0) || !sliceEqual(m.route, x.cells)) {
    return `two-cell ${m.twoCell} route ${String(m.route)}, recorded order ${x.order} cells ${String(x.cells)}`;
  }
  const msgs = x.messages ?? [];
  if (
    m.messages.length !== msgs.length ||
    m.messages.some((msg, i) => {
      const want = msgs[i];
      return want === undefined || msg.code !== want[0] || msg.side !== want[1];
    })
  ) {
    return `messages ${JSON.stringify(m.messages)}, recorded ${JSON.stringify(msgs)}`;
  }
  const live = x.live ?? [];
  if (m.units.size !== live.length) {
    return `${m.units.size} units, recorded ${live.length}`;
  }
  for (const r of live) {
    const got = liveRecord(m.units.get(r.index) ?? zeroLive());
    if (!bytesEqual(got, record(r))) {
      return `unit ${r.index} is ${hex(got)}, recorded ${hex(record(r))}`;
    }
  }
  for (let k = 0; k < 2; k++) {
    const want = new Map(x.occupants[k] ?? []);
    const got = m.occupants[k] ?? new Map<number, number>();
    if (!mapEqual(got, want)) {
      return `map ${k} occupants ${JSON.stringify([...got])}, recorded ${JSON.stringify([...want])}`;
    }
  }
  const u = m.units.get(g.unit) ?? zeroLive();
  if (m.refused) {
    seen.count("refused");
  } else if (m.twoCell) {
    seen.count("two-cell ship");
  } else if (
    m.route.length > 0 &&
    liveCell(u, g.side) === m.route[m.route.length - 1]
  ) {
    seen.count("arrived at the route's end");
  }
  return null;
}

describe("typed moves", () => {
  it("agree with every recorded move order", () => {
    const seen = new Seen();
    for (const [build, p] of Profiles) {
      for (const f of load(build, "steps.json", decodeStepFixture)) {
        const g = fixtureMovement(f);
        const [m, err] = attempt(() => move(p, g, f.order));
        const diff =
          m === null ? String(err) : movedAgrees(g, m, f.expected, seen);
        if (diff !== null) {
          seen.errors.push(`${f.id}: ${diff}`);
        }
      }
    }
    seen.need({
      "arrived at the route's end": 1460,
      refused: 16,
      "two-cell ship": 10,
    });
  });
});

describe("typed containers", () => {
  it("agree with every recorded load, unload and exit", () => {
    const seen = new Seen();
    for (const [build, p] of Profiles) {
      for (const f of load(build, "containers.json", decodeContainerFixture)) {
        const g = containerMovement(f);
        const [m, err] = attempt(() =>
          f.order.hold < 0
            ? move(p, g, {
                goal: f.order.goal,
                retarget: f.order.retarget,
                ...(f.order.stale === undefined
                  ? {}
                  : { stale: f.order.stale }),
              })
            : exit(p, g, f.order),
        );
        let diff =
          m === null ? String(err) : movedAgrees(g, m, f.expected, seen);
        if (diff === null && m !== null) {
          diff = holdsAgree(m.holds, f.expected.holds);
        }
        if (diff === null && m !== null && f.order.hold >= 0) {
          const want = new Map(f.expected.remaining ?? []);
          const got = m.remaining ?? new Map<number, number>();
          if (!mapEqual(got, want)) {
            diff = `exit reach ${JSON.stringify([...got])}, recorded ${JSON.stringify([...want])}`;
          }
          seen.count("exit");
        }
        if (diff !== null) {
          seen.errors.push(`${f.id}: ${diff}`);
        }
        if (m !== null && m.holds.size > 0) {
          seen.count("container records changed");
        }
      }
    }
    seen.need({
      "arrived at the route's end": 532,
      "container records changed": 440,
      exit: 156,
      refused: 524,
    });
  });
});

function holdsAgree(
  got: ReadonlyMap<number, Hold>,
  want: FixtureRecord[] | null,
): string | null {
  const w = want ?? [];
  if (got.size !== w.length) {
    return `${got.size} container records changed, recorded ${w.length}`;
  }
  for (const r of w) {
    const h = got.get(r.index);
    if (h === undefined || !bytesEqual(holdRecord(h), record(r))) {
      return `container record ${r.index} is ${JSON.stringify(h)}, recorded ${hex(record(r))}`;
    }
  }
  return null;
}

describe("typed phase changes", () => {
  it("agree with every recorded phase change", () => {
    const seen = new Seen();
    for (const [build, p] of Profiles) {
      for (const f of load(build, "visibility.json", decodeVisibilityFixture)) {
        const g = visibilityPhaseChange(f);
        const [ph, err] = attempt(() => profileChangePhase(p, g));
        const diff =
          ph === null ? String(err) : phasedAgrees(ph, f.expected, seen);
        if (diff !== null) {
          seen.errors.push(`${f.id}: ${diff}`);
        }
      }
    }
    seen.need({
      "container records changed": 62,
      "unit in both views": 778,
      "unit left the views": 704,
    });
  });
});

function phasedAgrees(
  ph: Phased,
  x: VisibilityExpected,
  seen: Seen,
): string | null {
  for (const r of x.live ?? []) {
    const u = ph.units.get(r.index);
    if (u === undefined || !bytesEqual(liveRecord(u), record(r))) {
      return `unit ${r.index} is ${JSON.stringify(u)}, recorded ${hex(record(r))}`;
    }
    if ((u.flags4 & 0xc000) !== 0) {
      seen.count("unit left the views");
    } else if (
      (ph.occupants[0].get(liveCell(u, 0)) ?? 0) === r.index &&
      (ph.occupants[1].get(liveCell(u, 0)) ?? 0) === r.index
    ) {
      seen.count("unit in both views");
    }
  }
  const diff = holdsAgree(ph.holds, x.holds);
  if (diff !== null) {
    return diff;
  }
  if (ph.holds.size > 0) {
    seen.count("container records changed");
  }
  for (let k = 0; k < 2; k++) {
    const want = x.players[k]?.[0x16] ?? 0;
    if (ph.phase[k] !== want) {
      return `side ${k} phase ${ph.phase[k] ?? 0}, recorded ${want}`;
    }
  }
  const msgs = x.messages ?? [];
  if (ph.count !== x.count || ph.messages.length !== msgs.length) {
    return `count ${ph.count} messages ${JSON.stringify(ph.messages)}, recorded ${x.count} ${JSON.stringify(msgs)}`;
  }
  return null;
}
