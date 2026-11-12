import { useEffect } from "preact/hooks";
import type { OfflineMission } from "../offline.ts";
import { missionLabel } from "../offline.ts";
import type { OfflineController } from "./controller.ts";
import "./offline.css";

export interface OfflinePanelProps {
  model: OfflineController;
  mission: OfflineMission & { name: string };
}
export const megabytes = (bytes: number) =>
  `${(bytes / 1_000_000).toFixed(1)} MB`;

const modes: Record<string, string> = {
  cp: "Central Powers",
  ec: "Entente",
  "2p": "Two players",
};

export function OfflinePanel({ model, mission }: OfflinePanelProps) {
  const key = `${mission.month}:${mission.mode}`;
  useEffect(() => {
    void model.select(mission);
    return () => {
      model.cancel();
    };
    // The page offers the map that is selected when it opens.
  }, [model, key]);
  const state = model.state.value,
    kept = model.kept.value,
    storage = model.storage.value;
  const label = `${missionLabel(mission)} · ${mission.name}`;
  let status = "";
  switch (state.phase) {
    case "checking":
      status = "Checking the download size…";
      break;
    case "available":
      status = `${label} is not on this device yet.`;
      break;
    case "downloading":
      status = `Downloading ${megabytes(state.done)} of ${megabytes(state.bytes)}…`;
      break;
    case "ready":
      status = `${label} is downloaded.`;
      break;
    case "error":
      status = state.note;
      break;
    case "idle":
      break;
  }
  return (
    <section class="offline-panel" aria-label="Offline maps">
      <h2>Selected map</h2>
      <div class="shell-well offline-box hl-well">
        <p>{label}</p>
        <p role="status" class="shell-muted">
          {status}
        </p>
        {state.phase === "downloading" && (
          <progress
            class="shell-progress"
            aria-label="Download progress"
            max={Math.max(1, state.bytes)}
            value={state.done}
          />
        )}
        <div class="offline-row">
          {state.phase === "downloading" ? (
            <button
              type="button"
              class="shell-button hl-nameplate"
              onClick={() => {
                model.cancel();
              }}
            >
              Cancel
            </button>
          ) : state.phase === "ready" ? (
            <button type="button" class="shell-button hl-nameplate" disabled>
              Downloaded
            </button>
          ) : (
            <button
              type="button"
              class="shell-button hl-nameplate"
              disabled={state.phase !== "available" && state.phase !== "error"}
              onClick={() => {
                if (state.phase === "error") void model.select(mission);
                else void model.download();
              }}
            >
              {state.phase === "available"
                ? `Download map · ${megabytes(state.bytes)}`
                : state.phase === "error"
                  ? "Try again"
                  : "Download map"}
            </button>
          )}
        </div>
      </div>
      <h2>On this device</h2>
      <div class="shell-well offline-box hl-well">
        {kept.length === 0 ? (
          <p class="shell-muted">No maps are kept for offline play.</p>
        ) : (
          <ul class="offline-list">
            {kept.map((entry) => (
              <li key={entry.key}>
                <span>
                  {entry.label} · {modes[entry.mode] ?? entry.mode}
                </span>
                <span class="shell-muted">{megabytes(entry.bytes)}</span>
                <button
                  type="button"
                  class="shell-button hl-nameplate"
                  disabled={state.phase === "downloading"}
                  onClick={() => {
                    void model.remove(entry.key);
                  }}
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>
        )}
        {kept.length > 0 && (
          <div class="offline-row">
            <button
              type="button"
              class="shell-button hl-nameplate"
              disabled={state.phase === "downloading"}
              onClick={() => {
                void model.removeAll();
              }}
            >
              Remove all · {megabytes(model.keptTotal.value)}
            </button>
          </div>
        )}
        <p class="shell-muted">
          {storage.used === null || storage.quota === null
            ? "Storage use is not reported by this browser."
            : `Browser storage: ${megabytes(storage.used)} used of ${megabytes(storage.quota)}.`}{" "}
          {storage.persistent === null
            ? ""
            : storage.persistent
              ? "Kept maps are protected from automatic clean-up."
              : "The browser may clear kept maps when space runs low."}
        </p>
      </div>
    </section>
  );
}
