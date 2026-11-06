import type { Settings } from "../storage/store.ts";
import { routeKey } from "./keys.ts";

interface Callbacks {
  mode(): Settings["playMode"];
  activeSide(): 0 | 1;
  blocked(): boolean;
  handover(): boolean;
  keyDown(side: 0 | 1, key: string, repeat: boolean): boolean;
  keyUp(side: 0 | 1, key: string): boolean;
  cameraKey(side: 0 | 1, key: string, down: boolean): boolean;
  reset(): void;
  leave(): void;
}

interface HeldKey {
  side: 0 | 1;
  key: string;
  camera: boolean;
}

function textControl(target: EventTarget | null): boolean {
  return (
    target instanceof Element &&
    (target.closest("input,textarea,select,[contenteditable=true]") !== null ||
      (target instanceof HTMLElement && target.isContentEditable))
  );
}

function reserved(target: EventTarget | null, key: string): boolean {
  if (/^F\d{1,2}$/i.test(key) || key.toLowerCase() === "escape") return false;
  if (textControl(target)) return true;
  return (
    target instanceof Element &&
    target.closest("button,summary") !== null &&
    [
      " ",
      "space",
      "enter",
      "arrowup",
      "arrowdown",
      "arrowleft",
      "arrowright",
    ].includes(key.toLowerCase())
  );
}

/** One document keyboard owner, with independent native A/B fire holds. */
export class SplitInput {
  private readonly abort = new AbortController();
  private readonly held = new Map<string, HeldKey>();
  private generation = 0;
  private disposed = false;

  constructor(private readonly callbacks: Callbacks) {
    const options = { signal: this.abort.signal };
    window.addEventListener(
      "keydown",
      (event) => {
        this.key(event, true);
      },
      options,
    );
    window.addEventListener(
      "keyup",
      (event) => {
        this.key(event, false);
      },
      options,
    );
    window.addEventListener(
      "blur",
      () => {
        this.reset();
      },
      options,
    );
    window.addEventListener(
      "focusin",
      (event) => {
        const target = event.target;
        if (
          this.held.size > 0 &&
          (textControl(target) ||
            (target instanceof Element &&
              target.closest("button,summary") !== null))
        )
          this.reset();
      },
      options,
    );
  }

  reset(): void {
    this.generation++;
    this.held.clear();
    this.callbacks.reset();
  }

  dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    this.abort.abort();
    this.reset();
  }

  private key(event: KeyboardEvent, down: boolean): void {
    if (this.disposed) return;
    if (this.callbacks.blocked()) {
      if (this.held.size > 0) this.reset();
      return;
    }
    if (this.callbacks.handover()) {
      if (this.held.size > 0) this.reset();
      if (
        down &&
        !event.repeat &&
        event.key.toLowerCase() === "q" &&
        !event.ctrlKey &&
        !event.altKey &&
        !event.metaKey &&
        !textControl(event.target)
      ) {
        this.callbacks.leave();
        event.preventDefault();
      }
      return;
    }
    if (reserved(event.target, event.key)) {
      if (this.held.size > 0) this.reset();
      return;
    }
    const id = event.code || event.key,
      owned = this.held.get(id);
    // A later modifier must never swallow an accepted key's release.
    if (!down) {
      if (owned === undefined) return;
      this.held.delete(id);
      const handled = owned.camera
        ? this.callbacks.cameraKey(owned.side, owned.key, false)
        : this.callbacks.keyUp(owned.side, owned.key);
      if (handled) event.preventDefault();
      return;
    }
    const mode = this.callbacks.mode(),
      bFire =
        mode === "simultaneous" &&
        (this.held.has("ControlLeft") || event.code === "ControlLeft");
    if (event.altKey || event.metaKey || (event.ctrlKey && !bFire)) return;
    const camera =
        mode !== "simultaneous" &&
        ["w", "a", "s", "d"].includes(event.key.toLowerCase()),
      route = camera
        ? { side: this.callbacks.activeSide(), key: event.key.toLowerCase() }
        : routeKey(mode, this.callbacks.activeSide(), event.key, event.code);
    if (route === null) return;
    const generation = this.generation,
      handled = camera
        ? this.callbacks.cameraKey(route.side, route.key, true)
        : this.callbacks.keyDown(route.side, route.key, event.repeat);
    if (handled) {
      // A synchronous callback may pause, hand over, or dispose the host.
      if (generation === this.generation)
        this.held.set(id, { ...route, camera });
      event.preventDefault();
    }
  }
}
