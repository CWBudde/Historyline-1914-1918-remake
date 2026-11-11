// Twin of internal/dosplay/works_test.go.

import { describe, expect, it } from "vitest";
import * as play from "./play.ts";
import { buildingKind, kindFactory } from "./capture.ts";
import { wireCharges } from "./wire.ts";
import { ambulanceHeal } from "./ambulance.ts";
import { type Coord, around } from "./hexgrid.ts";
import { type Hex, type Level, coordKey } from "./level.ts";
import { at } from "./goruntime.ts";
import { Extended, Full } from "./testing/tables.ts";
import { playData, shippedLevel } from "./testing/play.ts";

function grassLevel(units: readonly [Coord, string][]): Level {
  const lv: Level = {
    mapName: "grass",
    mapNo: 0,
    player: "Central Powers",
    extra: null,
    lines: Array.from({ length: 11 }, (_, l) =>
      Array.from({ length: 8 - (l % 2) }, (): Hex => ({
        terrain: "grass",
        infra: { ids: [], list: false },
        unit: "",
        content: null,
      })),
    ),
  };
  for (const [c, id] of units) at(at(lv.lines, c.line), c.col).unit = id;
  return lv;
}

function session(lv: Level, extendedUnits = false): play.Session {
  return play.newSession(lv, playData, {
    profile: Full,
    extended: Extended,
    extendedUnits,
  });
}

function passUntilAction(s: play.Session, side: number): void {
  while (play.phase(s, side) !== 2) {
    play.confirm(s, 0);
    play.confirm(s, 1);
    play.resolve(s);
  }
}

// resolveAll confirms both sides and resolves once, returning the events.
function resolveAll(s: play.Session): play.Event[] {
  play.confirm(s, 0);
  play.confirm(s, 1);
  return play.resolve(s);
}

const ofKind = (evs: readonly play.Event[], k: number): play.Event[] =>
  evs.filter((e) => e.kind === k);

const index = (s: play.Session, c: Coord): number =>
  play.view(s, 0).units.get(coordKey(c))?.index ?? -1;

describe("TestWiringParty", () => {
  it("wires up to three hexes of open ground, drawn as grass with barbed wire", () => {
    const from = { line: 4, col: 3 },
      to = at(around(from), 3);
    const s = session(
      grassLevel([
        [from, "CP_wiring"],
        [{ line: 9, col: 3 }, "EC_infantry"],
      ]),
    );
    const party = index(s, from);
    expect(play.wiring(s, party)).toBe(true);
    expect(play.charges(s, party)).toBe(wireCharges);
    passUntilAction(s, 0);
    play.canWire(s, 0, party);
    expect(play.targets(s, 0, party).has(coordKey(to))).toBe(true);
    const hexes = play.wireHexes(s, 0, party, to);
    expect(hexes).toHaveLength(3);
    expect(hexes).toContainEqual(to);
    let evs = play.attack(s, 0, party, to);
    expect(evs.map((e) => [e.kind, e.text])).toStrictEqual([
      [play.Tasked, "Wire laying ordered."],
    ]);
    evs = resolveAll(s);
    const w = ofKind(evs, play.Wired);
    expect(w).toHaveLength(1);
    expect(at(w, 0).from).toStrictEqual(from);
    expect(at(w, 0).to).toStrictEqual(to);
    expect(at(w, 0).hexes).toStrictEqual(hexes);
    const re = play.reworked(s);
    for (const c of hexes) {
      const h = re.get(coordKey(c));
      expect(h?.terrain).toBe("grass");
      expect(h?.infra.ids).toStrictEqual(["barbedWire"]);
    }
    expect(play.charges(s, party)).toBe(wireCharges - 1);
  });
});

describe("TestTunnellers", () => {
  it("blows the mine: the unit on the hex loses most of its group and the hex becomes a crater field", () => {
    const from = { line: 4, col: 3 },
      to = at(around(from), 3);
    const s = session(
      grassLevel([
        [from, "CP_tunnellers"],
        [to, "EC_infantry"],
      ]),
    );
    const tun = index(s, from),
      enemy = index(s, to);
    passUntilAction(s, 0);
    expect(play.tunneller(s, tun)).toBe(true);
    play.canTunnel(s, 0, tun);
    expect(play.targets(s, 0, tun).has(coordKey(to))).toBe(true);
    expect(at(play.attack(s, 0, tun, to), 0).text).toBe("Tunnelling ordered.");
    const blown: play.Event[] = [],
      mined: play.Event[] = [];
    for (let k = 0; k < 6; k++) {
      const evs = resolveAll(s);
      if (ofKind(evs, play.Tunnelled).length > 0) {
        const tn = play.tunnels(s, 0);
        expect(tn).toHaveLength(1);
        expect(at(tn, 0).mine).toStrictEqual(to);
        expect(at(tn, 0).from).toStrictEqual(from);
        expect(play.tunnels(s, 1)).toStrictEqual([]);
      }
      blown.push(...ofKind(evs, play.Blown));
      mined.push(...ofKind(evs, play.Mined));
      if (blown.length > 0) break;
    }
    expect(blown).toHaveLength(1);
    expect(at(blown, 0).to).toStrictEqual(to);
    expect(at(blown, 0).hexes).toStrictEqual([to]);
    expect(mined).toHaveLength(1);
    expect(at(mined, 0).unit).toBe(enemy);
    expect(at(mined, 0).loss).not.toBe(0);
    expect(at(mined, 0).text.endsWith("to a mine.")).toBe(true);
    const h = play.reworked(s).get(coordKey(to));
    expect(h?.terrain).toBe("mud");
    expect(h?.infra.ids).toStrictEqual([]);
    expect(play.tunnelOf(s, tun)).toBeNull();
    expect(play.tunnels(s, 0)).toStrictEqual([]);
  });
});

describe("TestFieldAmbulance", () => {
  it("heals the foot units of its side next to it once per turn", () => {
    const from = { line: 4, col: 3 },
      next = at(around(from), 3);
    const s = session(
      grassLevel([
        [from, "CP_ambulance"],
        [next, "CP_infantry"],
        [{ line: 9, col: 3 }, "EC_infantry"],
      ]),
    );
    const amb = index(s, from),
      inf = index(s, next);
    const u = s.game.units.get(inf);
    if (u === undefined) throw new Error("no infantry");
    u.size = 3;
    expect(play.ambulance(s, amb)).toBe(true);
    expect(play.ambulance(s, inf)).toBe(false);
    const healed: play.Event[] = [];
    for (let k = 0; k < 2; k++)
      healed.push(...ofKind(resolveAll(s), play.Healed));
    expect(healed).toHaveLength(1);
    expect(at(healed, 0).unit).toBe(inf);
    expect(at(healed, 0).gain).toBe(ambulanceHeal);
    expect(at(healed, 0).to).toStrictEqual(next);
    expect(at(healed, 0).text.endsWith("regain 1.")).toBe(true);
    expect(s.game.units.get(inf)?.size).toBe(3 + ambulanceHeal);
  });
});

describe("TestExtensionBuildsByMonth", () => {
  it("brings only the extension types of the level's month into the factories", () => {
    for (const c of [
      {
        month: "1918-02",
        has: [51, 52, 53, 57, 59, 60, 61, 62],
        not: [54, 55, 56, 58],
      },
      {
        month: "1915-05",
        has: [51, 52, 54, 59],
        not: [53, 55, 56, 57, 58],
      },
    ]) {
      const s = session(shippedLevel(c.month, "cp"), true);
      let factories = 0;
      for (const [i, bd] of s.game.buildings) {
        if (buildingKind(i) !== kindFactory) continue;
        factories++;
        for (const typ of c.has) expect(bd.builds, c.month).toContain(typ);
        for (const typ of c.not) expect(bd.builds, c.month).not.toContain(typ);
      }
      expect(factories).toBeGreaterThan(0);
    }
  });
});
