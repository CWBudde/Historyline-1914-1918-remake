// internal/origmaps: the maps of the original game's files, which the player supplies (the repository ships none
// of them), decoded into finmap levels: each map's MAP/*.FIN with its .SHP and .COM files and its CODES.DAT
// record. Files are keyed by their path in the game folder in capitals ("MAP/05.FIN", "CODES.DAT").
import * as finmap from "./finmap.ts";
import { GoError } from "./goruntime.ts";
import { type MoveTables } from "./move.ts";
import { unpack } from "./tpwm.ts";

export type Files = ReadonlyMap<string, Uint8Array>;

/** mapFiles are the paths of the files a map reads, in the game folder. */
export function mapFiles(files: Files): string[] {
  return [...files.keys()].filter((k) => /^MAP\/[^/]+\.FIN$/.test(k)).sort();
}

function unpacked(files: Files, path: string): Uint8Array | undefined {
  const b = files.get(path);
  if (b === undefined) return undefined;
  try {
    return unpack(b);
  } catch (e) {
    throw new GoError(`${path}: ${e instanceof Error ? e.message : String(e)}`);
  }
}

/**
 * read decodes every map of the full game in file name order (Go's origmaps.Read with the full build). It needs
 * CODES.DAT and at least one FIN file; a map's SHP and COM files are optional, as in the game.
 */
export function read(files: Files, tables: MoveTables): finmap.Level[] {
  const r = finmap.rulesFrom(tables);
  const codesRaw = unpacked(files, "CODES.DAT");
  if (codesRaw === undefined) throw new GoError("full: no CODES.DAT");
  const codes = finmap.decodeCodes(codesRaw);
  const fins = mapFiles(files);
  if (fins.length === 0) throw new GoError("full: no maps in MAP");
  return fins.map((fin) => {
    try {
      return level(files, fin, codes, r);
    } catch (e) {
      throw new GoError(`full: ${e instanceof Error ? e.message : String(e)}`);
    }
  });
}

function level(
  files: Files,
  fin: string,
  codes: finmap.Code[],
  r: finmap.Rules,
): finmap.Level {
  const stem = fin.slice("MAP/".length, -".FIN".length);
  const raw = unpacked(files, fin);
  if (raw === undefined) throw new GoError(`${fin}: missing`);
  const prefix = (e: unknown): never => {
    throw new GoError(`${stem}: ${e instanceof Error ? e.message : String(e)}`);
  };
  const input: finmap.Input = { index: -1, map: finmap.decodeMap(raw) };
  const used = [fin];
  const shop = unpacked(files, `MAP/${stem}.SHP`);
  if (shop !== undefined) {
    try {
      input.shop = finmap.decodeShop(shop);
    } catch (e) {
      prefix(e);
    }
    used.push(`MAP/${stem}.SHP`);
  }
  const com = unpacked(files, `MAP/${stem}.COM`);
  if (com !== undefined) {
    try {
      input.computer = finmap.decodeComputer(com);
    } catch (e) {
      prefix(e);
    }
    used.push(`MAP/${stem}.COM`);
  }
  if (/^[+-]?\d+$/.test(stem) && Number(stem) >= 0) {
    input.index = Number(stem);
    const code = codes[input.index];
    if (code !== undefined) {
      input.code = code;
      used.push("CODES.DAT");
    }
  }
  let l: finmap.Level;
  try {
    l = finmap.build(input, r);
  } catch (e) {
    return prefix(e);
  }
  const next = l.mission?.next ?? null;
  if (next !== null && next >= codes.length)
    throw new GoError(
      `${stem}: next map ${next}, CODES.DAT has ${codes.length}`,
    );
  l.source = { build: "full", map: stem, files: used };
  return l;
}
