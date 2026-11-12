// The original design's textures: the decoded original art packed into atlases, one texture per tile and frame,
// drawn pixel for pixel (nearest scaling).
import { Rectangle, Texture } from "pixi.js";
import type { OriginalArt, Pixels } from "./art.ts";

export interface OriginalTextures {
  /** tiles[season][record] */
  tiles: [Texture[], Texture[]];
  /** units[side][moved][frame] */
  units: [Texture[], Texture[]][];
  dispose(): void;
}

function atlas(images: readonly Pixels[], made: Texture[]): Texture[] {
  const cell = Math.max(1, ...images.map((p) => Math.max(p.width, p.height)));
  const cols = Math.ceil(Math.sqrt(images.length)),
    rows = Math.ceil(images.length / cols);
  const canvas = document.createElement("canvas");
  canvas.width = cols * cell;
  canvas.height = rows * cell;
  const ctx = canvas.getContext("2d");
  if (ctx === null) throw new Error("no 2D canvas for the original art");
  for (const [i, p] of images.entries())
    ctx.putImageData(
      new ImageData(new Uint8ClampedArray(p.rgba), p.width, p.height),
      (i % cols) * cell,
      Math.trunc(i / cols) * cell,
    );
  const base = Texture.from(canvas);
  base.source.scaleMode = "nearest";
  made.push(base);
  return images.map(
    (p, i) =>
      new Texture({
        source: base.source,
        frame: new Rectangle(
          (i % cols) * cell,
          Math.trunc(i / cols) * cell,
          p.width,
          p.height,
        ),
      }),
  );
}

export function originalTextures(art: OriginalArt): OriginalTextures {
  const made: Texture[] = [];
  return {
    tiles: [atlas(art.tiles[0], made), atlas(art.tiles[1], made)],
    units: art.units.map(
      ([still, moved]) =>
        [atlas(still, made), atlas(moved, made)] as [Texture[], Texture[]],
    ),
    dispose() {
      for (const t of made) t.destroy(true);
    },
  };
}
