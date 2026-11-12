// The imported original files of this document: their art for the original map design, decoded once and turned
// into textures when a battle first draws it.
import { decodeArt, type OriginalArt } from "./art.ts";
import { originalTextures, type OriginalTextures } from "./textures.ts";

let art: OriginalArt | null = null,
  textures: OriginalTextures | null = null;

/** setOriginalArt decodes the art of the imported files (null clears it); it throws on damaged art files. */
export function setOriginalArt(
  files: ReadonlyMap<string, Uint8Array> | null,
): boolean {
  textures?.dispose();
  textures = null;
  art = files === null ? null : decodeArt(files);
  return art !== null;
}
/** hasOriginalArt reports whether the original design can be drawn. */
export function hasOriginalArt(): boolean {
  return art !== null;
}
/** originalArt is the original design's textures, null without imported art. */
export function originalArt(): OriginalTextures | null {
  if (art === null) return null;
  textures ??= originalTextures(art);
  return textures;
}
