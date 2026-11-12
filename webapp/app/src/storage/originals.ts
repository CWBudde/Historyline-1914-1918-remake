/** The original game's files the player imported, kept in this browser's IndexedDB beside the saves (its own
 * database, so the saves' schema stays as it is). They never leave the browser. */
export interface StoredFile {
  path: string;
  bytes: ArrayBuffer;
}

const storeName = "files";

function failure(value: unknown): Error {
  return value instanceof Error ? value : new Error(String(value));
}

export class OriginalStore {
  private constructor(private readonly db: IDBDatabase) {
    db.onversionchange = () => {
      db.close();
    };
  }

  static open(
    factory?: IDBFactory,
    name = "historyline-original",
  ): Promise<OriginalStore> {
    const available =
      factory ?? (globalThis as { indexedDB?: IDBFactory }).indexedDB;
    if (!available)
      return Promise.reject(new Error("IndexedDB is unavailable"));
    return new Promise((resolve, reject) => {
      const request = available.open(name, 1);
      request.onupgradeneeded = () => {
        request.result.createObjectStore(storeName, { keyPath: "path" });
      };
      request.onblocked = () => {
        reject(
          new Error("IndexedDB upgrade is blocked; close other tabs and retry"),
        );
      };
      request.onerror = () => {
        reject(request.error ?? new Error("IndexedDB could not be opened"));
      };
      request.onsuccess = () => {
        const db = request.result;
        if (!db.objectStoreNames.contains(storeName)) {
          db.close();
          reject(new Error("Invalid IndexedDB storage schema"));
          return;
        }
        resolve(new OriginalStore(db));
      };
    });
  }

  private run<T>(
    mode: IDBTransactionMode,
    work: (store: IDBObjectStore, done: (value: T) => void) => void,
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      let tx: IDBTransaction;
      try {
        tx = this.db.transaction(storeName, mode);
      } catch (error) {
        reject(failure(error));
        return;
      }
      let value: T | undefined,
        has = false;
      tx.oncomplete = () => {
        if (has) resolve(value as T);
        else
          reject(new Error("IndexedDB transaction completed without a result"));
      };
      tx.onabort = () => {
        reject(tx.error ?? new Error("IndexedDB transaction aborted"));
      };
      try {
        work(tx.objectStore(storeName), (v) => {
          value = v;
          has = true;
        });
      } catch (error) {
        tx.abort();
        reject(failure(error));
      }
    });
  }

  /** files are the stored files by path; empty when nothing is imported. */
  files(): Promise<Map<string, Uint8Array>> {
    return this.run("readonly", (store, done) => {
      const request = store.getAll();
      request.onsuccess = () => {
        const out = new Map<string, Uint8Array>();
        for (const v of request.result as unknown[]) {
          const f = v as Partial<StoredFile>;
          if (typeof f.path === "string" && f.bytes instanceof ArrayBuffer)
            out.set(f.path, new Uint8Array(f.bytes));
        }
        done(out);
      };
    });
  }

  /** replace stores exactly these files, in one transaction. */
  replace(files: ReadonlyMap<string, Uint8Array>): Promise<void> {
    return this.run("readwrite", (store, done) => {
      store.clear();
      for (const [path, bytes] of files)
        store.put({ path, bytes: bytes.slice().buffer } satisfies StoredFile);
      done(undefined);
    });
  }

  clear(): Promise<void> {
    return this.run("readwrite", (store, done) => {
      store.clear();
      done(undefined);
    });
  }

  close(): void {
    this.db.close();
  }
}
