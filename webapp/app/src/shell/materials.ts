import type { UiAssets } from "../ui/assets.ts";

/**
 * The shell is the inside of a riveted armour plate (art/screens/armour, the screens catalog's armour family): the
 * plate covers the window, rivet strips run along its edges, and every control is a cast brass nameplate, dull at
 * rest and rubbed bright when selected. The art reaches the page's CSS as custom properties; until it has arrived
 * the page shows the plate's colour and plain brass, in the same sizes.
 */

/** The logical sizes of the parts drawn by the CSS (hlasset armour.go: half the trimmed master, or its scale). */
export const plateSize = { width: 836, height: 470 },
  nameplateHeight = 71,
  rivetStripHeight = 78;

/** The opening of armour/embrasure in fractions of the frame (its alpha, measured on the 2x art). */
export const embrasureHole = {
  left: 0.0556,
  top: 0.1479,
  right: 0.0569,
  bottom: 0.1003,
};
/** The frame's width over its height. */
export const embrasureAspect = 1494 / 967;

/**
 * shellUnit is the CSS pixels of one unit of the shell's layout, as shell.css sets --u: the window holds a 16:9 stage of
 * 712×400 units in landscape and a 340×560 column in portrait.
 */
export function shellUnit(
  width: number,
  height: number,
  scale: number,
): number {
  const u =
    width >= height
      ? Math.min(width / 712, height / 400)
      : Math.min(width / 340, height / 560);
  return Math.max(1.1, u) * scale;
}

/** The units of a nameplate's height in the main menu's column; pages scale theirs from it. */
export const menuPlate = 28;
/** The units of the rivet strips along the window's edges (a head is about half of it). */
export const rivetStrip = 24;

/**
 * useArmour is the CSS custom properties of the shell's armour: --hl-plate, --hl-nameplate,
 * --hl-nameplate-polished, --hl-rivets-row, --hl-rivets-column and, on the front page, --hl-paper, each a url()
 * once its art has arrived. --hl-brass is the nameplates' plain brass until then, transparent after.
 */
export function useArmour(
  assets: UiAssets,
  uiScale: number,
  newsprint: boolean,
): Record<string, string> {
  const w = window.innerWidth,
    h = window.innerHeight,
    u = shellUnit(w, h, uiScale);
  const url = (id: string, scale: number) =>
    assets.image("screens", id, scale, false, true).value.url;
  // The plate covers the window; the nameplates and rivets are drawn at the units of the layout.
  const parts: [string, string | null][] = [
    [
      "plate",
      // The plate is only ground under the parts: its 1x art may be enlarged up to 1.6 times.
      url(
        "armour/plate",
        Math.max(w / plateSize.width, h / plateSize.height) / 1.6,
      ),
    ],
    ["nameplate", url("armour/nameplate", (menuPlate * u) / nameplateHeight)],
    [
      "nameplate-polished",
      url("armour/nameplate-polished", (menuPlate * u) / nameplateHeight),
    ],
    [
      "rivets-row",
      url("armour/rivets-row", (rivetStrip * u) / rivetStripHeight),
    ],
    [
      "rivets-column",
      url("armour/rivets-column", (rivetStrip * u) / rivetStripHeight),
    ],
  ];
  // A newsprint tile spans 512 units, as the desktop game lays it.
  if (newsprint)
    parts.push([
      "paper",
      assets.image("screens", "surface/newsprint", u, true).value.url,
    ]);
  const style: Record<string, string> = {};
  for (const [name, part] of parts)
    if (part !== null) style[`--hl-${name}`] = `url("${part}")`;
  if (style["--hl-nameplate"] !== undefined)
    style["--hl-brass"] = "transparent";
  return style;
}
