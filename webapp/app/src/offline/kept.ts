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
  /**
   * The player downloaded the month on the Offline maps page, and only a Remove there deletes it. Absent (records from
   * before the flag) or false: the background keep stored it, and it leaves with the campaign's window.
   */
  pinned?: boolean;
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
    r["files"] !== null &&
    (r["pinned"] === undefined || typeof r["pinned"] === "boolean")
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
      const record = await this.get(key);
      if (record !== undefined) result.push(record);
    }
    return result.sort((a, b) => a.key.localeCompare(b.key));
  }
  private async get(key: string): Promise<KeptRecord | undefined> {
    try {
      const text = await this.storage?.get(key);
      const value: unknown = text === undefined ? undefined : JSON.parse(text);
      return isRecord(value) ? value : undefined;
    } catch {
      /* A damaged record is no kept month. */
      return undefined;
    }
  }
  async put(record: KeptRecord): Promise<void> {
    await this.storage?.put(record.key, JSON.stringify(record));
  }
  /** Records a month the background keep stored; a month the player pinned stays pinned. */
  async putBackground(record: KeptRecord): Promise<void> {
    const pinned = (await this.get(record.key))?.pinned === true;
    await this.put({ ...record, pinned });
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
  /**
   * Removes every month the player did not pin that is not within the campaign's window, one after the other,
   * so a file goes with the last of them that needed it and stays while a remaining month does. Returns their
   * keys.
   */
  async evict(
    within: ReadonlySet<string>,
    forget: (hashes: readonly string[]) => Promise<void>,
  ): Promise<string[]> {
    const leaving = (await this.list())
      .filter((r) => r.pinned !== true && !within.has(r.key))
      .map((r) => r.key);
    for (const key of leaving) await this.remove(key, forget);
    return leaving;
  }
}
