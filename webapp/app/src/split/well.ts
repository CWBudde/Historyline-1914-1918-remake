/**
 * The status wells under the split screen's panels (cmd/historyline/well.go): one line of capitals per side in the
 * panel's colour, numbers in amber. With the cursor on an own unit the human side's well shows the unit line
 * (group size, experience badge, name, the owner's flag); otherwise each well says what its side is doing, and
 * while the computer plans a thin bar along the bottom shows how far it got.
 */
import type { Rect } from "../camera.ts";
import type { PixelFace } from "../pixfont.ts";

/** How long a fresh message beats the unit line (Go: 180 frames). */
export const messageMs = 3000;

/** The status line's colours: pale green on the left, cream on the right, amber numbers, a grey bar. */
export const wellText = ["#bed392", "#fbe3c7"] as const;
export const wellNumbers = "#e7d365";
export const wellBar = "#7d7d7d";

export interface UnitLine {
  size: number;
  exp: number;
  name: string;
  /** The owner: 1 the Central Powers, 2 the Entente. */
  owner: number;
}
/** Well is what one status well shows: a line or the unit line, and the planning progress (negative: no bar). */
export interface Well {
  text: string;
  unit: UnitLine | null;
  progress: number;
}

/** WellInput is what the wells are worded from, per side. */
export interface WellInput {
  /** The prompt or phase-change line, already worded, or null. */
  stage: string | null;
  /** The end of the battle in the side's words, or null while it is open. */
  ended: string | null;
  /** The side's planning progress from 0 to 1 while the computer plans for it, or null. */
  planning: number | null;
  /** Whether the side's attack phase is being planned (weighing the attacks). */
  attack: boolean;
  /** Whether orders are being carried out on the map. */
  busy: boolean;
  /** Whether the side has confirmed its orders, and the other side's phase is attack. */
  ready: boolean;
  enemyAttacks: boolean;
  /** Whether the side is the one the player commands now. */
  human: boolean;
  /** A fresh message to the player, or "". */
  message: string;
  /** The player's own unit under the cursor, or null. */
  unit: UnitLine | null;
  /** The orders given this phase and the limit (0: none). */
  orders: number;
  limit: number;
}

const modeWord = (attack: boolean) => (attack ? "ATTACK" : "MOVEMENT");
function count(n: number, thing: string): string {
  return n === 1 ? `1 ${thing}` : `${n} ${thing}S`;
}

/** wellOf words one side's well as the desktop does (wellOf). */
export function wellOf(w: WellInput): Well {
  const out: Well = { text: "", unit: null, progress: -1 };
  if (w.stage !== null && w.ended === null) {
    out.text = w.stage.toUpperCase();
    return out;
  }
  if (w.ended !== null) out.text = w.ended;
  else if (w.planning !== null) {
    out.progress = w.planning;
    out.text = w.attack ? "WEIGHING THE ATTACKS" : "PLANNING THE MOVES";
  } else if (w.busy) out.text = "ORDERS UNDER WAY";
  else if (w.ready)
    out.text = w.enemyAttacks
      ? "WAITING: ENEMY ATTACKS"
      : "WAITING: ENEMY MOVES";
  else if (!w.human) out.text = `${modeWord(w.attack)} PHASE`;
  else if (w.message !== "") out.text = w.message;
  else if (w.unit !== null)
    out.unit = { ...w.unit, name: w.unit.name.toUpperCase() };
  else
    out.text = `${modeWord(w.attack)} PHASE  ${
      w.limit > 0 ? `${w.orders}/${w.limit} ORDERS` : count(w.orders, "ORDER")
    }`;
  out.text = out.text.toUpperCase();
  return out;
}

/** wellScale is the font scale of the wells: twice the UI scale, as large as the text area holds. */
export function wellScale(text: Rect, ui: number, face: PixelFace): number {
  return Math.max(
    1,
    Math.min(Math.floor(2 * ui + 1e-9), Math.trunc(text.height / face.height)),
  );
}

/** Flag bands: the German Empire's black, white and red across, the French blue, white and red upright. */
const flagCentral = ["#101010", "#eeeee6", "#c82020"],
  flagEntente = ["#2038a0", "#eeeee6", "#c82020"];

/** drawFlag draws owner's flag in the w × h rectangle at (x, y). */
export function drawFlag(
  ctx: CanvasRenderingContext2D,
  owner: number,
  x: number,
  y: number,
  w: number,
  h: number,
): void {
  const across = owner !== 2,
    bands = across ? flagCentral : flagEntente;
  for (const [i, c] of bands.entries()) {
    ctx.fillStyle = c;
    if (across) {
      const y0 = y + Math.trunc((h * i) / 3),
        y1 = y + Math.trunc((h * (i + 1)) / 3);
      ctx.fillRect(x, y0, w, y1 - y0);
    } else {
      const x0 = x + Math.trunc((w * i) / 3),
        x1 = x + Math.trunc((w * (i + 1)) / 3);
      ctx.fillRect(x0, y, x1 - x0, h);
    }
  }
}

/**
 * drawWell draws w into a well whose text area and bar track are given in the context's device pixels (drawWell):
 * the unit line or a centred line, then the progress bar. badge draws the experience badge centred on (cx, cy).
 */
export function drawWell(
  ctx: CanvasRenderingContext2D,
  face: PixelFace,
  side: 0 | 1,
  text: Rect,
  bar: Rect,
  ui: number,
  w: Well,
  badge: (exp: number, cx: number, cy: number, size: number) => void,
): void {
  const k = wellScale(text, ui, face),
    colour = wellText[side],
    th = face.height * k,
    y = text.y + Math.trunc((text.height - th) / 2);
  if (w.unit !== null) {
    const size = String(w.unit.size),
      fh = th,
      fw = Math.trunc((fh * 8) / 5);
    face.draw(ctx, size, text.x + k, y, k, wellNumbers);
    let sw = face.measure(size, k).width;
    if (w.unit.exp > 0) {
      const bs = Math.min(th * 1.8, text.height);
      badge(w.unit.exp, text.x + sw + 3 * k + bs / 2, y + th / 2, bs);
      sw += 3 * k + Math.trunc(bs);
    }
    const room = text.width - 2 * Math.max(sw + 2 * k, fw + 2 * k),
      name = face.fit(w.unit.name, k, room),
      nw = face.measure(name, k).width;
    face.draw(
      ctx,
      name,
      text.x + Math.trunc((text.width - nw) / 2),
      y,
      k,
      colour,
    );
    drawFlag(ctx, w.unit.owner, text.x + text.width - k - fw, y, fw, fh);
  } else {
    const line = face.fit(w.text, k, text.width - 2 * k),
      lw = face.measure(line, k).width;
    face.draw(
      ctx,
      line,
      text.x + Math.trunc((text.width - lw) / 2),
      y,
      k,
      colour,
      wellNumbers,
    );
  }
  if (w.progress >= 0) {
    ctx.fillStyle = wellBar;
    ctx.fillRect(
      bar.x,
      bar.y,
      Math.round(bar.width * Math.min(w.progress, 1)),
      bar.height,
    );
  }
}
