// Twin of internal/dosrules/gate_test.go, the casualty clauses of TestGate (repeatable, fixed-point,
// experience terms, divide error). The engagement clauses (losses, experience, blocking, pincer, ranged
// support, battle distance) read engagements.json and Engage; they follow with the order handler.

import { describe, expect, it } from "vitest";
import { ErrDivideError, casualties, percent } from "./casualties.ts";
import {
  type CasualtyExpected,
  casualtyInputs,
  casualtyResult,
} from "./fixtures.ts";
import { attempt, errorsIs } from "./goruntime.ts";
import { tablesOf } from "./profile.ts";
import { decodeCasualtyFixture, load } from "./testing/fixtures.ts";
import { Profiles } from "./testing/tables.ts";

// blockCap is the blocking cap of support.go (80 per eligible unit, at most 150).
const blockCap = 150;

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
