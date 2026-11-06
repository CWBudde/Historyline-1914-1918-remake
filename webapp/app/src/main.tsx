import {
  bind,
  scenario,
  play,
  mustTables,
  profilesOf,
  phaseMovement,
  hexgrid,
  level,
} from "@historyline/engine";
import { signal } from "@preact/signals";
import { Application } from "pixi.js";
import { render } from "preact";
import { Computer } from "./computer.ts";
import { AssetLoader } from "./assets.ts";
import { record } from "./assetManifest.ts";
import { MapBattle } from "./mapBattle.ts";
import { MapRenderer } from "./mapRenderer.ts";
import { MapInput } from "./mapInput.ts";
import { parseBakedLayers } from "./mapScene.ts";
import { CandidateArtError } from "./mapImages.ts";
import type { SpriteFamily } from "./unitFrame.ts";
import type { Point } from "./camera.ts";
import { BattleUI } from "./ui/BattleUI.tsx";
import { UiAssets } from "./ui/assets.ts";
import { layoutRing, placement } from "./ui/ring.ts";
import { BattleKeys } from "./input/keys.ts";
import { BattleScenes } from "./battle/scene.ts";

// Its lifetime follows the document, independently of battle/menu changes.
export const computer = new Computer();
const status = signal("Loading battle…"),
  revision = signal(0),
  candidate = signal(false),
  loading = signal(true),
  scale = signal(1),
  reduced = signal(
    window.matchMedia("(prefers-reduced-motion: reduce)").matches,
  ),
  scenes = signal(true),
  painted = signal(true),
  ring = signal<ReturnType<typeof layoutRing> | null>(null),
  leavePrompt = signal(false);
let battle: MapBattle | null = null,
  map: MapRenderer | null = null,
  input: MapInput | null = null,
  auxiliary: BattleScenes | null = null,
  uiAssets: UiAssets | null = null,
  families: Record<string, SpriteFamily> = {},
  ringAt: hexgrid.Coord | null = null,
  keys: BattleKeys | null = null,
  allowCandidateArt = false,
  generation = 0;
let motionAbort = new AbortController();

function closeRing(): void {
  ring.value = null;
  ringAt = null;
}
function design(): void {
  map?.design();
  painted.value = map?.backdrop ?? true;
}
function setReduced(value: boolean): void {
  reduced.value = value;
  battle?.playback.setReducedMotion(value);
  if (map !== null) map.dirty = true;
  revision.value++;
}
function setScenes(value: boolean): void {
  scenes.value = value;
  battle?.playback.setScenes(value);
  if (map !== null) map.dirty = true;
  revision.value++;
}
function requestLeave(): void {
  closeRing();
  keys?.cancel();
  leavePrompt.value = true;
}
function disposeBattle(): void {
  generation++;
  motionAbort.abort();
  input?.dispose();
  input = null;
  keys?.cancel();
  keys = null;
  if (battle !== null) battle.onChange = null;
  map?.dispose();
  map = null;
  auxiliary?.dispose();
  auxiliary = null;
  uiAssets?.dispose();
  uiAssets = null;
  battle = null;
  closeRing();
  leavePrompt.value = false;
  document.documentElement.classList.remove("in-battle");
  if (import.meta.env.DEV) delete window.__historyline;
  revision.value++;
}
function Controls() {
  const tick = revision.value;
  const current = battle;
  return current !== null && uiAssets !== null ? (
    <>
      <div inert={leavePrompt.value}>
        <BattleUI
          battle={current}
          assets={uiAssets}
          families={families}
          revision={revision}
          scale={scale}
          reduced={reduced}
          scenes={scenes}
          painted={painted}
          ring={ring}
          onScale={(value) => {
            scale.value = value;
            document.documentElement.style.setProperty(
              "--ui-scale",
              String(value),
            );
            closeRing();
          }}
          onReduced={setReduced}
          onScenes={setScenes}
          onDesign={design}
          onLeave={requestLeave}
          onFullscreen={() => {
            void fullscreen();
          }}
          onRingAction={(action) => {
            const at = ringAt ?? current.cursor;
            closeRing();
            current.run(action, at);
          }}
          onCloseRing={closeRing}
          onSceneHost={(host) => auxiliary?.attach("battle", host)}
          onStrategyHost={(host) => auxiliary?.attach("strategy", host)}
        />
      </div>
      {leavePrompt.value && (
        <div class="leave-shade">
          <section
            role="alertdialog"
            aria-modal="true"
            aria-labelledby="leave-title"
            class="banner"
          >
            <h2 id="leave-title">Leave this battle?</h2>
            <p>The current orders will be lost.</p>
            <div class="map-actions">
              <button
                onClick={() => {
                  disposeBattle();
                  status.value = "Battle closed.";
                }}
              >
                Leave battle
              </button>
              <button
                autoFocus
                onClick={() => {
                  leavePrompt.value = false;
                }}
              >
                Back to battle
              </button>
            </div>
          </section>
        </div>
      )}
    </>
  ) : (
    <div class="banner" data-revision={tick}>
      <h1>Historyline 1914-1918</h1>
      <p role="status">{status.value}</p>
      {!loading.value && (
        <button
          onClick={() => {
            void loadBattle(candidate.value || allowCandidateArt);
          }}
        >
          {candidate.value ? "Use candidate art" : "Start battle"}
        </button>
      )}
    </div>
  );
}
async function fullscreen(): Promise<void> {
  try {
    if (document.fullscreenElement !== null) await document.exitFullscreen();
    else await document.documentElement.requestFullscreen();
  } catch (error) {
    if (battle !== null) {
      battle.status = error instanceof Error ? error.message : String(error);
      revision.value++;
    }
  }
}

let app: Application;
async function loadBattle(allowCandidates = false): Promise<void> {
  disposeBattle();
  const run = generation;
  loading.value = true;
  status.value = "Loading battle…";
  motionAbort = new AbortController();
  let next: MapRenderer | null = null,
    nextScenes: BattleScenes | null = null,
    nextAssets: UiAssets | null = null;
  try {
    const loader = await AssetLoader.open(
      new URL("./index.json", document.baseURI),
      { allowCandidates },
    );
    const month = "1914-08",
      mode = "cp";
    const [files, baked, terrain, units, records, full, demo, spriteData] =
      await Promise.all([
        loader.levelFiles(month),
        loader.baked(month, mode),
        loader.rule("terrain.json"),
        loader.rule("units.json"),
        loader.rule("records.json"),
        loader.rule("full.json"),
        loader.rule("demo.json"),
        loader.rule("sprites.json"),
      ]);
    const scenarioFiles = new Map(
      [...files].filter(([name]) => name !== "news.json"),
    );
    const file = scenario.decode(
      scenarioFiles,
      new Set(Object.keys(record(terrain))),
    );
    const session = play.newSession(
      scenario.level(file, mode),
      bind.bindingData(units, records),
      {
        profile: profilesOf(mustTables(full), mustTables(demo)).full,
        computer: [false, true],
      },
    );
    const nextFamilies: Record<string, SpriteFamily> = {};
    for (const [id, value] of Object.entries(record(spriteData))) {
      const meta = record(value);
      if (
        typeof meta.family !== "string" ||
        (meta.long !== undefined && typeof meta.long !== "string")
      )
        throw new Error(`invalid sprite family ${id}`);
      nextFamilies[id] = {
        family: meta.family,
        ...(typeof meta.long === "string" ? { long: meta.long } : {}),
      };
    }
    const current = new MapBattle(session, 0, computer);
    current.playback.setReducedMotion(reduced.value);
    current.playback.setScenes(scenes.value);
    next = new MapRenderer(
      app,
      loader,
      parseBakedLayers(baked),
      current,
      nextFamilies,
      loader.index.levels[month]?.backdrop,
      (message) => {
        current.status = message;
        revision.value++;
      },
    );
    const renderer = next;
    nextAssets = new UiAssets(loader);
    nextScenes = new BattleScenes(
      current,
      loader,
      nextFamilies,
      () => scale.value,
      () => renderer.camera.visible(),
      (at) => {
        const [x, y] = hexgrid.center(at);
        renderer.camera.centerOn(x, y);
        renderer.dirty = true;
        current.setCursor(at);
      },
      (error) => {
        current.status = error.message;
        revision.value++;
      },
    );
    await Promise.all([
      next.initialize(),
      nextScenes.initialize(),
      nextAssets.fonts(loader),
    ]);
    if (run !== generation) {
      next.dispose();
      nextScenes.dispose();
      nextAssets.dispose();
      return;
    }
    map = next;
    auxiliary = nextScenes;
    uiAssets = nextAssets;
    families = nextFamilies;
    battle = current;
    painted.value = renderer.backdrop;
    allowCandidateArt = allowCandidates;
    candidate.value = false;
    status.value = "";
    document.documentElement.classList.add("in-battle");
    current.onChange = () => {
      renderer.dirty = true;
      auxiliary?.refresh();
      revision.value++;
    };
    const at = (p: Point) => {
      const world = renderer.camera.screenToWorld(p.x, p.y);
      const [coord, onMap] = hexgrid.pixelToHex(
        level.grid(session.level),
        world.x,
        world.y,
      );
      return onMap ? coord : null;
    };
    const seeCursor = () => {
      const [x, y] = hexgrid.center(current.cursor),
        v = renderer.camera.visible();
      if (
        x < v.x + 48 ||
        x > v.x + v.width - 48 ||
        y < v.y + 40 ||
        y > v.y + v.height - 40
      )
        renderer.camera.centerOn(x, y);
      renderer.dirty = true;
    };
    const close = () => {
      if (leavePrompt.value) leavePrompt.value = false;
      else if (ring.value !== null) closeRing();
      else current.cancel();
    };
    keys = new BattleKeys({
      step: (direction) => {
        if (leavePrompt.value) return;
        closeRing();
        current.stepCursor(direction);
        if (current.contents.box === null) seeCursor();
      },
      click: () => {
        if (leavePrompt.value) return;
        if (ring.value !== null) {
          closeRing();
          return;
        }
        current.fire();
      },
      run: (action) => {
        if (!leavePrompt.value) current.run(action);
      },
      placement: () => {
        const allowed = new Set(current.actions());
        const p = placement(
          play.phase(session, current.side) !== phaseMovement,
          allowed.has("open"),
          current.contents.box !== null,
        );
        return Object.fromEntries(
          Object.entries(p).map(([direction, action]) => [
            direction,
            allowed.has(action) ? action : "cancel",
          ]),
        );
      },
      shortcut: (key) => {
        const k = key.toLowerCase();
        if (k === "escape") {
          close();
          return true;
        }
        if (leavePrompt.value) return true;
        switch (k) {
          case "f1":
            closeRing();
            void current.requestPhase();
            break;
          case "f2":
            closeRing();
            current.requestSurrender();
            break;
          case "f8":
            design();
            break;
          case "f9":
            setScenes(!scenes.value);
            break;
          case "f11":
            void fullscreen();
            break;
          case "m":
            setReduced(!reduced.value);
            break;
          case "q":
            requestLeave();
            break;
          case "home":
          case "pageup":
            renderer.camera.centerOn(renderer.camera.world.width / 2, 0);
            renderer.dirty = true;
            break;
          case "end":
          case "pagedown":
            renderer.camera.centerOn(
              renderer.camera.world.width / 2,
              renderer.camera.world.height,
            );
            renderer.dirty = true;
            break;
          default:
            return false;
        }
        return true;
      },
    });
    input = new MapInput(app.canvas, renderer.camera, {
      changed: () => {
        closeRing();
        renderer.dirty = true;
      },
      hover: (p) => {
        if (ring.value === null && !leavePrompt.value) current.setCursor(at(p));
      },
      click: (p) => {
        if (leavePrompt.value) return;
        if (ring.value !== null) {
          closeRing();
          return;
        }
        const c = at(p);
        if (c !== null) current.click(c);
      },
      cancel: close,
      design,
      phase: () => {
        void current.requestPhase();
      },
      context: (p) => {
        if (leavePrompt.value) return;
        if (current.selected !== null || current.playback.busy()) {
          current.cancel();
          closeRing();
          return;
        }
        const c = at(p);
        if (c === null) return;
        current.setCursor(c);
        const actions = current.actions(c);
        if (actions.length === 0) return;
        const r = app.canvas.getBoundingClientRect();
        ringAt = c;
        ring.value = layoutRing(
          actions,
          {
            x: r.left + p.x / renderer.camera.scale,
            y: r.top + p.y / renderer.camera.scale,
          },
          r,
          scale.value,
        );
      },
      contextCancel: closeRing,
      keyDown: (key, repeat) => keys?.down(key, repeat) ?? false,
      keyUp: (key) => keys?.up(key) ?? false,
      reset: () => {
        keys?.cancel();
        closeRing();
      },
    });
    const motion = window.matchMedia("(prefers-reduced-motion: reduce)");
    motion.addEventListener(
      "change",
      (e) => {
        setReduced(e.matches);
      },
      { signal: motionAbort.signal },
    );
    if (import.meta.env.DEV)
      window.__historyline = { battle: current, map: renderer };
  } catch (error) {
    next?.dispose();
    nextScenes?.dispose();
    nextAssets?.dispose();
    status.value = error instanceof Error ? error.message : String(error);
    candidate.value = error instanceof CandidateArtError;
  } finally {
    if (run === generation) {
      loading.value = false;
      revision.value++;
    }
  }
}

async function start(): Promise<void> {
  const stage = document.getElementById("stage"),
    ui = document.getElementById("ui");
  if (stage === null || ui === null) throw new Error("missing #stage or #ui");
  app = new Application();
  await app.init({
    resolution: window.devicePixelRatio,
    autoDensity: true,
    antialias: true,
    background: 0x162630,
    width: stage.clientWidth,
    height: stage.clientHeight,
  });
  stage.appendChild(app.canvas);
  render(<Controls />, ui);
  const resize = () => {
    const dpr = window.devicePixelRatio,
      width = stage.clientWidth,
      height = stage.clientHeight;
    if (app.renderer.resolution !== dpr) app.renderer.resolution = dpr;
    if (app.screen.width !== width || app.screen.height !== height)
      app.renderer.resize(width, height);
    if (
      map !== null &&
      (map.camera.scale !== dpr ||
        map.camera.view.width !== width * dpr ||
        map.camera.view.height !== height * dpr)
    )
      map.resize(width, height, dpr);
  };
  window.addEventListener("resize", resize);
  app.ticker.add((ticker) => {
    resize();
    input?.tick(ticker.deltaMS);
    battle?.tick(ticker.deltaMS);
    map?.render();
  });
  window.addEventListener(
    "pagehide",
    () => {
      disposeBattle();
      computer.dispose();
      app.destroy(true);
    },
    { once: true },
  );
  await loadBattle();
}
declare global {
  interface Window {
    __historyline?: { battle: MapBattle; map: MapRenderer };
  }
}
void start().catch((error: unknown) => {
  status.value = error instanceof Error ? error.message : String(error);
});
