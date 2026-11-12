import type { AssetLoader } from "./assets.ts";
import {
  fileRef,
  parseImageCatalog,
  type FileRef,
  type Pyramid,
} from "./assetManifest.ts";

export interface OfflineMission {
  month: string;
  mode: string;
  next: string | null;
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

/** Best effort: one failed file must not stop the rest, nor reach the player. */
async function fetchAll(
  loader: AssetLoader,
  refs: readonly FileRef[],
  signal: AbortSignal,
): Promise<void> {
  let cursor = 0;
  await Promise.all(
    Array.from({ length: concurrency }, async () => {
      while (cursor < refs.length && !signal.aborted) {
        const ref = refs[cursor++];
        if (ref === undefined) return;
        try {
          await loader.store.get(ref, signal);
        } catch {
          /* The player sees the missing file when they reach it. */
        }
      }
    }),
  );
}

/** Stores the lite art of the given missions in the content cache, so they can be played offline. */
export async function keepForOffline(
  loader: AssetLoader,
  missions: readonly OfflineMission[],
  density: number,
  signal: AbortSignal,
): Promise<void> {
  const shared = await Promise.all(
    ["units", "interface", "screens"].map((name) =>
      catalog(loader, name, density),
    ),
  );
  await fetchAll(
    loader,
    shared.flatMap((entries) => [...entries.values()]),
    signal,
  );
  const terrain = await catalog(loader, "terrain", density);
  for (const mission of missions) {
    if (signal.aborted) return;
    const baked = await loader.baked(mission.month, mission.mode);
    await fetchAll(
      loader,
      [
        ...backdropRefs(await loader.backdrop(mission.month, signal)),
        ...bakedRefs(baked, terrain),
      ],
      signal,
    );
  }
}

/** The shell's words for a battle that could not load. */
export function battleLoadError(error: unknown, online: boolean): Error {
  if (!online)
    return Error(
      "This month is not available offline yet. Connect to the internet to play it.",
    );
  return error instanceof Error ? error : Error(String(error));
}
