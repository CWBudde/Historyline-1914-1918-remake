import { expect, it } from "vitest";
import { layout } from "./geometry.ts";
import { wellOf, wellScale, type WellInput } from "./well.ts";
import { PixelFace, parseMetrics } from "../pixfont.ts";

const base: WellInput = {
  stage: null,
  ended: null,
  planning: null,
  attack: false,
  busy: false,
  ready: false,
  enemyAttacks: false,
  human: true,
  message: "",
  unit: null,
  orders: 0,
  limit: 0,
};

it("words the wells as the desktop does", () => {
  expect(wellOf(base)).toEqual({
    text: "MOVEMENT PHASE  0 ORDERS",
    unit: null,
    progress: -1,
  });
  expect(wellOf({ ...base, orders: 1 }).text).toBe("MOVEMENT PHASE  1 ORDER");
  expect(wellOf({ ...base, orders: 3, limit: 6 }).text).toBe(
    "MOVEMENT PHASE  3/6 ORDERS",
  );
  const unit = { size: 6, exp: 2, name: "Supply car", owner: 1 };
  expect(wellOf({ ...base, unit }).unit).toEqual({
    ...unit,
    name: "SUPPLY CAR",
  });
  // A fresh message beats the unit line; state beats both.
  expect(wellOf({ ...base, unit, message: "Unit moved." })).toEqual({
    text: "UNIT MOVED.",
    unit: null,
    progress: -1,
  });
  expect(wellOf({ ...base, unit, ready: true }).text).toBe(
    "WAITING: ENEMY MOVES",
  );
  expect(wellOf({ ...base, ready: true, enemyAttacks: true }).text).toBe(
    "WAITING: ENEMY ATTACKS",
  );
  expect(wellOf({ ...base, busy: true, ready: true }).text).toBe(
    "ORDERS UNDER WAY",
  );
  expect(wellOf({ ...base, human: false, attack: true }).text).toBe(
    "ATTACK PHASE",
  );
  expect(wellOf({ ...base, human: false, planning: 0.25 })).toEqual({
    text: "PLANNING THE MOVES",
    unit: null,
    progress: 0.25,
  });
  expect(
    wellOf({ ...base, human: false, planning: 0, attack: true }).text,
  ).toBe("WEIGHING THE ATTACKS");
  expect(wellOf({ ...base, stage: "F1: end the phase" }).text).toBe(
    "F1: END THE PHASE",
  );
  // The end of the battle beats the prompt.
  expect(
    wellOf({ ...base, stage: "F1: END THE PHASE", ended: "DEFEAT - X" }).text,
  ).toBe("DEFEAT - X");
});

it("places the status line and bar inside the well's bevel, at twice the UI scale", () => {
  const f = new PixelFace(
    parseMetrics({
      height: 5,
      glyphs: {},
      fallback: { x: 0, width: 4, height: 5, advance: 5 },
    }),
  );
  for (const [w, h, ui, k] of [
    [1920, 1080, 1, 2],
    [2560, 1440, 1, 2],
    [3840, 2160, 2, 4],
    [1280, 720, 1.5, 3],
  ] as const) {
    const s = layout(w, h, ui);
    for (const side of [0, 1] as const) {
      const st = s.status[side],
        bw = Math.max(Math.round(2 * ui), 1),
        text = s.wellText(side),
        bar = s.wellBar(side);
      expect(text.x).toBeGreaterThan(st.x + bw);
      expect(text.y).toBeGreaterThan(st.y + bw);
      expect(text.x - st.x).toBe(st.x + st.width - (text.x + text.width));
      expect(text.y + text.height).toBeLessThanOrEqual(bar.y);
      expect(bar.y + bar.height).toBe(st.y + st.height - bw);
      expect(bar.height).toBe(Math.max(Math.round(4 * ui), 1));
      expect(text.height).toBeGreaterThanOrEqual(Math.round(10 * ui));
      expect(wellScale(text, ui, f)).toBe(k);
    }
  }
});
