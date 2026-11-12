export interface RenderResource<T> {
  value: T;
  bytes: number;
  dispose(): void;
}
interface Entry<T> {
  load: () => Promise<RenderResource<T>>;
  resource: RenderResource<T> | null;
  state: "queued" | "loading" | "ready" | "failed";
  used: number;
  error?: (error: Error) => void;
}

/** Network bytes have a separate disk cache. This bounds decoded CPU/GPU images
 * and concurrent decodes; resources used by the current frame stay resident. */
export class RenderCache<T> {
  private entries = new Map<string, Entry<T>>();
  private needed = new Set<string>();
  /** Needed resources a frame can do without: upgrades of images it has. */
  private optional = new Set<string>();
  private active = 0;
  private serial = 0;
  private stopped = false;
  bytes = 0;
  /** ready reports whether every resource of the current frame has settled: decoded, or failed for good. */
  get ready(): boolean {
    if (this.stopped) return false;
    for (const key of this.needed) {
      if (this.optional.has(key)) continue;
      const state = this.entries.get(key)?.state;
      if (state !== "ready" && state !== "failed") return false;
    }
    return true;
  }
  /** progress counts the current frame's settled resources. */
  progress(): { settled: number; total: number } {
    let settled = 0,
      total = 0;
    for (const key of this.needed) {
      if (this.optional.has(key)) continue;
      total++;
      const state = this.entries.get(key)?.state;
      if (state === "ready" || state === "failed") settled++;
    }
    return { settled, total };
  }
  constructor(
    private readonly budget: number,
    public concurrency: number,
    private readonly changed: () => void,
    private readonly error: (error: Error) => void = () => {},
  ) {}
  begin(): void {
    this.needed.clear();
    this.optional.clear();
  }
  peek(key: string): T | null {
    return this.entries.get(key)?.resource?.value ?? null;
  }
  /** get requests a resource for the current frame; an optional one (an upgrade) does not keep it from being complete. */
  get(
    key: string,
    load: () => Promise<RenderResource<T>>,
    error?: (error: Error) => void,
    optional = false,
  ): T | null {
    if (this.stopped) return null;
    if (optional && !this.needed.has(key)) this.optional.add(key);
    else this.optional.delete(key);
    this.needed.add(key);
    let e = this.entries.get(key);
    if (e === undefined) {
      e = {
        load,
        resource: null,
        state: "queued",
        used: ++this.serial,
        ...(error === undefined ? {} : { error }),
      };
      this.entries.set(key, e);
    } else e.used = ++this.serial;
    this.pump();
    return e.resource?.value ?? null;
  }
  /**
   * hold keeps a decoded resource for the current frame as a stand-in for one still decoding, so that it is neither
   * evicted nor replaced by nothing; false once it is gone.
   */
  hold(key: string): boolean {
    const e = this.entries.get(key);
    if (this.stopped || e?.state !== "ready") return false;
    this.needed.add(key);
    e.used = ++this.serial;
    return true;
  }
  end(): void {
    for (const [key, e] of this.entries)
      if (e.state === "queued" && !this.needed.has(key))
        this.entries.delete(key);
    this.evict();
    this.pump();
  }
  dispose(): void {
    this.stopped = true;
    for (const e of this.entries.values()) e.resource?.dispose();
    this.entries.clear();
    this.bytes = 0;
  }
  /** pump starts queued loads, what the frame shows before upgrades and prefetches. */
  private pump(): void {
    if (this.stopped) return;
    for (const later of [false, true])
      for (const [key, e] of this.entries) {
        if (this.active >= this.concurrency) return;
        if (e.state !== "queued" || this.optional.has(key) !== later) continue;
        e.state = "loading";
        this.active++;
        void this.load(key, e);
      }
  }
  private async load(key: string, e: Entry<T>): Promise<void> {
    try {
      const resource = await e.load();
      if (this.stopped || this.entries.get(key) !== e) resource.dispose();
      else {
        e.resource = resource;
        e.state = "ready";
        this.bytes += resource.bytes;
        this.evict();
        this.changed();
      }
    } catch (error) {
      e.state = "failed";
      if (!this.stopped)
        (e.error ?? this.error)(
          error instanceof Error ? error : new Error(String(error)),
        );
    } finally {
      this.active--;
      this.pump();
    }
  }
  private evict(): void {
    if (this.bytes <= this.budget) return;
    const idle = [...this.entries]
      .filter(([key, e]) => !this.needed.has(key) && e.state === "ready")
      .sort((a, b) => a[1].used - b[1].used);
    for (const [key, e] of idle) {
      if (this.bytes <= this.budget) break;
      if (e.resource !== null) {
        this.bytes -= e.resource.bytes;
        e.resource.dispose();
      }
      this.entries.delete(key);
    }
  }
}
