import type { Settings } from "../storage/store.ts";

const keySetB: Readonly<Record<string, string>> = {
  d: "ArrowUp",
  c: "ArrowDown",
  x: "ArrowLeft",
  v: "ArrowRight",
};
const first = new Set([
  "arrowup",
  "arrowdown",
  "arrowleft",
  "arrowright",
  " ",
  "space",
  "enter",
  "escape",
  "f1",
  "f2",
  "f8",
  "f9",
  "f11",
  "q",
  // The order menu's hex keys (w, a, s, d come as camera keys), the keypad's, the strip's Info and Map.
  "e",
  "1",
  "2",
  "3",
  "7",
  "8",
  "9",
  "i",
  "o",
  "contextmenu",
  "m",
  "home",
  "end",
  "pageup",
  "pagedown",
]);

/** The second native player uses a cross around C and left Control for fire.
 * Global shortcuts stay with key set A; hot-seat gives A to its recipient. */
export function routeKey(
  mode: Settings["playMode"],
  active: 0 | 1,
  key: string,
  code: string,
): { side: 0 | 1; key: string } | null {
  const lower = key.toLowerCase();
  if (mode === "simultaneous") {
    const mapped = keySetB[lower];
    if (mapped !== undefined) return { side: 1, key: mapped };
    if (code === "ControlLeft") return { side: 1, key: " " };
  }
  return first.has(lower)
    ? { side: mode === "simultaneous" ? 0 : active, key }
    : null;
}
