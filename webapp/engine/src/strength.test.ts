// Twin of internal/dosrules/strength_test.go: the shipped types use the primary strength; alternate strength
// belongs to the casualty routine's synthetic cases, not to a missing unit mechanic.

import { existsSync, readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import type { Fighter } from "./casualties.ts";
import { casualtyInputs } from "./fixtures.ts";
import { at } from "./goruntime.ts";
import { tablesOf } from "./profile.ts";
import { decodeCasualtyFixture, load } from "./testing/fixtures.ts";
import { repo } from "./testing/files.ts";
import { Profiles } from "./testing/tables.ts";
import { unpack } from "./testing/tpwm.ts";

describe("TestShippedStrengthData", () => {
  it.each(Profiles)("%s", (_build, p) => {
    for (const [i, typ] of tablesOf(p).types.entries()) {
      expect(
        typ.strengthFlags & 4,
        `type ${i} enables alternate strength`,
      ).toBe(0);
    }
  });
});

// Check the embedded strength data against the local original when present. Only numeric fields enter the
// assertions; original files stay local.
describe("TestStrengthDataMatchesOriginal", () => {
  it.for(Profiles)("%s", ([build, p], ctx) => {
    const dir =
      build === "demo" ? "original/demo/HL" : "original/extracted/dos-full";
    const path = `${repo}${dir}/UNIT.DAT`;
    if (!existsSync(path)) {
      ctx.skip("no local original unit table");
    }
    const b = unpack(new Uint8Array(readFileSync(path)));
    const recordLen = 69;
    const types = tablesOf(p).types;
    expect(b.length).toBe(recordLen * types.length);
    for (const [i, typ] of types.entries()) {
      const r = b.subarray(recordLen * i, recordLen * (i + 1));
      const group = at(r, 3);
      const flags = at(r, 15) | (at(r, 16) << 8);
      expect(
        [typ.group, typ.strengthFlags],
        `type ${i}: group/flags`,
      ).toStrictEqual([group, flags]);
      expect(at(r, 4), `type ${i}: alternate group`).toBe(0);
    }
  });
});

// These gates inspect recorded expectations, independently of the resolver, so the corpus must keep exercising
// alternate strength on both sides.
describe("TestAlternateStrengthEvidence", () => {
  it.each(Profiles)("%s", (build, p) => {
    const seen = new Map<string, number>(
      [
        "alternate attacker",
        "alternate defender",
        "both alternate",
        "attacker alternate loss",
        "defender alternate loss",
        "attacker primary cleared",
        "defender primary cleared",
        "attacker primary preserved",
        "defender primary preserved",
      ].map((k) => [k, 0]),
    );
    const count = (name: string): void => {
      seen.set(name, (seen.get(name) ?? 0) + 1);
    };
    const sizes = [
      Array.from({ length: 7 }, () => false),
      Array.from({ length: 7 }, () => false),
    ];
    for (const f of load(build, "casualties.json", decodeCasualtyFixture)) {
      if ((f.expected.error ?? "") !== "") {
        continue;
      }
      const [a, d] = casualtyInputs(f, tablesOf(p));
      if (
        (a.type.strengthFlags & 4) !== 0 &&
        (d.type.strengthFlags & 4) !== 0
      ) {
        count("both alternate");
      }
      const sides: [string, Fighter, number, number][] = [
        [
          "attacker",
          a,
          f.expected.attackerSizeAfter,
          f.expected.attackerAlternateAfter,
        ],
        [
          "defender",
          d,
          f.expected.defenderSizeAfter,
          f.expected.defenderAlternateAfter,
        ],
      ];
      for (const [side, [name, unit, size, alt]] of sides.entries()) {
        if ((unit.type.strengthFlags & 4) === 0) {
          expect(
            alt,
            `${f.id}: ordinary ${name} changed alternate strength`,
          ).toBe(unit.alt);
          continue;
        }
        count(`alternate ${name}`);
        if (unit.alt >= 0 && unit.alt <= 6) {
          at(sizes, side)[unit.alt] = true;
        }
        if (alt < unit.alt) {
          count(`${name} alternate loss`);
        }
        if (alt > 0) {
          expect(
            size,
            `${f.id}: surviving ${name} changed primary strength`,
          ).toBe(unit.size);
          count(`${name} primary preserved`);
        } else if (size === 0 && unit.size > 0) {
          count(`${name} primary cleared`);
        }
      }
    }
    for (const [side, values] of sizes.entries()) {
      for (const [size, present] of values.entries()) {
        expect(present, `no alternate strength ${size} on side ${side}`).toBe(
          true,
        );
      }
    }
    for (const name of [...seen.keys()].sort()) {
      expect(seen.get(name), `no recorded battle with ${name}`).toBeGreaterThan(
        0,
      );
    }
  });
});
