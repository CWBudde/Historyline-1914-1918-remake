/**
 * The order menu round a unit (docs/dos-screen.md "Web order menu"): its own orders lie as painted hex plates on
 * the six neighbouring cells, one fixed slot per order, the unit stays visible in the middle. Orders that do not
 * apply leave their cell free; orders that apply but are not possible now stay as a stripped plate with the reason.
 * End phase, map and info are not here: they are the nameplates of the order strip. Pure geometry and tables; the
 * battle runtime places it, BattleUI draws it.
 */
import { hexgrid } from "@historyline/engine";
import type { Point, Rect } from "../camera.ts";
import type { Action } from "./ring.ts";

/** A neighbouring cell of the unit, clockwise from the top (flat-topped hexes: no left or right neighbour). */
export type Slot = "n" | "ne" | "se" | "s" | "sw" | "nw";
export const slots: readonly Slot[] = ["n", "ne", "se", "s", "sw", "nw"];
/** Where the menu is hit: a slot's wedge, the centre disc over the unit (cancel), or nothing (the map). */
export type Hit = Slot | "centre" | null;

/** Each slot's neighbour centre relative to the unit's, in world px (hexgrid.around). */
const offsets: Readonly<Record<Slot, Point>> = {
  n: { x: 0, y: -2 * hexgrid.linePitch },
  ne: { x: hexgrid.halfPitch, y: -hexgrid.linePitch },
  se: { x: hexgrid.halfPitch, y: hexgrid.linePitch },
  s: { x: 0, y: 2 * hexgrid.linePitch },
  sw: { x: -hexgrid.halfPitch, y: hexgrid.linePitch },
  nw: { x: -hexgrid.halfPitch, y: -hexgrid.linePitch },
};
/** The keys of each slot: the hex on the left hand (Q W E over A S D) and on the numeric keypad (7 8 9 over 1 2 3). */
export const slotKeys: Readonly<Record<Slot, readonly [string, string]>> = {
  nw: ["q", "7"],
  n: ["w", "8"],
  ne: ["e", "9"],
  sw: ["a", "1"],
  s: ["s", "2"],
  se: ["d", "3"],
};
/** slotOfKey is the slot a key runs while the menu is open, or null. */
export function slotOfKey(key: string): Slot | null {
  const k = key.toLowerCase();
  for (const s of slots) if (slotKeys[s].includes(k)) return s;
  return null;
}

/**
 * slotTable is the order on each slot. On the map: up moves the unit (attacks in the attack phase), up-left opens
 * an own container. In a container's contents: up moves the picked unit out, up-left produces, down repairs. The
 * directions keep the desktop gestures where they still apply (up move, left open or produce, down repair).
 */
export function slotTable(
  contents: boolean,
  attack: boolean,
): Partial<Record<Slot, Action>> {
  return contents
    ? { n: "exit", nw: "produce", s: "repair" }
    : { n: attack ? "attack" : "move", nw: "open" };
}

/** An order on the menu: ready when reason is null, otherwise not possible now and why. */
export interface MenuItem {
  slot: Slot;
  action: Action;
  reason: string | null;
}
/**
 * menuItems are the table's orders that apply here (applies), each with the reason it is not possible now
 * (refusal, null when it is), in slot order.
 */
export function menuItems(
  table: Partial<Record<Slot, Action>>,
  applies: (a: Action) => boolean,
  refusal: (a: Action) => string | null,
): MenuItem[] {
  const out: MenuItem[] = [];
  for (const slot of slots) {
    const action = table[slot];
    if (action === undefined || !applies(action)) continue;
    out.push({ slot, action, reason: refusal(action) });
  }
  return out;
}

/** The plates' size rule in CSS px: a share of the cell, never a hit target under 48 px. */
export const plateRule = {
  mouse: { share: 0.7, min: 72, max: 160 },
  touch: { share: 0.92, min: 58, max: 200 },
};
/** A plate's width over its height: the hex cell's (96 × 80). */
export const plateAspect = hexgrid.hexWidth / hexgrid.hexHeight;

export interface PlacedItem extends MenuItem, Point {}
/** MenuLayout is the menu in viewport CSS px. */
export interface MenuLayout {
  /** The unit's centre, where the cursor ring stays. */
  unit: Point;
  /** The rosette's centre: the unit's, or moved inwards at a window edge. */
  centre: Point;
  plate: { width: number; height: number };
  /** The distance of a neighbour in CSS px (the wedges' unit). */
  step: number;
  items: PlacedItem[];
  touch: boolean;
  /** Where the menu may lie: the map's part of the window. */
  bounds: Rect;
}

/** plateWidth is the plate width for a hex cell cellWidth CSS px wide. */
export function plateWidth(cellWidth: number, touch: boolean): number {
  const r = touch ? plateRule.touch : plateRule.mouse;
  return Math.min(r.max, Math.max(r.min, r.share * cellWidth));
}

/**
 * layoutMenu places items round the unit at unit (CSS px) for hex cells cellWidth CSS px wide. Plates sit on the
 * neighbour centres; where the size rule makes them larger than the cell allows, the rosette grows past the
 * neighbours. If a plate would leave bounds the whole rosette moves inwards, its shape and directions intact.
 */
export function layoutMenu(
  items: readonly MenuItem[],
  unit: Point,
  cellWidth: number,
  touch: boolean,
  bounds: Rect,
): MenuLayout {
  if (!(cellWidth > 0)) throw new Error("invalid cell width");
  const width = plateWidth(cellWidth, touch),
    height = width / plateAspect,
    scale = cellWidth / hexgrid.hexWidth,
    grow = Math.max(1, width / (0.92 * cellWidth)),
    k = scale * grow;
  const shift = fitShift(
    items.map((i) => ({
      x: unit.x + offsets[i.slot].x * k,
      y: unit.y + offsets[i.slot].y * k,
    })),
    width,
    height,
    bounds,
  );
  const centre = { x: unit.x + shift.x, y: unit.y + shift.y };
  return {
    unit,
    centre,
    plate: { width, height },
    step: 2 * hexgrid.linePitch * k,
    items: items.map((i) => ({
      ...i,
      x: centre.x + offsets[i.slot].x * k,
      y: centre.y + offsets[i.slot].y * k,
    })),
    touch,
    bounds,
  };
}

/**
 * fitShift is the smallest move that keeps every plate (centres, size) inside bounds; where they cannot fit, the
 * top left wins.
 */
export function fitShift(
  centres: readonly Point[],
  width: number,
  height: number,
  bounds: Rect,
): Point {
  if (centres.length === 0) return { x: 0, y: 0 };
  const x0 = Math.min(...centres.map((c) => c.x)) - width / 2,
    x1 = Math.max(...centres.map((c) => c.x)) + width / 2,
    y0 = Math.min(...centres.map((c) => c.y)) - height / 2,
    y1 = Math.max(...centres.map((c) => c.y)) + height / 2;
  const into = (lo: number, hi: number, a: number, b: number) =>
    lo < a ? a - lo : hi > b ? Math.max(b - hi, a - lo) : 0;
  return {
    x: into(x0, x1, bounds.x, bounds.x + bounds.width),
    y: into(y0, y1, bounds.y, bounds.y + bounds.height),
  };
}

/** slotAt is the slot whose 60° wedge holds the direction (dx, dy), screen y down. */
export function slotAt(dx: number, dy: number): Slot {
  // Degrees clockwise from up; each slot's wedge is centred on a multiple of 60.
  const deg = ((Math.atan2(dx, -dy) * 180) / Math.PI + 360 + 30) % 360;
  return slots[Math.floor(deg / 60) % 6] ?? "n";
}

/** The wedges reach from half a neighbour's distance (inside: the centre disc) to this many distances. */
export const wedgeReach = 1.75;

/** hitMenu is what the point p hits: a slot's wedge, the centre disc, or nothing. Empty slots are hit too. */
export function hitMenu(m: MenuLayout, p: Point): Hit {
  const dx = p.x - m.centre.x,
    dy = p.y - m.centre.y,
    r = Math.hypot(dx, dy) / m.step;
  if (r < 0.5) return "centre";
  if (r > wedgeReach) return null;
  return slotAt(dx, dy);
}
/** itemAt is the menu item on slot s, if any. */
export function itemAt(m: MenuLayout, s: Hit): PlacedItem | null {
  return s === null || s === "centre"
    ? null
    : (m.items.find((i) => i.slot === s) ?? null);
}

/** A flick is a drag from the unit at least this far (CSS px); shorter is a tap. */
export const flickDistance = 24;
/** flickSlot is the slot a drag from from to to points at, or null while it is shorter than a flick. */
export function flickSlot(from: Point, to: Point): Slot | null {
  const dx = to.x - from.x,
    dy = to.y - from.y;
  return Math.hypot(dx, dy) < flickDistance ? null : slotAt(dx, dy);
}

/**
 * walk is where an arrow key moves the keyboard focus from at: the nearest plate (or the centre) that lies that
 * way, angles counting double against distance; at itself if nothing lies that way.
 */
export function walk(
  m: MenuLayout,
  at: Slot | "centre",
  key: "ArrowUp" | "ArrowDown" | "ArrowLeft" | "ArrowRight",
): Slot | "centre" {
  const pos = (s: Slot | "centre"): Point =>
      s === "centre" ? m.centre : (itemAt(m, s) ?? m.centre),
    from = pos(at),
    [ux, uy] =
      key === "ArrowUp"
        ? [0, -1]
        : key === "ArrowDown"
          ? [0, 1]
          : key === "ArrowLeft"
            ? [-1, 0]
            : [1, 0];
  let best: Slot | "centre" = at,
    score = Infinity;
  for (const s of [...m.items.map((i) => i.slot), "centre" as const]) {
    if (s === at) continue;
    const p = pos(s),
      dx = p.x - from.x,
      dy = p.y - from.y,
      along = dx * ux + dy * uy,
      across = Math.abs(dx * uy - dy * ux);
    if (along <= 1e-6) continue;
    const v = along + 2 * across;
    if (v < score) {
      score = v;
      best = s;
    }
  }
  return best;
}

/** Where the caption plate hangs: its anchor point and the side of it the plate lies on. */
export interface CaptionPlace extends Point {
  side: "above" | "below" | "left" | "right";
}
/** The caption plate's height in CSS px, with a reason line (for the room it needs above or below). */
const captionRoom = 52;
/**
 * captionPlace is where the nameplate of what is lit hangs: with a mouse from the lit plate's outer edge (beside
 * the side plates, above or below the upright ones), on touch above the rosette, clear of the thumb, and for the
 * centre above the rosette too; below it where the map has no room above. width is the caption's width, to keep
 * it inside bounds (the map's part of the window).
 */
export function captionPlace(
  m: MenuLayout,
  lit: Slot | "centre",
  width: number,
  bounds: Rect = m.bounds,
): CaptionPlace {
  const gap = 6,
    top =
      Math.min(m.centre.y - m.step / 2, ...m.items.map((i) => i.y)) -
      m.plate.height / 2 -
      gap,
    bottom =
      Math.max(m.centre.y + m.step / 2, ...m.items.map((i) => i.y)) +
      m.plate.height / 2 +
      gap,
    above = (x: number): CaptionPlace => ({
      x: clamp(x, bounds.x + width / 2, bounds.x + bounds.width - width / 2),
      y: top - captionRoom < bounds.y ? bottom : top,
      side: top - captionRoom < bounds.y ? "below" : "above",
    });

  const item = lit === "centre" ? null : itemAt(m, lit);
  if (m.touch || item === null) return above(m.centre.x);
  const hw = m.plate.width / 2,
    hh = m.plate.height / 2;
  switch (item.slot) {
    case "n":
      return above(item.x);
    case "s": {
      const y = item.y + hh + gap;
      return y + 40 > bounds.y + bounds.height
        ? above(item.x)
        : {
            x: clamp(
              item.x,
              bounds.x + width / 2,
              bounds.x + bounds.width - width / 2,
            ),
            y,
            side: "below",
          };
    }
    case "ne":
    case "se":
      return item.x + hw + gap + width > bounds.x + bounds.width
        ? above(item.x)
        : { x: item.x + hw + gap, y: item.y, side: "right" };
    case "nw":
    case "sw":
      return item.x - hw - gap - width < bounds.x
        ? above(item.x)
        : { x: item.x - hw - gap, y: item.y, side: "left" };
  }
}
function clamp(v: number, lo: number, hi: number): number {
  return lo > hi ? (lo + hi) / 2 : Math.min(hi, Math.max(lo, v));
}
