import { play } from "@historyline/engine";
import { MapBattle, type MapBattleCycle } from "../mapBattle.ts";
import { goEvents, type PendingPlayback } from "./pending.ts";

export type LocalMode = "simultaneous" | "hotseat";

/** One rule cycle, two independent panel controllers. Confirming never supplies
 * orders for the other human. The second confirmation resolves both views; in
 * hot-seat the resolution waits behind the curtain for the next human, who sees
 * it at once, and the other one keeps it for its next turn (as the desktop game). */
export class LocalMatch {
  readonly sides: [MapBattle, MapBattle];
  onChange: (() => void) | null = null;
  handover: 0 | 1 | null;
  activeSide: 0 | 1;
  private paused = false;
  private disposed = false;
  /** The hot-seat resolution a side has not seen yet, kept for a save: its game before and its events. */
  private kept: {
    side: 0 | 1;
    before: Uint8Array;
    events: play.Event[];
  } | null = null;

  constructor(
    readonly session: play.Session,
    readonly mode: LocalMode,
    activeSide = 0,
  ) {
    if (activeSide !== 0 && activeSide !== 1)
      throw Error("Invalid active side.");
    if (session.game.players.some((p) => p.computer))
      throw Error("A local match needs two human players.");
    this.activeSide = activeSide;
    this.handover = mode === "hotseat" ? activeSide : null;
    const cycle: MapBattleCycle = {
      confirm: (side) => {
        this.confirm(side);
      },
      changed: (side) => {
        this.changed(side);
      },
    };
    this.sides = [
      new MapBattle(session, 0, undefined, cycle),
      new MapBattle(session, 1, undefined, cycle),
    ];
    if (mode === "simultaneous") this.sides[1].surrenderKey = "F3";
    this.syncPause();
  }
  /** A save made at the handover resumes behind the next recipient's curtain. */
  saveSide(): 0 | 1 {
    return this.handover ?? this.activeSide;
  }
  /**
   * savedPending is the hot-seat resolution a side has not seen yet as a save keeps it (storage/store.ts
   * SavePending, the events in Go's JSON), null for none.
   */
  savedPending(): {
    side: 0 | 1;
    before: Uint8Array;
    events: Record<string, unknown>[];
  } | null {
    const k = this.kept;
    return k === null
      ? null
      : { side: k.side, before: k.before, events: goEvents(k.events) };
  }
  /** restore queues a saved battle's pending playback on its side's panel, shown at that side's takeover. */
  restore(p: PendingPlayback): void {
    if (this.disposed || this.mode !== "hotseat") return;
    const b = this.sides[p.side];
    b.playback.resolution(p.view, play.view(this.session, p.side), p.events);
    b.status = "Orders resolved.";
    this.kept = { side: p.side, before: p.before, events: p.events };
    this.syncPause();
  }
  takeOver(): void {
    if (this.disposed || this.paused || this.handover === null) return;
    this.activeSide = this.handover;
    this.handover = null;
    if (this.kept?.side === this.activeSide) this.kept = null; // its playback runs now
    if (play.ready(this.session, 0) && play.ready(this.session, 1))
      this.resolve();
    this.syncPause();
    this.onChange?.();
  }
  setPaused(paused: boolean): void {
    if (this.disposed || this.paused === paused) return;
    this.paused = paused;
    this.syncPause();
  }
  dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    this.onChange = null;
    for (const b of this.sides) b.dispose();
  }
  private syncPause(): void {
    for (const b of this.sides)
      b.setPaused(
        this.paused ||
          (this.mode === "hotseat" &&
            (this.handover !== null || b.side !== this.activeSide)),
      );
  }
  private changed(side: number): void {
    if (this.disposed) return;
    this.sides[side === 0 ? 1 : 0].refresh();
    this.onChange?.();
  }
  /** resolve plays the resolution on both panels; in hot-seat the side not at the controls keeps it. */
  private resolve(): void {
    const before = this.sides.map((b) => play.view(this.session, b.side)),
      saved = this.mode === "hotseat" ? play.save(this.session) : null,
      events = play.resolve(this.session);
    for (const b of this.sides) {
      const view = before[b.side];
      if (view === undefined) throw Error("Missing panel view.");
      b.playback.resolution(view, play.view(this.session, b.side), events);
      b.status = "Orders resolved.";
    }
    if (saved !== null)
      this.kept = {
        side: this.activeSide === 0 ? 1 : 0,
        before: saved,
        events: [...events],
      };
  }
  private confirm(side: number): void {
    if (this.disposed || this.paused || this.handover !== null)
      throw Error("The local match is paused.");
    play.confirm(this.session, side);
    const both = play.ready(this.session, 0) && play.ready(this.session, 1);
    if (both && this.mode === "simultaneous") this.resolve();
    else
      this.sides[side === 0 ? 0 : 1].status = both
        ? "Orders confirmed."
        : "Waiting for the other human to confirm.";
    if (this.mode === "hotseat") {
      this.handover = side === 0 ? 1 : 0;
      this.syncPause();
    }
  }
}
