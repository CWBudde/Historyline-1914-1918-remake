import { expect, it } from "vitest";
import {
  captionPlace,
  fitShift,
  flickSlot,
  hitMenu,
  itemAt,
  layoutMenu,
  menuItems,
  plateWidth,
  slotAt,
  slotOfKey,
  slotTable,
  walk,
  type MenuItem,
} from "./orderMenu.ts";

const window1080 = { x: 0, y: 0, width: 1920, height: 1080 };

it("puts each order on its fixed slot: up moves or attacks, up-left opens, contents add repair down", () => {
  expect(slotTable(false, false)).toEqual({ n: "move", nw: "open" });
  expect(slotTable(false, true)).toEqual({ n: "attack", nw: "open" });
  expect(slotTable(true, false)).toEqual({
    n: "exit",
    nw: "produce",
    s: "repair",
  });
});

it("leaves orders that do not apply absent and keeps impossible ones with their reason", () => {
  const items = menuItems(
    slotTable(true, false),
    (a) => a !== "produce",
    (a) => (a === "repair" ? "Nothing to repair." : null),
  );
  expect(items).toEqual([
    { slot: "n", action: "exit", reason: null },
    { slot: "s", action: "repair", reason: "Nothing to repair." },
  ]);
});

it("sizes plates by the cell: 0.7 with a mouse within 72-160 px, 0.92 on touch, never under 58 px wide", () => {
  expect(plateWidth(142, false)).toBeCloseTo(99.4);
  expect(plateWidth(60, false)).toBe(72);
  expect(plateWidth(300, false)).toBe(160);
  expect(plateWidth(66, true)).toBeCloseTo(60.72);
  expect(plateWidth(40, true)).toBe(58);
  // A plate is the cell's shape, so a 58 px plate is still 48 px across the flats.
  expect(58 / 1.2).toBeGreaterThanOrEqual(48);
});

const two: MenuItem[] = [
  { slot: "n", action: "move", reason: null },
  { slot: "nw", action: "open", reason: "Empty." },
];

it("places plates on the neighbour centres and grows the rosette past them when plates outgrow the cell", () => {
  const m = layoutMenu(two, { x: 960, y: 540 }, 142, false, window1080);
  const s = 142 / 96;
  expect(m.centre).toEqual({ x: 960, y: 540 });
  expect(itemAt(m, "n")).toMatchObject({ x: 960, y: 540 - 80 * s });
  expect(itemAt(m, "nw")).toMatchObject({ x: 960 - 71 * s, y: 540 - 40 * s });
  expect(m.step).toBeCloseTo(80 * s);
  // A 60 px phone cell with 58 px touch plates: the plates would overlap, so the neighbours move out.
  const phone = layoutMenu(two, { x: 200, y: 400 }, 50, true, {
    x: 0,
    y: 0,
    width: 390,
    height: 844,
  });
  expect(phone.plate.width).toBe(58);
  expect(400 - (itemAt(phone, "n")?.y ?? 0)).toBeGreaterThan(
    phone.plate.height,
  );
});

it("hits wedges, not plates: the centre disc cancels, empty wedges and the far map hit no item", () => {
  const m = layoutMenu(two, { x: 960, y: 540 }, 142, false, window1080);
  expect(hitMenu(m, { x: 960, y: 540 })).toBe("centre");
  expect(hitMenu(m, { x: 960 + 10, y: 540 - m.step * 0.4 })).toBe("centre");
  // Up, well outside the plate's own outline but inside the wedge.
  expect(hitMenu(m, { x: 960 + 30, y: 540 - m.step * 1.6 })).toBe("n");
  expect(hitMenu(m, { x: 960 - 60, y: 540 - 40 })).toBe("nw");
  expect(itemAt(m, hitMenu(m, { x: 960 + 80, y: 540 + 30 }))).toBeNull();
  expect(hitMenu(m, { x: 960, y: 540 - m.step * 2 })).toBeNull();
  // Every wedge of every plate is wider than 48 px at half its depth, even on a small phone cell.
  const phone = layoutMenu(two, { x: 200, y: 400 }, 50, true, {
    x: 0,
    y: 0,
    width: 390,
    height: 844,
  });
  const mid = ((0.5 + 1.75) / 2) * phone.step;
  expect((mid * Math.PI) / 3).toBeGreaterThan(48);
});

it("reads a flick by its direction in six 60° wedges once it is 24 px long", () => {
  const o = { x: 100, y: 100 };
  expect(flickSlot(o, { x: 100, y: 80 })).toBeNull();
  expect(flickSlot(o, { x: 102, y: 60 })).toBe("n");
  expect(flickSlot(o, { x: 140, y: 75 })).toBe("ne");
  expect(flickSlot(o, { x: 140, y: 125 })).toBe("se");
  expect(flickSlot(o, { x: 95, y: 140 })).toBe("s");
  expect(flickSlot(o, { x: 60, y: 125 })).toBe("sw");
  expect(flickSlot(o, { x: 60, y: 75 })).toBe("nw");
  expect(slotAt(1, -2)).toBe("n");
  expect(slotAt(1, -0.5)).toBe("ne");
});

it("shifts the whole rosette inwards at a window edge without mirroring a direction", () => {
  const m = layoutMenu(two, { x: 20, y: 30 }, 142, false, window1080);
  expect(m.unit).toEqual({ x: 20, y: 30 });
  for (const i of m.items) {
    expect(i.x - m.plate.width / 2).toBeGreaterThanOrEqual(-1e-9);
    expect(i.y - m.plate.height / 2).toBeGreaterThanOrEqual(-1e-9);
  }
  // Same shape: up is still straight above the centre, up-left still above and left of it.
  expect(itemAt(m, "n")?.x).toBeCloseTo(m.centre.x);
  expect(itemAt(m, "n")?.y ?? 0).toBeLessThan(m.centre.y);
  expect(itemAt(m, "nw")?.x ?? 0).toBeLessThan(m.centre.x);
  // Hits follow the moved rosette.
  expect(hitMenu(m, m.centre)).toBe("centre");
  expect(fitShift([{ x: 500, y: 500 }], 100, 80, window1080)).toEqual({
    x: 0,
    y: 0,
  });
  expect(fitShift([{ x: 1900, y: 1070 }], 100, 80, window1080)).toEqual({
    x: -30,
    y: -30,
  });
});

it("maps the hex keys and the keypad to the slots", () => {
  expect(slotOfKey("W")).toBe("n");
  expect(slotOfKey("q")).toBe("nw");
  expect(slotOfKey("3")).toBe("se");
  expect(slotOfKey("x")).toBeNull();
});

it("walks the keyboard focus to the plate that lies the arrow's way", () => {
  const three = layoutMenu(
    menuItems(
      slotTable(true, false),
      () => true,
      () => null,
    ),
    { x: 960, y: 540 },
    142,
    false,
    window1080,
  );
  expect(walk(three, "centre", "ArrowUp")).toBe("n");
  expect(walk(three, "centre", "ArrowDown")).toBe("s");
  expect(walk(three, "n", "ArrowLeft")).toBe("nw");
  expect(walk(three, "nw", "ArrowRight")).toBe("n");
  expect(walk(three, "s", "ArrowDown")).toBe("s");
});

it("hangs the caption beside side plates, above upright ones, and above the rosette on touch", () => {
  const items = menuItems(
    slotTable(true, false),
    () => true,
    () => null,
  );
  const m = layoutMenu(items, { x: 960, y: 540 }, 142, false, window1080);
  expect(captionPlace(m, "nw", 120, window1080).side).toBe("left");
  expect(captionPlace(m, "n", 120, window1080).side).toBe("above");
  expect(captionPlace(m, "s", 120, window1080).side).toBe("below");
  // No room on the left: above instead.
  const edge = layoutMenu(items, { x: 120, y: 540 }, 142, false, window1080);
  expect(captionPlace(edge, "nw", 160, window1080).side).toBe("above");
  const touch = layoutMenu(items, { x: 200, y: 500 }, 66, true, {
    x: 0,
    y: 0,
    width: 390,
    height: 844,
  });
  const c = captionPlace(touch, "s", 120, {
    x: 0,
    y: 0,
    width: 390,
    height: 844,
  });
  expect(c.side).toBe("above");
  expect(c.y).toBeLessThan(Math.min(...touch.items.map((i) => i.y)));
});

it("puts the caption below the rosette where the map has no room above it", () => {
  const phone = { x: 0, y: 120, width: 412, height: 500 };
  const m = layoutMenu(
    [{ slot: "n", action: "move", reason: null }],
    { x: 300, y: 200 },
    66,
    true,
    phone,
  );
  const c = captionPlace(m, "n", 120);
  expect(c.side).toBe("below");
  expect(c.y).toBeGreaterThan(m.centre.y);
});
