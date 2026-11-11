// Gas pioneers release a lasting gas cloud (gas.ts), an extension type that only a game on the extended rules
// fields (docs/new-units.md). Go's dosplay gas.go.

import * as rules from "./cycle.ts";
import type { BindingData, Board } from "./bind.ts";
import { cell as boundCell } from "./bind.ts";
import { liveCell } from "./board.ts";
import { buildingKind, kindFactory } from "./capture.ts";
import * as gases from "./gas.ts";
import { wiring } from "./wire.ts";
import type { Coord } from "./hexgrid.ts";
import { coordKey } from "./level.ts";
import { type Session } from "./playModel.ts";
import { levelHexes, refusal, wrap } from "./playEvent.ts";

/** originalTypes is the number of the original's unit types (UNIT.DAT); a dos index from here on is an extension type. */
export const originalTypes = 51;

/** gasPioneer reports whether unit releases gas instead of attacking (Go's Session.GasPioneer). */
export function gasPioneer(s: Session, unit: number): boolean {
  return gases.gasPioneer(rules.liveOf(s.game, unit));
}

/**
 * charges is what a construction unit, gas pioneers or a wiring party have left to work with, -1 for a unit without
 * charges.
 */
export function charges(s: Session, unit: number): number {
  const u = s.game.units.get(unit);
  if (
    u === undefined ||
    ((u.flags6 & 0x4000) === 0 && !gases.gasPioneer(u) && !wiring(u))
  )
    return -1;
  const b = u.hold & 0xff;
  return b >= 0x80 ? b - 0x100 : b;
}

/** canRelease throws why gas pioneer unit of side can release no gas now (none left); it returns when it can. */
export function canRelease(s: Session, side: number, unit: number): void {
  const u = s.game.units.get(unit);
  if (u === undefined || !gases.gasPioneer(u))
    throw refusal("not a gas pioneer");
  const e = rules.check(s.game, {
    kind: "attack",
    side,
    unit,
    cell: liveCell(u, side),
  });
  if (e instanceof rules.Refusal && e.code === rules.msgNoTarget) return;
  if (e !== null) throw wrap(e);
}

/** cloud is the hexes a cloud released by gas pioneer unit of side on hex to would cover, without ordering it. */
export function cloud(
  s: Session,
  side: number,
  unit: number,
  to: Coord,
): Coord[] {
  const c = boundCell(s.board, to);
  return c === null ? [] : levelHexes(s, gases.gasCells(s.game, side, unit, c));
}

/** gas are the gassed hexes (coordinate keys) and the turns each has left. Both sides see the cloud. */
export function gas(s: Session): Map<string, number> {
  const out = new Map<string, number>();
  for (const [c, t] of (s.game.gas ?? []).entries()) {
    if (t <= 0) continue;
    const [p] = levelHexes(s, [c]);
    if (p !== undefined) out.set(coordKey(p), t);
  }
  return out;
}

/** gasAt is the turns the cloud on hex has left, 0 for none. */
export function gasAt(s: Session, at: Coord): number {
  const c = boundCell(s.board, at);
  return c === null ? 0 : gases.gasAt(s.game, c);
}

/**
 * gasRates are the cloud's duration in turns and a foot unit's loss per turn in percent by the level's month
 * ("1915-04"; "" for the defaults): gas masks halve the loss from 1916, mustard gas doubles the duration from July
 * 1917.
 */
export function gasRates(month: string): [turns: number, loss: number] {
  let turns = gases.gasTurnsDefault,
    loss = gases.gasLossDefault;
  if (month >= "1916-01") loss = 15;
  if (month >= "1917-07") turns = 4;
  return [turns, loss];
}

/** builtIn reports whether a factory builds unit u in campaign month ("1915-04"; "" for any month; Go's data.Unit.BuiltIn). */
export function builtIn(
  u: Readonly<{ from: string; until: string }>,
  month: string,
): boolean {
  return (
    month === "" ||
    ((u.from === "" || month >= u.from) && (u.until === "" || month <= u.until))
  );
}

/** extensionTypes are the extension types of the remake's units that factories build in month ("" for any), ascending. */
export function extensionTypes(d: BindingData, month: string): number[] {
  const out = new Set<number>();
  for (const u of d.units.values())
    if (u.dos >= originalTypes && builtIn(u, month)) out.add(u.dos);
  return [...out].sort((a, b) => a - b);
}

/** usesExtension reports whether a board places or builds an extension type. */
export function usesExtension(b: Board): boolean {
  for (const u of b.units.values()) if (u.type >= originalTypes) return true;
  for (const bd of b.buildings.values())
    for (const t of [...bd.builds, ...(bd.computerBuilds ?? [])])
      if (t >= originalTypes) return true;
  return false;
}

/**
 * extend sets a new game on the extended rules up: with builds every factory also builds the extension types of the
 * month, and the cloud follows it.
 */
export function extend(
  g: rules.Game,
  d: BindingData,
  month: string,
  builds: boolean,
): void {
  if (builds) {
    const ext = extensionTypes(d, month);
    for (const [i, bd] of g.buildings) {
      if (buildingKind(i) !== kindFactory) continue;
      for (const t of ext) if (!bd.builds.includes(t)) bd.builds.push(t);
    }
  }
  const [turns, loss] = gasRates(month);
  if (turns !== gases.gasTurnsDefault) g.gasTurns = turns;
  if (loss !== gases.gasLossDefault) g.gasLoss = loss;
}
