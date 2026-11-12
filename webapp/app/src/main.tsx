import { play } from "@historyline/engine";
import { signal } from "@preact/signals";
import { Application, UPDATE_PRIORITY } from "pixi.js";
import { render } from "preact";
import { Computer } from "./computer.ts";
import { AssetLoader, assetDensity } from "./assets.ts";
import { battleLoadError, keepForOffline, offlineMissions } from "./offline.ts";
import { FrameMeter } from "./frameMeter.ts";
import { UiAssets } from "./ui/assets.ts";
import { SplitRuntime } from "./split/runtime.ts";
import { SplitUI } from "./split/SplitUI.tsx";
import {
  loadBattleData,
  missionSession,
  type BattleData,
} from "./battle/session.ts";
import {
  loadCampaign,
  setOriginals,
  type Campaign,
  type Mission,
} from "./shell/campaign.ts";
import {
  ShellController,
  type BattleStart,
  type OriginalHooks,
} from "./shell/controller.ts";
import { installViewport } from "./viewport.ts";
import { OriginalStore } from "./storage/originals.ts";
import { select, type Picked } from "./original/files.ts";
import {
  loadOriginalTables,
  originalMissions,
  type OriginalTables,
} from "./original/catalog.ts";
import { setOriginalArt } from "./original/state.ts";
import { ShellUI } from "./shell/ShellUI.tsx";
import {
  BrowserStore,
  defaultSettings,
  encodeEnvelope,
  decodeEnvelope,
  type Settings,
  type SaveEnvelope,
} from "./storage/store.ts";

// Document resources survive menu changes and individual battles.
export const computer = new Computer();
const revision = signal(0),
  status = signal("Preparing Historyline…"),
  bootBusy = signal(true);
let app: Application,
  appReady = false,
  closed = false,
  bootGeneration = 0;
let shell: ShellController | null = null,
  assets: UiAssets | null = null,
  runtime: SplitRuntime | null = null,
  store: BrowserStore | null = null,
  loader: AssetLoader | null = null,
  data: BattleData | null = null,
  campaign: Campaign | null = null;
let preferences = defaultSettings(
    window.matchMedia("(prefers-reduced-motion: reduce)").matches,
  ),
  storageNote = "";
let storageAttempt: Promise<void> | null = null,
  bootAbort = new AbortController(),
  lastEnd: ReturnType<typeof play.ended> = null,
  offlineKeep: { timer: number; abort: AbortController } | null = null,
  battleStarting = false;

const offlineDelay = 15_000;
/** Full art is held for the first battle; a player who stays in the menus gets it after this. */
const menuUpgradeDelay = 8_000;
function stopOfflineKeep(): void {
  if (offlineKeep === null) return;
  clearTimeout(offlineKeep.timer);
  offlineKeep.abort.abort();
  offlineKeep = null;
}
/** After the battle has settled, as its first load must stay small. */
function scheduleOfflineKeep(battle: SplitRuntime, mission: Mission): void {
  stopOfflineKeep();
  const active = loader,
    missions = campaign?.byKey;
  const connection = (navigator as { connection?: { saveData?: boolean } })
    .connection;
  if (active === null || missions === undefined || connection?.saveData) return;
  const abort = new AbortController();
  const density = assetDensity(battle.map.camera.zoom, battle.map.camera.scale);
  offlineKeep = {
    abort,
    timer: window.setTimeout(() => {
      keepForOffline(
        active,
        offlineMissions(mission, missions),
        density,
        abort.signal,
      ).catch(() => {
        /* Best effort: playing online is unaffected. */
      });
    }, offlineDelay),
  };
}
function message(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}
function discardBattle(): void {
  const old = runtime;
  runtime = null;
  lastEnd = null;
  old?.dispose();
  publishDebug();
}
function publishDebug(): void {
  if (!import.meta.env.DEV) return;
  if (shell === null) {
    delete window.__historyline;
    return;
  }
  window.__historyline = {
    shell,
    ...(runtime === null
      ? {}
      : { battle: runtime.battle, map: runtime.map, runtime }),
  };
}
function resize(): void {
  if (!appReady) return;
  const stage = document.getElementById("stage");
  if (stage === null) return;
  const dpr = window.devicePixelRatio,
    width = Math.max(1, stage.clientWidth),
    height = Math.max(1, stage.clientHeight);
  if (app.renderer.resolution !== dpr) app.renderer.resolution = dpr;
  if (app.screen.width !== width || app.screen.height !== height)
    app.renderer.resize(width, height);
  runtime?.resize(width, height, dpr);
}
function applyPreferences(settings: Settings): void {
  preferences = { ...settings };
  document.documentElement.style.setProperty(
    "--ui-scale",
    String(settings.uiScale),
  );
  runtime?.applyPreferences(settings);
  resize();
}
function checkEnd(): void {
  if (shell?.activeMission === null || shell === null || runtime === null)
    return;
  const end = play.ended(runtime.battle.session);
  if (end !== null && end !== lastEnd) {
    lastEnd = end;
    void shell.endBattle(end);
  }
}
function changed(): void {
  const battlePage = shell?.page === "battle",
    paused = !battlePage;
  document.documentElement.classList.toggle(
    "in-battle",
    runtime !== null && (battlePage || shell?.page === "leave"),
  );
  document.documentElement.classList.toggle(
    "original-layout",
    runtime?.original ?? false,
  );
  runtime?.setPaused(paused);
  resize();
  if (appReady) {
    if (battlePage) app.start();
    else app.stop();
  }
  publishDebug();
  revision.value++;
  checkEnd();
}
async function fullscreen(): Promise<void> {
  try {
    if (document.fullscreenElement !== null) await document.exitFullscreen();
    else await document.documentElement.requestFullscreen();
  } catch (error) {
    if (shell !== null) shell.note = message(error);
    else status.value = message(error);
    revision.value++;
  }
}
function Controls() {
  const tick = revision.value,
    current = shell,
    images = assets,
    battle = runtime;
  if (current === null || images === null)
    return (
      <div class="banner" data-revision={tick}>
        <h1>Historyline 1914-1918</h1>
        <p role="status">{status.value}</p>
        {!bootBusy.value && (
          <button
            onClick={() => {
              void bootstrap();
            }}
          >
            Retry
          </button>
        )}
      </div>
    );
  if (current.page !== "battle" || battle === null)
    return (
      <ShellUI
        shell={current}
        assets={images}
        revision={revision}
        onFullscreen={() => {
          void fullscreen();
        }}
        isFullscreen={document.fullscreenElement !== null}
      />
    );
  return (
    <>
      <SplitUI
        runtime={battle}
        assets={images}
        revision={revision}
        state={{
          busy: current.busy,
          note: current.note,
        }}
      />
      {battle.ended() && current.result !== null && (
        <div class="battle-ended">
          <p role="status">
            {current.result.text} · Score {current.result.score}
          </p>
          <button
            onClick={() => {
              current.showOutcome();
            }}
          >
            Continue
          </button>
        </div>
      )}
    </>
  );
}
// The original game's files the player imported (original/): kept in their own IndexedDB database.
let originalStore: OriginalStore | null = null,
  originalTables: OriginalTables | null = null,
  originalStatus = { maps: 0, art: false, skipped: [] as string[] };
/** offerOriginals makes the imported files' maps playable and their art drawable; it throws on bad files. */
async function offerOriginals(
  files: ReadonlyMap<string, Uint8Array>,
): Promise<void> {
  if (loader === null || data === null || campaign === null)
    throw Error("The game is not ready.");
  if (files.size === 0) {
    setOriginals(campaign, []);
    setOriginalArt(null);
    originalStatus = { maps: 0, art: false, skipped: [] };
    return;
  }
  originalTables ??= await loadOriginalTables(loader);
  const catalog = originalMissions(
    files,
    data,
    originalTables,
    [...campaign.byKey.values()].filter((m) => m.original === undefined),
  );
  const art = setOriginalArt(files);
  setOriginals(campaign, catalog.missions);
  originalStatus = {
    maps: new Set(catalog.missions.map((m) => m.month)).size,
    art,
    skipped: catalog.skipped,
  };
}
const originalHooks: OriginalHooks = {
  status: () => originalStatus,
  async import(picked: readonly Picked[]) {
    const chosen = select(picked);
    if (chosen.problem !== "") throw Error(chosen.problem);
    // Check every map binds before anything is stored.
    await offerOriginals(chosen.files);
    originalStore ??= await OriginalStore.open();
    await originalStore.replace(chosen.files);
    return `${originalStatus.maps} original maps imported${originalStatus.art ? " with their tiles and units" : ""}.${originalStatus.skipped.length > 0 ? ` Left out: ${originalStatus.skipped.join("; ")}.` : ""}`;
  },
  async remove() {
    originalStore ??= await OriginalStore.open();
    await originalStore.clear();
    await offerOriginals(new Map());
  },
};
async function restoreOriginals(): Promise<string> {
  try {
    originalStore ??= await OriginalStore.open();
    const files = await originalStore.files();
    if (files.size > 0) await offerOriginals(files);
    return "";
  } catch (error) {
    return `The imported original files could not be read: ${message(error)}`;
  }
}
async function initializeStorage(): Promise<void> {
  if (storageAttempt !== null) return storageAttempt;
  storageAttempt = (async () => {
    try {
      store = await BrowserStore.open();
      try {
        preferences = await store.loadSettings();
      } catch (error) {
        storageNote = `Saved settings could not be read: ${message(error)}`;
      }
    } catch (error) {
      storageNote = `Browser storage is unavailable: ${message(error)}`;
    }
  })();
  return storageAttempt;
}
function savedMission(envelope: SaveEnvelope): Mission {
  if (campaign === null || data === null)
    throw Error("The campaign is not ready.");
  const mission =
    campaign.byKey.get(envelope.map) ?? campaign.findCode(envelope.map);
  if (mission === null)
    throw Error(`The saved map ${envelope.map} is not in the campaign.`);
  // Parse against that setup before replacing a slot or the active battle.
  missionSession(
    mission,
    envelope.side,
    { limit: 0, computer: false },
    data,
    decodeEnvelope(envelope),
  );
  return mission;
}
async function beginBattle(
  request: BattleStart,
  signal: AbortSignal,
): Promise<void> {
  if (loader === null || data === null || shell === null)
    throw Error("The game is not ready.");
  const current = shell,
    activeLoader = loader,
    activeData = data;
  discardBattle();
  battleStarting = true;
  const next = await SplitRuntime.create(
    app,
    activeLoader,
    request,
    activeData,
    current.settings,
    computer,
    {
      changed,
      leave: () => {
        current.requestLeave();
      },
      continue: () => {
        current.showOutcome();
      },
      fullscreen: () => {
        void fullscreen();
      },
      preference: (patch) => {
        void current.updateSettings(patch);
      },
    },
    signal,
    (fraction, text) => {
      current.loading(fraction, text);
    },
  )
    .catch((error: unknown) => {
      if (signal.aborted) throw error;
      throw battleLoadError(error, navigator.onLine);
    })
    .finally(() => {
      battleStarting = false;
    });
  if (signal.aborted || current !== shell || closed) {
    next.dispose();
    signal.throwIfAborted();
    throw Error("Battle loading was cancelled.");
  }
  next.applyPreferences(current.settings);
  runtime = next;
  lastEnd = null;
  publishDebug();
  // An imported original map is local already; the offline cache keeps the shipped months.
  if (request.mission.original === undefined)
    scheduleOfflineKeep(next, request.mission);
}
async function bootstrap(): Promise<void> {
  const run = ++bootGeneration;
  bootAbort.abort();
  bootAbort = new AbortController();
  const signal = bootAbort.signal;
  bootBusy.value = true;
  status.value = "Loading the campaign…";
  if (shell !== null) shell.onChange = null;
  shell = null;
  discardBattle();
  assets?.dispose();
  assets = null;
  changed();
  let nextAssets: UiAssets | null = null;
  try {
    await initializeStorage();
    signal.throwIfAborted();
    // Full art waits for the first playable battle frame: a phone starts from
    // lite pixels at the lowest density, then upgrades.
    const nextLoader = await AssetLoader.open(
      new URL("./index.json", document.baseURI),
      { deferUpgrades: true },
    );
    signal.throwIfAborted();
    nextAssets = new UiAssets(nextLoader);
    const [nextCampaign, nextData] = await Promise.all([
      loadCampaign(nextLoader),
      loadBattleData(nextLoader),
      nextAssets.fonts(nextLoader),
    ]);
    signal.throwIfAborted();
    if (run !== bootGeneration || closed) {
      nextAssets.dispose();
      return;
    }
    loader = nextLoader;
    data = nextData;
    campaign = nextCampaign;
    assets = nextAssets;
    const originalNote = await restoreOriginals();
    const model = new ShellController(nextCampaign, store, preferences, {
      start: beginBattle,
      pause: (paused) => runtime?.setPaused(paused),
      discard: discardBattle,
      capture: () => {
        if (runtime === null || !runtime.canSave())
          throw Error(
            "The computer is planning. Stay in the battle until its orders are ready.",
          );
        return encodeEnvelope(
          runtime.request.mission.key,
          runtime.saveSide(),
          play.save(runtime.battle.session),
        );
      },
      canSave: () => runtime?.canSave() ?? false,
      validate: savedMission,
      settings: applyPreferences,
      originals: originalHooks,
    });
    shell = model;
    model.onChange = changed;
    applyPreferences(model.settings);
    if (storageNote !== "") model.note = storageNote;
    else if (originalNote !== "") model.note = originalNote;
    changed();
    window.setTimeout(() => {
      if (runtime === null && !battleStarting && nextLoader.enableUpgrades())
        changed();
    }, menuUpgradeDelay);
  } catch (error) {
    nextAssets?.dispose();
    if (run === bootGeneration && !signal.aborted) {
      status.value = message(error);
    }
  } finally {
    if (run === bootGeneration) {
      bootBusy.value = false;
      revision.value++;
    }
  }
}
function registerServiceWorker(): void {
  if (!import.meta.env.PROD || !("serviceWorker" in navigator)) return;
  navigator.serviceWorker
    .register(new URL("./sw.js", document.baseURI))
    .catch(() => {
      /* Without it the game still plays online. */
    });
}
async function start(): Promise<void> {
  registerServiceWorker();
  const stage = document.getElementById("stage"),
    ui = document.getElementById("ui");
  if (stage === null || ui === null) throw Error("Missing game stage.");
  app = new Application();
  render(<Controls />, ui);
  window.addEventListener(
    "pagehide",
    () => {
      closed = true;
      stopOfflineKeep();
      bootAbort.abort();
      if (shell !== null) shell.onChange = null;
      discardBattle();
      assets?.dispose();
      store?.close();
      originalStore?.close();
      computer.dispose();
      if (appReady) app.destroy(true);
    },
    { once: true },
  );
  await app.init({
    resolution: window.devicePixelRatio,
    autoDensity: true,
    antialias: true,
    background: 0x162630,
    width: window.innerWidth,
    height: window.innerHeight,
  });
  if (closed) {
    app.destroy(true);
    return;
  }
  appReady = true;
  app.stop();
  stage.appendChild(app.canvas);
  const frames = new FrameMeter();
  let firstPlayable = false;
  app.ticker.add(
    () => {
      frames.begin();
    },
    undefined,
    UPDATE_PRIORITY.INTERACTION + 1,
  );
  app.ticker.add((ticker) => {
    resize();
    runtime?.tick(ticker.deltaMS);
    checkEnd();
  });
  app.ticker.add(
    () => {
      frames.end();
      // Pixi submits its frame at LOW. Release full art only after the first
      // visible lite map has passed through that render.
      if (runtime?.ready !== true) return;
      if (!firstPlayable) {
        firstPlayable = true;
        performance.mark("historyline.first-playable");
      }
      if (loader?.enableUpgrades()) runtime.refreshImages();
    },
    undefined,
    UPDATE_PRIORITY.LOW - 1,
  );
  window.addEventListener("resize", () => {
    resize();
    revision.value++;
  });
  // Phones: the visual viewport shrinks under browser chrome and on-screen keyboards.
  installViewport(document, window, () => {
    resize();
    revision.value++;
  });
  window.addEventListener("fullscreenchange", () => {
    resize();
    revision.value++;
  });
  window
    .matchMedia("(prefers-reduced-motion: reduce)")
    .addEventListener("change", (event) => {
      if (shell !== null) void shell.updateSettings({ reduced: event.matches });
    });
  await bootstrap();
}
declare global {
  interface Window {
    __historyline?: {
      shell: ShellController;
      battle?: SplitRuntime["battle"];
      map?: SplitRuntime["map"];
      runtime?: SplitRuntime;
    };
  }
}
void start().catch((error: unknown) => {
  status.value = message(error);
  bootBusy.value = false;
  revision.value++;
});
