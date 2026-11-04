// Twin of internal/dosrules/ship_test.go.

import { describe, expect, it } from "vitest";
import { type Battlefield, type Unit, unitOf } from "./engage.ts";
import { type Order, resolveOrders } from "./orders.ts";
import { Profiles } from "./testing/tables.ts";

// A two-cell ship is one unit: battles against its rear leave the head the same strength, and the ship goes
// down whole.
describe.each(Profiles)("ship fights whole, %s build", (build, p) => {
  it("twelve attacks on the rear", () => {
    const ship = (type: number, cell: number, flags4: number): Unit => ({
      type,
      cell,
      size: 6,
      alt: 0,
      exp: 0,
      flags4,
      flags6: 0,
    });
    const b: Battlefield = {
      width: 10,
      height: 10,
      units: new Map([
        [0, ship(34, 22, 0xc9)],
        [1, ship(35, 12, 0x49)],
        [2, ship(32, 24, 0xca)],
        [3, ship(33, 23, 0x4a)],
      ]),
      terrain: new Map(),
      mapByte0: 0,
    };
    const orders: Order[] = Array.from({ length: 12 }, () => ({
      attacker: 0,
      target: 23,
    }));
    const r = resolveOrders(p, b, orders);
    const fought = r.orders.filter((o) => o.battle).length;
    const head = unitOf(r.units, 2);
    const rear = unitOf(r.units, 3);
    const what = `${build}: ${fought} battles leave head ${JSON.stringify(head)}, rear ${JSON.stringify(rear)}`;
    expect(fought, what).toBeGreaterThan(0);
    expect(head.size, what).toBe(rear.size);
    expect(head.exp, what).toBe(rear.exp);
    expect(head.size, what).not.toBe(6);
    const gone = r.gone.includes(3);
    expect(r.gone.includes(2), `${build}: gone ${r.gone.join(",")}`).toBe(gone);
    expect(rear.size === 0, `${build}: gone ${r.gone.join(",")}`).toBe(gone);
    expect(
      gone,
      `${build}: twelve attacks leave the destroyer at ${rear.size}`,
    ).toBe(true);
    // b is left as it was.
    expect(unitOf(b.units, 3).size).toBe(6);
  });
});
