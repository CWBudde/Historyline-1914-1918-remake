import {
  hexgrid,
  level,
  play,
  phaseMovement,
  bind,
  exitMovement,
  route as ruleRoute,
  tunnelTurns,
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
import { menuItems, slotTable, type MenuItem } from "./ui/orderMenu.ts";

/**
 * MapComputer plans side's orders off the frame (computer.ts): it snapshots the session when called, as postMessage
 * does, and resolves with the commands to replay on the live session.
 */
export interface MapComputer {
  commands(
    session: play.Session,
    side: number,
    progress?: (done: number, total: number) => void,
  ): Promise<play.Command[]>;
}

// The computer side plans as soon as its phase opens, in the worker on a snapshot of the session, while the player
// keeps giving orders on the live one; it then carries out its commands on the live session one by one, each move
// resting a moment and walking on the computer's own panel, and confirms after the last (cmd/historyline/plan.go).
// A command the live session refuses (the player changed something in between) makes it plan once more on the
// session as it is now; when that fails too it confirms the orders it has given, so the battle always goes on.

/** planMinMs is how long the planning shows at least, so its bar is seen to run (plan.go planMinFrames). */
export const planMinMs = (45 * 1000) / 60;
/** actRestMs is how long the computer's panel rests before each move (plan.go actPauseFrames). */
export const actRestMs = (12 * 1000) / 60;
/** actFastSpeed is how many times as fast the computer carries out its orders once the player waits for it. */
export const actFastSpeed = 3;

/** ComputerStage is where the computer's moves are seen: its own panel on the split screen. */
export interface ComputerStage {
  /** act shows one command's events; it reports whether a move now rests restMs and walks there. */
  act(events: readonly play.Event[], restMs: number): boolean;
  /** busy reports whether the panel still plays a move or a resolution. */
  busy(): boolean;
  /** pace sets how many times as fast its moves run. */
  pace(speed: number): void;
}

/** computerStage is battle's panel as the stage of its side's computer moves; moves only turn while not shown. */
export function computerStage(
  battle: MapBattle,
  shown: () => boolean,
): ComputerStage {
  return {
    act: (events, restMs) => battle.act(events, shown() ? restMs : null),
    busy: () => battle.playback.busy(),
    pace: (speed) => {
      battle.playback.pace = speed;
    },
  };
}

/** Turn is one planning of the computer side and the carrying-out of its commands. */
interface Turn {
  /** retry reports whether this planning follows a failed one. */
  retry: boolean;
  /** elapsed is how long the planning has shown, in ms. */
  elapsed: number;
  /** answer is the worker's, once it came. */
  answer: { commands: play.Command[] } | { error: unknown } | null;
  /** commands are those still to carry out, null while planning. */
  commands: play.Command[] | null;
}

/** Local humans share rules, while each keeps its own fair view and controls. */
export interface MapBattleCycle {
  confirm(side: number): void;
  changed(side: number): void;
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
  /** surrenderKey gives up the battle: F2, the right human's F3 in simultaneous play (split/match.ts). */
  surrenderKey: "F2" | "F3" = "F2";
  progress: { done: number; total: number } | null = null;
  status = "";
  version = 0;
  onChange: (() => void) | null = null;
  onResolution:
    | ((before: [play.View, play.View], events: readonly play.Event[]) => void)
    | null = null;
  /** stage is the computer's own panel, where its moves walk; null shows them nowhere. */
  stage: ComputerStage | null = null;
  private turn: Turn | null = null;
  private paused = false;
  private disposed = false;

  constructor(
    readonly session: play.Session,
    readonly side: number,
    private readonly computer?: MapComputer,
    private readonly cycle?: MapBattleCycle,
  ) {
    if (side !== 0 && side !== 1) throw new Error("Invalid battle side.");
    this.playback = new MapPlayback(side);
    this.contents = new ContentsController(
      session,
      side,
      () =>
        this.busy() ||
        this.paused ||
        this.disposed ||
        this.phasePrompt ||
        this.surrenderPrompt,
    );
    this.cursor = startHex(this.view(), side, session.level);
  }
  busy(): boolean {
    return this.playback.busy();
  }
  /** computerState is what the computer side does now: planning, carrying out its commands, or nothing. */
  computerState(): "planning" | "acting" | null {
    return this.turn === null
      ? null
      : this.turn.commands === null
        ? "planning"
        : "acting";
  }
  setPaused(paused: boolean): void {
    if (this.disposed || paused === this.paused) return;
    this.paused = paused;
    this.touch();
  }
  dispose(): void {
    this.disposed = true;
    // A late answer of the worker finds no turn and lands nothing.
    this.turn = null;
    this.progress = null;
    this.stage?.pace(1);
    this.stage = null;
    this.paused = false;
    this.onChange = null;
    this.onResolution = null;
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
      : actionTargets(this.session, this.side, this.selected);
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
        this.status = this.movement()
          ? `${unit.name}: pick where to move`
          : actionPrompt(this.session, this.side, unit, this.targets().size);
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
  /** showGame opens the game information, which also leads to the settings on the split screen. */
  showGame(): void {
    this.clearSelection();
    this.panel = "game";
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
  /**
   * opensOnly reports whether at holds an own unit that cannot be selected (it has its order) but still has
   * contents to open: a carrier that moved lets its load out.
   */
  opensOnly(at: hexgrid.Coord): boolean {
    const unit = this.unitAt(at);
    if (
      this.selected !== null ||
      this.contents.box !== null ||
      !this.canAct() ||
      unit?.side !== this.side ||
      !this.allowed("open", at)
    )
      return false;
    try {
      play.canOrder(this.session, this.side, unit.index);
      return false;
    } catch {
      return true;
    }
  }
  /** unitCell is the cell of the unit on at (its head where at is a two-cell unit's rear), else at. */
  unitCell(at: hexgrid.Coord): hexgrid.Coord {
    return this.unitAt(at)?.at ?? at;
  }
  /** holdsSelected reports
 whether at holds the selected unit (either cell of a unit with a rear). */
  holdsSelected(at: hexgrid.Coord): boolean {
    return (
      this.selected !== null &&
      this.exiting === null &&
      this.unitAt(at)?.index === this.selected
    );
  }
  /**
   * menu is the order menu of the cell at
 (ui/orderMenu.ts): the unit's own orders on their slots, each with why it
   * is not possible now; empty where the cell has no order of the side's (no own unit or container) or while a
   * prompt, a production list or a battle scene is open. End phase, map and info are the order strip's.
   */
  menu(at: hexgrid.Coord = this.cursor): MenuItem[] {
    const box = this.contents.box;
    if (
      this.phasePrompt ||
      this.surrenderPrompt ||
      box?.producing === true ||
      this.playback.scene() !== null ||
      play.ended(this.session) !== null
    )
      return [];
    const unit = this.unitAt(at),
      own = unit?.side === this.side ? unit : undefined,
      container =
        box === null
          ? play.contents(this.session, this.side, this.containerAt(at))
          : null,
      moving = this.movement(),
      held = (): string | null =>
        this.busy()
          ? "The orders are being carried out."
          : play.ready(this.session, this.side)
            ? "Orders confirmed: F1 opens them again."
            : null;
    return menuItems(
      slotTable(box !== null, !moving),
      (a) =>
        a === "move"
          ? own !== undefined &&
            // A unit that never moves (a bunker) has no Move at all.
            (play.typeInfo(this.session, this.side, own.type)?.speed ?? 1) > 0
          : a === "attack"
            ? own !== undefined
            : a === "open"
              ? container !== null
              : a === "produce"
                ? this.contents.options().length > 0
                : a === "exit" || a === "repair",
      (a) => {
        const c = this.contents.actions();
        if (a === "exit") return c.leave;
        if (a === "repair") return c.repair;
        if (a === "produce") return c.produce;
        const busy = held();
        if (busy !== null) return busy;
        if (a === "open")
          return container !== null &&
            (container.slots.length > 0 ||
              play.buildable(this.session, this.side, container.hold).length >
                0)
            ? null
            : "Nothing inside.";
        if (own === undefined) return "No unit of yours here.";
        try {
          play.canOrder(this.session, this.side, own.index);
          if (a === "attack")
            return actionTargets(this.session, this.side, own.index).size > 0
              ? null
              : "No target in range.";
          return play.destinations(this.session, this.side, own.index).size > 0
            ? null
            : "No cell in reach.";
        } catch (e) {
          return e instanceof Error ? e.message : String(e);
        }
      },
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
    if (this.paused || this.disposed) return;
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
    if (this.paused || this.disposed) return;
    if (play.ended(this.session) !== null || this.playback.busy()) return;
    this.clearSelection();
    this.phasePrompt = false;
    if (!this.surrenderPrompt) {
      this.surrenderPrompt = true;
      this.status = `Give up this battle? Press ${this.surrenderKey} again to confirm.`;
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
  /**
   * phase confirms the player's orders. The orders resolve once both sides are ready: at once where the computer
   * has confirmed already, else when it does (tick). A side without a computer gives no orders.
   */
  phase(): Promise<void> {
    if (this.paused || this.disposed) return Promise.resolve();
    if (this.cycle !== undefined) {
      if (!this.canAct()) return Promise.resolve();
      this.clearSelection();
      try {
        this.cycle.confirm(this.side);
      } catch (e) {
        this.refused(e);
      }
      this.touch();
      return Promise.resolve();
    }
    if (!this.canAct()) return Promise.resolve();
    this.clearSelection();
    try {
      play.confirm(this.session, this.side);
      const opponent = 1 - this.side;
      if (this.computer === undefined && !play.ready(this.session, opponent))
        play.confirm(this.session, opponent);
      this.resolveReady();
    } catch (e) {
      if (
        play.ended(this.session) === null &&
        play.ready(this.session, this.side)
      )
        play.unconfirm(this.session, this.side);
      this.refused(e);
    }
    this.touch();
    return Promise.resolve();
  }
  tick(ms: number): void {
    if (this.paused || this.disposed) return;
    const active = this.playback.busy() || this.playback.effects().length > 0,
      scene = this.playback.scene()?.battle,
      wasBusy = this.busy();
    this.playback.advance(ms);
    if (active) this.version++;
    this.computerTick(ms);
    if (scene !== this.playback.scene()?.battle || wasBusy !== this.busy())
      this.onChange?.();
  }
  /**
   * act shows the computer's events of one command on its side's panel: a move rests rest ms on its first hex and
   * then walks (with rest null it only turns), an order turns the unit to its target. It reports whether a move
   * rests or walks now.
   */
  act(events: readonly play.Event[], rest: number | null): boolean {
    if (this.disposed) return false;
    const units = [...this.view().units.values()];
    for (const e of events) {
      if (e.kind === play.Moved) {
        const unit = units.find((u) => u.index === e.unit && u.head === null);
        if (unit !== undefined) this.playback.move(e, unit, rest);
      } else this.turnTo(e);
    }
    this.refresh();
    return this.playback.walking();
  }
  /** A peer's order can change the fair map without changing this panel's UI. */
  refresh(): void {
    if (this.disposed) return;
    this.version++;
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
            actionTargets(this.session, this.side, unit.index).size > 0
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
      this.turnTo(e);
      this.status = e.text || fallback;
    }
  }
  /** turnTo turns the unit of an order towards its target. */
  private turnTo(e: play.Event): void {
    if (e.kind !== play.Ordered && e.kind !== play.Tasked) return;
    const [x, y] = hexgrid.center(e.from),
      [xx, yy] = hexgrid.center(e.to);
    this.playback.directions.set(e.unit, toward({ x, y }, { x: xx, y: yy }));
  }
  private movement(): boolean {
    return play.phase(this.session, this.side) === phaseMovement;
  }
  private canAct(): boolean {
    return (
      !this.busy() &&
      !this.paused &&
      !this.disposed &&
      !this.phasePrompt &&
      !this.surrenderPrompt &&
      play.ended(this.session) === null &&
      !play.ready(this.session, this.side)
    );
  }
  private touch(): void {
    this.refresh();
    if (!this.disposed) this.cycle?.changed(this.side);
  }
  private refused(error: unknown): void {
    this.status = error instanceof Error ? error.message : String(error);
    this.touch();
  }
  /**
   * computerTick runs one frame of the computer side: it starts planning once its phase opens and the resolution
   * has played, takes the worker's answer in after the planning has shown planMinMs, and carries out the commands.
   */
  private computerTick(ms: number): void {
    if (this.computer === undefined || this.cycle !== undefined) return;
    const s = this.session,
      opponent = 1 - this.side;
    if (play.ended(s) !== null) {
      if (this.turn !== null) this.endTurn();
      return;
    }
    const t = this.turn;
    if (t === null) {
      if (play.ready(s, opponent)) this.resolveReady();
      else if (!this.playback.resolving()) this.plan(false);
      return;
    }
    if (t.commands === null) {
      t.elapsed += ms;
      if (
        t.answer === null ||
        t.elapsed < (this.playback.reducedMotion ? 0 : planMinMs)
      )
        return;
      if ("error" in t.answer) {
        this.failed(t, t.answer.error);
        return;
      }
      t.commands = [...t.answer.commands];
      this.progress = null;
      this.touch();
    }
    this.carryOut(t, t.commands);
  }
  /** plan asks the worker for the computer side's commands on a snapshot of the session as it is now. */
  private plan(retry: boolean): void {
    const computer = this.computer;
    if (computer === undefined) return;
    const t: Turn = { retry, elapsed: 0, answer: null, commands: null };
    this.turn = t;
    this.progress = { done: 0, total: 1 };
    let pending: Promise<play.Command[]>;
    try {
      pending = computer.commands(
        this.session,
        1 - this.side,
        (done, total) => {
          if (this.turn !== t) return;
          this.progress = { done, total };
          this.touch();
        },
      );
    } catch (e) {
      pending = Promise.reject(e instanceof Error ? e : new Error(String(e)));
    }
    pending.then(
      (commands) => {
        if (this.turn === t) t.answer = { commands };
      },
      (error: unknown) => {
        if (this.turn === t) t.answer = { error };
      },
    );
    this.touch();
  }
  /**
   * carryOut replays the computer's next commands on the live session, one by one up to the next move its panel
   * walks, once the panel has played the last; after the last command the side confirms.
   */
  private carryOut(t: Turn, commands: play.Command[]): void {
    const stage = this.stage;
    stage?.pace(play.ready(this.session, this.side) ? actFastSpeed : 1);
    if (stage?.busy() === true) return;
    while (commands.length > 0) {
      const command = commands.slice(0, 1);
      commands.shift();
      let events: play.Event[];
      try {
        events = play.replay(this.session, command);
      } catch (e) {
        this.failed(t, e);
        return;
      }
      if (stage?.act(events, actRestMs) === true) {
        this.touch();
        return;
      }
    }
    this.endTurn();
  }
  /**
   * failed plans once more on the session as it is now after a failed planning or command; after a failed retry
   * the side confirms the orders it has given.
   */
  private failed(t: Turn, error: unknown): void {
    if (!t.retry) {
      this.stage?.pace(1);
      this.plan(true);
      return;
    }
    this.status = `Computer side: ${error instanceof Error ? error.message : String(error)}`;
    this.endTurn();
  }
  /** endTurn ends the computer's turn: the side confirms, and the orders resolve if the player is ready too. */
  private endTurn(): void {
    const opponent = 1 - this.side;
    this.turn = null;
    this.progress = null;
    this.stage?.pace(1);
    if (
      play.ended(this.session) === null &&
      !play.ready(this.session, opponent)
    )
      play.confirm(this.session, opponent);
    this.resolveReady();
    this.touch();
  }
  /** resolveReady resolves the orders once both sides are ready and plays the resolution back. */
  private resolveReady(): void {
    const s = this.session;
    if (play.ended(s) !== null || !play.ready(s, 0) || !play.ready(s, 1))
      return;
    const before: [play.View, play.View] = [play.view(s, 0), play.view(s, 1)],
      events = play.resolve(s);
    this.playback.resolution(
      before[this.side === 0 ? 0 : 1],
      play.view(s, this.side),
      events,
    );
    this.onResolution?.(before, events);
    this.status = effectStatus(events, this.side) ?? "Orders resolved.";
  }
}

/**
 * actionPrompt words the action phase's prompt for side's unit as the desktop does (dos.go selectUnit): what the
 * unit's order does, a work unit's charges, the tunnel a tunneller digs. It throws why a gas pioneer or a wiring
 * party has no order now.
 */
export function actionPrompt(
  s: play.Session,
  side: number,
  unit: play.UnitView,
  targets: number,
): string {
  const i = unit.index,
    name = unit.name;
  if (play.sapper(s, i)) return `${name}: pick a hex to dig or fill`;
  if (play.builder(s, i)) return `${name}: pick where to build a depot`;
  if (play.gasPioneer(s, i)) {
    play.canRelease(s, side, i);
    return `${name}: pick where to release gas (${String(play.charges(s, i))} left)`;
  }
  if (play.wiring(s, i)) {
    play.canWire(s, side, i);
    return targets === 0
      ? `${name}: no open ground to wire`
      : `${name}: pick where to lay wire (${String(play.charges(s, i))} left)`;
  }
  if (play.tunneller(s, i)) {
    const t = play.tunnelOf(s, i);
    return t === null
      ? `${name}: pick a hex to mine (it blows in ${String(tunnelTurns)} turns if they stay)`
      : `${name}: digging under ${String(t.mine.line)},${String(t.mine.col)} (${String(t.dug)}/${String(t.turns)} turns)`;
  }
  if (play.ambulance(s, i))
    return `${name}: heals the foot units next to it each turn`;
  return `${name}: pick a target`;
}

/**
 * actionTargets are the hexes side's unit may be ordered to in the action phase: none for tunnellers that dig a
 * tunnel already or a field ambulance, which has no order.
 */
export function actionTargets(
  s: play.Session,
  side: number,
  unit: number,
): Set<string> {
  if (play.ambulance(s, unit) || play.tunnelOf(s, unit) !== null)
    return new Set();
  return play.targets(s, side, unit);
}

/**
 * effectStatus words a resolution's lasting work for side (the extension's work units): gas released, wire laid,
 * the side's own tunnels (the enemy's stay hidden), mines, the losses to gas and mines, healing and the clearing;
 * null without any.
 */
export function effectStatus(
  events: readonly play.Event[],
  side: number,
): string | null {
  const parts: string[] = [];
  for (const e of events) {
    switch (e.kind) {
      case play.Released:
        parts.push(e.side === side ? e.text : "The enemy released gas.");
        break;
      case play.Wired:
        parts.push(e.side === side ? e.text : "The enemy laid wire.");
        break;
      case play.Tunnelled:
      case play.TunnelLost:
        if (e.side === side) parts.push(e.text);
        break;
      case play.Gassed:
      case play.Cleared:
      case play.Blown:
      case play.Mined:
      case play.Healed:
        parts.push(e.text);
    }
  }
  return parts.length === 0 ? null : parts.join(" ");
}

/**
 * startHex is where a side's cursor and camera start, as on the desktop (startHex): its own HQ, else its first unit
 * in line and column order, else the middle of the map.
 */
export function startHex(
  view: play.View,
  side: number,
  map: { lines: readonly unknown[][] },
): hexgrid.Coord {
  const hq = [...view.buildings]
    .filter(([, b]) => b.hq && b.owner === side + 1)
    .map(([key]) => play.coordFromKey(key))
    .sort((a, b) => a.line - b.line || a.col - b.col)[0];
  if (hq !== undefined) return { ...hq };
  const unit = [...view.units.values()]
    .filter((u) => u.side === side)
    .map((u) => u.at)
    .sort((a, b) => a.line - b.line || a.col - b.col)[0];
  if (unit !== undefined) return { ...unit };
  const lines = map.lines.length;
  return {
    line: Math.trunc(lines / 2),
    col: Math.trunc((map.lines[0]?.length ?? 1) / 2),
  };
}
