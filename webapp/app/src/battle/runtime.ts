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
import { layoutRing, placement, type Action } from "../ui/ring.ts";
import {
  cursorLook,
  fitRing,
  hitRing,
  newRing,
  type Ring,
} from "../gesture.ts";
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
  readonly ring = signal<ReturnType<typeof layoutRing> | null>(null);
  /** The unit families by type id: the art each unit is drawn with. */
  get families(): BattleData["families"] {
    return this.data.families;
  }
  private readonly auxiliary: BattleScenes;
  private readonly keys: BattleKeys;
  private input: MapInput | null = null;
  private ringAt: hexgrid.Coord | null = null;
  /** The ring menu drawn on the map, in world pixels, and the item under the pointer. */
  private ringWorld: Ring | null = null;
  private ringHover: number | null = null;
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
    );
    this.keys = new BattleKeys({
      step: (direction) => {
        this.closeRing();
        this.battle.stepCursor(direction);
        if (this.battle.contents.box === null) this.seeCursor();
      },
      click: () => {
        if (this.ring.value !== null) {
          this.closeRing();
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
    this.closeRing();
    this.auxiliary.refresh();
  }
  setPaused(paused: boolean): void {
    if (this.stopped || paused === this.paused) return;
    this.paused = paused;
    if (paused) {
      this.input?.dispose();
      this.input = null;
      this.keys.cancel();
      this.closeRing();
    }
    this.battle.setPaused(paused);
    if (!paused && this.options.pointer !== false)
      this.input = new MapInput(this.surface(), this.map.camera, {
        keyboard: this.options.keyboard ?? true,
        changed: () => {
          this.closeRing();
          this.map.dirty = true;
        },
        hover: (p) => {
          if (this.options.panel?.hover(p) === true) return;
          if (this.ringWorld !== null) {
            const hit = this.hitRing(p);
            this.ringHover = typeof hit === "number" ? hit : null;
          } else if (this.ring.value === null)
            this.battle.setCursor(this.at(p));
        },
        click: (p) => {
          if (this.options.panel?.click(p) === true) return;
          if (this.ringWorld !== null) {
            const hit = this.hitRing(p),
              action =
                typeof hit === "number" ? this.ringWorld.actions[hit] : null;
            if (action !== undefined && action !== null) this.runRing(action);
            else this.closeRing();
            return;
          }
          if (this.ring.value !== null) {
            this.closeRing();
            return;
          }
          if (this.ended()) {
            this.hooks.continue();
            return;
          }
          const at = this.at(p);
          if (at !== null) this.battle.click(at);
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
        context: (p) => {
          if (this.options.panel?.context(p) === true) return;
          this.context(p);
        },
        contextCancel: () => {
          this.closeRing();
        },
        keyDown: (key, repeat) => this.keyDown(key, repeat),
        keyUp: (key) => this.keyUp(key),
        reset: () => {
          this.keys.cancel();
          this.closeRing();
        },
      });
  }
  keyDown(key: string, repeat = false): boolean {
    return !this.paused && !this.stopped && this.keys.down(key, repeat);
  }
  keyUp(key: string): boolean {
    return !this.paused && !this.stopped && this.keys.up(key);
  }
  cancelInput(): void {
    this.input?.clearKeyboard();
    this.keys.cancel();
    this.closeRing();
  }
  cameraKey(key: string, down: boolean): boolean {
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
    if (draw) {
      this.map.setCursor(this.cursorDraw());
      this.map.render();
    }
  }
  /** cursorDraw is what the side's cursor shows now (cmd/historyline/cursor.go drawCursor, cursorPlate). */
  cursorDraw(): CursorDraw {
    const b = this.battle,
      key = `${b.version}/${this.keys.holding() ?? ""}/${JSON.stringify(this.ringWorld)}/${this.ringHover ?? ""}`;
    if (key !== this.cursorMemo.key)
      this.cursorMemo = { key, draw: this.lookOfCursor() };
    return this.cursorMemo.draw;
  }
  private lookOfCursor(): CursorDraw {
    const b = this.battle;
    if (this.options.cursor === false || this.stopped) return null;
    if (b.contents.box !== null || b.panel !== "orders") return null;
    if (b.playback.scene() !== null) return null;
    if (this.ringWorld !== null)
      return { kind: "menu", ring: this.ringWorld, hover: this.ringHover };
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
  /** hitRing is the ring menu's item under the panel point p, "cancel" on its centre, or null. */
  private hitRing(p: Point): number | "cancel" | null {
    if (this.ringWorld === null) return null;
    const w = this.map.camera.screenToWorld(p.x, p.y);
    return hitRing(this.ringWorld, w.x, w.y);
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
      ring: this.ring,
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
      onRingAction: (action) => {
        this.runRing(action);
      },
      onCloseRing: () => {
        this.closeRing();
      },
      onRingFocus: (action) => {
        const i =
          action === null
            ? -1
            : (this.ringWorld?.actions.indexOf(action) ?? -1);
        this.ringHover = i < 0 ? null : i;
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
    this.closeRing();
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
    else this.battle.fire();
  }
  private runRing(action: Action): void {
    const at = this.ringAt ?? this.battle.cursor;
    this.closeRing();
    this.battle.run(action, at);
  }
  private closeRing(): void {
    this.ring.value = null;
    this.ringAt = null;
    this.ringWorld = null;
    this.ringHover = null;
  }
  private cancel(): void {
    const b = this.battle;
    if (this.ring.value !== null) this.closeRing();
    else if (
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
  private context(p: Point): void {
    if (this.battle.selected !== null || this.battle.playback.busy()) {
      this.battle.cancel();
      this.closeRing();
      return;
    }
    const at = this.at(p);
    if (at === null) return;
    this.battle.setCursor(at);
    const actions = this.battle.actions(at);
    if (!actions.length) return;
    const r = this.surface().getBoundingClientRect();
    this.ringAt = at;
    this.ringWorld = fitRing(
      newRing(...hexgrid.center(at), actions),
      this.map.camera.visible(),
    );
    this.ringHover = null;
    this.ring.value = layoutRing(
      actions,
      {
        x: r.left + p.x / this.map.camera.scale,
        y: r.top + p.y / this.map.camera.scale,
      },
      r,
      this.scale.value,
    );
  }
  private shortcut(key: string): boolean {
    switch (key.toLowerCase()) {
      case "escape":
        this.cancel();
        break;
      case "f1":
        this.closeRing();
        void this.battle.requestPhase();
        break;
      case "f2":
        this.closeRing();
        this.battle.requestSurrender();
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

/** A level's painting is inline in the index or in a file of its own. */
function hasBackdrop(
  level: { backdrop?: unknown; backdropFile?: unknown } | undefined,
): boolean {
  return level?.backdrop !== undefined || level?.backdropFile !== undefined;
}
