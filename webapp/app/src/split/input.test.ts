import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { Camera } from "../camera.ts";
import { MapGesture } from "../mapInput.ts";
import { BattleKeys } from "../input/keys.ts";
import type { Settings } from "../storage/store.ts";
import { SplitInput } from "./input.ts";

class ElementTarget extends EventTarget {
  parentElement: ElementTarget | null = null;
  isContentEditable = false;
  constructor(readonly tagName: string) {
    super();
  }
  closest(selectors: string): ElementTarget | null {
    const tags = selectors.split(",").map((tag) => tag.trim().toUpperCase());
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

let win: EventTarget;
const owners: SplitInput[] = [];
beforeEach(() => {
  win = new EventTarget();
  vi.stubGlobal("window", win);
  vi.stubGlobal("Element", ElementTarget);
  vi.stubGlobal("HTMLElement", ElementTarget);
});
afterEach(() => {
  for (const input of owners) input.dispose();
  owners.length = 0;
  vi.unstubAllGlobals();
});

function key(
  type: "keydown" | "keyup",
  value: string,
  code = value,
  options: {
    ctrlKey?: boolean;
    altKey?: boolean;
    metaKey?: boolean;
    repeat?: boolean;
    target?: ElementTarget;
  } = {},
) {
  const event = new Event(type, { cancelable: true });
  const { target, ...properties } = options;
  Object.assign(event, {
    key: value,
    code,
    ctrlKey: false,
    altKey: false,
    metaKey: false,
    repeat: false,
    ...properties,
  });
  if (target !== undefined)
    Object.defineProperty(event, "target", { value: target });
  win.dispatchEvent(event);
  return event;
}

function setup(mode: Settings["playMode"] = "simultaneous", active: 0 | 1 = 0) {
  const state = {
      mode,
      active,
      blocked: false,
      handover: false,
      resets: 0,
      leaves: 0,
    },
    steps: string[] = [],
    clicks: number[] = [],
    actions: string[] = [],
    shortcuts: string[] = [],
    releases: string[] = [],
    cameraReleases: string[] = [],
    held = [new Set<string>(), new Set<string>()],
    cameras = [0, 1].map(() => {
      const camera = new Camera(
        { width: 4000, height: 4000 },
        { width: 600, height: 400 },
      );
      camera.centerOn(2000, 2000);
      return camera;
    }),
    gestures = cameras.map((camera) => new MapGesture(camera)),
    keys = [0, 1].map(
      (side) =>
        new BattleKeys({
          step: (direction) => {
            steps.push(`${side}/${direction}`);
          },
          click: () => {
            clicks.push(side);
          },
          run: (action) => {
            actions.push(`${side}/${action}`);
          },
          placement: () => ({
            left: "mode",
            up: "info",
            right: "map",
            down: "open",
          }),
          shortcut: (value) => {
            shortcuts.push(`${side}/${value}`);
            return true;
          },
        }),
    ),
    input = new SplitInput({
      mode: () => state.mode,
      activeSide: () => state.active,
      blocked: () => state.blocked,
      handover: () => state.handover,
      keyDown: (side, value, repeat) =>
        keys[side]?.down(value, repeat) ?? false,
      keyUp: (side, value) => {
        releases.push(`${side}/${value}`);
        return keys[side]?.up(value) ?? false;
      },
      cameraKey: (side, value, down) => {
        if (down) held[side]?.add(value);
        else {
          held[side]?.delete(value);
          cameraReleases.push(`${side}/${value}`);
        }
        return true;
      },
      reset: () => {
        state.resets++;
        for (const keySet of keys) keySet.cancel();
        for (const cameraKeys of held) cameraKeys.clear();
      },
      leave: () => {
        state.leaves++;
      },
    });
  owners.push(input);
  return {
    input,
    state,
    steps,
    clicks,
    actions,
    shortcuts,
    releases,
    cameraReleases,
    cameras,
    tick: () => {
      for (const side of [0, 1]) {
        const cameraKeys = held[side];
        if (cameraKeys !== undefined) gestures[side]?.keyboard(cameraKeys, 100);
      }
    },
  };
}

it("keeps A arrows, Space and F1 independent while B holds Left Control", () => {
  const s = setup();
  expect(
    key("keydown", "Control", "ControlLeft", { ctrlKey: true })
      .defaultPrevented,
  ).toBe(true);
  for (let i = 0; i < 2; i++) {
    expect(
      key("keydown", "ArrowRight", "ArrowRight", { ctrlKey: true })
        .defaultPrevented,
    ).toBe(true);
    key("keyup", "ArrowRight", "ArrowRight", { ctrlKey: true });
  }
  key("keydown", " ", "Space", { ctrlKey: true });
  key("keyup", " ", "Space", { ctrlKey: true });
  key("keydown", "F1", "F1", { ctrlKey: true });
  key("keyup", "F1", "F1", { ctrlKey: true });
  key("keyup", "Control", "ControlLeft");
  expect(s.steps).toEqual(["0/right", "0/right"]);
  expect(s.clicks).toEqual([0, 1]);
  expect(s.shortcuts).toEqual(["0/F1"]);
  expect(s.releases).toEqual([
    "0/ArrowRight",
    "0/ArrowRight",
    "0/ ",
    "0/F1",
    "1/ ",
  ]);
});

it("releases A keys pressed before B's modifier, including a pending Space click", () => {
  const s = setup();
  key("keydown", "ArrowRight");
  key("keydown", " ", "Space");
  key("keydown", "Control", "ControlLeft", { ctrlKey: true });
  key("keyup", "ArrowRight", "ArrowRight", { ctrlKey: true });
  key("keyup", " ", "Space", { ctrlKey: true });
  key("keyup", "Control", "ControlLeft");
  key("keydown", "ArrowRight");
  expect(s.steps).toEqual(["0/right", "0/right"]);
  expect(s.clicks).toEqual([0, 1]);
});

it.each([
  ["d", "KeyD", "info"],
  ["c", "KeyC", "open"],
  ["x", "KeyX", "mode"],
  ["v", "KeyV", "map"],
])(
  "uses B Control+%s for its native gesture and suppresses browser shortcuts",
  (value, code, action) => {
    const s = setup();
    key("keydown", "Control", "ControlLeft", { ctrlKey: true });
    expect(
      key("keydown", value, code, { ctrlKey: true }).defaultPrevented,
    ).toBe(true);
    key("keyup", value, code, { ctrlKey: true });
    key("keyup", "Control", "ControlLeft");
    expect(s.actions).toEqual([`1/${action}`]);
    expect(s.clicks).toEqual([]);
  },
);

it.each([
  ["single", 0],
  ["single", 1],
  ["hotseat", 0],
  ["hotseat", 1],
] as const)(
  "pans only the active camera with WASD and stops on release in %s side %s",
  (mode, active) => {
    const s = setup(mode, active),
      camera = s.cameras[active];
    if (camera === undefined) throw Error("Missing camera");
    const before = { ...camera.origin };
    for (const value of ["w", "a", "s", "d"]) {
      expect(
        key("keydown", value, `Key${value.toUpperCase()}`).defaultPrevented,
      ).toBe(true);
      const origin = { ...camera.origin };
      s.tick();
      expect(camera.origin).not.toEqual(origin);
      key("keyup", value, `Key${value.toUpperCase()}`);
      const released = { ...camera.origin };
      s.tick();
      expect(camera.origin).toEqual(released);
    }
    expect(s.cameraReleases).toEqual([
      `${active}/w`,
      `${active}/a`,
      `${active}/s`,
      `${active}/d`,
    ]);
    expect(s.steps).toEqual([]);
    expect(s.cameras[1 - active]?.origin).toEqual(before);
  },
);

it("keeps WASD camera panning disabled in simultaneous mode", () => {
  const s = setup(),
    before = s.cameras.map((camera) => ({ ...camera.origin }));
  for (const value of ["w", "a", "s", "d"]) {
    key("keydown", value, `Key${value.toUpperCase()}`);
    s.tick();
    key("keyup", value, `Key${value.toUpperCase()}`);
  }
  expect(s.cameras.map((camera) => camera.origin)).toEqual(before);
  expect(s.steps).toEqual(["1/up"]);
});

it("leaves right Control and unrelated browser modifier chords to the browser", () => {
  const s = setup();
  expect(
    key("keydown", "Control", "ControlRight", { ctrlKey: true })
      .defaultPrevented,
  ).toBe(false);
  for (const value of ["d", "x", "ArrowLeft", " ", "F1"])
    expect(
      key("keydown", value, value, { ctrlKey: true }).defaultPrevented,
    ).toBe(false);
  key("keyup", "Control", "ControlRight");
  for (const modifiers of [{ altKey: true }, { metaKey: true }])
    expect(
      key("keydown", "ArrowRight", "ArrowRight", modifiers).defaultPrevented,
    ).toBe(false);
  expect(s.steps).toEqual([]);
  expect(s.clicks).toEqual([]);
  expect(s.shortcuts).toEqual([]);
  expect(
    key("keydown", "Control", "ControlLeft", { ctrlKey: true })
      .defaultPrevented,
  ).toBe(true);
  expect(key("keydown", "l", "KeyL", { ctrlKey: true }).defaultPrevented).toBe(
    false,
  );
});

it("preserves editable fields and native button activation, including Take control", () => {
  const s = setup();
  for (const tag of ["INPUT", "TEXTAREA", "SELECT", "DIV"]) {
    const field = new ElementTarget(tag);
    field.isContentEditable = tag === "DIV";
    for (const value of ["w", "a", "s", "d", "m", "ArrowLeft", " "])
      expect(
        key("keydown", value, value, { target: field }).defaultPrevented,
      ).toBe(false);
  }
  for (const tag of ["BUTTON", "SUMMARY"]) {
    const target = new ElementTarget("SPAN");
    target.parentElement = new ElementTarget(tag);
    for (const value of [" ", "Enter", "ArrowRight"])
      expect(key("keydown", value, value, { target }).defaultPrevented).toBe(
        false,
      );
  }
  s.state.handover = true;
  for (const value of [" ", "Enter"])
    expect(
      key("keydown", value, value, { target: new ElementTarget("BUTTON") })
        .defaultPrevented,
    ).toBe(false);
  expect(s.steps).toEqual([]);
  expect(s.clicks).toEqual([]);
  expect(s.shortcuts).toEqual([]);
});

it.each([
  ["single", 0],
  ["single", 1],
  ["hotseat", 1],
] as const)(
  "keeps native function keys and Escape global in text controls for %s side %s",
  (mode, active) => {
    const s = setup(mode, active),
      shortcuts = ["F1", "F2", "F8", "F9", "F11", "Escape"];
    for (const tag of ["INPUT", "TEXTAREA", "SELECT", "DIV"]) {
      const target = new ElementTarget(tag);
      target.isContentEditable = tag === "DIV";
      for (const value of shortcuts) {
        expect(key("keydown", value, value, { target }).defaultPrevented).toBe(
          true,
        );
        expect(key("keyup", value, value, { target }).defaultPrevented).toBe(
          true,
        );
      }
    }
    const expected = Array.from({ length: 4 }, () =>
      shortcuts.map((value) => `${active}/${value}`),
    ).flat();
    expect(s.shortcuts).toEqual(expected);
    expect(s.releases).toEqual(expected);
    expect(s.clicks).toEqual([]);
    expect(s.steps).toEqual([]);
  },
);

it.each(["focusin", "keyup"])(
  "clears held camera and fire keys when text focus arrives through %s",
  (eventType) => {
    const s = setup("single"),
      target = new ElementTarget("INPUT"),
      before = s.cameras.map((camera) => ({ ...camera.origin }));
    key("keydown", " ", "Space");
    key("keydown", "w", "KeyW");
    if (eventType === "focusin") {
      const event = new Event("focusin");
      Object.defineProperty(event, "target", { value: target });
      win.dispatchEvent(event);
    }
    key("keyup", " ", "Space", { target });
    key("keyup", "w", "KeyW", { target });
    s.tick();
    expect(s.cameras.map((camera) => camera.origin)).toEqual(before);
    expect(s.clicks).toEqual([]);
    expect(s.state.resets).toBe(1);
    key("keydown", "Escape", "Escape", { target });
    key("keyup", "Escape", "Escape", { target });
    expect(s.shortcuts).toEqual(["0/Escape"]);
    expect(s.releases).toEqual(["0/Escape"]);
  },
);

it("preserves pointer context when focus moves without any owned keyboard hold", () => {
  const s = setup();
  for (const tag of [
    "BUTTON",
    "SUMMARY",
    "INPUT",
    "TEXTAREA",
    "SELECT",
    "DIV",
  ]) {
    const target = new ElementTarget(tag);
    target.isContentEditable = tag === "DIV";
    const event = new Event("focusin");
    Object.defineProperty(event, "target", { value: target });
    win.dispatchEvent(event);
  }
  // The host's reset callback also closes pointer rings; focus alone must
  // leave those contexts open after a touch hold synthesizes button focus.
  expect(s.state.resets).toBe(0);
  expect(s.clicks).toEqual([]);
  expect(s.actions).toEqual([]);
});

it("drops all held state on focus, blur and host reset without firing stale orders", () => {
  const s = setup();
  for (const trigger of ["focusin", "blur", "reset"]) {
    key("keydown", "Control", "ControlLeft", { ctrlKey: true });
    key("keydown", "x", "KeyX", { ctrlKey: true });
    if (trigger === "reset") s.input.reset();
    else {
      const event = new Event(trigger);
      if (trigger === "focusin")
        Object.defineProperty(event, "target", {
          value: new ElementTarget("INPUT"),
        });
      win.dispatchEvent(event);
    }
    key("keyup", "Control", "ControlLeft");
    expect(
      key("keydown", "d", "KeyD", { ctrlKey: true }).defaultPrevented,
    ).toBe(false);
  }
  expect(s.actions).toEqual([]);
  expect(s.clicks).toEqual([]);
  expect(s.state.resets).toBe(3);
});

it("blocks paused and handover orders and allows only unmodified Q to leave the handover", () => {
  const s = setup();
  key("keydown", " ", "Space");
  s.state.blocked = true;
  key("keyup", " ", "Space");
  key("keydown", "F1");
  s.state.blocked = false;
  s.state.handover = true;
  key("keydown", "Control", "ControlLeft", { ctrlKey: true });
  key("keydown", "q", "KeyQ", { ctrlKey: true });
  key("keydown", "q", "KeyQ");
  key("keydown", "q", "KeyQ", { repeat: true });
  expect(s.clicks).toEqual([]);
  expect(s.shortcuts).toEqual([]);
  expect(s.state.leaves).toBe(1);
  s.state.handover = false;
  key("keyup", " ", "Space");
  expect(s.clicks).toEqual([]);
});

it("releases an owned key despite a later modifier and removes listeners on dispose", () => {
  const s = setup("single");
  key("keydown", "w", "KeyW");
  key("keyup", "w", "KeyW", { ctrlKey: true });
  expect(s.cameraReleases).toEqual(["0/w"]);
  s.input.dispose();
  const resets = s.state.resets;
  s.input.dispose();
  key("keydown", "F1");
  win.dispatchEvent(new Event("blur"));
  expect(s.state.resets).toBe(resets);
  expect(s.shortcuts).toEqual([]);
  const next = setup("single");
  key("keydown", "F1");
  expect(next.shortcuts).toEqual(["0/F1"]);
  expect(s.shortcuts).toEqual([]);
});
