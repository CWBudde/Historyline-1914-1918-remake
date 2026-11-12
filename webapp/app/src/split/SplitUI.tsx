import { play, phaseMovement } from "@historyline/engine";
import { useCallback, useLayoutEffect, useRef, useState } from "preact/hooks";
import type { Signal } from "@preact/signals";
import type { TargetedKeyboardEvent } from "preact";
import { BattleUI } from "../ui/BattleUI.tsx";
import { makeInert } from "../ui/inert.ts";
import type { UiAssets } from "../ui/assets.ts";
import type { Rect } from "../camera.ts";
import type { SplitRuntime } from "./runtime.ts";
import { drawWell, type Well } from "./well.ts";
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
const ranks = ["", "01", "02", "03", "04", "05", "10"];
const badgeImages = new Map<string, HTMLImageElement>();
/** wellLabel is a well's line in words, for screen readers and tests. */
function wellLabel(w: Well): string {
  return w.unit === null ? w.text : `${w.unit.size} ${w.unit.name}`;
}
/** WellCanvas paints one status well in the pixel font over the frame's recess (cmd/historyline/well.go). */
function WellCanvas({
  runtime,
  assets,
  side,
  well,
  visible,
}: {
  runtime: SplitRuntime;
  assets: UiAssets;
  side: 0 | 1;
  well: Well;
  visible: boolean;
}) {
  const canvas = useRef<HTMLCanvasElement>(null),
    [, loaded] = useState(0),
    screen = runtime.screen,
    st = screen.status[side],
    exp = well.unit?.exp ?? 0,
    rank = exp > 0 ? ranks[Math.min(exp, ranks.length - 1)] : undefined,
    url =
      rank === undefined
        ? null
        : assets.image("screens", `rank/${rank}`, 2).value.url;
  let badge: HTMLImageElement | null = null;
  if (url !== null) {
    badge = badgeImages.get(url) ?? null;
    if (badge === null) {
      badge = new Image();
      badge.onload = () => {
        loaded((n) => n + 1);
      };
      badge.src = url;
      badgeImages.set(url, badge);
    }
  }
  useLayoutEffect(() => {
    const c = canvas.current,
      fonts = runtime.fonts;
    if (c === null) return;
    if (c.width !== st.width || c.height !== st.height) {
      c.width = st.width;
      c.height = st.height;
    }
    const ctx = c.getContext("2d");
    if (ctx === null) return;
    ctx.clearRect(0, 0, c.width, c.height);
    if (fonts === null || !visible) return;
    const shift = (r: Rect) => ({ ...r, x: r.x - st.x, y: r.y - st.y });
    drawWell(
      ctx,
      fonts.small,
      side,
      shift(screen.wellText(side)),
      shift(screen.wellBar(side)),
      screen.ui,
      well,
      (_, cx, cy, size) => {
        if (badge === null || !badge.complete || badge.naturalWidth === 0)
          return;
        const k = size / Math.max(badge.naturalWidth, badge.naturalHeight);
        ctx.imageSmoothingEnabled = true;
        ctx.drawImage(
          badge,
          cx - (badge.naturalWidth * k) / 2,
          cy - (badge.naturalHeight * k) / 2,
          badge.naturalWidth * k,
          badge.naturalHeight * k,
        );
      },
    );
  });
  return <canvas ref={canvas} class="split-well-canvas" aria-hidden="true" />;
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
          well = runtime.well(side),
          phase =
            play.phase(b.session, side) === phaseMovement
              ? "Movement"
              : "Attack",
          status = b.status || wellLabel(well);
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
              <WellCanvas
                runtime={runtime}
                assets={assets}
                side={side}
                well={well}
                visible={visible}
              />
              <span class="split-well-text">
                {visible ? status : "Waiting for the other player"}
              </span>
              {active && (
                <small class="split-well-text">
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
