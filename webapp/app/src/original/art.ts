// The original game's map art, decoded from the files the player imported (never shipped): the terrain tiles,
// summer (LIB/PARTS) and winter (LIB/PARTW), one per terrain record, and the unit sprites (LIB/UNIT), six frames
// per unit type clockwise from north, in the 00.PAL palette. The desktop's original design did the same.
import { origview, tpwm } from "@historyline/engine";

/** An RGBA image, row-major, alpha 0 where the art is transparent. */
export interface Pixels {
  width: number;
  height: number;
  rgba: Uint8ClampedArray;
}
export interface OriginalArt {
  /** tiles[season][record], season 0 summer, 1 winter. */
  tiles: [Pixels[], Pixels[]];
  /** units[side][moved][6 * type + facing], side 0, 1 and 2 (neutral), moved 0 or 1 (grey). */
  units: [Pixels[], Pixels[]][];
}

/** unitTypes is the number of unit types, each with six frames. */
export const unitTypes = 51;
/** The first palette entries of each side's unit colours: side 0 green, side 1 tan; neutral units keep their own. */
const sideColours = [0x20, 0x30] as const;

/** artFiles are the files the original design reads. */
export const artFiles = [
  "00.PAL",
  "LIB/PARTS.DAT",
  "LIB/PARTS.LIB",
  "LIB/PARTW.DAT",
  "LIB/PARTW.LIB",
  "LIB/UNIT.DAT",
  "LIB/UNIT.LIB",
] as const;

function u16(b: Uint8Array, i: number): number {
  return (b[i] ?? 0) | ((b[i + 1] ?? 0) << 8);
}
function u32(b: Uint8Array, i: number): number {
  return (u16(b, i) | (u16(b, i + 2) << 16)) >>> 0;
}
function name(b: Uint8Array, i: number): string {
  return String.fromCharCode(...b.subarray(i, i + 8));
}

/** directory maps a library's entry names to offsets: it ends in 8-byte names and 32-bit offsets. */
function directory(lib: Uint8Array): Map<string, number> {
  const at = new Map<string, number>();
  if (lib.length < 4) return at;
  for (let i = u32(lib, 0); i + 12 <= lib.length; i += 12)
    at.set(name(lib, i), u32(lib, i + 8));
  return at;
}

/** Pixels lie in four planes of every fourth column. */
function planar(i: number, w: number, h: number): [number, number] {
  const q = (w / 4) * h;
  return [
    Math.trunc(i / q) + ((i % q) % (w / 4)) * 4,
    Math.trunc((i % q) / (w / 4)),
  ];
}

/** decodeTiles cuts the tiles of a library out in the order of names, one per terrain record. */
export function decodeTiles(
  names: Uint8Array,
  lib: Uint8Array,
  pal: Uint8Array,
): Pixels[] {
  if (
    names.length !== 8 * origview.records ||
    pal.length < 768 ||
    lib.length < 4
  )
    throw new Error("unexpected tile file sizes");
  const at = directory(lib);
  const out: Pixels[] = [];
  for (let r = 0; r < origview.records; r++) {
    const n = name(names, 8 * r),
      off = at.get(n);
    if (off === undefined || off + 18 > lib.length)
      throw new Error(`record ${r}: no tile`);
    const key = lib[off + 8],
      w = u16(lib, off + 14),
      h = u16(lib, off + 16);
    if (lib.length < off + 18 + w * h)
      throw new Error(`record ${r}: tile cut short`);
    const rgba = new Uint8ClampedArray(4 * w * h);
    for (let i = 0; i < w * h; i++) {
      const p = lib[off + 18 + i] ?? 0;
      if (p === key) continue;
      const [x, y] = planar(i, w, h),
        o = 4 * (y * w + x);
      rgba[o] = pal[3 * p] ?? 0;
      rgba[o + 1] = pal[3 * p + 1] ?? 0;
      rgba[o + 2] = pal[3 * p + 2] ?? 0;
      rgba[o + 3] = 255;
    }
    out.push({ width: w, height: h, rgba });
  }
  return out;
}

/** decodeUnits decodes the unit frames by side (0, 1, 2) and moved (0, 1), in unit type order. */
export function decodeUnits(
  names: Uint8Array,
  lib: Uint8Array,
  pal: Uint8Array,
): [Pixels[], Pixels[]][] {
  if (names.length !== 8 * 6 * unitTypes || pal.length < 768)
    throw new Error("unexpected unit file sizes");
  const at = directory(lib);
  const out: [Pixels[], Pixels[]][] = [
    [[], []],
    [[], []],
    [[], []],
  ];
  for (let k = 0; k < 6 * unitTypes; k++) {
    const off = at.get(name(names, 8 * k));
    if (off === undefined || off + 18 > lib.length)
      throw new Error(`unit frame ${k}: no sprite`);
    const key = lib[off + 8],
      own = lib[off + 9] ?? 0,
      w = u16(lib, off + 14),
      h = u16(lib, off + 16);
    if (lib.length < off + 18 + (w * h) / 2)
      throw new Error(`unit frame ${k}: sprite cut short`);
    for (const [side, sets] of out.entries()) {
      const base = sideColours[side] ?? own;
      const im = new Uint8ClampedArray(4 * w * h),
        grey = new Uint8ClampedArray(4 * w * h);
      for (let i = 0; i < w * h; i++) {
        const byte = lib[off + 18 + (i >> 1)] ?? 0;
        const v = i % 2 === 0 ? byte >> 4 : byte & 15;
        if (v === key) continue;
        const [x, y] = planar(i, w, h),
          o = 4 * (y * w + x),
          c = 3 * (base + v);
        const r = pal[c] ?? 0,
          g = pal[c + 1] ?? 0,
          b = pal[c + 2] ?? 0;
        im.set([r, g, b, 255], o);
        const l = Math.trunc(((299 * r + 587 * g + 114 * b) * 3) / 4 / 1000);
        grey.set([l, l, l, 255], o);
      }
      sets[0].push({ width: w, height: h, rgba: im });
      sets[1].push({ width: w, height: h, rgba: grey });
    }
  }
  return out;
}

/** decodeArt decodes the original design's art from the imported files; null when any file is missing. */
export function decodeArt(
  files: ReadonlyMap<string, Uint8Array>,
): OriginalArt | null {
  const raw = (path: string): Uint8Array | undefined => {
    const b = files.get(path);
    return b === undefined ? undefined : tpwm.unpack(b);
  };
  const [pal, sNames, sLib, wNames, wLib, uNames, uLib] = artFiles.map(raw);
  if (
    pal === undefined ||
    sNames === undefined ||
    sLib === undefined ||
    wNames === undefined ||
    wLib === undefined ||
    uNames === undefined ||
    uLib === undefined
  )
    return null;
  return {
    tiles: [decodeTiles(sNames, sLib, pal), decodeTiles(wNames, wLib, pal)],
    units: decodeUnits(uNames, uLib, pal),
  };
}
