import { level, play, phaseMovement } from "@historyline/engine";
import { useCallback, useLayoutEffect, useRef } from "preact/hooks";
import type { Signal } from "@preact/signals";
import type { TargetedKeyboardEvent } from "preact";
import { BattleUI } from "../ui/BattleUI.tsx";
import { makeInert } from "../ui/inert.ts";
import type { UiAssets } from "../ui/assets.ts";
import type { Rect } from "../camera.ts";
import type { SplitRuntime } from "./runtime.ts";
import "./split.css";

interface Props {
  runtime: SplitRuntime;
  assets: UiAssets;
  revision: Signal<number>;
  state: { busy: boolean; note: string };
}
const name = (side: number) => (side === 0 ? "Central Powers" : "Entente");
function place(rect: Rect, dpr: number) {
  return {
    left: `${rect.x / dpr}px`,
    top: `${rect.y / dpr}px`,
    width: `${rect.width / dpr}px`,
    height: `${rect.height / dpr}px`,
  };
}
function Handover({ runtime, side }: { runtime: SplitRuntime; side: 0 | 1 }) {
  const take = useRef<HTMLButtonElement>(null),
    dialog = useRef<HTMLElement>(null);
  useLayoutEffect(() => {
    const stage = document.getElementById("stage"),
      previous = document.activeElement,
      release = stage === null ? () => {} : makeInert(stage);
    take.current?.focus();
    return () => {
      release();
      if (previous instanceof HTMLElement && previous.isConnected)
        previous.focus();
    };
  }, [side]);
  const keys = (e: TargetedKeyboardEvent<HTMLElement>) => {
    if (e.key !== "Tab") return;
    const controls =
        dialog.current?.querySelectorAll<HTMLButtonElement>("button"),
      first = controls?.[0],
      last = controls?.[1];
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault();
      last?.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault();
      first?.focus();
    }
  };
  return (
    <div class="hotseat-handover">
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="handover-title"
        ref={dialog}
        onKeyDown={keys}
      >
        <h1 id="handover-title">{name(side)} to play</h1>
        <p>
          Pass the device to the {name(side)} player. The battlefield stays
          covered until they take control.
        </p>
        <button
          ref={take}
          onClick={() => {
            runtime.takeOver();
          }}
        >
          Take control
        </button>
        <button
          onClick={() => {
            runtime.requestLeave();
          }}
        >
          Save or leave
        </button>
      </section>
    </div>
  );
}

export function SplitUI({ runtime, assets, revision, state }: Props) {
  const frameHost = useCallback(
    (node: HTMLDivElement | null) => {
      runtime.attachFrame(node);
    },
    [runtime],
  );
  const tick = revision.value,
    handover = runtime.handover;
  if (handover !== null)
    return <Handover key={handover} runtime={runtime} side={handover} />;
  if (!runtime.original)
    return <BattleUI {...runtime.props(assets, revision, state)} />;
  const dpr = window.devicePixelRatio;
  return (
    <div
      class="split-screen"
      data-revision={tick}
      data-mode={runtime.local?.mode ?? "single"}
    >
      {([0, 1] as const).map((side) => {
        const b = runtime.panels[side].battle,
          active = runtime.canInteract(side),
          visible = runtime.visible(side),
          unit = b.view().units.get(level.coordKey(b.cursor)),
          phase =
            play.phase(b.session, side) === phaseMovement
              ? "Movement"
              : "Attack",
          status =
            b.status ||
            (unit?.side === side
              ? `${unit.size} ${unit.name}`
              : `${phase} phase · Turn ${play.turn(b.session)}`);
        return (
          <div key={side}>
            <section
              class={`split-panel ${visible ? "" : "split-hidden"}`}
              data-side={side}
              aria-label={`${name(side)} panel`}
              style={place(runtime.screen.panels[side], dpr)}
              inert={
                !active &&
                !(runtime.local === null && b.playback.scene() !== null)
              }
            >
              {visible ? (
                <BattleUI
                  {...runtime.panelProps(side, assets, revision, state)}
                  embedded
                />
              ) : (
                <p class="split-covered">Waiting for the other player</p>
              )}
            </section>
            <button
              class={`split-well side-${side}`}
              data-side={side}
              style={place(runtime.screen.status[side], dpr)}
              disabled={
                !active || b.playback.busy() || play.ended(b.session) !== null
              }
              aria-label={`${name(side)} · ${phase} · ${play.ready(b.session, side) ? "Reopen orders" : "Ready"}`}
              onClick={() => {
                void b.requestPhase();
              }}
            >
              <span>{visible ? status : "Waiting for the other player"}</span>
              {active && (
                <small>
                  {runtime.local?.mode === "simultaneous" && side === 1
                    ? "Ctrl + Left"
                    : "F1"}
                </small>
              )}
            </button>
          </div>
        );
      })}
      <div class="split-frame-host" ref={frameHost} />
    </div>
  );
}
