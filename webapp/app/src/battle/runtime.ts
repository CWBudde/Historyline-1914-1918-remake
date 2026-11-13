import { hexgrid, level, play, phaseMovement } from "@historyline/engine";
import { signal, type Signal } from "@preact/signals";
import type { Application } from "pixi.js";
import type { AssetLoader } from "../assets.ts";
import type { Pyramid } from "../assetManifest.ts";
import type { Computer } from "../computer.ts";
import { MapBattle } from "../mapBattle.ts";
import { MapRenderer } from "../mapRenderer.ts";
import { MapInput } from "../mapInput.ts";
import { parseBakedLayers } from "../mapScene.ts";
import type { Point } from "../camera.ts";
import { BattleKeys } from "../input/keys.ts";
import { placement } from "../ui/ring.ts";
import {
  fitShift,
  flickSlot,
  hitMenu,
  itemAt,
  layoutMenu,
  slotOfKey,
  type Hit,
  type MenuItem,
  type MenuLayout,
  type Slot,
} from "../ui/orderMenu.ts";
import { cursorLook } from "../gesture.ts";
import type { CursorDraw } from "../mapRenderer.ts";
import type { BattleUIProps } from "../ui/types.ts";
import type { UiAssets } from "../ui/assets.ts";
import type { Settings } from "../storage/store.ts";
import { decodeEnvelope } from "../storage/store.ts";
import type { BattleStart } from "../shell/controller.ts";
import { BattleScenes } from "./scene.ts";
import {
  design,
  nextDesign,
  type Available,
  type Design,
} from "../original/design.ts";
import { hasOriginalArt, originalArt } from "../original/state.ts";
import { missionSession, type BattleData } from "./session.ts";

/** backdropMonth is the month whose painting a mission shows: an original map's month where one was made from it. */
function backdropMonth(m: BattleStart["mission"]): string {
  return m.original?.backdrop ?? m.month;
}
export interface BattleHooks {
  changed(): void;
  leave(): void;
  continue(): void;
  fullscreen(): void;
  preference(patch: Partial<Settings>): void;
}
export interface BattleRuntimeOptions {
  /** Split panels share a session, with their own controller and fair view. */
  battle?: MapBattle;
  surface?: HTMLElement;
  keyboard?: boolean;
  pointer?: boolean;
  hitTest?: (point: Point) => boolean;
  /** A view painted over the panel takes the pointer first (the split screen's contents); true when it took it. */
  panel?: {
    hover(point: Point): boolean;
    click(point: Point): boolean;
    context(point: Point): boolean;
  };
  /** Whether Escape with nothing to go back from opens the game information (and its settings). */
  escapeMenu?: boolean;
  /** Whether this panel shows its side's cursor (the player's panels); default true. */
  cursor?: boolean;
}
/** MenuView is the open order menu as BattleUI draws it (ui/orderMenu.ts). */
export interface MenuView {
  layout: MenuLayout;
  /** What the pointer, the thumb or the keyboard focus is on: it is lit and named on the caption plate. */
  lit: Slot | "centre" | null;
  /** False during the first moments of a flick, so a fast flick runs without the plates flashing. */
  visible: boolean;
  /** Opened from the keyboard: the first plate takes the focus. */
  focus: boolean;
  /** Changes with every opening, so the view knows a new menu from an update. */
  id: number;
}
/** A press that started on a cell with orders, read as a flick until it proves to be a tap. */
interface Flick {
  /** Where it started, in viewport CSS px. */
  start: Point;
  at: hexgrid.Coord;
  touch: boolean;
  /** Whether the menu was already open: then only a press on its centre starts a flick. */
  wasOpen: boolean;
  dragging: boolean;
  /** The plates show once this time (performance.now()) has passed. */
  showAt: number;
  timer: ReturnType<typeof setTimeout> | null;
}
/** How long a flick runs before its plates are drawn, in ms. */
const flickShow = 150;
/** A battle owns its map/input/GPU resources. Document fonts, settings,
 * worker and catalog outlive it, including a paused leave confirmation. */
export class BattleRuntime {
  readonly battle: MapBattle;
  readonly map: MapRenderer;
  readonly scale: Signal<number>;
  readonly reduced: Signal<boolean>;
  readonly scenes: Signal<boolean>;
  readonly painted: Signal<boolean>;
  /** The map design drawn now (original/design.ts). */
  readonly design = signal<Design>("painted");
  /** The order menu round a unit, null while closed. */
  readonly menu = signal<MenuView | null>(null);
  /** Whether the menu board is open (BattleUI MenuBoard). */
  readonly menuOpen = signal(false);
  /** Whether the orders board of the upright layout is open (BattleUI): Info opens it, a tap on the map closes it. */
  readonly sheetOpen = signal(false);
  /** The unit families by type id: the art each unit is drawn with. */
  get families(): BattleData["families"] {
    return this.data.families;
  }
  private readonly auxiliary: BattleScenes;
  private readonly keys: BattleKeys;
  private input: MapInput | null = null;
  /** The cell and orders of the open menu. */
  private menuAt: { at: hexgrid.Coord; items: MenuItem[] } | null = null;
  private menuOpened = 0;
  private flick: Flick | null = null;
  private cursorMemo: { key: string; draw: CursorDraw } = {
    key: "",
    draw: null,
  };
  private paused = true;
  private stopped = false;
  private preferences: Settings;
  private constructor(
    private readonly app: Application,
    private readonly loader: AssetLoader,
    readonly request: BattleStart,
    private readonly data: BattleData,
    preferences: Settings,
    computer: Computer,
    baked: unknown,
    backdrop: Pyramid | undefined,
    private readonly hooks: BattleHooks,
    private readonly options: BattleRuntimeOptions,
  ) {
    this.preferences = { ...preferences };
    this.scale = signal(preferences.uiScale);
    this.reduced = signal(preferences.reduced);
    this.scenes = signal(preferences.scenes);
    this.painted = signal(preferences.painted);
    if (options.battle !== undefined && options.battle.side !== request.side)
      throw Error("The panel controller belongs to another side.");
    const session =
      options.battle?.session ??
      missionSession(
        request.mission,
        request.side,
        {
          limit: preferences.limit,
          computer: preferences.opponent === "computer",
        },
        data,
        request.saved === undefined ? undefined : decodeEnvelope(request.saved),
      );
    this.battle =
      options.battle ??
      new MapBattle(
        session,
        request.side,
        session.game.players[1 - request.side]?.computer ? computer : undefined,
      );
    this.battle.setPaused(true);
    this.map = new MapRenderer(
      app,
      loader,
      baked === null ? null : parseBakedLayers(baked),
      this.battle,
      data.families,
      backdrop,
      (message) => {
        this.status(message);
      },
    );
    this.auxiliary = new BattleScenes(
      this.battle,
      loader,
      data.families,
      () => this.scale.value,
      () => this.map.camera.visible(),
      (at) => {
        this.map.camera.centerOn(...hexgrid.center(at));
        this.battle.setCursor(at);
        this.map.dirty = true;
      },
      (e) => {
        if (navigator.onLine) this.status(e.message);
      },
      () => this.design.value,
    );
    this.keys = new BattleKeys({
      step: (direction) => {
        this.closeMenu();
        this.battle.stepCursor(direction);
        if (this.battle.contents.box === null) this.seeCursor();
      },
      click: () => {
        if (this.menu.value !== null) {
          this.closeMenu();
          return;
        }
        this.fire();
      },
      run: (action) => {
        this.battle.run(action);
      },
      placement: () => {
        // Both native key sets can repeat the mode gesture to confirm. The
        // prompt suppresses ordinary actions, but must retain this gesture.
        if (this.battle.phasePrompt) return { left: "mode" };
        // As on the desktop, a gesture towards an action not possible here runs nothing but its refusal; only
        // a direction that is cancel itself goes back.
        return this.placementAt(this.battle.cursor);
      },
      shortcut: (key) => this.shortcut(key),
    });
    this.battle.onChange = () => {
      if (this.stopped) return;
      this.map.dirty = true;
      this.auxiliary.refresh();
      hooks.changed();
    };
    this.applyPreferences(preferences);
  }
  static async create(
    app: Application,
    loader: AssetLoader,
    request: BattleStart,
    data: BattleData,
    preferences: Settings,
    computer: Computer,
    hooks: BattleHooks,
    signal: AbortSignal,
    progress: (fraction: number, text: string) => void,
    options: BattleRuntimeOptions = {},
  ): Promise<BattleRuntime> {
    signal.throwIfAborted();
    progress(0.1, "Loading the battlefield…");
    // An imported original map has no baked layers: the renderer composes its tiles.
    const [baked, backdrop] = await Promise.all([
      request.mission.original === undefined
        ? loader.baked(request.mission.month, request.mission.mode)
        : null,
      // An original map without a painting of its own has no level in the index.
      loader.index.levels[backdropMonth(request.mission)] === undefined
        ? undefined
        : loader.backdrop(backdropMonth(request.mission), signal),
    ]);
    signal.throwIfAborted();
    progress(
      0.45,
      request.saved ? "Restoring saved orders…" : "Preparing the armies…",
    );
    const runtime = new BattleRuntime(
      app,
      loader,
      request,
      data,
      preferences,
      computer,
      baked,
      backdrop,
      hooks,
      options,
    );
    const aborted = () => {
      runtime.dispose();
    };
    signal.addEventListener("abort", aborted, { once: true });
    try {
      progress(0.65, "Preparing battle art…");
      await runtime.auxiliary.initialize();
      signal.throwIfAborted();
      progress(1, "Battle art ready.");
      return runtime;
    } catch (error) {
      runtime.dispose();
      throw error;
    } finally {
      signal.removeEventListener("abort", aborted);
    }
  }
  applyPreferences(settings: Settings): void {
    this.preferences = { ...settings };
    this.scale.value = settings.uiScale;
    this.reduced.value = settings.reduced;
    this.scenes.value = settings.scenes;
    this.painted.value = settings.painted;
    this.battle.playback.setReducedMotion(settings.reduced);
    this.battle.playback.setScenes(settings.scenes);
    this.design.value = design(settings, this.available());
    this.painted.value = this.design.value === "painted";
    this.map.backdrop = this.design.value === "painted";
    this.map.original = this.design.value === "original" ? originalArt() : null;
    this.map.dirty = true;
    this.closeMenu();
    this.auxiliary.refresh();
  }
  setPaused(paused: boolean): void {
    if (this.stopped || paused === this.paused) return;
    this.paused = paused;
    if (paused) {
      this.input?.dispose();
      this.input = null;
      this.keys.cancel();
      this.closeMenu();
    }
    this.battle.setPaused(paused);
    if (!paused && this.options.pointer !== false)
      this.input = new MapInput(this.surface(), this.map.camera, {
        keyboard: this.options.keyboard ?? true,
        changed: () => {
          this.closeMenu();
          this.map.dirty = true;
        },
        hover: (p) => {
          if (this.options.panel?.hover(p) === true) return;
          const m = this.menu.value;
          if (m !== null) {
            const hit = hitMenu(m.layout, this.css(p));
            this.light(
              hit === "centre" || itemAt(m.layout, hit) !== null ? hit : null,
            );
          } else this.battle.setCursor(this.at(p));
        },
        click: (p, pointerType) => {
          this.tap(p, pointerType);
        },
        flick: {
          start: (p, pointerType) => this.flickStart(p, pointerType),
          move: (p) => {
            this.flickMove(p);
          },
          end: (p, held) => {
            this.flickEnd(p, held);
          },
          cancel: () => {
            this.flickCancel();
          },
        },
        cancel: () => {
          this.cancel();
        },
        design: () => {
          this.cycleDesign();
        },
        phase: () => {
          void this.battle.requestPhase();
        },
        context: (p, pointerType) => {
          if (this.options.panel?.context(p) === true) return;
          this.context(p, pointerType);
        },
        contextCancel: () => {
          if (this.flick === null) this.closeMenu();
        },
        keyDown: (key, repeat) => this.keyDown(key, repeat),
        keyUp: (key) => this.keyUp(key),
        reset: () => {
          this.keys.cancel();
          if (!menuFocused()) this.closeMenu();
        },
      });
  }
  keyDown(key: string, repeat = false): boolean {
    if (this.paused || this.stopped) return false;
    // While the menu is open the hex keys run its slots (and do not pan or leave).
    const slot = this.menu.value === null ? null : slotOfKey(key);
    if (slot !== null) {
      if (!repeat) this.activate(slot);
      return true;
    }
    return this.keys.down(key, repeat);
  }
  keyUp(key: string): boolean {
    return !this.paused && !this.stopped && this.keys.up(key);
  }
  cancelInput(): void {
    this.input?.clearKeyboard();
    this.keys.cancel();
    // The keyboard's focus moving onto a plate drops held keys but keeps the menu it opened.
    if (!menuFocused()) this.closeMenu();
  }
  cameraKey(key: string, down: boolean): boolean {
    if (down && this.menu.value !== null && !this.paused && !this.stopped) {
      const slot = slotOfKey(key);
      if (slot !== null) {
        this.activate(slot);
        return true;
      }
    }
    return (
      !this.paused &&
      !this.stopped &&
      (this.input?.cameraKey(key, down) ?? false)
    );
  }
  saveSide(): 0 | 1 {
    return this.request.side;
  }
  tick(ms: number, draw = true): void {
    if (this.paused || this.stopped) return;
    this.input?.tick(
      ms,
      this.preferences.scroll === "fast"
        ? 2
        : this.preferences.scroll === "slow"
          ? 0.5
          : 1,
    );
    this.battle.tick(ms);
    // A blown mine or a gas release turns the view to it once, when its event plays (dos.go lookAt); in the split
    // layout each side's panel does so on its own playback.
    const look = this.battle.playback.takeLook();
    if (look !== null) {
      this.map.camera.centerOn(...hexgrid.center(look));
      this.map.dirty = true;
    }
    if (draw) {
      this.map.setCursor(this.cursorDraw());
      this.map.render();
    }
  }
  /** cursorDraw is what the side's cursor shows now (cmd/historyline/cursor.go drawCursor, cursorPlate). */
  cursorDraw(): CursorDraw {
    const b = this.battle,
      key = `${b.version}/${this.keys.holding() ?? ""}/${this.menuAt === null ? "" : level.coordKey(this.menuAt.at)}`;
    if (key !== this.cursorMemo.key)
      this.cursorMemo = { key, draw: this.lookOfCursor() };
    return this.cursorMemo.draw;
  }
  private lookOfCursor(): CursorDraw {
    const b = this.battle;
    if (this.options.cursor === false || this.stopped) return null;
    if (b.contents.box !== null || b.panel !== "orders") return null;
    if (b.playback.scene() !== null) return null;
    // The menu's plates are drawn by BattleUI; the unit keeps its cursor ring under them.
    if (this.menuAt !== null) return { kind: "ring", at: this.menuAt.at };
    const at = b.hover ?? b.cursor,
      movement = play.phase(b.session, b.side) === phaseMovement,
      reach = movement ? b.destinations() : b.targets(),
      pick =
        b.selected === null || reach.size === 0
          ? "cancel"
          : b.exiting !== null
            ? "exit"
            : movement
              ? "move"
              : "attack",
      entrance = play.entrance(b.session, at),
      onTarget =
        reach.has(level.coordKey(at)) ||
        (entrance !== null && reach.has(level.coordKey(entrance))),
      held = this.keys.holding(),
      raw = this.placementAt(at),
      picked = held === null ? "cancel" : raw[held],
      allowed = picked === "cancel" || b.actions(at).includes(picked),
      look = cursorLook(
        pick,
        onTarget,
        held !== null,
        held === "centre",
        picked,
        allowed,
      );
    return look === null
      ? { kind: "ring", at }
      : { kind: "plate", at, ...look };
  }
  /** placementAt is the gesture placement on at (gesture.ForMap, ForContents): the left gesture opens an own container. */
  private placementAt(at: hexgrid.Coord): ReturnType<typeof placement> {
    const b = this.battle,
      entrance = play.entrance(b.session, at);
    return placement(
      play.phase(b.session, b.side) !== phaseMovement,
      play.contents(b.session, b.side, entrance ?? at) !== null ||
        [...b.view().units.values()].some(
          (u) =>
            u.side === b.side &&
            u.carrier &&
            level.coordKey(u.at) === level.coordKey(at),
        ),
      b.contents.box !== null,
    );
  }
  resize(width: number, height: number, dpr: number): void {
    if (
      this.map.camera.scale !== dpr ||
      this.map.camera.view.width !== width * dpr ||
      this.map.camera.view.height !== height * dpr
    )
      this.map.resize(width, height, dpr);
  }
  canSave(): boolean {
    return !this.stopped && !this.battle.planningOrders();
  }
  ended(): boolean {
    return play.ended(this.battle.session) !== null && !this.battle.busy();
  }
  props(
    assets: UiAssets,
    revision: Signal<number>,
    settingsState: { busy: boolean; note: string } = { busy: false, note: "" },
  ): BattleUIProps {
    return {
      battle: this.battle,
      assets,
      families: this.data.families,
      revision,
      scale: this.scale,
      reduced: this.reduced,
      scenes: this.scenes,
      painted: this.painted,
      design: this.design,
      menu: this.menu,
      menuOpen: this.menuOpen,
      sheetOpen: this.sheetOpen,
      settingsBusy: settingsState.busy,
      settingsNote: settingsState.note,
      onScale: (uiScale) => {
        this.hooks.preference({ uiScale });
      },
      onReduced: (reduced) => {
        this.hooks.preference({ reduced });
      },
      onScenes: (scenes) => {
        this.hooks.preference({ scenes });
      },
      onDesign: () => {
        this.cycleDesign();
      },
      onLeave: () => {
        this.hooks.leave();
      },
      onFullscreen: () => {
        this.hooks.fullscreen();
      },
      onMenu: (hit) => {
        this.activate(hit);
      },
      onMenuFocus: (hit) => {
        this.light(hit);
      },
      onPhase: () => {
        this.closeMenu();
        void this.battle.requestPhase();
      },
      onMap: () => {
        this.closeMenu();
        this.battle.run("map");
      },
      onInfo: () => {
        this.info();
      },
      onSceneHost: (host) => {
        this.auxiliary.attach("battle", host);
      },
      onStrategyHost: (host) => {
        this.auxiliary.attach("strategy", host);
      },
    };
  }
  dispose(): void {
    if (this.stopped) return;
    this.stopped = true;
    this.input?.dispose();
    this.input = null;
    this.keys.cancel();
    this.closeMenu();
    this.battle.dispose();
    this.auxiliary.dispose();
    this.map.dispose();
  }
  /** available are the designs this battle can draw: a painting where its month has one, imported original art. */
  available(): Available {
    return {
      painted: hasBackdrop(
        this.loader.index.levels[backdropMonth(this.request.mission)],
      ),
      original: hasOriginalArt(),
      originalMap: this.request.mission.original !== undefined,
    };
  }
  /** cycleDesign is F8: the next available map design. */
  cycleDesign(): void {
    this.hooks.preference(nextDesign(this.preferences, this.available()));
  }
  private fire(): void {
    if (this.ended()) this.hooks.continue();
    // Enter on the selected unit itself opens its menu, as a click on it does.
    else if (this.onSelected(this.battle.cursor))
      this.openMenu(this.battle.cursor, { touch: coarse(), focus: true });
    else this.battle.fire();
  }
  /** info is the strip's Info: the unit or cell under the cursor, or the game where it has nothing to show. */
  private info(): void {
    this.closeMenu();
    const b = this.battle;
    if (b.panel === "unit" || b.panel === "game") b.cancel();
    else b.run("info", b.cursor);
  }
  /** onSelected reports whether at holds the selected unit. */
  private onSelected(at: hexgrid.Coord): boolean {
    return this.battle.holdsSelected(at);
  }

  /**
   * openMenu opens the order menu of at, if it has orders: plates sized for touch or a mouse, the camera panned
   * first where a plate would leave the view, then the whole rosette moved inwards. False when at has none.
   */
  private openMenu(
    cell: hexgrid.Coord,
    o: { touch: boolean; focus?: boolean; visible?: boolean },
  ): boolean {
    const items = this.battle.menu(cell);
    if (items.length === 0) return false;
    // Round the unit itself, also when its rear cell was hit.
    const at = this.battle.unitCell(cell);
    this.battle.setCursor(at);

    this.menuAt = { at, items };
    this.menuOpened++;
    this.menu.value = {
      layout: this.placeMenu(at, items, o.touch),
      lit: null,
      visible: o.visible ?? true,
      focus: o.focus ?? false,
      id: this.menuOpened,
    };
    this.map.dirty = true;
    return true;
  }
  /** placeMenu lays the menu of at out in viewport CSS px, panning the camera first if a plate is outside the view. */
  private placeMenu(
    at: hexgrid.Coord,
    items: MenuItem[],
    touch: boolean,
  ): MenuLayout {
    const camera = this.map.camera,
      dpr = camera.scale > 0 ? camera.scale : 1,
      cell = (hexgrid.hexWidth * camera.effective()) / dpr,
      r = this.surface().getBoundingClientRect(),
      inset = 4,
      x0 = Math.max(r.left, 0) + inset,
      y0 = Math.max(r.top, 0) + inset,
      bounds = {
        x: x0,
        y: y0,
        width: Math.min(r.right, window.innerWidth) - inset - x0,
        height: Math.min(r.bottom, window.innerHeight) - inset - y0,
      },
      unitAt = () => {
        const s = camera.worldToScreen(...hexgrid.center(at));
        return { x: r.left + s.x / dpr, y: r.top + s.y / dpr };
      };
    let m = layoutMenu(items, unitAt(), cell, touch, bounds);
    // The pan keeps the unit itself in view too, not only its plates.
    const pan = fitShift(
        [m.unit, ...items.map((i) => itemAt(m, i.slot) ?? m.unit)].map((p) => ({
          x: p.x - m.centre.x + m.unit.x,
          y: p.y - m.centre.y + m.unit.y,
        })),
        m.plate.width,
        m.plate.height,
        bounds,
      ),
      dx = pan.x,
      dy = pan.y;
    if (dx !== 0 || dy !== 0) {
      // Pan first, as far as the map goes; whatever is left moves the rosette.
      camera.scroll(-dx * dpr, -dy * dpr);
      this.map.dirty = true;
      m = layoutMenu(items, unitAt(), cell, touch, bounds);
    }
    return m;
  }
  /** light marks what the pointer or the focus is on. */
  private light(hit: Hit): void {
    const m = this.menu.value;
    if (m === null || m.lit === hit) return;
    this.menu.value = { ...m, lit: hit };
  }
  /**
   * activate runs what a menu slot holds: a ready order closes the menu and runs; one not possible now stays lit
   * and names its reason; the centre cancels (the selection too); an empty slot does nothing.
   */
  private activate(hit: Hit): void {
    const m = this.menu.value,
      open = this.menuAt;
    if (m === null || open === null) return;
    if (hit === "centre") {
      this.closeMenu();
      if (this.battle.selected !== null) this.battle.cancel();
      return;
    }
    const item = itemAt(m.layout, hit);
    if (item === null) return;
    if (item.reason !== null) {
      this.menu.value = { ...m, lit: item.slot, visible: true };
      this.status(reasonText(item.reason));
      return;
    }
    this.closeMenu();
    this.battle.run(item.action, open.at);
  }
  private closeMenu(): void {
    if (this.flick?.timer != null) clearTimeout(this.flick.timer);
    this.flick = null;
    if (this.menu.value === null && this.menuAt === null) return;
    this.menu.value = null;
    this.menuAt = null;
    this.map.dirty = true;
  }
  /** css is the panel point p (device px) in viewport CSS px. */
  private css(p: Point): Point {
    const r = this.surface().getBoundingClientRect(),
      dpr = this.map.camera.scale > 0 ? this.map.camera.scale : 1;
    return { x: r.left + p.x / dpr, y: r.top + p.y / dpr };
  }
  /**
   * tap is a click or tap on the map. With the menu open it runs the wedge's order, cancels on the centre, and
   * outside the menu closes it and goes on as a plain click where a unit is selected (a tap on a destination
   * moves there). A click on the selected unit opens its menu; a tap on an own unit selects it and opens it, and
   * so does a click on a loaded carrier that has its order, for its contents.
   */
  private tap(p: Point, pointerType: string): void {
    this.sheetOpen.value = false;
    if (this.options.panel?.click(p) === true) return;
    const m = this.menu.value;
    if (m !== null) {
      const hit = hitMenu(m.layout, this.css(p));
      if (hit === "centre" || itemAt(m.layout, hit) !== null) {
        this.activate(hit);
        return;
      }
      this.closeMenu();
      if (this.battle.selected === null) return;
    }
    if (this.ended()) {
      this.hooks.continue();
      return;
    }
    const at = this.at(p);
    if (at === null) return;
    const touch = pointerType !== "mouse";
    if (this.onSelected(at)) {
      this.openMenu(at, { touch });
      return;
    }
    this.battle.click(at);
    if ((touch && this.onSelected(at)) || this.battle.opensOnly(at))
      this.openMenu(at, { touch });
  }
  private flickStart(p: Point, pointerType: string): boolean {
    const b = this.battle;
    if (this.paused || this.stopped || b.contents.box !== null) return false;
    if (b.panel !== "orders" || b.busy()) return false;
    const start = this.css(p),
      m = this.menu.value,
      touch = pointerType !== "mouse";
    let at: hexgrid.Coord | null;
    if (m !== null && this.menuAt !== null) {
      if (hitMenu(m.layout, start) !== "centre") return false;
      at = this.menuAt.at;
    } else {
      at = this.at(p);
      if (at === null || b.menu(at).length === 0) return false;
    }
    const f: Flick = {
      start,
      at,
      touch,
      wasOpen: m !== null,
      dragging: false,
      showAt: performance.now() + flickShow,
      timer: null,
    };
    f.timer = setTimeout(() => {
      f.timer = null;
      const v = this.menu.value;
      if (this.flick === f && v !== null && !v.visible)
        this.menu.value = { ...v, visible: true };
    }, flickShow);
    this.flick = f;
    return true;
  }
  private flickMove(p: Point): void {
    const f = this.flick;
    if (f === null) return;
    const to = this.css(p),
      slot = flickSlot(f.start, to);
    if (slot === null) {
      if (f.dragging) this.light("centre");
      return;
    }
    if (!f.dragging) {
      f.dragging = true;
      if (
        this.menu.value === null &&
        !this.openMenu(f.at, {
          touch: f.touch,
          visible: performance.now() >= f.showAt,
        })
      ) {
        this.flickCancel();
        return;
      }
    }
    const m = this.menu.value;
    if (m !== null) this.light(itemAt(m.layout, slot) === null ? null : slot);
  }
  private flickEnd(p: Point, held: boolean): void {
    const f = this.flick;
    if (f === null) return;
    if (f.timer !== null) clearTimeout(f.timer);
    this.flick = null;
    if (!f.dragging) {
      // A press that never became a flick: a tap, unless a long press already opened the menu.
      if (!held) this.tap(p, f.touch ? "touch" : "mouse");
      return;
    }
    const m = this.menu.value,
      slot = flickSlot(f.start, this.css(p));
    if (m === null) return;
    if (slot === null || itemAt(m.layout, slot) === null) {
      // Dragged back to the unit or towards a free cell: nothing runs; a menu opened by the flick goes.
      if (!f.wasOpen && !held) this.closeMenu();
      return;
    }
    this.activate(slot);
  }
  private flickCancel(): void {
    const f = this.flick;
    if (f === null) return;
    if (f.timer !== null) clearTimeout(f.timer);
    this.flick = null;
    if (!f.wasOpen) this.closeMenu();
  }
  private cancel(): void {
    const b = this.battle;
    if (this.menu.value !== null) this.closeMenu();
    else if (this.menuOpen.value) {
      this.menuOpen.value = false;
      this.hooks.changed();
    } else if (this.sheetOpen.value) {
      this.sheetOpen.value = false;
      this.hooks.changed();
    } else if (
      this.options.escapeMenu === true &&
      b.selected === null &&
      b.contents.box === null &&
      b.panel === "orders" &&
      !b.phasePrompt &&
      !b.surrenderPrompt &&
      !b.playback.busy()
    )
      b.showGame();
    else b.cancel();
  }
  private status(message: string): void {
    if (!this.stopped) {
      this.battle.status = message;
      this.hooks.changed();
    }
  }
  private at(p: Point): hexgrid.Coord | null {
    if (this.options.hitTest?.(p) === false) return null;
    const w = this.map.camera.screenToWorld(p.x, p.y),
      [at, onMap] = hexgrid.pixelToHex(
        level.grid(this.battle.session.level),
        w.x,
        w.y,
      );
    return onMap ? at : null;
  }
  private seeCursor(): void {
    const [x, y] = hexgrid.center(this.battle.cursor),
      v = this.map.camera.visible();
    if (
      x < v.x + 48 ||
      x > v.x + v.width - 48 ||
      y < v.y + 40 ||
      y > v.y + v.height - 40
    )
      this.map.camera.centerOn(x, y);
    this.map.dirty = true;
  }
  /**
   * context is a right click or a long press. A long press during a press on a unit opens its menu (or, on a plate,
   * explains it); a right click closes an open menu, steps back from a selection, or opens the cell's menu.
   */
  private context(p: Point, pointerType: string): void {
    const f = this.flick,
      m = this.menu.value;
    if (f !== null) {
      if (m === null) this.openMenu(f.at, { touch: f.touch });
      else this.menu.value = { ...m, visible: true };
      return;
    }
    if (m !== null) {
      const hit = hitMenu(m.layout, this.css(p));
      if (pointerType === "touch" && itemAt(m.layout, hit) !== null) {
        // A long press on a plate explains it.
        this.light(hit);
        const item = itemAt(m.layout, hit);
        if (item?.reason != null) this.status(reasonText(item.reason));
        return;
      }
      this.closeMenu();
      return;
    }
    if (this.battle.selected !== null || this.battle.playback.busy()) {
      this.battle.cancel();
      return;
    }
    const at = this.at(p);
    if (at === null) return;
    this.battle.setCursor(at);
    this.openMenu(at, { touch: pointerType === "touch" });
  }
  private shortcut(key: string): boolean {
    switch (key.toLowerCase()) {
      case "escape":
        this.cancel();
        break;
      case "f1":
        this.closeMenu();
        void this.battle.requestPhase();
        break;
      case "f2":
        this.closeMenu();
        this.battle.requestSurrender();
        break;
      case "i":
        this.info();
        break;
      case "o":
        this.closeMenu();
        if (this.battle.panel === "map") this.battle.cancel();
        else this.battle.run("map");
        break;
      case "contextmenu":
        if (this.menu.value !== null) this.closeMenu();
        else
          this.openMenu(this.battle.cursor, { touch: coarse(), focus: true });
        break;
      case "f8":
        this.cycleDesign();
        break;
      case "f9":
        this.hooks.preference({ scenes: !this.scenes.value });
        break;
      case "m":
        this.hooks.preference({ reduced: !this.reduced.value });
        break;
      case "f11":
        this.hooks.fullscreen();
        break;
      case "q":
        this.hooks.leave();
        break;
      case "home":
      case "pageup":
        this.map.camera.centerOn(this.map.camera.world.width / 2, 0);
        this.map.dirty = true;
        break;
      case "end":
      case "pagedown":
        this.map.camera.centerOn(
          this.map.camera.world.width / 2,
          this.map.camera.world.height,
        );
        this.map.dirty = true;
        break;
      default:
        return false;
    }
    return true;
  }
  private surface(): HTMLElement {
    return this.options.surface ?? this.app.canvas;
  }
}

/** menuFocused reports whether the keyboard focus is on the order menu. */
function menuFocused(): boolean {
  return (
    typeof document !== "undefined" &&
    document.activeElement?.closest(".order-menu") != null
  );
}
/** coarse reports
 whether the main pointer is a finger: menus opened from the keyboard then get touch-sized plates. */
function coarse(): boolean {
  return (
    typeof matchMedia === "function" && matchMedia("(pointer: coarse)").matches
  );
}
/** reasonText is a refusal as a sentence: capitalised, with a full stop. */
export function reasonText(reason: string): string {
  const t = reason.trim();
  if (t === "") return t;
  return (
    t.charAt(0).toUpperCase() + t.slice(1) + (/[.!?]$/u.test(t) ? "" : ".")
  );
}

/** A level's painting is inline in the index or in a file of its own. */
function hasBackdrop(
  level: { backdrop?: unknown; backdropFile?: unknown } | undefined,
): boolean {
  return level?.backdrop !== undefined || level?.backdropFile !== undefined;
}
