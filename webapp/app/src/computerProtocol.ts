import type { play } from "@historyline/engine";

// Session contains data only: postMessage snapshots its Maps, records and pending routes.
export interface PlanRequest {
  kind: "plan";
  id: number;
  session: play.Session;
  side: number;
}
export interface PlayOutRequest {
  kind: "playout";
  id: number;
  session: play.Session;
  maxTurns: number;
}
export type ComputerRequest = PlanRequest | PlayOutRequest;
export type PlanResponse =
  | { kind: "progress"; id: number; done: number; total: number }
  | { kind: "result"; id: number; commands: play.Command[] }
  | { kind: "error"; id: number; message: string };
