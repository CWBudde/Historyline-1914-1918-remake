// Energy. Each side has one energy account that its buildings pay into every command cycle and that any of its
// buildings draws on: a repair costs 15 energy, 50 for a construction unit, restores the group size, keeps the
// experience and resets a construction unit's charges to 2; a factory builds a unit of its build list for the
// type's price (UNIT.DAT +0x3f, 255 never sold). What a building pays and what it builds are the scenario's
// (Game.buildings). Repairs and builds are given in the side's action phase, as the manual's action cycle; a new
// unit leaves the building from the next movement phase on.
//
// Assumed, not verified:
//   - Repairs and builds are given in an own building; they mark the unit moved (one order per unit and phase)
//     but do not enter the order list, so the move limit does not count them.
//   - A repair leaves the alternate strength as it is.
//   - A new unit takes the lowest free live record, the side's highest serial plus one, a free slot of the side
//     and the building's first cell as its position, as an arrival into an own building leaves it; a
//     construction unit starts with 2 charges. A transporter takes as +0x0a the lowest transporter record (below
//     holdFlag40) that none holds, as a scenario's transporter gets one: flag 0x20 with the classes its type
//     carries and its side, the building's first cell as its position (2*cell+1) and itself as the occupant.

import { type Live, destroyed, zeroLive } from "./board.ts";
import {
  type Hold,
  emptyHold,
  holdFlag40,
  inBuilding,
  slotFree,
} from "./container.ts";
import {
  type Command,
  type Event,
  type Game,
  cmdRepair,
  evBuilt,
  evCredit,
  evRepaired,
  holdEqual,
  holdOf,
  liveOf,
  newEvent,
  phaseAction,
  refuse,
} from "./cycle.ts";
import { serial } from "./capture.ts";
import { gasCharges, gasType } from "./gas.ts";
import { type GoError, at, setAt } from "./goruntime.ts";
import { div } from "./int.ts";
import { type MoveType, owns } from "./move.ts";
import { type Profile, tablesOf } from "./profile.ts";
import { liveUnits } from "./visibility.ts";
import { flagWire, typeFlag, wireCharges } from "./wire.ts";

/** Building is what the scenario gives a building record: the energy it credits each cycle and the types it builds (none: it only repairs). */
export interface Building {
  credits: number;
  /** Go's nil (JSON omitempty) is empty. */
  builds: number[];
  /** The types a computer side may build here; absent or empty: the default of its side. */
  computerBuilds?: number[];
}

/** builderCharges are a construction unit's charges after a repair. */
export const builderCharges = 2;

/** profileBuilder reports whether typ is a construction unit (Go's Profile.builder). */
export function profileBuilder(p: Profile, typ: number): boolean {
  return p.builders.includes(typ);
}

/**
 * charges are the charges a unit of typ holds in live +0x0a when placed, built or repaired, and whether it holds
 * any (Go's Profile.charges).
 */
export function charges(p: Profile, typ: number): [number, boolean] {
  if (profileBuilder(p, typ)) {
    return [builderCharges, true];
  }
  if (gasType(p, typ)) {
    return [gasCharges, true];
  }
  if (typeFlag(p, typ, flagWire)) {
    return [wireCharges, true];
  }
  return [0, false];
}

/** repairCost is the energy a repair of a unit of typ costs. */
export function repairCost(p: Profile, typ: number): number {
  return profileBuilder(p, typ) ? p.builderRepair : p.repair;
}

/** checkBuilding is check for a repair or a build. */
export function checkBuilding(g: Game, c: Command): GoError | null {
  if (at(g.players, c.side).phase !== phaseAction) {
    return refuse(0, "repairs and builds are given in the action phase");
  }
  const hold = c.hold ?? 0;
  const b = g.buildings.get(hold);
  const h = g.holds.get(hold);
  if (b === undefined || h === undefined) {
    return refuse(0, "no such building");
  }
  if ((h.flags & 2) !== 0 || !owns(h.flags, c.side)) {
    return refuse(0, "not the side's building");
  }
  if (c.kind === cmdRepair) {
    const slot = c.slot ?? 0;
    if (slot < 0 || slot >= 7) {
      return refuse(0, "no such slot");
    }
    const u = g.units.get(at(h.slots, 7 * c.side + slot));
    if (u === undefined) {
      return refuse(0, "the slot is empty");
    }
    if ((u.flags4 & 2) !== 0 || !owns(u.flags4, c.side)) {
      return refuse(0, "not the side's unit");
    }
    if ((u.flags4 & 0x200) !== 0) {
      return refuse(0, "the unit has its order");
    }
    if (at(g.energy, c.side) < repairCost(g.profile, u.type)) {
      return refuse(0, "not enough energy");
    }
    return null;
  }
  const typ = c.type ?? 0;
  const types = tablesOf(g.profile).types;
  if (
    !b.builds.includes(typ) ||
    typ < 0 ||
    typ >= types.length ||
    at(types, typ).cost === 0xff
  ) {
    return refuse(0, "the building does not build the type");
  }
  if (at(g.energy, c.side) < at(types, typ).cost) {
    return refuse(0, "not enough energy");
  }
  if (freeHoldSlot(h, c.side) < 0) {
    return refuse(0, "no free slot");
  }
  if (freeRecord(g) < 0) {
    return refuse(0, "no free unit record");
  }
  if (
    (at(moveTypes(g.profile), typ).class & 0x1000) !== 0 &&
    freeCarrier(g) < 0
  ) {
    return refuse(0, "no free container record");
  }
  return null;
}

// moveTypes is Go's p.Tables.Move.Types, which Go dereferences without a check: no movement tables panic.
function moveTypes(p: Profile): MoveType[] {
  const m = tablesOf(p).move;
  if (m === null) {
    throw new TypeError(
      "runtime error: invalid memory address or nil pointer dereference",
    );
  }
  return m.types;
}

/** freeHoldSlot is the first free slot of side's part of h, -1 if none (Go's freeSlot on a Hold). */
export function freeHoldSlot(h: Readonly<Hold>, side: number): number {
  for (let k = 0; k < 7; k++) {
    if (at(h.slots, 7 * side + k) > slotFree) {
      return k;
    }
  }
  return -1;
}

/**
 * freeRecord is the lowest unused live record; once every record has been used, it is the lowest record holding
 * a destroyed unit. It is -1 if none is available.
 */
export function freeRecord(g: Game): number {
  for (let i = 0; i < liveUnits; i++) {
    if (!g.units.has(i)) {
      return i;
    }
  }
  for (let i = 0; i < liveUnits; i++) {
    if (destroyed(g.units.get(i) ?? zeroLive())) {
      return i;
    }
  }
  return -1;
}

/**
 * freeCarrier is the lowest unused transporter record; once every record has been used, it is the lowest cleared
 * record. It is -1 if none is available.
 */
export function freeCarrier(g: Game): number {
  const empty = emptyHold();
  for (let i = 0; i < holdFlag40; i++) {
    const h = g.holds.get(i);
    if (h === undefined || holdEqual(h, empty)) {
      return i;
    }
  }
  for (let i = 0; i < holdFlag40; i++) {
    const h = g.holds.get(i);
    if (h !== undefined && (h.flags & 0x8000) !== 0) {
      return i;
    }
  }
  return -1;
}

/** repair carries out a checked repair. */
export function repair(g: Game, c: Command): Event[] {
  const hold = c.hold ?? 0;
  const i = at(holdOf(g, hold).slots, 7 * c.side + (c.slot ?? 0));
  const u = liveOf(g, i);
  setAt(g.energy, c.side, at(g.energy, c.side) - repairCost(g.profile, u.type));
  u.size = at(tablesOf(g.profile).types, u.type).group;
  const [n, ok] = charges(g.profile, u.type);
  if (ok) {
    u.hold = n;
  }
  u.flags4 |= 0x200;
  g.units.set(i, u);
  return [
    newEvent({
      kind: evRepaired,
      side: c.side,
      unit: i,
      cell: div(at(holdOf(g, hold).cells, 0), 2),
      energy: at(g.energy, c.side),
    }),
  ];
}

/** build carries out a checked build. */
export function build(g: Game, c: Command): Event[] {
  const typ = c.type ?? 0;
  const hold = c.hold ?? 0;
  const m = at(moveTypes(g.profile), typ);
  const h = holdOf(g, hold);
  const i = freeRecord(g);
  const k = freeHoldSlot(h, c.side);
  const first = at(h.cells, 0);
  const u: Live = {
    ...zeroLive(),
    budget: m.budget,
    size: at(tablesOf(g.profile).types, typ).group,
    flags4: m.class | c.side | 0x4200,
    flags6: m.flags,
    type: typ,
    serial: serial(g, c.side),
    hold: 0xff,
    pos: [first, first],
  };
  const [n, ok] = charges(g.profile, typ);
  if (ok) {
    u.hold = n;
  }
  if ((m.class & 0x1000) !== 0) {
    u.hold = freeCarrier(g);
    const r = emptyHold();
    setAt(r.cells, 0, first + 1);
    r.flags = 0x20 | m.carries | c.side;
    r.occupant = i;
    g.holds.set(u.hold, r);
  }
  g.units.set(i, u);
  setAt(h.slots, 7 * c.side + k, i);
  g.holds.set(hold, h);
  const mp = at(g.maps, c.side);
  const a = first + 1;
  if (at(mp, a) === 0xff) {
    setAt(mp, a, inBuilding);
  }
  setAt(
    g.energy,
    c.side,
    at(g.energy, c.side) - at(tablesOf(g.profile).types, typ).cost,
  );
  return [
    newEvent({
      kind: evBuilt,
      side: c.side,
      unit: i,
      cell: div(first, 2),
      energy: at(g.energy, c.side),
    }),
  ];
}

/** income is the energy side's buildings pay it each cycle: the credits of every building record it owns. */
export function income(g: Game, side: number): number {
  let paid = 0;
  for (const [i, b] of g.buildings) {
    const h = g.holds.get(i);
    if (h !== undefined && (h.flags & 2) === 0 && owns(h.flags, side)) {
      paid += b.credits;
    }
  }
  return paid;
}

/** credit pays the side that enters its action phase at the phase change ph its income, so each side is paid once per cycle; a side credited nothing gets no event. */
export function credit(g: Game, ph: Readonly<Event>): Event[] {
  const evs: Event[] = [];
  for (let k = 0; k < 2; k++) {
    const paid = income(g, k);
    if (at(g.players, k).phase !== phaseAction || paid === 0) {
      continue;
    }
    setAt(g.energy, k, at(g.energy, k) + paid);
    evs.push(
      newEvent({
        kind: evCredit,
        side: k,
        energy: at(g.energy, k),
        phase: [ph.phase[0], ph.phase[1]],
        turn: ph.turn,
      }),
    );
  }
  return evs;
}
