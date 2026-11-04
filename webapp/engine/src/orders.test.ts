// Twin of internal/dosrules/orders_test.go.

import { describe, expect, it } from "vitest";
import {
  type OrdersFixture,
  ordersBattlefield,
  ordersList,
  ordersResult,
} from "./fixtures.ts";
import { at, attempt } from "./goruntime.ts";
import { resolveOrders } from "./orders.ts";
import { decodeOrdersFixture, load, replay } from "./testing/fixtures.ts";
import { Profiles, moveTables } from "./testing/tables.ts";

// TestOrdersFixtures replays the recorded attack lists (orders.json): each one resolves to the handler's
// recorded outcome, and the corpus shows every behaviour of the loop the spec names.
replay(
  "orders.json",
  decodeOrdersFixture,
  183,
  (p, f) => {
    const got = ordersResult(
      ...attempt(() => resolveOrders(p, ordersBattlefield(f), ordersList(f))),
    );
    expect(got, f.id).toStrictEqual(f.expected);
  },
  "pair",
  "supporter",
  "random",
);

describe.each(Profiles)("orders behaviours, %s build", (build) => {
  it("the corpus shows every behaviour of the loop", () => {
    const seen = new Map<string, number>([
      ["order skipped", 0],
      ["target destroyed before its order", 0],
      ["surviving target fought again", 0],
      ["attacker destroyed before a later order", 0],
      ["other side moved", 0],
      ["other side's list present", 0],
    ]);
    for (const f of load(build, "orders.json", decodeOrdersFixture)) {
      ordersBehaviour(f, seen);
    }
    for (const [name, n] of seen) {
      expect(n, `no recorded battle with ${name}`).toBeGreaterThan(0);
    }
  });
});

// ordersBehaviour counts what a recorded list shows.
function ordersBehaviour(f: OrdersFixture, seen: Map<string, number>): void {
  const count = (name: string): void => {
    seen.set(name, (seen.get(name) ?? 0) + 1);
  };
  const os = f.expected.orders ?? [];
  const orders = f.orders ?? [];
  for (const [i, o] of os.entries()) {
    if (!o.battle) {
      count("order skipped");
      if (orders.slice(0, i).some((q) => q[1] === at(orders, i)[1])) {
        count("target destroyed before its order");
      }
      continue;
    }
    const prev = os[i - 1];
    if (prev?.battle === true && prev.defender === o.defender) {
      count("surviving target fought again");
    }
    for (const k of o.destroyed ?? []) {
      if (k === at(orders, i)[0] && i + 1 < os.length) {
        count("attacker destroyed before a later order");
      }
    }
  }
  if ((f.seen ?? []).length > 0) {
    count("other side moved");
  }
  if ((f.idle ?? []).length > 0) {
    count("other side's list present");
  }
}

// TestTerrainClasses checks the combat class of every terrain record: one of the nine classes, and the same
// in the winter table, which differs from the summer one in the cost only.
describe.each(Profiles)("terrain classes, %s build", (build, p) => {
  it("every terrain record", () => {
    const m = moveTables(p);
    for (const [i, tr] of m.terrains.entries()) {
      expect(tr.class, `${build}: terrain ${i}`).toBeGreaterThanOrEqual(0);
      expect(tr.class, `${build}: terrain ${i}`).toBeLessThanOrEqual(8);
      if (m.winter !== null) {
        expect(at(m.winter, i).class, `${build}: terrain ${i} in winter`).toBe(
          tr.class,
        );
      }
    }
  });
});
