import type { BindingData, Board } from "./bind.ts";
import type {
  Game,
  Command as RuleCommand,
  Event as RuleEvent,
} from "./cycle.ts";
import { newEvent as newRuleEvent } from "./cycle.ts";
import type { Coord } from "./hexgrid.ts";
import type { Level } from "./level.ts";
import type { Profile } from "./profile.ts";
import type { End } from "./victory.ts";

export const Moved = 0,
  Refused = 1,
  Ordered = 2,
  Confirmed = 3,
  Unconfirmed = 4,
  Battle = 5,
  Skipped = 6,
  Revealed = 7,
  PhaseChanged = 8,
  Credited = 9,
  Repaired = 10,
  Built = 11,
  Ended = 12,
  Captured = 13,
  Tasked = 14,
  Dug = 15,
  Constructed = 16,
  /** Gas pioneers released a cloud: from the pioneers, to the hex, hexes the cloud, text. */
  Released = 17,
  /** A unit lost strength to gas: unit, to its hex, loss, text. */
  Gassed = 18,
  /** The gas left hexes: hexes, text. */
  Cleared = 19;
export const Infantry = 0,
  Armour = 1,
  Artillery = 2,
  Naval = 3,
  Air = 4;
export const HQ = 0,
  Depot = 1,
  Factory = 2;
export type Command = Omit<RuleCommand, "kind"> & {
  kind: RuleCommand["kind"] | "confirm" | "unconfirm" | "resolve" | "surrender";
};
/**
 * Options are a session's game options. profile plays the original units; extended (dosrules' Extended, the
 * full profile on the extension's tables) binds every level and plays a level that places or builds an
 * extension unit, or every level with extendedUnits, which also brings the extension units into every factory
 * (docs/new-units.md).
 */
export interface Options {
  profile: Profile;
  limit?: number;
  computer?: [boolean, boolean];
  extended?: Profile;
  extendedUnits?: boolean;
}
export interface Session {
  game: Game;
  board: Board;
  level: Level;
  data: BindingData;
  version: number;
  routes: [Event[], Event[]];
  ids: Map<number, string>;
  log: Command[];
  light: number[];
  lightWater: number;
  lightDone: boolean;
}
export interface UnitView {
  index: number;
  type: number;
  id: string;
  name: string;
  domain: number;
  side: number;
  owner: number;
  at: Coord;
  size: number;
  full: number;
  alt: number;
  exp: number;
  budget: number;
  armour: number;
  moved: boolean;
  held: boolean;
  carrier: boolean;
  destroyed: boolean;
  hold: number;
  in: number;
  slot: number;
  target: Coord | null;
  rear: Coord | null;
  head: Coord | null;
}
export interface BuildingView {
  hold: number;
  owner: number;
  kind: number;
  hq: boolean;
  units: UnitView[];
}
export interface View {
  side: number;
  units: Map<string, UnitView>;
  buildings: Map<string, BuildingView>;
}
export interface SlotView {
  slot: number;
  unit: UnitView | null;
}
export interface Fighter {
  unit: UnitView;
  before: number;
  after: number;
  expBefore: number;
  expAfter: number;
  destroyed: boolean;
  altBefore: number;
  altAfter: number;
}
export interface BattleInfo {
  attacker: Fighter;
  defender: Fighter;
  returnFire: boolean;
  pincer: number;
  blocking: number;
}
export type RawEvent = Omit<RuleEvent, "kind"> & {
  kind: RuleEvent["kind"] | "";
};
export function zeroRawEvent(): RawEvent {
  return { ...newRuleEvent({ kind: "moved" }), kind: "" };
}
export interface Event {
  kind: number;
  side: number;
  unit: number;
  path: Coord[];
  from: Coord;
  to: Coord;
  battle: BattleInfo | null;
  phase: [number, number];
  turn: number;
  energy: number;
  end: End | null;
  hold: number;
  text: string;
  raw: RawEvent;
  rear: Coord[];
  hexes: Coord[];
  /** The strength a unit lost to gas (Gassed). */
  loss: number;
}
export function newEvent(fields: Partial<Event> = {}): Event {
  return {
    kind: Moved,
    side: 0,
    unit: 0,
    path: [],
    from: { line: 0, col: 0 },
    to: { line: 0, col: 0 },
    battle: null,
    phase: [0, 0],
    turn: 0,
    energy: 0,
    end: null,
    hold: 0,
    text: "",
    raw: zeroRawEvent(),
    rear: [],
    hexes: [],
    loss: 0,
    ...fields,
  };
}
/** A Go Event struct copy: its arrays and coordinates are values, its slice backing arrays remain shared. */
export function copyEvent(e: Event): Event {
  return {
    ...e,
    from: { ...e.from },
    to: { ...e.to },
    phase: [...e.phase],
    raw: { ...e.raw, phase: [...e.raw.phase] },
  };
}
export function flagSide(flags: number): number {
  return flags & 1 ? 1 : flags & 2 ? 2 : 0;
}
export function faction(side: number): number {
  return [1, 2, 0][side] ?? 0;
}
export function sideOf(player: number): number {
  return player === 1 ? 0 : player === 2 ? 1 : 2;
}
export function coordFromKey(key: string): Coord {
  const [line, col] = key.split(",").map(Number);
  return { line: line ?? 0, col: col ?? 0 };
}
export function coords(set: ReadonlySet<string>): Coord[] {
  return [...set]
    .map(coordFromKey)
    .sort((a, b) => a.line - b.line || a.col - b.col);
}
/** The cell-to-hex conversion also retains the coordinate of omitted map-edge cells, like Go Board.Coord. */
export function hex(s: Session, cell: number): Coord {
  if (cell < 0 || cell >= s.board.width * s.board.height)
    return { line: 0, col: 0 };
  const col = cell % s.board.width;
  return {
    line: 2 * Math.trunc(cell / s.board.width) + (col % 2),
    col: Math.trunc(col / 2),
  };
}
