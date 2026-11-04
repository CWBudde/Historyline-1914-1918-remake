import { div, rem } from "./int.ts";
import type { Profile } from "./profile.ts";

/**
 * hexDistance is the distance helper (0e40:2421 in the demo) on its own argument order: two (x, y) pairs of a
 * map whose odd x sit half a cell lower, where x is the column. The movement code calls it that way.
 */
export function hexDistance(
  x1: number,
  y1: number,
  x2: number,
  y2: number,
): number {
  if (y1 > y2) {
    [x1, x2, y1, y2] = [x2, x1, y2, y1];
  }
  let lo: number;
  let hi: number;
  if (x2 >= x1) {
    lo = y1 - div(x1, 2);
    hi = y2 - div(x2, 2);
  } else {
    [x1, x2] = [x2, x1];
    lo = y1 - (div(x1, 2) + rem(x1, 2));
    hi = y2 - (div(x2, 2) + rem(x2, 2));
  }
  return x2 - x1 + Math.max(0, hi - lo);
}

/** distance is the battle distance between two cells of a map of the given width (Go's Profile.Distance). */
export function distance(
  p: Profile,
  width: number,
  a: number,
  b: number,
): number {
  if (p.transposedDistance) {
    return hexDistance(
      div(a, width),
      rem(a, width),
      div(b, width),
      rem(b, width),
    );
  }
  return hexDistance(
    rem(a, width),
    div(a, width),
    rem(b, width),
    div(b, width),
  );
}

/**
 * layoutDistance is the hex distance between two cells of a map of the given width on the map's own layout,
 * whatever the profile's battle distance (Go's Profile.LayoutDistance; it reads nothing of the profile).
 */
export function layoutDistance(
  _p: Profile,
  width: number,
  a: number,
  b: number,
): number {
  return hexDistance(
    rem(a, width),
    div(a, width),
    rem(b, width),
    div(b, width),
  );
}
