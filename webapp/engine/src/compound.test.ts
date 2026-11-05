import { expect, it } from "vitest";
import { at, compoundPiece, compounds } from "./compound.ts";

it.each([
  ["hqN", "hq", "N", ""],
  ["hqSE_w", "hq", "SE", "_w"],
  ["cmpW", "cmp", "W", ""],
  ["factoryE_w", "factory", "E", "_w"],
  ["mountS", "mount", "S", ""],
])(
  "splits %s into its compound, piece and season",
  (terrain, prefix, piece, suffix) => {
    const result = compoundPiece(terrain);
    expect(result?.compound.prefix).toBe(prefix);
    expect(result?.piece).toBe(piece);
    expect(result?.suffix).toBe(suffix);
  },
);
it.each(["mountain", "hqX", "grass"])("%s is no compound", (terrain) => {
  expect(compoundPiece(terrain)).toBeNull();
});
it("stamps the HQ and diamonds in Go's hexes", () => {
  const n = { line: 4, col: 3 };
  const diamond = {
    N: n,
    W: { line: 5, col: 2 },
    E: { line: 5, col: 3 },
    S: { line: 6, col: 3 },
  };
  const hq = {
    N: n,
    NW: diamond.W,
    NE: diamond.E,
    C: diamond.S,
    SW: { line: 7, col: 2 },
    SE: { line: 7, col: 3 },
    S: { line: 8, col: 3 },
  };
  for (const cp of compounds) {
    const pieces = cp.prefix === "hq" ? hq : diamond;
    expect(
      Object.fromEntries(Object.keys(cp.pieces).map((p) => [p, at(cp, n, p)])),
    ).toEqual(pieces);
    expect(cp.entrance).toBe(
      cp.prefix === "mount" ? "" : cp.prefix === "factory" ? "E" : "N",
    );
  }
});
