import { GoError, at } from "./goruntime.ts";
import { int16, int32, mul32 } from "./int.ts";
import { type Profile, type Type, alternate, tablesOf } from "./profile.ts";

/**
 * ErrDivideError is the 8086 divide error the casualty routine raises on a zero divisor (a zero-armour unit);
 * the game would stop with a runtime error.
 */
export class ErrDivideError extends GoError {
  constructor() {
    super("divide error");
  }
}

// Target-flag and live-flag bits the casualty routine reads.
/** Target classes in TargetFlags, AttackMask and the live word. */
export const classBits = 0x3c;
/** Live word of the target: a boat. */
export const waterTarget = 0x08;
/** Live word of the target: an aircraft. */
export const airTarget = 0x10;
/** TargetFlags: never supports, never returns fire. */
export const longRange = 0x40;
/** TargetFlags: never returns fire. */
export const noReturn = 0x80;
/** Live +6 of the attacker: a flamethrower (extension types only, wire.ts), the source of Context.flame. */
export const flagFlame = 0x0800;

/** Fighter is one side of a battle: its type and its live values. */
export interface Fighter {
  type: Type;
  /** Live group strength. */
  size: number;
  /** Live alternate strength. */
  alt: number;
  /** Experience, 0-6 in the game (read as an unsigned byte). */
  exp: number;
  /** Live flag word +4: side bits 1/2, target class bits. */
  flags: number;
}

/** Context are the battle inputs the handler supplies. */
export interface Context {
  distance: number;
  /** Terrain classes 0-8. */
  aTerrain: number;
  dTerrain: number;
  /** Support percentages (pincer a byte, blocking at most 150). */
  pincer: number;
  blocking: number;
  returnFire: boolean;
  /** The attacker is a flamethrower (an extension type, wire.ts), which the defender's terrain does not shelter from. */
  flame: boolean;
}

/** Result holds the strengths after the battle. */
export interface Result {
  attackerSize: number;
  attackerAlt: number;
  defenderSize: number;
  defenderAlt: number;
}

// The routine's arithmetic is the 32-bit runtime helpers' on Go int64 values; every operand stays a safe
// integer, products go through mul32 (int32 of the full product) and sums through int32.

/** quo is the runtime's signed 32-bit division, truncating toward zero; a zero divisor raises the divide error. */
export function quo(a: number, b: number): number {
  if (b === 0) {
    throw new ErrDivideError();
  }
  return int32(Math.trunc(a / b));
}

/** percent raises v by pct percent in 16.16 fixed point and keeps the signed high word: 20 raised by 150 is 49, not 50. */
export function percent(v: number, pct: number): number {
  const fx = (v & 0xffff) << 16;
  return int16(int32(fx + mul32(quo(fx, 100), pct)) >> 16);
}

/** strengthTerm is base·size + base·F·size/div + (base/2)·term, F = (exp<<8)/6 + 1. */
export function strengthTerm(
  base: number,
  size: number,
  exp: number,
  term: number,
  div: number,
): number {
  const f = quo(exp << 8, 6) + 1;
  return int32(
    mul32(base, size) +
      quo(mul32(mul32(base, f), size), div) +
      mul32(quo(base, 2), term),
  );
}

/** weapon is the power and encoded range a type uses against a target with the live flag word flags. */
export function weapon(t: Type, flags: number): [power: number, rng: number] {
  if ((flags & airTarget) !== 0) {
    return [t.airPower, t.airRange];
  }
  if ((flags & waterTarget) !== 0) {
    return [t.waterPower, t.surfaceRange];
  }
  return [t.groundPower, t.surfaceRange];
}

/** groupSize is the strength that fights: the alternate strength for such types, clamped to 1..6. */
function groupSize(f: Fighter): number {
  const n = alternate(f.type) ? f.alt : f.size;
  return Math.min(Math.max(n, 1), 6);
}

/**
 * firesBack reports whether a defender of type d returns the fire of an attacker of type a: only when the
 * context allows it (next to the attacker), its type neither never returns fire nor is long range, and the
 * attacker's classes include it (2451:0037-009e in the demo).
 */
export function firesBack(a: Type, d: Type, returnFire: boolean): boolean {
  return (
    returnFire &&
    (d.targetFlags & (noReturn | longRange)) === 0 &&
    (a.attackMask & d.targetFlags & classBits) !== 0
  );
}

/**
 * casualties resolves one battle as the casualty routine (2451:0004 in the demo) does (Go's
 * Profile.Casualties). The routine makes no random draw. The defender fires back unless the context forbids
 * it, its type never returns fire or is long range, or the attacker's classes do not include it. A zero
 * divisor throws ErrDivideError (Go recovers its panic into the returned error).
 */
export function casualties(
  p: Profile,
  a: Fighter,
  d: Fighter,
  c: Context,
): Result {
  const terms = tablesOf(p).terms;
  const suppressed = !firesBack(a.type, d.type, c.returnFire);

  const sizeA = groupSize(a);
  const sizeD = groupSize(d);
  const [aPower, rng] = weapon(a.type, d.flags);
  let power = aPower;
  if (rng > 1) {
    power = int32(power - quo(mul32(3, mul32(power, c.distance - 1)), 4 * rng));
  }
  const [counter] = weapon(d.type, a.flags);
  const expA = a.exp + 1;
  const expD = d.exp + 1;
  const armourA = a.type.armour;
  let armourD = d.type.armour;

  power = percent(power, c.pincer);
  armourD = percent(armourD, c.blocking);

  const aTerms = at(terms, c.aTerrain);
  let dTerms = at(terms, c.dTerrain);
  if (c.flame && dTerms.defence > 0) {
    dTerms = { ...dTerms, defence: 0 };
  }
  const attackA = strengthTerm(power, sizeA, expA, aTerms.attack, 256);
  const defenceA = strengthTerm(armourA, sizeA, expA, aTerms.defence, 256);
  const attackD = strengthTerm(counter, sizeD, expD, dTerms.attack, 256);
  const defenceD = strengthTerm(armourD, sizeD, expD, dTerms.defence, 512);

  // Experience parameters: the attacker's is E, 2 below 1 (which a byte cannot reach) and 6 above 7; the
  // defender's is 2 up to 1 and 6 from 7.
  let pa = expA;
  let pd = expD;
  if (expA < 1) {
    pa = 2;
  } else if (expA > 7) {
    pa = 6;
  }
  if (expD <= 1) {
    pd = 2;
  } else if (expD >= 7) {
    pd = 6;
  }
  let lossD = quo(int32(attackA - defenceD), quo(mul32(armourD, 8 - pa), 2));
  let lossA = quo(int32(attackD - defenceA), quo(mul32(armourA, 8 - pd), 2));

  // At least one loss when the attack beats the per-unit defence and half the armour; 7-11 losses become 6;
  // never more than twice the opponent's strength, never below zero or above the own strength.
  if (lossD <= 0 && int32(quo(defenceD, sizeD) + quo(armourD, 2)) < attackA) {
    lossD = 1;
  }
  if (lossA <= 0 && int32(quo(defenceA, sizeA) + quo(armourA, 2)) < attackD) {
    lossA = 1;
  }
  if (lossD > 6 && lossD < 12) {
    lossD = 6;
  }
  if (lossA > 6 && lossA < 12) {
    lossA = 6;
  }
  lossD = Math.min(Math.max(Math.min(lossD, 2 * sizeA), 0), sizeD);
  lossA = Math.min(Math.max(Math.min(lossA, 2 * sizeD), 0), sizeA);

  const res: Result = {
    attackerSize: a.size,
    attackerAlt: a.alt,
    defenderSize: d.size,
    defenderAlt: d.alt,
  };
  if (!suppressed) {
    [res.attackerSize, res.attackerAlt] = writeBack(a, sizeA - lossA);
  }
  [res.defenderSize, res.defenderAlt] = writeBack(d, sizeD - lossD);
  return res;
}

/**
 * writeBack stores the remaining strength as a byte: into the alternate strength for such types, where zero
 * also clears the group strength.
 */
function writeBack(f: Fighter, left: number): [size: number, alt: number] {
  const v = left & 0xff;
  if (!alternate(f.type)) {
    return [v, f.alt];
  }
  if (v === 0) {
    return [0, 0];
  }
  return [f.size, v];
}
