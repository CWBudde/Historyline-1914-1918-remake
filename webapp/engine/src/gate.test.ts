// Twin of internal/dosrules/gate_test.go: TestGate on each recorded build, its casualty clauses (repeatable,
// fixed-point, experience terms, divide error) on casualties.json and its engagement clauses (losses,
// experience, blocking, pincer, ranged support, battle distance) on engagements.json. Every clause reads the
// recorded expectations, never the resolver's own results.

import { describe, expect, it } from "vitest";
import { ErrDivideError, casualties, percent } from "./casualties.ts";
import { distance } from "./distance.ts";
import {
  type Battlefield,
  type Unit,
  engage,
  strength,
  unitOf,
} from "./engage.ts";
import {
  type CasualtyExpected,
  type EngagementExpected,
  type EngagementFixture,
  casualtyInputs,
  casualtyResult,
  engagementBattlefield,
} from "./fixtures.ts";
import { at, attempt, errorsIs } from "./goruntime.ts";
import { type Profile, type Type, newProfile, tablesOf } from "./profile.ts";
import { ErrUndefined, blockCap, blockStep, pincerWeights } from "./support.ts";
import {
  decodeCasualtyFixture,
  decodeEngagementFixture,
  load,
} from "./testing/fixtures.ts";
import { Profiles } from "./testing/tables.ts";

describe.each(Profiles)("gate, %s build", (build, p) => {
  const cs = load(build, "casualties.json", decodeCasualtyFixture);

  // Identical battle states give identical casualty results, in any order.
  it("repeatable", () => {
    const run = (i: number): CasualtyExpected => {
      const f = cs[i];
      if (f === undefined) {
        throw new Error(`no fixture ${i}`);
      }
      const [a, d, c] = casualtyInputs(f, tablesOf(p));
      return casualtyResult(...attempt(() => casualties(p, a, d, c)));
    };
    const first = cs.map((_, i) => run(i));
    for (let i = cs.length - 1; i >= 0; i--) {
      expect(run(i), cs[i]?.id).toStrictEqual(first[i]);
    }
  });

  // Fixed-point rounding: the 16.16 percentage truncates, so 20 with 150% gives 49, and the casualty corpus
  // holds runs at the blocking cap.
  it("fixed-point", () => {
    expect(percent(20, 150)).toBe(49);
    const n = cs.filter(
      (f) => f.context.blocking === blockCap && f.expected.error === undefined,
    ).length;
    expect(n, "no recorded casualty run at blocking 150").toBeGreaterThan(0);
  });

  // Experience terms: the casualty corpus covers experience 0-6 on both sides.
  it("experience terms", () => {
    const a = [0, 0, 0, 0, 0, 0, 0];
    const d = [0, 0, 0, 0, 0, 0, 0];
    for (const f of cs) {
      if (f.expected.error !== undefined) {
        continue;
      }
      if (f.attacker.experience <= 6) {
        a[f.attacker.experience] = (a[f.attacker.experience] ?? 0) + 1;
      }
      if (f.defender.experience <= 6) {
        d[f.defender.experience] = (d[f.defender.experience] ?? 0) + 1;
      }
    }
    for (let e = 0; e < 7; e++) {
      expect(a[e], `experience ${e}: attacker runs`).toBeGreaterThan(0);
      expect(d[e], `experience ${e}: defender runs`).toBeGreaterThan(0);
    }
  });

  // Zero armour is the recorded divide error, and the resolver raises it too.
  it("divide error", () => {
    let n = 0;
    for (const f of cs) {
      if (f.expected.error === undefined) {
        continue;
      }
      n++;
      expect(f.expected.error, f.id).toBe("divide-error");
      const [a, d, c] = casualtyInputs(f, tablesOf(p));
      const [, err] = attempt(() => casualties(p, a, d, c));
      expect(errorsIs(err, ErrDivideError), f.id).toBe(true);
    }
    expect(n, "no recorded divide error").toBeGreaterThan(0);
  });
});

// recorded is a battle of the engagement corpus: the two units before the order and the recorded state after
// it.
interface Recorded {
  id: string;
  b: Battlefield;
  a: Readonly<Unit>;
  d: Readonly<Unit>;
  ta: Type;
  td: Type;
  e: EngagementExpected;
}

function recordedOf(p: Profile, f: EngagementFixture): Recorded {
  const b = engagementBattlefield(f);
  const a = unitOf(b.units, f.attacker);
  const d = unitOf(b.units, f.expected.defender);
  const types = tablesOf(p).types;
  return {
    id: f.id,
    b,
    a,
    d,
    ta: at(types, a.type),
    td: at(types, d.type),
    e: f.expected,
  };
}

function strengths(r: Recorded): [number, number, number, number] {
  return [
    strength(r.ta, r.a.size, r.a.alt),
    strength(r.ta, r.e.attackerSize, r.e.attackerAlternateSize),
    strength(r.td, r.d.size, r.d.alt),
    strength(r.td, r.e.defenderSize, r.e.defenderAlternateSize),
  ];
}

// require fails for every named behaviour the corpus never shows.
function require(seen: Map<string, number>): void {
  for (const name of [...seen.keys()].sort()) {
    expect(seen.get(name), `no recorded battle with ${name}`).toBeGreaterThan(
      0,
    );
  }
}

function counter(
  ...names: string[]
): [Map<string, number>, (name: string) => void] {
  const seen = new Map(names.map((n) => [n, 0]));
  return [seen, (name) => seen.set(name, (seen.get(name) ?? 0) + 1)];
}

describe.each(Profiles)("gate engagements, %s build", (build, p) => {
  const es = load(build, "engagements.json", decodeEngagementFixture);
  const battles = es
    .filter((f) => f.expected.battle && f.expected.error === undefined)
    .map((f) => recordedOf(p, f));

  it("holds the recorded battles", () => {
    expect(es).toHaveLength(446);
    expect(battles.length).toBeGreaterThan(0);
  });

  // Both sides' recorded losses: strength lost and units destroyed, on either side.
  it("losses", () => {
    const [seen, count] = counter(
      "attacker loss",
      "defender loss",
      "attacker destroyed",
      "defender destroyed",
    );
    for (const r of battles) {
      const [beforeA, afterA, beforeD, afterD] = strengths(r);
      expect(
        afterA > beforeA || afterD > beforeD,
        `${r.id}: strength grew (${beforeA}→${afterA}, ${beforeD}→${afterD})`,
      ).toBe(false);
      if (afterA < beforeA) {
        count("attacker loss");
      }
      if (afterD < beforeD) {
        count("defender loss");
      }
      for (const idx of r.e.destroyed ?? []) {
        count(
          idx === r.e.defender ? "defender destroyed" : "attacker destroyed",
        );
      }
    }
    require(seen);
  });

  // Both sides' recorded experience transitions follow the spec's award: the opponent of a side that lost
  // strength gains 1 unless that side's live +6 word has bit 0x20, 1 more when it was destroyed, capped at 6.
  it("experience", () => {
    const [seen, count] = counter(
      "attacker gain",
      "defender gain",
      "gain 1 (loss)",
      "gain 2 (kill)",
      "no award bit",
      "cap at 6",
    );
    const gain = (lost: boolean, dead: boolean, flags6: number): number =>
      (lost && (flags6 & 0x20) === 0 ? 1 : 0) + (dead ? 1 : 0);
    for (const r of battles) {
      const [beforeA, afterA, beforeD, afterD] = strengths(r);
      const gainA = gain(afterD < beforeD, afterD === 0, r.d.flags6);
      const gainD = gain(afterA < beforeA, afterA === 0, r.a.flags6);
      const wantA = Math.min(6, r.a.exp + gainA);
      const wantD = Math.min(6, r.d.exp + gainD);
      expect(
        [r.e.attackerExperience, r.e.defenderExperience],
        `${r.id}: recorded experience against the award rule`,
      ).toStrictEqual([wantA, wantD]);
      for (const s of [
        {
          exp: r.a.exp,
          after: r.e.attackerExperience,
          gain: gainA,
          lost: afterD < beforeD,
          opponent6: r.d.flags6,
          name: "attacker gain",
        },
        {
          exp: r.d.exp,
          after: r.e.defenderExperience,
          gain: gainD,
          lost: afterA < beforeA,
          opponent6: r.a.flags6,
          name: "defender gain",
        },
      ]) {
        if (s.after > s.exp) {
          count(s.name);
        }
        if (s.lost && (s.opponent6 & 0x20) !== 0) {
          count("no award bit");
        } else if (s.gain === 1 && s.exp < 6) {
          count("gain 1 (loss)");
        } else if (s.gain === 2 && s.exp < 5) {
          count("gain 2 (kill)");
        }
        if (s.gain > 0 && s.exp + s.gain > 6) {
          count("cap at 6");
        }
      }
    }
    require(seen);
  });

  // Blocking (manual: a quarter of the strength; DOS: 80 per eligible unit, capped at 150): the recorded
  // values are 0, 80 and 150 only.
  it("blocking", () => {
    const [seen, count] = counter("blocking 80", "blocking 150");
    for (const r of battles) {
      switch (r.e.blocking) {
        case 0:
          break;
        case blockStep:
          count("blocking 80");
          break;
        case blockCap:
          count("blocking 150");
          break;
        default:
          expect.fail(`${r.id}: recorded blocking ${r.e.blocking}`);
      }
    }
    require(seen);
  });

  // Pincer: a byte sum of the weights 25/50/100/50/25.
  it("pincer", () => {
    const sums = new Set<number>();
    for (let mask = 0; mask < 1 << 5; mask++) {
      let s = 0;
      for (let k = 0; k < 5; k++) {
        if ((mask & (1 << k)) !== 0) {
          s += at(pincerWeights, k + 1);
        }
      }
      sums.add(s & 0xff);
    }
    const [seen, count] = counter(
      "pincer 25",
      "pincer 50",
      "pincer 100",
      "pincer above 200",
    );
    for (const r of battles) {
      expect(
        sums.has(r.e.pincer),
        `${r.id}: recorded pincer ${r.e.pincer} is no sum of the weights`,
      ).toBe(true);
      if (r.e.pincer === 25) {
        count("pincer 25");
      } else if (r.e.pincer === 50) {
        count("pincer 50");
      } else if (r.e.pincer === 100) {
        count("pincer 100");
      } else if (r.e.pincer > 200) {
        count("pincer above 200");
      }
    }
    require(seen);
  });

  // Ranged support: a battle that is not adjacent has no pincer and no blocking. The map-edge case, which
  // reads an uninitialised stack byte in the original, stays gated off: the resolver refuses it.
  it("ranged support", () => {
    const [seen, count] = counter("ranged battle");
    for (const r of battles) {
      if (r.e.returnFire) {
        continue;
      }
      count("ranged battle");
      expect(
        [r.e.pincer, r.e.blocking],
        `${r.id}: ranged battle with support`,
      ).toStrictEqual([0, 0]);
    }
    require(seen);

    // A defender on the top row of a 9×9 map, attacked from the centre: its off-map neighbour reads map byte
    // 0, which names the attacker.
    const unit = (cell: number, flags4: number): Unit => ({
      type: 18,
      cell,
      size: 6,
      alt: 0,
      exp: 0,
      flags4,
      flags6: 0,
    });
    const b: Battlefield = {
      width: 9,
      height: 9,
      units: new Map([
        [0, unit(40, 5)],
        [1, unit(4, 6)],
      ]),
      terrain: new Map(),
      mapByte0: 0,
    };
    const [, err] = attempt(() => engage(p, b, 0, 4));
    expect(
      errorsIs(err, ErrUndefined),
      `map-edge ranged battle: ${String(err)}, want ErrUndefined`,
    ).toBe(true);
  });

  // Battle distance: the recorded corpus holds adjacent battles that took the range penalty, and the profile
  // reproduces the transposed call.
  it("battle distance", () => {
    expect(
      p.transposedDistance,
      "profile does not reproduce the transposed battle distance",
    ).toBe(true);
    const [seen, count] = counter("adjacent battle at transposed distance 2");
    const layout = newProfile({ tables: p.tables });
    for (const r of battles) {
      if (
        r.e.returnFire &&
        distance(layout, r.b.width, r.a.cell, r.d.cell) === 1 &&
        distance(p, r.b.width, r.a.cell, r.d.cell) === 2
      ) {
        count("adjacent battle at transposed distance 2");
      }
    }
    require(seen);
  });
});
