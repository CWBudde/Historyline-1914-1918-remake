import { play, phaseMovement } from "@historyline/engine";
import { useCallback, useLayoutEffect, useRef, useState } from "preact/hooks";
import type { Signal } from "@preact/signals";
import type { TargetedKeyboardEvent } from "preact";
import { BattleUI } from "../ui/BattleUI.tsx";
import { OrderStrip } from "../ui/OrderMenu.tsx";
import { makeInert } from "../ui/inert.ts";
import type { UiAssets } from "../ui/assets.ts";
import type { Rect } from "../camera.ts";
import type { SplitRuntime } from "./runtime.ts";
import { drawWell, type Well } from "./well.ts";
import {
  contentsLayout,
  drawContents,
  drawGameInfo,
  drawUnitInfo,
  gameInfoScale,
  panelScale,
  plateRect,
  unitInfoLayout,
} from "./overlay.ts";
import { RivetSheet, Tile, type Surfaces } from "./painter.ts";
import { spriteID } from "../unitFrame.ts";
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
/**
 * stripPlace puts the order strip at the outer end of a panel's status strip, clear of the well's flag: nameplates at least
 * 26 px tall, centred on the strip, so they overhang its bevel where the strip is thinner than that.
 */
function stripPlace(rect: Rect, dpr: number): Record<string, string> {
  const h = rect.height / dpr,
    ph = Math.max(26, Math.min(64, h * 1.1));
  return {
    position: "absolute",
    right: `${window.innerWidth - (rect.x + rect.width) / dpr + 8 + h * 1.1}px`,
    top: `${rect.y / dpr + h / 2 - ph / 2}px`,
    "--strip-ph": `${ph}px`,
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
/**
 * PanelOverlay paints the view open in side's panel as the desktop does (internal/dospanel) over the map; the
 * battle UI's DOM view stays beneath it, visually hidden, for screen readers and tests.
 */
const pictures = new Map<string, HTMLImageElement>();
/** picture is a decoded image of the screen or unit art at 2x, null until it has loaded. */
function picture(
  assets: UiAssets,
  catalog: string,
  id: string,
  loaded: () => void,
): HTMLImageElement | null {
  const url = assets.image(catalog, id, 2).value.url;
  if (url === null) return null;
  let img = pictures.get(url);
  if (img === undefined) {
    img = new Image();
    img.onload = loaded;
    img.src = url;
    pictures.set(url, img);
  }
  return img.complete && img.naturalWidth > 0 ? img : null;
}
const surfaceCache = new Map<HTMLImageElement, Tile | RivetSheet>();
function pixels(img: HTMLImageElement): ImageData | null {
  const c = document.createElement("canvas");
  c.width = img.naturalWidth;
  c.height = img.naturalHeight;
  const ctx = c.getContext("2d", { willReadFrequently: true });
  if (ctx === null) return null;
  ctx.drawImage(img, 0, 0);
  return ctx.getImageData(0, 0, c.width, c.height);
}
/** panelSurfaces are the painted steel and rivets the plates are modulated with (dospanel.SetSurfaces). */
function panelSurfaces(assets: UiAssets, loaded: () => void): Surfaces {
  const of = <T extends Tile | RivetSheet>(
    id: string,
    make: (p: ImageData) => T,
  ): T | null => {
    const img = picture(assets, "screens", id, loaded);
    if (img === null) return null;
    let s = surfaceCache.get(img);
    if (s === undefined) {
      const p = pixels(img);
      if (p === null) return null;
      s = make(p);
      surfaceCache.set(img, s);
    }
    return s as T;
  };
  return {
    steel: of("surface/steel", (p) => new Tile(p)),
    rivets: of("surface/rivets", (p) => new RivetSheet(p)),
  };
}
function PanelOverlay({
  runtime,
  assets,
  side,
}: {
  runtime: SplitRuntime;
  assets: UiAssets;
  side: 0 | 1;
  /** The UI revision: the signals integration skips a render whose props are unchanged. */
  tick: number;
}) {
  const [, loaded] = useState(0),
    redraw = () => {
      loaded((n) => n + 1);
    },
    canvas = useRef<HTMLCanvasElement>(null),
    view = runtime.overlay(side),
    screen = runtime.screen,
    panel = screen.panels[side],
    surfaces = panelSurfaces(assets, redraw),
    unit = view?.kind === "unit" ? view : null,
    portrait =
      unit?.family == null
        ? null
        : picture(assets, "screens", `portrait/${unit.family}`, redraw),
    sprite =
      unit?.family == null
        ? null
        : picture(
            assets,
            "units",
            spriteID(unit.family, unit.owner, "n", false),
            redraw,
          );
  useLayoutEffect(() => {
    const c = canvas.current,
      fonts = runtime.fonts;
    if (c === null) return;
    if (c.width !== panel.width || c.height !== panel.height) {
      c.width = panel.width;
      c.height = panel.height;
    }
    const ctx = c.getContext("2d");
    if (ctx === null) return;
    ctx.clearRect(0, 0, c.width, c.height);
    if (fonts === null || view === null) return;
    const local = { x: 0, y: 0, width: panel.width, height: panel.height },
      room = plateRect(screen, side),
      base = panelScale(screen);
    if (view.kind === "game")
      drawGameInfo(ctx, local, gameInfoScale(room, base), view.info, fonts);
    else if (view.kind === "map") {
      const pic = document.createElement("canvas");
      pic.width = view.picture.width;
      pic.height = view.picture.height;
      pic
        .getContext("2d")
        ?.putImageData(
          new ImageData(
            view.picture.data,
            view.picture.width,
            view.picture.height,
          ),
          0,
          0,
        );
      ctx.fillStyle = "rgb(48,48,36)";
      ctx.fillRect(0, 0, panel.width, panel.height);
      ctx.imageSmoothingEnabled = false;
      ctx.drawImage(
        pic,
        view.rect.x - panel.x,
        view.rect.y - panel.y,
        pic.width * view.ui,
        pic.height * view.ui,
      );
    } else if (view.kind === "contents")
      drawContents(
        ctx,
        local,
        contentsLayout(room, base).scale,
        view.contents,
        fonts,
        surfaces,
        (slot, photo) =>
          slot.family === null
            ? null
            : photo
              ? picture(assets, "screens", `portrait/${slot.family}`, redraw)
              : picture(
                  assets,
                  "units",
                  spriteID(slot.family, slot.owner, "n", false),
                  redraw,
                ),
      );
    else
      drawUnitInfo(
        ctx,
        local,
        unitInfoLayout(room, base).scale,
        view.info,
        fonts,
        surfaces,
        portrait,
        sprite,
      );
  });
  return (
    <canvas
      ref={canvas}
      class="split-overlay"
      hidden={view === null}
      data-view={view?.kind ?? "map"}
      aria-hidden="true"
    />
  );
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
                <>
                  <BattleUI
                    {...runtime.panelProps(side, assets, revision, state)}
                    embedded
                  />
                  <PanelOverlay
                    runtime={runtime}
                    assets={assets}
                    side={side}
                    tick={tick}
                  />
                </>
              ) : (
                <p class="split-covered">Waiting for the other player</p>
              )}
            </section>
            <div
              class={`split-well side-${side}`}
              data-side={side}
              role="status"
              style={place(runtime.screen.status[side], dpr)}
              aria-label={`${name(side)} · ${phase} · ${play.ready(b.session, side) ? "Orders confirmed" : "Giving orders"}`}
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
            </div>
            {active && visible && (
              <OrderStrip
                p={runtime.panelProps(side, assets, revision, state)}
                className="split-strip"
                keys={
                  runtime.local?.mode === "simultaneous" && side === 1
                    ? { phase: "Ctrl ←" }
                    : { info: "I", map: "O", phase: "F1" }
                }
                style={stripPlace(runtime.screen.status[side], dpr)}
              />
            )}
          </div>
        );
      })}
      <div class="split-frame-host" ref={frameHost} />
    </div>
  );
}
