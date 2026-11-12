import { expect, it } from "vitest";
import { level, play } from "@historyline/engine";
import { Full } from "../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../engine/src/testing/play.ts";
import { MapBattle } from "./mapBattle.ts";
import { aimReach } from "./aim.ts";

it("keeps the aimed order's destinations and the unit's own cell in colour, nothing while none is aimed", () => {
  const s = play.newSession(shippedLevel("1914-08"), playData, {
      profile: Full,
      computer: [false, true],
    }),
    b = new MapBattle(s, 0);
  expect(aimReach(b, b.view())).toBeNull();
  const unit = [...b.view().units.values()].find((u) => {
    if (u.side !== 0 || u.head !== null) return false;
    try {
      return play.destinations(s, 0, u.index).size > 0;
    } catch {
      return false;
    }
  });
  if (unit === undefined) throw new Error("no movable unit");
  b.click(unit.at);
  expect(b.selected).toBe(unit.index);
  const reach = aimReach(b, b.view());
  expect(reach).not.toBeNull();
  expect(reach?.has(level.coordKey(unit.at))).toBe(true);
  for (const key of b.destinations()) expect(reach?.has(key)).toBe(true);
  // Only the destinations, the unit's cell and buildings' other cells: far from all of the map.
  expect(reach?.size).toBeLessThan(b.destinations().size + 20);
  b.cancel();
  expect(aimReach(b, b.view())).toBeNull();
});
