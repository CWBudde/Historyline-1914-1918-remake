// Twin of internal/hexgrid/hexgrid_test.go.

import { describe, expect, it } from "vitest";
import { DecodeError, array, int, object } from "./decode.ts";
import {
  type Coord,
  type Grid,
  around,
  center,
  contains,
  cube,
  distance,
  neighbors,
  size,
} from "./hexgrid.ts";
import { readJSON } from "./testing/files.ts";

interface GoldenLevel {
  width: number;
  lines: number;
  hexes: { hex: Coord; neighbors: Coord[] }[];
}

function pair(v: unknown, what: string): Coord {
  const p = array(v, what);
  if (p.length !== 2) {
    throw new DecodeError(`${what}: not a pair`);
  }
  return { line: int(p[0], what), col: int(p[1], what) };
}

// golden is testdata/adjacent.json, a frozen golden of the Python remake's adjacentHexes in hlrData.py.
function loadGolden(): Map<string, GoldenLevel> {
  const raw = object(
    readJSON("internal/hexgrid/testdata/adjacent.json"),
    "golden",
  );
  const g = new Map<string, GoldenLevel>();
  for (const [name, v] of Object.entries(raw)) {
    const lv = object(v, name, ["width", "lines", "hexes"]);
    g.set(name, {
      width: int(lv["width"], `${name}.width`),
      lines: int(lv["lines"], `${name}.lines`),
      hexes: array(lv["hexes"], `${name}.hexes`).map((h) => {
        const o = object(h, name, ["hex", "neighbors"]);
        return {
          hex: pair(o["hex"], `${name}.hex`),
          neighbors: array(o["neighbors"], `${name}.neighbors`).map((n) =>
            pair(n, `${name}.neighbors`),
          ),
        };
      }),
    });
  }
  return g;
}

const golden = loadGolden();
const key = (c: Coord): string => `${c.line},${c.col}`;

describe("hexgrid", () => {
  it("Neighbors match Python on every hex of adjacent.json", () => {
    expect(golden.size).toBe(4);
    // The grids are the sizes of the Python remake's levels.
    let total = 0;
    for (const [name, lv] of golden) {
      const grid: Grid = { width: lv.width, lines: lv.lines };
      let count = 0;
      for (const h of lv.hexes) {
        expect(contains(grid, h.hex), `${name}: ${key(h.hex)}`).toBe(true);
        expect(neighbors(grid, h.hex), `${name}: ${key(h.hex)}`).toStrictEqual(
          h.neighbors,
        );
        count++;
      }
      expect(count, name).toBe(size(grid));
      total += count;
    }
    expect(total).toBe(1788);
  });

  it("edge cases", () => {
    const grid: Grid = { width: 8, lines: 47 }; // PULSE
    const c = (line: number, col: number): Coord => ({ line, col });
    const containsCases: [Coord, boolean][] = [
      [c(0, 0), true],
      [c(0, 7), true], // last hex of an even line
      [c(1, 6), true], // last hex of an odd line
      [c(1, 7), false], // odd lines are one hex shorter
      [c(46, 7), true], // last line is even
      [c(47, 0), false], // past the last line
      [c(-1, 0), false],
      [c(0, -1), false],
      [c(0, 8), false],
    ];
    for (const [hex, want] of containsCases) {
      expect(contains(grid, hex), key(hex)).toBe(want);
    }
    const neighborCases: [Coord, Coord[]][] = [
      [c(0, 0), [c(1, 0), c(2, 0)]],
      [c(0, 7), [c(2, 7), c(1, 6)]],
      [c(1, 0), [c(0, 1), c(2, 1), c(3, 0), c(2, 0), c(0, 0)]],
      [c(1, 6), [c(0, 7), c(2, 7), c(3, 6), c(2, 6), c(0, 6)]],
      [c(2, 0), [c(1, 0), c(3, 0), c(4, 0), c(0, 0)]],
      [c(2, 7), [c(4, 7), c(3, 6), c(1, 6), c(0, 7)]],
      [c(20, 3), [c(19, 3), c(21, 3), c(22, 3), c(21, 2), c(19, 2), c(18, 3)]],
      [c(21, 3), [c(20, 4), c(22, 4), c(23, 3), c(22, 3), c(20, 3), c(19, 3)]],
      [c(46, 0), [c(45, 0), c(44, 0)]],
      [c(46, 7), [c(45, 6), c(44, 7)]],
    ];
    for (const [hex, want] of neighborCases) {
      expect(neighbors(grid, hex), key(hex)).toStrictEqual(want);
    }
    expect(size(grid)).toBe(24 * 8 + 23 * 7);
  });

  // The cube distance equals the breadth-first step count over the Python adjacency, which is what gui.py
  // calculateDistance computes.
  it("Distance matches BFS", () => {
    for (const [name, lv] of golden) {
      if (name.startsWith("BACK-")) {
        continue; // same size as PULSE
      }
      const adj = new Map<string, Coord[]>();
      const hexes = new Map<string, Coord>();
      for (const h of lv.hexes) {
        adj.set(key(h.hex), h.neighbors);
        hexes.set(key(h.hex), h.hex);
      }
      for (const from of hexes.values()) {
        const dist = new Map<string, number>([[key(from), 0]]);
        const queue: Coord[] = [from];
        for (let i = 0; i < queue.length; i++) {
          const cur = queue[i] as Coord;
          const d = dist.get(key(cur)) ?? 0;
          for (const n of adj.get(key(cur)) ?? []) {
            if (!dist.has(key(n))) {
              dist.set(key(n), d + 1);
              queue.push(n);
            }
          }
        }
        expect(dist.size, `${name}: from ${key(from)}`).toBe(adj.size);
        for (const [to, want] of dist) {
          const got = distance(from, hexes.get(to) ?? from);
          if (got !== want) {
            expect.fail(
              `${name}: Distance(${key(from)}, ${to}) = ${got}, BFS ${want}`,
            );
          }
        }
      }
    }
  });

  it("cube invariant", () => {
    const grid: Grid = { width: 12, lines: 47 };
    for (let line = 0; line < grid.lines; line++) {
      for (let col = 0; col < grid.width; col++) {
        const c: Coord = { line, col };
        if (!contains(grid, c)) {
          continue;
        }
        const [q, r, s] = cube(c);
        expect(q + r + s, key(c)).toBe(0);
        for (const n of neighbors(grid, c)) {
          expect(distance(c, n), `${key(c)} to ${key(n)}`).toBe(1);
        }
      }
    }
  });

  it("Around is geometric", () => {
    // Around lists all six neighbours, on the map or not, in the order NE, SE, S, SW, NW, N; their centres
    // lie in exactly those directions.
    const want = [
      [71, -40],
      [71, 40],
      [0, 80],
      [-71, 40],
      [-71, -40],
      [0, -80],
    ];
    for (const c of [
      { line: 0, col: 0 },
      { line: 1, col: 0 },
      { line: 4, col: 3 },
      { line: 7, col: 5 },
    ]) {
      const [cx, cy] = center(c);
      expect(
        around(c).map((n) => {
          const [nx, ny] = center(n);
          return [nx - cx, ny - cy];
        }),
        key(c),
      ).toStrictEqual(want);
    }
    const g: Grid = { width: 3, lines: 4 };
    const origin: Coord = { line: 0, col: 0 };
    const on = around(origin).filter((n) => contains(g, n));
    expect(neighbors(g, origin)).toStrictEqual(on);
  });

  it("never gives -0 for negative coordinates", () => {
    const [q, r, s] = cube({ line: -2, col: 0 });
    expect([q, r, s]).toStrictEqual([0, -1, 1]);
    expect(cube({ line: 0, col: 0 })).toStrictEqual([0, 0, 0]);
    expect(distance({ line: -4, col: -1 }, { line: -4, col: -1 })).toBe(0);
  });
});
