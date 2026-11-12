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
  private active = 0;
  private serial = 0;
  private stopped = false;
  bytes = 0;
  get ready(): boolean {
    if (this.stopped) return false;
    for (const key of this.needed)
      if (this.entries.get(key)?.state !== "ready") return false;
    return true;
  }
  constructor(
    private readonly budget: number,
    private readonly concurrency: number,
    private readonly changed: () => void,
    private readonly error: (error: Error) => void = () => {},
  ) {}
  begin(): void {
    this.needed.clear();
  }
  peek(key: string): T | null {
    return this.entries.get(key)?.resource?.value ?? null;
  }
  get(
    key: string,
    load: () => Promise<RenderResource<T>>,
    error?: (error: Error) => void,
  ): T | null {
    if (this.stopped) return null;
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
  private pump(): void {
    if (this.stopped) return;
    for (const [key, e] of this.entries) {
      if (this.active >= this.concurrency) break;
      if (e.state !== "queued") continue;
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
