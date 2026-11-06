/** Browser persistence and the native internal/saves.Envelope file format. */
export interface Settings {
  uiScale: number;
  reduced: boolean;
  scenes: boolean;
  painted: boolean;
  scroll: "normal" | "fast" | "slow";
  opponent: "computer" | "passive";
  layout: "responsive" | "original";
  playMode: "single" | "simultaneous" | "hotseat";
  limit: 0 | 5 | 10 | 15 | 20;
  allowCandidates: boolean;
}

export interface SaveEnvelope {
  map: string;
  side: 0 | 1;
  data: string;
}

export interface SaveSlot {
  slot: string;
  name: string;
  savedAt: number;
  envelope: SaveEnvelope;
}

export interface ScoreEntry {
  level: string;
  score: number;
}

// Large enough for a complete game, bounded before JSON/base64 allocations.
const maxSaveBytes = 8 * 1024 * 1024;
const maxBase64Length = 4 * Math.ceil(maxSaveBytes / 3);
const maxEnvelopeLength = maxBase64Length + 2048;

function object(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function label(value: unknown, field: string, max = 128): string {
  if (
    typeof value !== "string" ||
    value.length > max ||
    value.trim() === "" ||
    Array.from(value).some(
      (char) => char.charCodeAt(0) < 32 || char.charCodeAt(0) === 127,
    )
  )
    throw new Error(`Invalid ${field}`);
  return value;
}

export function defaultSettings(systemReduced = false): Settings {
  return {
    uiScale: 1,
    reduced: systemReduced,
    scenes: true,
    painted: true,
    scroll: "normal",
    opponent: "computer",
    layout: "responsive",
    playMode: "single",
    limit: 0,
    allowCandidates: false,
  };
}

/** Missing fields receive defaults; malformed present fields throw, never coerce. */
export function validateSettings(value: unknown): Settings {
  if (!object(value)) throw new Error("Invalid settings");
  const result = defaultSettings();
  for (const key of [
    "reduced",
    "scenes",
    "painted",
    "allowCandidates",
  ] as const) {
    if (!Object.hasOwn(value, key)) continue;
    const field = value[key];
    if (typeof field !== "boolean") throw new Error(`Invalid settings ${key}`);
    result[key] = field;
  }
  if (Object.hasOwn(value, "uiScale")) {
    const scale = value.uiScale;
    if (typeof scale !== "number" || ![0.75, 1, 1.25, 1.5, 2].includes(scale))
      throw new Error("Invalid settings uiScale");
    result.uiScale = scale;
  }
  if (Object.hasOwn(value, "scroll")) {
    const scroll = value.scroll;
    if (scroll !== "normal" && scroll !== "fast" && scroll !== "slow")
      throw new Error("Invalid settings scroll");
    result.scroll = scroll;
  }
  if (Object.hasOwn(value, "opponent")) {
    const opponent = value.opponent;
    if (opponent !== "computer" && opponent !== "passive")
      throw new Error("Invalid settings opponent");
    result.opponent = opponent;
  }
  if (Object.hasOwn(value, "layout")) {
    const layout = value.layout;
    if (layout !== "responsive" && layout !== "original")
      throw new Error("Invalid settings layout");
    result.layout = layout;
  }
  if (Object.hasOwn(value, "playMode")) {
    const playMode = value.playMode;
    if (
      playMode !== "single" &&
      playMode !== "simultaneous" &&
      playMode !== "hotseat"
    )
      throw new Error("Invalid settings playMode");
    result.playMode = playMode;
  }
  if (Object.hasOwn(value, "limit")) {
    const limit = value.limit;
    if (
      limit !== 0 &&
      limit !== 5 &&
      limit !== 10 &&
      limit !== 15 &&
      limit !== 20
    )
      throw new Error("Invalid settings limit");
    result.limit = limit;
  }
  return result;
}

function validatePayload(bytes: Uint8Array): void {
  if (bytes.length === 0 || bytes.length > maxSaveBytes)
    throw new Error("Invalid saved game size");
  const value: unknown = JSON.parse(
    new TextDecoder("utf-8", { fatal: true }).decode(bytes),
  );
  if (!object(value)) throw new Error("Saved game must be a JSON object");
}

function base64(bytes: Uint8Array): string {
  const chunks: string[] = [];
  for (let i = 0; i < bytes.length; i += 0x8000)
    chunks.push(String.fromCharCode(...bytes.subarray(i, i + 0x8000)));
  return btoa(chunks.join(""));
}

function envelope(value: unknown): SaveEnvelope {
  if (!object(value)) throw new Error("Invalid save envelope");
  const map = label(value.map, "save map");
  if (value.side !== 0 && value.side !== 1)
    throw new Error("Invalid save side");
  const data = value.data;
  if (
    typeof data !== "string" ||
    data.length === 0 ||
    data.length > maxBase64Length ||
    data.length % 4 !== 0 ||
    /[^A-Za-z0-9+/=]/.test(data)
  )
    throw new Error("Invalid save base64 data");
  return { map, side: value.side, data };
}

export function encodeEnvelope(
  map: string,
  side: 0 | 1,
  bytes: Uint8Array,
): SaveEnvelope {
  validatePayload(bytes);
  return envelope({ map, side, data: base64(bytes) });
}

/** Structural checks only: the caller must load the game against the chosen mission. */
export function decodeEnvelope(value: SaveEnvelope): Uint8Array {
  const checked = envelope(value);
  let binary: string;
  try {
    binary = atob(checked.data);
  } catch {
    throw new Error("Invalid save base64 data");
  }
  if (binary.length > maxSaveBytes || btoa(binary) !== checked.data)
    throw new Error("Invalid save base64 data");
  const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0));
  validatePayload(bytes);
  return bytes;
}

export function serializeSave(value: SaveEnvelope): string {
  const checked = envelope(value);
  decodeEnvelope(checked);
  return JSON.stringify(checked, null, 2).replace(
    /[<>&\u2028\u2029]/g,
    (char) => `\\u${char.charCodeAt(0).toString(16).padStart(4, "0")}`,
  );
}

export function parseSave(text: string): SaveEnvelope {
  if (text.length > maxEnvelopeLength)
    throw new Error("Save file is too large");
  const checked = envelope(JSON.parse(text));
  decodeEnvelope(checked);
  return checked;
}

function saveSlot(value: unknown): SaveSlot {
  if (!object(value)) throw new Error("Invalid save slot");
  const slot = label(value.slot, "save slot");
  const name = label(value.name, "save name", 120);
  const savedAt = value.savedAt;
  if (
    typeof savedAt !== "number" ||
    !Number.isSafeInteger(savedAt) ||
    savedAt < 0
  )
    throw new Error("Invalid save date");
  const checked = envelope(value.envelope);
  decodeEnvelope(checked);
  return { slot, name, savedAt, envelope: checked };
}

function scoreEntry(value: unknown): ScoreEntry {
  if (!object(value)) throw new Error("Invalid score entry");
  const level = label(value.level, "score level");
  if (typeof value.score !== "number" || !Number.isSafeInteger(value.score))
    throw new Error("Invalid score value");
  return { level, score: value.score };
}

type RequestHandler = <T>(
  request: IDBRequest<T>,
  read: (value: unknown) => void,
) => void;

function storageError(value: unknown): Error {
  return value instanceof Error ? value : new Error(String(value));
}

/** All writes resolve on transaction completion; failed or stale storage always rejects. */
export class BrowserStore {
  private unavailable: Error | undefined;

  private constructor(private readonly db: IDBDatabase) {
    db.onversionchange = () => {
      this.unavailable = new Error(
        "IndexedDB version changed; reopen browser storage",
      );
      db.close();
    };
    db.onclose = () => {
      this.unavailable ??= new Error(
        "IndexedDB connection closed unexpectedly",
      );
    };
  }

  static async open(
    factory?: IDBFactory,
    name = "historyline",
  ): Promise<BrowserStore> {
    // lib.dom declares indexedDB as always present; it can be absent or disabled at runtime.
    const available =
      factory ?? (globalThis as { indexedDB?: IDBFactory }).indexedDB;
    if (!available) throw new Error("IndexedDB is unavailable");
    label(name, "database name");
    return new Promise((resolve, reject) => {
      const request = available.open(name, 1);
      let refused = false;
      request.onblocked = () => {
        refused = true;
        reject(
          new Error("IndexedDB upgrade is blocked; close other tabs and retry"),
        );
      };
      request.onupgradeneeded = () => {
        if (refused) {
          request.transaction?.abort();
          return;
        }
        const db = request.result;
        try {
          db.createObjectStore("saves", { keyPath: "slot" });
          db.createObjectStore("prefs", { keyPath: "key" });
          db.createObjectStore("scores", { keyPath: "level" });
        } catch (error) {
          refused = true;
          request.transaction?.abort();
          reject(storageError(error));
        }
      };
      request.onerror = () => {
        refused = true;
        reject(request.error ?? new Error("IndexedDB could not be opened"));
      };
      request.onsuccess = () => {
        const db = request.result;
        if (refused) {
          db.close();
          return;
        }
        // Refuse an unrelated or damaged database rather than losing a later write.
        if (
          ["saves", "prefs", "scores"].some(
            (store) => !db.objectStoreNames.contains(store),
          )
        ) {
          db.close();
          reject(new Error("Invalid IndexedDB storage schema"));
          return;
        }
        resolve(new BrowserStore(db));
      };
    });
  }

  private transaction<T>(
    name: string,
    mode: IDBTransactionMode,
    work: (
      store: IDBObjectStore,
      done: (value: T) => void,
      read: RequestHandler,
    ) => void,
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      if (this.unavailable) {
        reject(this.unavailable);
        return;
      }
      let tx: IDBTransaction;
      try {
        tx = this.db.transaction(name, mode);
      } catch (error) {
        reject(storageError(error));
        return;
      }
      let value: T;
      let hasValue = false;
      let failure: Error | undefined;
      const abort = (error: unknown) => {
        failure = storageError(error);
        try {
          tx.abort();
        } catch {
          reject(failure);
        }
      };
      tx.onabort = () => {
        reject(
          failure ?? tx.error ?? new Error("IndexedDB transaction aborted"),
        );
      };
      tx.oncomplete = () => {
        if (hasValue) resolve(value);
        else
          reject(new Error("IndexedDB transaction completed without a result"));
      };
      const read: RequestHandler = (request, callback) => {
        request.onsuccess = () => {
          try {
            const result: unknown = request.result;
            callback(result);
          } catch (error) {
            abort(error);
          }
        };
      };
      try {
        work(
          tx.objectStore(name),
          (result) => {
            value = result;
            hasValue = true;
          },
          read,
        );
      } catch (error) {
        abort(error);
      }
    });
  }

  listSaves(): Promise<SaveSlot[]> {
    return this.transaction("saves", "readonly", (store, done, read) => {
      read(store.getAll(), (value) => {
        if (!Array.isArray(value)) throw new Error("Invalid saves list");
        done(value.map(saveSlot).sort((a, b) => a.slot.localeCompare(b.slot)));
      });
    });
  }

  getSave(slot: string): Promise<SaveSlot | undefined> {
    return this.transaction("saves", "readonly", (store, done, read) => {
      read(store.get(label(slot, "save slot")), (value) => {
        done(value === undefined ? undefined : saveSlot(value));
      });
    });
  }

  putSave(value: SaveSlot): Promise<void> {
    return this.transaction("saves", "readwrite", (store, done) => {
      store.put(saveSlot(value));
      done(undefined);
    });
  }

  deleteSave(slot: string): Promise<void> {
    return this.transaction("saves", "readwrite", (store, done) => {
      store.delete(label(slot, "save slot"));
      done(undefined);
    });
  }

  loadSettings(): Promise<Settings> {
    return this.transaction("prefs", "readonly", (store, done, read) => {
      read(store.get("settings"), (value) => {
        if (value === undefined) {
          const motion = (globalThis as { matchMedia?: typeof matchMedia })
            .matchMedia;
          done(
            defaultSettings(
              motion?.("(prefers-reduced-motion: reduce)").matches,
            ),
          );
          return;
        }
        if (!object(value)) throw new Error("Invalid stored settings");
        done(validateSettings(value.value));
      });
    });
  }

  saveSettings(value: Settings): Promise<void> {
    return this.transaction("prefs", "readwrite", (store, done) => {
      store.put({ key: "settings", value: validateSettings(value) });
      done(undefined);
    });
  }

  listScores(): Promise<ScoreEntry[]> {
    return this.transaction("scores", "readonly", (store, done, read) => {
      read(store.getAll(), (value) => {
        if (!Array.isArray(value)) throw new Error("Invalid score list");
        done(
          value.map(scoreEntry).sort((a, b) => a.level.localeCompare(b.level)),
        );
      });
    });
  }

  recordScore(level: string, score: number): Promise<boolean> {
    return this.transaction("scores", "readwrite", (store, done, read) => {
      const entry = scoreEntry({ level, score });
      read(store.get(entry.level), (old) => {
        if (old !== undefined && scoreEntry(old).score >= entry.score) {
          done(false);
          return;
        }
        store.put(entry);
        done(true);
      });
    });
  }

  close(): void {
    this.unavailable ??= new Error("IndexedDB storage is closed");
    this.db.close();
  }
}
