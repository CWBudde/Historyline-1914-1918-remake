import { ai, play } from "@historyline/engine";
import type { ComputerRequest, PlanResponse } from "./computerProtocol.ts";

function send(message: PlanResponse): void {
  globalThis.postMessage(message);
}
globalThis.addEventListener(
  "message",
  (event: MessageEvent<ComputerRequest>) => {
    const r = event.data;
    try {
      const mark = play.mark(r.session);
      if (r.kind === "playout") ai.playOut(r.session, [{}, {}], r.maxTurns);
      else
        ai.commands(
          {
            progress: (done, total) => {
              send({ kind: "progress", id: r.id, done, total });
            },
          },
          r.session,
          r.side,
        );
      send({
        kind: "result",
        id: r.id,
        commands: play.commandsSince(r.session, mark),
      });
    } catch (e) {
      send({
        kind: "error",
        id: r.id,
        message: e instanceof Error ? e.message : String(e),
      });
    }
  },
);
