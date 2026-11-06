import type { Action, Direction, Placement } from "../ui/ring.ts";

// Native arrow processing order: up, left, right, down. Each browser event
// supplies one newly pressed/released key to that state machine.
const arrows: readonly (readonly [string, Direction])[] = [
  ["ArrowUp", "up"],
  ["ArrowLeft", "left"],
  ["ArrowRight", "right"],
  ["ArrowDown", "down"],
];
const opposite: Partial<Record<Direction, Direction>> = {
  up: "down",
  down: "up",
  left: "right",
  right: "left",
};
function canonical(key: string): string {
  if (key === "Space" || key === " ") return " ";
  return key.length === 1 ? key.toLowerCase() : key;
}

/** Browser key transitions for the first side's native gesture key set. */
export class BattleKeys {
  private readonly held = new Map<string, boolean>();
  private fire = false;
  private direction: Direction = "centre";
  private picked = false;

  constructor(
    private readonly callbacks: {
      step(dir: Direction): void;
      click(): void;
      run(action: Action): void;
      placement(): Partial<Placement>;
      shortcut(key: string): boolean;
    },
  ) {}

  down(key: string, repeat = false): boolean {
    const normalized = canonical(key);
    const direction = arrows.find(([name]) => name === normalized)?.[1];
    const core =
      direction !== undefined || normalized === " " || normalized === "Enter";
    const held = this.held.has(normalized);
    if (held || repeat) {
      // Browser repeat supplies native cursor repeat; a held fire pick is
      // changed only by a fresh arrow press, never by its repeat events.
      if (held && repeat && direction !== undefined && !this.fire)
        this.callbacks.step(direction);
      return this.held.get(normalized) ?? core;
    }
    this.held.set(normalized, core);
    if (direction !== undefined) {
      if (this.fire) {
        const next =
          opposite[this.direction] === direction ? "centre" : direction;
        this.direction = next;
        this.picked ||= next !== "centre";
      } else this.callbacks.step(direction);
      return true;
    }
    if (normalized === " ") {
      this.fire = true;
      this.direction = "centre";
      this.picked = false;
      return true;
    }
    if (normalized === "Enter") {
      this.callbacks.click();
      return true;
    }
    const handled = this.callbacks.shortcut(key);
    // A shortcut may cancel this input context. Do not restore keys that
    // its callback just cleared, including during a blur or panel change.
    if (this.held.has(normalized)) this.held.set(normalized, handled);
    return handled;
  }

  up(key: string): boolean {
    const normalized = canonical(key);
    const handled =
      this.held.get(normalized) ??
      (normalized === " " ||
        normalized === "Enter" ||
        arrows.some(([name]) => name === normalized));
    this.held.delete(normalized);
    if (normalized === " " && this.fire) {
      const direction = this.direction,
        picked = this.picked;
      this.resetFire();
      if (direction === "centre" && !picked) this.callbacks.click();
      else
        this.callbacks.run(
          direction === "centre"
            ? "cancel"
            : (this.callbacks.placement()[direction] ?? "cancel"),
        );
    }
    return handled;
  }

  /** Drop a hold without running an action; use on blur/context changes. */
  cancel(): void {
    this.held.clear();
    this.resetFire();
  }

  private resetFire(): void {
    this.fire = false;
    this.direction = "centre";
    this.picked = false;
  }
}
