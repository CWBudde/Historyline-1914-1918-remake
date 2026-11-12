import { bind, hexgrid, level, play, phaseMovement } from "@historyline/engine";
import { Container, Graphics, type Application } from "pixi.js";
import type { Signal } from "@preact/signals";
import type { AssetLoader } from "../assets.ts";
import type { Computer } from "../computer.ts";
import { BattleRuntime, type BattleHooks } from "../battle/runtime.ts";
import { missionSession, type BattleData } from "../battle/session.ts";
import { MapBattle } from "../mapBattle.ts";
import {
  decodeEnvelope,
  splitByDefault,
  type Layout,
  type Settings,
} from "../storage/store.ts";
import type { BattleStart } from "../shell/controller.ts";
import type { UiAssets } from "../ui/assets.ts";
import type { BattleUIProps } from "../ui/types.ts";
import type { Point, Rect } from "../camera.ts";
import { LocalMatch } from "./match.ts";
import { SplitFrame } from "./frame.ts";
import {
  layout,
  defaultZoom,
  zooms,
  type Screen,
  type View,
} from "./geometry.ts";
import {
  alignCamera,
  cameraRest,
  centreView,
  pairRange,
  rowRange,
} from "./camera.ts";
import { SplitInput } from "./input.ts";
import { loadPixelFonts, type PixelFonts } from "../pixfont.ts";
import { messageMs, wellOf, type Well } from "./well.ts";

const sides = [0, 1] as const;
function position(element: HTMLElement, rect: Rect, dpr: number): void {
  Object.assign(element.style, {
    left: `${rect.x / dpr}px`,
    top: `${rect.y / dpr}px`,
    width: `${rect.width / dpr}px`,
    height: `${rect.height / dpr}px`,
  });
}

/** The document's battle host. Both layouts keep the same session and panel
 * controllers when the layout changes; only the viewports and DOM hosts move. */
export class SplitRuntime {
  private readonly wrappers = [new Container(), new Container()] as const;
  private readonly masks = [new Graphics(), new Graphics()] as const;
  private readonly surfaces: [HTMLDivElement, HTMLDivElement];
  readonly frame: SplitFrame;
  screen: Screen;
  views: [View, View];
  private width = 1;
  private height = 1;
  private dpr = 1;
  private zoom = defaultZoom;
  private zoomIntent = 1;
  private intents: [Point | null, Point | null] = [null, null];
  private aligned: [Point | null, Point | null] = [null, null];
  private alignedZoom: [number, number] = [1, 1];
  private paused = true;
  private stopped = false;
  private updating = false;
  private readonly keyboard: SplitInput;
  private frameGeneration = 0;
  private layoutKey = "";
  /** The pixel font of the wells and overlays. */
  fonts: PixelFonts | null = null;
  /** Each side's last message and when it came, so that it beats the unit line only while fresh. */
  private readonly messages = [
    { text: "", at: 0, fresh: false },
    { text: "", at: 0, fresh: false },
  ];
  private constructor(
    private readonly app: Application,
    loader: AssetLoader,
    readonly request: BattleStart,
    readonly panels: [BattleRuntime, BattleRuntime],
    readonly local: LocalMatch | null,
    private settings: Settings,
    private readonly hooks: BattleHooks,
    surfaces: [HTMLDivElement, HTMLDivElement],
  ) {
    this.surfaces = surfaces;
    this.frame = new SplitFrame(loader, () => {
      this.hooks.changed();
    });
    this.frame.canvas.className = "split-frame";
    this.screen = layout(1, 1);
    this.views = [this.screen.view(0, 0, 0), this.screen.view(1, 0, 0)];
    const stage = this.app.canvas.parentElement;
    if (stage === null) throw Error("The map stage is unavailable.");
    for (const side of sides) {
      const wrapper = this.wrappers[side],
        mask = this.masks[side];
      wrapper.addChild(this.panels[side].map.root, mask);
      wrapper.mask = mask;
      app.stage.addChild(wrapper);
      stage.appendChild(surfaces[side]);
    }
    stage.appendChild(this.frame.canvas);
    if (local === null) {
      const primary = this.panels[request.side].battle,
        peer = this.panels[request.side === 0 ? 1 : 0].battle;
      primary.onResolution = (before, events) => {
        peer.playback.resolution(
          before[peer.side === 0 ? 0 : 1],
          play.view(peer.session, peer.side),
          events,
        );
      };
    }
    this.keyboard = this.installKeyboard();
    this.resize(app.screen.width, app.screen.height, app.renderer.resolution);
  }
  static async create(
    app: Application,
    loader: AssetLoader,
    request: BattleStart,
    data: BattleData,
    settings: Settings,
    computer: Computer,
    hooks: BattleHooks,
    signal: AbortSignal,
    progress: (fraction: number, text: string) => void,
  ): Promise<SplitRuntime> {
    signal.throwIfAborted();
    const session = missionSession(
      request.mission,
      request.side,
      {
        limit: settings.limit,
        computer:
          request.mission.mode !== "2p" && settings.opponent === "computer",
      },
      data,
      request.saved === undefined ? undefined : decodeEnvelope(request.saved),
    );
    const local =
      request.mission.mode === "2p" &&
      !session.game.players.some((p) => p.computer)
        ? new LocalMatch(
            session,
            settings.playMode === "hotseat" ? "hotseat" : "simultaneous",
            request.side,
          )
        : null;
    const controllers: [MapBattle, MapBattle] = local?.sides ?? [
      new MapBattle(
        session,
        0,
        request.side === 0 && session.game.players[1].computer
          ? computer
          : undefined,
      ),
      new MapBattle(
        session,
        1,
        request.side === 1 && session.game.players[0].computer
          ? computer
          : undefined,
      ),
    ];
    const surfaces: [HTMLDivElement, HTMLDivElement] = [
      document.createElement("div"),
      document.createElement("div"),
    ];
    for (const [side, surface] of surfaces.entries()) {
      surface.className = "split-input";
      surface.dataset.side = String(side);
      surface.setAttribute(
        "aria-label",
        `${side === 0 ? "Central Powers" : "Entente"} map`,
      );
    }
    let host: SplitRuntime | null = null;
    const made: BattleRuntime[] = [];
    try {
      // Sequential resource construction keeps cancellation ownership explicit.
      for (const side of sides) {
        const panel = await BattleRuntime.create(
          app,
          loader,
          { ...request, side },
          data,
          settings,
          computer,
          {
            ...hooks,
            changed: () => {
              host?.panelChanged(side);
            },
          },
          signal,
          (n, text) => {
            progress((side + n) / 3, text);
          },
          {
            battle: controllers[side],
            surface: surfaces[side],
            keyboard: false,
            pointer:
              local?.mode === "simultaneous"
                ? side === 0
                : local !== null || side === request.side,
            hitTest: (point) => host?.hitTest(side, point) ?? true,
          },
        );
        made.push(panel);
      }
      const [a, b] = made;
      if (a === undefined || b === undefined)
        throw Error("Both battle panels are required.");
      signal.throwIfAborted();
      host = new SplitRuntime(
        app,
        loader,
        request,
        [a, b],
        local,
        { ...settings },
        hooks,
        surfaces,
      );
      const ready = host;
      const aborted = () => {
        ready.dispose();
      };
      signal.addEventListener("abort", aborted, { once: true });
      try {
        signal.throwIfAborted();
        progress(0.7, "Loading map art…");
        // The split layout has now set each camera's real viewport and zoom;
        // select catalogue densities once, without fetching the full canvas's
        // unused tier before the panels are aligned.
        const [, , fonts] = await Promise.all([
          a.map.initialize(),
          b.map.initialize(),
          loadPixelFonts(loader),
        ]);
        ready.fonts = fonts;
        signal.throwIfAborted();
        progress(0.8, "Drawing the battlefield…");
        return ready;
      } finally {
        signal.removeEventListener("abort", aborted);
      }
    } catch (error) {
      host?.dispose();
      for (const panel of made) panel.dispose();
      local?.dispose();
      for (const surface of surfaces) surface.remove();
      throw error;
    }
  }
  /** paintedDesign reports whether the map shows its painting, which the split screen frames slim. */
  private get paintedDesign(): boolean {
    return this.panels[0].design.value === "painted";
  }
  /** original reports whether the battle shows the original split screen (the auto layout on a wide window). */
  get original(): boolean {
    const layout = this.settings.layout;
    return (
      layout === "original" ||
      (layout === "auto" &&
        typeof window !== "undefined" &&
        splitByDefault(window.innerWidth, window.innerHeight)) ||
      this.local?.mode === "simultaneous"
    );
  }
  get activeSide(): 0 | 1 {
    return this.local?.activeSide ?? this.request.side;
  }
  get handover(): 0 | 1 | null {
    return this.local?.handover ?? null;
  }
  get battle(): MapBattle {
    return this.panels[this.activeSide].battle;
  }
  get map(): BattleRuntime["map"] {
    return this.panels[this.activeSide].map;
  }
  get ready(): boolean {
    const visible = sides.filter((side) => this.visible(side));
    return (
      !this.stopped &&
      visible.length > 0 &&
      visible.every((side) => this.panels[side].map.ready)
    );
  }
  saveSide(): 0 | 1 {
    return this.local?.saveSide() ?? this.request.side;
  }
  canSave(): boolean {
    return this.panels.every((p) => p.canSave());
  }
  ended(): boolean {
    return (
      play.ended(this.battle.session) !== null &&
      this.handover === null &&
      (this.local?.mode === "hotseat"
        ? !this.battle.busy()
        : !this.panels.some((p) => p.battle.busy()))
    );
  }
  canInteract(side: 0 | 1): boolean {
    return (
      !this.paused &&
      this.handover === null &&
      (this.local?.mode === "simultaneous" || side === this.activeSide)
    );
  }
  visible(side: 0 | 1): boolean {
    if (this.handover !== null) return false;
    if (this.local?.mode === "hotseat") return side === this.activeSide;
    return this.original || side === this.activeSide;
  }
  takeOver(): void {
    if (this.paused || this.stopped) return;
    this.keyboard.reset();
    this.local?.takeOver();
    this.syncPause();
    this.hooks.changed();
  }
  requestLeave(): void {
    this.hooks.leave();
  }
  setPaused(paused: boolean): void {
    if (this.stopped || paused === this.paused) return;
    this.paused = paused;
    this.syncPause();
  }
  private syncPause(): void {
    if (this.updating) return;
    this.updating = true;
    try {
      if (this.paused || this.handover !== null) this.keyboard.reset();
      this.local?.setPaused(this.paused);
      for (const side of sides) {
        const hiddenHuman =
          this.local?.mode === "hotseat" &&
          (this.handover !== null || side !== this.activeSide);
        this.panels[side].setPaused(this.paused || hiddenHuman);
        this.surfaces[side].style.pointerEvents = this.canInteract(side)
          ? "auto"
          : "none";
      }
    } finally {
      this.updating = false;
    }
  }
  private panelChanged(side: 0 | 1): void {
    if (this.stopped || this.updating) return;
    if (this.local === null) {
      this.updating = true;
      const primary = this.panels[side].battle,
        peer = this.panels[side === 0 ? 1 : 0].battle,
        scene = primary.playback.scene()?.battle,
        key =
          scene === undefined
            ? null
            : `${scene.attacker.unit.index}/${scene.defender.unit.index}`;
      if (key !== null || !primary.busy()) {
        let other = peer.playback.scene()?.battle;
        while (
          other !== undefined &&
          `${other.attacker.unit.index}/${other.defender.unit.index}` !== key
        ) {
          if (!peer.playback.skipScene()) break;
          other = peer.playback.scene()?.battle;
        }
      }
      peer.refresh();
      this.updating = false;
    }
    this.syncPause();
    this.hooks.changed();
  }
  applyPreferences(settings: Settings): void {
    const geometry =
      this.settings.layout !== settings.layout ||
      this.settings.uiScale !== settings.uiScale ||
      this.settings.painted !== settings.painted ||
      this.settings.originalArt !== settings.originalArt;
    this.settings = { ...settings };
    for (const panel of this.panels) panel.applyPreferences(settings);
    if (geometry) {
      this.intents = [null, null];
      this.layoutKey = "";
    }
    this.resize(this.width, this.height, this.dpr);
    this.hooks.changed();
  }
  props(
    assets: UiAssets,
    revision: Signal<number>,
    state = { busy: false, note: "" },
  ): BattleUIProps {
    return this.withLayout(
      this.panels[this.activeSide].props(assets, revision, state),
    );
  }
  panelProps(
    side: 0 | 1,
    assets: UiAssets,
    revision: Signal<number>,
    state: { busy: boolean; note: string },
  ): BattleUIProps {
    return {
      ...this.withLayout(this.panels[side].props(assets, revision, state)),
      showScene: this.local !== null || side !== this.request.side,
      readOnly: this.local === null && side !== this.request.side,
    };
  }
  attachFrame(node: HTMLDivElement | null): void {
    if (node === null) this.frame.canvas.remove();
    else if (!this.stopped) node.appendChild(this.frame.canvas);
  }
  private withLayout(props: BattleUIProps): BattleUIProps {
    return {
      ...props,
      layout: this.settings.layout,
      ...(this.local?.mode === "simultaneous"
        ? {}
        : {
            onLayout: (value: Layout) => {
              this.hooks.preference({ layout: value });
            },
          }),
    };
  }
  resize(width: number, height: number, dpr: number): void {
    if (this.stopped) return;
    const key = `${width}/${height}/${dpr}/${this.original}/${this.settings.uiScale}/${this.paintedDesign}`;
    if (key === this.layoutKey) return;
    this.layoutKey = key;
    const changed =
      this.width !== width || this.height !== height || this.dpr !== dpr;
    this.width = width;
    this.height = height;
    this.dpr = dpr;
    if (changed) this.intents = [null, null];
    if (!this.original) {
      this.frame.canvas.hidden = true;
      for (const side of sides) {
        this.wrappers[side].position.set(0, 0);
        this.masks[side].clear().rect(0, 0, width, height).fill(0xffffff);
        this.panels[side].map.camera.bounded = true;
        this.panels[side].map.camera.viewport(width, height, dpr);
        this.panels[side].resize(width, height, dpr);
        position(
          this.surfaces[side],
          { x: 0, y: 0, width: width * dpr, height: height * dpr },
          dpr,
        );
      }
    } else {
      this.frame.canvas.hidden = false;
      const base = layout(
        Math.round(width * dpr),
        Math.round(height * dpr),
        this.settings.uiScale * dpr,
      );
      const grid = bind.playable(this.battle.session.board),
        [ww, hh] = hexgrid.bounds(grid);
      this.zoom = base.clampZoom(this.zoom);
      if (this.paintedDesign)
        this.zoom = Math.max(this.zoom, base.fillZoom(ww, hh));
      this.screen = base.zoomed(this.zoom);
      this.zoomIntent = zooms[this.zoom] ?? 1;
      for (const side of sides) {
        const panel = this.screen.panels[side];
        this.wrappers[side].position.set(panel.x / dpr, panel.y / dpr);
        this.masks[side]
          .clear()
          .rect(0, 0, panel.width / dpr, panel.height / dpr)
          .fill(0xffffff);
        position(this.surfaces[side], panel, dpr);
        if (this.intents[side] === null) {
          const b = this.panels[side].battle,
            rest = centreView(
              grid,
              this.screen.cols,
              this.screen.rows,
              b.cursor,
            );
          this.views[side] = this.screen.view(side, rest.pair, rest.row);
          if (this.paintedDesign)
            this.views[side] = this.screen.slimFit(this.views[side], side, {
              x: 0,
              y: 0,
              width: ww,
              height: hh,
            });
          alignCamera(this.panels[side].map.camera, this.views[side], dpr);
          this.intents[side] = { ...this.panels[side].map.camera.origin };
          this.aligned[side] = { ...this.panels[side].map.camera.origin };
          this.alignedZoom[side] = this.panels[side].map.camera.zoom;
        }
      }
      const run = ++this.frameGeneration;
      void this.frame
        .initialize(this.settings.uiScale, dpr)
        .catch((error: unknown) => {
          if (!this.stopped && run === this.frameGeneration) {
            this.battle.status =
              error instanceof Error ? error.message : String(error);
            this.hooks.changed();
          }
        });
      this.alignViews();
    }
    for (const side of sides) {
      this.wrappers[side].visible = this.visible(side);
      this.surfaces[side].hidden = !this.visible(side);
      this.panels[side].map.dirty = true;
    }
    this.syncPause();
  }
  private alignViews(): void {
    const grid = bind.playable(this.battle.session.board),
      [ww, hh] = hexgrid.bounds(grid);
    let zoomChanged = false;
    for (const side of sides) {
      const camera = this.panels[side].map.camera,
        aligned = this.aligned[side],
        intent = this.intents[side];
      if (aligned === null || intent === null) continue;
      intent.x += camera.origin.x - aligned.x;
      intent.y += camera.origin.y - aligned.y;
      if (Math.abs(camera.zoom - this.alignedZoom[side]) > 1e-7) {
        this.zoomIntent *= camera.zoom / this.alignedZoom[side];
        zoomChanged = true;
      }
    }
    if (zoomChanged) {
      let index = defaultZoom,
        distance = Infinity;
      for (const [i, factor] of zooms.entries()) {
        const d = Math.abs(Math.log(factor / this.zoomIntent));
        if (d < distance) {
          index = i;
          distance = d;
        }
      }
      const base = layout(
        Math.round(this.width * this.dpr),
        Math.round(this.height * this.dpr),
        this.settings.uiScale * this.dpr,
      );
      this.zoom = base.clampZoom(index);
      if (this.paintedDesign)
        this.zoom = Math.max(this.zoom, base.fillZoom(ww, hh));
      this.screen = base.zoomed(this.zoom);
      if (this.zoom !== index) this.zoomIntent = zooms[this.zoom] ?? 1;
    }
    for (const side of sides) {
      const camera = this.panels[side].map.camera,
        intent = this.intents[side],
        old = { ...camera.origin },
        oldZoom = camera.zoom;
      if (intent !== null) camera.origin = { ...intent };
      const rest = cameraRest(camera, this.screen, side, grid);
      let view = this.screen.view(side, rest.pair, rest.row);
      if (this.paintedDesign)
        view = this.screen.slimFit(view, side, {
          x: 0,
          y: 0,
          width: ww,
          height: hh,
        });
      this.views[side] = view;
      alignCamera(camera, view, this.dpr);
      if (intent !== null) {
        const [x0, x1] = pairRange(grid, this.screen.cols),
          [y0, y1] = rowRange(grid, this.screen.rows);
        if (
          (rest.pair === x0 && intent.x < camera.origin.x) ||
          (rest.pair === x1 && intent.x > camera.origin.x)
        )
          intent.x = camera.origin.x;
        if (
          (rest.row === y0 && intent.y < camera.origin.y) ||
          (rest.row === y1 && intent.y > camera.origin.y)
        )
          intent.y = camera.origin.y;
      }
      this.aligned[side] = { ...camera.origin };
      this.alignedZoom[side] = camera.zoom;
      if (
        old.x !== camera.origin.x ||
        old.y !== camera.origin.y ||
        oldZoom !== camera.zoom
      )
        this.panels[side].map.dirty = true;
    }
    this.frame.paint(
      this.screen,
      this.views,
      grid,
      this.paintedDesign,
      this.dpr,
    );
  }
  tick(ms: number): void {
    if (this.paused || this.stopped) return;
    const now = performance.now();
    for (const m of this.messages)
      if (m.fresh && now - m.at >= messageMs) {
        m.fresh = false;
        this.hooks.changed();
      }
    if (this.original) this.alignViews();
    for (const side of sides) {
      this.wrappers[side].visible = this.visible(side);
      this.surfaces[side].hidden = !this.visible(side);
      this.panels[side].tick(ms, this.visible(side));
    }
  }
  /**
   * prepare draws the visible maps behind the loading page without running the battle, so that their images decode
   * before the battle shows; it reports the share of the first frame's images decoded, 1 once complete.
   */
  prepare(): number {
    if (this.stopped) return 0;
    if (this.original) this.alignViews();
    let settled = 0,
      total = 0;
    for (const side of sides) {
      const visible = this.visible(side);
      this.wrappers[side].visible = visible;
      this.surfaces[side].hidden = !visible;
      if (!visible) continue;
      const map = this.panels[side].map;
      map.render();
      const p = map.images.cache.progress();
      settled += p.settled;
      total += p.total;
    }
    return this.ready ? 1 : total === 0 ? 0 : Math.min(settled / total, 0.99);
  }
  /** human reports whether side is commanded by the player now (its well shows the unit line and messages). */
  human(side: 0 | 1): boolean {
    return this.local?.mode === "simultaneous" || side === this.activeSide;
  }
  /** well is side's status well now (cmd/historyline/well.go wellOf). */
  well(side: 0 | 1): Well {
    const b = this.panels[side].battle,
      s = b.session,
      human = this.human(side),
      own = this.panels[this.request.side].battle,
      end = play.ended(s),
      m = this.messages[side];
    if (m !== undefined && b.status !== m.text) {
      m.text = b.status;
      m.at = performance.now();
      m.fresh = b.status !== "";
    }
    const prompter = [0, 1].find(
        (i) =>
          this.panels[i as 0 | 1].battle.phasePrompt ||
          this.panels[i as 0 | 1].battle.surrenderPrompt,
      ),
      prompt =
        prompter === undefined
          ? null
          : prompter === side
            ? b.surrenderPrompt
              ? "F2: GIVE UP THE BATTLE"
              : this.local?.mode === "simultaneous" && side === 1
                ? "CTRL+LEFT: END THE PHASE"
                : "F1: END THE PHASE"
            : this.human(side)
              ? null
              : "FIRE: BACK TO THE ORDERS";
    let unit = null;
    if (human) {
      const at = b.hover ?? b.cursor;
      let u = b.view().units.get(level.coordKey(at));
      if (u?.head !== null && u?.head !== undefined)
        u = b.view().units.get(level.coordKey(u.head));
      if (u !== undefined && u.side === side && b.playback.moving() === null)
        unit = { size: u.size, exp: u.exp, name: u.name, owner: u.owner };
    }
    const planning =
      this.local === null &&
      side !== this.request.side &&
      own.planningOrders() &&
      own.progress !== null
        ? own.progress.total > 0
          ? own.progress.done / own.progress.total
          : 0
        : null;
    return wellOf({
      stage: prompt,
      ended: end === null || !end.over ? null : endText(end, side),
      planning,
      attack: play.phase(s, side) !== phaseMovement,
      busy: b.playback.busy(),
      // The desktop's computer plans as soon as a phase opens and then waits on the player; here it plans once the
      // player confirms, so its well says from the start that it waits.
      ready:
        play.ready(s, side) ||
        (this.local === null && side !== this.request.side),
      enemyAttacks: play.phase(s, 1 - side) !== phaseMovement,
      human,
      message: m?.fresh === true ? b.status : "",
      unit,
      orders: play.stats(s).orders[side],
      limit: s.game.limit,
    });
  }
  refreshImages(): void {
    for (const panel of this.panels) panel.map.dirty = true;
  }
  private installKeyboard(): SplitInput {
    return new SplitInput({
      mode: () => this.local?.mode ?? "single",
      activeSide: () => this.activeSide,
      blocked: () => this.paused || this.stopped,
      handover: () => this.handover !== null,
      keyDown: (side, key, repeat) => this.panels[side].keyDown(key, repeat),
      keyUp: (side, key) => this.panels[side].keyUp(key),
      cameraKey: (side, key, down) => this.panels[side].cameraKey(key, down),
      reset: () => {
        for (const panel of this.panels) panel.cancelInput();
      },
      leave: () => {
        this.hooks.leave();
      },
    });
  }
  dispose(): void {
    if (this.stopped) return;
    this.stopped = true;
    this.frameGeneration++;
    this.keyboard.dispose();
    this.local?.dispose();
    for (const panel of this.panels) panel.dispose();
    for (const wrapper of this.wrappers) wrapper.destroy({ children: true });
    for (const surface of this.surfaces) surface.remove();
    this.frame.dispose();
    this.frame.canvas.remove();
  }
  private hitTest(side: 0 | 1, point: Point): boolean {
    if (!this.original) return true;
    const view = this.views[side],
      p = { x: point.x + view.panel.x, y: point.y + view.panel.y },
      w = view.screenToWorld(p.x, p.y),
      grid = bind.playable(this.battle.session.board),
      [coord, onMap] = hexgrid.pixelToHex(grid, w.x, w.y);
    if (!onMap) return false;
    if (this.paintedDesign) {
      const o = this.screen.slimOpening(side),
        c = view.worldToScreen(...hexgrid.center(coord));
      return (
        p.x >= o.x &&
        p.y >= o.y &&
        p.x < o.x + o.width &&
        p.y < o.y + o.height &&
        c.x >= o.x &&
        c.y >= o.y &&
        c.x < o.x + o.width &&
        c.y < o.y + o.height
      );
    }
    return view.shows(grid, coord);
  }
}

const sideName = (side: number) => (side === 0 ? "Central Powers" : "Entente");
/** endText is the end of the battle in side's words (dos.go endText). */
function endText(
  e: { winner: number; won: [string, string]; score: number },
  side: number,
): string {
  if (e.winner !== 0 && e.winner !== 1) return "THE BATTLE ENDS IN A DRAW";
  const how =
      e.won[e.winner] === "hq"
        ? "hold the enemy HQ"
        : e.won[e.winner] === "surrender"
          ? "accepted the surrender"
          : "destroyed every enemy unit",
    reason = `${sideName(e.winner)} ${how}`;
  return (
    e.winner === side
      ? `VICTORY - ${reason} (score ${e.score})`
      : `DEFEAT - ${reason}`
  ).toUpperCase();
}
