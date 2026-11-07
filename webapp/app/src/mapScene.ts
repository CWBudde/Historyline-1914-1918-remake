import { hexgrid, level } from "@historyline/engine";
import { fileRef, record, type FileRef } from "./assetManifest.ts";
import type { Point } from "./camera.ts";

export interface Placed {
  id: string;
  x: number;
  y: number;
}
export interface SceneCell {
  background: string;
  overlays: string[];
  standing: string[];
  shadows: string[];
  scatter: Placed[];
  ground: Placed[];
  cluster: number;
  under: string;
}
export interface Cluster {
  kind: string;
  cells: hexgrid.Coord[];
  stamp: string;
  standing: string;
}
export interface MapScene {
  grid: hexgrid.Grid;
  cells: SceneCell[][];
  clusters: Cluster[];
}
export interface Raster {
  x: number;
  y: number;
  width: number;
  height: number;
  divisor: number;
  file: FileRef;
}
export interface SeasonMask {
  at: hexgrid.Coord;
  x: number;
  y: number;
  seasons: string[];
  weights: Raster[];
}
export interface BakedBackground {
  id: string;
  width: number;
  height: number;
  density: number;
  logicalAnchor: [number, number];
  lite: FileRef;
  full: FileRef;
  tileOnly?: boolean;
}
export interface BakedLayers {
  scene: MapScene;
  seasons: Map<string, MapScene>;
  water: Map<string, Raster>;
  blend: Map<string, SeasonMask>;
  backgrounds: Map<string, Map<number, BakedBackground>>;
  seasonBakes: SeasonBakes;
}
export interface SeasonBakes {
  layers: Map<string, Map<string, Map<number, BakedBackground>>>;
  ground: Map<string, Placed[]>;
  standing: Map<string, Placed[]>;
}
export type TerrainPass =
  "bleed" | "background" | "stamp" | "infrastructure" | "shadow" | "standing";
export interface TerrainDraw {
  key: string;
  id: string;
  at: Point;
  coord: hexgrid.Coord;
  pass: TerrainPass;
}

const list = (v: unknown): unknown[] => {
  if (v === null || v === undefined) return [];
  if (!Array.isArray(v)) throw new Error("map scene: expected array");
  return v as unknown[];
};
const text = (v: unknown): string => {
  if (typeof v !== "string") throw new Error("map scene: expected string");
  return v;
};
const number = (v: unknown): number => {
  if (typeof v !== "number" || !Number.isFinite(v))
    throw new Error("map scene: expected number");
  return v;
};
const integer = (v: unknown): number => {
  const n = number(v);
  if (!Number.isSafeInteger(n)) throw new Error("map scene: expected integer");
  return n;
};
const coord = (v: unknown): hexgrid.Coord => {
  const r = record(v);
  return { line: integer(r.Line), col: integer(r.Col) };
};
const strings = (v: unknown) => list(v).map(text);
const placed = (v: unknown): Placed[] =>
  list(v).map((value) => {
    const p = record(value);
    return { id: text(p.ID), x: number(p.X), y: number(p.Y) };
  });
function scene(v: unknown): MapScene {
  const r = record(v),
    g = record(r.Grid),
    grid = { width: integer(g.Width), lines: integer(g.Lines) };
  const cells = list(r.Cells).map((row) =>
    list(row).map((value) => {
      const c = record(value);
      return {
        background: text(c.Background),
        overlays: strings(c.Overlays),
        standing: strings(c.Standing),
        shadows: strings(c.Shadows),
        scatter: placed(c.Scatter),
        ground: placed(c.Ground),
        cluster: integer(c.Cluster),
        under: text(c.Under),
      };
    }),
  );
  if (
    cells.length !== grid.lines ||
    cells.some((row, line) => row.length !== grid.width - (line % 2))
  )
    throw new Error("map scene: invalid row geometry");
  const clusters = list(r.Clusters).map((value) => {
    const c = record(value);
    return {
      kind: text(c.Kind),
      cells: list(c.Cells).map(coord),
      stamp: text(c.Stamp),
      standing: text(c.Standing),
    };
  });
  for (const row of cells)
    for (const c of row)
      if (c.cluster < 0 || c.cluster > clusters.length)
        throw new Error("map scene: invalid cluster");
  return { grid, cells, clusters };
}
function raster(v: unknown): Raster {
  const r = record(v);
  return {
    x: integer(r.x),
    y: integer(r.y),
    width: integer(r.width),
    height: integer(r.height),
    divisor: integer(r.divisor),
    file: fileRef(r.file),
  };
}
function backgrounds(v: unknown): Map<number, BakedBackground> {
  return new Map(
    Object.entries(record(v ?? {})).map(([density, v]) => {
      const b = record(v),
        d = integer(b.density),
        anchor = list(b.logicalAnchor);
      if (![1, 2, 4].includes(d) || String(d) !== density)
        throw new Error("map scene: invalid background density");
      if (anchor.length !== 2)
        throw new Error("map scene: invalid background anchor");
      const width = integer(b.width),
        height = integer(b.height);
      if (width <= 0 || height <= 0)
        throw new Error("map scene: invalid background size");
      return [
        d,
        {
          id: text(b.id),
          width,
          height,
          density: d,
          logicalAnchor: [number(anchor[0]), number(anchor[1])] as [
            number,
            number,
          ],
          lite: fileRef(b.lite),
          full: fileRef(b.full),
          ...(b.tileOnly === true ? { tileOnly: true } : {}),
        },
      ];
    }),
  );
}
export function parseBakedLayers(v: unknown): BakedLayers {
  const r = record(v);
  if (r.schema !== 1) throw new Error("unsupported baked map schema");
  const seasons = record(r.seasonBakes ?? {});
  return {
    scene: scene(r.scene),
    backgrounds: new Map(
      Object.entries(record(r.backgrounds ?? {})).map(([key, value]) => [
        key,
        backgrounds(value),
      ]),
    ),
    seasonBakes: {
      layers: new Map(
        Object.entries(record(seasons.layers ?? {})).map(([key, passes]) => [
          key,
          new Map(
            Object.entries(record(passes)).map(([pass, value]) => {
              if (!["background", "infrastructure", "standing"].includes(pass))
                throw new Error("map scene: invalid season pass");
              return [pass, backgrounds(value)];
            }),
          ),
        ]),
      ),
      ground: new Map(
        Object.entries(record(seasons.ground ?? {})).map(([key, value]) => [
          key,
          placed(value),
        ]),
      ),
      standing: new Map(
        Object.entries(record(seasons.standing ?? {})).map(([key, value]) => [
          key,
          placed(value),
        ]),
      ),
    },
    seasons: new Map(
      Object.entries(record(r.seasons)).map(([s, value]) => [s, scene(value)]),
    ),
    water: new Map(
      Object.entries(record(r.water)).map(([c, value]) => [c, raster(value)]),
    ),
    blend: new Map(
      Object.entries(record(r.blend)).map(([c, value]) => {
        const m = record(value);
        return [
          c,
          {
            at: coord(m.at),
            x: integer(m.x),
            y: integer(m.y),
            seasons: strings(m.seasons),
            weights: list(m.weights).map(raster),
          },
        ];
      }),
    ),
  };
}

export function summerID(id: string): string {
  return id.replace(/_(au|sp)$/u, "");
}
export function shadowID(id: string): string {
  return "standing/shadow/" + summerID(id).replace(/^standing\//u, "");
}
function clusteredStanding(id: string): boolean {
  const name = summerID(id).replace(/^standing\//u, "");
  return (
    [
      "mud",
      "shelled_w",
      "stone",
      "stoney_w",
      "mountain",
      "mountain_w",
      "water",
      "waterDeep",
      "waterSea",
      "waterStones",
    ].includes(name) || name.startsWith("lakeside")
  );
}

export function terrainDraws(
  s: MapScene,
  visible: readonly hexgrid.Coord[],
  baked?: SeasonBakes,
  changed: ReadonlySet<string> = new Set(),
): TerrainDraw[] {
  const kept = new Map(visible.map((c) => [level.coordKey(c), c]));
  for (const c of visible) {
    const cell = s.cells[c.line]?.[c.col],
      cluster = cell === undefined ? undefined : s.clusters[cell.cluster - 1];
    const anchor = cluster?.cells[0];
    if (cluster?.stamp !== "" && anchor !== undefined)
      kept.set(level.coordKey(anchor), anchor);
  }
  const cells = [...kept.values()].sort(
      (a, b) => a.line - b.line || a.col - b.col,
    ),
    result: TerrainDraw[] = [];
  for (const pass of [
    "bleed",
    "background",
    "stamp",
    "infrastructure",
    "shadow",
    "standing",
  ] as const) {
    for (const c of cells) {
      const cell = s.cells[c.line]?.[c.col];
      if (cell === undefined) continue;
      const cluster = s.clusters[cell.cluster - 1],
        stamped = cluster !== undefined && cluster.stamp !== "";
      const atAnchor =
        stamped &&
        cluster.cells[0]?.line === c.line &&
        cluster.cells[0].col === c.col;
      const standing = stamped
        ? [
            ...(atAnchor && cluster.standing !== "" ? [cluster.standing] : []),
            ...cell.standing.filter((id) => !clusteredStanding(id)),
          ]
        : cell.standing;
      const [x, y] = hexgrid.center(c);
      const coordKey = level.coordKey(c),
        selected = changed.has(coordKey) ? undefined : baked;
      const layer = selected?.layers
        .get(coordKey)
        ?.get(pass === "bleed" ? "background" : pass);
      const mixed = () => {
        const image = layer?.values().next().value;
        if (image !== undefined) add(image.id);
      };
      let index = 0;
      const add = (id: string, dx = 0, dy = 0) => {
        if (id !== "")
          result.push({
            key: `${pass}/${c.line},${c.col}/${index++}`,
            id,
            at: { x: x + dx, y: y + dy },
            coord: c,
            pass,
          });
      };
      switch (pass) {
        case "bleed":
        case "background":
          if (layer !== undefined) mixed();
          else add(stamped ? cell.under : cell.background);
          break;
        case "stamp":
          if (atAnchor) add(cluster.stamp);
          if (!stamped)
            for (const p of selected?.ground.get(coordKey) ?? cell.ground)
              add(p.id, p.x, p.y);
          break;
        case "infrastructure":
          if (layer !== undefined) mixed();
          else for (const id of cell.overlays) add(id);
          break;
        case "shadow":
          for (const id of standing) {
            const shadow = shadowID(id);
            if (cell.shadows.includes(shadow)) add(shadow);
          }
          if (!stamped)
            for (const p of cell.scatter) add(shadowID(p.id), p.x, p.y);
          break;
        case "standing":
          if (layer !== undefined) mixed();
          else for (const id of standing) add(id);
          if (!stamped)
            for (const p of selected?.standing.get(coordKey) ?? cell.scatter)
              add(p.id, p.x, p.y);
          break;
      }
    }
  }
  return result;
}
