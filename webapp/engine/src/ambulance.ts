// A field ambulance (live +6 bit 0x0080, an extension type) treats the wounded. Once per game turn, with the gas
// (turnEffects), every foot unit (live +6 bit 0x01) next to an ambulance of its side on the side's map regains
// ambulanceHeal strength, up to its full group; a unit next to two ambulances regains it once. Neither the
// ambulance nor the unit may be held in a building or a transporter: a depot repairs there instead. The ambulance
// itself never fights (it has no weapon and never returns fire). (Go's ambulance.go.)

import { type Live, destroyed, liveCell } from "./board.ts";
import {
  type Event,
  type EventKind,
  type Game,
  liveOf,
  newEvent,
  phaseAction,
  sortedKeys,
  vacate,
} from "./cycle.ts";
import { gasTurn, ownerOf } from "./gas.ts";
import { at } from "./goruntime.ts";
import { tablesOf } from "./profile.ts";
import { neighbours } from "./support.ts";
import { mineTurn } from "./tunnel.ts";

/** Live +6: a field ambulance (extension types only). */
export const flagMedic = 0x0080;
/** AmbulanceHeal is the strength a unit next to an ambulance regains per turn. */
export const ambulanceHeal = 1;

/** Unit of side regained gain strength from an ambulance. */
export const evHealed = "healed" satisfies EventKind;

/** ambulance reports whether the unit is a field ambulance (Go's Live.Ambulance). */
export function ambulance(u: Readonly<Live>): boolean {
  return (u.flags6 & flagMedic) !== 0;
}

/** healTurn is the ambulances' turn, in unit order of the ambulances. */
export function healTurn(g: Game): Event[] {
  const evs: Event[] = [];
  const healed = new Set<number>();
  for (const i of sortedKeys(g.units)) {
    const a = liveOf(g, i);
    const side = ownerOf(a);
    if (
      !ambulance(a) ||
      destroyed(a) ||
      (at(a.action, 1) & 4) !== 0 ||
      side > 1 ||
      (a.flags4 & 0x4000) !== 0
    ) {
      continue;
    }
    for (const c of neighbours(liveCell(a, side), g.width, g.height)) {
      if (c < 0) {
        continue;
      }
      const j = at(at(g.maps, side), 2 * c + 1);
      const found = g.units.get(j);
      if (
        found === undefined ||
        healed.has(j) ||
        destroyed(found) ||
        (at(found.action, 1) & 4) !== 0 ||
        ownerOf(found) !== side ||
        (found.flags6 & 0x01) === 0 ||
        (found.flags4 & 0x4000) !== 0
      ) {
        continue;
      }
      const u = liveOf(g, j);
      const full = at(tablesOf(g.profile).types, u.type).group;
      const gain = Math.min(ambulanceHeal, full - u.size);
      if (gain <= 0) {
        continue;
      }
      u.size += gain;
      g.units.set(j, u);
      healed.add(j);
      evs.push(newEvent({ kind: evHealed, side, unit: j, cell: c, gain }));
    }
  }
  return evs;
}

/**
 * turnEffects are the extension units' effects once per game turn, in the resolution whose phase change counts
 * the turn (side 0's action phase): the tunnels (mineTurn), the gas (gasTurn), then the ambulances (healTurn). The
 * units they destroy leave the map before the phase change marks them.
 */
export function turnEffects(g: Game): Event[] {
  if (g.players[0].phase !== phaseAction) {
    return [];
  }
  const gone: number[] = [];
  const evs = mineTurn(g, gone);
  evs.push(...gasTurn(g, gone));
  vacate(g, gone);
  evs.push(...healTurn(g));
  return evs;
}
