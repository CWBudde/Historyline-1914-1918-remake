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

// Its lifetime follows the document, independently of battle/menu changes.
export const computer = new Computer();
const status = signal("Loading battle…"),
  revision = signal(0),
  candidate = signal(false),
  loading = signal(true);
let battle: MapBattle | null = null,
  map: MapRenderer | null = null,
  input: MapInput | null = null;
let motionAbort = new AbortController();

function Controls() {
  const tick = revision.value;
  const current = battle,
    renderer = map;
  return (
    <div class="banner" data-revision={tick}>
      <h1>Historyline 1914-1918</h1>
      {current === null ? (
        <p role="status">{status.value}</p>
      ) : (
        <>
          <p>
            {current.session.level.mapName} · Turn {play.turn(current.session)}{" "}
            ·{" "}
            {play.phase(current.session, current.side) === phaseMovement
              ? "Movement"
              : "Action"}
          </p>
          <p role="status">{status.value || current.status}</p>
          <div class="map-actions">
            <button
              disabled={current.busy() || play.ended(current.session) !== null}
              onClick={() => {
                void current.phase().then(() => {
                  revision.value++;
                });
                revision.value++;
              }}
            >
              Ready (F1)
            </button>
            <button
              onClick={() => {
                renderer?.design();
                revision.value++;
              }}
            >
              {renderer?.backdrop ? "Show tiles" : "Show painting"} (F8)
            </button>
          </div>
          <p class="map-help">
            Drag or WASD to scroll · Wheel or pinch to zoom · Tap a unit, then
            its destination or target
          </p>
        </>
      )}
      {candidate.value && (
        <button
          disabled={loading.value}
          onClick={() => {
            void loadBattle(true);
          }}
        >
          Use candidate art
        </button>
      )}
    </div>
  );
}

let app: Application;
async function loadBattle(allowCandidates = false): Promise<void> {
  loading.value = true;
  status.value = "Loading battle…";
  motionAbort.abort();
  motionAbort = new AbortController();
  input?.dispose();
  input = null;
  map?.dispose();
  map = null;
  battle = null;
  revision.value++;
  let next: MapRenderer | null = null;
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
    const families: Record<string, SpriteFamily> = {};
    for (const [id, value] of Object.entries(record(spriteData))) {
      const meta = record(value);
      if (
        typeof meta.family !== "string" ||
        (meta.long !== undefined && typeof meta.long !== "string")
      )
        throw new Error(`invalid sprite family ${id}`);
      families[id] = {
        family: meta.family,
        ...(typeof meta.long === "string" ? { long: meta.long } : {}),
      };
    }
    const current = new MapBattle(session, 0, computer);
    const motion = window.matchMedia("(prefers-reduced-motion: reduce)");
    current.playback.setReducedMotion(motion.matches);
    next = new MapRenderer(
      app,
      loader,
      parseBakedLayers(baked),
      current,
      families,
      loader.index.levels[month]?.backdrop,
      (message) => {
        status.value = message;
      },
    );
    await next.initialize();
    map = next;
    battle = current;
    candidate.value = false;
    status.value = "";
    const renderer = next;
    const at = (p: { x: number; y: number }) => {
      const world = renderer.camera.screenToWorld(p.x, p.y);
      const [coord, onMap] = hexgrid.pixelToHex(
        level.grid(session.level),
        world.x,
        world.y,
      );
      return onMap ? coord : null;
    };
    input = new MapInput(app.canvas, renderer.camera, {
      changed: () => {
        renderer.dirty = true;
      },
      hover: (p) => {
        const c = at(p);
        if (JSON.stringify(c) !== JSON.stringify(current.hover)) {
          current.hover = c;
          renderer.dirty = true;
        }
      },
      click: (p) => {
        const c = at(p);
        if (c !== null) {
          current.click(c);
          revision.value++;
        }
      },
      cancel: () => {
        current.cancel();
        revision.value++;
      },
      design: () => {
        renderer.design();
        revision.value++;
      },
      phase: () => {
        void current.phase().then(() => {
          revision.value++;
        });
        revision.value++;
      },
    });
    motion.addEventListener(
      "change",
      (e) => {
        current.playback.setReducedMotion(e.matches);
        renderer.dirty = true;
      },
      { signal: motionAbort.signal },
    );
    if (import.meta.env.DEV)
      window.__historyline = { battle: current, map: renderer };
  } catch (error) {
    next?.dispose();
    status.value = error instanceof Error ? error.message : String(error);
    candidate.value = error instanceof CandidateArtError;
  } finally {
    loading.value = false;
    revision.value++;
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
    width: window.innerWidth,
    height: window.innerHeight,
  });
  stage.appendChild(app.canvas);
  render(<Controls />, ui);
  const resize = () => {
    const dpr = window.devicePixelRatio;
    if (app.renderer.resolution !== dpr) app.renderer.resolution = dpr;
    if (
      app.screen.width !== window.innerWidth ||
      app.screen.height !== window.innerHeight
    )
      app.renderer.resize(window.innerWidth, window.innerHeight);
    if (
      map !== null &&
      (map.camera.scale !== dpr ||
        map.camera.view.width !== window.innerWidth * dpr ||
        map.camera.view.height !== window.innerHeight * dpr)
    )
      map.resize(window.innerWidth, window.innerHeight, dpr);
  };
  window.addEventListener("resize", resize);
  app.ticker.add((ticker) => {
    resize();
    input?.tick(ticker.deltaMS);
    const busy = battle?.busy();
    battle?.tick(ticker.deltaMS);
    map?.render();
    if (battle?.busy() !== busy) revision.value++;
  });
  window.addEventListener(
    "pagehide",
    () => {
      input?.dispose();
      motionAbort.abort();
      map?.dispose();
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
