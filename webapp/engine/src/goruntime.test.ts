import { describe, expect, it } from "vitest";
import { GoError, at, attempt, errorsIs } from "./goruntime.ts";

class ErrSentinel extends GoError {
  constructor() {
    super("sentinel");
  }
}

describe("Go errors", () => {
  it("attempt returns the value or the GoError", () => {
    expect(attempt(() => 3)).toEqual([3, null]);
    const [v, err] = attempt(() => {
      throw new ErrSentinel();
    });
    expect(v).toBeNull();
    expect(err).toBeInstanceOf(ErrSentinel);
    expect(err?.name).toBe("ErrSentinel");
  });

  it("attempt lets panics through", () => {
    expect(() =>
      attempt(() => {
        throw new TypeError("nil");
      }),
    ).toThrow(TypeError);
  });

  it("errorsIs follows the cause chain", () => {
    const wrapped = new GoError("load: sentinel", { cause: new ErrSentinel() });
    expect(errorsIs(wrapped, ErrSentinel)).toBe(true);
    expect(errorsIs(new GoError("other"), ErrSentinel)).toBe(false);
    expect(errorsIs(null, ErrSentinel)).toBe(false);
  });
});

describe("at", () => {
  it("indexes inside the slice and panics outside", () => {
    expect(at([4, 5], 1)).toBe(5);
    expect(() => at([4, 5], 2)).toThrow(RangeError);
    expect(() => at([4, 5], -1)).toThrow(RangeError);
  });
});
