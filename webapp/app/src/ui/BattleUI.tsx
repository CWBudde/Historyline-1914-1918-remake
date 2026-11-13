import { designNames } from "../original/design.ts";
import { level, play, phaseMovement } from "@historyline/engine";
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "preact/hooks";
import type { ComponentChildren, TargetedKeyboardEvent } from "preact";
import { spriteID } from "../unitFrame.ts";
import { shellUnit } from "../shell/materials.ts";
import type { BattleUIProps } from "./types.ts";
import { layoutNames } from "./layout.tsx";
import { OrderMenu, OrderStrip } from "./OrderMenu.tsx";
import "../shell/fonts.css";
import "./armour.css";
import "./battle.css";

/*
 * The battle screen in the worn army metal of the menus (armour.css): the map seen through an opening in a riveted
 * plate, a steel strip with a signal lamp, the side's plaque, the phase stencilled and the brass plates of the global
 * orders; what opens (the situation, the small map, a unit's type, a building's contents, the menu) is a steel board
 * bolted over the map, its values stencilled in dark wells. Upright, the map runs from edge to edge between the strip
 * along the top and a row of the orders' plates; the unit, the ground and the queued attacks open on Info as a board.
 */

type Props = BattleUIProps;
export const ownerName = (owner: number): string =>
  owner === 1 ? "Central Powers" : owner === 2 ? "Entente" : "Neutral";
const phaseName = (phase: number): string =>
  phase === phaseMovement ? "Movement" : "Attack";
const groundName = (id: string): string =>
  id
    .replace(/_w$/u, " (winter)")
    .replace(/([a-z])([A-Z])/gu, "$1 $2")
    .replace(/([a-z])([0-9])/giu, "$1 $2")
    .replace(/_/gu, " ")
    .replace(/^./u, (letter) => letter.toUpperCase());
const scales = [0.75, 1, 1.25, 1.5, 2] as const;
const layoutShort: Record<string, string> = {
  auto: "Automatic",
  single: "Single map",
  original: "Split screen",
};

/** usePortrait follows whether the window stands upright: the strip at the top, the plates below the map. */
export function usePortrait(): boolean {
  const query = "(orientation: portrait)",
    [portrait, setPortrait] = useState(
      () => typeof window !== "undefined" && window.matchMedia(query).matches,
    );
  useEffect(() => {
    const m = window.matchMedia(query),
      update = () => {
        setPortrait(m.matches);
      };
    update();
    m.addEventListener("change", update);
    return () => {
      m.removeEventListener("change", update);
    };
  }, []);
  return portrait;
}

/** Plaque is the enamel plaque of a side (art/screens/armour), a painted flag until the art has arrived. */
export function Plaque({
  assets,
  owner,
  uiScale,
  className = "",
}: {
  assets: Props["assets"];
  owner: number;
  uiScale: number;
  className?: string;
}) {
  if (owner !== 1 && owner !== 2)
    return (
      <span
        class={`battle-flag owner-${owner} ${className}`}
        role="img"
        aria-label={ownerName(owner)}
      />
    );
  // The menus' plaques at the menus' size, so the battle takes the art they already fetched.
  const id = owner === 1 ? "armour/plaque-germany" : "armour/plaque-france",
    scale =
      (56 * shellUnit(window.innerWidth, window.innerHeight, uiScale)) /
      (owner === 1 ? 169 : 164),
    image = assets.image("screens", id, scale, false, true).value;
  return image.url === null ? (
    <span
      class={`battle-flag owner-${owner} ${className}`}
      role="img"
      aria-label={ownerName(owner)}
    />
  ) : (
    <img
      class={`battle-plaque ${className}`}
      src={image.url}
      alt={ownerName(owner)}
      draggable={false}
    />
  );
}

/** Lamp is the strip's signal lamp: lit while its side may give orders. */
export function Lamp({ lit }: { lit: boolean }) {
  return <span class="hl-lamp" data-lit={lit} aria-hidden="true" />;
}

/** MenuPlate opens and closes the battle's menu board. */
export function MenuPlate({ p }: { p: Props }) {
  const open = p.menuOpen.value;
  return (
    <button
      type="button"
      class="hl-nameplate battle-menu-plate"
      aria-expanded={open}
      aria-controls={`battle-menu-${p.battle.side}`}
      onClick={() => {
        p.menuOpen.value = !open;
      }}
    >
      Menu
    </button>
  );
}

function Art({
  p,
  catalog,
  id,
  alt,
  className = "",
}: {
  p: Props;
  catalog: string;
  id: string;
  alt: string;
  className?: string;
}) {
  const image = p.assets.image(catalog, id, p.scale.value).value;
  return image.url === null ? (
    <span
      class={`battle-art battle-art-empty ${className}`}
      role={alt === "" ? undefined : "img"}
      aria-label={alt || undefined}
    />
  ) : (
    <img
      class={`battle-art ${className}`}
      src={image.url}
      alt={alt}
      draggable={false}
    />
  );
}

function UnitArt({
  p,
  unit,
  portrait = false,
}: {
  p: Props;
  unit: play.UnitView;
  portrait?: boolean;
}) {
  const family = p.families[unit.id]?.family;
  if (family === undefined) return null;
  return (
    <Art
      p={p}
      catalog={portrait ? "screens" : "units"}
      id={
        portrait
          ? `portrait/${family}`
          : spriteID(family, unit.owner, "n", false)
      }
      alt={portrait ? `${unit.name} portrait` : ""}
      className={portrait ? "battle-photo" : "battle-unit-sprite"}
    />
  );
}

/** Values is a list of names and values stencilled in a well, the values in amber. */
function Values({
  rows,
  className = "",
}: {
  rows: readonly (readonly [string, ComponentChildren])[];
  className?: string;
}) {
  return (
    <dl class={`battle-values hl-well ${className}`}>
      {rows.map(([name, value]) => (
        <div key={name}>
          <dt>{name}</dt>
          <dd>{value}</dd>
        </div>
      ))}
    </dl>
  );
}

function UnitCard({
  p,
  unit,
  full = false,
}: {
  p: Props;
  unit: play.UnitView;
  full?: boolean;
}) {
  const b = p.battle;
  // A panel never turns a visible enemy record into private type/budget information.
  if (unit.side !== b.side) return null;
  const info = full ? play.typeInfo(b.session, b.side, unit.type) : null,
    rank = ["", "01", "02", "03", "04", "05", "10"][
      Math.min(6, Math.max(0, unit.exp))
    ];
  return (
    <section class="battle-unit-card" aria-label={`${unit.name} information`}>
      <div class="battle-unit-heading">
        <div class="battle-photo-well hl-well">
          <UnitArt p={p} unit={unit} portrait />
        </div>
        <div class="battle-unit-name">
          <h3>{unit.name}</h3>
          <span class="battle-owner">
            <Plaque
              assets={p.assets}
              owner={unit.owner}
              uiScale={p.scale.value}
            />
            {ownerName(unit.owner)}
          </span>
          <UnitArt p={p} unit={unit} />
        </div>
      </div>
      <Values
        rows={[
          [
            "Group",
            <>
              {unit.size} / {unit.full}
            </>,
          ],
          [
            "Experience",
            <>
              {unit.exp}
              {rank && (
                <Art
                  p={p}
                  catalog="screens"
                  id={`rank/${rank}`}
                  alt=""
                  className="battle-rank"
                />
              )}
            </>,
          ],
          ["Budget", unit.budget],
          [
            "Order",
            unit.target !== null
              ? `Target ${unit.target.line}, ${unit.target.col}`
              : unit.moved
                ? "Given"
                : "Available",
          ],
        ]}
      />
      {info !== null && (
        <>
          <h4>
            Weapons <small>power / range</small>
          </h4>
          <Values
            className="battle-weapons"
            rows={(
              [
                ["Air", info.air],
                ["Ground", info.ground],
                ["Water", info.water],
              ] as const
            ).map(
              ([name, weapon]) =>
                [
                  name,
                  <>
                    {weapon.power} / {weapon.range}
                  </>,
                ] as const,
            )}
          />
          <Values
            rows={[
              ["Armour", info.armour],
              ["Speed", info.speed],
              ["Weight", info.weight],
              ["Maximum group", info.group],
            ]}
          />
        </>
      )}
    </section>
  );
}

function Terrain({ p }: { p: Props }) {
  const b = p.battle,
    at = b.hover ?? b.cursor,
    key = level.coordKey(at),
    h =
      play.reworked(b.session).get(key) ??
      b.session.level.lines[at.line]?.[at.col];
  if (h === undefined) return null;
  const entrance = play.entrance(b.session, at),
    building = b.view().buildings.get(level.coordKey(entrance ?? at)),
    moving = play.phase(b.session, b.side) === phaseMovement,
    legal = moving ? b.destinations() : b.targets(),
    allowed =
      legal.has(key) ||
      (entrance !== null && legal.has(level.coordKey(entrance)));
  return (
    <section
      class="battle-terrain"
      aria-labelledby={`battle-terrain-title-${p.battle.side}`}
    >
      <h3 id={`battle-terrain-title-${p.battle.side}`}>
        Terrain{" "}
        <small>
          {at.line}, {at.col}
        </small>
      </h3>
      <Values
        rows={[
          ["Ground", groundName(h.terrain)],
          [
            "Features",
            h.infra.ids.length === 0
              ? "None"
              : h.infra.ids.map(groundName).join(", "),
          ],
          ...(building === undefined
            ? []
            : [
                [
                  "Owner",
                  <>
                    <Plaque
                      assets={p.assets}
                      owner={building.owner}
                      uiScale={p.scale.value}
                    />{" "}
                    {ownerName(building.owner)}
                  </>,
                ] as const,
              ]),
        ]}
      />
      {b.selected === null ? (
        <p class="battle-muted">
          Select your unit to see legal moves or targets.
        </p>
      ) : (
        <>
          <p class="battle-notice">
            {allowed
              ? moving
                ? "Legal movement destination"
                : "Legal action target"
              : moving
                ? "Not a legal movement destination"
                : "Not a legal action target"}
          </p>
          <button
            type="button"
            class="hl-nameplate"
            disabled={!allowed}
            onClick={() => {
              b.click(at);
            }}
          >
            {moving ? "Move here" : "Target this hex"}
          </button>
        </>
      )}
    </section>
  );
}

/** ownUnit is the side's unit the orders are about: the selected one, else the one under the cursor. */
function ownUnit(b: Props["battle"]): play.UnitView | null {
  const view = b.view(),
    selected =
      b.exiting?.unit ??
      [...view.units.values()].find((u) => u.index === b.selected);
  let cursorUnit = view.units.get(level.coordKey(b.hover ?? b.cursor));
  if (cursorUnit?.head !== null && cursorUnit?.head !== undefined)
    cursorUnit = view.units.get(level.coordKey(cursorUnit.head));
  return selected?.side === b.side
    ? selected
    : cursorUnit?.side === b.side
      ? cursorUnit
      : null;
}

/** selection is the upright strip's second line: the own unit and its group, else the ground, else the turn. */
function selection(b: Props["battle"]): string {
  const own = ownUnit(b);
  if (own !== null) return `${own.name} · ${own.size}/${own.full}`;
  const at = b.hover ?? b.cursor,
    h =
      play.reworked(b.session).get(level.coordKey(at)) ??
      b.session.level.lines[at.line]?.[at.col];
  return h === undefined
    ? `Turn ${play.turn(b.session)}`
    : groundName(h.terrain);
}

function Orders({ p }: { p: Props }) {
  const b = p.battle,
    view = b.view(),
    own = ownUnit(b),
    blocked =
      b.busy() ||
      b.phasePrompt ||
      b.surrenderPrompt ||
      play.ended(b.session) !== null,
    orders = b.orders();
  return (
    <>
      <section
        class="battle-orders"
        aria-labelledby={`battle-orders-title-${p.battle.side}`}
      >
        <h2 id={`battle-orders-title-${p.battle.side}`} class="battle-hidden">
          Orders
        </h2>
        {own === null ? (
          <p class="battle-muted">
            Tap your unit, then its destination or target.
          </p>
        ) : (
          <>
            <UnitCard p={p} unit={own} />
            <div class="battle-button-row">
              <button
                type="button"
                class="hl-nameplate"
                disabled={blocked || play.ready(b.session, b.side)}
                onClick={() => {
                  b.click(own.at);
                }}
              >
                Select unit
              </button>
              <button
                type="button"
                class="hl-nameplate"
                disabled={blocked}
                onClick={() => {
                  b.showInfo(own.at);
                }}
              >
                Type information
              </button>
              {own.carrier && (
                <button
                  type="button"
                  class="hl-nameplate"
                  disabled={blocked}
                  onClick={() => {
                    b.openContents(own.at);
                  }}
                >
                  Contents
                </button>
              )}
            </div>
          </>
        )}
        {b.exiting !== null && (
          <p class="battle-notice">
            Moving out: choose a highlighted destination, or cancel.
          </p>
        )}
        {b.selected !== null && (
          <button
            type="button"
            class="hl-nameplate battle-wide-plate"
            onClick={() => {
              b.cancel();
            }}
          >
            Cancel selection
          </button>
        )}
      </section>
      <Terrain p={p} />
      <section class="battle-queue" aria-label="Your queued attack orders">
        <h3>
          Queued attacks <small>{orders.length}</small>
        </h3>
        {orders.length === 0 ? (
          <p class="battle-muted">No attacks queued.</p>
        ) : (
          <ol class="battle-order-list hl-well">
            {orders.map((order, index) => (
              <li key={index}>
                {[...view.units.values()].find((u) => u.index === order.unit)
                  ?.name ?? "Unit"}{" "}
                <span class="hl-amber">
                  → {order.to.line}, {order.to.col}
                </span>
              </li>
            ))}
          </ol>
        )}
      </section>
    </>
  );
}

function Contents({ p }: { p: Props }) {
  const b = p.battle,
    c = b.contents,
    box = c.box;
  const list = useRef<HTMLOListElement>(null);
  useEffect(() => {
    list.current
      ?.querySelector('[aria-current="true"]')
      ?.scrollIntoView({ block: "nearest" });
  }, [box?.pick, box?.producing]);
  if (box === null) return null;
  const slots = c.slots(),
    unit = c.unit(),
    resources = c.resources(),
    actions = c.actions(),
    profile = b.session.game.profile,
    repairCost =
      unit !== null && profile.builders.includes(unit.type)
        ? profile.builderRepair
        : profile.repair,
    options = c.options();
  return (
    <section
      class="battle-contents"
      aria-labelledby={`battle-contents-title-${p.battle.side}`}
    >
      <div class="battle-section-heading">
        <h2 id={`battle-contents-title-${p.battle.side}`}>
          {groundName(c.name())}
        </h2>
        <button
          type="button"
          class="hl-nameplate"
          onClick={() => {
            b.cancel();
          }}
        >
          {box.producing ? "Back to slots" : "Close"}
        </button>
      </div>
      <Values
        rows={[
          ["Resources", resources.stored],
          ["Total energy", resources.total],
        ]}
      />
      <div class="battle-slots" role="group" aria-label="Seven container slots">
        {Array.from({ length: play.Slots }, (_, slot) => {
          const u = slots[slot]?.unit;
          return (
            <button
              type="button"
              key={slot}
              class={`battle-slot hl-well${box.slot === slot ? " selected" : ""}`}
              aria-pressed={box.slot === slot}
              aria-label={`Slot ${slot + 1}: ${u?.name ?? "empty"}${u === null || u === undefined ? "" : `, group ${u.size} of ${u.full}`}`}
              onClick={() => {
                b.selectSlot(slot);
              }}
            >
              <span class="battle-slot-number">{slot + 1}</span>
              {u === undefined || u === null ? (
                <span class="battle-slot-empty">Empty</span>
              ) : (
                <>
                  <UnitArt p={p} unit={u} />
                  <span>{u.name}</span>
                  <small>
                    {u.size} / {u.full}
                  </small>
                </>
              )}
            </button>
          );
        })}
      </div>
      {unit === null ? (
        <p class="battle-muted">Slot {box.slot + 1} is empty.</p>
      ) : (
        <UnitCard p={p} unit={unit} />
      )}
      {box.producing ? (
        <>
          <h3>Production</h3>
          <p class="battle-muted">Builds use the side's energy account.</p>
          <ol class="battle-production-list hl-well" ref={list}>
            {options.map((option, index) => (
              <li key={`${option.type}/${index}`}>
                <button
                  type="button"
                  class="hl-nameplate"
                  aria-current={box.pick === index ? "true" : undefined}
                  disabled={option.reason !== null}
                  onFocus={() => {
                    c.selectBuild(index);
                    p.revision.value++;
                  }}
                  onClick={() => {
                    b.produce(option.type);
                  }}
                >
                  <span>{option.name}</span>
                  <strong>{option.cost} energy</strong>
                </button>
                {option.reason !== null && (
                  <p class="battle-reason">{option.reason}</p>
                )}
              </li>
            ))}
          </ol>
        </>
      ) : (
        <div class="battle-content-actions">
          <div>
            <button
              type="button"
              class="hl-nameplate"
              disabled={actions.leave !== null}
              onClick={() => {
                b.run("exit");
              }}
            >
              Move out
            </button>
            {actions.leave !== null && (
              <p class="battle-reason">{actions.leave}</p>
            )}
          </div>
          <div>
            <button
              type="button"
              class="hl-nameplate"
              disabled={actions.repair !== null}
              onClick={() => {
                b.run("repair");
              }}
            >
              Repair · {repairCost} energy
            </button>
            {actions.repair !== null && (
              <p class="battle-reason">{actions.repair}</p>
            )}
          </div>
          {options.length > 0 && (
            <div>
              <button
                type="button"
                class="hl-nameplate"
                disabled={actions.produce !== null}
                onClick={() => {
                  b.run("produce");
                }}
              >
                Produce
              </button>
              {actions.produce !== null && (
                <p class="battle-reason">{actions.produce}</p>
              )}
            </div>
          )}
        </div>
      )}
    </section>
  );
}

/** GameInfo is the situation: the battle's state, then the two sides' holdings side by side. */
function GameInfo({ p }: { p: Props }) {
  const b = p.battle,
    s = play.stats(b.session),
    plaque = (owner: number) => (
      <Plaque assets={p.assets} owner={owner} uiScale={p.scale.value} />
    ),
    holdings = [s.sides[0], s.sides[1], s.map] as const;
  return (
    <section
      class="battle-game"
      aria-labelledby={`battle-game-title-${p.battle.side}`}
    >
      <h2 id={`battle-game-title-${p.battle.side}`}>Situation</h2>
      <Values
        className="battle-situation"
        rows={[
          ["Turn", s.round],
          ["Level", s.level || String(s.levelNo).padStart(2, "0")],
          ["Phase", phaseName(s.mode[b.side] ?? phaseMovement)],
          ["Orders", s.orders[b.side]],
          ...(p.best === undefined ? [] : [["Best score", p.best] as const]),
          ["Score", s.score],
          ["Limit", s.limit === 0 ? "No limit" : s.limit],
        ]}
      />
      <div class="battle-holdings hl-well">
        <table>
          <thead>
            <tr>
              <td />
              <th scope="col">{plaque(play.faction(0))}</th>
              <th scope="col">{plaque(play.faction(1))}</th>
              <th scope="col">Whole map</th>
            </tr>
          </thead>
          <tbody>
            {(
              [
                ["Units", (h: (typeof holdings)[number]) => h.units],
                ["Depots", (h: (typeof holdings)[number]) => h.depots],
                ["Factories", (h: (typeof holdings)[number]) => h.factories],
              ] as const
            ).map(([name, value]) => (
              <tr key={name}>
                <th scope="row">{name}</th>
                {holdings.map((h, i) => (
                  <td key={i}>{value(h)}</td>
                ))}
              </tr>
            ))}
            <tr class="battle-gap">
              <th scope="row">Energy</th>
              <td>{s.energy[0]}</td>
              <td>{s.energy[1]}</td>
              <td />
            </tr>
            <tr>
              <th scope="row">Income</th>
              <td>{s.income[0]}</td>
              <td>{s.income[1]}</td>
              <td />
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  );
}

/** A choice on the menu: a plate that steps through its values, its key beside the name. */
function MenuChoice({
  label,
  keys,
  onClick,
  disabled,
  pressed,
  expanded,
  controls,
}: {
  label: string;
  keys?: string;
  onClick: () => void;
  disabled?: boolean;
  pressed?: boolean;
  expanded?: boolean;
  controls?: string;
}) {
  return (
    <button
      type="button"
      class="hl-nameplate battle-wide-plate"
      aria-label={label}
      aria-keyshortcuts={keys}
      aria-pressed={pressed}
      aria-expanded={expanded}
      aria-controls={controls}
      disabled={disabled}
      onClick={onClick}
    >
      {keys !== undefined && <kbd>{keys}</kbd>}
      <span>{label}</span>
    </button>
  );
}

/**
 * MenuBoard is the battle's menu: the settings as plates that step through their values, the keys and touch, then
 * surrender and leaving. It stays in the document while closed (hidden) so its settings keep their state.
 */
function MenuBoard({ p }: { p: Props }) {
  const open = p.menuOpen.value,
    board = useRef<HTMLElement>(null),
    sheet = useRef<HTMLFieldSetElement>(null),
    lastFocus = useRef<HTMLElement | null>(null),
    opener = useRef<Element | null>(null),
    heldFocus = useRef(false),
    [help, setHelp] = useState(false),
    side = p.battle.side;
  // Opened: the first plate takes the focus; closed: it goes back to what had it.
  useLayoutEffect(() => {
    if (open) {
      opener.current = document.activeElement;
      sheet.current?.querySelector<HTMLButtonElement>("button")?.focus();
    } else {
      setHelp(false);
      const back = opener.current;
      opener.current = null;
      if (
        back instanceof HTMLElement &&
        back.isConnected &&
        (board.current?.contains(document.activeElement) === true ||
          document.activeElement === document.body)
      )
        back.focus();
    }
  }, [open]);
  // Saving a setting disables the plates briefly; the focus waits on the board and comes back.
  useLayoutEffect(() => {
    const last = lastFocus.current;
    if (p.settingsBusy) {
      if (
        last?.isConnected &&
        sheet.current?.contains(last) &&
        (document.activeElement === last ||
          document.activeElement === document.body)
      ) {
        board.current?.focus();
        heldFocus.current = true;
      }
    } else if (heldFocus.current) {
      if (
        open &&
        document.activeElement === board.current &&
        last?.isConnected &&
        !last.matches(":disabled")
      )
        last.focus();
      heldFocus.current = false;
    }
  }, [p.settingsBusy]);
  const scale = p.scale.value,
    layout = p.layout ?? "auto";
  return (
    <section
      ref={board}
      id={`battle-menu-${side}`}
      class="battle-settings battle-menu hl-board"
      aria-labelledby={`battle-menu-title-${side}`}
      hidden={!open}
      tabIndex={-1}
      onKeyDown={(event) => {
        if (event.key === "Escape" && open) {
          p.menuOpen.value = false;
          event.stopPropagation();
        }
      }}
    >
      <h2 id={`battle-menu-title-${side}`}>Menu</h2>
      <fieldset
        class="battle-settings-sheet"
        ref={sheet}
        disabled={p.settingsBusy ?? false}
        aria-label="Battle settings"
        onFocusCapture={(event) => {
          if (event.target instanceof HTMLElement)
            lastFocus.current = event.target;
        }}
      >
        <MenuChoice
          label={`UI scale: ${scale * 100}%`}
          onClick={() => {
            const at = scales.indexOf(scale as (typeof scales)[number]);
            p.onScale(scales[(at + 1) % scales.length] ?? 1);
          }}
        />
        {p.onLayout !== undefined && (
          <MenuChoice
            label={`Layout: ${layoutShort[layout] ?? layout}`}
            onClick={() => {
              const at = layoutNames.findIndex(([value]) => value === layout),
                next = layoutNames[(at + 1) % layoutNames.length]?.[0];
              if (next !== undefined) p.onLayout?.(next);
            }}
          />
        )}
        <MenuChoice
          label={`Motion: ${p.reduced.value ? "Reduced" : "Full"}`}
          keys="M"
          onClick={() => {
            p.onReduced(!p.reduced.value);
          }}
        />
        <MenuChoice
          label={`Battle scenes: ${p.scenes.value ? "On" : "Off"}`}
          keys="F9"
          onClick={() => {
            p.onScenes(!p.scenes.value);
          }}
        />
        <MenuChoice
          label={
            p.design === undefined
              ? `Map design: ${p.painted.value ? "Painted" : "Tiles"}`
              : `Map design: ${designNames[p.design.value]}`
          }
          keys="F8"
          onClick={() => {
            p.onDesign();
          }}
        />
        <MenuChoice
          label="Fullscreen"
          keys="F11"
          onClick={() => {
            p.onFullscreen();
          }}
        />
        <MenuChoice
          label="Keys and touch"
          expanded={help}
          controls={`battle-help-${side}`}
          onClick={() => {
            setHelp(!help);
          }}
        />
        {help && <Help id={`battle-help-${side}`} />}
        <div class="battle-menu-end">
          <MenuChoice
            label="Surrender"
            keys="F2"
            onClick={() => {
              p.menuOpen.value = false;
              p.battle.requestSurrender();
            }}
            disabled={p.battle.busy() || play.ended(p.battle.session) !== null}
          />
          <MenuChoice
            label="Leave battle"
            onClick={() => {
              p.menuOpen.value = false;
              p.onLeave();
            }}
          />
        </div>
        {(p.settingsNote || p.settingsBusy) && (
          <p class="battle-reason" role="status">
            {p.settingsNote || "Saving settings…"}
          </p>
        )}
      </fieldset>
    </section>
  );
}

function Help({ id }: { id: string }) {
  return (
    <div id={id} class="battle-help hl-well">
      <p>
        Tap your unit for its orders: plates on the cells round it. Tap a plate,
        or flick from the unit towards it. Click or right-click a unit with a
        mouse. Drag elsewhere to pan; wheel or pinch to zoom.
      </p>
      <p>
        With the orders open, Q W E / A S D (or the keypad's 7 8 9 / 1 2 3) run
        the plate up-left, up, up-right, down-left, down, down-right. WASD pans
        otherwise. Arrows move the cursor; Space/Enter selects, Enter on the
        selected unit opens its orders. Home/End and Page Up/Down look at the
        map ends.
      </p>
      <p>
        F1 ends the phase, I shows information, O the map, F2 surrenders, Escape
        goes back. M: reduced motion; F8: map design; F9: battle scenes; F11:
        fullscreen.
      </p>
    </div>
  );
}

function Confirm({ p }: { p: Props }) {
  const confirm = useRef<HTMLButtonElement>(null),
    dialog = useRef<HTMLElement>(null),
    b = p.battle;
  useEffect(() => {
    if (p.embedded) return;
    const previous = document.activeElement;
    confirm.current?.focus();
    return () => {
      if (previous instanceof HTMLElement && previous.isConnected)
        previous.focus();
    };
  }, []);
  const trap = (event: TargetedKeyboardEvent<HTMLElement>) => {
    if (p.embedded) return;
    if (event.key !== "Tab") return;
    const buttons = dialog.current?.querySelectorAll<HTMLButtonElement>(
      "button:not(:disabled)",
    );
    if (buttons === undefined || buttons.length === 0) return;
    const first = buttons[0],
      last = buttons[buttons.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last?.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first?.focus();
    }
  };
  return (
    <div class="battle-modal-shade">
      <section
        class="battle-confirm hl-board"
        role="alertdialog"
        aria-modal={p.embedded ? undefined : "true"}
        aria-labelledby={`battle-confirm-title-${p.battle.side}`}
        ref={dialog}
        onKeyDown={trap}
      >
        <h2 id={`battle-confirm-title-${p.battle.side}`}>
          {b.surrenderPrompt ? "Give up this battle?" : "End this phase?"}
        </h2>
        <p>
          {b.surrenderPrompt
            ? "The other side will win by surrender."
            : "Confirm your orders, or return to the map."}
        </p>
        <div class="battle-button-row">
          <button
            type="button"
            class="hl-nameplate"
            ref={confirm}
            onClick={() => {
              if (b.surrenderPrompt) b.requestSurrender();
              else void b.requestPhase();
            }}
          >
            {b.surrenderPrompt ? "Surrender (F2)" : "Confirm (F1)"}
          </button>
          <button
            type="button"
            class="hl-nameplate"
            onClick={() => {
              b.cancel();
            }}
          >
            Back to orders
          </button>
        </div>
      </section>
    </div>
  );
}

/** Board is what opens over the map: a steel board with its heading and a plate that closes it. */
function Board({
  kind,
  children,
}: {
  kind: string;
  children: ComponentChildren;
}) {
  return (
    <div class={`battle-board hl-board battle-board-${kind}`}>{children}</div>
  );
}

/** Responsive DOM battle controls; map/scene rendering is owned by the supplied hosts. */
export function BattleUI(p: BattleUIProps) {
  // Keep Pixi host refs stable when text/status signals repaint this component.
  const props = useRef(p);
  props.current = p;
  const onSceneHost = useCallback((element: HTMLDivElement | null) => {
    props.current.onSceneHost(element);
  }, []);
  const onStrategyHost = useCallback((element: HTMLDivElement | null) => {
    props.current.onStrategyHost(element);
  }, []);
  const portrait = usePortrait(),
    embedded = p.embedded === true;
  const revision = p.revision.value,
    b = p.battle,
    s = b.session,
    phase = play.phase(s, b.side),
    ready = play.ready(s, b.side),
    end = play.ended(s),
    scene = p.showScene === false ? null : b.playback.scene(),
    progress =
      b.progress === null
        ? null
        : b.progress.total > 0
          ? Math.min(1, Math.max(0, b.progress.done / b.progress.total))
          : 0,
    orders = play.stats(s).orders[b.side] ?? 0,
    menuOpen = p.menuOpen.value;
  const skipScene = () => {
    if (!b.playback.skipScene()) return;
    b.version++;
    if (b.onChange !== null) b.onChange();
    else p.revision.value++;
  };
  const controlKeys = (event: TargetedKeyboardEvent<HTMLDivElement>) => {
    if (
      [
        " ",
        "Enter",
        "ArrowUp",
        "ArrowDown",
        "ArrowLeft",
        "ArrowRight",
        "Home",
        "End",
        "PageUp",
        "PageDown",
      ].includes(event.key) &&
      event.target instanceof Element &&
      event.target.closest("button,input,select,summary") !== null
    )
      event.stopPropagation();
  };
  const status =
    b.status ||
    (end !== null
      ? "The battle is over."
      : ready
        ? "Waiting for the opponent."
        : `${phaseName(phase)} phase`);
  const orderFigure = `${orders}${s.game.limit > 0 ? `/${s.game.limit}` : ""}`,
    orderWord = orders === 1 && s.game.limit === 0 ? "order" : "orders",
    ordersText = (
      <>
        <span>{orderFigure}</span> <span>{orderWord}</span>
      </>
    );
  // What opens over the map; in a wide window a building's contents stay in the side board.
  const contents = b.contents.box !== null,
    centre =
      scene !== null
        ? "scene"
        : contents
          ? embedded || portrait
            ? "contents"
            : null
          : b.panel === "unit" && b.infoUnit?.side === b.side
            ? "unit"
            : b.panel === "game"
              ? "game"
              : b.panel === "map"
                ? "map"
                : null;
  const close = (label: string) => (
    <div class="battle-board-foot">
      <button
        type="button"
        class={`hl-nameplate ${centre === "scene" || centre === "map" ? "battle-host-close" : "battle-board-close"}`}
        onClick={() => {
          if (centre === "scene") skipScene();
          else b.cancel();
        }}
      >
        {label}
      </button>
    </div>
  );
  const board =
    centre === "scene" && scene !== null ? (
      <Board kind="scene">
        <section class="battle-scene-panel" aria-label="Battle scene">
          <h2>Battle</h2>
          <div class="battle-scene-host hl-well" ref={onSceneHost} />
          {close("Skip battle scene")}
        </section>
      </Board>
    ) : centre === "contents" ? (
      <Board kind="contents">
        <Contents p={p} />
      </Board>
    ) : centre === "unit" && b.infoUnit !== null ? (
      <Board kind="unit">
        <h2>Unit information</h2>
        <UnitCard p={p} unit={b.infoUnit} full />
        {close("Close")}
      </Board>
    ) : centre === "game" ? (
      <Board kind="game">
        <GameInfo p={p} />
        {close("Close")}
      </Board>
    ) : centre === "map" ? (
      <Board kind="map">
        <section class="battle-strategy-panel" aria-label="Strategic map">
          <h2>Strategic map</h2>
          <p class="battle-muted">
            Tap the map to look there. Press Escape to return.
          </p>
          <div class="battle-strategy-host hl-well" ref={onStrategyHost} />
          {close("Close map")}
        </section>
      </Board>
    ) : null;
  const planning = progress !== null && (
    <progress
      class="battle-planning"
      aria-label="Computer planning"
      max={1}
      value={progress}
    />
  );
  const statusLine = (
    <p
      class="battle-status"
      role={embedded ? undefined : "status"}
      aria-live={embedded ? undefined : "polite"}
    >
      {status}
    </p>
  );
  const opponent = (
    <div class="battle-opponent-status battle-muted">
      Opponent: {phaseName(play.phase(s, 1 - b.side))} ·{" "}
      {play.ready(s, 1 - b.side) ? "Ready" : "Giving orders"}
      {ready ? " · Your orders confirmed" : ""}
    </div>
  );
  // Upright, Info opens and closes the orders board; what Info shows in a wide window is its Situation plate.
  const sheet = portrait && !embedded && p.sheetOpen.value && centre === null;
  // What opens over the map, or the menu, takes the orders board's place for good.
  useEffect(() => {
    if ((centre !== null || menuOpen) && p.sheetOpen.value)
      p.sheetOpen.value = false;
  }, [centre, menuOpen]);
  const strip = !embedded && p.readOnly !== true && (
    <OrderStrip
      p={
        portrait
          ? {
              ...p,
              onInfo: () => {
                p.sheetOpen.value = !p.sheetOpen.value;
              },
            }
          : p
      }
      className="battle-order-strip"
      sheet={portrait ? p.sheetOpen.value : undefined}
    />
  );
  return (
    <div
      class="battle-ui"
      data-revision={revision}
      data-panel={b.panel}
      data-scene={scene !== null}
      data-readonly={p.readOnly ?? false}
      data-layout={embedded ? "panel" : portrait ? "upright" : "wide"}
      data-board={centre ?? (menuOpen ? "menu" : sheet ? "orders" : "none")}
      style={{ "--ui-scale": p.scale.value }}
      onKeyDown={controlKeys}
    >
      {!embedded && (
        <>
          <div class="battle-housing hl-plate" aria-hidden="true" />
          <div class="battle-window" aria-hidden="true" />
        </>
      )}
      {!embedded && (
        <header
          class="battle-top battle-bar hud-strip"
          inert={p.readOnly ?? false}
          aria-label="Battle status"
        >
          <Lamp lit={!ready && end === null && !b.busy()} />
          <Plaque
            assets={p.assets}
            owner={play.faction(b.side)}
            uiScale={p.scale.value}
          />
          {portrait ? (
            <>
              <h1 class="battle-hidden">
                {s.level.mapName || "Historyline 1914–1918"}
              </h1>
              <div class="battle-bar-lines">
                {statusLine}
                <p class="battle-summary">{selection(b)}</p>
              </div>
              <span class="battle-order-count hl-amber">{ordersText}</span>
            </>
          ) : (
            <>
              <h1 class="battle-hidden">
                {s.level.mapName || "Historyline 1914–1918"}
              </h1>
              <div class="battle-bar-status hud-status">
                {statusLine}
                <span class="battle-order-count hl-amber">{ordersText}</span>
              </div>
            </>
          )}
          <span class="battle-spring" />
          {!portrait && strip}
          <MenuPlate p={p} />
          {planning}
        </header>
      )}

      {portrait && !embedded ? (
        // Upright, the plate below the map is only the row of the global orders' plates.
        <aside class="battle-dock" aria-label="Battle controls">
          {strip}
        </aside>
      ) : (
        <aside class="battle-dock hl-board" aria-label="Battle controls">
          {embedded && (
            <div class="battle-phase-row">
              {statusLine}
              <span class="battle-order-count hl-amber">{ordersText}</span>
            </div>
          )}
          {opponent}
          <div class="battle-dock-body">
            {contents && !embedded ? <Contents p={p} /> : <Orders p={p} />}
          </div>
        </aside>
      )}

      {sheet && (
        <Board kind="orders">
          {opponent}
          <Orders p={p} />
          <div class="battle-board-foot">
            <button
              type="button"
              class="hl-nameplate"
              onClick={() => {
                p.sheetOpen.value = false;
                b.showGame();
              }}
            >
              Situation
            </button>
            <button
              type="button"
              class="hl-nameplate battle-board-close"
              onClick={() => {
                p.sheetOpen.value = false;
              }}
            >
              Close
            </button>
          </div>
        </Board>
      )}

      {board}

      <MenuBoard p={p} />

      {p.readOnly !== true && <OrderMenu p={p} />}

      {(b.phasePrompt || b.surrenderPrompt) && <Confirm p={p} />}
    </div>
  );
}
