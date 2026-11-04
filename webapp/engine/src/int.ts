// Go's fixed-width integers and integer division on JS numbers. The rules engine computes in Go's 64-bit int,
// which plain numbers hold exactly below 2^53; the narrower types of the original's records wrap. Every helper
// takes a safe integer: the bitwise wraps reduce it modulo 2^32 first, which keeps the low bits exact.
//
// None of them returns -0 (Go has no negative zero, and Vitest's toStrictEqual tells -0 from 0): JS gives -0
// for Math.trunc(-0.5), -4 % 2 and -x of 0, so the port divides with div and rem and negates as 0 - x.

/** int8 is Go's int8(x): the low byte, signed. */
export function int8(x: number): number {
  return (x << 24) >> 24;
}

/** uint8 is Go's uint8(x) (byte): the low byte. */
export function uint8(x: number): number {
  return x & 0xff;
}

/** int16 is Go's int16(x): the low word, signed. */
export function int16(x: number): number {
  return (x << 16) >> 16;
}

/** uint16 is Go's uint16(x): the low word. */
export function uint16(x: number): number {
  return x & 0xffff;
}

/** int32 is Go's int32(x): the low 32 bits, signed. */
export function int32(x: number): number {
  return x | 0;
}

/** uint32 is Go's uint32(x): the low 32 bits. */
export function uint32(x: number): number {
  return x >>> 0;
}

/**
 * mul32 is int32(a*b) for any safe integers a and b, exact where a*b itself would leave the safe range: the
 * low 32 bits of a product depend only on the low 32 bits of its factors.
 */
export function mul32(a: number, b: number): number {
  return Math.imul(a, b);
}

/** IntegerDivideError is Go's runtime panic on an integer division by zero. */
export class IntegerDivideError extends RangeError {
  constructor() {
    super("runtime error: integer divide by zero");
    this.name = "IntegerDivideError";
  }
}

/** div is Go's integer a / b: truncated toward zero. It panics (throws) on a zero divisor. */
export function div(a: number, b: number): number {
  if (b === 0) {
    throw new IntegerDivideError();
  }
  const q = Math.trunc(a / b);
  return q === 0 ? 0 : q;
}

/** rem is Go's integer a % b: the sign of a. It panics (throws) on a zero divisor. */
export function rem(a: number, b: number): number {
  if (b === 0) {
    throw new IntegerDivideError();
  }
  const r = a % b;
  return r === 0 ? 0 : r;
}
