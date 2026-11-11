import type { Event as RuleEvent } from "./cycle.ts";
import { Refusal as RuleRefusal } from "./cycle.ts";
import { GoError } from "./goruntime.ts";
import { liveCell } from "./board.ts";
import * as rules from "./cycle.ts";
import { coord as boundCoord } from "./bind.ts";
import type { Coord } from "./hexgrid.ts";
import {
  type Session,
  type Event,
  type UnitView,
  type Fighter,
  newEvent,
  hex,
  Moved,
  Refused,
  Ordered,
  Confirmed,
  Unconfirmed,
  Battle,
  Skipped,
  PhaseChanged,
  Credited,
  Repaired,
  Built,
  Ended,
  Captured,
  Tasked,
  Dug,
  Constructed,
  Released,
  Gassed,
  Cleared,
  Wired,
  Tunnelled,
  TunnelLost,
  Blown,
  Mined,
  Healed,
} from "./playModel.ts";
import { kindOf, slots, typeName, unitID } from "./playView.ts";

const messages: Readonly<Record<number, string>> = {
  0: "",
  0x34: "Orders open again.",
  0x35: "Movement phase.",
  0x36: "Action phase.",
  0x37: "An enemy unit holds that hex.",
  0x38: "Move complete.",
  0x39: "A friendly unit holds that hex.",
  0x3a: "That transport cannot take this unit now.",
  0x3b: "There is no room left there.",
  0x40: "The unit cannot stop on that terrain.",
  0x45: "No target there.",
  0x48: "No route leads there.",
  0x49: "Attack ordered.",
  0x53: "No orders left this phase.",
  0x6e: "Trench work ordered.",
  0x46: "Depot construction ordered.",
  0x47: "No more depots can be built.",
  0x4b: "There is no room for a depot here.",
  0x4c: "The unit has nothing left to build with.",
  0x100: "Gas release ordered.",
  0x101: "The unit has no gas left.",
  0x102: "Wire laying ordered.",
  0x103: "The unit has no wire left.",
  0x104: "Tunnelling ordered.",
  0x105: "The tunnellers are digging a mine already.",
};
/** levelHexes are the level hexes of cells, leaving out cells off the level. */
export function levelHexes(s: Session, cells: readonly number[]): Coord[] {
  return cells.flatMap((c) => {
    const p = boundCoord(s.board, c);
    return p === null ? [] : [p];
  });
}
export function messageText(code: number): string {
  return messages[code] ?? `Message 0x${code.toString(16)}.`;
}
export class OrderError extends GoError {
  readonly refusal: RuleRefusal;
  constructor(r: RuleRefusal) {
    super(
      r.code === 0 && r.reason !== ""
        ? r.reason.charAt(0).toUpperCase() + r.reason.slice(1) + "."
        : messageText(r.code),
      { cause: r },
    );
    this.refusal = r;
  }
}
export function wrap(e: unknown): unknown {
  return e instanceof RuleRefusal ? new OrderError(e) : e;
}
/** The events before a refused command accompany the error, like Go's (events, error). */
export class ReplayError extends GoError {
  readonly events: Event[];
  constructor(index: number, kind: string, cause: GoError, events: Event[]) {
    super(`dosplay: replay of command ${index} (${kind}): ${cause.message}`, {
      cause,
    });
    this.events = events;
  }
}
export function refusal(reason: string): OrderError {
  return new OrderError(new RuleRefusal(0, reason));
}
function text(r: RuleEvent): string {
  return messageText(r.messages.find((m) => m.side === r.side)?.code ?? 0);
}
export interface Battles {
  views: Map<number, UnitView>;
  state: Map<number, [number, number, number]>;
  half: Map<number, number>;
}
function fighter(
  b: Battles,
  idx: number,
  size: number,
  alt: number,
  exp: number,
  gone: number[],
): Fighter {
  const st = b.state.get(idx) ?? [0, 0, 0],
    v = b.views.get(idx);
  if (v === undefined) throw new GoError(`dosplay: battle unit ${idx}`);
  const unit = { ...v, size: st[0], alt: st[1], exp: st[2] };
  const f: Fighter = {
    unit,
    before: st[0],
    after: size,
    altBefore: st[1],
    altAfter: alt,
    expBefore: st[2],
    expAfter: exp,
    destroyed: gone.includes(idx),
  };
  b.state.set(idx, [size, alt, exp]);
  const other = b.half.get(idx);
  if (other !== undefined) b.state.set(other, [size, alt, exp]);
  return f;
}
export function moved(s: Session, r: RuleEvent, start: number): Event {
  const ev = newEvent({
    kind: Moved,
    side: r.side,
    unit: r.unit,
    phase: r.phase,
    turn: r.turn,
    raw: r,
    to: hex(s, r.cell),
  });
  if (r.kind === "refused" || r.route.length === 0) {
    ev.kind = Refused;
    ev.text = text(r) || "The unit cannot go there.";
    if (start >= 0) ev.from = hex(s, start);
    return ev;
  }
  ev.path = [
    ...(start >= 0 ? [hex(s, start)] : []),
    ...r.route.slice(0, -1).map((c) => hex(s, c)),
    ev.to,
  ];
  ev.from = ev.path[0] ?? { line: 0, col: 0 };
  return ev;
}
export function event(
  s: Session,
  r: RuleEvent,
  b: Battles | null = null,
): Event {
  const ev = newEvent({
    side: r.side,
    unit: r.unit,
    phase: r.phase,
    turn: r.turn,
    energy: r.energy,
    end: r.end,
    raw: r,
  });
  switch (r.kind) {
    case "attack":
      ev.kind = Ordered;
      ev.to = hex(s, r.cell);
      ev.text = text(r);
      ev.from = hex(s, liveCell(rules.liveOf(s.game, r.unit), r.side));
      break;
    case "ready":
      ev.kind = Confirmed;
      ev.text = "Orders confirmed.";
      break;
    case "unready":
      ev.kind = Unconfirmed;
      ev.text = text(r);
      break;
    case "battle": {
      ev.kind = Battle;
      const o = r.outcome;
      if (o === null) break;
      if (b !== null) {
        const attacker = fighter(
            b,
            r.unit,
            o.attackerSize,
            o.attackerAlt,
            o.attackerExp,
            o.destroyed,
          ),
          defender = fighter(
            b,
            o.defender,
            o.defenderSize,
            o.defenderAlt,
            o.defenderExp,
            o.destroyed,
          );
        ev.battle = {
          attacker,
          defender,
          returnFire: o.returned,
          pincer: o.pincer,
          blocking: o.blocking,
        };
        ev.from = attacker.unit.at;
        ev.to = defender.unit.at;
      }
      break;
    }
    case "skipped":
      ev.kind = Skipped;
      ev.to = hex(s, r.cell);
      ev.from = b?.views.get(r.unit)?.at ?? ev.from;
      break;
    case "phase":
      ev.kind = PhaseChanged;
      ev.text = `Turn ${r.turn}.`;
      break;
    case "credit":
      ev.kind = Credited;
      break;
    case "repaired":
      ev.kind = Repaired;
      ev.to = hex(s, r.cell);
      break;
    case "built":
      ev.kind = Built;
      ev.to = hex(s, r.cell);
      break;
    case "end":
      ev.kind = Ended;
      break;
    case "captured": {
      ev.kind = Captured;
      ev.to = hex(s, r.cell);
      ev.hold = r.hold;
      const found = slots(s, r.hold, r.side).length - 1,
        t =
          ["HQ captured", "Depot captured", "Factory captured"][
            kindOf(r.hold)
          ] ?? "";
      ev.text =
        t +
        (found === 1
          ? ", 1 unit found."
          : found > 1
            ? `, ${found} units found.`
            : ".");
      break;
    }
    case "work":
      ev.kind = Tasked;
      ev.to = hex(s, r.cell);
      ev.text = text(r);
      ev.from = hex(s, liveCell(rules.liveOf(s.game, r.unit), r.side));
      break;
    case "dug":
    case "filled":
    case "halfDepot":
    case "depot":
      ev.kind = r.kind === "dug" || r.kind === "filled" ? Dug : Constructed;
      ev.to = hex(s, r.cell);
      ev.text =
        r.kind === "dug"
          ? "Trench dug."
          : r.kind === "filled"
            ? "Trench filled."
            : r.kind === "halfDepot"
              ? "Half a depot built."
              : "Depot built.";
      if (r.kind === "depot") ev.hold = r.hold;
      ev.from = hex(s, liveCell(rules.liveOf(s.game, r.unit), r.side));
      ev.hexes = r.cells.flatMap((c) => {
        const p = boundCoord(s.board, c);
        return p === null ? [] : [p];
      });
      break;
    case "gasCloud":
      ev.kind = Released;
      ev.to = hex(s, r.cell);
      ev.text = "Gas released.";
      ev.from = hex(s, liveCell(rules.liveOf(s.game, r.unit), r.side));
      ev.hexes = levelHexes(s, r.cells);
      break;
    case "gassed": {
      ev.kind = Gassed;
      ev.to = hex(s, r.cell);
      ev.loss = r.loss;
      const u = rules.liveOf(s.game, r.unit),
        name = typeName(s, unitID(s, r.unit), u.type);
      ev.text =
        u.size === 0
          ? `${name} destroyed by gas.`
          : `${name} lose ${String(r.loss)} to gas.`;
      break;
    }
    case "gasCleared":
      ev.kind = Cleared;
      ev.hexes = levelHexes(s, r.cells);
      ev.text = "The gas has cleared.";
      break;
    case "wired":
      ev.kind = Wired;
      ev.to = hex(s, r.cell);
      ev.hexes = levelHexes(s, r.cells);
      ev.text = "Wire laid.";
      ev.from = hex(s, liveCell(rules.liveOf(s.game, r.unit), r.side));
      break;
    case "tunnel":
      ev.kind = Tunnelled;
      ev.to = hex(s, r.cell);
      ev.text = "Tunnelling begun.";
      ev.from = hex(s, liveCell(rules.liveOf(s.game, r.unit), r.side));
      break;
    case "tunnelLost":
      ev.kind = TunnelLost;
      ev.to = hex(s, r.cell);
      ev.text = "The tunnel was abandoned.";
      break;
    case "mine":
      ev.kind = Blown;
      ev.to = hex(s, r.cell);
      ev.hexes = levelHexes(s, r.cells);
      ev.text = "A mine went up.";
      break;
    case "mined": {
      ev.kind = Mined;
      ev.to = hex(s, r.cell);
      ev.loss = r.loss;
      const u = rules.liveOf(s.game, r.unit),
        name = typeName(s, unitID(s, r.unit), u.type);
      ev.text =
        u.size === 0
          ? `${name} destroyed by a mine.`
          : `${name} lose ${String(r.loss)} to a mine.`;
      break;
    }
    case "healed": {
      ev.kind = Healed;
      ev.to = hex(s, r.cell);
      ev.gain = r.gain;
      const u = rules.liveOf(s.game, r.unit);
      ev.text = `${typeName(s, unitID(s, r.unit), u.type)} regain ${String(r.gain)}.`;
      break;
    }
    case "moved":
    case "refused":
      return moved(s, r, -1);
  }
  return ev;
}
