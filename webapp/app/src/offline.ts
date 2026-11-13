import {
  AssetStorageFullError,
  isStorageFull,
  type AssetStore,
} from "./assetFetch.ts";
import { unitDensity, type AssetLoader } from "./assets.ts";
import {
  fileRef,
  parseImageCatalog,
  type FileRef,
  type Pyramid,
} from "./assetManifest.ts";
import type { KeptRecord, KeptRecords } from "./offline/kept.ts";

export interface OfflineMission {
  month: string;
  mode: string;
  next: string | null;
  news?: { label: string } | null;
}
const keptMonths = 3;
const concurrency = 2;

/** The played mission and the next ones of its campaign chain. */
export function offlineMissions<T extends OfflineMission>(
  mission: T,
  byKey: ReadonlyMap<string, T>,
): T[] {
  const result: T[] = [];
  const seen = new Set<string>();
  let current: T | undefined = mission;
  while (current !== undefined && result.length < keptMonths) {
    const key = `${current.month}:${current.mode}`;
    if (seen.has(key)) break;
    seen.add(key);
    result.push(current);
    current = current.next === null ? undefined : byKey.get(current.next);
  }
  return result;
}

/** Baked layers nest image references and terrain IDs at several depths. */
export function bakedRefs(
  baked: unknown,
  terrain: ReadonlyMap<string, FileRef>,
): FileRef[] {
  const refs = new Map<string, FileRef>();
  const walk = (value: unknown): void => {
    if (typeof value === "string") {
      const ref = terrain.get(value);
      if (ref !== undefined) refs.set(ref.sha256, ref);
    } else if (Array.isArray(value)) {
      for (const item of value) walk(item);
    } else if (typeof value === "object" && value !== null) {
      const entries = Object.entries(value);
      for (const [key, item] of entries) {
        if (key === "lite") {
          const ref = fileRef(item);
          refs.set(ref.sha256, ref);
        } else if (key !== "full") walk(item);
      }
    }
  };
  walk(baked);
  return [...refs.values()];
}
export function backdropRefs(backdrop: Pyramid | undefined): FileRef[] {
  return (backdrop?.mips ?? []).flatMap((mip) =>
    mip.tiles.map((tile) => tile.file),
  );
}

async function catalog(
  loader: AssetLoader,
  name: string,
  density: number,
): Promise<Map<string, FileRef>> {
  const pack = loader.index.packs[name];
  if (pack === undefined) return new Map();
  // A catalog without this density falls back to its largest, as the map does.
  const usable = Object.keys(pack)
    .map(Number)
    .sort((a, b) => a - b);
  const chosen = usable.find((d) => d >= density) ?? usable.at(-1);
  const pair = chosen === undefined ? undefined : pack[String(chosen)];
  if (pair === undefined || chosen === undefined) return new Map();
  const parsed = parseImageCatalog(await loader.store.json(pair.lite), chosen);
  return new Map(parsed.entries.map((entry) => [entry.id, entry.file]));
}

/** The content cache as the offline keep uses it. */
export type OfflineStore = Pick<AssetStore, "keep" | "forget">;

/**
 * Fetches the files into the content cache, two at a time. A failed file counts as a failure; with strict set
 * it stops the rest and throws, otherwise the others go on. A full storage always stops the rest and throws
 * (isStorageFull). Returns the number of failures.
 */
export async function fetchRefs(
  store: Pick<AssetStore, "keep">,
  refs: readonly FileRef[],
  signal: AbortSignal,
  options: { strict?: boolean; onFile?: (ref: FileRef) => void } = {},
): Promise<number> {
  let cursor = 0,
    failures = 0;
  const stop: { error?: Error } = {};
  await Promise.all(
    Array.from({ length: concurrency }, async () => {
      while (
        cursor < refs.length &&
        !signal.aborted &&
        stop.error === undefined
      ) {
        const ref = refs[cursor++];
        if (ref === undefined) return;
        try {
          await store.keep(ref, signal);
          options.onFile?.(ref);
        } catch (error: unknown) {
          failures++;
          if (isStorageFull(error))
            stop.error =
              error instanceof AssetStorageFullError
                ? error
                : new AssetStorageFullError(ref.path, error);
          else if (options.strict === true)
            stop.error = error instanceof Error ? error : Error(String(error));
        }
      }
    }),
  );
  signal.throwIfAborted();
  if (stop.error !== undefined) throw stop.error;
  return failures;
}

function unique(refs: readonly FileRef[]): FileRef[] {
  return [...new Map(refs.map((r) => [r.sha256, r])).values()];
}

export interface OfflinePlan {
  /** Units, interface and screens: the art every month draws. */
  shared: FileRef[];
  /** What each mission needs beyond the shared art. */
  months: { mission: OfflineMission; refs: FileRef[] }[];
}

/** What playing the missions offline needs: the lite art at the density the map draws, the paintings at theirs. */
export async function planOffline(
  loader: AssetLoader,
  missions: readonly OfflineMission[],
  density: number,
  signal: AbortSignal,
  paintingDensity = density,
): Promise<OfflinePlan> {
  const shared = await Promise.all(
    ["units", "interface", "screens"].map((name) =>
      catalog(loader, name, name === "units" ? unitDensity(density) : density),
    ),
  );
  const refs = shared.flatMap((entries) => [...entries.values()]);
  // The first battle draws the lowest density until its full art may come.
  for (const name of ["units", "interface", "screens"])
    refs.push(...(await catalog(loader, name, 0)).values());
  if (paintingDensity !== density)
    refs.push(...(await catalog(loader, "screens", paintingDensity)).values());
  // The menus draw their pages from the screens at their own density, and the split screen its pixel fonts.
  for (const font of Object.values(loader.index.pixelFonts))
    refs.push(font.image, font.metrics);
  const terrain = await catalog(loader, "terrain", density);
  for (const entry of (await catalog(loader, "terrain", 0)).values())
    refs.push(entry);
  const months: OfflinePlan["months"] = [];
  for (const mission of missions) {
    signal.throwIfAborted();
    const baked = await loader.baked(mission.month, mission.mode);
    months.push({
      mission,
      refs: unique([
        ...backdropRefs(await loader.backdrop(mission.month, signal)),
        ...bakedRefs(baked, terrain),
      ]),
    });
  }
  return { shared: unique(refs), months };
}

/** The record of a kept month: every file it needs, shared art included. */
export function keptRecord(
  mission: OfflineMission,
  plan: OfflinePlan,
  refs: readonly FileRef[],
  now = Date.now(),
): KeptRecord {
  const files: Record<string, number> = {};
  for (const ref of [...plan.shared, ...refs]) files[ref.sha256] = ref.bytes;
  return {
    key: `${mission.month}:${mission.mode}`,
    month: mission.month,
    mode: mission.mode,
    label: missionLabel(mission),
    keptAt: now,
    files,
  };
}
export function missionLabel(mission: OfflineMission): string {
  return mission.news?.label ?? mission.month;
}

/** The Network Information API, which not every browser has and TypeScript's DOM types leave out. */
interface ConnectionSource {
  connection?: { saveData?: boolean };
}
/** Reports whether the browser asks to save data: the background keep then waits for the player's say. */
export function saveDataRequested(
  source: ConnectionSource | undefined = typeof navigator === "undefined"
    ? undefined
    : (navigator as ConnectionSource),
): boolean {
  return source?.connection?.saveData === true;
}

/** What a background keep did: the months it recorded and evicted, and whether the storage ran full. */
export interface KeepResult {
  kept: string[];
  evicted: string[];
  full: boolean;
}

/**
 * Stores the lite art of the given missions in the content cache, so they can be played offline; the
 * screens also at paintingDensity, the density the menus draw them at. See keepPlanned.
 */
export async function keepForOffline(
  loader: AssetLoader,
  missions: readonly OfflineMission[],
  density: number,
  signal: AbortSignal,
  paintingDensity = density,
  kept?: KeptRecords,
): Promise<KeepResult> {
  const plan = await planOffline(
    loader,
    missions,
    density,
    signal,
    paintingDensity,
  );
  return keepPlanned(loader.store, plan, signal, kept);
}

/**
 * Fetches the plan's files and records in kept each month that arrived complete, which is what the Offline maps
 * page lists; a month the player pinned there stays pinned. A full storage stops the keep quietly. Then every
 * month the player did not pin that left the window, the plan's months, is evicted with the files no
 * remaining month needs. A keep cancelled with its battle evicts nothing.
 */
export async function keepPlanned(
  store: OfflineStore,
  plan: OfflinePlan,
  signal: AbortSignal,
  kept?: KeptRecords,
): Promise<KeepResult> {
  const result: KeepResult = { kept: [], evicted: [], full: false };
  try {
    const sharedFailures = await fetchRefs(store, plan.shared, signal);
    for (const { mission, refs } of plan.months) {
      if (signal.aborted) return result;
      const failures = await fetchRefs(store, refs, signal);
      if (failures + sharedFailures === 0 && kept !== undefined) {
        const record = keptRecord(mission, plan, refs);
        await kept.putBackground(record);
        result.kept.push(record.key);
      }
    }
  } catch (error) {
    if (!isStorageFull(error)) throw error;
    result.full = true;
  }
  if (kept === undefined || signal.aborted) return result;
  const within = new Set(
    plan.months.map(({ mission }) => `${mission.month}:${mission.mode}`),
  );
  result.evicted = await kept.evict(within, (hashes) => store.forget(hashes));
  return result;
}

/** The shell's words for a battle that could not load. */
export function battleLoadError(error: unknown, online: boolean): Error {
  if (!online)
    return Error(
      "This month is not available offline yet. Connect to the internet to play it.",
    );
  return error instanceof Error ? error : Error(String(error));
}
