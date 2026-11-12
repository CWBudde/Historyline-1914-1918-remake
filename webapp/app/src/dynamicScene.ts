import { hexgrid, level, play } from "@historyline/engine";
import { shadowID, type MapScene, type SceneCell } from "./mapScene.ts";

interface Trench {
  id: string;
  ports: number;
  ring: boolean;
}
function trench(ids: readonly string[]): Trench | null {
  for (const id of ids) {
    if (!id.startsWith("trench") || id.length === 6) continue;
    const rest = id.slice(6);
    if (rest === "0") return { id, ports: 0, ring: true };
    let ports = 0;
    for (const digit of rest) {
      if (digit < "1" || digit > "6") break;
      ports |= 1 << (Number(digit) - 1);
    }
    return ports === 0 ? null : { id, ports, ring: false };
  }
  return null;
}
function digits(mask: number): string {
  let out = "";
  for (let i = 0; i < 6; i++) if ((mask & (1 << i)) !== 0) out += String(i + 1);
  return out;
}
function ringID(links: number): string {
  return "trench0" + (links === 0 ? "" : "-" + digits(links));
}
function drawnInfra(
  lv: level.Level,
  c: hexgrid.Coord,
  has: (id: string) => boolean,
): string[] {
  const own = hexAt(lv, c).infra.ids,
    piece = trench(own);
  if (piece === null) return [...own];
  let links = 0,
    spurs = 0;
  for (const [i, n] of hexgrid.around(c).entries()) {
    const bit = 1 << i;
    if (!hexgrid.contains(level.grid(lv), n)) links |= piece.ports & bit;
    else {
      const other = trench(hexAt(lv, n).infra.ids);
      if (
        other !== null &&
        ((piece.ports & bit) !== 0 ||
          (other.ports & (1 << ((i + 3) % 6))) !== 0)
      )
        links |= bit;
      else spurs |= piece.ports & bit;
    }
  }
  let id = piece.ring
    ? ringID(links)
    : "trench" + digits(links) + (spurs === 0 ? "" : "-s" + digits(spurs));
  if (links === piece.ports && spurs === 0) id = piece.id;
  if (
    !has("tile/" + id) &&
    !piece.ring &&
    spurs === 0 &&
    has("tile/" + ringID(links))
  )
    id = ringID(links);
  if (id === piece.id || !has("tile/" + id)) return [...own];
  return own.map((s) => (s === piece.id ? id : s));
}
function hexAt(lv: level.Level, c: hexgrid.Coord): level.Hex {
  const h = lv.lines[c.line]?.[c.col];
  if (h === undefined) throw new Error("dynamic scene: hex outside map");
  return h;
}
function sceneAt(scene: MapScene, c: hexgrid.Coord): SceneCell {
  const h = scene.cells[c.line]?.[c.col];
  if (h === undefined) throw new Error("dynamic scene: hex outside scene");
  return h;
}
function mapSeason(lv: level.Level): string {
  const s = lv.extra?.season;
  if (s === "summer" || s === "autumn" || s === "spring" || s === "winter")
    return s;
  let votes = 0;
  for (const line of lv.lines)
    for (const h of line)
      if (!/^(water|lakeside)/.test(h.terrain))
        votes += h.terrain.endsWith("_w") ? 1 : -1;
  return votes > 0 ? "winter" : "summer";
}
function seasonAt(lv: level.Level, c: hexgrid.Coord, fallback: string): string {
  return lv.extra?.seasons.get(level.coordKey(c)) ?? fallback;
}
function seasonal(
  id: string,
  season: string,
  has: (id: string) => boolean,
  winter: boolean,
): string {
  if (season === "winter" && winter && !id.endsWith("_w") && has(id + "_w"))
    return id + "_w";
  if ((season === "autumn" || season === "spring") && !id.endsWith("_w")) {
    const candidate = id + (season === "autumn" ? "_au" : "_sp");
    if (has(candidate)) return candidate;
  }
  return id;
}
function terrain(lv: level.Level, c: hexgrid.Coord): string {
  const h = hexAt(lv, c),
    override = lv.extra?.seasons.get(level.coordKey(c));
  if (override === undefined) return h.terrain;
  const kind = h.terrain.replace(/_w$/, "");
  return override === "winter" ? kind + "_w" : kind;
}
function background(
  lv: level.Level,
  c: hexgrid.Coord,
  has: (id: string) => boolean,
  season: string,
): string {
  const t = terrain(lv, c);
  let id = "tile/" + t;
  if (t === "grass") {
    const variant = (c.line % 4) * 4 + ((c.col + 3 * c.line) % 4);
    const detail = `tile/grass-detail-${String(variant).padStart(2, "0")}`;
    if (has(detail)) id = detail;
  }
  return seasonal(id, season, has, false);
}
function composeCell(
  lv: level.Level,
  c: hexgrid.Coord,
  has: (id: string) => boolean,
  season: string,
): SceneCell {
  const t = terrain(lv, c),
    h = hexAt(lv, c);
  const standing: string[] = [];
  const add = (id: string): void => {
    if (has(id)) standing.push(seasonal(id, season, has, true));
  };
  add("standing/" + t);
  for (const id of h.infra.ids) add("standing/" + id);
  return {
    background: background(lv, c, has, season),
    overlays: drawnInfra(lv, c, has).map((id) =>
      seasonal("tile/" + id, season, has, true),
    ),
    standing,
    shadows: standing.map(shadowID).filter(has),
    scatter: [],
    ground: [],
    cluster: 0,
    under: "",
  };
}
function site(
  lv: level.Level,
  scene: MapScene,
  c: hexgrid.Coord,
  has: (id: string) => boolean,
  fallback: string,
): void {
  const t = terrain(lv, c);
  if (!/^cmpN(_w)?$/.test(t)) return;
  const suffix = t.endsWith("_w") ? "_w" : "";
  const group = [
    c,
    hexgrid.around(c)[1],
    { line: c.line + 2, col: c.col },
    hexgrid.around(c)[3],
  ];
  for (const [i, p] of group.entries())
    if (
      p === undefined ||
      !hexgrid.contains(scene.grid, p) ||
      terrain(lv, p) !== "cmp" + "NESW".charAt(i) + suffix
    )
      return;
  const vacant = hexAt(lv, c).content === null;
  const sites =
    vacant && ["N", "E", "S", "W"].every((p) => has("tile/site" + p + suffix));
  for (const [i, p] of group.entries()) {
    if (p === undefined) continue;
    sceneAt(scene, p).background = seasonal(
      "tile/" + (sites ? "site" : "cmp") + "NESW".charAt(i) + suffix,
      seasonAt(lv, p, fallback),
      has,
      false,
    );
  }
}
function sameVisual(a: level.Hex, b: level.Hex): boolean {
  return (
    a.terrain === b.terrain &&
    JSON.stringify(a.infra.ids) === JSON.stringify(b.infra.ids) &&
    (a.content === null) === (b.content === null)
  );
}

/** Recompose only work's grass/trench/depot cells and trench neighbours; keep
 * the authored scene for every other terrain rather than inventing contexts. */
export function reworkScene(
  base: MapScene,
  original: level.Level,
  reworked: ReadonlyMap<string, level.Hex>,
  has: (id: string) => boolean,
): { scene: MapScene; level: level.Level; changed: Set<string> } {
  const lv = structuredClone(original),
    scene = structuredClone(base),
    changed = new Set<string>();
  const edits = new Map<string, hexgrid.Coord>();
  for (const [key, h] of reworked) {
    const c = play.coordFromKey(key);
    if (
      !Number.isSafeInteger(c.line) ||
      !Number.isSafeInteger(c.col) ||
      level.coordKey(c) !== key ||
      !hexgrid.contains(base.grid, c)
    )
      throw new Error("dynamic scene: invalid reworked hex");
    if (!sameVisual(hexAt(original, c), h)) edits.set(key, c);
    const row = lv.lines[c.line];
    if (row !== undefined) row[c.col] = structuredClone(h);
  }
  const fallback = mapSeason(lv),
    nearby = new Map(edits);
  for (const [key, c] of edits) {
    const old = sceneAt(base, c),
      cluster = scene.clusters[old.cluster - 1];
    if (cluster !== undefined && cluster.stamp !== "") {
      cluster.stamp = "";
      cluster.standing = "";
      for (const p of cluster.cells) changed.add(level.coordKey(p));
    }
    const row = scene.cells[c.line];
    if (row !== undefined)
      row[c.col] = composeCell(lv, c, has, seasonAt(lv, c, fallback));
    changed.add(key);
    for (const n of hexgrid.around(c))
      if (hexgrid.contains(scene.grid, n)) nearby.set(level.coordKey(n), n);
  }
  for (const [key, c] of nearby) {
    if (edits.has(key) || trench(hexAt(lv, c).infra.ids) === null) continue;
    const cell = sceneAt(scene, c),
      drawn = drawnInfra(lv, c, has)
        .filter((id) => id.startsWith("trench"))
        .map((id) =>
          seasonal("tile/" + id, seasonAt(lv, c, fallback), has, true),
        );
    let index = 0;
    const overlays = cell.overlays.flatMap((id) =>
      id.startsWith("tile/trench")
        ? drawn[index++] === undefined
          ? []
          : [drawn[index - 1] ?? ""]
        : [id],
    );
    overlays.push(...drawn.slice(index));
    if (JSON.stringify(overlays) !== JSON.stringify(cell.overlays)) {
      cell.overlays = overlays;
      changed.add(key);
    }
  }
  // A half-to-complete transition can change all four tiles although only the
  // N piece's content changed. No other building groups need recomposition.
  for (const [line, row] of lv.lines.entries())
    for (const [col] of row.entries()) {
      const c = { line, col };
      if (nearby.has(level.coordKey(c))) site(lv, scene, c, has, fallback);
    }
  for (const [line, row] of scene.cells.entries())
    for (const [col, cell] of row.entries())
      if (JSON.stringify(cell) !== JSON.stringify(base.cells[line]?.[col]))
        changed.add(level.coordKey({ line, col }));
  return { scene, level: lv, changed };
}

/**
 * composeScene composes every hex of a level without a baked scene (an imported original map): each hex its
 * own background, overlays and standing art, as reworkScene composes the hexes the work changed.
 */
export function composeScene(
  lv: level.Level,
  has: (id: string) => boolean,
): MapScene {
  const fallback = mapSeason(lv);
  return {
    grid: level.grid(lv),
    cells: lv.lines.map((line, l) =>
      line.map((_, c) =>
        composeCell(
          lv,
          { line: l, col: c },
          has,
          seasonAt(lv, { line: l, col: c }, fallback),
        ),
      ),
    ),
    clusters: [],
  };
}
