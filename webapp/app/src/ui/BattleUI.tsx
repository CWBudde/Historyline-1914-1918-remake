import { designNames } from "../original/design.ts";
import { level, play, phaseMovement } from "@historyline/engine";
import { useCallback, useEffect, useLayoutEffect, useRef } from "preact/hooks";
import type { TargetedKeyboardEvent } from "preact";
import { spriteID } from "../unitFrame.ts";
import type { BattleUIProps } from "./types.ts";
import type { Action } from "./ring.ts";
import "./battle.css";

type Props = BattleUIProps;
const ownerName = (owner: number): string =>
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
const actionName: Record<Action, string> = {
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

function Flag({ owner }: { owner: number }) {
  return (
    <span
      class={`battle-flag owner-${owner}`}
      role="img"
      aria-label={ownerName(owner)}
    />
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
        <UnitArt p={p} unit={unit} portrait />
        <div>
          <h3>{unit.name}</h3>
          <Flag owner={unit.owner} />
          <UnitArt p={p} unit={unit} />
        </div>
      </div>
      <dl class="battle-values">
        <div>
          <dt>Group</dt>
          <dd>
            {unit.size} / {unit.full}
          </dd>
        </div>
        <div>
          <dt>Experience</dt>
          <dd>
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
          </dd>
        </div>
        <div>
          <dt>Budget</dt>
          <dd>{unit.budget}</dd>
        </div>
        <div>
          <dt>Order</dt>
          <dd>
            {unit.target !== null
              ? `Target ${unit.target.line}, ${unit.target.col}`
              : unit.moved
                ? "Given"
                : "Available"}
          </dd>
        </div>
      </dl>
      {info !== null && (
        <>
          <h4>
            Weapons <small>power / range</small>
          </h4>
          <dl class="battle-values battle-weapons">
            {(
              [
                ["Air", info.air],
                ["Ground", info.ground],
                ["Water", info.water],
              ] as const
            ).map(([name, weapon]) => (
              <div key={name}>
                <dt>{name}</dt>
                <dd>
                  {weapon.power} / {weapon.range}
                </dd>
              </div>
            ))}
          </dl>
          <dl class="battle-values">
            {(
              [
                ["Armour", info.armour],
                ["Speed", info.speed],
                ["Weight", info.weight],
                ["Maximum group", info.group],
              ] as const
            ).map(([name, value]) => (
              <div key={name}>
                <dt>{name}</dt>
                <dd>{value}</dd>
              </div>
            ))}
          </dl>
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
      <dl class="battle-values">
        <div>
          <dt>Ground</dt>
          <dd>{groundName(h.terrain)}</dd>
        </div>
        <div>
          <dt>Features</dt>
          <dd>
            {h.infra.ids.length === 0
              ? "None"
              : h.infra.ids.map(groundName).join(", ")}
          </dd>
        </div>
        {building !== undefined && (
          <div>
            <dt>Owner</dt>
            <dd>
              <Flag owner={building.owner} /> {ownerName(building.owner)}
            </dd>
          </div>
        )}
      </dl>
      {b.selected === null ? (
        <p class="battle-muted">
          Select your unit to see legal moves or targets.
        </p>
      ) : (
        <>
          <p>
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

function Orders({ p }: { p: Props }) {
  const b = p.battle,
    view = b.view(),
    selected =
      b.exiting?.unit ??
      [...view.units.values()].find((u) => u.index === b.selected);
  let cursorUnit = view.units.get(level.coordKey(b.hover ?? b.cursor));
  if (cursorUnit?.head !== null && cursorUnit?.head !== undefined)
    cursorUnit = view.units.get(level.coordKey(cursorUnit.head));
  const own =
      selected?.side === b.side
        ? selected
        : cursorUnit?.side === b.side
          ? cursorUnit
          : null,
    blocked =
      b.busy() ||
      b.phasePrompt ||
      b.surrenderPrompt ||
      play.ended(b.session) !== null,
    orders = b.orders();
  return (
    <>
      <section aria-labelledby={`battle-orders-title-${p.battle.side}`}>
        <h2 id={`battle-orders-title-${p.battle.side}`}>Orders</h2>
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
                disabled={blocked || play.ready(b.session, b.side)}
                onClick={() => {
                  b.click(own.at);
                }}
              >
                Select unit
              </button>
              <button
                type="button"
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
            onClick={() => {
              b.cancel();
            }}
          >
            Cancel selection
          </button>
        )}
      </section>
      <Terrain p={p} />
      <section aria-label="Your queued attack orders">
        <h3>
          Queued attacks <small>{orders.length}</small>
        </h3>
        {orders.length === 0 ? (
          <p class="battle-muted">No attacks queued.</p>
        ) : (
          <ol class="battle-order-list">
            {orders.map((order, index) => (
              <li key={index}>
                {[...view.units.values()].find((u) => u.index === order.unit)
                  ?.name ?? "Unit"}{" "}
                → {order.to.line}, {order.to.col}
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
          onClick={() => {
            b.cancel();
          }}
        >
          {box.producing ? "Back to slots" : "Close"}
        </button>
      </div>
      <dl class="battle-values">
        <div>
          <dt>Resources</dt>
          <dd>{resources.stored}</dd>
        </div>
        <div>
          <dt>Total energy</dt>
          <dd>{resources.total}</dd>
        </div>
      </dl>
      <div class="battle-slots" role="group" aria-label="Seven container slots">
        {Array.from({ length: play.Slots }, (_, slot) => {
          const u = slots[slot]?.unit;
          return (
            <button
              type="button"
              key={slot}
              class={`battle-slot${box.slot === slot ? " selected" : ""}`}
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
          <ol class="battle-production-list" ref={list}>
            {options.map((option, index) => (
              <li key={`${option.type}/${index}`}>
                <button
                  type="button"
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

function GameInfo({ p }: { p: Props }) {
  const b = p.battle,
    s = play.stats(b.session);
  return (
    <section aria-labelledby={`battle-game-title-${p.battle.side}`}>
      <h2 id={`battle-game-title-${p.battle.side}`}>Battle information</h2>
      <dl class="battle-values">
        <div>
          <dt>Turn</dt>
          <dd>{s.round}</dd>
        </div>
        <div>
          <dt>Level</dt>
          <dd>{s.level || String(s.levelNo).padStart(2, "0")}</dd>
        </div>
        <div>
          <dt>Phase</dt>
          <dd>{phaseName(s.mode[b.side] ?? phaseMovement)}</dd>
        </div>
        <div>
          <dt>Score</dt>
          <dd>{s.score}</dd>
        </div>
        <div>
          <dt>Orders</dt>
          <dd>{s.orders[b.side]}</dd>
        </div>
        <div>
          <dt>Limit</dt>
          <dd>{s.limit === 0 ? "No limit" : s.limit}</dd>
        </div>
      </dl>
      <div class="battle-holdings">
        {[0, 1, 2].map((side) => {
          const h = side === 2 ? s.map : s.sides[side];
          if (h === undefined) return null;
          return (
            <section key={side}>
              <h3>
                {side === 2 ? "Whole map" : ownerName(play.faction(side))}
              </h3>
              <dl class="battle-values">
                <div>
                  <dt>Units</dt>
                  <dd>{h.units}</dd>
                </div>
                <div>
                  <dt>Depots</dt>
                  <dd>{h.depots}</dd>
                </div>
                <div>
                  <dt>Factories</dt>
                  <dd>{h.factories}</dd>
                </div>
                {side < 2 && (
                  <>
                    <div>
                      <dt>Energy</dt>
                      <dd>{s.energy[side]}</dd>
                    </div>
                    <div>
                      <dt>Income</dt>
                      <dd>{s.income[side]}</dd>
                    </div>
                  </>
                )}
              </dl>
            </section>
          );
        })}
      </div>
    </section>
  );
}

function Settings({ p }: { p: Props }) {
  const details = useRef<HTMLDetailsElement>(null),
    summary = useRef<HTMLElement>(null),
    sheet = useRef<HTMLFieldSetElement>(null),
    lastFocus = useRef<HTMLElement | null>(null),
    heldFocus = useRef(false);
  useLayoutEffect(() => {
    const last = lastFocus.current;
    if (p.settingsBusy) {
      if (
        last?.isConnected &&
        sheet.current?.contains(last) &&
        (document.activeElement === last ||
          document.activeElement === document.body)
      ) {
        summary.current?.focus();
        heldFocus.current = true;
      }
    } else if (heldFocus.current) {
      if (
        details.current?.open &&
        document.activeElement === summary.current &&
        last?.isConnected &&
        !last.matches(":disabled")
      )
        last.focus();
      heldFocus.current = false;
    }
  }, [p.settingsBusy]);
  return (
    <details
      ref={details}
      class="battle-settings"
      onKeyDown={(event) => {
        if (event.key === "Escape" && event.currentTarget.open) {
          event.currentTarget.open = false;
          event.stopPropagation();
        }
      }}
    >
      <summary ref={summary}>Settings</summary>
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
        <label>
          UI scale
          <select
            value={p.scale.value}
            onChange={(event) => {
              p.onScale(Number(event.currentTarget.value));
            }}
          >
            {[0.75, 1, 1.25, 1.5, 2].map((value) => (
              <option key={value} value={value}>
                {value * 100}%
              </option>
            ))}
          </select>
        </label>
        {p.onLayout !== undefined && (
          <label>
            Layout
            <select
              value={p.layout ?? "responsive"}
              onChange={(event) => {
                p.onLayout?.(
                  event.currentTarget.value === "original"
                    ? "original"
                    : "responsive",
                );
              }}
            >
              <option value="responsive">Responsive</option>
              <option value="original">Original split screen</option>
            </select>
          </label>
        )}
        <label>
          <input
            type="checkbox"
            checked={p.reduced.value}
            onChange={(event) => {
              p.onReduced(event.currentTarget.checked);
            }}
          />
          Reduced motion (M)
        </label>
        <label>
          <input
            type="checkbox"
            checked={p.scenes.value}
            onChange={(event) => {
              p.onScenes(event.currentTarget.checked);
            }}
          />
          Battle scenes (F9)
        </label>
        <button
          type="button"
          onClick={() => {
            p.onDesign();
          }}
        >
          {p.design === undefined
            ? `${p.painted.value ? "Show terrain tiles" : "Show painting"} (F8)`
            : `Map design: ${designNames[p.design.value]} (F8)`}
        </button>
        <button
          type="button"
          onClick={() => {
            p.onFullscreen();
          }}
        >
          Fullscreen (F11)
        </button>
        <button
          type="button"
          onClick={() => {
            p.battle.requestSurrender();
          }}
          disabled={p.battle.busy() || play.ended(p.battle.session) !== null}
        >
          Surrender (F2)
        </button>
        <button
          type="button"
          onClick={() => {
            p.onLeave();
          }}
        >
          Leave battle
        </button>
        {(p.settingsNote || p.settingsBusy) && (
          <p class="battle-reason" role="status">
            {p.settingsNote || "Saving settings…"}
          </p>
        )}
      </fieldset>
    </details>
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
        class="battle-confirm"
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
  const revision = p.revision.value,
    b = p.battle,
    s = b.session,
    phase = play.phase(s, b.side),
    ready = play.ready(s, b.side),
    end = play.ended(s),
    scene = p.showScene === false ? null : b.playback.scene(),
    ring = p.ring.value,
    progress =
      b.progress === null
        ? null
        : b.progress.total > 0
          ? Math.min(1, Math.max(0, b.progress.done / b.progress.total))
          : 0,
    blocked = b.playback.busy() || end !== null;
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
  return (
    <div
      class="battle-ui"
      data-revision={revision}
      data-panel={b.panel}
      data-scene={scene !== null}
      data-readonly={p.readOnly ?? false}
      style={{ "--ui-scale": p.scale.value }}
      onKeyDown={controlKeys}
    >
      <header class="battle-top" inert={p.readOnly ?? false}>
        <div class="battle-title">
          <h1>{s.level.mapName || "Historyline 1914–1918"}</h1>
          <span>
            <Flag owner={play.faction(b.side)} /> Turn {play.turn(s)} ·{" "}
            {phaseName(phase)} · {play.energy(s, b.side)} energy
          </span>
        </div>
        <button
          type="button"
          class="battle-ready"
          disabled={blocked}
          aria-pressed={ready}
          onClick={() => {
            void b.requestPhase();
          }}
        >
          {ready ? "Reopen orders" : "Ready"} <kbd>F1</kbd>
        </button>
        <Settings p={p} />
        {progress !== null && (
          <progress
            class="battle-planning"
            aria-label="Computer planning"
            max={1}
            value={progress}
          />
        )}
      </header>

      <aside class="battle-dock" aria-label="Battle controls">
        <p class="battle-status" role="status" aria-live="polite">
          {b.status ||
            (end !== null
              ? "The battle is over."
              : ready
                ? "Waiting for the opponent."
                : `${phaseName(phase)} phase`)}
        </p>
        <div class="battle-phase-row">
          <span>
            Your orders: {play.stats(s).orders[b.side]}
            {s.game.limit > 0 ? ` / ${s.game.limit}` : ""}
          </span>
          <span>{ready ? "Confirmed" : "Open"}</span>
        </div>
        <div class="battle-opponent-status">
          Opponent: {phaseName(play.phase(s, 1 - b.side))} ·{" "}
          {play.ready(s, 1 - b.side) ? "Ready" : "Giving orders"}
        </div>
        <nav
          class="battle-panel-nav"
          aria-label="Battle views"
          inert={p.readOnly ?? false}
        >
          {(b.panel !== "orders" || b.contents.box !== null) && (
            <button
              type="button"
              onClick={() => {
                b.cancel();
              }}
            >
              Back
            </button>
          )}
          <button
            type="button"
            disabled={b.busy() || b.phasePrompt || b.surrenderPrompt}
            onClick={() => {
              b.showInfo();
            }}
          >
            Info
          </button>
          <button
            type="button"
            disabled={b.busy() || b.phasePrompt || b.surrenderPrompt}
            onClick={() => {
              b.run("map");
            }}
          >
            Map
          </button>
        </nav>
        {scene !== null ? (
          <section class="battle-scene-panel" aria-label="Battle scene">
            <h2>Battle</h2>
            <div class="battle-scene-host" ref={onSceneHost} />
            <button type="button" class="battle-host-close" onClick={skipScene}>
              Skip battle scene
            </button>
          </section>
        ) : b.contents.box !== null ? (
          <Contents p={p} />
        ) : b.panel === "unit" && b.infoUnit?.side === b.side ? (
          <>
            <h2>Unit information</h2>
            <UnitCard p={p} unit={b.infoUnit} full />
          </>
        ) : b.panel === "game" ? (
          <GameInfo p={p} />
        ) : b.panel === "map" ? (
          <section class="battle-strategy-panel" aria-label="Strategic map">
            <h2>Strategic map</h2>
            <p>Tap the map to look there. Press Escape to return.</p>
            <div class="battle-strategy-host" ref={onStrategyHost} />
            <button
              type="button"
              class="battle-host-close"
              onClick={() => {
                b.cancel();
              }}
            >
              Close map
            </button>
          </section>
        ) : (
          <Orders p={p} />
        )}
        <details class="battle-help">
          <summary>Keys and touch</summary>
          <p>
            Tap to select, move or attack. Drag to pan; wheel or pinch to zoom.
            Right-click or hold a touch for actions.
          </p>
          <p>
            WASD pans. Arrows move the cursor; Space/Enter selects. Hold Space
            and choose a direction for an action. Home/End and Page Up/Down look
            at the map ends.
          </p>
          <p>
            F1 ends the phase, F2 surrenders, Escape goes back. M: reduced
            motion; F8: map design; F9: battle scenes; F11: fullscreen.
          </p>
        </details>
      </aside>

      {ring !== null && (
        <div class="battle-ring" role="group" aria-label="Actions">
          {ring.items.map((item) => (
            <button
              type="button"
              key={item.action}
              class="battle-ring-action"
              style={{
                left: `${item.x}px`,
                top: `${item.y}px`,
                width: `${96 * ring.scale}px`,
                height: `${80 * ring.scale}px`,
              }}
              onClick={() => {
                p.onRingAction(item.action);
              }}
            >
              {actionName[item.action]}
            </button>
          ))}
          <button
            type="button"
            class="battle-ring-action battle-ring-cancel"
            style={{
              left: `${ring.x}px`,
              top: `${ring.y}px`,
              width: `${96 * ring.scale}px`,
              height: `${80 * ring.scale}px`,
            }}
            onClick={() => {
              p.onCloseRing();
            }}
          >
            Cancel
          </button>
        </div>
      )}
      {(b.phasePrompt || b.surrenderPrompt) && <Confirm p={p} />}
    </div>
  );
}
