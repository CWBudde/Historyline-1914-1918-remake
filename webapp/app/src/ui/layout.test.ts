import { expect, it } from "vitest";
import { layoutNames, layoutValue } from "./layout.tsx";

it("offers every layout and reads only those back", () => {
  expect(layoutNames.map(([value]) => value)).toEqual([
    "auto",
    "single",
    "original",
  ]);
  expect(layoutValue("single")).toBe("single");
  expect(layoutValue("responsive")).toBeNull();
});
