import { level, play } from "@historyline/engine";
import type { MapBattle } from "./mapBattle.ts";

/** How bright the ground out of an aimed order's reach stays (cursor.go darkKeep). */
export const darkKeep = 0.62;
/**
 * aimReach is what keeps its colour while an order is aimed (cursor.go reach): the destinations or targets with
 * every cell of the buildings among them, and the unit's own cell; null while none is aimed.
 */
export function aimReach(
  b: Pick<
    MapBattle,
    "selected" | "exiting" | "destinations" | "targets" | "session"
  >,
  view: play.View,
): Set<string> | null {
  if (b.selected === null) return null;
  const moves = b.destinations(),
    targets = b.targets();
  if (moves.size === 0 && targets.size === 0) return null;
  const out = new Set<string>(),
    from =
      b.exiting?.from ??
      [...view.units.values()].find((u) => u.index === b.selected)?.at;
  if (from !== undefined) out.add(level.coordKey(from));
  for (const key of [...moves, ...targets]) {
    out.add(key);
    for (const c of play.footprint(b.session, play.coordFromKey(key)))
      out.add(level.coordKey(c));
  }
  return out;
}
