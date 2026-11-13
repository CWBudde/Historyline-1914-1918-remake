import { expect, it } from "vitest";
import { Painter } from "../split/painter.ts";
import { battleLayout } from "./layout.ts";
import { paintBattle } from "./plate.ts";

/** painted paints the plate into a w × h buffer and reads its pixels. */
function painted(w: number, h: number) {
  const px = new Uint8ClampedArray(4 * w * h),
    l = battleLayout({ x: 0, y: 0, width: w, height: h });
  paintBattle(new Painter(px, w, h, l.scale, { steel: null, rivets: null }), l);
  const at = (x: number, y: number) => {
    const i = 4 * (Math.round(y) * w + Math.round(x));
    return [px[i] ?? 0, px[i + 1] ?? 0, px[i + 2] ?? 0, px[i + 3] ?? 0];
  };
  return { px, l, at };
}
const luma = ([r = 0, g = 0, b = 0]: number[]) =>
  0.299 * r + 0.587 * g + 0.114 * b;

it("covers every pixel with opaque steel, also beyond the scene's block", () => {
  // A host wider than 152:180: the steel still reaches its edges.
  const { px } = painted(200, 180);
  for (let i = 3; i < px.length; i += 4) expect(px[i]).toBe(255);
});
it("paints the hex as dark sepia ground, lit in its middle, in a steel rim", () => {
  const { l, at } = painted(304, 360);
  for (const s of l.sides) {
    const cx = s.hex.x + s.hex.width / 2,
      cy = s.hex.y + s.hex.height / 2,
      middle = at(cx, cy),
      inner = at(cx, s.hex.y + s.hex.height * 0.2),
      rim = at(cx, s.hex.y + 1);
    // Brown ground: red over green over blue; lighter in the middle than near the edge.
    expect(middle[0]).toBeGreaterThan(middle[1] ?? 0);
    expect(middle[1]).toBeGreaterThan(middle[2] ?? 0);
    expect(luma(middle)).toBeGreaterThan(luma(inner));
    // The rim is grey steel, not ground.
    expect(Math.abs((rim[0] ?? 0) - (rim[1] ?? 0))).toBeLessThan(8);
  }
});
it("sets the photo in a dark recess and brass rivets on the seam", () => {
  const { l, at } = painted(304, 360);
  const s = l.sides[0],
    well = at(s.photoWell.x + 1, s.photoWell.y + s.photoWell.height / 2),
    plate = at(s.half.x + s.half.width / 2, s.half.y + 4),
    rivet = l.rivets[0];
  expect(luma(well)).toBeLessThan(luma(plate) - 20);
  if (rivet === undefined) throw Error("no rivets");
  const head = at(rivet.x - 1, rivet.y - 1);
  // Brass: warm, red well over blue.
  expect((head[0] ?? 0) - (head[2] ?? 0)).toBeGreaterThan(40);
});
