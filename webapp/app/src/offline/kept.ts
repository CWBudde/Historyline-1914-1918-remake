/** What each kept month holds in the content cache, so removing one deletes only what no other month needs. */
export interface KeptRecord {
  /** month:mode */
  key: string;
  month: string;
  mode: string;
  label: string;
  keptAt: number;
  /** Every file the month needs, shared art included: SHA-256 to bytes. */
  files: Record<string, number>;
}

/** Where the records live: a small text store, the Cache API in the browser. */
export interface RecordStorage {
  keys(): Promise<string[]>;
  get(key: string): Promise<string | undefined>;
  put(key: string, text: string): Promise<void>;
  delete(key: string): Promise<void>;
}

const recordsCache = "historyline-offline-v1";
const recordsPath = "/__historyline_offline__/";

export function browserRecordStorage(): RecordStorage | null {
  if (typeof caches === "undefined") return null;
  const open = caches.open(recordsCache);
  const url = (key: string) =>
    new URL(recordsPath + encodeURIComponent(key), location.origin).href;
  return {
    keys: async () =>
      (await (await open).keys()).map((r) =>
        decodeURIComponent(new URL(r.url).pathname.slice(recordsPath.length)),
      ),
    get: async (key) => (await (await open).match(url(key)))?.text(),
    put: async (key, text) => {
      await (await open).put(url(key), new Response(text));
    },
    delete: async (key) => {
      await (await open).delete(url(key));
    },
  };
}

function isRecord(value: unknown): value is KeptRecord {
  if (typeof value !== "object" || value === null) return false;
  const r = value as Record<string, unknown>;
  return (
    typeof r["key"] === "string" &&
    typeof r["month"] === "string" &&
    typeof r["mode"] === "string" &&
    typeof r["label"] === "string" &&
    typeof r["keptAt"] === "number" &&
    typeof r["files"] === "object" &&
    r["files"] !== null
  );
}

/** The files only this record needs: what removing it may delete. */
export function exclusiveFiles(
  records: readonly KeptRecord[],
  key: string,
): { hashes: string[]; bytes: number } {
  const others = new Set<string>();
  for (const r of records)
    if (r.key !== key)
      for (const hash of Object.keys(r.files)) others.add(hash);
  const mine = records.find((r) => r.key === key);
  const hashes = Object.keys(mine?.files ?? {}).filter((h) => !others.has(h));
  return {
    hashes,
    bytes: hashes.reduce((sum, h) => sum + (mine?.files[h] ?? 0), 0),
  };
}

/** The bytes of all files any record holds, each counted once. */
export function keptBytes(records: readonly KeptRecord[]): number {
  const all = new Map<string, number>();
  for (const r of records)
    for (const [hash, bytes] of Object.entries(r.files)) all.set(hash, bytes);
  return [...all.values()].reduce((a, b) => a + b, 0);
}

export class KeptRecords {
  constructor(private readonly storage: RecordStorage | null) {}
  async list(): Promise<KeptRecord[]> {
    if (this.storage === null) return [];
    const result: KeptRecord[] = [];
    for (const key of await this.storage.keys()) {
      try {
        const text = await this.storage.get(key);
        const value: unknown =
          text === undefined ? undefined : JSON.parse(text);
        if (isRecord(value)) result.push(value);
      } catch {
        /* A damaged record is no kept month. */
      }
    }
    return result.sort((a, b) => a.key.localeCompare(b.key));
  }
  async put(record: KeptRecord): Promise<void> {
    await this.storage?.put(record.key, JSON.stringify(record));
  }
  /**
   * Forgets the month and deletes, through forget, exactly the files no other kept month needs. The record
   * goes first: an interrupted removal leaves unlisted files, never a listed month with missing files.
   */
  async remove(
    key: string,
    forget: (hashes: readonly string[]) => Promise<void>,
  ): Promise<number> {
    const records = await this.list();
    const { hashes, bytes } = exclusiveFiles(records, key);
    await this.storage?.delete(key);
    await forget(hashes);
    return bytes;
  }
}
