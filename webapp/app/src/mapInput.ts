import type { Camera, Point } from "./camera.ts";

interface Pointer {
  start: Point;
  at: Point;
  moved: boolean;
}
const midpoint = (a: Point, b: Point): Point => ({
  x: (a.x + b.x) / 2,
  y: (a.y + b.y) / 2,
});
const distance = (a: Point, b: Point) => Math.hypot(a.x - b.x, a.y - b.y);

/** Stateful pointer arbitration; it changes only the camera, never game rules. */
export class MapGesture {
  private pointers = new Map<number, Pointer>();
  constructor(readonly camera: Camera) {}
  private physical(x: number, y: number): Point {
    const d = this.camera.scale > 0 ? this.camera.scale : 1;
    return { x: x * d, y: y * d };
  }
  down(id: number, x: number, y: number): void {
    const at = this.physical(x, y),
      multiple = this.pointers.size > 0;
    if (multiple) for (const p of this.pointers.values()) p.moved = true;
    this.pointers.set(id, { start: at, at, moved: multiple });
  }
  move(id: number, x: number, y: number): boolean {
    const p = this.pointers.get(id);
    if (p === undefined) return false;
    const pair = [...this.pointers.values()].slice(0, 2),
      a = pair[0],
      b = pair[1];
    const oldMid =
      a !== undefined && b !== undefined ? midpoint(a.at, b.at) : null;
    const oldDistance =
      a !== undefined && b !== undefined ? distance(a.at, b.at) : 0;
    const previous = p.at,
      wasMoved = p.moved;
    p.at = this.physical(x, y);
    if (oldMid !== null && a !== undefined && b !== undefined) {
      const mid = midpoint(a.at, b.at),
        span = distance(a.at, b.at);
      if (oldDistance > 0 && span > 0)
        this.camera.zoomAt(span / oldDistance, oldMid.x, oldMid.y);
      this.camera.scroll(oldMid.x - mid.x, oldMid.y - mid.y);
      return true;
    }
    p.moved ||=
      distance(p.start, p.at) >=
      4 * (this.camera.scale > 0 ? this.camera.scale : 1);
    if (!p.moved) return false;
    const from = wasMoved ? previous : p.start;
    this.camera.scroll(from.x - p.at.x, from.y - p.at.y);
    return true;
  }
  up(id: number, x: number, y: number): Point | null {
    const p = this.pointers.get(id);
    if (p === undefined) return null;
    this.move(id, x, y);
    const click = !p.moved && this.pointers.size === 1 ? p.at : null;
    this.pointers.delete(id);
    return click;
  }
  cancel(): void {
    this.pointers.clear();
  }
  wheel(delta: number, mode: number, x: number, y: number): void {
    const pixels = delta * (mode === 1 ? 100 / 3 : mode === 2 ? 800 : 1);
    const p = this.physical(x, y);
    this.camera.zoomAt(
      Math.pow(1.1, -Math.max(-1000, Math.min(1000, pixels)) / 100),
      p.x,
      p.y,
    );
  }
  keyboard(keys: ReadonlySet<string>, ms: number): boolean {
    const x =
      Number(keys.has("arrowright") || keys.has("d")) -
      Number(keys.has("arrowleft") || keys.has("a"));
    const y =
      Number(keys.has("arrowdown") || keys.has("s")) -
      Number(keys.has("arrowup") || keys.has("w"));
    if (x === 0 && y === 0) return false;
    const step =
      ((12 * Math.min(100, Math.max(0, ms))) / (1000 / 60)) *
      (this.camera.scale > 0 ? this.camera.scale : 1);
    this.camera.scroll(x * step, y * step);
    return true;
  }
}

export interface MapInputCallbacks {
  changed(): void;
  hover(point: Point): void;
  click(point: Point): void;
  cancel(): void;
  design(): void;
  phase(): void;
}

/** DOM event lifetime belongs to a map view; closing it releases every handler. */
export class MapInput {
  readonly gesture: MapGesture;
  private keys = new Set<string>();
  private abort = new AbortController();
  constructor(
    canvas: HTMLCanvasElement,
    camera: Camera,
    private callbacks: MapInputCallbacks,
  ) {
    this.gesture = new MapGesture(camera);
    const options = { signal: this.abort.signal };
    const point = (e: { clientX: number; clientY: number }) => {
      const r = canvas.getBoundingClientRect();
      return { x: e.clientX - r.left, y: e.clientY - r.top };
    };
    canvas.addEventListener(
      "pointerdown",
      (e) => {
        if (e.button !== 0) return;
        const p = point(e);
        canvas.setPointerCapture(e.pointerId);
        this.gesture.down(e.pointerId, p.x, p.y);
      },
      options,
    );
    canvas.addEventListener(
      "pointermove",
      (e) => {
        const p = point(e);
        if (this.gesture.move(e.pointerId, p.x, p.y)) callbacks.changed();
        callbacks.hover({ x: p.x * camera.scale, y: p.y * camera.scale });
      },
      options,
    );
    canvas.addEventListener(
      "pointerup",
      (e) => {
        const p = point(e),
          tap = this.gesture.up(e.pointerId, p.x, p.y);
        if (canvas.hasPointerCapture(e.pointerId))
          canvas.releasePointerCapture(e.pointerId);
        callbacks.changed();
        if (tap !== null) callbacks.click(tap);
      },
      options,
    );
    canvas.addEventListener(
      "pointercancel",
      () => {
        this.gesture.cancel();
      },
      options,
    );
    canvas.addEventListener(
      "contextmenu",
      (e) => {
        e.preventDefault();
        this.gesture.cancel();
        callbacks.cancel();
      },
      options,
    );
    canvas.addEventListener(
      "wheel",
      (e) => {
        e.preventDefault();
        const p = point(e);
        this.gesture.wheel(e.deltaY, e.deltaMode, p.x, p.y);
        callbacks.changed();
      },
      { ...options, passive: false },
    );
    window.addEventListener(
      "keydown",
      (e) => {
        const target = e.target;
        if (
          target instanceof HTMLElement &&
          (target.isContentEditable ||
            ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName))
        )
          return;
        const key = e.key.toLowerCase();
        if (
          [
            "arrowleft",
            "arrowright",
            "arrowup",
            "arrowdown",
            "w",
            "a",
            "s",
            "d",
          ].includes(key)
        ) {
          e.preventDefault();
          this.keys.add(key);
        }
        if (!e.repeat) {
          if (key === "escape") callbacks.cancel();
          else if (key === "f8") {
            e.preventDefault();
            callbacks.design();
          } else if (key === "f1") {
            e.preventDefault();
            callbacks.phase();
          }
        }
      },
      options,
    );
    window.addEventListener(
      "keyup",
      (e) => this.keys.delete(e.key.toLowerCase()),
      options,
    );
    window.addEventListener(
      "blur",
      () => {
        this.keys.clear();
        this.gesture.cancel();
      },
      options,
    );
  }
  tick(ms: number): void {
    if (this.gesture.keyboard(this.keys, ms)) this.callbacks.changed();
  }
  dispose(): void {
    this.abort.abort();
    this.keys.clear();
    this.gesture.cancel();
  }
}
