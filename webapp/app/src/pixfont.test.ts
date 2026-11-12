import { expect, it } from "vitest";
import { PixelFace, parseMetrics } from "./pixfont.ts";

/** A face like the shipped Small one: capitals and digits 3 wide, the dot 1, the box 4. */
function face(): PixelFace {
  const glyphs: Record<
    string,
    { x: number; width: number; height: number; advance: number }
  > = {};
  let x = 5;
  for (const r of "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 .:-") {
    const width = r === "." || r === ":" ? 1 : 3;
    glyphs[r] = { x, width, height: 5, advance: width + 1 };
    x += width + 1;
  }
  return new PixelFace(
    parseMetrics({
      width: x - 1,
      height: 5,
      glyphs,
      fallback: { x: 0, width: 4, height: 5, advance: 5 },
      plain: { Ä: "A", Ö: "O", Ü: "U" },
      expand: { ß: "SS", "…": "..." },
    }),
  );
}

it("measures up to the last glyph's last column at integer scales", () => {
  const f = face();
  expect(f.measure("AB", 1)).toEqual({ width: 7, height: 5 });
  expect(f.measure("AB", 2)).toEqual({ width: 14, height: 10 });
  expect(f.measure("A.", 3)).toEqual({ width: 15, height: 15 });
  expect(f.measure("", 2).width).toBe(0);
  // Scales below 1 count as 1, fractions are dropped.
  expect(f.measure("A", 0.5).width).toBe(3);
  expect(f.measure("A", 2.7).width).toBe(6);
});

it("draws lower case as capitals, umlauts plain, ß spelled out and unknown runes as the box", () => {
  const f = face();
  expect(f.layout("ab").width).toBe(f.layout("AB").width);
  expect(f.layout("Ö").glyphs[0]?.glyph).toEqual(
    f.layout("O").glyphs[0]?.glyph,
  );
  expect(f.layout("ß").glyphs).toHaveLength(2);
  expect(f.layout("€").glyphs[0]?.glyph.width).toBe(4);
  expect(f.layout("A1").glyphs.map((g) => g.digit)).toEqual([false, true]);
});

it("fits text into a width with dots, never wider, as long as it can", () => {
  const f = face(),
    scale = 2,
    s = "Waiting for movement-mode",
    full = f.measure(s, scale).width,
    dots = f.measure("...", scale).width;
  expect(f.fit(s, scale, full)).toBe(s);
  for (let max = full - 1; max >= 0; max -= 7) {
    const got = f.fit(s, scale, max);
    expect(f.measure(got, scale).width).toBeLessThanOrEqual(max);
    if (max < dots) {
      expect(got).toBe("");
      continue;
    }
    expect(got.endsWith("...")).toBe(true);
    const head = got.slice(0, -3);
    expect(s.startsWith(head)).toBe(true);
    expect(head.endsWith(" ")).toBe(false);
    const longer = s.slice(0, head.length + 1).trimEnd() + "...";
    if (longer.length > got.length)
      expect(f.measure(longer, scale).width).toBeGreaterThan(max);
  }
});
