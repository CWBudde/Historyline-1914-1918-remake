import { useState } from "preact/hooks";
import type { ShellController } from "./controller.ts";
import type { Mission } from "./campaign.ts";
import { fromDrop, fromFileList } from "../original/pick.ts";

/** OriginalPage imports the original game's files from the player's own copy and lists its maps. */
export function OriginalPage({
  shell: s,
  busy,
  button,
}: {
  shell: ShellController;
  busy: boolean;
  button: (
    label: string,
    action: () => void | Promise<void>,
    disabled?: boolean,
  ) => preact.JSX.Element;
}) {
  const [over, setOver] = useState(false);
  const hooks = s.originals;
  if (hooks === null) return <p>The original game cannot be imported here.</p>;
  const status = hooks.status();
  const take = (
    read: Promise<Parameters<ShellController["importOriginals"]>[0]>,
  ) => {
    void read
      .then((picked) => s.importOriginals(picked))
      .catch((error: unknown) => {
        s.note = error instanceof Error ? error.message : String(error);
        s.onChange?.();
      });
  };
  // A map is listed once, as the side the original gives the player; Start then offers both sides.
  const maps = (s.campaign.originals ?? []).filter(
    (m) =>
      m.mode !== "2p" &&
      m.side === (m.original?.map.mission?.humanSides[0] ?? 0),
  );
  const groups: [string, Mission[]][] = [
    ["Central Powers campaign", []],
    ["Entente campaign", []],
    ["Two-player maps", []],
  ];
  for (const m of maps)
    groups[m.original?.map.mission?.players === 2 ? 2 : m.side]?.[1].push(m);
  return (
    <>
      <p>
        Play the original game&apos;s maps with its own tiles: choose the folder
        of your copy of Historyline 1914–1918 (the one with MAP and LIB) or its
        files. They stay in this browser and are never uploaded.
      </p>
      <div
        class={`shell-drop${over ? " shell-drop-over" : ""}`}
        onDragOver={(event) => {
          event.preventDefault();
          setOver(true);
        }}
        onDragLeave={() => {
          setOver(false);
        }}
        onDrop={(event) => {
          event.preventDefault();
          setOver(false);
          if (event.dataTransfer !== null && !busy)
            take(fromDrop(event.dataTransfer));
        }}
      >
        <label
          class={`shell-import hl-nameplate${busy ? " shell-disabled" : ""}`}
        >
          <span>Choose game folder</span>
          <input
            type="file"
            multiple
            disabled={busy}
            aria-label="Choose the original game's folder"
            ref={(input) => {
              // A folder picker; Preact would set the property, which wants a boolean.
              input?.setAttribute("webkitdirectory", "");
            }}
            onChange={(event) => {
              const input = event.currentTarget,
                files = Array.from(input.files ?? []);
              input.value = "";
              take(fromFileList(files));
            }}
          />
        </label>
        <label
          class={`shell-import hl-nameplate${busy ? " shell-disabled" : ""}`}
        >
          <span>Choose files</span>
          <input
            type="file"
            multiple
            disabled={busy}
            aria-label="Choose the original game's files"
            onChange={(event) => {
              const input = event.currentTarget,
                files = Array.from(input.files ?? []);
              input.value = "";
              take(fromFileList(files));
            }}
          />
        </label>
        <p class="shell-muted">or drop the folder here</p>
      </div>
      <p class="shell-muted" role="status">
        {status.maps === 0
          ? "No original files imported."
          : `${status.maps} original maps imported${status.art ? ", with the original tiles and units" : " (without the tile and unit art: the LIB folder and 00.PAL are missing)"}.`}
      </p>
      {groups.map(([title, list]) =>
        list.length === 0 ? null : (
          <section key={title} class="shell-original-maps" aria-label={title}>
            <h2>{title}</h2>
            <div class="shell-original-list">
              {list.map((m) =>
                button(
                  `${m.month.slice(m.month.indexOf("/") + 1)} · ${m.code}${m.season === "winter" ? " · winter" : ""}`,
                  () => {
                    s.selectOriginal(m.key);
                  },
                ),
              )}
            </div>
          </section>
        ),
      )}
      <footer class="shell-actions">
        {status.maps > 0 &&
          button("Remove imported files", () => s.removeOriginals())}
        {button("Back", () => {
          s.back();
        })}
      </footer>
    </>
  );
}
