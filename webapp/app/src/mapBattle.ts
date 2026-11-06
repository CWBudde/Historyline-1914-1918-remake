import {
  hexgrid,
  level,
  play,
  phaseMovement,
  bind,
  exitMovement,
  route as ruleRoute,
} from "@historyline/engine";
import { MapPlayback } from "./mapPlayback.ts";
import { toward } from "./unitFrame.ts";
import { ContentsController, type ExitAim } from "./contents/controller.ts";
import {
  placement,
  ringActions,
  type Action,
  type Direction,
} from "./ui/ring.ts";

export interface MapComputer {
  commands(
    session: play.Session,
    side: number,
    progress?: (done: number, total: number) => void,
  ): Promise<play.Command[]>;
}

/** Orders and panels share one fair side view; only play mutates the rules. */
export class MapBattle {
  readonly playback: MapPlayback;
  readonly contents: ContentsController;
  selected: number | null = null;
  exiting: ExitAim | null = null;
  hover: hexgrid.Coord | null = null;
  cursor: hexgrid.Coord;
  panel: "orders" | "unit" | "game" | "map" = "orders";
  infoUnit: play.UnitView | null = null;
  phasePrompt = false;
  surrenderPrompt = false;
  progress: { done: number; total: number } | null = null;
  status = "";
  version = 0;
  onChange: (() => void) | null = null;
  private planning = false;

  constructor(
    readonly session: play.Session,
    readonly side: number,
    private readonly computer?: MapComputer,
  ) {
    if (side !== 0 && side !== 1) throw new Error("Invalid battle side.");
    this.playback = new MapPlayback(side);
    this.contents = new ContentsController(
      session,
      side,
      () => this.busy() || this.phasePrompt || this.surrenderPrompt,
    );
    this.cursor = {
      ...([...this.view().units.values()].find((u) => u.side === side)?.at ?? {
        line: 1,
        col: 0,
      }),
    };
  }
  busy(): boolean {
    return this.planning || this.playback.busy();
  }
  view(): play.View {
    return this.playback.shown(play.view(this.session, this.side));
  }
  cancel(): void {
    if (this.playback.skipWalk() || this.playback.skipScene()) {
      this.touch();
      return;
    }
    if (this.phasePrompt || this.surrenderPrompt) {
      this.phasePrompt = false;
      this.surrenderPrompt = false;
      this.status = "Back to the orders.";
      this.touch();
      return;
    }
    if (this.contents.box !== null) {
      this.contents.close();
      this.status = this.contents.message;
      this.touch();
      return;
    }
    this.clearSelection();
    this.status = "";
    this.touch();
  }
  setCursor(at: hexgrid.Coord | null): void {
    if (at === null) {
      if (this.hover !== null) {
        this.hover = null;
        this.touch();
      }
      return;
    }
    if (
      level.coordKey(at) === level.coordKey(this.cursor) &&
      this.hover !== null
    )
      return;
    this.cursor = { ...at };
    this.hover = { ...at };
    this.touch();
  }
  stepCursor(direction: Direction): void {
    if (this.playback.busy() || this.phasePrompt || this.surrenderPrompt)
      return;
    if (this.contents.box !== null) {
      this.contents.step(direction);
      this.touch();
      return;
    }
    const ns = hexgrid.around(this.cursor),
      odd = (this.cursor.line & 1) === 1;
    const at =
      direction === "up"
        ? ns[5]
        : direction === "down"
          ? ns[2]
          : direction === "left"
            ? ns[odd ? 3 : 4]
            : direction === "right"
              ? ns[odd ? 1 : 0]
              : undefined;
    if (
      at !== undefined &&
      hexgrid.contains(level.grid(this.session.level), at)
    ) {
      const cell = bind.cell(this.session.board, at);
      if (cell !== null && !bind.edge(this.session.board, cell))
        this.setCursor(at);
    }
  }
  destinations(): Set<string> {
    if (!this.canAct() || this.selected === null || !this.movement())
      return new Set();
    if (this.exiting !== null) return new Set(this.exiting.destinations);
    try {
      return play.destinations(this.session, this.side, this.selected);
    } catch (e) {
      this.refused(e);
      return new Set();
    }
  }
  targets(): Set<string> {
    return !this.canAct() || this.selected === null || this.movement()
      ? new Set()
      : play.targets(this.session, this.side, this.selected);
  }
  route(to: hexgrid.Coord): hexgrid.Coord[] {
    if (!this.destinations().has(level.coordKey(to)) || this.selected === null)
      return [];
    try {
      if (this.exiting === null)
        return play.route(this.session, this.side, this.selected, to);
      const m = exitMovement(
        this.session.game,
        this.side,
        this.exiting.hold,
        this.exiting.slot,
      );
      play.patchLight(this.session, m, this.selected);
      const target = bind.cell(this.session.board, to);
      if (target === null) return [];
      return (ruleRoute(this.session.game.profile, m, target) ?? []).flatMap(
        (cell) => {
          const c = bind.coord(this.session.board, cell);
          return c === null ? [] : [c];
        },
      );
    } catch (e) {
      this.refused(e);
      return [];
    }
  }
  orders(): play.Event[] {
    return play.orders(this.session, this.side);
  }
  fire(): void {
    const box = this.contents.box;
    if (box === null) this.click(this.cursor);
    else if (box.producing) {
      const option = this.contents.options()[box.pick];
      if (option !== undefined) this.produce(option.type);
    } else if (this.contents.unit() !== null) this.run("exit");
  }
  click(at: hexgrid.Coord): void {
    if (this.playback.skipWalk() || this.playback.skipScene()) {
      this.touch();
      return;
    }
    if (this.phasePrompt || this.surrenderPrompt) {
      this.cancel();
      return;
    }
    if (!this.canAct()) {
      this.status = play.ready(this.session, this.side)
        ? "Orders confirmed: F1 opens them again."
        : this.busy()
          ? "The orders are being carried out."
          : this.status;
      this.touch();
      return;
    }
    this.cursor = { ...at };
    this.hover = { ...at };
    if (this.selected !== null) {
      const moving = this.movement(),
        legal = moving ? this.destinations() : this.targets(),
        entrance = this.containerAt(at);
      const target = legal.has(level.coordKey(entrance)) ? entrance : at;
      if (legal.has(level.coordKey(target))) {
        const index = this.selected,
          exit = this.exiting,
          unit =
            exit?.unit ??
            [...this.view().units.values()].find((u) => u.index === index);
        this.clearSelection();
        try {
          const events =
            exit !== null
              ? play.exit(this.session, this.side, exit.hold, exit.slot, target)
              : moving
                ? play.move(this.session, this.side, index, target)
                : play.attack(this.session, this.side, index, target);
          this.receive(events, unit, moving ? "Unit moved." : "Order queued.");
        } catch (e) {
          this.refused(e);
        }
        this.touch();
        return;
      }
    }
    this.clearSelection();
    this.status = "";
    const unit = this.unitAt(at);
    if (unit?.side === this.side) {
      try {
        play.canOrder(this.session, this.side, unit.index);
        this.selected = unit.index;
        this.status = `${unit.name}: ${this.movement() ? "pick where to move" : play.sapper(this.session, unit.index) ? "pick a hex to dig or fill" : play.builder(this.session, unit.index) ? "pick where to build a depot" : "pick a target"}`;
      } catch (e) {
        this.refused(e);
      }
    } else if (
      this.view().buildings.get(level.coordKey(this.containerAt(at)))?.owner ===
      play.faction(this.side)
    )
      this.openContents(at);
    this.touch();
  }
  openContents(at: hexgrid.Coord = this.cursor): void {
    this.clearSelection();
    this.contents.open(at);
    this.status = this.contents.message;
    this.touch();
  }
  selectSlot(slot: number): void {
    try {
      this.contents.select(slot);
    } catch (e) {
      this.refused(e);
    }
    this.touch();
  }
  produce(type: number): void {
    try {
      this.receive(this.contents.build(type), undefined, this.contents.message);
      this.status = this.contents.message;
    } catch (e) {
      this.refused(e);
    }
    this.touch();
  }
  showInfo(at: hexgrid.Coord = this.cursor): void {
    const unit = this.unitAt(at);
    if (unit !== undefined && unit.side !== this.side) {
      this.status = "Information is only available for your own units.";
      this.touch();
      return;
    }
    this.clearSelection();
    if (unit !== undefined) {
      this.panel = "unit";
      this.infoUnit = structuredClone(unit);
    } else this.panel = "game";
    this.touch();
  }
  actions(at: hexgrid.Coord = this.cursor): Action[] {
    if (
      this.selected !== null ||
      this.phasePrompt ||
      this.surrenderPrompt ||
      this.contents.box?.producing === true ||
      this.playback.scene() !== null
    )
      return [];
    return ringActions(
      placement(
        !this.movement(),
        play.contents(this.session, this.side, this.containerAt(at)) !== null,
        this.contents.box !== null,
      ),
      (a) => this.allowed(a, at),
    );
  }
  run(action: Action, at: hexgrid.Coord = this.cursor): void {
    if (action === "cancel") {
      this.cancel();
      return;
    }
    if (!this.allowed(action, at)) {
      this.status = "That action is unavailable now.";
      this.touch();
      return;
    }
    try {
      switch (action) {
        case "move":
        case "attack":
          this.clearSelection();
          this.click(at);
          return;
        case "mode":
          void this.requestPhase();
          return;
        case "open":
          this.openContents(at);
          return;
        case "info":
          this.showInfo(at);
          return;
        case "map":
          this.clearSelection();
          this.panel = "map";
          break;
        case "produce":
          this.contents.showProduction();
          this.status = this.contents.message;
          break;
        case "repair":
          this.receive(
            this.contents.repair(),
            this.contents.unit() ?? undefined,
            this.contents.message,
          );
          this.status = this.contents.message;
          break;
        case "exit": {
          const exit = this.contents.prepareExit();
          this.clearSelection();
          this.exiting = exit;
          this.selected = exit.unit.index;
          this.cursor = { ...exit.from };
          this.hover = { ...exit.from };
          this.status = `${exit.unit.name} leaves: pick where to go`;
          break;
        }
      }
    } catch (e) {
      this.refused(e);
    }
    this.touch();
  }
  async requestPhase(): Promise<void> {
    if (play.ended(this.session) !== null) return;
    if (play.ready(this.session, this.side) && !this.playback.busy()) {
      play.unconfirm(this.session, this.side);
      this.phasePrompt = false;
      this.status = "Orders open again.";
      this.touch();
      return;
    }
    if (this.playback.busy()) {
      this.status = "The orders are being carried out.";
      this.touch();
      return;
    }
    this.clearSelection();
    this.surrenderPrompt = false;
    if (!this.phasePrompt) {
      this.phasePrompt = true;
      this.status =
        "Confirm the mode change with F1, or cancel to return to orders.";
      this.touch();
      return;
    }
    this.phasePrompt = false;
    await this.phase();
  }
  requestSurrender(): void {
    if (play.ended(this.session) !== null || this.playback.busy()) return;
    this.clearSelection();
    this.phasePrompt = false;
    if (!this.surrenderPrompt) {
      this.surrenderPrompt = true;
      this.status = "Give up this battle? Press F2 again to confirm.";
      this.touch();
      return;
    }
    this.surrenderPrompt = false;
    try {
      const events = play.surrender(this.session, this.side);
      this.receive(events, undefined, "The battle is over.");
    } catch (e) {
      this.refused(e);
    }
    this.touch();
  }
  async phase(): Promise<void> {
    if (this.planning) {
      if (
        !play.ready(this.session, this.side) &&
        play.ended(this.session) === null
      ) {
        play.confirm(this.session, this.side);
        this.touch();
      }
      return;
    }
    if (!this.canAct()) return;
    this.clearSelection();
    this.planning = true;
    this.progress = { done: 0, total: 1 };
    this.touch();
    try {
      play.confirm(this.session, this.side);
      const opponent = 1 - this.side;
      if (!play.ready(this.session, opponent)) {
        if (this.computer === undefined) play.confirm(this.session, opponent);
        else {
          const commands = await this.computer.commands(
            this.session,
            opponent,
            (done, total) => {
              this.progress = { done, total };
              this.touch();
            },
          );
          if (play.ended(this.session) !== null) return;
          play.replay(this.session, commands);
        }
      }
      if (play.ready(this.session, 0) && play.ready(this.session, 1)) {
        const before = play.view(this.session, this.side),
          events = play.resolve(this.session);
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
      this.progress = null;
      this.touch();
    }
  }
  tick(ms: number): void {
    const active = this.playback.busy() || this.playback.effects().length > 0,
      scene = this.playback.scene()?.battle,
      wasBusy = this.busy();
    this.playback.advance(ms);
    if (active) this.version++;
    if (scene !== this.playback.scene()?.battle || wasBusy !== this.busy())
      this.onChange?.();
  }
  private allowed(action: Action, at: hexgrid.Coord): boolean {
    if (action === "cancel" || action === "info" || action === "map")
      return true;
    if (action === "repair" || action === "exit" || action === "produce") {
      const a = this.contents.actions();
      return (
        (action === "repair"
          ? a.repair
          : action === "exit"
            ? a.leave
            : a.produce) === null
      );
    }
    if (action === "mode")
      return !this.playback.busy() && play.ended(this.session) === null;
    if (this.busy() || play.ended(this.session) !== null) return false;
    if (action === "open") {
      const c = play.contents(this.session, this.side, this.containerAt(at));
      return (
        c !== null &&
        (c.slots.length > 0 ||
          play.buildable(this.session, this.side, c.hold).length > 0)
      );
    }
    if (!this.canAct()) return false;
    const unit = this.unitAt(at);
    if (unit?.side !== this.side) return false;
    try {
      play.canOrder(this.session, this.side, unit.index);
      return action === "attack"
        ? !this.movement() &&
            play.targets(this.session, this.side, unit.index).size > 0
        : this.movement() &&
            play.destinations(this.session, this.side, unit.index).size > 0;
    } catch {
      return false;
    }
  }
  private unitAt(at: hexgrid.Coord): play.UnitView | undefined {
    const view = this.view();
    let unit = view.units.get(level.coordKey(at));
    if (unit?.head !== null && unit?.head !== undefined)
      unit = view.units.get(level.coordKey(unit.head));
    return unit;
  }
  private containerAt(at: hexgrid.Coord): hexgrid.Coord {
    const unit = this.unitAt(at);
    return unit?.side === this.side && unit.carrier
      ? unit.at
      : (play.entrance(this.session, at) ?? at);
  }
  private clearSelection(): void {
    this.selected = null;
    this.exiting = null;
    this.contents.close(true);
    this.panel = "orders";
    this.infoUnit = null;
  }
  private receive(
    events: readonly play.Event[],
    unit: play.UnitView | undefined,
    fallback: string,
  ): void {
    for (const e of events) {
      if (unit !== undefined) this.playback.move(e, unit);
      if (e.kind === play.Ordered || e.kind === play.Tasked) {
        const [x, y] = hexgrid.center(e.from),
          [xx, yy] = hexgrid.center(e.to);
        this.playback.directions.set(
          e.unit,
          toward({ x, y }, { x: xx, y: yy }),
        );
      }
      this.status = e.text || fallback;
    }
  }
  private movement(): boolean {
    return play.phase(this.session, this.side) === phaseMovement;
  }
  private canAct(): boolean {
    return (
      !this.busy() &&
      !this.phasePrompt &&
      !this.surrenderPrompt &&
      play.ended(this.session) === null &&
      !play.ready(this.session, this.side)
    );
  }
  private touch(): void {
    this.version++;
    this.onChange?.();
  }
  private refused(error: unknown): void {
    this.status = error instanceof Error ? error.message : String(error);
    this.touch();
  }
}
