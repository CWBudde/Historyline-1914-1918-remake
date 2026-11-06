import { play } from "@historyline/engine";
import { signal } from "@preact/signals";
import { Application } from "pixi.js";
import { render } from "preact";
import { Computer } from "./computer.ts";
import { AssetLoader } from "./assets.ts";
import { parseImageCatalog } from "./assetManifest.ts";
import { UiAssets } from "./ui/assets.ts";
import { SplitRuntime } from "./split/runtime.ts";
import { SplitUI } from "./split/SplitUI.tsx";
import {
  loadBattleData,
  newBattleSession,
  type BattleData,
} from "./battle/session.ts";
import { loadCampaign, type Campaign, type Mission } from "./shell/campaign.ts";
import { ShellController, type BattleStart } from "./shell/controller.ts";
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
  lastEnd: ReturnType<typeof play.ended> = null;

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
  newBattleSession(
    mission.file,
    mission.mode,
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
  );
  if (signal.aborted || current !== shell || closed) {
    next.dispose();
    signal.throwIfAborted();
    throw Error("Battle loading was cancelled.");
  }
  next.applyPreferences(current.settings);
  runtime = next;
  lastEnd = null;
  publishDebug();
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
    const nextLoader = await AssetLoader.open(
      new URL("./index.json", document.baseURI),
    );
    await Promise.all(
      ["screens", "units", "terrain"].flatMap((name) => {
        const pair = nextLoader.index.packs[name]?.["1"];
        if (pair === undefined) throw Error(`Missing ${name} art catalog.`);
        return [pair.lite, pair.full].map(async (ref) =>
          parseImageCatalog(await nextLoader.store.json(ref), 1),
        );
      }),
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
    });
    shell = model;
    model.onChange = changed;
    applyPreferences(model.settings);
    if (storageNote !== "") model.note = storageNote;
    changed();
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
async function start(): Promise<void> {
  const stage = document.getElementById("stage"),
    ui = document.getElementById("ui");
  if (stage === null || ui === null) throw Error("Missing game stage.");
  app = new Application();
  render(<Controls />, ui);
  window.addEventListener(
    "pagehide",
    () => {
      closed = true;
      bootAbort.abort();
      if (shell !== null) shell.onChange = null;
      discardBattle();
      assets?.dispose();
      store?.close();
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
  app.ticker.add((ticker) => {
    resize();
    runtime?.tick(ticker.deltaMS);
    checkEnd();
  });
  window.addEventListener("resize", () => {
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
