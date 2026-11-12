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
import "./shell.css";
import { OriginalPage } from "./OriginalPage.tsx";
import { design, designNames, nextDesign } from "../original/design.ts";
import { hasOriginalArt } from "../original/state.ts";

export interface ShellUIProps {
  shell: ShellController;
  assets: UiAssets;
  revision: Signal<number>;
  onFullscreen: () => void;
  isFullscreen: boolean;
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
  closed: "HISTORYLINE 1914–1918",
};
const maxImportBytes = 4 * Math.ceil((8 * 1024 * 1024) / 3) + 2048;
const controls =
  'button:not(:disabled), input:not(:disabled):not([type="hidden"]), select:not(:disabled), [tabindex="0"]';
/**
 * pageScene is the painting behind a page. The screen/* mock-ups of the whole interface stay out: the desktop game
 * draws only the scene (cmd/historyline/app.go).
 */
function pageScene(s: ShellController): string {
  switch (s.page) {
    case "main":
    case "code":
    case "closed":
    case "news":
      return "backdrop/western-front";
    case "values":
    case "control":
      return "scene/options";
    case "extras":
    case "scores":
    case "saves":
    case "original":
      return "scene/save";
    case "country":
      return "scene/side-selection";
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
}: {
  p: ShellUIProps;
  id: string;
  alt?: string;
  className?: string;
  /** The painting covers the window: its art is chosen for the drawn size. */
  cover?: boolean;
}) {
  const ui = p.shell.settings.uiScale,
    image = p.assets.image(
      "screens",
      id,
      cover ? paintingScale(window.innerWidth, window.innerHeight, ui) : ui,
      cover,
    ).value;
  return image.url === null ? (
    <span
      class={`shell-art shell-art-empty ${className}`}
      role={alt ? "img" : undefined}
      aria-label={
        alt ? `${alt}${image.error ? " (image unavailable)" : ""}` : undefined
      }
      title={image.error ?? undefined}
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
function Unit({ p, card }: { p: ShellUIProps; card: UnitCard }) {
  return (
    <article class="shell-unit">
      <Art
        p={p}
        id={`portrait/${card.family}`}
        alt={`${card.name} portrait`}
        className="shell-portrait"
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
  const busy = s.busy || reading;
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
  ) => (
    <button
      type="button"
      class="shell-button"
      disabled={disabled}
      onClick={() => {
        run(action);
      }}
    >
      {label}
    </button>
  );
  const back = () =>
    button("Back", () => {
      s.back();
    });

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
  const scene = pageScene(s);
  const wide = [
    "country",
    "news",
    "units",
    "saves",
    "leave",
    "original",
  ].includes(page);
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
          <p class="shell-years">1 9 1 4 · 1 9 1 8</p>
          <nav class="shell-menu" aria-label="Main menu">
            {button("Start", () => {
              s.start();
            })}
            {button(`Code  ${s.mission.code}`, () => {
              s.open("code");
            })}
            {button("Game values", () => {
              s.open("values");
            })}
            {button("Control", () => {
              s.open("control");
            })}
            {button("Extras", () => {
              s.open("extras");
            })}
          </nav>
          <p class="shell-muted">
            {s.mission.news?.label ?? s.mission.month} · {s.mission.name}
          </p>
          <footer class="shell-actions">
            {button("Exit", () => {
              s.open("closed");
            })}
          </footer>
        </>
      );
      break;
    case "code":
      content = (
        <>
          <p>Type the code of a map you have reached, then press Enter.</p>
          <form
            class="shell-menu"
            onSubmit={(event) => {
              event.preventDefault();
              run(() => {
                s.submitCode(code);
              });
            }}
          >
            <label class="shell-field">
              Map code
              <input
                value={code}
                maxLength={12}
                autoComplete="off"
                autoCapitalize="characters"
                spellcheck={false}
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
            <button type="submit" class="shell-button" disabled={busy || !code}>
              Confirm code
            </button>
          </form>
          <footer class="shell-actions">{back()}</footer>
        </>
      );
      break;
    case "values":
      content = (
        <>
          <div class="shell-menu">
            <p class="shell-fixed">Rules · Original (DOS)</p>
            {button(`Battles · ${s.settings.reduced ? "Short" : "Full"}`, () =>
              change({ reduced: !s.settings.reduced }),
            )}
            <label class="shell-field">
              Play mode
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
            <p class="shell-fixed">Player A · Human</p>
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
              )
            ) : (
              <p class="shell-fixed">Player B · Human</p>
            )}
            <p class="shell-fixed">Depots · Own only</p>
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
            )}
          </div>
          <p class="shell-muted">Fewer orders score higher.</p>
          <footer class="shell-actions">{back()}</footer>
        </>
      );
      break;
    case "control":
      content = (
        <>
          <div class="shell-menu">
            <label class="shell-field">
              Layout
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
            )}
            {button(`Scrolling · ${s.settings.scroll}`, () =>
              change({
                scroll:
                  s.settings.scroll === "normal"
                    ? "fast"
                    : s.settings.scroll === "fast"
                      ? "slow"
                      : "normal",
              }),
            )}
            {button(`Motion · ${s.settings.reduced ? "Reduced" : "Full"}`, () =>
              change({ reduced: !s.settings.reduced }),
            )}
            {button(`Battle scenes · ${s.settings.scenes ? "On" : "Off"}`, () =>
              change({ scenes: !s.settings.scenes }),
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
              );
            })()}
            <label class="shell-field">
              UI scale
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
          <p class="shell-muted">
            Two humans uses the original split screen. Player A: mouse or arrows
            + Space. Player B: D/C/X/V + Ctrl. Hot-seat passes control between
            players.
          </p>
          <footer class="shell-actions">{back()}</footer>
        </>
      );
      break;
    case "extras":
      content = (
        <>
          <nav class="shell-menu" aria-label="Extras">
            {button("Score list", () => s.openStorage("scores"))}
            {button("Load game", () => s.openStorage("saves"))}
            {s.originals !== null &&
              button("Original game", () => {
                s.open("original");
              })}
          </nav>
          <footer class="shell-actions">{back()}</footer>
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
                class="shell-country"
                data-side={side}
                disabled={busy}
                onClick={() => {
                  run(() => s.chooseSide(side));
                }}
              >
                <Art
                  p={p}
                  id={side === 0 ? "country/germany" : "country/france"}
                  className="shell-country-flag"
                />
                <strong>{side === 0 ? "Germany" : "France"}</strong>
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
                  />
                  <figcaption>{article.caption}</figcaption>
                </figure>
                <section class="shell-orders">
                  <h3>Orders of the day</h3>
                  <p>{s.mission.briefing}</p>
                </section>
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
            <footer class="shell-actions">
              {button("Continue", () => s.continueNews())}
              {back()}
            </footer>
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
            <p>
              Page {s.unitPage + 1} of{" "}
              {Math.max(1, Math.ceil(s.mission.newUnits.length / 4))}
            </p>
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
                  class="shell-save-slot"
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
                    class={`shell-import ${busy || !s.storageAvailable() ? "shell-disabled" : ""}`}
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
  return (
    <div
      ref={root}
      data-revision={revision}
      onFocusCapture={(event) => {
        if (event.target instanceof HTMLElement)
          lastFocused.current = event.target;
      }}
      class={`shell-ui shell-page-${page}${wide ? " shell-wide" : ""}`}
      style={{ "--shell-scale": s.settings.uiScale }}
      tabIndex={-1}
      role={page === "leave" ? "dialog" : undefined}
      aria-modal={page === "leave" ? true : undefined}
      aria-labelledby="shell-title"
      aria-busy={busy}
    >
      <div class="shell-backdrop" aria-hidden="true">
        {scene && <Art p={p} id={scene} className="shell-scene" cover />}
      </div>
      <main class="shell-sheet">
        <header class="shell-heading">
          <h1 id="shell-title">{title}</h1>
        </header>
        {content}
        {!s.storageAvailable() &&
          ["saves", "scores", "leave"].includes(page) && (
            <p class="shell-notice" role="status">
              Browser storage is unavailable. Saved games and scores cannot be
              stored.
            </p>
          )}
        {(s.note || fileError || busy) && (
          <p class="shell-notice" role="status" aria-live="polite">
            {fileError || s.note || "Please wait…"}
          </p>
        )}
      </main>
    </div>
  );
}
