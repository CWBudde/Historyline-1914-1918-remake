import * as rules from "./cycle.ts";
import { destroyed } from "./board.ts";
import { income } from "./energy.ts";
import { score } from "./victory.ts";
import { tablesOf } from "./profile.ts";
import { at } from "./goruntime.ts";
import { type Session, Depot, Factory, flagSide, sideOf } from "./playModel.ts";
import { typeID, typeName, typeDomain, buildings } from "./playView.ts";
import { wrap } from "./playEvent.ts";

export interface Weapon {
  power: number;
  range: number;
}
export interface TypeView {
  type: number;
  id: string;
  name: string;
  domain: number;
  air: Weapon;
  ground: Weapon;
  water: Weapon;
  armour: number;
  speed: number;
  weight: number;
  group: number;
  cost: number;
}
export function typeInfo(
  s: Session,
  side: number,
  typ: number,
): TypeView | null {
  const ts = tablesOf(s.game.profile),
    t = ts.types[typ];
  if (t === undefined) return null;
  const mt = ts.move?.types[typ],
    id = typeID(s, typ, side);
  return {
    type: typ,
    id,
    name: typeName(s, id, typ),
    domain: typeDomain(ts, typ),
    air: { power: t.airPower, range: t.airRange - 1 },
    ground: { power: t.groundPower, range: t.surfaceRange - 1 },
    water: { power: t.waterPower, range: t.surfaceRange - 1 },
    armour: t.armour,
    speed: mt?.speed ?? 0,
    weight: mt?.weight ?? 0,
    group: t.group,
    cost: t.cost === 255 ? -1 : t.cost,
  };
}
export interface Holdings {
  units: number;
  depots: number;
  factories: number;
}
export interface Stats {
  round: number;
  level: string;
  levelNo: number;
  mode: [number, number];
  score: number;
  sides: [Holdings, Holdings, Holdings];
  map: Holdings;
  energy: [number, number];
  income: [number, number];
  orders: [number, number];
  limit: number;
}
export function stats(s: Session): Stats {
  const g = s.game,
    empty = (): Holdings => ({ units: 0, depots: 0, factories: 0 });
  const st: Stats = {
    round: g.turn,
    level: s.level.mapName,
    levelNo: s.level.mapNo,
    mode: [g.players[0].phase, g.players[1].phase],
    score: score(g),
    sides: [empty(), empty(), empty()],
    map: empty(),
    energy: [...g.energy],
    income: [income(g, 0), income(g, 1)],
    orders: [g.players[0].orders.length, g.players[1].orders.length],
    limit: g.limit,
  };
  for (const u of g.units.values())
    if (!destroyed(u)) at(st.sides, flagSide(u.flags4)).units++;
  for (const b of buildings(s, -1).values()) {
    const h = at(st.sides, sideOf(b.owner));
    if (b.kind === Depot) h.depots++;
    else if (b.kind === Factory) h.factories++;
  }
  for (const h of st.sides) {
    st.map.units += h.units;
    st.map.depots += h.depots;
    st.map.factories += h.factories;
  }
  return st;
}
export function canRepair(
  s: Session,
  side: number,
  hold: number,
  slot: number,
): void {
  const e = rules.check(s.game, { kind: "repair", side, hold, slot, cell: 0 });
  if (e !== null) throw wrap(e);
}
export function canBuild(
  s: Session,
  side: number,
  hold: number,
  type: number,
): void {
  const e = rules.check(s.game, { kind: "build", side, hold, type, cell: 0 });
  if (e !== null) throw wrap(e);
}
export interface BuildOption {
  type: number;
  id: string;
  name: string;
  cost: number;
}
export function buildable(
  s: Session,
  side: number,
  hold: number,
): BuildOption[] {
  const b = s.game.buildings.get(hold);
  if ((side !== 0 && side !== 1) || b === undefined || hold < 70) return [];
  return b.builds.flatMap((typ) => {
    const t = tablesOf(s.game.profile).types[typ];
    if (t === undefined || t.cost === 255) return [];
    const id = typeID(s, typ, side);
    return id === ""
      ? []
      : [{ type: typ, id, name: typeName(s, id, typ), cost: t.cost }];
  });
}
