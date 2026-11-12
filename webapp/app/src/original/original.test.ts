// The original game's import: file sorting and the design cycle always; the art, the maps as missions and their
// battles with the local original files (original/extracted/dos-full, git-ignored), skipped without them.
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { play, tpwm } from "@historyline/engine";
import { decodeArt, unitTypes } from "./art.ts";
import { gamePath, select } from "./files.ts";
import { design, nextDesign } from "./design.ts";
import { loadOriginalTables, originalMissions } from "./catalog.ts";
import { loadBattleData, missionSession } from "../battle/session.ts";

const repo = fileURLToPath(new URL("../../../../", import.meta.url));
const dir = `${repo}original/extracted/dos-full/`;
const local = existsSync(`${dir}CODES.DAT`);

/** picked are the local original folder's files as a folder pick gives them. */
function picked() {
  const out: { path: string; bytes: Uint8Array }[] = [];
  const walk = (rel: string) => {
    for (const name of readdirSync(dir + rel, { withFileTypes: true })) {
      const path = rel === "" ? name.name : `${rel}/${name.name}`;
      if (name.isDirectory()) {
        if (["MAP", "LIB"].includes(name.name.toUpperCase())) walk(path);
      } else
        out.push({
          path: `dos-full/${path}`,
          bytes: new Uint8Array(readFileSync(dir + path)),
        });
    }
  };
  walk("");
  return out;
}

const rules: Record<string, string> = {
  "units.json": "internal/data/units.json",
  "records.json": "internal/dosbind/records.json",
  "full.json": "internal/dosrules/full.json",
  "demo.json": "internal/dosrules/demo.json",
  "extension.json": "internal/dosrules/extension.json",
  "terrain.json": "internal/data/terrain.json",
  "looks.json": "internal/origview/looks.json",
};
const rule = (name: string): Promise<unknown> => {
  if (name === "sprites.json") return Promise.resolve({});
  const path = rules[name];
  if (path === undefined) throw Error(`unexpected rule ${name}`);
  return Promise.resolve(JSON.parse(readFileSync(repo + path, "utf8")));
};

describe("picked files", () => {
  const bytes = new Uint8Array(1);
  it("take their place in the game folder by folder and name", () => {
    expect(gamePath({ path: "HL/MAP/05.fin", bytes })).toBe("MAP/05.FIN");
    expect(gamePath({ path: "05.SHP", bytes })).toBe("MAP/05.SHP");
    expect(gamePath({ path: "HL/CODES.DAT", bytes })).toBe("CODES.DAT");
    expect(gamePath({ path: "HL/LIB/PARTS.LIB", bytes })).toBe("LIB/PARTS.LIB");
    expect(gamePath({ path: "HL/UNIT.DAT", bytes })).toBeNull();
    expect(gamePath({ path: "HL/HL.EXE", bytes })).toBeNull();
    expect(gamePath({ path: "HL/SEQ/05.FIN", bytes })).toBeNull();
    // Without folders, the sprite names' UNIT.DAT is told from the unit table by its size.
    expect(gamePath({ path: "UNIT.DAT", bytes })).toBeNull();
    expect(
      gamePath({ path: "UNIT.DAT", bytes: new Uint8Array(8 * 6 * unitTypes) }),
    ).toBe("LIB/UNIT.DAT");
  });
  it("need CODES.DAT and a map", () => {
    expect(select([]).problem).toMatch(/CODES\.DAT/);
    expect(select([{ path: "CODES.DAT", bytes }]).problem).toMatch(/maps/);
    const s = select([
      { path: "CODES.DAT", bytes },
      { path: "MAP/00.FIN", bytes },
    ]);
    expect([s.problem, s.maps, s.art]).toEqual(["", 1, false]);
  });
});

describe("map designs", () => {
  const all = { painted: true, original: true, originalMap: false };
  it("cycle through what a battle has", () => {
    let s = { painted: true, originalArt: false };
    const seen: string[] = [];
    for (let i = 0; i < 4; i++) {
      seen.push(design(s, all));
      s = nextDesign(s, all);
    }
    expect(seen).toEqual(["painted", "tiles", "original", "painted"]);
    const none = { painted: false, original: false, originalMap: false };
    expect(design({ painted: true, originalArt: true }, none)).toBe("tiles");
    expect(nextDesign({ painted: false, originalArt: false }, none)).toEqual({
      painted: false,
      originalArt: false,
    });
  });
  it("show an original map without a painting in the original art", () => {
    const a = { painted: false, original: true, originalMap: true };
    const s = { painted: true, originalArt: false };
    expect(design(s, a)).toBe("original");
    expect(design(nextDesign(s, a), a)).toBe("tiles");
  });
});

describe.skipIf(!local)("the local original files", () => {
  it("decode every tile and unit frame", () => {
    const files = select(picked()).files;
    const art = decodeArt(files);
    expect(art).not.toBeNull();
    for (const season of art?.tiles ?? []) {
      expect(season).toHaveLength(175);
      for (const t of season) {
        expect([t.width, t.height]).toEqual([24, 24]);
        expect(t.rgba.some((v, i) => i % 4 === 3 && v === 255)).toBe(true);
      }
    }
    for (const side of art?.units ?? [])
      for (const frames of side) {
        expect(frames).toHaveLength(6 * unitTypes);
        for (const f of frames)
          expect(f.rgba.some((v, i) => i % 4 === 3 && v === 255)).toBe(true);
      }
    // A plain pick of the same files sorts the same way.
    const plain = picked().map((p) => ({
      path: p.path.split("/").at(-1) ?? "",
      bytes: p.bytes,
    }));
    expect([...select(plain).files.keys()].sort()).toEqual(
      [...files.keys()].sort(),
    );
    expect(tpwm.hasMagic(files.get("CODES.DAT") ?? new Uint8Array())).toBe(
      true,
    );
  });
  it("offer every map as missions that play against the computer", async () => {
    const s = select(picked());
    expect(s.problem).toBe("");
    expect(s.art).toBe(true);
    const data = await loadBattleData({ rule });
    const catalog = originalMissions(
      s.files,
      data,
      await loadOriginalTables({ rule }),
    );
    const maps = new Set(catalog.missions.map((m) => m.month));
    expect(maps.size).toBe(72);
    expect(catalog.skipped).toEqual([]);
    const first = catalog.missions.find((m) => m.key === "original/01:cp");
    expect(first?.next).toBe("original/02:cp");
    expect(catalog.missions.some((m) => m.mode === "2p")).toBe(true);
    if (first === undefined) throw Error("no map 01");
    for (const side of [0, 1] as const) {
      const session = missionSession(
        first,
        side,
        { limit: 0, computer: true },
        data,
      );
      expect(session.game.players[1 - side]?.computer).toBe(true);
      const saved = play.save(session);
      const loaded = missionSession(
        first,
        side,
        { limit: 0, computer: true },
        data,
        saved,
      );
      expect(play.save(loaded)).toEqual(saved);
    }
  });
});
