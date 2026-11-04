import { describe, expect, it } from "vitest";
import {
  IntegerDivideError,
  div,
  int16,
  int32,
  int8,
  mul32,
  rem,
  uint16,
  uint32,
  uint8,
} from "./int.ts";

describe("fixed-width wraps", () => {
  it("int8 and uint8", () => {
    expect([0, 127, 128, 255, 256, -1, -128, -129, 0x1ff].map(int8)).toEqual([
      0, 127, -128, -1, 0, -1, -128, 127, -1,
    ]);
    expect([0, 255, 256, -1, -256, 0x1234].map(uint8)).toEqual([
      0, 255, 0, 255, 0, 0x34,
    ]);
  });

  it("int16 and uint16", () => {
    expect([32767, 32768, 65535, 65536, -32769, 0x12345].map(int16)).toEqual([
      32767, -32768, -1, 0, 32767, 0x2345,
    ]);
    expect([65535, 65536, -1, 0x12345].map(uint16)).toEqual([
      65535, 0, 65535, 0x2345,
    ]);
  });

  it("int32 and uint32, also above 2^32", () => {
    expect(int32(2 ** 31)).toBe(-(2 ** 31));
    expect(int32(2 ** 32 + 5)).toBe(5);
    expect(int32(-(2 ** 31) - 1)).toBe(2 ** 31 - 1);
    expect(int32(3 * 2 ** 40 + 7)).toBe(7);
    expect(uint32(-1)).toBe(2 ** 32 - 1);
    expect(uint32(2 ** 32)).toBe(0);
  });

  it("mul32 wraps the full product", () => {
    expect(mul32(0x10000, 0x10000)).toBe(0);
    expect(mul32(0x7fffffff, 2)).toBe(-2);
    expect(mul32(-3, 5)).toBe(-15);
    // A product beyond 2^53: (2^31-1)^2 = 2^62 - 2^32 + 1, low 32 bits 1.
    expect(mul32(2 ** 31 - 1, 2 ** 31 - 1)).toBe(1);
  });

  it("never gives -0", () => {
    for (const v of [int8(-0), uint8(-0), int16(-0), int32(-0), uint32(-0)]) {
      expect(Object.is(v, 0)).toBe(true);
    }
  });
});

describe("Go integer division", () => {
  it("div truncates toward zero", () => {
    expect([div(7, 2), div(-7, 2), div(7, -2), div(-7, -2)]).toEqual([
      3, -3, -3, 3,
    ]);
    expect(div(-(2 ** 31), -1)).toBe(2 ** 31);
    expect(div(2 ** 40 + 1, 2)).toBe(2 ** 39);
  });

  it("rem takes the dividend's sign", () => {
    expect([rem(7, 3), rem(-7, 3), rem(7, -3), rem(-7, -3)]).toEqual([
      1, -1, 1, -1,
    ]);
  });

  it("never gives -0", () => {
    expect(Object.is(div(-1, 2), 0)).toBe(true);
    expect(Object.is(div(0, -5), 0)).toBe(true);
    expect(Object.is(rem(-4, 2), 0)).toBe(true);
  });

  it("panics on a zero divisor", () => {
    expect(() => div(1, 0)).toThrow(IntegerDivideError);
    expect(() => rem(1, 0)).toThrow(IntegerDivideError);
  });
});
