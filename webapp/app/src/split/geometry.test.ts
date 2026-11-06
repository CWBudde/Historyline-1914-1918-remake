import { expect, it } from "vitest";
import { hexgrid } from "@historyline/engine";
import {
  blockSize,
  defaultZoom,
  fit,
  layout,
  outline,
  rivets,
  View,
  zooms,
} from "./geometry.ts";

it.each([
  [1280, 720, 1, 632, 680, 8, 8, 1],
  [1920, 1080, 2, 944, 1000, 8, 8, 1000 / 680],
  [3840, 2160, 2, 1904, 2080, 8, 8, 2080 / 680],
  [3440, 1440, 1, 1712, 1400, 11, 8, 1400 / 680],
  [1024, 768, 1, 504, 728, 7, 8, 504 / 522],
  // A phone falls below native MinWindow: preserve the scale floor and fewer columns.
  [390, 844, 1, 187, 804, 4, 19, 0.5],
])(
  "fits native %dx%d at UI %s",
  (width, height, ui, pw, ph, cols, rows, scale) => {
    const s = layout(width, height, ui);
    expect(s.panels).toEqual([
      { x: 0, y: 0, width: pw, height: ph },
      { x: width - pw, y: 0, width: pw, height: ph },
    ]);
    expect([s.cols, s.rows]).toEqual([cols, rows]);
    expect(s.cellScale).toBeCloseTo(scale, 12);
    expect(s.fitScale).toBe(s.cellScale);
    expect(s.zoom).toBe(defaultZoom);
    expect(s.zoomed(defaultZoom)).toEqual(s);
  },
);

it("tiles odd window widths and native status wells without rounding gaps", () => {
  const s = layout(1281, 720);
  expect(s.divider).toEqual({ x: 632, y: 0, width: 17, height: 688 });
  expect(s.status).toEqual([
    { x: 0, y: 688, width: 632, height: 28 },
    { x: 649, y: 688, width: 632, height: 28 },
  ]);
  expect(s.bolt).toEqual({ x: 632, y: 694, width: 16, height: 16 });
  expect(layout(1920, 1080, 2).bolt).toEqual({
    x: 944,
    y: 1028,
    width: 32,
    height: 32,
  });
});

it.each([
  [0, 0],
  [1, 1],
  [50, 30],
  [-5, 10],
])("keeps native positive cell fits for undersized %dx%d", (w, h) => {
  const s = layout(w, h, 0);
  expect(s.width).toBe(Math.max(w, 1));
  expect(s.height).toBe(Math.max(h, 1));
  expect(s.ui).toBe(1);
  expect(s.cellScale).toBe(0.5);
  expect(s.cols).toBeGreaterThanOrEqual(1);
  expect(s.rows).toBeGreaterThanOrEqual(1);
  expect(s.panels[0].width).toBeGreaterThanOrEqual(1);
  expect(s.panels[0].height).toBeGreaterThanOrEqual(1);
});

it("fits exact blocks and the unshifted one-column special case", () => {
  expect(blockSize(7, 8)).toEqual({ width: 522, height: 680 });
  expect(blockSize(1, 3)).toEqual({ width: 96, height: 240 });
  expect(fit({ x: 37, y: 11, width: 261, height: 340 }, 0.5)).toEqual({
    cols: 7,
    rows: 8,
  });
  expect(fit({ x: 0, y: 0, width: 96, height: 240 }, 1)).toEqual({
    cols: 1,
    rows: 3,
  });
  expect(fit({ x: 0, y: 0, width: 1, height: 1 }, 2)).toEqual({
    cols: 1,
    rows: 1,
  });
});

it("scrolls by whole pairs and rows and enumerates both line parities", () => {
  const v = new View({ x: 0, y: 0, width: 632, height: 680 }, 680 / 600, 3, 2);
  expect([v.cols, v.rows, v.cells().length]).toEqual([7, 7, 49]);
  expect(v.origin()).toEqual({ x: 426, y: 160 });
  expect(v.offset()).toEqual({ x: 20, y: 0 });
  expect(v.cell(0, 0)).toEqual({ line: 4, col: 3 });
  expect(v.cell(1, 0)).toEqual({ line: 5, col: 3 });
  expect(v.cell(2, 0)).toEqual({ line: 4, col: 4 });
  expect(v.cell(0, 1)).toEqual({ line: 6, col: 3 });
  expect(v.cell(6, 6)).toEqual({ line: 16, col: 6 });
  expect(v.cell(5, 6)).toEqual({ line: 17, col: 5 });
  expect(v.cells().slice(0, 8)).toEqual([
    { line: 4, col: 3 },
    { line: 6, col: 3 },
    { line: 8, col: 3 },
    { line: 10, col: 3 },
    { line: 12, col: 3 },
    { line: 14, col: 3 },
    { line: 16, col: 3 },
    { line: 5, col: 3 },
  ]);
  const g = { width: 40, lines: 40 };
  expect(v.shows(g, { line: 17, col: 6 })).toBe(false);
  expect(v.shows(g, { line: 16, col: 6 })).toBe(true);
  expect(v.shows(g, { line: 3, col: 3 })).toBe(false);
  expect(v.shows(g, { line: 4, col: 2 })).toBe(false);
});

it("uses native rounding away from zero for a block larger than its panel", () => {
  const v = new View({ x: 0, y: 0, width: 1, height: 1 }, 1, 0, 0);
  expect(v.offset()).toEqual({ x: -48, y: -40 });
});

it("keeps native value semantics when callers change derived views and screens", () => {
  const s = layout(1280, 720);
  const z = s.zoomed(8);
  z.panels[0].x = 300;
  z.divider.x = 301;
  z.status[1].x = 302;
  z.bolt.x = 303;
  expect(s.panels[0].x).toBe(0);
  expect(s.divider.x).toBe(632);
  expect(s.status[1].x).toBe(648);
  expect(s.bolt.x).toBe(632);
  const v = s.view(0, 0, 0);
  v.panel.x = 100;
  expect(s.panels[0].x).toBe(0);
  const fitted = s.slimFit(v, 0, { x: 0, y: 0, width: 1000, height: 1000 });
  fitted.panel.x = 200;
  expect(v.panel.x).toBe(100);
});

it("round trips shifted views in either panel at every usable zoom", () => {
  for (const [w, h, ui] of [
    [1280, 720, 1],
    [390, 844, 1],
    [1920, 1080, 2],
  ] as const) {
    const s = layout(w, h, ui);
    for (const z of s.zoomLevels()) {
      for (const side of [0, 1] as const) {
        const v = s.zoomed(z).view(side, -2, 3);
        v.shiftX = 0.25;
        v.shiftY = -1.5;
        expect(v.origin()).toEqual({ x: -283.75, y: 238.5 });
        for (const [x, y] of [
          [0, 0],
          [431.5, 217.25],
          [1000, 900],
        ] as const) {
          const p = v.worldToScreen(x, y);
          const q = v.screenToWorld(p.x, p.y);
          expect(q.x).toBeCloseTo(x, 9);
          expect(q.y).toBeCloseTo(y, 9);
        }
      }
    }
  }
});

it("closes off-board cells and does not outline shared visible cell edges", () => {
  const g = { width: 1, lines: 1 };
  const v = new View({ x: 0, y: 0, width: 238, height: 200 }, 1, -1, -1);
  expect(v.cells().length).toBe(6);
  expect(v.cells().filter((c) => v.shows(g, c))).toEqual([{ line: 0, col: 0 }]);
  expect(v.shows({ width: 2, lines: 2 }, { line: 1, col: 1 })).toBe(false);
  expect(outline(v, g).map(({ a, b }) => [a, b])).toEqual([
    [
      { x: 142, y: 120 },
      { x: 167, y: 80 },
    ],
    [
      { x: 167, y: 80 },
      { x: 213, y: 80 },
    ],
    [
      { x: 213, y: 80 },
      { x: 238, y: 120 },
    ],
    [
      { x: 238, y: 120 },
      { x: 213, y: 160 },
    ],
    [
      { x: 213, y: 160 },
      { x: 167, y: 160 },
    ],
    [
      { x: 167, y: 160 },
      { x: 142, y: 120 },
    ],
  ]);
  const two = new View({ x: 0, y: 0, width: 167, height: 120 }, 1, 0, 0);
  expect(outline(two, { width: 2, lines: 2 })).toHaveLength(10);
  expect(rivets(two, { width: 2, lines: 2 })).toHaveLength(10);
  expect(outline(two, { width: 0, lines: 0 })).toEqual([]);
  expect(rivets(two, { width: 0, lines: 0 })).toEqual([]);
});

it("points unit outline normals into cells and sets rivets back into steel", () => {
  for (const scale of [0.5, 1, 1.7333]) {
    const v = new View(
      { x: 37, y: 11, width: 96 * scale, height: 80 * scale },
      scale,
      0,
      0,
    );
    const g = { width: 1, lines: 1 };
    const centre = v.worldToScreen(48, 40);
    for (const { a, b, n } of outline(v, g)) {
      expect(Math.hypot(n.x, n.y)).toBeCloseTo(1, 12);
      expect((centre.x - a.x) * n.x + (centre.y - a.y) * n.y).toBeGreaterThan(
        0,
      );
      expect((b.x - a.x) * n.x + (b.y - a.y) * n.y).toBeCloseTo(0, 10);
    }
    const rs = rivets(v, g);
    expect(rs).toHaveLength(6);
    const off =
      (Math.max(4.5 * scale, 1.5) + Math.max(2 * scale, 1) + 1) /
      Math.sin((56 * Math.PI) / 180);
    expect(rs[0]?.x).toBeCloseTo(37 - off, 10);
    expect(rs[0]?.y).toBeCloseTo(11 + 40 * scale, 10);
  }
});

it.each([
  [1280, 720, 1, 0, 13],
  [1280, 720, 1.5, 5, 13],
  [1280, 720, 2, 7, 12],
  [1920, 1080, 2, 4, 13],
  [390, 844, 1, 7, 8],
])("restricts native usable zooms at %dx%d UI %s", (w, h, ui, lo, hi) => {
  const s = layout(w, h, ui);
  expect(s.zoomLevels()).toEqual(
    Array.from({ length: hi - lo + 1 }, (_, i) => lo + i),
  );
  expect(s.clampZoom(-100)).toBe(lo);
  expect(s.clampZoom(100)).toBe(hi);
  const z = s.zoomed(100);
  expect(z.zoom).toBe(hi);
  expect(z.zoomLevels()).toEqual(s.zoomLevels());
  expect(z.cellScale).toBeCloseTo(s.fitScale * (zooms[hi] ?? 0), 12);
  expect(z.panels).toEqual(s.panels);
  expect(z.divider).toEqual(s.divider);
  expect(z.status).toEqual(s.status);
  expect(z.bolt).toEqual(s.bolt);
  expect(s.zoom).toBe(defaultZoom);
});

it("leaves only the top and outer rims in slim openings", () => {
  const s = layout(1280, 720);
  expect(s.slimOpening(0)).toEqual({ x: 6, y: 6, width: 626, height: 674 });
  expect(s.slimOpening(1)).toEqual({ x: 648, y: 6, width: 626, height: 674 });
  expect(layout(1920, 1080, 2).slimOpening(1)).toEqual({
    x: 976,
    y: 12,
    width: 932,
    height: 988,
  });
  expect(layout(1, 1).slimOpening(0)).toEqual({
    x: 6,
    y: 6,
    width: 0,
    height: 0,
  });
});

it("fits slim openings onto each map end by whole device pixels", () => {
  const s = layout(1920, 1080);
  const world = { x: 13, y: 23, width: 1090, height: 1880 };
  for (const side of [0, 1] as const) {
    const o = s.slimOpening(side);
    for (const [pair, row] of [
      [0, 0],
      [4, 0],
      [0, 18],
      [4, 18],
    ] as const) {
      const original = s.view(side, pair, row);
      original.shiftX = 1000;
      original.shiftY = -1000;
      const v = s.slimFit(original, side, world);
      expect(original.shiftX).toBe(1000);
      expect(original.shiftY).toBe(-1000);
      expect(v.pair).toBe(pair);
      expect(v.row).toBe(row);
      const a = v.screenToWorld(o.x, o.y);
      const b = v.screenToWorld(o.x + o.width, o.y + o.height);
      expect(a.x).toBeGreaterThanOrEqual(world.x - 1e-6);
      expect(a.y).toBeGreaterThanOrEqual(world.y - 1e-6);
      expect(b.x).toBeLessThanOrEqual(world.x + world.width + 1e-6);
      expect(b.y).toBeLessThanOrEqual(world.y + world.height + 1e-6);
      expect(v.shiftX * v.scale).toBeCloseTo(Math.round(v.shiftX * v.scale), 9);
      expect(v.shiftY * v.scale).toBeCloseTo(Math.round(v.shiftY * v.scale), 9);
    }
  }
  const inside = s.slimFit(s.view(0, 10, 10), 0, {
    x: 0,
    y: 0,
    width: 10000,
    height: 10000,
  });
  expect(inside.shiftX).toBeCloseTo(0, 12);
  expect(inside.shiftY).toBeCloseTo(0, 12);
  const small = s.slimFit(s.view(1, 0, 0), 1, {
    x: 0,
    y: 0,
    width: 300,
    height: 200,
  });
  const o = s.slimOpening(1);
  const a = small.worldToScreen(0, 0);
  const b = small.worldToScreen(300, 200);
  expect(Math.abs(a.x - o.x - (o.x + o.width - b.x))).toBeLessThanOrEqual(1);
  expect(Math.abs(a.y - o.y - (o.y + o.height - b.y))).toBeLessThanOrEqual(1);
});

it("stops zooming out before the world uncovers the slim opening", () => {
  const s = layout(1920, 1080);
  expect(s.fillZoom(1090, 1880)).toBe(2);
  expect(s.fillZoom(5000, 5000)).toBe(0);
  expect(s.fillZoom(200, 200)).toBe(defaultZoom);
  const [width, height] = hexgrid.bounds({ width: 8, lines: 46 });
  expect(s.fillZoom(width, height)).toBe(2);
});
