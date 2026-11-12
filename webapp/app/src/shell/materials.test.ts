import { expect, it } from "vitest";
import { embrasureAspect, embrasureHole, shellUnit } from "./materials.ts";

it("fits a 712×400 stage in landscape and a 340×560 column upright", () => {
  expect(shellUnit(1920, 1080, 1)).toBeCloseTo(1920 / 712, 3);
  // An ultrawide window is held by its height, a 16:10 one by its width.
  expect(shellUnit(2560, 1080, 1)).toBeCloseTo(2.7, 3);
  expect(shellUnit(1920, 1200, 1)).toBeCloseTo(1920 / 712, 3);
  expect(shellUnit(390, 844, 1)).toBeCloseTo(390 / 340, 3);
  expect(shellUnit(320, 200, 1.5)).toBeCloseTo(1.65, 3);
});

it("keeps the embrasure's opening inside its frame", () => {
  const { left, right, top, bottom } = embrasureHole;
  expect(left + right).toBeLessThan(0.2);
  expect(top + bottom).toBeLessThan(0.5);
  // The opening is wider than the menu painting's 16:10, so the painting covers it by its width.
  const hole = (embrasureAspect * (1 - left - right)) / (1 - top - bottom);
  expect(hole).toBeGreaterThan(1.6);
});
