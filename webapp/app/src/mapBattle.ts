import { hexgrid, level, play, phaseMovement } from "@historyline/engine";
import { MapPlayback } from "./mapPlayback.ts";
import { toward } from "./unitFrame.ts";

export interface MapComputer {
  commands(session: play.Session, side: number): Promise<play.Command[]>;
}

/** Map orders use only the human side's view; the rules own the actual battle. */
export class MapBattle {
  readonly playback: MapPlayback;
  selected: number | null = null;
  hover: hexgrid.Coord | null = null;
  status = "";
  version = 0;
  private planning = false;

  constructor(
    readonly session: play.Session,
    readonly side: number,
    private readonly computer?: MapComputer,
  ) {
    if (side !== 0 && side !== 1) throw new Error("Invalid battle side.");
    this.playback = new MapPlayback(side);
  }

  busy(): boolean {
    return this.planning || this.playback.busy();
  }
  view(): play.View {
    return this.playback.shown(play.view(this.session, this.side));
  }
  cancel(): void {
    this.selected = null;
    this.status = "";
    this.version++;
  }
  destinations(): Set<string> {
    if (!this.canAct() || this.selected === null || !this.movement())
      return new Set();
    try {
      return play.destinations(this.session, this.side, this.selected);
    } catch (e) {
      this.refused(e);
      return new Set();
    }
  }
  targets(): Set<string> {
    if (!this.canAct() || this.selected === null || this.movement())
      return new Set();
    return play.targets(this.session, this.side, this.selected);
  }
  route(to: hexgrid.Coord): hexgrid.Coord[] {
    if (!this.destinations().has(level.coordKey(to)) || this.selected === null)
      return [];
    try {
      return play.route(this.session, this.side, this.selected, to);
    } catch (e) {
      this.refused(e);
      return [];
    }
  }
  orders(): play.Event[] {
    return play.orders(this.session, this.side);
  }
  click(at: hexgrid.Coord): void {
    if (!this.canAct()) return;
    const key = level.coordKey(at);
    if (this.selected !== null) {
      const moving = this.movement();
      if ((moving ? this.destinations() : this.targets()).has(key)) {
        const idx = this.selected;
        const unit = [
          ...play.view(this.session, this.side).units.values(),
        ].find((u) => u.index === idx);
        this.selected = null;
        try {
          const events = moving
            ? play.move(this.session, this.side, idx, at)
            : play.attack(this.session, this.side, idx, at);
          for (const e of events) {
            if (unit !== undefined) this.playback.move(e, unit);
            if (e.kind === play.Ordered || e.kind === play.Tasked) {
              const [x, y] = hexgrid.center(e.from);
              const [xx, yy] = hexgrid.center(e.to);
              this.playback.directions.set(
                e.unit,
                toward({ x, y }, { x: xx, y: yy }),
              );
            }
            this.status = e.text || (moving ? "Unit moved." : "Order queued.");
          }
        } catch (e) {
          this.refused(e);
        }
        this.version++;
        return;
      }
    }
    this.selected = null;
    this.status = "";
    const shown = play.view(this.session, this.side);
    let unit = shown.units.get(key);
    if (unit?.head !== null && unit?.head !== undefined)
      unit = shown.units.get(level.coordKey(unit.head));
    if (unit !== undefined && unit.side === this.side) {
      try {
        play.canOrder(this.session, this.side, unit.index);
        this.selected = unit.index;
        this.status = `${unit.name}: ${this.movement() ? "pick where to move" : "pick a target"}`;
      } catch (e) {
        this.refused(e);
      }
    }
    this.version++;
  }
  async phase(): Promise<void> {
    if (!this.canAct()) return;
    this.cancel();
    this.planning = true;
    this.version++;
    try {
      play.confirm(this.session, this.side);
      const opponent = 1 - this.side;
      if (!play.ready(this.session, opponent)) {
        if (this.computer === undefined) play.confirm(this.session, opponent);
        else {
          const commands = await this.computer.commands(this.session, opponent);
          play.replay(this.session, commands);
        }
      }
      if (play.ready(this.session, 0) && play.ready(this.session, 1)) {
        const before = play.view(this.session, this.side);
        const events = play.resolve(this.session);
        this.playback.resolution(
          before,
          play.view(this.session, this.side),
          events,
        );
        this.status = "Orders resolved.";
      }
    } catch (e) {
      if (
        play.ended(this.session) === null &&
        play.ready(this.session, this.side)
      )
        play.unconfirm(this.session, this.side);
      this.refused(e);
    } finally {
      this.planning = false;
      this.version++;
    }
  }
  tick(ms: number): void {
    const active = this.playback.busy() || this.playback.effects().length > 0;
    this.playback.advance(ms);
    if (active) this.version++;
  }
  private movement(): boolean {
    return play.phase(this.session, this.side) === phaseMovement;
  }
  private canAct(): boolean {
    return (
      !this.busy() &&
      play.ended(this.session) === null &&
      !play.ready(this.session, this.side)
    );
  }
  private refused(error: unknown): void {
    this.status = error instanceof Error ? error.message : String(error);
    this.version++;
  }
}
