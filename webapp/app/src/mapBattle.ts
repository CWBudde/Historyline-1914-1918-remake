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

export interface MapComputer {
  commands(
    session: play.Session,
    side: number,
    progress?: (done: number, total: number) => void,
  ): Promise<play.Command[]>;
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
  progress: { done: number; total: number } | null = null;
  status = "";
  version = 0;
  onChange: (() => void) | null = null;
  onResolution:
    | ((before: [play.View, play.View], events: readonly play.Event[]) => void)
    | null = null;
  private planning = false;
  private paused = false;
  private disposed = false;
  private resumeGate: Promise<void> | null = null;
  private resume: (() => void) | null = null;

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
    return this.planning || this.playback.busy();
  }
  planningOrders(): boolean {
    return this.planning;
  }
  setPaused(paused: boolean): void {
    if (this.disposed || paused === this.paused) return;
    this.paused = paused;
    if (paused)
      this.resumeGate = new Promise((resolve) => {
        this.resume = resolve;
      });
    else {
      this.resume?.();
      this.resume = null;
      this.resumeGate = null;
    }
    this.touch();
  }
  dispose(): void {
    this.disposed = true;
    this.paused = false;
    this.resume?.();
    this.resume = null;
    this.resumeGate = null;
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
    if (this.paused || this.disposed) return;
    if (this.cycle !== undefined) {
      if (!this.canAct()) return;
      this.clearSelection();
      try {
        this.cycle.confirm(this.side);
      } catch (e) {
        this.refused(e);
      }
      this.touch();
      return;
    }
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
          while (this.isPaused()) await this.resumeGate;
          if (this.isDisposed()) return;
          if (play.ended(this.session) !== null) return;
          play.replay(this.session, commands);
        }
      }
      if (play.ready(this.session, 0) && play.ready(this.session, 1)) {
        const before: [play.View, play.View] = [
            play.view(this.session, 0),
            play.view(this.session, 1),
          ],
          events = play.resolve(this.session);
        this.playback.resolution(
          before[this.side === 0 ? 0 : 1],
          play.view(this.session, this.side),
          events,
        );
        this.onResolution?.(before, events);
        this.status = effectStatus(events, this.side) ?? "Orders resolved.";
      }
    } catch (e) {
      while (this.isPaused()) await this.resumeGate;
      if (this.isDisposed()) return;
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
    if (this.paused || this.disposed) return;
    const active = this.playback.busy() || this.playback.effects().length > 0,
      scene = this.playback.scene()?.battle,
      wasBusy = this.busy();
    this.playback.advance(ms);
    if (active) this.version++;
    if (scene !== this.playback.scene()?.battle || wasBusy !== this.busy())
      this.onChange?.();
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
      !this.paused &&
      !this.disposed &&
      !this.phasePrompt &&
      !this.surrenderPrompt &&
      play.ended(this.session) === null &&
      !play.ready(this.session, this.side)
    );
  }
  private isPaused(): boolean {
    return this.paused;
  }
  private isDisposed(): boolean {
    return this.disposed;
  }
  private touch(): void {
    this.refresh();
    if (!this.disposed) this.cycle?.changed(this.side);
  }
  private refused(error: unknown): void {
    this.status = error instanceof Error ? error.message : String(error);
    this.touch();
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
