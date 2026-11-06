import { expect, it } from "vitest";
import { routeKey } from "./keys.ts";

it("routes native key sets A and B independently in simultaneous play", () => {
  for (const key of [
    "ArrowUp",
    "ArrowDown",
    "ArrowLeft",
    "ArrowRight",
    " ",
    "Enter",
  ])
    expect(routeKey("simultaneous", 0, key, "")).toEqual({ side: 0, key });
  for (const [key, mapped] of [
    ["d", "ArrowUp"],
    ["c", "ArrowDown"],
    ["x", "ArrowLeft"],
    ["v", "ArrowRight"],
    ["D", "ArrowUp"],
    ["Control", " "],
  ] as const)
    expect(
      routeKey("simultaneous", 0, key, key === "Control" ? "ControlLeft" : ""),
    ).toEqual({ side: 1, key: mapped });
  expect(routeKey("simultaneous", 0, "Control", "ControlRight")).toBeNull();
});
it("gives the active hot-seat recipient the complete first key set", () => {
  for (const side of [0, 1] as const) {
    expect(routeKey("hotseat", side, "ArrowRight", "ArrowRight")).toEqual({
      side,
      key: "ArrowRight",
    });
    expect(routeKey("hotseat", side, "F1", "F1")).toEqual({ side, key: "F1" });
    expect(routeKey("hotseat", side, "d", "KeyD")).toBeNull();
  }
});
it("routes single-player shortcuts to its human side and excludes browser shortcuts", () => {
  for (const key of [
    "Escape",
    "F1",
    "F2",
    "F8",
    "F9",
    "F11",
    "q",
    "m",
    "Home",
    "End",
    "PageUp",
    "PageDown",
  ])
    expect(routeKey("single", 1, key, "")).toEqual({ side: 1, key });
  for (const key of ["Tab", "F5", "r", "Meta", "Control"])
    expect(routeKey("single", 1, key, "")).toBeNull();
});
