import type { Point } from "../camera.ts";

/** A single stationary touch may open a menu; this class never issues orders.
 * Positions and thresholds are CSS pixels, independent of the camera's DPR.
 */
export class PressHold {
  private readonly pointers = new Map<number, Point>();
  private readonly suppressed = new Set<number>();
  private pending: { id: number; start: Point } | null = null;
  private timer: ReturnType<typeof setTimeout> | null = null;
  private fired: number | null = null;

  constructor(
    private readonly onHold: (id: number, point: Point) => void,
    private readonly onInterrupted: () => void = () => {},
    private readonly delayMS = 450,
    private readonly slopCSS = 4,
  ) {}

  down(id: number, point: Point, pointerType: string): void {
    if (this.pointers.has(id)) {
      this.move(id, point);
      return;
    }
    this.pointers.set(id, { ...point });
    if (this.pointers.size > 1) {
      this.clearPending();
      for (const pointer of this.pointers.keys()) this.suppressed.add(pointer);
      this.interrupt();
      return;
    }
    if (pointerType !== "touch") return;
    const pending = { id, start: { ...point } };
    this.pending = pending;
    this.timer = setTimeout(() => {
      if (this.pending !== pending) return;
      this.pending = null;
      this.timer = null;
      const position = this.pointers.get(id);
      if (this.pointers.size !== 1 || position === undefined) return;
      this.fired = id;
      this.suppressed.add(id);
      this.onHold(id, { ...position });
    }, this.delayMS);
  }

  /** A fired hold keeps its menu; otherwise the parent may pan/pinch. */
  move(id: number, point: Point): boolean {
    if (!this.pointers.has(id)) return false;
    this.pointers.set(id, { ...point });
    if (this.fired === id) return true;
    const pending = this.pending;
    if (
      pending?.id === id &&
      Math.hypot(point.x - pending.start.x, point.y - pending.start.y) >=
        this.slopCSS
    ) {
      this.suppressed.add(id);
      this.clearPending();
    }
    return false;
  }

  /** A normal fired release leaves the opened menu's lifecycle to the parent. */
  up(id: number): boolean {
    if (!this.pointers.has(id)) return false;
    const suppress = this.suppressed.has(id);
    if (this.pending?.id === id) this.clearPending();
    if (this.fired === id) this.fired = null;
    this.pointers.delete(id);
    this.suppressed.delete(id);
    return suppress;
  }

  /** Blur/disposal drops timers and interrupts an active menu exactly once. */
  cancel(): void {
    this.clearPending();
    this.pointers.clear();
    this.suppressed.clear();
    this.interrupt();
  }

  positions(): ReadonlyMap<number, Point> {
    const positions = new Map<number, Point>();
    for (const [id, point] of this.pointers) positions.set(id, { ...point });
    return positions;
  }

  holding(): boolean {
    return this.fired !== null;
  }

  private clearPending(): void {
    if (this.timer !== null) clearTimeout(this.timer);
    this.timer = null;
    this.pending = null;
  }

  private interrupt(): void {
    if (this.fired === null) return;
    this.fired = null;
    this.onInterrupted();
  }
}
