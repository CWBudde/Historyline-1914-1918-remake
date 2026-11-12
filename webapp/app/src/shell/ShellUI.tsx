import type { Signal } from "@preact/signals";
import type { ComponentChildren } from "preact";
import { useEffect, useLayoutEffect, useRef, useState } from "preact/hooks";
import { paintingScale, type UiAssets } from "../ui/assets.ts";
import { makeInert } from "../ui/inert.ts";
import { LayoutOptions, layoutValue } from "../ui/layout.tsx";
import type { UnitCard } from "./campaign.ts";
import {
  saveSlots,
  type ShellController,
  type ShellPage,
} from "./controller.ts";
import "../ui/armour.css";
import "./shell.css";
import { OriginalPage } from "./OriginalPage.tsx";
import { design, designNames, nextDesign } from "../original/design.ts";
import { hasOriginalArt } from "../original/state.ts";
import { embrasureHole, shellUnit, useArmour } from "./materials.ts";
import "./fonts.css";

export interface ShellUIProps {
  shell: ShellController;
  assets: UiAssets;
  revision: Signal<number>;
  onFullscreen: () => void;
  isFullscreen: boolean;
  /** The offline maps page's content: download, progress and what is kept (src/offline). */
  offlinePanel?: () => ComponentChildren;
}
const titles: Record<ShellPage, string> = {
  main: "HISTORYLINE",
  code: "MAP CODE",
  values: "GAME VALUES",
  control: "CONTROL",
  extras: "EXTRAS",
  country: "SELECT COUNTRY",
  news: "THE WAR CHRONICLE",
  units: "NEW AT THE FRONT",
  loading: "LOADING",
  battle: "",
  leave: "LEAVE THE BATTLE?",
  outcome: "",
  scores: "SCORE LIST",
  saves: "SAVED GAMES",
  original: "ORIGINAL GAME",
  offline: "OFFLINE MAPS",
  closed: "HISTORYLINE 1914–1918",
};
/** Help is the two lines the column shows for the focused control, as the desktop game's menus do. */
type Help = readonly [string, string?];
/** The pages whose controls stand in the steel column at the left, over the painting. */
const columnPages: readonly ShellPage[] = [
  "main",
  "code",
  "values",
  "control",
  "extras",
  "scores",
  "loading",
  "outcome",
  "closed",
];
/** The wide pages on a riveted steel board. */
const boards: readonly ShellPage[] = ["saves", "original", "offline", "leave"];
/** The pages with the two flags over the painting. */
const flagPages: readonly ShellPage[] = ["main", "values", "control", "extras"];
const maxImportBytes = 4 * Math.ceil((8 * 1024 * 1024) / 3) + 2048;
const controls =
  'button:not(:disabled), input:not(:disabled):not([type="hidden"]), select:not(:disabled), [tabindex="0"]';
/**
 * pageScene is the painting seen through the embrasure of a columned page. The screen/* mock-ups of the whole
 * interface stay out: the desktop game draws only the scene (cmd/historyline/app.go).
 */
function pageScene(s: ShellController): string {
  switch (s.page) {
    case "main":
    case "code":
    case "closed":
      return "backdrop/western-front";
    case "values":
    case "control":
      return "scene/options";
    case "extras":
    case "scores":
      return "scene/save";
    case "loading":
      return "scene/mission-selection";
    case "outcome":
      return s.result?.won ? "scene/victory" : "scene/defeat";
    default:
      return "";
  }
}
function Art({
  p,
  id,
  alt = "",
  className = "",
  cover = false,
  drawn = false,
  scale,
  lite = false,
}: {
  p: ShellUIProps;
  id: string;
  alt?: string;
  className?: string;
  /** The painting covers the window: its art is chosen for the drawn size. */
  cover?: boolean;
  /** The art is drawn at the layout's unit (shell.css --u) rather than its own size. */
  drawn?: boolean;
  /** The drawn size over the logical size, where the page knows it. */
  scale?: number;
  /** Only the lite art: the armour's painted parts. */
  lite?: boolean;
}) {
  const ui = p.shell.settings.uiScale,
    image = p.assets.image(
      "screens",
      id,
      scale ??
        (cover
          ? paintingScale(window.innerWidth, window.innerHeight, ui)
          : drawn
            ? shellUnit(window.innerWidth, window.innerHeight, ui)
            : ui),
      cover || scale !== undefined,
      lite,
    ).value;
  return image.url === null ? (
    <span
      class={`shell-art shell-art-empty ${className}`}
      role={alt ? "img" : undefined}
      aria-label={
        alt ? `${alt}${image.error ? " (image unavailable)" : ""}` : undefined
      }
      title={alt ? (image.error ?? undefined) : undefined}
    />
  ) : (
    <img
      class={`shell-art ${className}`}
      src={image.url}
      alt={alt}
      draggable={false}
    />
  );
}
/**
 * enlarge is how far the embrasure's art may be drawn above its size before the next density is fetched: on a small
 * window the 1x art, a fifth enlarged, keeps the first playable frame within its transfer budget.
 */
const enlarge = 1.2;
/** The units of the embrasure's width in landscape; upright it spans the column. */
const embrasureWidth = 470;
/**
 * Embrasure is the heavy frame cut into the plate with its shutter swung up out of sight, the page's painting seen through
 * its opening. The painting is chosen for the opening's drawn width; it reaches under the frame's lip.
 */
function Embrasure({ p, scene }: { p: ShellUIProps; scene: string }) {
  const w = window.innerWidth,
    h = window.innerHeight,
    u = shellUnit(w, h, p.shell.settings.uiScale),
    frame = w >= h ? embrasureWidth * u : 0.86 * (w - 48 * u),
    hole = frame * (1 - embrasureHole.left - embrasureHole.right);
  return (
    <div class="shell-embrasure" aria-hidden="true">
      <div class="shell-embrasure-view">
        {scene && (
          <Art
            p={p}
            id={scene}
            className="shell-scene"
            scale={hole / 640 / enlarge}
            lite
          />
        )}
      </div>
      <Art
        p={p}
        id="armour/embrasure"
        className="shell-embrasure-frame"
        scale={frame / 747 / enlarge}
        lite
      />
    </div>
  );
}
/** The two enamel plaques of the sides. */
function Plaques({ p }: { p: ShellUIProps }) {
  const u = shellUnit(
    window.innerWidth,
    window.innerHeight,
    p.shell.settings.uiScale,
  );
  return (
    <div class="shell-plaques">
      <Art
        p={p}
        id="armour/plaque-germany"
        alt="Germany"
        className="shell-plaque"
        scale={(56 * u) / 169}
        lite
      />
      <Art
        p={p}
        id="armour/plaque-france"
        alt="France"
        className="shell-plaque"
        scale={(56 * u) / 164}
        lite
      />
    </div>
  );
}

function Unit({ p, card }: { p: ShellUIProps; card: UnitCard }) {
  return (
    <article class="shell-unit shell-well hl-well">
      <Art
        p={p}
        id={`portrait/${card.family}`}
        alt={`${card.name} portrait`}
        className="shell-portrait"
        drawn
      />
      <div class="shell-unit-values">
        <h2>{card.name}</h2>
        <p class="shell-muted">{card.country}</p>
        <p>
          Armour {card.armour} · Speed {card.speed}
        </p>
        <p>{card.attack}</p>
        {card.reach && <p>{card.reach}</p>}
      </div>
      <p class="shell-unit-note">{card.note}</p>
    </article>
  );
}

/** Semantic controller changes and progressive art repaint only the DOM shell. */
export function ShellUI(p: ShellUIProps) {
  const revision = p.revision.value;
  const s = p.shell,
    page = s.page;
  const root = useRef<HTMLDivElement>(null);
  const lastFocused = useRef<HTMLElement | null>(null);
  const wasBusy = useRef(false);
  const [fileError, setFileError] = useState("");
  const [reading, setReading] = useState(false);
  const [code, setCode] = useState("");
  const [help, showHelp] = useState<Help | null>(null);
  /** The notice the help strip no longer shows: the player moved on to another control. */
  const [dismissed, setDismissed] = useState("");
  const noticeSettled = useRef(false);
  const materials = useArmour(p.assets, s.settings.uiScale, page === "news");
  const busy = s.busy || reading;
  const notice = fileError || s.note || (busy ? "Please wait…" : "");
  /**
   * A notice takes the help strip until the pointer or the focus goes to another control. The focus a page sets on
   * arrival comes right after the notice and does not count.
   */
  const lastNotice = useRef(notice);
  if (lastNotice.current !== notice) {
    lastNotice.current = notice;
    noticeSettled.current = false;
  }
  useEffect(() => {
    const timer = setTimeout(() => {
      noticeSettled.current = true;
    }, 300);
    return () => {
      clearTimeout(timer);
    };
  }, [notice]);
  const setHelp = (lines: Help | null) => {
    showHelp(lines);
    if (lines !== null && noticeSettled.current) setDismissed(notice);
  };
  /** Focus and pointer show a control's help in the column. */
  const helpOf = (lines: Help | undefined) =>
    lines === undefined
      ? {}
      : {
          onFocus: () => {
            setHelp(lines);
          },
          onMouseEnter: () => {
            setHelp(lines);
          },
        };
  /** fieldHelp is helpOf for a label around its control. */
  const fieldHelp = (lines: Help) => ({
    onFocusCapture: () => {
      setHelp(lines);
    },
    onMouseEnter: () => {
      setHelp(lines);
    },
  });
  const run = (action: () => void | Promise<void>) => {
    setFileError("");
    try {
      const pending = action();
      if (pending !== undefined)
        void pending.catch((error: unknown) => {
          setFileError(error instanceof Error ? error.message : String(error));
        });
    } catch (error) {
      setFileError(error instanceof Error ? error.message : String(error));
    }
  };
  const button = (
    label: string,
    action: () => void | Promise<void>,
    disabled = busy,
    lines?: Help,
  ) => (
    <button
      type="button"
      class="shell-button hl-nameplate"
      disabled={disabled}
      onClick={() => {
        run(action);
      }}
      {...helpOf(lines)}
    >
      {label}
    </button>
  );
  const back = (lines: Help = ["Back to the previous page."]) =>
    button(
      "Back",
      () => {
        s.back();
      },
      busy,
      lines,
    );

  useLayoutEffect(() => {
    // IndexedDB operations disable controls briefly. Restore the same control
    // when Chromium has moved focus to the document during that operation.
    const previous = lastFocused.current;
    if (
      wasBusy.current &&
      !busy &&
      document.activeElement === document.body &&
      previous?.isConnected &&
      root.current?.contains(previous) &&
      !previous.matches(":disabled")
    )
      previous.focus();
    wasBusy.current = busy;
  }, [busy]);

  useLayoutEffect(() => {
    setFileError("");
    setHelp(null);
    if (page === "code") setCode("");
    const node = root.current;
    if (!node || page === "battle") return;
    const preferred =
      page === "country"
        ? node.querySelector<HTMLElement>(`[data-side="${s.side}"]`)
        : page === "code"
          ? node.querySelector<HTMLElement>("input")
          : page === "leave"
            ? node.querySelector<HTMLElement>(".shell-actions button")
            : null;
    (preferred ?? node.querySelector<HTMLElement>(controls) ?? node).focus();
    node.scrollTop = 0;
  }, [page, s]);

  useEffect(() => {
    const node = root.current;
    if (!node || page !== "leave") return;
    const savedFocus = document.activeElement;
    const release: (() => void)[] = [];
    // Preserve existing inert states on the canvas and other document controls.
    for (
      let branch: HTMLElement = node;
      branch.parentElement;
      branch = branch.parentElement
    ) {
      for (const sibling of branch.parentElement.children) {
        if (sibling !== branch && sibling instanceof HTMLElement) {
          release.push(makeInert(sibling));
        }
      }
      if (branch.parentElement === document.body) break;
    }
    const retainFocus = (event: FocusEvent) => {
      if (event.target instanceof Node && !node.contains(event.target))
        (node.querySelector<HTMLElement>(controls) ?? node).focus();
    };
    document.addEventListener("focusin", retainFocus);
    return () => {
      document.removeEventListener("focusin", retainFocus);
      for (const unlock of release) unlock();
      if (
        savedFocus instanceof HTMLElement &&
        savedFocus.isConnected &&
        !savedFocus.closest("[inert]")
      )
        savedFocus.focus();
    };
  }, [page]);

  useEffect(() => {
    const node = root.current;
    if (!node || page === "battle") return;
    const key = (event: KeyboardEvent) => {
      if (event.ctrlKey || event.altKey || event.metaKey || event.isComposing)
        return;
      event.stopPropagation();
      const target = event.target;
      const editable =
        target instanceof HTMLElement &&
        (target.matches("input, textarea, select") || target.isContentEditable);
      if (event.key === "Escape") {
        event.preventDefault();
        if (!s.busy && !reading) s.back();
        return;
      }
      if (event.key === "F11") {
        event.preventDefault();
        p.onFullscreen();
        return;
      }
      const list = Array.from(node.querySelectorAll<HTMLElement>(controls));
      const active = list.findIndex((item) => item === document.activeElement);
      if (event.key === "Tab" && page === "leave") {
        if (list.length === 0) {
          event.preventDefault();
          node.focus();
        } else if (event.shiftKey && active <= 0) {
          event.preventDefault();
          list.at(-1)?.focus();
        } else if (
          !event.shiftKey &&
          (active === list.length - 1 || active < 0)
        ) {
          event.preventDefault();
          list[0]?.focus();
        }
        return;
      }
      if (editable) return;
      if (
        ["ArrowDown", "ArrowRight", "ArrowUp", "ArrowLeft"].includes(event.key)
      ) {
        event.preventDefault();
        const direction =
          event.key === "ArrowUp" || event.key === "ArrowLeft" ? -1 : 1;
        list[(active + direction + list.length) % list.length]?.focus();
      } else if (
        (event.key === "Enter" || event.key === " ") &&
        !(target instanceof HTMLButtonElement)
      ) {
        event.preventDefault();
        list[0]?.click();
      }
    };
    document.addEventListener("keydown", key, true);
    return () => {
      document.removeEventListener("keydown", key, true);
    };
  }, [page, s, reading, p.onFullscreen]);

  if (page === "battle") return null;
  const u = shellUnit(
    window.innerWidth,
    window.innerHeight,
    s.settings.uiScale,
  );
  const scene = pageScene(s);
  const wide = !columnPages.includes(page);
  const title =
    page === "outcome"
      ? s.result?.draw
        ? "DRAW"
        : s.result?.won
          ? "VICTORY"
          : "DEFEAT"
      : titles[page];
  const change = (patch: Parameters<ShellController["updateSettings"]>[0]) =>
    s.updateSettings(patch);
  const importFile = async (slot: string, file: File | undefined) => {
    if (!file || busy) return;
    setFileError("");
    setReading(true);
    try {
      if (file.size > maxImportBytes) throw new Error("Save file is too large");
      if (file.size === 0) throw new Error("Save file is empty");
      await s.importSlot(slot, await file.text());
    } catch (error) {
      setFileError(error instanceof Error ? error.message : String(error));
    } finally {
      setReading(false);
    }
  };
  const exportFile = async (slot: string) => {
    const text = await s.exportSlot(slot);
    if (text === null) return;
    const url = URL.createObjectURL(
      new Blob([text], { type: "application/json" }),
    );
    const link = document.createElement("a");
    link.href = url;
    link.download = `historyline-slot-${slot}.json`;
    document.body.append(link);
    link.click();
    link.remove();
    setTimeout(() => {
      URL.revokeObjectURL(url);
    }, 0);
  };
  let content: ComponentChildren;
  switch (page) {
    case "main":
      content = (
        <>
          <nav class="shell-menu" aria-label="Main menu">
            {button(
              "Start",
              () => {
                s.start();
              },
              busy,
              [
                "Begin the battle for the map",
                `${s.mission.name}, ${s.mission.news?.label ?? s.mission.month}.`,
              ],
            )}
            {button(
              `Code  ${s.mission.code}`,
              () => {
                s.open("code");
              },
              busy,
              ["Type the code of a map you have reached, then press Enter."],
            )}
            {button(
              "Game values",
              () => {
                s.open("values");
              },
              busy,
              ["Battles, players and the move limit."],
            )}
            {button(
              "Control",
              () => {
                s.open("control");
              },
              busy,
              ["Layout, screen, scrolling and the map's design."],
            )}
            {button(
              "Extras",
              () => {
                s.open("extras");
              },
              busy,
              ["Score list, saved games and the original game's maps."],
            )}
            {button(
              "Exit",
              () => {
                s.open("closed");
              },
              busy,
              ["Leave Historyline 1914–1918."],
            )}
          </nav>
        </>
      );
      break;
    case "code":
      content = (
        <>
          <form
            class="shell-menu"
            onSubmit={(event) => {
              event.preventDefault();
              run(() => {
                s.submitCode(code);
              });
            }}
          >
            <label class="shell-field shell-code hl-nameplate">
              <span>Map code</span>
              <input
                value={code}
                maxLength={12}
                autoComplete="off"
                autoCapitalize="characters"
                spellcheck={false}
                {...helpOf([
                  "Type the code of a map you have reached, then press Enter.",
                ])}
                onInput={(event) => {
                  const value = event.currentTarget.value
                    .toUpperCase()
                    .replace(/[^A-Z0-9]/g, "")
                    .slice(0, 12);
                  setCode(value);
                  s.typed = value;
                }}
              />
            </label>
            <button
              type="submit"
              class="shell-button hl-nameplate"
              disabled={busy || !code}
              {...helpOf(["Open the map of this code."])}
            >
              Confirm code
            </button>
          </form>
          <footer class="shell-actions">
            {back(["Back to the main menu."])}
          </footer>
        </>
      );
      break;
    case "values":
      content = (
        <>
          <div class="shell-menu shell-menu-tight">
            <p class="shell-fixed hl-nameplate">Rules · Original (DOS)</p>
            {button(
              `Battles · ${s.settings.reduced ? "Short" : "Full"}`,
              () => change({ reduced: !s.settings.reduced }),
              busy,
              [
                "Full: units walk and fire in full.",
                "Short: moves and fire are instant.",
              ],
            )}
            <label
              class="shell-field hl-nameplate"
              {...fieldHelp([
                "Two humans: the split screen, B on D/C/X/V + Ctrl.",
                "Hot-seat: the sides take turns.",
              ])}
            >
              <span>Play mode</span>
              <select
                value={s.settings.playMode}
                disabled={busy}
                onChange={(event) => {
                  const playMode = event.currentTarget.value;
                  if (
                    playMode === "single" ||
                    playMode === "simultaneous" ||
                    playMode === "hotseat"
                  )
                    run(() => change({ playMode }));
                }}
              >
                <option value="single">Single player</option>
                <option value="simultaneous">Two humans</option>
                <option value="hotseat">Hot-seat</option>
              </select>
            </label>
            <p class="shell-fixed hl-nameplate">Player A · Human</p>
            {s.settings.playMode === "single" ? (
              button(
                `Player B · ${s.settings.opponent === "computer" ? "Computer" : "Passive"}`,
                () =>
                  change({
                    opponent:
                      s.settings.opponent === "computer"
                        ? "passive"
                        : "computer",
                  }),
                busy,
                [
                  "Computer: the computer commands.",
                  "Passive: it confirms without orders.",
                ],
              )
            ) : (
              <p class="shell-fixed hl-nameplate">Player B · Human</p>
            )}
            <p class="shell-fixed hl-nameplate">Depots · Own only</p>
            {button(
              `Moves · ${s.settings.limit === 0 ? "No limit" : `${s.settings.limit} a phase`}`,
              () => {
                const limits = [0, 5, 10, 15, 20] as const;
                return change({
                  limit:
                    limits[
                      (limits.indexOf(s.settings.limit) + 1) % limits.length
                    ] ?? 0,
                });
              },
              busy,
              [
                "Orders each side may give a phase.",
                "Fewer orders score higher.",
              ],
            )}
          </div>
          <footer class="shell-actions">
            {back(["Back to the main menu."])}
          </footer>
        </>
      );
      break;
    case "control":
      content = (
        <>
          <div class="shell-menu shell-menu-tight">
            <label
              class="shell-field hl-nameplate"
              {...fieldHelp([
                "The battle screen: the split screen of the original or one map.",
              ])}
            >
              <span>Layout</span>
              <select
                value={s.settings.layout}
                disabled={busy}
                onChange={(event) => {
                  const layout = layoutValue(event.currentTarget.value);
                  if (layout !== null) run(() => change({ layout }));
                }}
              >
                <LayoutOptions />
              </select>
            </label>
            {button(
              `Fullscreen · ${p.isFullscreen ? "On" : "Off"}`,
              p.onFullscreen,
              busy,
              ["Window or full screen (F11)."],
            )}
            {button(
              `Scrolling · ${s.settings.scroll}`,
              () =>
                change({
                  scroll:
                    s.settings.scroll === "normal"
                      ? "fast"
                      : s.settings.scroll === "fast"
                        ? "slow"
                        : "normal",
                }),
              busy,
              ["How fast the arrow keys move the map."],
            )}
            {button(
              `Motion · ${s.settings.reduced ? "Reduced" : "Full"}`,
              () => change({ reduced: !s.settings.reduced }),
              busy,
              ["Full or reduced animation."],
            )}
            {button(
              `Battle scenes · ${s.settings.scenes ? "On" : "Off"}`,
              () => change({ scenes: !s.settings.scenes }),
              busy,
              ["Show the battle scenes between units, or skip them."],
            )}
            {(() => {
              // Painted where a month has a painting; Original once the original game's files are imported.
              const a = {
                painted: true,
                original: hasOriginalArt(),
                originalMap: false,
              };
              return button(
                `Map design · ${designNames[design(s.settings, a)]}`,
                () => change(nextDesign(s.settings, a)),
                busy,
                [
                  "The painted map, its tiles, or the original game's art once imported.",
                ],
              );
            })()}
            <label
              class="shell-field hl-nameplate"
              {...fieldHelp(["The size of the menus and panels."])}
            >
              <span>UI scale</span>
              <select
                value={s.settings.uiScale}
                disabled={busy}
                onChange={(event) => {
                  const scale = Number(event.currentTarget.value);
                  run(() => change({ uiScale: scale }));
                }}
              >
                {[0.75, 1, 1.25, 1.5, 2].map((scale) => (
                  <option key={scale} value={scale}>
                    {scale * 100}%
                  </option>
                ))}
              </select>
            </label>
          </div>
          <footer class="shell-actions">
            {back(["Back to the main menu."])}
          </footer>
        </>
      );
      break;
    case "extras":
      content = (
        <>
          <nav class="shell-menu" aria-label="Extras">
            {button("Score list", () => s.openStorage("scores"), busy, [
              "The best score of each map.",
            ])}
            {button("Load game", () => s.openStorage("saves"), busy, [
              "Load, export or import a saved",
              "battle.",
            ])}
            {s.originals !== null &&
              button(
                "Original game",
                () => {
                  s.open("original");
                },
                busy,
                ["Play the original game's maps from your own copy."],
              )}
            {p.offlinePanel !== undefined &&
              button(
                "Offline maps",
                () => {
                  s.open("offline");
                },
                busy,
                ["Keep maps on this device to play without a connection."],
              )}
          </nav>
          <footer class="shell-actions">
            {back(["Back to the main menu."])}
          </footer>
        </>
      );
      break;
    case "country":
      content = (
        <>
          <div class="shell-countries">
            {([1, 0] as const).map((side) => (
              <button
                key={side}
                type="button"
                class="shell-well shell-country hl-well"
                data-side={side}
                disabled={busy}
                onClick={() => {
                  run(() => s.chooseSide(side));
                }}
              >
                <Art
                  p={p}
                  id={
                    side === 0
                      ? "armour/plaque-germany"
                      : "armour/plaque-france"
                  }
                  className="shell-country-flag"
                  scale={(200 * u) / 169}
                  lite
                />
                <strong class="shell-nameplate hl-nameplate">
                  {side === 0 ? "Germany" : "France"}
                </strong>
                <span>{side === 0 ? "The Central Powers" : "The Entente"}</span>
              </button>
            ))}
          </div>
          <footer class="shell-actions">{back()}</footer>
        </>
      );
      break;
    case "news": {
      const article = s.mission.news;
      content =
        article === null ? (
          button("Continue", () => s.continueNews())
        ) : (
          <div class="shell-paper">
            <header class="shell-paper-header">
              <div class="shell-edition">
                <span>No. {s.mission.issue}</span>
                <span>{article.label}</span>
                <span>Price one penny</span>
              </div>
              <h2>{article.headline}</h2>
              <p class="shell-deck">{article.deck}</p>
            </header>
            <div class="shell-news-layout">
              <div class="shell-news-picture">
                <figure>
                  <Art
                    p={p}
                    id={`news/${s.mission.month}`}
                    alt={article.caption}
                    className="shell-news-photo"
                    drawn
                  />
                  <figcaption>{article.caption}</figcaption>
                </figure>
                <section class="shell-orders">
                  <h3>Orders of the day</h3>
                  <p>{s.mission.briefing}</p>
                </section>
                <footer class="shell-actions">
                  {button("Continue", () => s.continueNews(), busy, [
                    "On to the battle.",
                  ])}
                  {back()}
                </footer>
              </div>
              <div class="shell-news-copy">
                {article.body.map((paragraph, i) => (
                  <p key={i}>{paragraph}</p>
                ))}
                <section class="shell-briefs">
                  <h3>In brief</h3>
                  <ul>
                    {article.briefs.map((brief, i) => (
                      <li key={i}>{brief}</li>
                    ))}
                  </ul>
                </section>
              </div>
            </div>
          </div>
        );
      break;
    }
    case "units":
      content = (
        <>
          <div class="shell-section-heading">
            <p>
              {s.side === 0 ? "The Central Powers" : "The Entente"} ·{" "}
              {s.mission.news?.label ?? s.mission.month}
            </p>
            {s.mission.newUnits.length > 4 && (
              <p>
                Page {s.unitPage + 1} of{" "}
                {Math.ceil(s.mission.newUnits.length / 4)}
              </p>
            )}
          </div>
          <div class="shell-units">
            {s.cards.map((card) => (
              <Unit p={p} card={card} key={card.id} />
            ))}
          </div>
          <footer class="shell-actions">
            {s.unitPage + 1 < Math.ceil(s.mission.newUnits.length / 4) &&
              button("Next page", () => {
                s.nextUnits();
              })}
            {button("To battle", () => s.toBattle())}
            {back()}
          </footer>
        </>
      );
      break;
    case "loading":
      content = (
        <>
          <p class="shell-muted">{s.mission.news?.label ?? s.mission.month}</p>
          <h2>{s.mission.name}</h2>
          <p>{s.mission.briefing}</p>
          <p>The map is being prepared.</p>
          <progress
            class="shell-progress"
            max={1}
            value={s.loadingProgress}
            aria-label="Preparing battle"
          />
          <p role="status">{s.loadingText}</p>
          <footer class="shell-actions">
            {s.note && button("Retry", () => s.toBattle())}
            {button("Back", () => {
              s.cancelLoading();
            })}
          </footer>
        </>
      );
      break;
    case "leave":
      content = (
        <>
          <p>The battle is given up and lost.</p>
          <p>Stay to go on commanding.</p>
          <div class="shell-save-leave">
            {saveSlots.map((slot) =>
              button(
                `Save and leave · Slot ${slot}`,
                () => s.saveAndLeave(slot),
                busy || !s.canSave() || !s.storageAvailable(),
              ),
            )}
          </div>
          <footer class="shell-actions">
            {button("Stay", () => {
              s.stay();
            })}
            {button("Leave", () => {
              s.leave();
            })}
          </footer>
        </>
      );
      break;
    case "outcome":
      content = (
        <>
          <p>{s.result?.text}</p>
          {s.result?.won && s.mission.mode !== "2p" && (
            <p>Score {s.result.score}</p>
          )}
          {s.result?.next ? (
            <p>
              The code of the next map is <strong>{s.result.next}</strong>.
            </p>
          ) : s.result?.won && s.mission.mode !== "2p" ? (
            <p>There is no further map yet.</p>
          ) : null}
          <footer class="shell-actions">
            {button("Continue", () => {
              s.continueOutcome();
            })}
          </footer>
        </>
      );
      break;
    case "scores":
      content = (
        <>
          {s.scores.length === 0 ? (
            <p>No scores are kept yet.</p>
          ) : (
            <table class="shell-scores">
              <thead>
                <tr>
                  <th scope="col">Map</th>
                  <th scope="col">Best score</th>
                </tr>
              </thead>
              <tbody>
                {s.scores.map((entry) => (
                  <tr key={entry.level}>
                    <th scope="row">{entry.level}</th>
                    <td>{entry.score}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <footer class="shell-actions">{back()}</footer>
        </>
      );
      break;
    case "saves":
      content = (
        <>
          <div class="shell-save-slots">
            {saveSlots.map((slot) => {
              const saved = s.slots.find((item) => item.slot === slot);
              return (
                <section
                  key={slot}
                  class="shell-save-slot shell-well hl-well"
                  aria-labelledby={`shell-slot-${slot}`}
                >
                  <h2 id={`shell-slot-${slot}`}>Slot {slot}</h2>
                  <p>{saved?.name ?? "Empty slot"}</p>
                  {saved && (
                    <p class="shell-muted">
                      <time dateTime={new Date(saved.savedAt).toISOString()}>
                        {new Date(saved.savedAt).toLocaleString()}
                      </time>
                    </p>
                  )}
                  <div class="shell-actions">
                    {button(
                      "Load",
                      () => s.loadSlot(slot),
                      busy || !saved || !s.storageAvailable(),
                    )}
                    {button(
                      "Export",
                      () => exportFile(slot),
                      busy || !saved || !s.storageAvailable(),
                    )}
                  </div>
                  <label
                    class={`shell-import hl-nameplate ${busy || !s.storageAvailable() ? "shell-disabled" : ""}`}
                  >
                    <span>{saved ? "Import / replace" : "Import"}</span>
                    <input
                      type="file"
                      accept=".json,application/json"
                      aria-label={`Import${saved ? " and replace" : ""} slot ${slot}`}
                      disabled={busy || !s.storageAvailable()}
                      onChange={(event) => {
                        const input = event.currentTarget,
                          file = input.files?.[0];
                        input.value = "";
                        void importFile(slot, file);
                      }}
                    />
                  </label>
                </section>
              );
            })}
          </div>
          <footer class="shell-actions">{back()}</footer>
        </>
      );
      break;
    case "original":
      content = <OriginalPage shell={s} busy={busy} button={button} />;
      break;
    case "offline":
      content = (
        <>
          {p.offlinePanel?.()}
          <footer class="shell-actions">{back(["Back to extras."])}</footer>
        </>
      );
      break;
    case "closed":
      content = (
        <>
          <p>You have left Historyline.</p>
          {button("Return to the main menu", () => {
            s.back();
          })}
        </>
      );
      break;
  }
  const notices = (
    <>
      {!s.storageAvailable() && ["saves", "scores", "leave"].includes(page) && (
        <p class="shell-notice" role="status">
          Browser storage is unavailable. Saved games and scores cannot be
          stored.
        </p>
      )}
      {notice && wide && (
        <p class="shell-notice" role="status" aria-live="polite">
          {notice}
        </p>
      )}
    </>
  );
  return (
    <div
      ref={root}
      data-revision={revision}
      onFocusCapture={(event) => {
        if (event.target instanceof HTMLElement)
          lastFocused.current = event.target;
      }}
      class={`shell-ui shell-page-${page} ${wide ? "shell-wide" : "shell-columned"}`}
      style={{ "--shell-scale": s.settings.uiScale, ...materials }}
      tabIndex={-1}
      role={page === "leave" ? "dialog" : undefined}
      aria-modal={page === "leave" ? true : undefined}
      aria-labelledby="shell-title"
      aria-busy={busy}
    >
      <div class="shell-rivets" aria-hidden="true" />
      <main
        class={`shell-sheet${boards.includes(page) ? " shell-panel hl-well" : ""}`}
      >
        <header class="shell-heading">
          {page === "main" ? (
            <h1 id="shell-title" class="shell-title">
              <Art
                p={p}
                id="armour/title"
                alt="Historyline 1914–1918"
                scale={(252 * u) / 712}
                lite
              />
            </h1>
          ) : (
            <h1 id="shell-title">{title}</h1>
          )}
        </header>
        {flagPages.includes(page) && <Plaques p={p} />}
        {wide ? (
          <>
            {content}
            {notices}
          </>
        ) : (
          <>
            <div class="shell-column">
              {content}
              {notices}
            </div>
            <Embrasure p={p} scene={scene} />
            <aside class="shell-help hl-nameplate">
              {notice && notice !== dismissed ? (
                <p class="shell-help-notice" role="status" aria-live="polite">
                  {notice}
                </p>
              ) : (
                help !== null && (
                  <>
                    <p>{help[0]}</p>
                    {help[1] !== undefined && <p>{help[1]}</p>}
                  </>
                )
              )}
            </aside>
          </>
        )}
      </main>
    </div>
  );
}
