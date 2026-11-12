// Twin of internal/finmap's and internal/origmaps' tests: synthetic files always, the local original maps where
// present (skipped without them; only numbers enter the assertions).
import { describe, expect, it } from "vitest";
import * as finmap from "./finmap.ts";
import * as origmaps from "./origmaps.ts";
import { moveTablesOf } from "./board.ts";
import { Full } from "./testing/tables.ts";
import { originalFiles } from "./testing/original.ts";
import { hasMagic, unpack } from "./tpwm.ts";

const rules = finmap.rulesFrom(moveTablesOf(Full));
const grass = 0;
const types = moveTablesOf(Full).types;
const typeWith = (want: (cls: number) => boolean): number => {
  const t = types.findIndex((ty) => want(ty.class));
  if (t < 0) throw new Error("no such type");
  return t;
};

function fin(width: number, height: number, cells: [number, number][]) {
  const b = new Uint8Array(4 + 2 * width * height);
  b[1] = width;
  b[3] = height;
  for (let i = 0; i < width * height; i++) {
    const [t, o] = cells[i] ?? [grass, 0xff];
    b[4 + 2 * i] = t;
    b[5 + 2 * i] = o;
  }
  return b;
}

describe("tpwm", () => {
  it("passes plain data and unpacks literals and references", () => {
    expect(unpack(Uint8Array.of(1, 2, 3))).toEqual(Uint8Array.of(1, 2, 3));
    // "TPWM", size 6 LE, flags 0b0010_0000: two literals, then a reference distance 2 length 3, one literal.
    const packed = Uint8Array.of(
      0x54,
      0x50,
      0x57,
      0x4d,
      6,
      0,
      0,
      0,
      0x20,
      0x61,
      0x62,
      0x00,
      0x02,
      0x63,
    );
    expect(hasMagic(packed)).toBe(true);
    expect(new TextDecoder().decode(unpack(packed))).toBe("ababac");
    expect(() => unpack(packed.subarray(0, 12))).toThrow(/truncated/);
    expect(() => unpack(Uint8Array.of(...packed, 0))).toThrow(/trailing/);
  });
});

describe("finmap", () => {
  it("decodes a map and creates units in scan order", () => {
    const infantry = typeWith((c) => (c & 0x10c2) === 0);
    const m = finmap.decodeMap(
      fin(4, 2, [
        [grass, 0xff],
        [grass, 2 * infantry + 1],
        [grass, 0xff],
        [grass, 0xff],
        [grass, 2 * infantry],
      ]),
    );
    expect([m.width, m.height]).toEqual([4, 2]);
    const l = finmap.build({ index: -1, map: m }, rules);
    expect(l.units).toEqual([
      { cell: [0, 1], type: infantry, side: 1 },
      { cell: [1, 0], type: infantry, side: 0 },
    ]);
    expect(l.mission).toBeUndefined();
    expect(l.containers).toBeNull();
  });
  it("gives a two-cell ship its rear and a transporter its hold", () => {
    const ship = typeWith((c) => (c & 0xc0) === 0xc0),
      carrier = typeWith((c) => (c & 0x10c0) === 0x1000);
    const cells: [number, number][] = Array.from({ length: 12 }, () => [
      grass,
      0xff,
    ]);
    cells[5] = [grass, 2 * ship];
    cells[1] = [grass, 2 * (ship + 1)];
    cells[9] = [grass, 2 * ship + 1];
    cells[3] = [grass, 2 * carrier];
    // An allied ship's rear is a row below the head: none there.
    expect(() =>
      finmap.build(
        { index: -1, map: finmap.decodeMap(fin(4, 3, cells)) },
        rules,
      ),
    ).toThrow(/no rear/);
    cells[9] = [grass, 0xff];
    const l = finmap.build(
      { index: -1, map: finmap.decodeMap(fin(4, 3, cells)) },
      rules,
    );
    expect(l.units).toEqual([
      { cell: [0, 3], type: carrier, side: 0, container: 0 },
      { cell: [1, 1], type: ship, side: 0, rear: [0, 1] },
    ]);
  });
  it("decodes shop records, computer builds and codes", () => {
    const shop = new Uint8Array(finmap.typeCount + 1 + 12);
    shop[3] = 1;
    shop[finmap.typeCount] = 1;
    shop.set(
      [1, 0, 0, 20, 5, 4, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff],
      finmap.typeCount + 1,
    );
    const s = finmap.decodeShop(shop);
    expect(s.locked[3]).toBe(true);
    expect(s.records[0]).toMatchObject({
      side: 1,
      kind: 0,
      index: 0,
      energy: 20,
    });
    expect(() => finmap.decodeShop(shop.subarray(1))).toThrow(/SHP/);
    const com = new Uint8Array(6 * finmap.typeCount);
    com[6 * 7 + 3] = 4;
    expect(finmap.computerBuilds(finmap.decodeComputer(com))).toEqual([7]);
    const codes = finmap.decodeCodes(
      Uint8Array.of(0x11, 0x12, 0x13, 0x14, 0x15, 2, 1, 1, 2, 0),
    );
    expect(codes).toEqual([
      {
        password: "ABCDE",
        players: 1,
        step: 1,
        winter: true,
        campaignEnd: false,
      },
    ]);
    expect(() => finmap.decodeCodes(new Uint8Array(9))).toThrow(/whole/);
  });
});

const files = originalFiles();
describe.skipIf(files === null)("the local original maps", () => {
  it("decode as the full game's 73 maps with their missions", () => {
    const levels = origmaps.read(files ?? new Map(), moveTablesOf(Full));
    expect(levels.map((l) => l.source.map)).toEqual([
      ...Array.from({ length: 72 }, (_, i) => String(i).padStart(2, "0")),
      "STATMAP",
    ]);
    for (const l of levels) {
      expect(l.terrain).toHaveLength(l.height);
    }
    expect(levels.filter((l) => l.units.length > 0).length).toBeGreaterThan(70);
    const first = levels[0];
    expect(first?.mission?.players).toBe(1);
    expect(first?.mission?.humanSides).toEqual([0]);
    expect(levels[24]?.mission?.humanSides).toEqual([1]);
    expect(
      levels.filter((l) => l.mission?.players === 2).length,
    ).toBeGreaterThan(0);
    expect(levels[72]?.mission).toBeUndefined();
  });
  it("needs CODES.DAT and a map", () => {
    const t = moveTablesOf(Full);
    expect(() => origmaps.read(new Map(), t)).toThrow(/CODES/);
    const codes = files?.get("CODES.DAT") ?? new Uint8Array();
    expect(() => origmaps.read(new Map([["CODES.DAT", codes]]), t)).toThrow(
      /no maps/,
    );
  });
});
