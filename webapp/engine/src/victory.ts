// The end of a battle and its score. A side wins by capturing the other side's HQ (capture.ts), by destroying all
// of its units, or when the other side surrenders (surrender); the capture wins over an annihilation in the same
// resolution (0661:3d03); the score starts as the defence values of all units and falls as units disappear. The
// winner's final score adds 500 for annihilation and 100 for concealed inventories, multiplies by 4, 3 or 2 for a
// move limit of 4, 8 or 16 and is capped at 32,500.
//
// Assumed, not verified:
//   - A unit's defence value is its type's armour; the score counts every unit not destroyed, of either side or
//     neutral, in containers too.
//   - A side is annihilated when none of its units is left.
//   - Both sides annihilated at once is a draw without a score.
//   - The bonuses are added before the limit's factor.

import { destroyed } from "./board.ts";
import { type Event, type Game, evEnd, newEvent } from "./cycle.ts";
import { at, setAt } from "./goruntime.ts";
import { owns } from "./move.ts";
import { tablesOf } from "./profile.ts";

/** MaxScore caps the final score. */
export const maxScore = 32500;

/** EndReason is how a side won; "" where it did not. */
export type EndReason = "" | "hq" | "annihilation" | "surrender";

/** The ways to win. */
export const endHQ = "hq" satisfies EndReason;
export const endAnnihilation = "annihilation" satisfies EndReason;
export const endSurrender = "surrender" satisfies EndReason;

/**
 * End is the state of a battle's end: whether it is over, the winner (-1 while it is open or after a draw), how
 * each side won, and the winner's final score.
 */
export interface End {
  over: boolean;
  winner: number;
  won: [EndReason, EndReason];
  score: number;
}

/** score is the current score: the armour of every unit not destroyed (Go's Game.Score). */
export function score(g: Game): number {
  const types = tablesOf(g.profile).types;
  let s = 0;
  for (const u of g.units.values()) {
    if (!destroyed(u)) {
      s += at(types, u.type).armour;
    }
  }
  return s;
}

/** end is the battle's end as the state stands; resolve decides it, with hq the side that captured an HQ in the resolution (-1 none). */
export function end(g: Game, hq: number): End {
  const r: End = { over: false, winner: -1, won: ["", ""], score: 0 };
  for (let k = 0; k < 2; k++) {
    if (annihilated(g, 1 - k)) {
      setAt(r.won, k, endAnnihilation);
    }
  }
  if (hq >= 0) {
    r.won = ["", ""];
    setAt(r.won, hq, endHQ);
  }
  if (r.won[0] !== "" && r.won[1] !== "") {
    r.over = true;
  } else if (r.won[0] !== "" || r.won[1] !== "") {
    r.over = true;
    r.winner = r.won[0] === "" ? 1 : 0;
    r.score = finalScore(
      score(g),
      at(r.won, r.winner) === endAnnihilation,
      g.concealed,
      g.limit,
    );
  }
  return r;
}

/**
 * surrender ends the battle at once: side gives up and the other side wins by surrender at the current score,
 * without the annihilation bonus. It returns the end event, or none (Go's nil is empty) while the battle is
 * already over or the side is not one of the two.
 */
export function surrender(g: Game, side: number): Event[] {
  if (g.ended !== null || (side !== 0 && side !== 1)) {
    return [];
  }
  const e: End = { over: false, winner: 1 - side, won: ["", ""], score: 0 };
  setAt(e.won, e.winner, endSurrender);
  e.over = true;
  e.score = finalScore(score(g), false, g.concealed, g.limit);
  g.ended = e;
  return [
    newEvent({
      kind: evEnd,
      side: e.winner,
      end: e,
      phase: [g.players[0].phase, g.players[1].phase],
      turn: g.turn,
    }),
  ];
}

/** annihilated reports whether side has no unit left. */
export function annihilated(g: Game, side: number): boolean {
  for (const u of g.units.values()) {
    if (owns(u.flags4, side) && !destroyed(u)) {
      return false;
    }
  }
  return true;
}

/** finalScore is the winner's score from the current one. */
export function finalScore(
  s: number,
  annihilation: boolean,
  concealed: boolean,
  limit: number,
): number {
  if (annihilation) {
    s += 500;
  }
  if (concealed) {
    s += 100;
  }
  switch (limit) {
    case 4:
      s *= 4;
      break;
    case 8:
      s *= 3;
      break;
    case 16:
      s *= 2;
      break;
  }
  return Math.min(s, maxScore);
}
