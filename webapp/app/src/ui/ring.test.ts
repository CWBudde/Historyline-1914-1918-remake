import { expect, it } from "vitest";
import { placement, ringActions } from "./ring.ts";

it("keeps the desktop's key gestures: clockwise map and contents actions with disabled actions omitted", () => {
  expect(ringActions(placement(false, false), () => true)).toEqual([
    "move",
    "map",
    "info",
    "mode",
  ]);
  expect(ringActions(placement(true, true), () => true)).toEqual([
    "attack",
    "map",
    "info",
    "open",
  ]);
  expect(
    ringActions(placement(false, true, true), (a) => a !== "repair"),
  ).toEqual(["exit", "produce"]);
});
