import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { Camera, type Point } from "./camera.ts";
import { MapInput, type MapInputCallbacks } from "./mapInput.ts";
import { BattleKeys } from "./input/keys.ts";

// Native EventTarget/AbortSignal exercise the installed listeners and teardown.
// The canvas/element surface supplies only DOM geometry, capture and ancestry.
class ElementTarget extends EventTarget {
  parentElement: ElementTarget | null = null;
  isContentEditable = false;
  constructor(readonly tagName: string) {
    super();
  }
  closest(selectors: string): ElementTarget | null {
    const tags = selectors.split(",").map((s) => s.trim().toUpperCase());
    if (tags.includes(this.tagName)) return this;
    for (
      let node: ElementTarget | null = this.parentElement;
      node !== null;
      node = node.parentElement
    )
      if (tags.includes(node.tagName)) return node;
    return null;
  }
}
class CanvasTarget extends ElementTarget {
  readonly captured = new Set<number>();
  constructor() {
    super("CANVAS");
  }
  getBoundingClientRect() {
    return {
      left: 10,
      top: 20,
      width: 400,
      height: 300,
      right: 410,
      bottom: 320,
    };
  }
  setPointerCapture(id: number) {
    this.captured.add(id);
  }
  hasPointerCapture(id: number) {
    return this.captured.has(id);
  }
  releasePointerCapture(id: number) {
    this.captured.delete(id);
  }
}
let win: EventTarget;
const inputs: MapInput[] = [];
beforeEach(() => {
  vi.useFakeTimers();
  win = new EventTarget();
  vi.stubGlobal("window", win);
  vi.stubGlobal("HTMLElement", ElementTarget);
  vi.stubGlobal("Element", ElementTarget);
});
afterEach(() => {
  for (const input of inputs) input.dispose();
  inputs.length = 0;
  vi.clearAllTimers();
  vi.useRealTimers();
  vi.unstubAllGlobals();
});
function setup(extras: Partial<MapInputCallbacks> = {}, withContext = true) {
  const camera = new Camera(
    { width: 4000, height: 3000 },
    { width: 800, height: 600 },
  );
  camera.setScale(2);
  camera.centerOn(2000, 1500);
  const canvas = new CanvasTarget(),
    clicks: Point[] = [],
    contexts: Point[] = [],
    hovers: Point[] = [];
  const calls = {
    changed: 0,
    cancel: 0,
    contextCancel: 0,
    reset: 0,
    phase: 0,
    design: 0,
  };
  const input = new MapInput(canvas as unknown as HTMLCanvasElement, camera, {
    changed: () => {
      calls.changed++;
    },
    hover: (p) => {
      hovers.push(p);
    },
    click: (p) => {
      clicks.push(p);
    },
    cancel: () => {
      calls.cancel++;
    },
    design: () => {
      calls.design++;
    },
    phase: () => {
      calls.phase++;
    },
    ...(withContext
      ? {
          context: (p: Point) => {
            contexts.push(p);
          },
          contextCancel: () => {
            calls.contextCancel++;
          },
          reset: () => {
            calls.reset++;
          },
        }
      : {}),
    ...extras,
  });
  inputs.push(input);
  const pointer = (
    type: string,
    id: number,
    x: number,
    y: number,
    pointerType = "touch",
    button = 0,
  ) => {
    const e = new Event(type, { cancelable: true });
    Object.assign(e, {
      pointerId: id,
      clientX: x + 10,
      clientY: y + 20,
      pointerType,
      button,
    });
    Object.defineProperty(e, "target", { value: canvas });
    if (type === "pointerdown") win.dispatchEvent(e);
    canvas.dispatchEvent(e);
    return e;
  };
  const context = (x: number, y: number, pointerType?: string) => {
    const e = new Event("contextmenu", { cancelable: true });
    Object.assign(e, {
      clientX: x + 10,
      clientY: y + 20,
      ...(pointerType === undefined ? {} : { pointerType }),
    });
    canvas.dispatchEvent(e);
    return e;
  };
  return {
    input,
    camera,
    canvas,
    clicks,
    contexts,
    hovers,
    calls,
    pointer,
    context,
  };
}
function key(
  type: string,
  key: string,
  target?: ElementTarget,
  repeat = false,
) {
  const e = new Event(type, { cancelable: true });
  Object.assign(e, { key, repeat });
  if (target !== undefined)
    Object.defineProperty(e, "target", { value: target });
  win.dispatchEvent(e);
  return e;
}
function focus(target: ElementTarget) {
  const e = new Event("focusin");
  Object.defineProperty(e, "target", { value: target });
  win.dispatchEvent(e);
}
function compatibilityClick(target: ElementTarget, detail = 1): Event {
  const e = new Event("click", { cancelable: true });
  Object.assign(e, { detail, pointerType: "touch" });
  Object.defineProperty(e, "target", { value: target });
  win.dispatchEvent(e);
  return e;
}
function overlayDown(id: number, x: number, y: number): Event {
  const e = new Event("pointerdown", { cancelable: true });
  Object.assign(e, {
    pointerId: id,
    clientX: x + 10,
    clientY: y + 20,
    pointerType: "touch",
    button: 0,
  });
  Object.defineProperty(e, "target", { value: new ElementTarget("BUTTON") });
  win.dispatchEvent(e);
  return e;
}
it("captures a second held touch from an overlay button into the existing pinch arbitration", () => {
  const s = setup();
  s.pointer("pointerdown", 3, 100, 220);
  vi.advanceTimersByTime(450);
  s.pointer("pointermove", 3, 110, 220);
  let buttonDowns = 0;
  win.addEventListener("pointerdown", () => {
    buttonDowns++;
  });
  const anchor = s.camera.screenToWorld(380, 440);
  expect(overlayDown(4, 270, 220).defaultPrevented).toBe(true);
  expect(buttonDowns).toBe(0);
  expect(s.canvas.captured.has(4)).toBe(true);
  expect(s.calls.contextCancel).toBe(1);
  s.pointer("pointermove", 4, 310, 220);
  s.pointer("pointermove", 3, 70, 220);
  expect(s.camera.zoom).toBeCloseTo(1.5, 9);
  expect(s.camera.screenToWorld(380, 440).x).toBeCloseTo(anchor.x, 9);
  s.pointer("pointerup", 3, 70, 220);
  s.pointer("pointerup", 4, 310, 220);
  expect(s.clicks).toEqual([]);
  expect(compatibilityClick(new ElementTarget("BUTTON")).defaultPrevented).toBe(
    true,
  );
});
it("does not hijack overlay touches outside the canvas or after the held pointer releases", () => {
  const s = setup();
  s.pointer("pointerdown", 1, 100, 100);
  vi.advanceTimersByTime(450);
  expect(overlayDown(2, 500, 100).defaultPrevented).toBe(false);
  expect(s.canvas.captured.has(2)).toBe(false);
  expect(s.calls.contextCancel).toBe(0);
  s.pointer("pointerup", 1, 100, 100);
  expect(overlayDown(3, 100, 100).defaultPrevented).toBe(false);
  expect(s.canvas.captured.has(3)).toBe(false);
  expect(compatibilityClick(new ElementTarget("BUTTON")).defaultPrevented).toBe(
    false,
  );
});

it.each(["hold", "drag", "multi"] as const)(
  "blocks a retargeted compatibility click after suppressed touch %s",
  (kind) => {
    const add = vi.spyOn(win, "addEventListener");
    const s = setup();
    s.pointer("pointerdown", 1, 100, 100);
    if (kind === "hold") vi.advanceTimersByTime(450);
    else if (kind === "drag") s.pointer("pointermove", 1, 110, 100);
    else s.pointer("pointerdown", 2, 200, 100);
    s.pointer("pointerup", 1, kind === "drag" ? 110 : 100, 100);
    if (kind === "multi") s.pointer("pointerup", 2, 200, 100);
    // This button mounts only after the canvas pointerup. A downstream UI
    // listener must never receive the synthesized click retargeted onto it.
    const button = new ElementTarget("BUTTON");
    let activations = 0;
    win.addEventListener("click", () => {
      activations++;
    });
    expect(compatibilityClick(button).defaultPrevented).toBe(true);
    expect(activations).toBe(0);
    expect(
      add.mock.calls.some(
        ([type, , options]) =>
          type === "click" &&
          typeof options === "object" &&
          options.capture === true,
      ),
    ).toBe(true);
  },
);
it("does not consume keyboard/programmatic detail-zero clicks while suppressing the pending touch click", () => {
  const s = setup();
  s.pointer("pointerdown", 1, 100, 100);
  vi.advanceTimersByTime(450);
  s.pointer("pointerup", 1, 100, 100);
  const button = new ElementTarget("BUTTON");
  let activations = 0;
  win.addEventListener("click", () => {
    activations++;
  });
  expect(compatibilityClick(button, 0).defaultPrevented).toBe(false);
  expect(activations).toBe(1);
  expect(compatibilityClick(button, 1).defaultPrevented).toBe(true);
  expect(activations).toBe(1);
});
it("allows the next intentional tap after a fresh pointerdown outside the canvas", () => {
  const s = setup();
  s.pointer("pointerdown", 1, 100, 100);
  vi.advanceTimersByTime(450);
  s.pointer("pointerup", 1, 100, 100);
  const button = new ElementTarget("BUTTON");
  const down = new Event("pointerdown");
  Object.defineProperty(down, "target", { value: button });
  win.dispatchEvent(down);
  expect(compatibilityClick(button).defaultPrevented).toBe(false);
});
it("does not debounce ordinary taps and releases the global click guard on disposal", () => {
  const s = setup();
  s.pointer("pointerdown", 1, 100, 100);
  s.pointer("pointerup", 1, 100, 100);
  const button = new ElementTarget("BUTTON");
  expect(compatibilityClick(button).defaultPrevented).toBe(false);
  s.pointer("pointerdown", 2, 100, 100);
  vi.advanceTimersByTime(450);
  s.pointer("pointerup", 2, 100, 100);
  s.input.dispose();
  expect(compatibilityClick(button).defaultPrevented).toBe(false);
});

it("opens a touch hold in physical pixels and retains its ring on movement/release", () => {
  const s = setup(),
    before = { ...s.camera.origin };
  s.pointer("pointerdown", 1, 100, 100);
  vi.advanceTimersByTime(450);
  expect(s.contexts).toEqual([{ x: 200, y: 200 }]);
  s.pointer("pointermove", 1, 150, 120);
  s.pointer("pointerup", 1, 150, 120);
  expect(s.camera.origin).toEqual(before);
  expect(s.clicks).toEqual([]);
  expect(s.calls.changed).toBe(0);
  expect(s.calls.contextCancel).toBe(0);
  expect(s.context(150, 120, "touch").defaultPrevented).toBe(true);
  expect(s.contexts).toHaveLength(1);
});
it("keeps short touch taps and cancels their timers without camera-change callbacks", () => {
  const s = setup();
  s.pointer("pointerdown", 1, 10, 10);
  vi.advanceTimersByTime(449);
  s.pointer("pointerup", 1, 11, 10);
  vi.advanceTimersByTime(1000);
  expect(s.clicks).toEqual([{ x: 22, y: 20 }]);
  expect(s.contexts).toEqual([]);
  expect(s.calls.changed).toBe(0);
});
it("normal touch drag pans and suppresses a later hold or tap", () => {
  const s = setup(),
    before = { ...s.camera.origin };
  s.pointer("pointerdown", 1, 100, 100);
  s.pointer("pointermove", 1, 104, 100);
  s.pointer("pointerup", 1, 104, 100);
  vi.advanceTimersByTime(1000);
  expect(s.camera.origin).not.toEqual(before);
  expect(s.clicks).toEqual([]);
  expect(s.contexts).toEqual([]);
  expect(s.calls.changed).toBe(1);
});
it("rebuilds a stale held gesture from current CSS positions before pinch", () => {
  const s = setup();
  s.pointer("pointerdown", 1, 100, 100);
  vi.advanceTimersByTime(450);
  s.pointer("pointermove", 1, 150, 100);
  const anchor = s.camera.screenToWorld(400, 200);
  s.pointer("pointerdown", 2, 250, 100);
  expect(s.calls.contextCancel).toBe(1);
  s.pointer("pointermove", 2, 300, 100);
  expect(s.camera.zoom).toBeCloseTo(1.5, 9);
  expect(s.camera.screenToWorld(450, 200).x).toBeCloseTo(anchor.x, 9);
  s.pointer("pointerup", 1, 150, 100);
  s.pointer("pointerup", 2, 300, 100);
  expect(s.clicks).toEqual([]);
});
it("keeps ordinary two-touch pinch and suppresses both taps before a hold fires", () => {
  const s = setup();
  s.pointer("pointerdown", 1, 100, 100);
  s.pointer("pointerdown", 2, 200, 100);
  s.pointer("pointermove", 1, 50, 100);
  s.pointer("pointermove", 2, 250, 100);
  expect(s.camera.zoom).toBeCloseTo(2, 9);
  s.pointer("pointerup", 1, 50, 100);
  s.pointer("pointerup", 2, 250, 100);
  vi.advanceTimersByTime(1000);
  expect(s.contexts).toEqual([]);
  expect(s.clicks).toEqual([]);
});
it("ignores synthetic touch context menus but opens a genuine later right-click", () => {
  const s = setup();
  s.pointer("pointerdown", 1, 20, 30);
  expect(s.context(20, 30).defaultPrevented).toBe(true);
  expect(s.contexts).toEqual([]);
  vi.advanceTimersByTime(450);
  s.pointer("pointerup", 1, 20, 30);
  s.context(20, 30);
  expect(s.contexts).toEqual([{ x: 40, y: 60 }]);
  s.pointer("pointerdown", 2, 80, 90, "mouse", 2);
  s.context(80, 90, "mouse");
  expect(s.contexts.at(-1)).toEqual({ x: 160, y: 180 });
  expect(s.contexts).toHaveLength(2);
});
it("preserves W5 context cancellation and ordinary long taps without context hooks", () => {
  const s = setup({}, false);
  s.context(5, 6);
  expect(s.calls.cancel).toBe(1);
  s.pointer("pointerdown", 1, 20, 30);
  vi.advanceTimersByTime(450);
  s.pointer("pointerup", 1, 20, 30);
  expect(s.clicks).toEqual([{ x: 40, y: 60 }]);
});
it.each(["blur", "pointercancel", "dispose"])(
  "%s clears keys, holds, captures and ring without a delayed callback",
  (kind) => {
    const s = setup();
    s.pointer("pointerdown", 1, 20, 30);
    vi.advanceTimersByTime(450);
    if (kind === "dispose") s.input.dispose();
    else if (kind === "blur") win.dispatchEvent(new Event("blur"));
    else s.pointer("pointercancel", 1, 20, 30);
    expect(s.calls.reset).toBe(1);
    expect(s.calls.contextCancel).toBe(1);
    expect(s.canvas.captured.size).toBe(0);
    s.pointer("pointerup", 1, 20, 30);
    vi.advanceTimersByTime(1000);
    expect(s.clicks).toEqual([]);
    expect(s.contexts).toHaveLength(1);
  },
);
it("disposal removes native event listeners and pending hold timers", () => {
  const s = setup();
  s.pointer("pointerdown", 1, 20, 30);
  s.input.dispose();
  s.pointer("pointerdown", 2, 40, 50);
  s.context(40, 50);
  key("keydown", "F1");
  vi.advanceTimersByTime(1000);
  expect(s.contexts).toEqual([]);
  expect(s.calls.phase).toBe(0);
  expect(s.calls.reset).toBe(1);
  expect(s.canvas.captured.size).toBe(0);
});
it("delegates native keyboard hooks once and does not also run fallback camera/phase actions", () => {
  const events: string[] = [],
    s = setup({
      keyDown: (k, r) => {
        events.push(`down:${k}:${String(r)}`);
        return k !== "x";
      },
      keyUp: (k) => {
        events.push(`up:${k}`);
        return true;
      },
    }),
    before = { ...s.camera.origin };
  expect(key("keydown", "ArrowUp").defaultPrevented).toBe(true);
  s.input.tick(16);
  expect(s.camera.origin).toEqual(before);
  expect(key("keyup", "ArrowUp").defaultPrevented).toBe(true);
  key("keydown", "F1");
  expect(s.calls.phase).toBe(0);
  expect(key("keydown", "x").defaultPrevented).toBe(false);
  expect(events).toEqual([
    "down:ArrowUp:false",
    "up:ArrowUp",
    "down:F1:false",
    "down:x:false",
  ]);
  expect(s.calls.changed).toBe(0);
});
it("ignores editable and nested button targets while allowing global function shortcuts", () => {
  const events: string[] = [],
    s = setup({
      keyDown: (k) => {
        events.push(k);
        return true;
      },
      keyUp: () => true,
    });
  const button = new ElementTarget("BUTTON"),
    span = new ElementTarget("SPAN");
  span.parentElement = button;
  const editable = new ElementTarget("DIV");
  editable.isContentEditable = true;
  for (const target of [
    new ElementTarget("INPUT"),
    new ElementTarget("TEXTAREA"),
    new ElementTarget("SELECT"),
    button,
    span,
    editable,
  ]) {
    key("keydown", " ", target);
    key("keyup", " ", target);
    key("keydown", "Enter", target);
    key("keyup", "Enter", target);
  }
  expect(events).toEqual([]);
  key("keydown", "F1", button);
  expect(events).toEqual(["F1"]);
  expect(s.clicks).toEqual([]);
});
it("resets held fire when keyboard focus changes rather than releasing it into an order", () => {
  let clicks = 0,
    resets = 0;
  const keys = new BattleKeys({
    step: () => {},
    click: () => {
      clicks++;
    },
    run: () => {
      clicks++;
    },
    placement: () => ({ up: "move" }),
    shortcut: () => false,
  });
  setup({
    keyDown: (k, r) => keys.down(k, r),
    keyUp: (k) => keys.up(k),
    reset: () => {
      resets++;
      keys.cancel();
    },
  });
  key("keydown", " ");
  const button = new ElementTarget("BUTTON");
  focus(button);
  key("keyup", " ", button);
  expect(clicks).toBe(0);
  expect(resets).toBe(1);
  key("keydown", "Enter");
  key("keyup", "Enter");
  expect(clicks).toBe(1);
});
it("preserves W5 fallback shortcuts and releases camera keys after focus changes", () => {
  const s = setup({}, false),
    before = { ...s.camera.origin };
  key("keydown", "ArrowRight");
  s.input.tick(1000 / 60);
  expect(s.camera.origin.x).toBeGreaterThan(before.x);
  key("keyup", "ArrowRight", new ElementTarget("INPUT"));
  const stopped = { ...s.camera.origin };
  s.input.tick(16);
  expect(s.camera.origin).toEqual(stopped);
  key("keydown", "F1");
  key("keydown", "F1", undefined, true);
  key("keydown", "F8");
  key("keydown", "Escape");
  expect([s.calls.phase, s.calls.design, s.calls.cancel]).toEqual([1, 1, 1]);
});
it("falls back to WASD camera pan once when the native key hook declines it", () => {
  const s = setup({ keyDown: () => false, keyUp: () => false });
  const before = { ...s.camera.origin };
  key("keydown", "w");
  key("keydown", "w", undefined, true);
  s.input.tick(1000 / 60);
  expect(s.camera.origin.y).toBeCloseTo(before.y - 12, 9);
  expect(s.calls.changed).toBe(1);
  key("keyup", "w");
  const stopped = { ...s.camera.origin };
  s.input.tick(16);
  expect(s.camera.origin).toEqual(stopped);
});
it("keeps Escape/M global on buttons and summaries without intercepting native activation", () => {
  const events: string[] = [];
  setup({
    keyDown: (k) => {
      events.push(`down:${k}`);
      return true;
    },
    keyUp: (k) => {
      events.push(`up:${k}`);
      return true;
    },
  });
  for (const target of [
    new ElementTarget("BUTTON"),
    new ElementTarget("SUMMARY"),
  ]) {
    for (const activation of [" ", "Enter", "ArrowDown"]) {
      expect(key("keydown", activation, target).defaultPrevented).toBe(false);
      expect(key("keyup", activation, target).defaultPrevented).toBe(false);
    }
    expect(key("keydown", "Escape", target).defaultPrevented).toBe(true);
    key("keyup", "Escape", target);
    expect(key("keydown", "m", target).defaultPrevented).toBe(true);
    key("keyup", "m", target);
  }
  expect(events).toEqual([
    "down:Escape",
    "up:Escape",
    "down:m",
    "up:m",
    "down:Escape",
    "up:Escape",
    "down:m",
    "up:m",
  ]);
});
it("reserves text-entry M while Escape still reaches the modal cancellation hook", () => {
  const events: string[] = [];
  setup({
    keyDown: (k) => {
      events.push(k);
      return true;
    },
    keyUp: () => true,
  });
  for (const target of [
    new ElementTarget("INPUT"),
    new ElementTarget("TEXTAREA"),
    new ElementTarget("SELECT"),
  ]) {
    key("keydown", "m", target);
    key("keyup", "m", target);
    expect(key("keydown", "Escape", target).defaultPrevented).toBe(true);
    key("keyup", "Escape", target);
  }
  expect(events).toEqual(["Escape", "Escape", "Escape"]);
});

it("lets a split-screen router own the keyboard while retaining map pointer controls", () => {
  const down = vi.fn(() => true),
    up = vi.fn(() => true),
    s = setup({ keyboard: false, keyDown: down, keyUp: up });
  const before = { ...s.camera.origin };
  for (const value of ["ArrowRight", "d", "F1", "F8", "Escape"]) {
    expect(key("keydown", value).defaultPrevented).toBe(false);
    s.input.tick(16);
    key("keyup", value);
  }
  expect(down).not.toHaveBeenCalled();
  expect(up).not.toHaveBeenCalled();
  expect(s.camera.origin).toEqual(before);
  expect(s.calls.phase).toBe(0);
  expect(s.calls.design).toBe(0);
  expect(s.calls.cancel).toBe(0);
  s.pointer("pointerdown", 1, 100, 100, "mouse");
  s.pointer("pointerup", 1, 100, 100, "mouse");
  expect(s.clicks).toEqual([{ x: 200, y: 200 }]);
});

it("accepts camera keys from the split router and releases them on focus or disposal", () => {
  const s = setup({ keyboard: false }),
    before = { ...s.camera.origin };
  expect(s.input.cameraKey("d", true)).toBe(true);
  s.input.tick(16);
  expect(s.camera.origin.x).toBeGreaterThan(before.x);
  s.input.cameraKey("d", false);
  const released = { ...s.camera.origin };
  s.input.tick(16);
  expect(s.camera.origin).toEqual(released);
  s.input.cameraKey("w", true);
  focus(new ElementTarget("INPUT"));
  s.input.tick(16);
  expect(s.camera.origin).toEqual(released);
  s.input.cameraKey("s", true);
  s.input.clearKeyboard();
  s.input.tick(16);
  expect(s.camera.origin).toEqual(released);
  expect(s.input.cameraKey("r", true)).toBe(false);
  s.input.dispose();
  expect(s.input.cameraKey("d", true)).toBe(false);
});
