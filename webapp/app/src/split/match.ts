import { play } from "@historyline/engine";
import { MapBattle, type MapBattleCycle } from "../mapBattle.ts";

export type LocalMode = "simultaneous" | "hotseat";

/** One rule cycle, two independent panel controllers. Confirming never supplies
 * orders for the other human. The second confirmation resolves both views. */
export class LocalMatch {
  readonly sides: [MapBattle, MapBattle];
  onChange: (() => void) | null = null;
  handover: 0 | 1 | null;
  activeSide: 0 | 1;
  private paused = false;
  private disposed = false;

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
    this.syncPause();
  }
  /** A save made at the handover resumes behind the next recipient's curtain. */
  saveSide(): 0 | 1 {
    return this.handover ?? this.activeSide;
  }
  takeOver(): void {
    if (this.disposed || this.paused || this.handover === null) return;
    this.activeSide = this.handover;
    this.handover = null;
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
  private confirm(side: number): void {
    if (this.disposed || this.paused || this.handover !== null)
      throw Error("The local match is paused.");
    play.confirm(this.session, side);
    if (play.ready(this.session, 0) && play.ready(this.session, 1)) {
      const before = this.sides.map((b) => play.view(this.session, b.side)),
        events = play.resolve(this.session);
      for (const b of this.sides) {
        const view = before[b.side];
        if (view === undefined) throw Error("Missing panel view.");
        b.playback.resolution(view, play.view(this.session, b.side), events);
        b.status = "Orders resolved.";
      }
    } else
      this.sides[side === 0 ? 0 : 1].status =
        "Waiting for the other human to confirm.";
    if (this.mode === "hotseat") {
      this.handover = side === 0 ? 1 : 0;
      this.syncPause();
    }
  }
}
