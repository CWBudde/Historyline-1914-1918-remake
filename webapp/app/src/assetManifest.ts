/** The Go exporter owns this schema; URLs and hashes are data, never directory discovery. */
export interface FileRef {
  path: string;
  sha256: string;
  bytes: number;
}
export interface QualityPair {
  lite: FileRef;
  full: FileRef;
}
export interface ImageEntry {
  id: string;
  file: FileRef;
  width: number;
  height: number;
  density: number;
  logicalAnchor: [number, number];
}
export interface ImageCatalog {
  schema: 1;
  entries: ImageEntry[];
}
export interface Tile {
  x: number;
  y: number;
  width: number;
  height: number;
  sourceX: number;
  sourceY: number;
  sourceWidth: number;
  sourceHeight: number;
  file: FileRef;
}
export interface Mip {
  scale: number;
  width: number;
  height: number;
  tiles: Tile[];
}
export interface Pyramid {
  width: number;
  height: number;
  tileSize: number;
  mips: Mip[];
}
export interface LevelAssets {
  files: Record<string, FileRef>;
  layers: Record<string, FileRef>;
  backdrop?: Pyramid;
  backdropFile?: FileRef;
}
export interface AssetIndex {
  schema: 1;
  rules: Record<string, FileRef>;
  levels: Record<string, LevelAssets>;
  packs: Record<string, Record<string, QualityPair>>;
  pixelFonts: Record<string, { image: FileRef; metrics: FileRef }>;
  fonts: Record<string, FileRef>;
  fontLicense: FileRef;
}

export function record(v: unknown): Record<string, unknown> {
  if (typeof v !== "object" || v === null || Array.isArray(v))
    throw new Error("asset manifest: expected object");
  return v as Record<string, unknown>;
}
function positive(v: unknown): number {
  if (typeof v !== "number" || !Number.isSafeInteger(v) || v <= 0)
    throw new Error("asset manifest: expected positive integer");
  return v;
}
function nonnegative(v: unknown): number {
  if (v === 0) return 0;
  return positive(v);
}
function array(v: unknown): unknown[] {
  if (!Array.isArray(v)) throw new Error("asset manifest: expected array");
  return v as unknown[];
}
export function fileRef(v: unknown): FileRef {
  const r = record(v);
  if (
    typeof r.sha256 !== "string" ||
    !/^[a-f0-9]{64}$/.test(r.sha256) ||
    typeof r.path !== "string" ||
    !new RegExp(`^blobs/${r.sha256}\\.[a-z0-9]+$`).test(r.path)
  )
    throw new Error("invalid asset reference");
  return { path: r.path, sha256: r.sha256, bytes: nonnegative(r.bytes) };
}
function refs(v: unknown): Record<string, FileRef> {
  return Object.fromEntries(
    Object.entries(record(v)).map(([k, r]) => [k, fileRef(r)]),
  );
}
export function parsePyramid(v: unknown): Pyramid {
  const p = record(v);
  const width = positive(p.width),
    height = positive(p.height),
    tileSize = positive(p.tileSize);
  const mips = array(p.mips).map((value, i) => {
    const m = record(value);
    const scale = positive(m.scale),
      w = positive(m.width),
      h = positive(m.height);
    if (
      scale !== 2 ** i ||
      w !== Math.ceil(width / scale) ||
      h !== Math.ceil(height / scale)
    )
      throw new Error("asset pyramid: inconsistent mip");
    const seen = new Set<string>();
    const tiles = array(m.tiles).map((value) => {
      const t = record(value);
      const x = nonnegative(t.x),
        y = nonnegative(t.y),
        tw = positive(t.width),
        th = positive(t.height);
      const key = `${x},${y}`;
      const sourceX = nonnegative(t.sourceX),
        sourceY = nonnegative(t.sourceY),
        sourceWidth = positive(t.sourceWidth),
        sourceHeight = positive(t.sourceHeight);
      if (
        sourceX !== Math.max(0, x - 2) ||
        sourceY !== Math.max(0, y - 2) ||
        sourceX + sourceWidth !== Math.min(w, x + tw + 2) ||
        sourceY + sourceHeight !== Math.min(h, y + th + 2)
      )
        throw new Error("asset pyramid: invalid gutter");
      if (
        x >= w ||
        y >= h ||
        x % tileSize !== 0 ||
        y % tileSize !== 0 ||
        tw !== Math.min(tileSize, w - x) ||
        th !== Math.min(tileSize, h - y) ||
        seen.has(key)
      )
        throw new Error("asset pyramid: invalid tile coverage");
      seen.add(key);
      return {
        x,
        y,
        width: tw,
        height: th,
        sourceX,
        sourceY,
        sourceWidth,
        sourceHeight,
        file: fileRef(t.file),
      };
    });
    if (tiles.length !== Math.ceil(w / tileSize) * Math.ceil(h / tileSize))
      throw new Error("asset pyramid: missing tiles");
    return { scale, width: w, height: h, tiles };
  });
  const last = mips.at(-1);
  if (last === undefined || last.width > tileSize || last.height > tileSize)
    throw new Error("asset pyramid: missing coarsest mip");
  return { width, height, tileSize, mips };
}
export function parseAssetIndex(v: unknown): AssetIndex {
  const r = record(v);
  if (r.schema !== 1) throw new Error("unsupported asset schema");
  const levels = Object.fromEntries(
    Object.entries(record(r.levels)).map(([month, value]) => {
      const l = record(value);
      const result: LevelAssets = {
        files: refs(l.files),
        layers: refs(l.layers),
      };
      for (const mode of ["cp", "ec", "2p"]) {
        if (
          result.files[`${mode}.json`] === undefined ||
          result.layers[mode] === undefined
        )
          throw new Error("asset level: missing setup");
      }
      if (
        result.files["map.json"] === undefined ||
        result.files["news.json"] === undefined
      )
        throw new Error("asset level: missing map/news");
      if (l.backdrop !== undefined && l.backdropFile !== undefined)
        throw new Error("asset level: ambiguous backdrop metadata");
      if (l.backdrop !== undefined) result.backdrop = parsePyramid(l.backdrop);
      if (l.backdropFile !== undefined)
        result.backdropFile = fileRef(l.backdropFile);
      return [month, result];
    }),
  );
  const packs = Object.fromEntries(
    Object.entries(record(r.packs)).map(([name, value]) => [
      name,
      Object.fromEntries(
        Object.entries(record(value)).map(([density, value]) => {
          if (!["1", "2", "4"].includes(density))
            throw new Error("asset pack: invalid density");
          const p = record(value);
          return [density, { lite: fileRef(p.lite), full: fileRef(p.full) }];
        }),
      ),
    ]),
  );
  for (const name of ["units", "terrain", "screens", "interface"]) {
    if (
      packs[name]?.["1"] === undefined ||
      packs[name]["2"] === undefined ||
      (name !== "screens" && packs[name]["4"] === undefined)
    )
      throw new Error(`asset pack: missing densities for ${name}`);
  }
  const pixelFonts = Object.fromEntries(
    Object.entries(record(r.pixelFonts)).map(([h, value]) => {
      const p = record(value);
      return [h, { image: fileRef(p.image), metrics: fileRef(p.metrics) }];
    }),
  );
  for (const h of ["5", "7"]) {
    if (pixelFonts[h] === undefined)
      throw new Error("asset font: missing pixel face");
  }
  return {
    schema: 1,
    rules: refs(r.rules),
    levels,
    packs,
    pixelFonts,
    fonts: refs(r.fonts),
    fontLicense: fileRef(r.fontLicense),
  };
}
export function parseImageCatalog(v: unknown, density: number): ImageCatalog {
  const c = record(v);
  if (c.schema !== 1) throw new Error("unsupported image catalog");
  const seen = new Set<string>();
  const entries = array(c.entries).map((value) => {
    const e = record(value);
    const w = positive(e.width),
      h = positive(e.height),
      d = positive(e.density);
    const a = array(e.logicalAnchor);
    if (
      typeof e.id !== "string" ||
      e.id === "" ||
      seen.has(e.id) ||
      d !== density ||
      w % d !== 0 ||
      h % d !== 0 ||
      a.length !== 2 ||
      a.some((v) => typeof v !== "number" || !Number.isFinite(v))
    )
      throw new Error("asset catalog: invalid image geometry or duplicate ID");
    seen.add(e.id);
    return {
      id: e.id,
      file: fileRef(e.file),
      width: w,
      height: h,
      density: d,
      logicalAnchor: a as [number, number],
    };
  });
  return { schema: 1, entries };
}
