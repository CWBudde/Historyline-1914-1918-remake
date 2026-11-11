// The extension's work units beyond the gas pioneers (wire.ts, tunnel.ts, ambulance.ts; docs/new-units.md): the
// wiring party lays barbed wire, the tunnellers dig a mine under a neighbouring hex, the field ambulance heals the
// foot units next to it. A flamethrower attacks as any unit does. Go's dosplay works.go.

import * as rules from "./cycle.ts";
import { cell as boundCell } from "./bind.ts";
import { liveCell } from "./board.ts";
import { sapper as ruleSapper } from "./sapper.ts";
import { builder as ruleBuilder } from "./builder.ts";
import { gasPioneer as ruleGasPioneer } from "./gas.ts";
import * as wires from "./wire.ts";
import * as tunnelRules from "./tunnel.ts";
import { ambulance as ruleAmbulance } from "./ambulance.ts";
import type { Coord } from "./hexgrid.ts";
import { type Session, hex } from "./playModel.ts";
import { levelHexes, refusal, wrap } from "./playEvent.ts";

/** wiring reports whether unit lays barbed wire instead of attacking (Go's Session.Wiring). */
export function wiring(s: Session, unit: number): boolean {
  return wires.wiring(rules.liveOf(s.game, unit));
}

/** tunneller reports whether unit digs mines instead of attacking (Go's Session.Tunneller). */
export function tunneller(s: Session, unit: number): boolean {
  return tunnelRules.tunneller(rules.liveOf(s.game, unit));
}

/** ambulance reports whether unit is a field ambulance, which heals the foot units next to it once per turn. */
export function ambulance(s: Session, unit: number): boolean {
  return ruleAmbulance(rules.liveOf(s.game, unit));
}

/** flamethrower reports whether unit's attacks ignore the defender's cover. */
export function flamethrower(s: Session, unit: number): boolean {
  return wires.flamethrower(rules.liveOf(s.game, unit));
}

/**
 * worker reports whether unit's action order is work on a neighbouring hex, not an attack: a sapper's, a
 * construction unit's, gas pioneers', a wiring party's or tunnellers'.
 */
export function worker(s: Session, unit: number): boolean {
  const u = rules.liveOf(s.game, unit);
  return (
    ruleSapper(u) ||
    ruleBuilder(u) ||
    ruleGasPioneer(u) ||
    wires.wiring(u) ||
    tunnelRules.tunneller(u)
  );
}

// canWork throws why a work unit of side has no work order now (no charge left, a tunnel dug already); it returns
// when it has one somewhere: its own cell is never a target, so a refusal for no target there passes.
function canWork(s: Session, side: number, unit: number): void {
  const u = rules.liveOf(s.game, unit);
  const e = rules.check(s.game, {
    kind: "attack",
    side,
    unit,
    cell: liveCell(u, side),
  });
  if (e instanceof rules.Refusal && e.code === rules.msgNoTarget) return;
  if (e !== null) throw wrap(e);
}

/** canWire throws why wiring party unit of side can lay no wire now (none left); it returns when it can. */
export function canWire(s: Session, side: number, unit: number): void {
  const u = s.game.units.get(unit);
  if (u === undefined || !wires.wiring(u)) throw refusal("not a wiring party");
  canWork(s, side, unit);
}

/** canTunnel throws why tunnellers unit of side can start no tunnel now (one dug already); it returns when they can. */
export function canTunnel(s: Session, side: number, unit: number): void {
  const u = s.game.units.get(unit);
  if (u === undefined || !tunnelRules.tunneller(u))
    throw refusal("not tunnellers");
  canWork(s, side, unit);
}

/** wireHexes are the hexes wiring party unit of side would wire from its hex on hex to, without ordering it. */
export function wireHexes(
  s: Session,
  side: number,
  unit: number,
  to: Coord,
): Coord[] {
  const c = boundCell(s.board, to);
  return c === null
    ? []
    : levelHexes(s, wires.wireCells(s.game, side, unit, c));
}

/** Tunnel is a mine side's tunnellers dig: the tunnellers (unit), the hex they dig from and the hex the mine lies under, the turns dug and the turns it takes. */
export interface Tunnel {
  unit: number;
  from: Coord;
  mine: Coord;
  dug: number;
  turns: number;
}

/** tunnelsOf are the mines side digs; the other side's stay hidden (Go's Session.Tunnels). */
export function tunnels(s: Session, side: number): Tunnel[] {
  return (s.game.tunnels ?? [])
    .filter((t) => t.side === side)
    .map((t) => ({
      unit: t.unit,
      from: hex(s, t.from),
      mine: hex(s, t.cell),
      dug: t.dug,
      turns: tunnelRules.tunnelTurns,
    }));
}

/** tunnelOf is the mine unit digs, null where it digs none. */
export function tunnelOf(s: Session, unit: number): Tunnel | null {
  const t = tunnelRules.tunnelOf(s.game, unit);
  return t === null
    ? null
    : {
        unit,
        from: hex(s, t.from),
        mine: hex(s, t.cell),
        dug: t.dug,
        turns: tunnelRules.tunnelTurns,
      };
}
