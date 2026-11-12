import { useEffect, useState } from "preact/hooks";
import type { UiAssets } from "../ui/assets.ts";
import {
  brownGain,
  gradeTile,
  rivetImage,
  seamTile,
  steelGain,
  type Pixels,
} from "./steel.ts";

/**
 * The shell's plates take their materials from the screens catalog's surfaces, as the desktop game's menus do: a
 * graded steel tile for every steel tone, a graded tile for the brown name plates, a rivet head and the hull
 * plating's seams. They are made once per source image and handed to the page's CSS as custom properties.
 */
type Kind = "steel" | "brown" | "rivet" | "seams";
const made = new Map<string, Promise<string>>();
/** Few sources are ever in use; older ones are released. */
const keep = 12;

function context(width: number, height: number): CanvasRenderingContext2D {
  const canvas = document.createElement("canvas");
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext("2d", { willReadFrequently: true });
  if (ctx === null) throw new Error("no 2D canvas");
  return ctx;
}
async function read(url: string): Promise<Pixels> {
  const bitmap = await createImageBitmap(await (await fetch(url)).blob());
  try {
    const ctx = context(bitmap.width, bitmap.height);
    ctx.drawImage(bitmap, 0, 0);
    const data = ctx.getImageData(0, 0, bitmap.width, bitmap.height);
    return { width: data.width, height: data.height, data: data.data };
  } finally {
    bitmap.close();
  }
}
async function encode(pixels: Pixels): Promise<string> {
  const ctx = context(pixels.width, pixels.height),
    image = ctx.createImageData(pixels.width, pixels.height);
  image.data.set(pixels.data);
  ctx.putImageData(image, 0, 0);
  const blob = await new Promise<Blob | null>((done) => {
    ctx.canvas.toBlob(done, "image/png");
  });
  if (blob === null) throw new Error("canvas encoding failed");
  return URL.createObjectURL(blob);
}
function make(kind: Kind, source: string): Promise<string> {
  const key = `${kind} ${source}`;
  let url = made.get(key);
  if (url !== undefined) return url;
  url = read(source).then((pixels) =>
    encode(
      kind === "steel"
        ? gradeTile(pixels, steelGain)
        : kind === "brown"
          ? gradeTile(pixels, brownGain)
          : kind === "rivet"
            ? rivetImage(pixels)
            : seamTile(pixels),
    ),
  );
  made.set(key, url);
  url.catch(() => made.delete(key));
  while (made.size > keep) {
    const oldest = made.keys().next().value;
    if (oldest === undefined) break;
    const old = made.get(oldest);
    made.delete(oldest);
    void old?.then(
      (u) => {
        URL.revokeObjectURL(u);
      },
      () => undefined,
    );
  }
  return url;
}

/** shellUnit is the CSS pixels of one reference pixel of the desktop game's 640×400 menus, as shell.css sets it. */
export function shellUnit(
  width: number,
  height: number,
  scale: number,
): number {
  const u =
    width >= height
      ? Math.min(width / 640, height / 400)
      : Math.min(width / 340, height / 560);
  return Math.max(1.1, u) * scale;
}

/**
 * useShellMaterials is the CSS custom properties of the shell's materials: --hl-steel, --hl-brown, --hl-rivet,
 * --hl-seams and --hl-paper, each a url() once its art has arrived. Until then the plates show their plain colours.
 */
export function useShellMaterials(
  assets: UiAssets,
  uiScale: number,
): Record<string, string> {
  const u = shellUnit(window.innerWidth, window.innerHeight, uiScale);
  // A steel tile spans 256 reference pixels, a newsprint tile 512, as the desktop game lays them (density 1: 512 px).
  const steel = assets.image("screens", "surface/steel", u / 2, true).value.url,
    paper = assets.image("screens", "surface/newsprint", u, true).value.url,
    rivets = assets.image("screens", "surface/rivets", 1).value.url;
  const [urls, setUrls] = useState<Partial<Record<Kind, string>>>({});
  useEffect(() => {
    let live = true;
    const sources: [Kind, string | null][] = [
      ["steel", steel],
      ["brown", steel],
      ["rivet", rivets],
      ["seams", rivets],
    ];
    for (const [kind, source] of sources)
      if (source !== null)
        make(kind, source).then(
          (url) => {
            if (live) setUrls((now) => ({ ...now, [kind]: url }));
          },
          () => {
            // Without the painting the plates keep their plain steel.
          },
        );
    return () => {
      live = false;
    };
  }, [steel, rivets]);
  const style: Record<string, string> = {};
  for (const kind of ["steel", "brown", "rivet", "seams"] as const) {
    const url = urls[kind];
    if (url !== undefined) style[`--hl-${kind}`] = `url("${url}")`;
  }
  if (paper !== null) style["--hl-paper"] = `url("${paper}")`;
  return style;
}
