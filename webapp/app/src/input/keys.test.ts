import { expect, it } from "vitest";
import type { Action, Direction, Placement } from "../ui/ring.ts";
import { BattleKeys } from "./keys.ts";

function fixture() {
  const events: string[] = [],
    shortcuts: string[] = [];
  const state: {
    placement: Partial<Placement>;
    shortcut: (key: string) => boolean;
  } = {
    placement: {
      centre: "cancel",
      up: "move",
      left: "open",
      right: "map",
      down: "info",
    },
    shortcut: (key) => key === "F1",
  };
  const keys = new BattleKeys({
    step: (dir: Direction) => {
      events.push(`step:${dir}`);
    },
    click: () => {
      events.push("click");
    },
    run: (action: Action) => {
      events.push(`run:${action}`);
    },
    placement: () => state.placement,
    shortcut: (key) => {
      shortcuts.push(key);
      return state.shortcut(key);
    },
  });
  return { keys, events, shortcuts, state };
}

it("steps with fresh arrows and browser repeats but ignores duplicate nonrepeat downs", () => {
  const { keys, events } = fixture();
  expect(keys.down("ArrowUp")).toBe(true);
  keys.down("ArrowUp");
  keys.down("ArrowUp", true);
  expect(keys.up("ArrowUp")).toBe(true);
  keys.down("ArrowLeft");
  keys.up("ArrowLeft");
  keys.down("ArrowRight");
  keys.up("ArrowRight");
  keys.down("ArrowDown");
  expect(events).toEqual([
    "step:up",
    "step:up",
    "step:left",
    "step:right",
    "step:down",
  ]);
});
it("clicks a plain Space tap once regardless of repeats or duplicate releases", () => {
  const { keys, events } = fixture();
  expect(keys.down(" ")).toBe(true);
  keys.down(" ", true);
  keys.down(" ");
  expect(events).toEqual([]);
  expect(keys.up(" ")).toBe(true);
  keys.up(" ");
  expect(events).toEqual(["click"]);
});
it("runs a picked action on fire release even after the arrow was released", () => {
  const { keys, events } = fixture();
  keys.down(" ");
  keys.down("ArrowUp");
  keys.up("ArrowUp");
  expect(events).toEqual([]);
  keys.up(" ");
  expect(events).toEqual(["run:move"]);
});
it.each([
  ["ArrowUp", "ArrowDown"],
  ["ArrowDown", "ArrowUp"],
  ["ArrowLeft", "ArrowRight"],
  ["ArrowRight", "ArrowLeft"],
])(
  "returns from %s to centre on opposite %s and cancels rather than clicks",
  (first, opposite) => {
    const { keys, events } = fixture();
    keys.down(" ");
    keys.down(first);
    keys.down(opposite);
    keys.up(" ");
    expect(events).toEqual(["run:cancel"]);
  },
);
it("jumps to perpendicular directions, using newly pressed arrows rather than held priority", () => {
  const { keys, events } = fixture();
  keys.down(" ");
  keys.down("ArrowUp");
  keys.down("ArrowLeft");
  keys.down("ArrowDown");
  keys.up(" ");
  expect(events).toEqual(["run:info"]);
});
it("does not treat an arrow held before fire as a gesture pick", () => {
  const { keys, events } = fixture();
  keys.down("ArrowDown");
  keys.down(" ");
  keys.down("ArrowDown", true);
  keys.up(" ");
  expect(events).toEqual(["step:down", "click"]);
  keys.down("ArrowDown", true);
  expect(events).toEqual(["step:down", "click", "step:down"]);
});
it("never repicks on repeat, but a released arrow may be picked again", () => {
  const { keys, events } = fixture();
  keys.down(" ");
  keys.down("ArrowUp");
  keys.down("ArrowDown");
  keys.down("ArrowUp", true);
  keys.up("ArrowUp");
  keys.down("ArrowUp");
  keys.up(" ");
  expect(events).toEqual(["run:move"]);
});
it.each([undefined, "cancel"] as const)(
  "cancels a picked action whose placement is %s",
  (action) => {
    const { keys, events, state } = fixture();
    state.placement = action === undefined ? {} : { up: action };
    keys.down(" ");
    keys.down("ArrowUp");
    keys.up(" ");
    expect(events).toEqual(["run:cancel"]);
  },
);
it("checks placement at release so newly disabled actions cannot run", () => {
  const { keys, events, state } = fixture();
  keys.down(" ");
  keys.down("ArrowUp");
  state.placement = {};
  keys.up(" ");
  expect(events).toEqual(["run:cancel"]);
});
it("allows an action enabled before release, matching the native tracker", () => {
  const { keys, events, state } = fixture();
  state.placement = {};
  keys.down(" ");
  keys.down("ArrowUp");
  state.placement = { up: "attack" };
  keys.up(" ");
  expect(events).toEqual(["run:attack"]);
});
it("plain fire still clicks when the current placement has no directions", () => {
  const { keys, events, state } = fixture();
  state.placement = {};
  keys.down(" ");
  keys.up(" ");
  expect(events).toEqual(["click"]);
});
it("Enter clicks once on keydown and becomes available again after release", () => {
  const { keys, events } = fixture();
  expect(keys.down("Enter")).toBe(true);
  keys.down("Enter", true);
  keys.down("Enter");
  keys.up("Enter");
  keys.up("Enter");
  keys.down("Enter");
  expect(events).toEqual(["click", "click"]);
});
it("forwards shortcuts and preserves handled flags without duplicating held actions", () => {
  const { keys, shortcuts } = fixture();
  expect(keys.down("F1")).toBe(true);
  expect(keys.down("F1", true)).toBe(true);
  expect(keys.down("F1")).toBe(true);
  expect(keys.up("F1")).toBe(true);
  expect(keys.down("x")).toBe(false);
  expect(keys.down("x", true)).toBe(false);
  expect(keys.up("x")).toBe(false);
  expect(keys.up("unknown")).toBe(false);
  expect(shortcuts).toEqual(["F1", "x"]);
});
it("leaves native shortcut policy to the callback and preserves original key strings", () => {
  const { keys, shortcuts, events, state } = fixture();
  state.shortcut = () => true;
  for (const key of [
    "F1",
    "F2",
    "F8",
    "F9",
    "F11",
    "M",
    "q",
    "Escape",
    "Home",
    "End",
    "PageUp",
    "PageDown",
  ]) {
    expect(keys.down(key)).toBe(true);
    expect(keys.up(key)).toBe(true);
  }
  expect(shortcuts).toEqual([
    "F1",
    "F2",
    "F8",
    "F9",
    "F11",
    "M",
    "q",
    "Escape",
    "Home",
    "End",
    "PageUp",
    "PageDown",
  ]);
  expect(events).toEqual([]);
});
it("deduplicates a held letter across Shift changes without rewriting callback arguments", () => {
  const { keys, shortcuts, state } = fixture();
  state.shortcut = () => true;
  expect(keys.down("M")).toBe(true);
  expect(keys.down("m", true)).toBe(true);
  expect(keys.up("m")).toBe(true);
  keys.down("m");
  expect(shortcuts).toEqual(["M", "m"]);
});
it("cancel drops a pending hold silently and ignores stale repeat events until a fresh press", () => {
  const { keys, events } = fixture();
  keys.down(" ");
  keys.down("ArrowUp");
  keys.cancel();
  keys.up(" ");
  keys.up("ArrowUp");
  keys.down(" ", true);
  keys.up(" ");
  keys.down("Enter", true);
  keys.up("Enter");
  keys.down("ArrowUp", true);
  keys.up("ArrowUp");
  expect(events).toEqual([]);
  keys.down(" ");
  keys.up(" ");
  expect(events).toEqual(["click"]);
});
it("supports parent context cancellation from a shortcut without later running the old pick", () => {
  const { keys, events, state } = fixture();
  state.shortcut = (key) => {
    if (key !== "Escape") return false;
    keys.cancel();
    return true;
  };
  keys.down(" ");
  keys.down("ArrowLeft");
  expect(keys.down("Escape")).toBe(true);
  keys.up(" ");
  expect(events).toEqual([]);
  keys.down("Escape", true);
  keys.down(" ");
  keys.up(" ");
  expect(events).toEqual(["click"]);
});
it("accepts the browser Space key and code spelling as the same held key", () => {
  const { keys, events } = fixture();
  keys.down("Space");
  keys.down(" ", true);
  keys.up(" ");
  expect(events).toEqual(["click"]);
});
it("does not implement the second human's D/C/X/V/control key set", () => {
  const { keys, events } = fixture();
  for (const key of ["d", "c", "x", "v", "Control"]) {
    expect(keys.down(key)).toBe(false);
    keys.up(key);
  }
  expect(events).toEqual([]);
});
it("resets fire state before callbacks so a new gesture may begin inside a run callback", () => {
  const events: string[] = [];
  const keys: BattleKeys = new BattleKeys({
    step: () => {},
    click: () => {
      events.push("click");
    },
    run: (action) => {
      events.push(action);
      keys.down(" ");
    },
    placement: () => ({ up: "move" }),
    shortcut: () => false,
  });
  keys.down(" ");
  keys.down("ArrowUp");
  keys.up(" ");
  keys.up(" ");
  expect(events).toEqual(["move", "click"]);
});
