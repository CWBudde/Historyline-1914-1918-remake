import type { Camera, Point } from "./camera.ts";
import { PressHold } from "./input/press.ts";

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
  keyboard(keys: ReadonlySet<string>, ms: number, speed = 1): boolean {
    const x =
      Number(keys.has("arrowright") || keys.has("d")) -
      Number(keys.has("arrowleft") || keys.has("a"));
    const y =
      Number(keys.has("arrowdown") || keys.has("s")) -
      Number(keys.has("arrowup") || keys.has("w"));
    if (x === 0 && y === 0) return false;
    const step =
      ((12 * Math.min(100, Math.max(0, ms))) / (1000 / 60)) *
      (this.camera.scale > 0 ? this.camera.scale : 1) *
      speed;
    this.camera.scroll(x * step, y * step);
    return true;
  }
}

export interface MapInputCallbacks {
  /** A split screen installs one router for both native key sets. */
  keyboard?: boolean;
  changed(): void;
  hover(point: Point): void;
  /** A tap or click that did not pan; pointerType is the PointerEvent's ("mouse", "touch", "pen"). */
  click(point: Point, pointerType: string): void;
  cancel(): void;
  design(): void;
  phase(): void;
  keyDown?(key: string, repeat: boolean): boolean;
  keyUp?(key: string): boolean;
  reset?(): void;
  context?(point: Point, pointerType: string): void;
  contextCancel?(): void;
  /**
   * A press the battle reads as a flick from a unit instead of a pan (start returns true): its moves, its release
   * (held: a long press fired during it) and its cancellation. Points are device px like the others.
   */
  flick?: {
    start(point: Point, pointerType: string): boolean;
    move(point: Point): void;
    end(point: Point, held: boolean): void;
    cancel(): void;
  };
}

/** DOM event lifetime belongs to a map view; closing it releases every handler. */
export class MapInput {
  readonly gesture: MapGesture;
  private keys = new Set<string>();
  private activeKeys = new Set<string>();
  private abort = new AbortController();
  private press: PressHold;
  private touchContext = false;
  /** The pointer of a press read as a flick, which never pans. */
  private flickId: number | null = null;
  private pointerTypes = new Map<number, string>();
  private suppressTouchClick = false;
  private disposed = false;
  constructor(
    private readonly canvas: HTMLElement,
    camera: Camera,
    private callbacks: MapInputCallbacks,
  ) {
    this.gesture = new MapGesture(camera);
    const physical = (p: Point) => {
      const d = camera.scale > 0 ? camera.scale : 1;
      return { x: p.x * d, y: p.y * d };
    };
    this.press = new PressHold(
      (_id, p) => {
        if (this.activeKeys.size > 0 || this.keys.size > 0)
          this.resetKeyboard();
        callbacks.context?.(physical(p), "touch");
      },
      () => callbacks.contextCancel?.(),
    );
    const options = { signal: this.abort.signal };
    // A suppressed touch pointerup can still synthesize a click on a newly
    // mounted DOM control. Catch it before that retargeted control sees it.
    // A fresh pointerdown anywhere starts a new intentional tap.
    window.addEventListener(
      "pointerdown",
      (e) => {
        this.suppressTouchClick = false;
        // A fired ring can cover the canvas. Its second touch belongs to
        // the ongoing map gesture, rather than activating an overlay button.
        if (
          !this.press.holding() ||
          e.pointerType !== "touch" ||
          e.button !== 0 ||
          e.target === canvas ||
          this.press.positions().has(e.pointerId)
        )
          return;
        const rect = canvas.getBoundingClientRect();
        if (
          e.clientX < rect.left ||
          e.clientX >= rect.right ||
          e.clientY < rect.top ||
          e.clientY >= rect.bottom
        )
          return;
        e.preventDefault();
        e.stopImmediatePropagation();
        this.beginPointer(e);
      },
      { ...options, capture: true },
    );
    window.addEventListener(
      "click",
      (e) => {
        if (this.suppressTouchClick && e.detail > 0) {
          this.suppressTouchClick = false;
          e.preventDefault();
          e.stopImmediatePropagation();
        }
      },
      { ...options, capture: true },
    );
    const point = (e: { clientX: number; clientY: number }) => {
      const r = canvas.getBoundingClientRect();
      return { x: e.clientX - r.left, y: e.clientY - r.top };
    };
    canvas.addEventListener(
      "pointerdown",
      (e) => {
        this.beginPointer(e);
      },
      options,
    );
    canvas.addEventListener(
      "pointermove",
      (e) => {
        const p = point(e);
        if (e.pointerId === this.flickId) {
          this.press.move(e.pointerId, p);
          callbacks.flick?.move(physical(p));
          return;
        }
        if (this.press.move(e.pointerId, p)) return;
        this.cameraAction(() => this.gesture.move(e.pointerId, p.x, p.y));
        callbacks.hover(physical(p));
      },
      options,
    );
    canvas.addEventListener(
      "pointerup",
      (e) => {
        const known = this.press.positions().has(e.pointerId);
        const p = point(e);
        const pointerType = this.pointerTypes.get(e.pointerId) ?? e.pointerType;
        this.pointerTypes.delete(e.pointerId);
        if (e.pointerId === this.flickId) {
          this.press.move(e.pointerId, p);
          const held = this.press.holding();
          this.press.up(e.pointerId);
          this.flickId = null;
          if (e.pointerType === "touch") this.suppressTouchClick = true;
          if (this.canvas.hasPointerCapture(e.pointerId))
            this.canvas.releasePointerCapture(e.pointerId);
          callbacks.flick?.end(physical(p), held);
          return;
        }
        const held = this.press.move(e.pointerId, p);
        const suppress = this.press.up(e.pointerId);
        let tap: Point | null;
        if (held) {
          this.gesture.cancel();
          tap = null;
        } else
          tap = this.cameraAction(() => this.gesture.up(e.pointerId, p.x, p.y));
        if (known && e.pointerType === "touch" && (suppress || tap === null))
          this.suppressTouchClick = true;
        if (canvas.hasPointerCapture(e.pointerId))
          canvas.releasePointerCapture(e.pointerId);
        if (!suppress && tap !== null) callbacks.click(tap, pointerType);
      },
      options,
    );
    canvas.addEventListener(
      "pointercancel",
      () => {
        this.resetInput();
      },
      options,
    );
    canvas.addEventListener(
      "contextmenu",
      (e) => {
        e.preventDefault();
        // Browsers can synthesize contextmenu before or after a touch hold.
        // A real mouse pointerdown switches modality before its right-click.
        if (
          this.touchContext ||
          ("pointerType" in e && e.pointerType === "touch")
        )
          return;
        this.resetInput();
        if (callbacks.context !== undefined)
          callbacks.context(physical(point(e)), "mouse");
        else callbacks.cancel();
      },
      options,
    );
    canvas.addEventListener(
      "wheel",
      (e) => {
        e.preventDefault();
        const p = point(e);
        this.cameraAction(() => {
          this.gesture.wheel(e.deltaY, e.deltaMode, p.x, p.y);
        });
      },
      { ...options, passive: false },
    );
    window.addEventListener(
      "keydown",
      (e) => {
        if (callbacks.keyboard === false) return;
        if (reserved(e.target, e.key)) {
          if (this.activeKeys.size > 0 || this.keys.size > 0)
            this.resetKeyboard();
          return;
        }
        if (this.press.positions().size === 0) this.touchContext = false;
        const key = e.key.toLowerCase();
        if (callbacks.keyDown !== undefined) {
          this.activeKeys.add(key);
          if (callbacks.keyDown(e.key, e.repeat)) {
            e.preventDefault();
            return;
          }
          this.activeKeys.delete(key);
        }
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
          this.activeKeys.add(key);
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
      (e) => {
        if (callbacks.keyboard === false) return;
        const key = e.key.toLowerCase();
        const held = this.activeKeys.delete(key) || this.keys.has(key);
        this.keys.delete(key);
        if (reserved(e.target, e.key)) {
          if (held || this.activeKeys.size > 0 || this.keys.size > 0)
            this.resetKeyboard();
          return;
        }
        if (callbacks.keyUp?.(e.key)) e.preventDefault();
      },
      options,
    );
    window.addEventListener(
      "focusin",
      (e) => {
        if (
          editable(e.target) &&
          (this.activeKeys.size > 0 || this.keys.size > 0)
        )
          this.resetKeyboard();
      },
      options,
    );
    window.addEventListener(
      "blur",
      () => {
        this.resetInput();
      },
      options,
    );
  }
  tick(ms: number, speed = 1): void {
    this.cameraAction(() => this.gesture.keyboard(this.keys, ms, speed));
  }
  /** Continuous panning when a shared split-screen router owns DOM keys. */
  cameraKey(key: string, down: boolean): boolean {
    const lower = key.toLowerCase();
    if (this.disposed || !["w", "a", "s", "d"].includes(lower)) return false;
    if (down) {
      this.keys.add(lower);
      this.activeKeys.add(lower);
    } else {
      this.keys.delete(lower);
      this.activeKeys.delete(lower);
    }
    return true;
  }
  clearKeyboard(): void {
    this.resetKeyboard();
  }
  dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    this.abort.abort();
    this.resetInput();
  }
  private beginPointer(e: PointerEvent): void {
    this.touchContext = e.pointerType === "touch";
    if (e.button !== 0) return;
    const rect = this.canvas.getBoundingClientRect();
    const p = { x: e.clientX - rect.left, y: e.clientY - rect.top };
    this.canvas.setPointerCapture(e.pointerId);
    this.pointerTypes.set(e.pointerId, e.pointerType);
    if (this.flickId !== null) {
      // A second finger turns a flick into a pinch.
      this.flickId = null;
      this.callbacks.flick?.cancel();
      this.press.down(e.pointerId, p, e.pointerType);
      this.gesture.cancel();
      for (const [id, at] of this.press.positions())
        this.gesture.down(id, at.x, at.y);
      return;
    }
    if (
      this.press.positions().size === 0 &&
      this.callbacks.flick?.start(this.physicalPoint(p), e.pointerType) === true
    ) {
      this.flickId = e.pointerId;
      this.press.down(
        e.pointerId,
        p,
        this.callbacks.context === undefined ? "mouse" : e.pointerType,
      );
      return;
    }
    const held = this.press.holding();
    this.press.down(
      e.pointerId,
      p,
      this.callbacks.context === undefined ? "mouse" : e.pointerType,
    );
    if (this.disposed) return;
    if (held && this.press.positions().size > 1) {
      // Held moves were intentionally not fed to MapGesture. Rebuild
      // from current CSS positions before the first pinch delta.
      this.gesture.cancel();
      for (const [id, at] of this.press.positions())
        this.gesture.down(id, at.x, at.y);
    } else this.gesture.down(e.pointerId, p.x, p.y);
  }
  private resetKeyboard(): void {
    this.keys.clear();
    this.activeKeys.clear();
    this.callbacks.reset?.();
  }
  private physicalPoint(p: Point): Point {
    const d = this.gesture.camera.scale > 0 ? this.gesture.camera.scale : 1;
    return { x: p.x * d, y: p.y * d };
  }
  private resetInput(): void {
    this.resetKeyboard();
    if (this.flickId !== null) {
      this.flickId = null;
      this.callbacks.flick?.cancel();
    }
    this.pointerTypes.clear();

    const held = this.press.holding();
    for (const id of this.press.positions().keys()) {
      if (this.canvas.hasPointerCapture(id))
        this.canvas.releasePointerCapture(id);
    }
    this.gesture.cancel();
    this.press.cancel();
    if (!held) this.callbacks.contextCancel?.();
  }
  private cameraAction<T>(action: () => T): T {
    const camera = this.gesture.camera;
    const before = {
      x: camera.origin.x,
      y: camera.origin.y,
      zoom: camera.zoom,
    };
    const result = action();
    if (
      before.x !== camera.origin.x ||
      before.y !== camera.origin.y ||
      before.zoom !== camera.zoom
    )
      this.callbacks.changed();
    return result;
  }
}

function editable(target: EventTarget | null): target is Element {
  return (
    target instanceof Element &&
    (target.closest("input,textarea,select,button,summary") !== null ||
      (target instanceof HTMLElement && target.isContentEditable))
  );
}
function reserved(target: EventTarget | null, key: string): boolean {
  if (!editable(target) || functionKey(key) || key.toLowerCase() === "escape")
    return false;
  if (
    target.closest("input,textarea,select") !== null ||
    (target instanceof HTMLElement && target.isContentEditable)
  )
    return true;
  return [
    " ",
    "space",
    "enter",
    "arrowup",
    "arrowleft",
    "arrowright",
    "arrowdown",
  ].includes(key.toLowerCase());
}
function functionKey(key: string): boolean {
  return /^F\d{1,2}$/i.test(key);
}
