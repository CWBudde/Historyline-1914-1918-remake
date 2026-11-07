// The command cycle. Both sides plan at the same time: the side in the movement phase moves its units, one order
// at a time and at once, on its own map; the side in the action phase gives attack orders, which wait in its
// order list. When both sides have confirmed, the order handler resolves the action side's list and the phase
// change brings the two maps into line and swaps the phases. Game holds that state; every command and the
// resolution change it at once and return the events in the order they happened, so no state waits on an
// animation.
//
// Go's methods on *Game take the game first (g.Apply(c) is apply(g, c)). Go copies Live and Hold values out of
// the game's maps and writes them back; the port does the same (liveOf, holdOf), so a value read earlier stays
// stale exactly where Go's does. Go's nil slices are empty arrays, its nil pointers null.

import {
  type Live,
  type Message,
  changed,
  copyLive,
  decodeHold,
  decodeLive,
  destroyed,
  liveCell,
  liveRecord,
  messages,
  moveTablesOf,
  zeroLive,
} from "./board.ts";
import { builder, checkBuilder, orderDepot } from "./builder.ts";
import { weapon, classBits } from "./casualties.ts";
import {
  type Hold,
  containerRecords,
  emptyHold,
  exitOn,
  exitQuery,
  holdFlag40,
  holdFlag100,
  holdFlag400,
  holdRecord,
  holdRecords,
  holdTable,
  inBuilding,
  slotFree,
} from "./container.ts";
import { markKinds } from "./capture.ts";
import { type Battlefield, type Outcome, type Unit } from "./engage.ts";
import {
  type Building,
  builderCharges,
  checkBuilding,
  credit,
  build as buildUnit,
  profileBuilder,
  repair,
} from "./energy.ts";
import { GoError, at, setAt } from "./goruntime.ts";
import { div } from "./int.ts";
import { layoutDistance } from "./distance.ts";
import {
  type MoveTables,
  type Movement,
  type Piece,
  newMovement,
  owns,
} from "./move.ts";
import { type Order, resolveOrders } from "./orders.ts";
import { type Profile, setsUp, tablesOf } from "./profile.ts";
import type { MoveOrder, StepRun } from "./record.ts";
import { executeOn, moveStateOn } from "./route.ts";
import {
  actionTrench,
  msgTrench,
  sapper,
  trenchTarget,
  work,
} from "./sapper.ts";
import { type End, end } from "./victory.ts";
import { type Sighted, type PhaseChange, changePhase } from "./visibility.ts";

/** The phases (player record +0x16). */
export const phaseMovement = 1;
export const phaseAction = 2;

/** Status messages of the cycle. */
/** The attack order: no target in the cell. */
export const msgNoTarget = 0x45;
/** The attack order taken. */
export const msgAttack = 0x49;
/** The move limit reached. */
export const msgLimit = 0x53;
/** The fire button cancels the confirmation. */
export const msgReadyUndone = 0x34;

/** Player is a side's part of the cycle: its player record (32e2:4d88 + 0x31·side in the demo) and its order list. */
export interface Player {
  /** +0x16: phaseMovement or phaseAction. */
  phase: number;
  /** +0x17 = 5: the side has confirmed its orders. */
  ready: boolean;
  /** 32e2:4d29/4d40 bit 2; the move limit holds for human sides only. */
  computer: boolean;
  /**
   * The side's order list (32e2:4d3a + 0x17·side): the units of its moves, exits and attacks in the order
   * given. Its length is the count (4d3e) the move limit compares. Only the action side's list is resolved.
   */
  orders: number[];
}

/** newPlayer is Go's Player{}. */
export function newPlayer(): Player {
  return { phase: 0, ready: false, computer: false, orders: [] };
}

/**
 * Game is the state of a battle the command cycle works on: both sides' maps (two numbers per cell, the terrain
 * record and the occupant, 0xff empty), the live records of the placed units and the container records, the
 * players, the turn counter (32e2:4cab), the move limit (the option 1653:253c: 4, 8 or 16 orders per side and
 * phase; 0 none), the option "with concealed inventories", which counts for the score, the battle's end once a
 * resolution has decided it (null while it is open), each side's energy account and the scenario's buildings by
 * container record (energy.ts). The fields are Go's, in Go's order.
 */
export interface Game {
  profile: Profile;
  width: number;
  height: number;
  maps: [number[], number[]];
  units: Map<number, Live>;
  holds: Map<number, Hold>;
  players: [Player, Player];
  turn: number;
  limit: number;
  concealed: boolean;
  ended: End | null;
  energy: [number, number];
  /** Go's nil map is an empty Map. */
  buildings: Map<number, Building>;
}

/** Placement is a unit as a scenario places it: its type, owner (0 or 1, 2 neutral), cell and strength; absent fields are 0. */
export interface Placement {
  type: number;
  owner?: number;
  cell?: number;
  /** 0: the type's full group. */
  size?: number;
  alt?: number;
  exp?: number;
}

/** liveOf is Go's g.Units[i]: a copy of the unit's live record, the zero record where there is none. */
export function liveOf(g: Game, i: number): Live {
  const u = g.units.get(i);
  return u === undefined ? zeroLive() : copyLive(u);
}

/** zeroHold is Go's Hold{}: every byte zero, the slots too. */
export function zeroHold(): Hold {
  return {
    slots: Array.from({ length: 14 }, () => 0),
    cells: [0, 0, 0, 0],
    flags: 0,
    occupant: 0,
  };
}

/** copyHold is Go's assignment of a Hold: a copy that shares nothing with h. */
export function copyHold(h: Readonly<Hold>): Hold {
  return {
    slots: [...h.slots],
    cells: [...h.cells],
    flags: h.flags,
    occupant: h.occupant,
  };
}

/** holdOf is Go's g.Holds[r]: a copy of the container record, the zero record where there is none. */
export function holdOf(g: Game, r: number): Hold {
  const h = g.holds.get(r);
  return h === undefined ? zeroHold() : copyHold(h);
}

/** holdEqual is Go's == on two Hold values. */
export function holdEqual(a: Readonly<Hold>, b: Readonly<Hold>): boolean {
  return (
    a.flags === b.flags &&
    a.occupant === b.occupant &&
    a.slots.length === b.slots.length &&
    a.slots.every((v, i) => v === b.slots[i]) &&
    a.cells.length === b.cells.length &&
    a.cells.every((v, i) => v === b.cells[i])
  );
}

/** sortedKeys are a map's keys, ascending (Go's slices.Sorted(maps.Keys(m))). */
export function sortedKeys(m: ReadonlyMap<number, unknown>): number[] {
  return [...m.keys()].sort((a, b) => a - b);
}

// placementString is Go's %+v of a Placement.
function placementString(u: Placement): string {
  return `{Type:${u.type} Owner:${u.owner ?? 0} Cell:${u.cell ?? 0} Size:${u.size ?? 0} Alt:${u.alt ?? 0} Exp:${u.exp ?? 0}}`;
}

/**
 * newGame sets up a battle on terrain (a terrain record per cell): both maps show every unit, side 0 starts in
 * the movement phase and side 1 in the action phase (0661:0bb2). A unit's live record gets what the constructor
 * (159e:0006) is known to copy: the budget (type byte 0), the class word (+0x0d) with the owner's bit into +4 and
 * the flags word (+0x11) into +6; its hold is 0xff.
 *
 * The container state follows from holds, as an arrival leaves it:
 *   - A unit in a slot of a record is held (+4 0x4000), its position on both sides is the record's first cell
 *     and no map shows it; its cell is read only to find a transporter's own record. A building record
 *     (holdFlag40 and up) with a unit in a side's slots shows inBuilding on that side's map at its first cell
 *     where no unit stands.
 *   - A transporter (+4 0x1000) gets as +0x0a the lowest record of flag 0x20 whose first cell is its own
 *     (2*cell+1) that no transporter of a lower index took; without one newGame fails.
 *   - A construction unit (Profile.builders) starts with 2 charges in +0x0a.
 *
 * It throws a GoError where the scene does not fit together, ErrNoMovement without movement tables.
 */
export function newGame(
  p: Profile,
  width: number,
  height: number,
  terrain: readonly number[],
  units: ReadonlyMap<number, Placement>,
  holds: ReadonlyMap<number, Hold> = new Map(),
): Game {
  const t = moveTablesOf(p);
  if (terrain.length !== width * height) {
    throw new GoError(
      `dosrules: ${terrain.length} terrain records for ${width * height} cells`,
    );
  }
  const g: Game = {
    profile: p,
    width,
    height,
    maps: [[], []],
    units: new Map(),
    holds: new Map(),
    players: [newPlayer(), newPlayer()],
    turn: 0,
    limit: 0,
    concealed: false,
    ended: null,
    energy: [0, 0],
    buildings: new Map(),
  };
  for (let k = 0; k < 2; k++) {
    const m: number[] = [];
    for (const tr of terrain) {
      m.push(tr, 0xff);
    }
    setAt(g.maps, k, m);
  }
  const recs = sortedKeys(holds);
  const holdAt = (r: number): Hold => {
    const h = holds.get(r);
    return h === undefined ? zeroHold() : h;
  };
  const inRec = new Map<number, number>(); // held unit -> its record
  for (const r of recs) {
    const h = holdAt(r);
    if (r < 0 || r >= holdRecords) {
      throw new GoError(`dosrules: container record ${r}`);
    }
    for (const u of h.slots) {
      if (u > slotFree) {
        continue;
      }
      if (!units.has(u)) {
        throw new GoError(
          `dosrules: container record ${r} holds unit ${u}, which is not placed`,
        );
      }
      const o = inRec.get(u);
      if (o !== undefined && o !== r) {
        throw new GoError(
          `dosrules: unit ${u} is held in records ${o} and ${r}`,
        );
      }
      inRec.set(u, r);
    }
    g.holds.set(r, copyHold(h));
  }
  const taken = new Set<number>();
  for (const i of sortedKeys(units)) {
    const u = units.get(i) ?? { type: 0 };
    const owner = u.owner ?? 0;
    const cell = u.cell ?? 0;
    const r = inRec.get(i);
    const held = r !== undefined;
    if (
      i < 0 ||
      i > slotFree ||
      (!held && (cell < 0 || cell >= width * height)) ||
      u.type < 0 ||
      u.type >= t.types.length ||
      owner < 0 ||
      owner > 2
    ) {
      throw new GoError(`dosrules: unit ${i}: ${placementString(u)}`);
    }
    const typ = at(t.types, u.type);
    let size = u.size ?? 0;
    if (size === 0) {
      size = at(tablesOf(p).types, u.type).group;
    }
    const l: Live = {
      ...zeroLive(),
      budget: typ.budget,
      exp: u.exp ?? 0,
      size,
      alt: u.alt ?? 0,
      flags4: typ.class | owner,
      flags6: typ.flags,
      type: u.type,
      hold: 0xff,
    };
    if (profileBuilder(p, u.type)) {
      l.hold = builderCharges;
    }
    if ((l.flags4 & 0x1000) !== 0) {
      l.hold = -1;
      for (const k of recs) {
        const h = holdAt(k);
        if (
          !taken.has(k) &&
          (h.flags & 0x20) !== 0 &&
          at(h.cells, 0) === 2 * cell + 1
        ) {
          l.hold = k;
          taken.add(k);
          break;
        }
      }
      if (l.hold < 0) {
        throw new GoError(
          `dosrules: transporter ${i} at cell ${cell} has no container record`,
        );
      }
    }
    for (let k = 0; k < 2; k++) {
      if (held) {
        setAt(l.pos, k, at(holdAt(r).cells, 0));
        continue;
      }
      setAt(l.pos, k, 2 * cell + 1);
      setAt(at(g.maps, k), 2 * cell + 1, i);
    }
    if (held) {
      l.flags4 |= 0x4000;
    }
    g.units.set(i, l);
  }
  for (const r of recs) {
    const h = holdAt(r);
    const a = at(h.cells, 0) + 1;
    if (r < holdFlag40 || a < 1 || a >= 2 * width * height) {
      continue;
    }
    for (let k = 0; k < 2; k++) {
      const m = at(g.maps, k);
      if (
        h.slots.slice(7 * k, 7 * k + 7).some((u) => u <= slotFree) &&
        at(m, a) === 0xff
      ) {
        setAt(m, a, inBuilding);
      }
    }
  }
  markKinds(g);
  g.players[0].phase = phaseMovement;
  g.players[1].phase = phaseAction;
  return g;
}

/** CommandKind names an order. */
export type CommandKind = "move" | "exit" | "attack" | "repair" | "build";

/** The orders a side gives. */
/** Unit to cell. */
export const cmdMove = "move" satisfies CommandKind;
/** The unit in slot of hold to cell. */
export const cmdExit = "exit" satisfies CommandKind;
/** Unit against cell. */
export const cmdAttack = "attack" satisfies CommandKind;
/** The unit in slot of building hold. */
export const cmdRepair = "repair" satisfies CommandKind;
/** A unit of type in building hold. */
export const cmdBuild = "build" satisfies CommandKind;

/** Command is an order of side. The optional fields are Go's omitempty ones; absent is 0. */
export interface Command {
  kind: CommandKind;
  side: number;
  unit?: number;
  hold?: number;
  slot?: number;
  cell: number;
  type?: number;
}

/** Refusal is a command the cycle does not take. Code is the status message the original shows, 0 where it shows none. */
export class Refusal extends GoError {
  readonly code: number;
  readonly reason: string;

  constructor(code: number, reason: string) {
    super(
      code !== 0
        ? `dosrules: ${reason} (message 0x${code.toString(16)})`
        : `dosrules: ${reason}`,
    );
    this.code = code;
    this.reason = reason;
  }
}

/** refuse is a Refusal (Go's refuse). */
export function refuse(code: number, reason: string): Refusal {
  return new Refusal(code, reason);
}

/** ErrNotReady is thrown by resolve while a side has not confirmed. */
export class ErrNotReady extends GoError {
  constructor() {
    super("dosrules: both sides must confirm before the orders resolve");
  }
}

/** ErrOver is thrown by resolve once the battle has ended. */
export class ErrOver extends GoError {
  constructor() {
    super("dosrules: the battle is over");
  }
}

/** EventKind names an event. */
export type EventKind =
  | "moved"
  | "refused"
  | "attack"
  | "ready"
  | "unready"
  | "battle"
  | "skipped"
  | "phase"
  | "end"
  | "repaired"
  | "built"
  | "credit"
  | "captured"
  | "work"
  | "dug"
  | "filled"
  | "halfDepot"
  | "depot";

/** A move or an exit ran: route. */
export const evMoved = "moved" satisfies EventKind;
/** The order code or the move start refused it: messages. */
export const evRefused = "refused" satisfies EventKind;
/** An attack order entered the list. */
export const evAttack = "attack" satisfies EventKind;
export const evReady = "ready" satisfies EventKind;
export const evUnready = "unready" satisfies EventKind;
/** An attack order resolved: outcome. */
export const evBattle = "battle" satisfies EventKind;
/** An attack order whose cell held no unit. */
export const evSkipped = "skipped" satisfies EventKind;
/** The phase change: phase, turn. */
export const evPhase = "phase" satisfies EventKind;
/** The battle ended: side is the winner (-1 a draw), end. */
export const evEnd = "end" satisfies EventKind;
/** A unit repaired: energy left. */
export const evRepaired = "repaired" satisfies EventKind;
/** A unit built: unit, the building's cell, energy left. */
export const evBuilt = "built" satisfies EventKind;
/** The side entering its action phase credited: energy. */
export const evCredit = "credit" satisfies EventKind;
/** Side took building record hold at cell with unit (capture.ts). */
export const evCaptured = "captured" satisfies EventKind;
/** A sapper's or construction unit's work order on cell was taken: messages. */
export const evWork = "work" satisfies EventKind;
/** A sapper dug a trench at cell: cells took trench records (sapper.ts). */
export const evDug = "dug" satisfies EventKind;
/** A sapper filled a trench at cell: cells took terrainFilled (sapper.ts). */
export const evFilled = "filled" satisfies EventKind;
/** A construction unit laid half a depot at cell, its anchor: cells (N, E, S, W) took the half depot's records (builder.ts). */
export const evHalfDepot = "halfDepot" satisfies EventKind;
/** A construction unit completed the depot at cell: cells took the depot's records and hold is its new container record (builder.ts). */
export const evDepot = "depot" satisfies EventKind;

/**
 * Event is one thing that happened, in order. Every field is there, with Go's zero value where Go leaves it
 * (route, messages and cells empty for Go's nil, outcome and end null); Go's JSON omits the empty ones.
 */
export interface Event {
  kind: EventKind;
  side: number;
  unit: number;
  cell: number;
  route: number[];
  outcome: Outcome | null;
  end: End | null;
  messages: Message[];
  phase: [number, number];
  turn: number;
  energy: number;
  hold: number;
  cells: number[];
}

/** newEvent is a Go Event literal: the fields not given are zero. */
export function newEvent(fields: Partial<Event> & { kind: EventKind }): Event {
  return {
    side: 0,
    unit: 0,
    cell: 0,
    route: [],
    outcome: null,
    end: null,
    messages: [],
    phase: [0, 0],
    turn: 0,
    energy: 0,
    hold: 0,
    cells: [],
    ...fields,
  };
}

/** gameUnitOf is the unit a command orders (Go's Game.unitOf); it throws a Refusal for an empty or unknown slot. */
export function gameUnitOf(g: Game, c: Command): number {
  if (c.kind !== cmdExit) {
    return c.unit ?? 0;
  }
  const slot = c.slot ?? 0;
  const h = g.holds.get(c.hold ?? 0);
  if (h === undefined || slot < 0 || slot >= 7) {
    throw refuse(0, "no such slot");
  }
  const u = at(h.slots, 7 * c.side + slot);
  if (u > slotFree) {
    throw refuse(0, "the slot is empty");
  }
  return u;
}

/**
 * check is the validation of a command, without changing the state; it returns the error (Go's error result),
 * null where the command is taken.
 *   - The side must not have confirmed (0661:3b81: a confirmed side's input only cancels the confirmation).
 *   - The unit must be one of the side's (0ca0:020c), not neutral (live +4 bit 2) and without the moved flag
 *     (0x200): one order per unit and phase (0ca0:0f59).
 *   - Moves and exits belong to the movement phase, attacks to the action phase (0ca0:0f59: the move symbol in
 *     phase 1, the fist in phase 2).
 *   - A human side's list must be shorter than the move limit (0ca0:1757, 0661:3349, message 0x53); a computer
 *     side has none.
 *   - An attack needs a unit in the target cell of the side's map (0ad1:000c, message 0x45) that is a target of
 *     the unit (target). Assumed, not verified: as for a move, the attacker must be on the side's map, not held
 *     in a container and the occupant of its cell. A sapper's order needs a neighbouring cell it can dig or fill
 *     instead (sapper.ts), a construction unit's a charge and a depot site (builder.ts).
 *   - A repair or a build needs an own building, the action phase and the energy (energy.ts).
 *
 * The move order code refuses a destination itself (evRefused), as it does in the game.
 */
export function check(g: Game, c: Command): GoError | null {
  if (c.side !== 0 && c.side !== 1) {
    return new GoError(`dosrules: side ${c.side}`);
  }
  const p = at(g.players, c.side);
  if (g.ended !== null) {
    return refuse(0, "the battle is over");
  }
  if (p.ready) {
    return refuse(0, "the side has confirmed its orders");
  }
  if (c.kind === cmdRepair || c.kind === cmdBuild) {
    return checkBuilding(g, c);
  }
  let idx: number;
  try {
    idx = gameUnitOf(g, c);
  } catch (e) {
    if (e instanceof GoError) {
      return e;
    }
    throw e;
  }
  const found = g.units.get(idx);
  if (found === undefined) {
    return refuse(0, "no such unit");
  }
  const u = copyLive(found);
  if ((u.flags4 & 2) !== 0 || !owns(u.flags4, c.side)) {
    return refuse(0, "not the side's unit");
  }
  if ((u.flags4 & 0x200) !== 0) {
    return refuse(0, "the unit has its order");
  }
  if ((u.flags4 & 0xc0) === 0x40) {
    // A two-cell ship's rear (+4 0x40 without the head's 0x80) moves and fights with its head.
    return refuse(0, "the ship's rear follows its head");
  }
  switch (c.kind) {
    case cmdMove:
    case cmdExit:
      if (p.phase !== phaseMovement) {
        return refuse(0, "moves are given in the movement phase");
      }
      if (
        c.kind === cmdMove &&
        ((u.flags4 & 0x4000) !== 0 ||
          occupant(g, c.side, liveCell(u, c.side)) !== idx)
      ) {
        return refuse(0, "the unit is not on the map");
      }
      break;
    case cmdAttack:
      if (p.phase !== phaseAction) {
        return refuse(0, "attacks are given in the action phase");
      }
      if (
        (u.flags4 & 0x4000) !== 0 ||
        occupant(g, c.side, liveCell(u, c.side)) !== idx
      ) {
        return refuse(0, "the unit is not on the map");
      }
      break;
    default:
      return new GoError(`dosrules: command ${JSON.stringify(c.kind)}`);
  }
  if (c.cell < 0 || c.cell >= g.width * g.height) {
    return new GoError(`dosrules: cell ${c.cell}`);
  }
  if (g.limit > 0 && !p.computer && p.orders.length >= g.limit) {
    return refuse(msgLimit, "the move limit is reached");
  }
  if (c.kind === cmdAttack && builder(u)) {
    return checkBuilder(g, c.side, u, idx, c.cell);
  }
  if (c.kind === cmdAttack && sapper(u)) {
    if (!trenchTarget(g, c.side, u, c.cell)) {
      return refuse(msgNoTarget, "no trench work there");
    }
  } else if (c.kind === cmdAttack) {
    const t = occupant(g, c.side, c.cell);
    const v = g.units.get(t);
    if (t > slotFree || v === undefined || !target(g, c.side, u, v)) {
      return refuse(msgNoTarget, "no target in the cell");
    }
  }
  return null;
}

/**
 * target reports whether the unit a of side may attack the unit v where the side's map shows them. This is an
 * assumed rule, not a verified one: v is not the side's, a's target flags cover a class of v's live word, and v
 * lies within the range of a's weapon against that class (the air range against aircraft, the surface range
 * otherwise; a range byte is one more than the distance it reaches, so 1 reaches nothing).
 */
export function target(
  g: Game,
  side: number,
  a: Readonly<Live>,
  v: Readonly<Live>,
): boolean {
  if (owns(v.flags4, side)) {
    return false;
  }
  const t = at(tablesOf(g.profile).types, a.type);
  if ((t.targetFlags & v.flags4 & classBits) === 0) {
    return false;
  }
  const [, rng] = weapon(t, v.flags4);
  const d = layoutDistance(
    g.profile,
    g.width,
    liveCell(a, side),
    liveCell(v, side),
  );
  return d >= 1 && d < rng;
}

/**
 * weaponAgainst is the weapon of type typ against a unit with the live word flags4 (Go's Profile.WeaponAgainst):
 * the air power and range against aircraft, the water power and the surface range against ships, the ground
 * power and the surface range otherwise (a range is encoded, one more than the distance it reaches).
 */
export function weaponAgainst(
  p: Profile,
  typ: number,
  flags4: number,
): [power: number, rng: number] {
  return weapon(at(tablesOf(p).types, typ), flags4);
}

/** occupant is the occupant byte of a cell of side's map. */
export function occupant(g: Game, side: number, cell: number): number {
  if (cell < 0 || cell >= g.width * g.height) {
    return 0xff;
  }
  return at(at(g.maps, side), 2 * cell + 1);
}

/**
 * apply checks a command and carries it out; it throws the error check returns.
 *   - A move or an exit runs at once, as in the game, where the unit moves during the phase while no other move
 *     may start (4c9e bit 0x80): the order code, the move start and the steps on the side's map (execute,
 *     moveExit). The move start appends the unit to the side's list.
 *   - An attack order (0ad1:000c, the attack kind) sets the moved flag (live +4 0x200), +0x11 to the target,
 *     +0x13 = 9 and +0x14 = 2, and appends the unit to the list (message 0x49). A sapper's order sets the moved
 *     flag, +0x11 and +0x13 = 0x11 and stays off the list (message 0x6e, evWork); so does a construction
 *     unit's (message 0x46, builder.ts).
 *
 * A repair or a build is carried out at once (energy.ts).
 *
 * Nothing withdraws an order once given: the cancel (10cc:0001) only clears the reach and selection marks and
 * undoes an exit in progress.
 */
export function apply(g: Game, c: Command): Event[] {
  const err = check(g, c);
  if (err !== null) {
    throw err;
  }
  switch (c.kind) {
    case cmdRepair:
      return repair(g, c);
    case cmdBuild:
      return buildUnit(g, c);
  }
  const unit = c.unit ?? 0;
  if (c.kind === cmdAttack && builder(liveOf(g, unit))) {
    return [orderDepot(g, c.side, unit, c.cell)];
  }
  if (c.kind === cmdAttack && sapper(liveOf(g, unit))) {
    const u = liveOf(g, unit);
    u.flags4 |= 0x200;
    u.target = 2 * c.cell;
    setAt(u.action, 0, actionTrench);
    g.units.set(unit, u);
    return [
      newEvent({
        kind: evWork,
        side: c.side,
        unit,
        cell: c.cell,
        messages: [{ code: msgTrench, side: c.side }],
      }),
    ];
  }
  if (c.kind === cmdAttack) {
    const u = liveOf(g, unit);
    u.flags4 |= 0x200;
    u.target = 2 * c.cell;
    setAt(u.action, 0, 9);
    setAt(u.action, 1, 2);
    g.units.set(unit, u);
    at(g.players, c.side).orders.push(unit);
    return [
      newEvent({
        kind: evAttack,
        side: c.side,
        unit,
        cell: c.cell,
        messages: [{ code: msgAttack, side: c.side }],
      }),
    ];
  }
  return gameMove(g, c);
}

/**
 * movement is the movement query of unit for side on the game's state (Go's Game.Movement): the side's map with
 * its units, the units held in containers, the budget of live byte 0 and the flag mask the player's order code
 * passes (0xffff for live +6 bit 1, else 6 for a side-1 unit and 3 for others). Its holds are the game's.
 */
export function movement(g: Game, side: number, unit: number): Movement {
  const mp = at(g.maps, side);
  const m = newMovement({
    width: g.width,
    height: g.height,
    terrain: [],
    holds: g.holds,
    unit,
    side,
  });
  const piece = (u: Readonly<Live>, cell: number): Piece => ({
    type: u.type,
    cell,
    flags4: u.flags4,
    flags6: u.flags6,
    hold: u.hold,
  });
  for (let c = 0; c < g.width * g.height; c++) {
    m.terrain.push(at(mp, 2 * c));
    const occ = at(mp, 2 * c + 1);
    if (occ <= slotFree) {
      const u = g.units.get(occ);
      if (u !== undefined) {
        m.units.set(occ, piece(u, c));
      }
    }
  }
  for (const i of sortedKeys(g.units)) {
    const u = liveOf(g, i);
    if ((u.flags4 & 0x4000) !== 0) {
      m.held.set(i, piece(u, liveCell(u, side)));
    }
  }
  const u = liveOf(g, unit);
  m.budget = u.budget;
  m.block = 3;
  if ((u.flags6 & 1) !== 0) {
    m.block = 0xffff;
  } else if ((u.flags4 & 1) !== 0) {
    m.block = 6;
  }
  return m;
}

/**
 * exitMovement is the movement query of the exit from slot of side's part of container record hold, as the
 * exit runs its reach search (moveExit): the unit in the slot at the record's first cell (one less for a
 * transporter, flag 0x20), budget 8 (4 for live +6 bit 0x10) and the flag mask 3 (6 for live +4 bit 1, with
 * 0x18 for live +6 bit 4). The unit is not on the side's map yet; its start cell is among the reached cells, and
 * the order code refuses it as a destination. It throws a GoError for no such slot, a Refusal for an empty one.
 */
export function exitMovement(
  g: Game,
  side: number,
  hold: number,
  slot: number,
): Movement {
  const h = g.holds.get(hold);
  if ((side !== 0 && side !== 1) || h === undefined || slot < 0 || slot >= 7) {
    throw new GoError(
      `dosrules: no slot ${slot} of side ${side} in container record ${hold}`,
    );
  }
  const u = at(h.slots, 7 * side + slot);
  if (u > slotFree) {
    throw refuse(0, "the slot is empty");
  }
  const [m] = exitQuery(movement(g, side, u), holdRecord(h), slot, (i) =>
    liveRecord(liveOf(g, i)),
  );
  return m;
}

/**
 * holdAt is the container record of a cell of side's map: the occupant's +0x0a where it is a transporter (+4
 * 0x1000), else the building record of the cell's terrain flag as the end-cell check finds it (0x400 and 0x100
 * by the cells the records list, 0x40 by flag bit 1). It reports false for a cell without one.
 */
export function holdAt(
  g: Game,
  side: number,
  cell: number,
): [rec: number, ok: boolean] {
  const t = moveTablesOrNull(g.profile);
  if (
    t === null ||
    (side !== 0 && side !== 1) ||
    cell < 0 ||
    cell >= g.width * g.height
  ) {
    return [0, false];
  }
  const u = g.units.get(occupant(g, side, cell));
  if (u !== undefined && (u.flags4 & 0x1000) !== 0) {
    return [u.hold, u.hold >= 0 && u.hold < holdRecords];
  }
  const tr = at(at(g.maps, side), 2 * cell);
  if (tr < 0 || tr >= t.terrains.length) {
    return [0, false];
  }
  const find = (base: number): [number, boolean] => {
    for (let i = 0; i < holdTable; i++) {
      const h = g.holds.get(base + i) ?? emptyHold();
      if (h.cells.includes(2 * cell)) {
        return [base + i, true];
      }
    }
    return [0, false];
  };
  const f = at(t.terrains, tr).flags;
  if ((f & 0x400) !== 0) {
    return find(holdFlag400);
  }
  if ((f & 0x100) !== 0) {
    return find(holdFlag100);
  }
  if ((f & 0x40) !== 0) {
    // Go's HoldFlag40 + f&1: & binds tighter than + there.
    return [holdFlag40 + (f & 1), true];
  }
  return [0, false];
}

/** moveTablesOrNull is Go's p.moveTables() where only its success counts: null for ErrNoMovement. */
export function moveTablesOrNull(p: Profile): MoveTables | null {
  return p.tables?.move ?? null;
}

/** bytesOf is a []int as bytes (Go's byte(x) of each). */
export function bytesOf(v: readonly number[]): Uint8Array {
  const b = new Uint8Array(v.length);
  for (const [i, x] of v.entries()) {
    b[i] = x & 0xff;
  }
  return b;
}

/**
 * gameMove runs a checked move or exit on the game's records and keeps what it leaves: the live records, the
 * side's map, the container records and the order list (Go's Game.move).
 */
export function gameMove(g: Game, c: Command): Event[] {
  const t = moveTablesOf(g.profile);
  const unit = gameUnitOf(g, c);
  const live = new Map<number, Uint8Array>();
  for (const [i, u] of g.units) {
    live.set(i, liveRecord(u));
  }
  const s = moveStateOn(
    t,
    movement(g, c.side, unit),
    live,
    [bytesOf(at(g.maps, c.side)), bytesOf(at(g.maps, 1 - c.side))],
    containerRecords(g.holds),
  );
  const o: MoveOrder = { goal: c.cell, retarget: -1 };
  let r: StepRun;
  if (c.kind === cmdExit) {
    r = exitOn(t, s, { ...o, hold: c.hold ?? 0, slot: c.slot ?? 0 });
  } else {
    r = executeOn(t, s, o);
  }
  for (const i of [...g.units.keys()]) {
    g.units.set(i, decodeLive(s.live.get(i) ?? new Uint8Array(0)));
  }
  const own = at(g.maps, c.side);
  for (const [i, b] of s.maps[0].entries()) {
    setAt(own, i, b);
  }
  const empty = emptyHold();
  for (const [i, b] of s.holds) {
    const h = decodeHold(b);
    if (g.holds.has(i) || !holdEqual(h, empty)) {
      g.holds.set(i, h);
    }
  }
  at(g.players, c.side).orders.push(...s.orders);
  const ev = newEvent({
    kind: evMoved,
    side: c.side,
    unit,
    cell: c.cell,
    route: [...r.cells],
    messages: messages(r.messages),
  });
  if ((r.order & 1) === 0 || (r.start & 1) === 0) {
    ev.kind = evRefused;
  }
  return [ev];
}

/** confirm is the double-arrow symbol: the side's orders are complete (+0x17 = 5, 0ca0:16b6). */
export function confirm(g: Game, side: number): Event[] {
  at(g.players, side).ready = true;
  return [newEvent({ kind: evReady, side })];
}

/**
 * unconfirm is the fire button on a confirmed side (0661:3b81): the side may give orders again (message 0x34).
 * A side that has not confirmed gets no event (Go's nil is empty).
 */
export function unconfirm(g: Game, side: number): Event[] {
  const p = at(g.players, side);
  if (!p.ready) {
    return [];
  }
  p.ready = false;
  return [
    newEvent({
      kind: evUnready,
      side,
      messages: [{ code: msgReadyUndone, side }],
    }),
  ];
}

/** action is the side in the action phase; it throws a GoError where the phases are no pair. */
export function action(g: Game): number {
  const [p0, p1] = g.players;
  if (p0.phase === phaseAction && p1.phase === phaseMovement) {
    return 0;
  }
  if (p1.phase === phaseAction && p0.phase === phaseMovement) {
    return 1;
  }
  throw new GoError(`dosrules: phases ${p0.phase}/${p1.phase}`);
}

/**
 * battlefield is the action side's map as the order handler fights on it: the units its map shows and the
 * combat class of every cell's terrain record (byte 5).
 */
export function battlefield(g: Game, side: number): Battlefield {
  const t = moveTablesOf(g.profile);
  const mp = at(g.maps, side);
  const b: Battlefield = {
    width: g.width,
    height: g.height,
    units: new Map(),
    terrain: new Map(),
    mapByte0: at(mp, 0),
  };
  for (let c = 0; c < g.width * g.height; c++) {
    const tr = at(mp, 2 * c);
    if (tr < t.terrains.length) {
      b.terrain.set(c, at(t.terrains, tr).class);
    }
    const occ = at(mp, 2 * c + 1);
    if (occ <= slotFree) {
      const u = g.units.get(occ);
      if (u !== undefined) {
        const x: Unit = {
          type: u.type,
          cell: c,
          size: u.size,
          alt: u.alt,
          exp: u.exp,
          flags4: u.flags4,
          flags6: u.flags6,
        };
        b.units.set(occ, x);
      }
    }
  }
  return b;
}

/**
 * resolve is the order handler once both sides have confirmed (F1, or two computer sides, 0661:3cfb). The
 * action side's list resolves on its map, in list order (resolveOrders): each attack against the cell in the
 * unit's +0x11, with the strengths and experience it leaves, and every destroyed unit taken off that map with
 * bit 4 of its +0x14 set (02dc:2083) and out of every container slot, a transporter's cargo destroyed with it
 * (vacate). The action pass then hands the buildings the movement side's units entered to that side (capture,
 * evCaptured) and carries out the sappers' work on the action side's map (trench, evDug and evFilled) and the
 * construction units' (builder.ts, evHalfDepot and evDepot). Then the phase change (changePhase) brings both
 * maps into line, marks the destroyed units 0x8000, swaps the phases, clears both lists and the confirmations
 * and counts the turn. It also clears every unit's +4 bits 0x0f00 (moved among them), its order cell +0x11 and
 * its action bytes +0x13 to +0x16 (02dc:1cf0, 1ebc), except that a unit that must set up (type +0x0f bit 0:
 * medium, heavy and train artillery) and arrived on the map this phase (+0x14 bit 0), not destroyed and not
 * held, keeps the moved flag: it gives no order in the next phase (02dc:1cf9-1d60). The side entering its action phase is then credited
 * its buildings' energy (evCredit). A destroyed attacker cannot have a later order: it is listed once, and only
 * return fire in its own battle destroys it. When the battle has ended (end), an evEnd event follows the phase
 * change and no further command or resolution is taken.
 *
 * It throws ErrOver after the end, ErrNotReady while a side has not confirmed, and the errors of the
 * resolution (Go returns the events so far beside them; the port drops them).
 */
export function resolve(g: Game): Event[] {
  if (g.ended !== null) {
    throw new ErrOver();
  }
  if (!g.players[0].ready || !g.players[1].ready) {
    throw new ErrNotReady();
  }
  const a = action(g);
  const evs = attack(g, a);
  const [caps, hq] = work(g, a);
  evs.push(...caps);
  const ev = gameChangePhase(g, a);
  evs.push(...ev);
  const ph = at(ev, 0);
  evs.push(...credit(g, ph));
  const e = end(g, hq);
  if (e.over) {
    g.ended = e;
    evs.push(
      newEvent({
        kind: evEnd,
        side: e.winner,
        end: e,
        phase: [ph.phase[0], ph.phase[1]],
        turn: g.turn,
      }),
    );
  }
  return evs;
}

/** attack resolves side a's list on its map. */
export function attack(g: Game, a: number): Event[] {
  const b = battlefield(g, a);
  const orders: Order[] = [];
  for (const u of at(g.players, a).orders) {
    orders.push({ attacker: u, target: div(liveOf(g, u).target, 2) });
  }
  const r = resolveOrders(g.profile, b, orders);
  const evs: Event[] = [];
  for (const [i, o] of r.orders.entries()) {
    const ord = at(orders, i);
    const ev = newEvent({
      kind: evSkipped,
      side: a,
      unit: ord.attacker,
      cell: ord.target,
    });
    if (o.battle) {
      ev.kind = evBattle;
      ev.outcome = o;
    }
    evs.push(ev);
  }
  const keys = sortedKeys(r.units);
  for (const i of keys) {
    const u = r.units.get(i);
    if (u === undefined) {
      continue;
    }
    const l = liveOf(g, i);
    l.size = u.size;
    l.alt = u.alt;
    l.exp = u.exp;
    g.units.set(i, l);
  }
  for (const i of keys) {
    const u = r.units.get(i);
    if (u === undefined) {
      continue;
    }
    // A ship half whose other half that map does not hold: it is one unit.
    const p = (u.flags4 & 0x80) !== 0 ? i + 1 : i - 1;
    if ((u.flags4 & 0x40) === 0 || r.units.has(p)) {
      continue;
    }
    const found = g.units.get(p);
    if (found !== undefined && (found.flags4 & 0x40) !== 0) {
      const l = copyLive(found);
      l.size = u.size;
      l.alt = u.alt;
      l.exp = u.exp;
      g.units.set(p, l);
      if (r.gone.includes(i)) {
        r.gone.push(p);
      }
    }
  }
  const mp = at(g.maps, a);
  for (const i of r.gone) {
    const l = liveOf(g, i);
    setAt(l.action, 1, at(l.action, 1) | 4);
    g.units.set(i, l);
    for (let c = 0; c < g.width * g.height; c++) {
      if (at(mp, 2 * c + 1) === i) {
        setAt(mp, 2 * c + 1, 0xff);
      }
    }
  }
  vacate(g, r.gone);
  return evs;
}

/**
 * vacate takes the units an attack list destroyed out of the container records. The battles hit a unit at the
 * cell the attacking side's map shows, so one that went into a building or onto a transporter during the
 * movement phase is destroyed while it holds a slot; the phase change only frees slots of units marked before
 * it, so the dead unit would stay listed (and repairable) for two phases. A destroyed transporter's cargo is
 * destroyed with it and its record cleared, as the destroy helper of a capture does (remove).
 */
export function vacate(g: Game, goneIn: readonly number[]): void {
  const dead = new Set<number>();
  const gone = [...goneIn];
  for (let next = gone.shift(); next !== undefined; next = gone.shift()) {
    const i = next;
    if (dead.has(i)) {
      continue;
    }
    dead.add(i);
    const u = liveOf(g, i);
    if ((u.flags4 & 0x4000) !== 0) {
      for (const r of sortedKeys(g.holds)) {
        const h = holdOf(g, r);
        if (!h.slots.includes(i)) {
          continue;
        }
        for (const [k, s] of h.slots.entries()) {
          if (s === i) {
            h.slots[k] = 0xff;
          }
        }
        g.holds.set(r, h);
      }
    }
    const found = g.holds.get(u.hold);
    if (
      (u.flags4 & 0x1000) === 0 ||
      found === undefined ||
      (found.flags & 0x8000) !== 0
    ) {
      continue;
    }
    const c = copyHold(found);
    for (const [k, s] of c.slots.entries()) {
      const v = g.units.get(s);
      if (v !== undefined && s <= slotFree && !dead.has(s)) {
        const w = copyLive(v);
        w.size = 0;
        setAt(w.action, 1, at(w.action, 1) | 4);
        g.units.set(s, w);
        gone.push(s);
      }
      c.slots[k] = 0xff;
    }
    c.flags = 0x8000;
    g.holds.set(u.hold, c);
  }
}

/** gameChangePhase is the phase change after side a's attacks (Go's Game.changePhase). */
export function gameChangePhase(g: Game, a: number): Event[] {
  const pc: PhaseChange = {
    width: g.width,
    height: g.height,
    maps: [g.maps[0], g.maps[1]],
    units: new Map(),
    holds: g.holds,
    attacker: a,
    phase: [g.players[0].phase, g.players[1].phase],
    count: g.turn,
  };
  for (const [i, u] of g.units) {
    const s: Sighted = {
      flags4: u.flags4,
      size: u.size,
      pos: [u.pos[0], u.pos[1]],
      facing: [u.facing[0], u.facing[1]],
      hold: u.hold,
    };
    pc.units.set(i, s);
  }
  const pr = changePhase(pc);
  const types = tablesOf(g.profile).types;
  for (const i of [...g.units.keys()]) {
    const d = decodeLive(pr.live.get(i) ?? new Uint8Array(0));
    const l = liveOf(g, i);
    l.flags4 = d.flags4 & 0xf0ff;
    l.size = d.size;
    l.pos = d.pos;
    l.facing = d.facing;
    if (
      (l.flags4 & 0xc000) === 0 &&
      (at(l.action, 1) & 1) !== 0 &&
      setsUp(at(types, l.type))
    ) {
      l.flags4 |= 0x200; // 02dc:1cf9-1d60
    }
    l.target = 0;
    for (let k = 0; k < 4; k++) {
      setAt(l.action, k, 0);
    }
    g.units.set(i, l);
  }
  for (let k = 0; k < 2; k++) {
    setAt(g.maps, k, Array.from(at(pr.maps, k)));
  }
  for (const [i, h] of changed(containerRecords(g.holds), pr.holds)) {
    g.holds.set(i, h);
  }
  for (let k = 0; k < 2; k++) {
    const p = at(g.players, k);
    p.phase = at(at(pr.players, k), 0x16);
    p.ready = false;
    p.orders = [];
  }
  g.turn = pr.count;
  return [
    newEvent({
      kind: evPhase,
      side: a,
      messages: messages(pr.messages),
      phase: [g.players[0].phase, g.players[1].phase],
      turn: g.turn,
    }),
  ];
}

/** gameDestroyed are the units the last resolutions destroyed and the phase change marked, ascending (Go's Game.Destroyed). */
export function gameDestroyed(g: Game): number[] {
  const out: number[] = [];
  for (const [i, u] of g.units) {
    if (destroyed(u)) {
      out.push(i);
    }
  }
  return out.sort((a, b) => a - b);
}
