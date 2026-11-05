// Test-only bridge: execute the actual browser worker entry on a real Node worker thread.
import { parentPort } from "node:worker_threads";
const port = parentPort;
if (port === null) throw new Error("computerThread must run in a worker");
Object.assign(globalThis, {
  addEventListener: (
    name: string,
    listener: (event: { data: unknown }) => void,
  ) => {
    port.on(name, (data: unknown) => {
      listener({ data });
    });
  },
  postMessage: (data: unknown) => {
    port.postMessage(data);
  },
});
await import("../computer.worker.ts");
