// Twin of internal/hexgrid/pixel_test.go.

import { describe, expect, it } from "vitest";
import { array, int, object } from "./decode.ts";
import {
  type Coord,
  type Grid,
  around,
  bounds,
  center,
  centerY,
  colPitch,
  contains,
  corners,
  linePitch,
  pixelToHex,
} from "./hexgrid.ts";
import { readJSON } from "./testing/files.ts";

const show = (c: Coord): string => `(${c.line},${c.col})`;

describe("hexgrid pixels", () => {
  it("Center round trip", () => {
    const g: Grid = { width: 8, lines: 23 };
    for (let line = 0; line < g.lines; line++) {
      for (let col = 0; col < g.width; col++) {
        const c: Coord = { line, col };
        if (!contains(g, c)) {
          continue;
        }
        const [x, y] = center(c);
        for (const [dx, dy] of [
          [0, 0],
          [-20, 0],
          [20, 0],
          [0, -30],
          [0, 30],
          [10, 20],
        ] as const) {
          const [got, ok] = pixelToHex(g, x + dx, y + dy);
          expect([got, ok], `hex ${show(c)} offset ${dx},${dy}`).toStrictEqual([
            c,
            true,
          ]);
        }
      }
    }
  });

  it("PixelToHex outside", () => {
    const g: Grid = { width: 8, lines: 23 };
    for (const [x, y] of [
      [-5, 40],
      [-100, -100],
      [5000, 40],
      [48, 23 * 40 + 200],
      [48, -60],
    ] as const) {
      const [c, ok] = pixelToHex(g, x, y);
      expect(ok, `point ${x},${y} resolved to ${show(c)}`).toBe(false);
    }
    // Odd lines are one hex shorter: the slot right of the last odd hex is empty.
    let [x, y] = center({ line: 1, col: 8 });
    expect(pixelToHex(g, x, y)[1]).toBe(false);
    // A hex of an odd line reaches half a hex left of column 0 of the even lines.
    [x, y] = center({ line: 1, col: 0 });
    expect(pixelToHex(g, x, y)).toStrictEqual([{ line: 1, col: 0 }, true]);
  });

  // Compares with findHex of main.py. Python's diagonal edges are only approximate (its intercept is 23, not
  // the 40 of the hex outline), so disagreement is tolerated only where the point is within 12 pixels of a
  // cell edge.
  it("PixelToHex matches Python on every sample of pixel.json", () => {
    const raw = object(
      readJSON("internal/hexgrid/testdata/pixel.json"),
      "pixel",
    );
    const g: Grid = {
      width: int(raw["width"], "width"),
      lines: int(raw["lines"], "lines"),
    };
    const samples = array(raw["samples"], "samples").map((s) => {
      const p = array(s, "sample").map((v) => int(v, "sample"));
      expect(p).toHaveLength(4);
      return p as [number, number, number, number];
    });
    expect(samples).toHaveLength(17303);
    let same = 0;
    let edge = 0;
    let skipped = 0;
    for (const [sx, sy, line, col] of samples) {
      // Right of the last column's flat top Python rejects even-line hexes and accepts the missing last hex
      // of odd lines; its edge handling there is simply wrong.
      if (sx >= (g.width - 1) * colPitch + 74) {
        skipped++;
        continue;
      }
      const [got, ok] = pixelToHex(g, sx, sy);
      const want: Coord = { line, col };
      const pyOK = line >= 0;
      if (
        ok === pyOK &&
        (!ok || (got.line === want.line && got.col === want.col))
      ) {
        same++;
        continue;
      }
      const m = edgeMargin(g, sx, sy);
      if (m > 12) {
        expect.fail(
          `(${sx},${sy}): got ${show(got)}/${ok}, Python ${show(want)}/${pyOK}, ${m.toFixed(1)} px from an edge`,
        );
      }
      edge++;
    }
    expect(same + edge + skipped).toBe(samples.length);
    expect(same).toBeGreaterThan(0);
  });

  it("Corners and Bounds", () => {
    const g: Grid = { width: 8, lines: 23 };
    let maxX = 0;
    let maxY = 0;
    for (let line = 0; line < g.lines; line++) {
      for (let col = 0; col < g.width; col++) {
        const c: Coord = { line, col };
        if (!contains(g, c)) {
          continue;
        }
        for (const [px, py] of corners(c)) {
          maxX = Math.max(maxX, px);
          maxY = Math.max(maxY, py);
        }
      }
    }
    expect(bounds(g)).toStrictEqual([maxX, maxY]);
    // Hex (0, 0) touches the origin: left vertex at x 0, top edge at y 0.
    const cs = corners({ line: 0, col: 0 });
    expect([cs[0], cs[1], cs[3]]).toStrictEqual([
      [0, 40],
      [25, 0],
      [96, 40],
    ]);
    // Neighbours tile without gaps: the top-right corner of a hex is the left vertex of its NE neighbour,
    // its right vertex the neighbour's bottom-left corner.
    const c: Coord = { line: 4, col: 3 };
    const ne = corners(around(c)[0] ?? c);
    const own = corners(c);
    expect([own[2], own[3]]).toStrictEqual([ne[0], ne[5]]);
    expect(bounds({ width: 0, lines: 0 })).toStrictEqual([0, 0]);
  });
});

// edgeMargin is half the difference between the distances to the nearest and the second nearest hex centre,
// which is the distance to the cell edge.
function edgeMargin(g: Grid, x: number, y: number): number {
  let best = Infinity;
  let second = Infinity;
  const line0 = Math.floor((y - centerY) / linePitch);
  for (let line = line0 - 2; line <= line0 + 3; line++) {
    for (let col = -2; col < g.width + 2; col++) {
      const [cx, cy] = center({ line, col });
      const d = Math.hypot(x - cx, y - cy);
      if (d < best) {
        second = best;
        best = d;
      } else if (d < second) {
        second = d;
      }
    }
  }
  return (second - best) / 2;
}
