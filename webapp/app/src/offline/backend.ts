import type { AssetLoader } from "../assets.ts";
import { fetchRefs, keptRecord, planOffline } from "../offline.ts";
import type { OfflineServices, StorageFigures } from "./controller.ts";
import { browserRecordStorage, KeptRecords } from "./kept.ts";

interface StorageApi {
  estimate?: () => Promise<{ usage?: number; quota?: number }>;
  persist?: () => Promise<boolean>;
  persisted?: () => Promise<boolean>;
}

export function browserRecords(): KeptRecords {
  return new KeptRecords(browserRecordStorage());
}

/** The Offline maps page's services on the running game's loader and the browser's storage. */
export function browserServices(
  loader: () => AssetLoader | null,
  records: KeptRecords,
  densities: () => { map: number; paintings: number },
): OfflineServices {
  const active = (): AssetLoader => {
    const l = loader();
    if (l === null) throw new Error("The game is not ready.");
    return l;
  };
  const storage = (): StorageApi | undefined =>
    typeof navigator === "undefined" ? undefined : navigator.storage;
  return {
    records: () => records.list(),
    async plan(mission, signal) {
      const { map, paintings } = densities();
      const plan = await planOffline(
        active(),
        [mission],
        map,
        signal,
        paintings,
      );
      const month = plan.months[0];
      if (month === undefined) throw new Error("This map has no files.");
      return {
        record: keptRecord(mission, plan, month.refs),
        files: [...plan.shared, ...month.refs],
      };
    },
    async fetch(files, signal, onFile) {
      await fetchRefs(active(), files, signal, { strict: true, onFile });
    },
    commit: (record) => records.put(record),
    async remove(key) {
      await records.remove(key, (hashes) => active().store.forget(hashes));
    },
    discard: (hashes) => active().store.forget(hashes),
    async storage(): Promise<StorageFigures> {
      const api = storage();
      const [estimate, persistent] = await Promise.all([
        api?.estimate?.().catch(() => undefined),
        api?.persisted?.().catch(() => undefined),
      ]);
      return {
        used: estimate?.usage ?? null,
        quota: estimate?.quota ?? null,
        persistent: persistent ?? null,
      };
    },
    async persist() {
      await storage()?.persist?.();
    },
  };
}
