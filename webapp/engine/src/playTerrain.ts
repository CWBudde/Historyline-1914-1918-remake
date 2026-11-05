import * as rules from "./cycle.ts";
import * as sappers from "./sapper.ts";
import { builder as ruleBuilder } from "./builder.ts";
import { cell, coord } from "./bind.ts";
import { type Coord, around } from "./hexgrid.ts";
import { type Hex, coordKey } from "./level.ts";
import { type Session, faction } from "./playModel.ts";
import { wrap, refusal } from "./playEvent.ts";
import { at } from "./goruntime.ts";
import { liveCell } from "./board.ts";

export function sapper(s: Session, unit: number): boolean {
  return sappers.sapper(rules.liveOf(s.game, unit));
}
export function builder(s: Session, unit: number): boolean {
  return ruleBuilder(rules.liveOf(s.game, unit));
}
export function canConstruct(s: Session, side: number, unit: number): void {
  const u = s.game.units.get(unit);
  if (u === undefined || !ruleBuilder(u))
    throw refusal("not a construction unit");
  const e = rules.check(s.game, {
    kind: "attack",
    side,
    unit,
    cell: liveCell(u, side),
  });
  if (e instanceof rules.Refusal && e.code === rules.msgNoTarget) return;
  if (e !== null) throw wrap(e);
}
export function work(
  s: Session,
  side: number,
  unit: number,
  to: Coord,
): { hexes: Coord[]; recs: number[]; fill: boolean } {
  const target = cell(s.board, to),
    out: { hexes: Coord[]; recs: number[]; fill: boolean } = {
      hexes: [],
      recs: [],
      fill: false,
    };
  if (target === null || !sapper(s, unit)) return out;
  const [cells, recs, fill] = sappers.workCells(s.game, side, unit, target);
  out.fill = fill;
  for (const [k, c] of cells.entries()) {
    const p = coord(s.board, c);
    if (p !== null) {
      out.hexes.push(p);
      out.recs.push(at(recs, k));
    }
  }
  return out;
}
const depotGround: Readonly<Record<number, number>> = {
  15: 3,
  16: 3,
  17: 3,
  18: 5,
  19: 6,
  20: 4,
  21: 7,
  22: 9,
  23: 10,
  24: 8,
};
function groundOf(rec: number): number {
  return sappers.trench(rec)
    ? 1
    : rec === sappers.terrainFilled
      ? 2
      : (depotGround[rec] ?? 0);
}
function seasonal(s: Session, t: string): string {
  return s.board.winter ? t + "_w" : t;
}
function trenchPiece(s: Session, c: Coord): string {
  let id = "trench";
  const ns = around(c);
  for (const k of [0, 1, 3, 4]) {
    const pos = cell(s.board, at(ns, k));
    if (pos !== null && sappers.trench(at(s.game.maps[0], 2 * pos)))
      id += String(k + 1);
  }
  return id === "trench" ? "trench0" : id;
}
export function reworked(s: Session): Map<string, Hex> {
  const out = new Map<string, Hex>(),
    m = s.game.maps[0];
  for (const [pos, orig] of s.board.terrain.entries()) {
    const rec = at(m, 2 * pos),
      c = coord(s.board, pos);
    if (c === null || groundOf(rec) === groundOf(orig)) continue;
    const h = { ...at(at(s.level.lines, c.line), c.col) },
      g = groundOf(rec);
    if (g >= 3) {
      h.terrain = seasonal(s, "cmp" + "NESW".charAt((g - 3) % 4));
      h.infra = { ids: [], list: false };
      h.content = null;
      if (rec === 15 || rec === 16 || rec === 17)
        h.content = {
          owner: faction(rec === 15 ? 2 : rec === 16 ? 0 : 1),
          units: null,
        };
      out.set(coordKey(c), h);
      continue;
    }
    const ids = h.infra.ids.filter((id) => !id.startsWith("trench"));
    h.terrain = seasonal(s, "grass");
    if (g === 1) ids.push(trenchPiece(s, c));
    h.infra = { ids, list: ids.length > 1 };
    out.set(coordKey(c), h);
  }
  return out;
}
