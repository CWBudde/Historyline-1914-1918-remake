import { describe, expect, it } from "vitest";
import {
  brass,
  gradeTile,
  ramp,
  rivetHeads,
  rivetImage,
  rivetCentre,
  rivetRadius,
  rivetSize,
  seamTile,
  steelDark,
  steelLight,
  type Pixels,
} from "./steel.ts";

function pixels(
  width: number,
  height: number,
  at: (x: number, y: number) => [number, number, number, number],
): Pixels {
  const data = new Uint8ClampedArray(4 * width * height);
  for (let y = 0; y < height; y++)
    for (let x = 0; x < width; x++) data.set(at(x, y), 4 * (y * width + x));
  return { width, height, data };
}
function rgba(p: Pixels, x: number, y: number): number[] {
  const i = 4 * (y * p.width + x);
  return Array.from(p.data.slice(i, i + 4));
}
/** CSS overlay of source s over a backdrop b at most half bright: 2·s·b. */
function overlay(b: number, s: number): number {
  return (2 * s * b) / 255;
}

describe("menu steel", () => {
  it("ramps from the desktop game's dark to light steel", () => {
    expect(ramp(0)).toEqual(steelDark);
    expect(ramp(1)).toEqual(steelLight);
    expect(ramp(0.3)).toEqual([79, 79, 69]);
  });

  it("grades a tile so overlay over a colour modulates it like texture.Modulate", () => {
    // Two texels: one at the mean's half, one at one and a half times it.
    const tile = pixels(2, 1, (x) =>
      x === 0 ? [50, 50, 50, 255] : [150, 150, 150, 255],
    );
    const graded = gradeTile(tile, 0.6);
    const base = ramp(0.3)[0];
    // c·(1 + 0.6·(−0.5)) and c·(1 + 0.6·0.5)
    expect(overlay(base, graded.data[0] ?? 0)).toBeCloseTo(base * 0.7, 0);
    expect(overlay(base, graded.data[4] ?? 0)).toBeCloseTo(base * 1.3, 0);
    expect(rgba(graded, 1, 0)[3]).toBe(255);
  });

  it("leaves a flat tile neutral", () => {
    const graded = gradeTile(
      pixels(4, 4, () => [90, 100, 80, 255]),
      0.9,
    );
    expect(new Set(graded.data.filter((_, i) => i % 4 !== 3))).toEqual(
      new Set([128]),
    );
  });
});

describe("rivets", () => {
  // A sheet of 4×2 cells of 16 texels with a round head of radius 5 in each.
  const sheet = pixels(64, 32, (x, y) => {
    const cx = (x % 16) - 7.5,
      cy = (y % 16) - 7.5;
    return Math.hypot(cx, cy) < 5 ? [200, 150, 60, 255] : [0, 0, 0, 0];
  });

  it("finds each cell's head and the heads' mean colour", () => {
    const { heads, mean } = rivetHeads(sheet);
    expect(heads).toHaveLength(8);
    expect(heads[5]?.cx).toBeCloseTo(16 + 8, 5);
    expect(heads[5]?.cy).toBeCloseTo(16 + 8, 5);
    expect(heads[0]?.r).toBeGreaterThan(4.5);
    expect(heads[0]?.r).toBeLessThan(5.5);
    expect(mean).toEqual([200, 150, 60]);
  });

  it("cuts one head out round, as light as brass, with its shadow", () => {
    const image = rivetImage(sheet);
    expect(image.width).toBe(rivetSize);
    const centre = rgba(image, rivetCentre, rivetCentre),
      luma = (c: readonly number[]) =>
        0.299 * (c[0] ?? 0) + 0.587 * (c[1] ?? 0) + 0.114 * (c[2] ?? 0);
    expect(centre[3]).toBe(255);
    expect(luma(centre)).toBeCloseTo(luma(brass), -1);
    // Clear beyond the head to the top left, shaded to the bottom right.
    expect(rgba(image, rivetCentre - rivetRadius - 1, rivetCentre)[3]).toBe(0);
    const shadow = rgba(
      image,
      Math.round(rivetCentre + rivetRadius * 1.1),
      Math.round(rivetCentre + rivetRadius * 0.4),
    );
    expect(shadow[3]).toBeGreaterThan(0);
    expect(shadow[3]).toBeLessThan(255);
  });

  it("draws a domed brass head without a sheet", () => {
    const image = rivetImage(null),
      lit = rgba(image, rivetCentre - 5, rivetCentre - 5),
      dark = rgba(image, rivetCentre + 8, rivetCentre + 8);
    expect(lit[0] ?? 0).toBeGreaterThan(dark[0] ?? 0);
  });
});

describe("hull plating seams", () => {
  it("lines the tile's top and left edge and rivets the crossing in every corner", () => {
    const tile = seamTile(null, 20, 2);
    expect(tile.width).toBe(40);
    // The dark line, then the lit one, along the left edge.
    expect(rgba(tile, 0, 20).slice(0, 3)).toEqual([...ramp(0.1)]);
    expect(rgba(tile, 5, 20).slice(0, 3)).toEqual([...ramp(0.42)]);
    // Clear plate between the seams.
    expect(rgba(tile, 20, 20)[3]).toBe(0);
    // The rivet's head reaches into the far corner too, so the tile wraps.
    expect(rgba(tile, 39, 39)[3]).toBeGreaterThan(0);
  });
});
