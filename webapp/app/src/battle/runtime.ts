import { hexgrid, level, play, phaseMovement } from "@historyline/engine";
import { signal, type Signal } from "@preact/signals";
import type { Application } from "pixi.js";
import type { AssetLoader } from "../assets.ts";
import type { Computer } from "../computer.ts";
import { MapBattle } from "../mapBattle.ts";
import { MapRenderer } from "../mapRenderer.ts";
import { MapInput } from "../mapInput.ts";
import { parseBakedLayers } from "../mapScene.ts";
import type { Point } from "../camera.ts";
import { BattleKeys } from "../input/keys.ts";
import { layoutRing, placement, type Action } from "../ui/ring.ts";
import type { BattleUIProps } from "../ui/types.ts";
import type { UiAssets } from "../ui/assets.ts";
import type { Settings } from "../storage/store.ts";
import { decodeEnvelope } from "../storage/store.ts";
import type { BattleStart } from "../shell/controller.ts";
import { BattleScenes } from "./scene.ts";
import { newBattleSession, type BattleData } from "./session.ts";

export interface BattleHooks {
  changed(): void;
  leave(): void;
  continue(): void;
  fullscreen(): void;
  preference(patch: Partial<Settings>): void;
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
  readonly ring = signal<ReturnType<typeof layoutRing> | null>(null);
  private readonly auxiliary: BattleScenes;
  private readonly keys: BattleKeys;
  private input: MapInput | null = null;
  private ringAt: hexgrid.Coord | null = null;
  private paused = true;
  private stopped = false;
  private preferences: Settings;
  private constructor(
    private readonly app: Application,
    loader: AssetLoader,
    readonly request: BattleStart,
    private readonly data: BattleData,
    preferences: Settings,
    computer: Computer,
    baked: unknown,
    private readonly hooks: BattleHooks,
  ) {
    this.preferences = { ...preferences };
    this.scale = signal(preferences.uiScale);
    this.reduced = signal(preferences.reduced);
    this.scenes = signal(preferences.scenes);
    this.painted = signal(preferences.painted);
    const session = newBattleSession(
      request.mission.file,
      request.mission.mode,
      request.side,
      {
        limit: preferences.limit,
        computer: preferences.opponent === "computer",
      },
      data,
      request.saved === undefined ? undefined : decodeEnvelope(request.saved),
    );
    this.battle = new MapBattle(
      session,
      request.side,
      session.game.players[1 - request.side]?.computer ? computer : undefined,
    );
    this.battle.setPaused(true);
    this.map = new MapRenderer(
      app,
      loader,
      parseBakedLayers(baked),
      this.battle,
      data.families,
      loader.index.levels[request.mission.month]?.backdrop,
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
        this.status(e.message);
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
        const allowed = new Set(this.battle.actions()),
          p = placement(
            play.phase(session, request.side) !== phaseMovement,
            allowed.has("open"),
            this.battle.contents.box !== null,
          );
        return Object.fromEntries(
          Object.entries(p).map(([dir, action]) => [
            dir,
            allowed.has(action) ? action : "cancel",
          ]),
        );
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
  ): Promise<BattleRuntime> {
    signal.throwIfAborted();
    progress(0.1, "Loading the battlefield…");
    const baked = await loader.baked(
      request.mission.month,
      request.mission.mode,
    );
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
      hooks,
    );
    const aborted = () => {
      runtime.dispose();
    };
    signal.addEventListener("abort", aborted, { once: true });
    try {
      progress(0.65, "Preparing map and battle art…");
      await Promise.all([
        runtime.map.initialize(),
        runtime.auxiliary.initialize(),
      ]);
      signal.throwIfAborted();
      progress(1, "Ready.");
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
    this.map.backdrop = settings.painted;
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
    if (!paused)
      this.input = new MapInput(this.app.canvas, this.map.camera, {
        changed: () => {
          this.closeRing();
          this.map.dirty = true;
        },
        hover: (p) => {
          if (this.ring.value === null) this.battle.setCursor(this.at(p));
        },
        click: (p) => {
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
          this.hooks.preference({ painted: !this.painted.value });
        },
        phase: () => {
          void this.battle.requestPhase();
        },
        context: (p) => {
          this.context(p);
        },
        contextCancel: () => {
          this.closeRing();
        },
        keyDown: (key, repeat) => this.keys.down(key, repeat),
        keyUp: (key) => this.keys.up(key),
        reset: () => {
          this.keys.cancel();
          this.closeRing();
        },
      });
  }
  tick(ms: number): void {
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
    this.map.render();
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
        this.hooks.preference({ painted: !this.painted.value });
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
  }
  private cancel(): void {
    if (this.ring.value !== null) this.closeRing();
    else this.battle.cancel();
  }
  private status(message: string): void {
    if (!this.stopped) {
      this.battle.status = message;
      this.hooks.changed();
    }
  }
  private at(p: Point): hexgrid.Coord | null {
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
    const r = this.app.canvas.getBoundingClientRect();
    this.ringAt = at;
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
        this.hooks.preference({ painted: !this.painted.value });
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
}
