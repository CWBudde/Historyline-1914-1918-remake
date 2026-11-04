// Go's error values and runtime checks in the port. A Go function that returns (T, error) throws a GoError in
// TS and returns T; attempt turns it back into Go's pair where a caller (a fixture adapter) wants both. Go's
// panics (an index out of range, a nil dereference) stay other exceptions, which attempt never catches.

/** GoError is an error a Go function returns, as opposed to a panic. Sentinels (Go's ErrX) subclass it. */
export class GoError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = new.target.name;
  }
}

/** GoResult is Go's (value, error) pair. */
export type GoResult<T> = readonly [T, null] | readonly [null, GoError];

/** attempt runs f and returns its value, or the GoError it threw; other exceptions (panics) pass through. */
export function attempt<T>(f: () => T): GoResult<T> {
  try {
    return [f(), null];
  } catch (e) {
    if (e instanceof GoError) {
      return [null, e];
    }
    throw e;
  }
}

/**
 * errorsIs is Go's errors.Is for a sentinel class: err or an error in its cause chain (Go's %w wrapping is
 * `new GoError("prefix: " + err.message, { cause: err })`) is an instance of cls.
 */
export function errorsIs(
  err: unknown,
  cls: abstract new (...args: never[]) => Error,
): boolean {
  let e: unknown = err;
  while (e instanceof Error) {
    const cause: unknown = e.cause;
    if (e instanceof cls) {
      return true;
    }
    e = cause;
  }
  return false;
}

/** at is Go's bounds-checked a[i]: it panics (throws a RangeError) outside the slice. */
export function at<T>(a: ArrayLike<T>, i: number): T {
  if (!Number.isInteger(i) || i < 0 || i >= a.length) {
    throw new RangeError(
      `runtime error: index out of range [${i}] with length ${a.length}`,
    );
  }
  return a[i] as T;
}
