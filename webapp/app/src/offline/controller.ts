import { signal } from "@preact/signals";
import type { FileRef } from "../assetManifest.ts";
import type { OfflineMission } from "../offline.ts";
import { exclusiveFiles, keptBytes, type KeptRecord } from "./kept.ts";

/** A month's files and the record that lists them once they all arrived. */
export interface MonthPlan {
  record: KeptRecord;
  files: FileRef[];
}
export interface StorageFigures {
  used: number | null;
  quota: number | null;
  persistent: boolean | null;
}
export interface OfflineServices {
  records(): Promise<KeptRecord[]>;
  plan(mission: OfflineMission, signal: AbortSignal): Promise<MonthPlan>;
  /** Fetches every file or throws; onFile reports each one that arrived. */
  fetch(
    files: readonly FileRef[],
    signal: AbortSignal,
    onFile: (ref: FileRef) => void,
  ): Promise<void>;
  commit(record: KeptRecord): Promise<void>;
  /** Forgets the month and deletes the files no other kept month needs. */
  remove(key: string): Promise<void>;
  /** Deletes files from the content cache without a record. */
  discard(hashes: readonly string[]): Promise<void>;
  storage(): Promise<StorageFigures>;
  /** Asks the browser not to evict the content cache. */
  persist(): Promise<void>;
}
export interface OfflineState {
  phase: "idle" | "checking" | "available" | "downloading" | "ready" | "error";
  key: string;
  /** Bytes still to fetch for the month. */
  bytes: number;
  /** Bytes fetched so far while downloading. */
  done: number;
  note: string;
}
export interface KeptView {
  key: string;
  label: string;
  mode: string;
  /** What removing the month frees: its files no other kept month needs. */
  bytes: number;
}

const idle: OfflineState = {
  phase: "idle",
  key: "",
  bytes: 0,
  done: 0,
  note: "",
};
const keyOf = (m: OfflineMission) => `${m.month}:${m.mode}`;

/**
 * The Offline maps page's model. One selected map at a time; a selection or cancel invalidates the work in
 * flight, so a stale result never updates the state. A download is recorded only when every file arrived:
 * cancelling or failing deletes the files that no kept month holds.
 */
export class OfflineController {
  readonly state = signal<OfflineState>(idle);
  readonly kept = signal<KeptView[]>([]);
  readonly storage = signal<StorageFigures>({
    used: null,
    quota: null,
    persistent: null,
  });
  /** The bytes of all kept files, shared art counted once. */
  readonly keptTotal = signal(0);
  private records: KeptRecord[] = [];
  private mission: OfflineMission | null = null;
  private pending: MonthPlan | null = null;
  private toFetch: FileRef[] = [];
  private active: AbortController | null = null;
  private generation = 0;
  constructor(private readonly services: OfflineServices) {}

  get busy(): boolean {
    return (
      this.state.value.phase === "checking" ||
      this.state.value.phase === "downloading"
    );
  }
  private current(run: number): boolean {
    return run === this.generation;
  }
  private set(patch: Partial<OfflineState>): void {
    this.state.value = { ...this.state.value, ...patch };
  }
  private invalidate(): AbortController {
    this.generation++;
    this.active?.abort();
    this.active = new AbortController();
    return this.active;
  }

  /** Reads the kept months and the storage figures. */
  async refresh(): Promise<void> {
    const [records, storage] = await Promise.all([
      this.services.records(),
      this.services.storage(),
    ]);
    this.records = records;
    this.kept.value = records.map((r) => ({
      key: r.key,
      label: r.label,
      mode: r.mode,
      bytes: exclusiveFiles(records, r.key).bytes,
    }));
    this.keptTotal.value = keptBytes(records);
    this.storage.value = storage;
  }

  /** Selects the map the page offers: Downloaded when it is kept, else how much it needs. */
  async select(mission: OfflineMission): Promise<void> {
    const key = keyOf(mission);
    if (this.mission !== null && keyOf(this.mission) === key && this.busy)
      return;
    const { signal } = this.invalidate();
    const run = this.generation;
    this.mission = mission;
    this.pending = null;
    this.toFetch = [];
    this.state.value = { ...idle, phase: "checking", key };
    try {
      await this.refresh();
      if (!this.current(run)) return;
      if (this.records.some((r) => r.key === key)) {
        this.set({ phase: "ready" });
        return;
      }
      const plan = await this.services.plan(mission, signal);
      if (!this.current(run)) return;
      const held = new Set(this.records.flatMap((r) => Object.keys(r.files)));
      this.pending = plan;
      this.toFetch = plan.files.filter((f) => !held.has(f.sha256));
      this.set({
        phase: "available",
        bytes: this.toFetch.reduce((sum, f) => sum + f.bytes, 0),
      });
    } catch (error) {
      if (this.current(run)) this.fail(error);
    }
  }
  private fail(error: unknown): void {
    this.set({
      phase: "error",
      note: error instanceof Error ? error.message : String(error),
    });
  }

  async download(): Promise<void> {
    const plan = this.pending,
      files = this.toFetch;
    if (plan === null || this.busy) return;
    const { signal } = this.invalidate();
    const run = this.generation;
    this.set({ phase: "downloading", done: 0, note: "" });
    // The prompt can only be answered inside the player's gesture.
    const persisting = this.services.persist().catch(() => undefined);
    let done = 0;
    try {
      await this.services.fetch(files, signal, (ref) => {
        done += ref.bytes;
        if (this.current(run)) this.set({ done });
      });
      signal.throwIfAborted();
      await this.services.commit(plan.record);
      await persisting;
      if (!this.current(run)) return;
      await this.refresh();
      this.pending = null;
      this.set({ phase: "ready" });
    } catch (error) {
      await this.services
        .discard(files.map((f) => f.sha256))
        .catch(() => undefined);
      if (this.current(run)) {
        this.fail(error);
      }
    }
  }

  /** Stops the download; what arrived without a record is deleted. */
  cancel(): void {
    if (this.state.value.phase !== "downloading") return;
    this.active?.abort();
    this.generation++;
    this.set({ phase: "available", done: 0, note: "" });
    void this.services
      .discard(this.toFetch.map((f) => f.sha256))
      .then(() => this.refresh())
      .catch(() => undefined);
  }

  async remove(key: string): Promise<void> {
    await this.services.remove(key);
    await this.after(key);
  }
  async removeAll(): Promise<void> {
    for (const r of this.records) await this.services.remove(r.key);
    await this.after(null);
  }
  /** The selected map is not kept any more after its removal: offer it again. */
  private async after(removed: string | null): Promise<void> {
    await this.refresh();
    const mission = this.mission;
    if (
      mission !== null &&
      (removed === null || removed === keyOf(mission)) &&
      !this.busy
    )
      await this.select(mission);
  }
}
