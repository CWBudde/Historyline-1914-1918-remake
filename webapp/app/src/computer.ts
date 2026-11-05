import type { play } from "@historyline/engine";
import type { ComputerRequest, PlanResponse } from "./computerProtocol.ts";

export interface ComputerWorker {
  postMessage(request: ComputerRequest): void;
  addEventListener(
    type: "message",
    listener: (event: { data: PlanResponse }) => void,
  ): void;
  addEventListener(
    type: "error",
    listener: (event: { message: string }) => void,
  ): void;
  terminate(): void;
}
interface Pending {
  resolve: (commands: play.Command[]) => void;
  reject: (error: Error) => void;
  progress: ((done: number, total: number) => void) | undefined;
}
/** Plans a snapshot off the frame loop. The caller replays returned commands onto its live session. */
export class Computer {
  private readonly worker: ComputerWorker;
  private readonly pending = new Map<number, Pending>();
  private next = 0;
  private stopped = false;
  constructor(
    worker: ComputerWorker = new Worker(
      new URL("./computer.worker.ts", import.meta.url),
      { type: "module" },
    ),
  ) {
    this.worker = worker;
    worker.addEventListener("message", ({ data: r }) => {
      const p = this.pending.get(r.id);
      if (p === undefined) return;
      if (r.kind === "progress") {
        p.progress?.(r.done, r.total);
        return;
      }
      this.pending.delete(r.id);
      if (r.kind === "result") p.resolve(r.commands);
      else p.reject(new Error(r.message));
    });
    worker.addEventListener("error", (e) => {
      this.stop(new Error(e.message));
    });
  }
  commands(
    session: play.Session,
    side: number,
    progress?: (done: number, total: number) => void,
  ): Promise<play.Command[]> {
    return this.request(session, { kind: "plan", side }, progress);
  }
  playOut(session: play.Session, maxTurns: number): Promise<play.Command[]> {
    if (!Number.isSafeInteger(maxTurns))
      return Promise.reject(new Error("Invalid playout limit."));
    return this.request(session, { kind: "playout", maxTurns });
  }
  private request(
    session: play.Session,
    order:
      { kind: "plan"; side: number } | { kind: "playout"; maxTurns: number },
    progress?: (done: number, total: number) => void,
  ): Promise<play.Command[]> {
    if (this.stopped) return Promise.reject(new Error("Computer stopped."));
    const id = this.next++;
    return new Promise((resolve, reject) => {
      this.pending.set(id, { resolve, reject, progress });
      try {
        this.worker.postMessage({ ...order, id, session });
      } catch (e) {
        this.pending.delete(id);
        reject(e instanceof Error ? e : new Error(String(e)));
      }
    });
  }
  private stop(error: Error): void {
    if (this.stopped) return;
    this.stopped = true;
    this.worker.terminate();
    for (const p of this.pending.values()) p.reject(error);
    this.pending.clear();
  }
  dispose(): void {
    this.stop(new Error("Computer stopped."));
  }
}
