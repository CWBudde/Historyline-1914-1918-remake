// Fixed terrain groups from internal/level/compound.go.
import { at as element } from "./goruntime.ts";
import { type Coord, around } from "./hexgrid.ts";

export interface Compound {
  prefix: string;
  entrance: string;
  pieces: Readonly<Record<string, readonly number[]>>;
}
const diamond = { N: [], W: [3], E: [1], S: [2] };
export const compounds: readonly Compound[] = [
  {
    prefix: "hq",
    entrance: "N",
    pieces: {
      N: [],
      NW: [3],
      NE: [1],
      C: [2],
      SW: [2, 3],
      SE: [2, 1],
      S: [2, 2],
    },
  },
  { prefix: "cmp", entrance: "N", pieces: diamond },
  { prefix: "factory", entrance: "E", pieces: diamond },
  { prefix: "mount", entrance: "", pieces: diamond },
];
export function at(cp: Compound, n: Coord, piece: string): Coord {
  for (const d of cp.pieces[piece] ?? []) n = element(around(n), d);
  return n;
}
export function compoundPiece(
  terrain: string,
): { compound: Compound; piece: string; suffix: string } | null {
  const suffix = terrain.endsWith("_w") ? "_w" : "";
  const base = suffix === "" ? terrain : terrain.slice(0, -2);
  for (const compound of compounds) {
    if (!base.startsWith(compound.prefix)) continue;
    const piece = base.slice(compound.prefix.length);
    if (Object.hasOwn(compound.pieces, piece))
      return { compound, piece, suffix };
  }
  return null;
}
