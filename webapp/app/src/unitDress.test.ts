import { createHash } from "node:crypto";
import { expect, it } from "vitest";
import {
  dress,
  dressMargin,
  padded,
  rimFor,
  type RGBARaster,
  type Rim,
} from "./unitDress.ts";

// Hashes of straight RGBA bytes from native unitsprite.Dress on synthetic().
// Cover fractional alpha, disc dilation, blur, both rims and the flying shift.
const native = [
  [
    1,
    false,
    "dark",
    "8f98322dc4851f7dacfd6cc204e421f4e030abbfe142d9094eca6bc35d354355",
  ],
  [
    1,
    false,
    "light",
    "06dfcc0eee3fe15a7a151b6a740441e961291053b53aadd5bf6f6f53a4e48024",
  ],
  [
    1,
    true,
    "dark",
    "8900a107e9a070e7b1038ac95f37ee955459aadeaa15603bfc7f23c0eb131e21",
  ],
  [
    1,
    true,
    "light",
    "47c5609d83df8dc1eaf78e463f5e433bed2aa14687c32cb96ca5205772b77fb9",
  ],
  [
    2,
    false,
    "dark",
    "4699034358f302ba8518b2b6e7acc0feaa5c64f2a5c53c6cdc181ff452fcb41d",
  ],
  [
    2,
    false,
    "light",
    "347dbdc3a99a508dc17ab6d1ff044248f98367434d4aedd7a606f7bf5f7e507a",
  ],
  [
    2,
    true,
    "dark",
    "76233ce715d80f4b6b74e1d527f535775b6c3704fdb29dcd0d611c3e0126a7a1",
  ],
  [
    2,
    true,
    "light",
    "8d13ff022807e4e3e7781baa2fe4191324f87a665361f76718bdaa56a60d3f91",
  ],
  [
    4,
    false,
    "dark",
    "a9160c01dc6b93bcb1d3231067d1733ec36a0312901a19e60a07730c110e3138",
  ],
  [
    4,
    false,
    "light",
    "6dd4df7ad30951a7c660142faafd7b3470881201711318d58447c3434d697a77",
  ],
  [
    4,
    true,
    "dark",
    "b1e1a3ca77963eb2a467a4dd2221c5750031ad15bba72e116a94bc79198bda60",
  ],
  [
    4,
    true,
    "light",
    "b6a9ae40f4f7938159728fef3c50010105964a15f613bd2bbdd92ee0925bb6fb",
  ],
] as const;
function synthetic(d: number): RGBARaster {
  const width = 5 * d,
    height = 4 * d,
    data = new Uint8ClampedArray(width * height * 4);
  for (let y = 0; y < height; y++)
    for (let x = 0; x < width; x++) {
      const a =
        (x + 2 * y) % 7 === 0
          ? 0
          : (x + y) % 4 === 0
            ? 81
            : x >= d && y >= d
              ? 255
              : 145;
      data.set(
        [20 + ((x * 7) % 200), 40 + ((y * 11) % 180), 170 - ((x + y) % 90), a],
        (y * width + x) * 4,
      );
    }
  return { width, height, data };
}
for (const [d, flying, rim, sha] of native) {
  it(`matches native RGBA at ${d}x, flying ${String(flying)}, rim ${rim}`, () => {
    const src = synthetic(d),
      before = src.data.slice(),
      out = dress(src, d, flying, rim);
    expect([out.width, out.height]).toEqual([21 * d, 20 * d]);
    expect(dressMargin(d)).toBe(8 * d);
    expect(createHash("sha256").update(out.data).digest("hex")).toBe(sha);
    expect(src.data).toEqual(before);
  });
}
function pixel(src: RGBARaster, x: number, y: number): number[] {
  return Array.from(
    src.data.slice((y * src.width + x) * 4, (y * src.width + x) * 4 + 4),
  );
}
it("keeps opaque body pixels and draws a disc rather than square rim", () => {
  const src = {
    width: 1,
    height: 1,
    data: new Uint8ClampedArray([245, 40, 20, 255]),
  };
  for (const [rim, colour] of [
    ["dark", [16, 18, 12, 255]],
    ["light", [196, 190, 162, 255]],
  ] as const) {
    const out = dress(src, 1, false, rim);
    expect(pixel(out, 8, 8)).toEqual([245, 40, 20, 255]);
    expect(pixel(out, 6, 8)).toEqual(colour);
    expect(pixel(out, 6, 6)[3]).toBeLessThan(128);
  }
});
it("keeps empty sprites empty and preserves their hidden RGB", () => {
  const out = dress(
    { width: 1, height: 1, data: new Uint8ClampedArray([123, 25, 50, 0]) },
    2,
    true,
    "light",
  );
  expect(pixel(out, 16, 16)).toEqual([123, 25, 50, 0]);
  for (let i = 3; i < out.data.length; i += 4) expect(out.data[i]).toBe(0);
});
const terrains: readonly [string, boolean, Rim][] = [
  ["forest", false, "light"],
  ["forest_w", false, "light"],
  ["water", false, "light"],
  ["waterDeep", false, "light"],
  ["waterSea", false, "light"],
  ["waterStones", false, "light"],
  ["forest", true, "dark"],
  ["water", true, "dark"],
  ["waterSea", true, "dark"],
  ["grass", false, "dark"],
  ["grass_w", false, "dark"],
  ["lakeside12", false, "dark"],
];
it.each(terrains)(
  "selects native rim on %s with infrastructure %s",
  (terrain, infra, rim) => {
    expect(rimFor(terrain, infra)).toBe(rim);
  },
);

// Native cmd/historyline padded() bytes converted from premultiplied RGBA to
// straight RGBA through Go's color.NRGBAModel, ready for browser ImageData.
const nativePadding = [
  [
    0,
    false,
    "7c914e233404a4231242ed9fd906d08183944f1bdef424b201e54200adcc7562",
  ],
  [0, true, "7c914e233404a4231242ed9fd906d08183944f1bdef424b201e54200adcc7562"],
  [
    1,
    false,
    "48ff67ee22b9abc792ce09ec6db26d3c5bce59acf1b4ee539f2486221f0626db",
  ],
  [1, true, "78ff88a14030f4ec5d3bc075f616169066a85ba325cbbff760a9ae45ca878db1"],
  [
    3,
    false,
    "66df5f27f8df2f8fb4e41ef056ae868750e651f45529fe238738d99d56357c53",
  ],
  [3, true, "dcda99f0a82cb4d93f7c6d3432bb6cfff6126bc96a5d88be70092c108fb92c7d"],
] as const;
for (const [pad, bleed, hash] of nativePadding) {
  it(`matches native padding with ${pad} texels and bleed ${String(bleed)}`, () => {
    const src = synthetic(4),
      before = src.data.slice(),
      out = padded(src, pad, bleed);
    expect([out.width, out.height]).toEqual([20 + 2 * pad, 16 + 2 * pad]);
    expect(createHash("sha256").update(out.data).digest("hex")).toBe(hash);
    expect(src.data).toEqual(before);
  });
}
it("grows one opaque texel per step and keeps native neighbour colour priority", () => {
  const src = { width: 7, height: 1, data: new Uint8ClampedArray(28) };
  src.data.set([255, 20, 10, 255], 3 * 4);
  const out = padded(src, 1, true);
  expect(pixel(out, 3, 0)).toEqual([255, 20, 10, 255]);
  expect(pixel(out, 6, 1)).toEqual([0, 0, 0, 0]);
  const choices = {
    width: 3,
    height: 1,
    data: new Uint8ClampedArray([
      240, 10, 10, 255, 0, 0, 0, 0, 10, 20, 240, 255,
    ]),
  };
  expect(pixel(padded(choices, 1, true), 2, 1)).toEqual([10, 20, 240, 255]);
});
