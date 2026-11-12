import { play } from "@historyline/engine";
import { useEffect, useLayoutEffect, useRef } from "preact/hooks";
import type { TargetedKeyboardEvent } from "preact";
import type { BattleUIProps } from "./types.ts";
import type { Action } from "./ring.ts";
import {
  captionPlace,
  slotKeys,
  walk,
  type Slot,
  type PlacedItem,
} from "./orderMenu.ts";
import "./orderMenu.css";

/** The name of each order on its caption plate and for screen readers. */
export const orderName: Record<Action, string> = {
  cancel: "Cancel",
  move: "Move",
  attack: "Attack",
  mode: "End phase",
  open: "Contents",
  map: "Map",
  info: "Info",
  exit: "Move out",
  produce: "Produce",
  repair: "Repair",
};
/** The stencil symbol of each menu order (art/screens/plates). */
const icon: Partial<Record<Action, string>> = {
  move: "move",
  attack: "attack",
  open: "contents",
  exit: "exit",
  produce: "produce",
  repair: "repair",
  cancel: "cancel",
};
/** The logical width of a plate and of a symbol in the screens catalog's plates family (hlasset plates.go). */
const plateLogical = 160,
  iconLogical = 96,
  /** The caption plate's logical height (armour/nameplate-polished). */
  nameplateLogical = 70;

/** sentence is a refusal as the caption's second line. */
function sentence(reason: string): string {
  const t = reason.trim();
  return t.charAt(0).toUpperCase() + t.slice(1);
}

/**
 * artUrl is the url() of a screens-catalog part at a drawn scale (lite art only: the parts are painted, their
 * lossless art would cost a lot for no visible gain), "none" until it has arrived.
 */
function artUrl(p: BattleUIProps, id: string, scale: number): string {
  const url = p.assets.image("screens", id, scale, false, true).value.url;
  return url === null ? "none" : `url("${url}")`;
}

/**
 * OrderMenu draws the open order menu: a worn steel hex plate per order on its neighbouring cell, lit under the
 * pointer or focus, stripped and rusty where the order is not possible now, and one brass caption plate naming
 * what is lit. The plates take no pointer events (the battle hit-tests wedges through the map); they are buttons
 * for the keyboard and screen readers.
 */
export function OrderMenu({ p }: { p: BattleUIProps }) {
  const m = p.menu.value,
    root = useRef<HTMLDivElement>(null),
    previous = useRef<Element | null>(null),
    width = m?.layout.plate.width ?? 100,
    scale = width / plateLogical;
  // Opened from the keyboard: focus the first ready plate; on closing, give the focus back.
  useLayoutEffect(() => {
    if (m === null || !m.focus) return;
    previous.current = document.activeElement;
    const first =
      root.current?.querySelector<HTMLButtonElement>(
        '.order-plate:not([aria-disabled="true"])',
      ) ?? root.current?.querySelector<HTMLButtonElement>(".order-plate");
    first?.focus();
  }, [m?.id]);
  useEffect(
    () => () => {
      const prev = previous.current;
      previous.current = null;
      if (
        prev instanceof HTMLElement &&
        prev.isConnected &&
        root.current?.contains(document.activeElement) !== false
      )
        prev.focus();
    },
    [m?.id],
  );
  if (m === null) return null;
  // The art is asked for only once a menu opens, so none of it weighs on the first playable frame.
  const faces = {
      idle: artUrl(p, "plates/plate", scale),
      lit: artUrl(p, "plates/plate-lit", scale),
      stripped: artUrl(p, "plates/plate-stripped", scale),
      attack: artUrl(p, "plates/plate-attack", scale),
      attackLit: artUrl(p, "plates/plate-attack-lit", scale),
    },
    symbols: Record<string, string> = {};
  for (const name of Object.values(icon))
    symbols[name] = artUrl(
      p,
      `plates/icon-${name}`,
      (0.56 * width) / iconLogical,
    );

  const l = m.layout,
    h = l.plate.height,
    litItem =
      m.lit === null || m.lit === "centre"
        ? null
        : (l.items.find((i) => i.slot === m.lit) ?? null),
    keys = (event: TargetedKeyboardEvent<HTMLDivElement>) => {
      if (
        event.key !== "ArrowUp" &&
        event.key !== "ArrowDown" &&
        event.key !== "ArrowLeft" &&
        event.key !== "ArrowRight"
      )
        return;
      const at =
        (event.target instanceof HTMLElement &&
          (event.target.dataset["slot"] as Slot | "centre" | undefined)) ||
        "centre";
      const to = walk(l, at, event.key);
      root.current
        ?.querySelector<HTMLButtonElement>(`[data-slot="${to}"]`)
        ?.focus();
      event.preventDefault();
      event.stopPropagation();
    };
  const plate = (item: PlacedItem) => {
    const attack = item.action === "attack",
      disabled = item.reason !== null,
      lit = m.lit === item.slot,
      face = disabled
        ? faces.stripped
        : attack
          ? lit
            ? faces.attackLit
            : faces.attack
          : lit
            ? faces.lit
            : faces.idle,
      name = orderName[item.action],
      [key, pad] = slotKeys[item.slot];
    return (
      <button
        type="button"
        key={item.slot}
        role="menuitem"
        class="order-plate"
        data-slot={item.slot}
        data-action={item.action}
        data-lit={lit}
        aria-disabled={disabled ? "true" : undefined}
        aria-label={disabled ? `${name}: ${sentence(item.reason ?? "")}` : name}
        aria-keyshortcuts={`${key.toUpperCase()} ${pad}`}
        style={{
          left: `${item.x - width / 2}px`,
          top: `${item.y - h / 2}px`,
          width: `${width}px`,
          height: `${h}px`,
          "--face": face,
          "--symbol": symbols[icon[item.action] ?? "cancel"] ?? "none",
        }}
        onClick={() => {
          p.onMenu(item.slot);
        }}
        onFocus={() => {
          p.onMenuFocus(item.slot);
        }}
        onBlur={() => {
          p.onMenuFocus(null);
        }}
      />
    );
  };
  // The caption: the lit plate's name (and why it is not possible now), the key on a desktop.
  const caption =
    m.lit === null
      ? null
      : m.lit === "centre"
        ? { name: "Cancel", reason: null, key: "Esc" }
        : litItem === null
          ? null
          : {
              name: orderName[litItem.action],
              reason: litItem.reason,
              key: slotKeys[litItem.slot][0].toUpperCase(),
            };
  // The caption grows with the plates (the split screen's HUD at 4K), within touch- and mouse-sized limits.
  const captionHeight = Math.min(56, Math.max(l.touch ? 40 : 32, 0.36 * h)),
    captionWidth =
      caption === null
        ? 0
        : Math.max(
            caption.name.length * 0.4 * captionHeight + 2.1 * captionHeight,
            caption.reason === null
              ? 0
              : caption.reason.length * 0.2 * captionHeight + captionHeight,
          );
  const place = m.lit === null ? null : captionPlace(l, m.lit, captionWidth),
    captionFace = artUrl(
      p,
      "armour/nameplate-polished",
      captionHeight / nameplateLogical,
    );

  const shifted = Math.hypot(l.centre.x - l.unit.x, l.centre.y - l.unit.y) > 1;
  return (
    <div
      ref={root}
      class="order-menu"
      role="menu"
      aria-label="Orders"
      data-visible={m.visible}
      data-touch={l.touch}
      onKeyDown={keys}
    >
      {shifted && (
        <span
          class="order-bracket"
          aria-hidden="true"
          style={{
            left: `${l.unit.x}px`,
            top: `${l.unit.y}px`,
            width: `${Math.hypot(l.centre.x - l.unit.x, l.centre.y - l.unit.y)}px`,
            transform: `rotate(${Math.atan2(l.centre.y - l.unit.y, l.centre.x - l.unit.x)}rad)`,
          }}
        />
      )}
      {l.items.map(plate)}
      <button
        type="button"
        role="menuitem"
        class="order-centre"
        data-slot="centre"
        data-lit={m.lit === "centre"}
        aria-label="Cancel"
        aria-keyshortcuts="Escape"
        style={{
          left: `${l.centre.x - (0.5 * l.step * 96) / 80}px`,
          top: `${l.centre.y - l.step / 2}px`,
          width: `${(l.step * 96) / 80}px`,
          height: `${l.step}px`,
        }}
        onClick={() => {
          p.onMenu("centre");
        }}
        onFocus={() => {
          p.onMenuFocus("centre");
        }}
        onBlur={() => {
          p.onMenuFocus(null);
        }}
      />
      {caption !== null && place !== null && (
        <div
          class="order-caption"
          role="status"
          data-side={place.side}
          style={{
            left: `${place.x}px`,
            top: `${place.y}px`,
            "--face": captionFace,
            "--ph": `${captionHeight}px`,
          }}
        >
          <span class="order-caption-name">
            {!l.touch && <kbd>{caption.key}</kbd>}
            {caption.name}
          </span>
          {caption.reason !== null && (
            <span class="order-caption-reason">{sentence(caption.reason)}</span>
          )}
        </div>
      )}
    </div>
  );
}

/** The order strip's three global orders and their keys. */
const strip = [
  { id: "info", name: "Info" },
  { id: "map", name: "Map" },
  { id: "phase", name: "End phase" },
] as const;

/**
 * OrderStrip is the row of brass nameplates (armour.css) for the orders that are not about a unit: Info, Map and End
 * phase, each with its key. It lies on the steel strip under the map (upright: along the plate's bottom edge) and on
 * the split screen's status strips. End phase is never next to a unit, so no flick reaches it.
 */
export function OrderStrip({
  p,
  className = "",
  keys = { info: "I", map: "O", phase: "F1" },
  style,
}: {
  p: BattleUIProps;
  className?: string;
  /** The key shown on each nameplate (a second simultaneous player has only its End phase gesture). */
  keys?: Partial<Record<"info" | "map" | "phase", string>>;
  style?: Record<string, string>;
}) {
  const b = p.battle,
    ready = play.ready(b.session, b.side),
    blocked = b.playback.busy() || play.ended(b.session) !== null;
  return (
    <nav
      class={`order-strip ${className}`}
      aria-label="Global orders"
      style={style}
    >
      {strip.map((s) => {
        const phase = s.id === "phase",
          label = phase ? (ready ? "Reopen orders" : "End phase") : s.name,
          key = keys[s.id];
        return (
          <button
            type="button"
            key={s.id}
            class={`hl-nameplate order-nameplate${phase ? " battle-ready" : ""}`}
            data-order={s.id}
            aria-pressed={phase ? ready : undefined}
            aria-keyshortcuts={
              key === "F1" || key?.length === 1 ? key : undefined
            }
            aria-label={label}
            disabled={
              phase ? blocked : b.busy() || b.phasePrompt || b.surrenderPrompt
            }
            onClick={() => {
              if (s.id === "phase") p.onPhase();
              else if (s.id === "map") p.onMap();
              else p.onInfo();
            }}
          >
            {key !== undefined && <kbd>{key}</kbd>}
            <span>{label}</span>
          </button>
        );
      })}
    </nav>
  );
}
