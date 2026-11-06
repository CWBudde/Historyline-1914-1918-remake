import { afterEach, beforeEach, expect, it, vi } from "vitest";
import type { Point } from "../camera.ts";
import { PressHold } from "./press.ts";

beforeEach(() => {
  vi.useFakeTimers();
});
afterEach(() => {
  vi.clearAllTimers();
  vi.useRealTimers();
});
function fixture(delay?: number, slop?: number) {
  const holds: { id: number; point: Point }[] = [],
    interruptions: number[] = [];
  const press = new PressHold(
    (id, point) => {
      holds.push({ id, point });
    },
    () => {
      interruptions.push(1);
    },
    delay,
    slop,
  );
  return { press, holds, interruptions };
}

it("fires one stationary touch hold after exactly450ms and suppresses its release", () => {
  const { press, holds, interruptions } = fixture();
  press.down(1, { x: 20, y: 30 }, "touch");
  expect(press.holding()).toBe(false);
  vi.advanceTimersByTime(449);
  expect(holds).toEqual([]);
  vi.advanceTimersByTime(1);
  expect(holds).toEqual([{ id: 1, point: { x: 20, y: 30 } }]);
  expect(press.holding()).toBe(true);
  vi.advanceTimersByTime(1000);
  expect(holds).toHaveLength(1);
  expect(press.up(1)).toBe(true);
  expect(press.holding()).toBe(false);
  expect(press.positions().size).toBe(0);
  expect(interruptions).toEqual([]);
  expect(press.up(1)).toBe(false);
});
it("keeps an ordinary short touch available as a tap and removes its timer", () => {
  const { press, holds } = fixture();
  press.down(1, { x: 10, y: 10 }, "touch");
  vi.advanceTimersByTime(449);
  expect(press.up(1)).toBe(false);
  vi.advanceTimersByTime(1000);
  expect(holds).toEqual([]);
  expect(press.positions().size).toBe(0);
});
it("allows subthreshold motion and fires at the latest CSS position", () => {
  const { press, holds } = fixture();
  press.down(1, { x: 10, y: 10 }, "touch");
  vi.advanceTimersByTime(200);
  expect(press.move(1, { x: 13, y: 10 })).toBe(false);
  expect(press.positions().get(1)).toEqual({ x: 13, y: 10 });
  vi.advanceTimersByTime(250);
  expect(holds).toEqual([{ id: 1, point: { x: 13, y: 10 } }]);
});
it.each([
  { x: 4, y: 0 },
  { x: 0, y: 4 },
  { x: 3, y: 3 },
])("cancels when CSS movement reaches the threshold: %j", (point) => {
  const { press, holds } = fixture();
  press.down(1, { x: 0, y: 0 }, "touch");
  expect(press.move(1, point)).toBe(false);
  press.move(1, { x: 0, y: 0 });
  vi.advanceTimersByTime(1000);
  expect(holds).toEqual([]);
  expect(press.up(1)).toBe(true);
  expect(press.holding()).toBe(false);
});
it("measures displacement from the original press rather than accumulating tiny moves", () => {
  const { press, holds } = fixture();
  press.down(1, { x: 0, y: 0 }, "touch");
  press.move(1, { x: 1, y: 0 });
  press.move(1, { x: 2, y: 0 });
  press.move(1, { x: 3, y: 0 });
  press.move(1, { x: 4, y: 0 });
  vi.advanceTimersByTime(500);
  expect(holds).toEqual([]);
  expect(press.up(1)).toBe(true);
});
it("keeps a fired menu open while moving and suppresses only that pointer's camera pan", () => {
  const { press, holds, interruptions } = fixture();
  press.down(1, { x: 0, y: 0 }, "touch");
  vi.advanceTimersByTime(450);
  expect(press.move(1, { x: 80, y: 90 })).toBe(true);
  expect(press.move(99, { x: 1, y: 1 })).toBe(false);
  expect(press.positions().get(1)).toEqual({ x: 80, y: 90 });
  expect(holds).toHaveLength(1);
  expect(interruptions).toEqual([]);
  expect(press.up(1)).toBe(true);
});
it("cancels a pending hold for two pointers and suppresses both eventual taps", () => {
  const { press, holds, interruptions } = fixture();
  press.down(1, { x: 10, y: 10 }, "touch");
  vi.advanceTimersByTime(300);
  press.down(2, { x: 50, y: 50 }, "touch");
  expect([...press.positions()]).toEqual([
    [1, { x: 10, y: 10 }],
    [2, { x: 50, y: 50 }],
  ]);
  expect(press.move(1, { x: 11, y: 12 })).toBe(false);
  vi.advanceTimersByTime(1000);
  expect(holds).toEqual([]);
  expect(interruptions).toEqual([]);
  expect(press.up(1)).toBe(true);
  expect(press.up(2)).toBe(true);
});
it("interrupts a fired hold once and leaves both positions available for pinch rebuilding", () => {
  const seen: ReadonlyMap<number, Point>[] = [];
  const press: PressHold = new PressHold(
    () => {},
    () => {
      seen.push(press.positions());
    },
  );
  press.down(1, { x: 10, y: 10 }, "touch");
  vi.advanceTimersByTime(450);
  press.move(1, { x: 12, y: 14 });
  press.down(2, { x: 50, y: 60 }, "touch");
  expect(press.holding()).toBe(false);
  expect(seen).toHaveLength(1);
  expect([...atMap(seen, 0)]).toEqual([
    [1, { x: 12, y: 14 }],
    [2, { x: 50, y: 60 }],
  ]);
  expect(press.move(1, { x: 13, y: 15 })).toBe(false);
  press.down(3, { x: 80, y: 90 }, "touch");
  expect(seen).toHaveLength(1);
  expect(press.up(1)).toBe(true);
  expect(press.up(2)).toBe(true);
  expect(press.up(3)).toBe(true);
});
function atMap(
  values: readonly ReadonlyMap<number, Point>[],
  index: number,
): ReadonlyMap<number, Point> {
  const value = values[index];
  if (value === undefined) throw new Error("missing callback snapshot");
  return value;
}
it("does not restart the hold on the remaining finger after a pinch begins", () => {
  const { press, holds } = fixture();
  press.down(1, { x: 0, y: 0 }, "touch");
  press.down(2, { x: 20, y: 20 }, "touch");
  expect(press.up(2)).toBe(true);
  vi.advanceTimersByTime(1000);
  expect(holds).toEqual([]);
  expect(press.up(1)).toBe(true);
  press.down(1, { x: 5, y: 6 }, "touch");
  vi.advanceTimersByTime(450);
  expect(holds).toEqual([{ id: 1, point: { x: 5, y: 6 } }]);
});
it("cancels timers and pointer state silently before firing, including disposal", () => {
  const { press, holds, interruptions } = fixture();
  press.down(1, { x: 0, y: 0 }, "touch");
  vi.advanceTimersByTime(449);
  press.cancel();
  press.cancel();
  vi.advanceTimersByTime(1000);
  expect(holds).toEqual([]);
  expect(interruptions).toEqual([]);
  expect(press.positions().size).toBe(0);
  expect(press.holding()).toBe(false);
  expect(press.up(1)).toBe(false);
});
it("cancel interrupts an active fired hold only once and removes all state before callbacks", () => {
  const sizes: number[] = [];
  const press: PressHold = new PressHold(
    () => {},
    () => {
      sizes.push(press.positions().size);
      press.cancel();
    },
  );
  press.down(1, { x: 2, y: 3 }, "touch");
  vi.advanceTimersByTime(450);
  press.cancel();
  press.cancel();
  expect(sizes).toEqual([0]);
  expect(press.holding()).toBe(false);
  expect(press.positions().size).toBe(0);
  vi.advanceTimersByTime(1000);
  expect(sizes).toEqual([0]);
});
it("normal release transfers the open menu lifecycle to the parent", () => {
  const { press, interruptions } = fixture();
  press.down(1, { x: 0, y: 0 }, "touch");
  vi.advanceTimersByTime(450);
  expect(press.up(1)).toBe(true);
  press.cancel();
  expect(interruptions).toEqual([]);
});
it("starts a fresh hold after the previous fired gesture ends", () => {
  const { press, holds, interruptions } = fixture();
  press.down(1, { x: 1, y: 2 }, "touch");
  vi.advanceTimersByTime(450);
  press.up(1);
  press.down(2, { x: 3, y: 4 }, "touch");
  vi.advanceTimersByTime(450);
  expect(holds).toEqual([
    { id: 1, point: { x: 1, y: 2 } },
    { id: 2, point: { x: 3, y: 4 } },
  ]);
  expect(interruptions).toEqual([]);
});
it.each(["mouse", "pen", ""])(
  "never schedules a hold for pointer type %s",
  (pointerType) => {
    const { press, holds } = fixture();
    press.down(1, { x: 0, y: 0 }, pointerType);
    vi.advanceTimersByTime(1000);
    expect(press.move(1, { x: 8, y: 9 })).toBe(false);
    expect(holds).toEqual([]);
    expect(press.up(1)).toBe(false);
  },
);
it("treats a simultaneous mouse and touch as multiple involved pointers", () => {
  const { press, holds } = fixture();
  press.down(1, { x: 0, y: 0 }, "mouse");
  press.down(2, { x: 20, y: 20 }, "touch");
  vi.advanceTimersByTime(1000);
  expect(holds).toEqual([]);
  expect(press.up(1)).toBe(true);
  expect(press.up(2)).toBe(true);
});
it("uses configured delay and CSS slop without a DPR conversion", () => {
  const { press, holds } = fixture(100, 8);
  press.down(1, { x: 0, y: 0 }, "touch");
  press.move(1, { x: 7, y: 0 });
  vi.advanceTimersByTime(99);
  expect(holds).toEqual([]);
  vi.advanceTimersByTime(1);
  expect(holds).toEqual([{ id: 1, point: { x: 7, y: 0 } }]);
});
it("ignores duplicate down events and does not postpone the pending timer", () => {
  const { press, holds } = fixture();
  press.down(1, { x: 0, y: 0 }, "touch");
  vi.advanceTimersByTime(300);
  press.down(1, { x: 1, y: 0 }, "touch");
  vi.advanceTimersByTime(150);
  expect(holds).toEqual([{ id: 1, point: { x: 1, y: 0 } }]);
});
it("does not expose mutable internal positions or retain mutable input points", () => {
  const { press, holds } = fixture(),
    point = { x: 1, y: 2 };
  press.down(1, point, "touch");
  point.x = 100;
  const snapshot = press.positions(),
    at = snapshot.get(1);
  if (at === undefined) throw new Error("missing point");
  at.x = 200;
  expect(press.positions().get(1)).toEqual({ x: 1, y: 2 });
  vi.advanceTimersByTime(450);
  expect(holds).toEqual([{ id: 1, point: { x: 1, y: 2 } }]);
});
it("registers the fired state before onHold so parent cancellation is safe", () => {
  const holds: number[] = [],
    interruptions: number[] = [];
  const press: PressHold = new PressHold(
    (id) => {
      holds.push(id);
      press.cancel();
    },
    () => {
      interruptions.push(1);
    },
  );
  press.down(1, { x: 0, y: 0 }, "touch");
  vi.advanceTimersByTime(450);
  expect(holds).toEqual([1]);
  expect(interruptions).toEqual([1]);
  expect(press.holding()).toBe(false);
  expect(press.positions().size).toBe(0);
});
