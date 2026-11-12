// The original game's files the player picks (a folder, several files or a drop), sorted into the paths the
// engine reads ("CODES.DAT", "MAP/05.FIN", "LIB/PARTS.LIB"). Everything else is left out: nothing but these
// files is kept in the browser, and nothing is uploaded.
import { tpwm } from "@historyline/engine";
import { artFiles, unitTypes } from "./art.ts";

export interface Picked {
  /** The path the browser gives (webkitRelativePath or a dropped entry's path), else the file name. */
  path: string;
  bytes: Uint8Array;
}

const art = new Set<string>(artFiles);

/** gamePath is the path a picked file takes in the game folder, or null for a file the game does not need here. */
export function gamePath(p: Picked): string | null {
  const parts = p.path.replace(/\\/g, "/").toUpperCase().split("/");
  const file = parts.at(-1) ?? "",
    dir = parts.length > 1 ? (parts.at(-2) ?? "") : "";
  if (/^[0-9A-Z_]{1,8}\.(FIN|SHP|COM)$/.test(file))
    return dir === "" || dir === "MAP" ? `MAP/${file}` : null;
  if (file === "CODES.DAT" || file === "00.PAL")
    return dir === "MAP" || dir === "LIB" ? null : file;
  if (!art.has(`LIB/${file}`)) return null;
  if (dir === "LIB") return `LIB/${file}`;
  if (dir !== "") return null;
  // A plain pick has no folders: the game folder holds another UNIT.DAT (the unit table), told apart by size.
  if (file === "UNIT.DAT") {
    try {
      if (tpwm.unpack(p.bytes).length !== 8 * 6 * unitTypes) return null;
    } catch {
      return null;
    }
  }
  return `LIB/${file}`;
}

export interface Selection {
  files: Map<string, Uint8Array>;
  maps: number;
  art: boolean;
  /** Why the selection cannot be imported; "" when it can. */
  problem: string;
}

/** select sorts picked files; a selection needs CODES.DAT and at least one map. */
export function select(picked: readonly Picked[]): Selection {
  const files = new Map<string, Uint8Array>();
  for (const p of picked) {
    const path = gamePath(p);
    if (path !== null && !files.has(path)) files.set(path, p.bytes);
  }
  const maps = [...files.keys()].filter((k) => k.endsWith(".FIN")).length;
  const problem = !files.has("CODES.DAT")
    ? "CODES.DAT is missing: pick the game's folder (the one with MAP and LIB) or its files."
    : maps === 0
      ? "No maps found: the MAP folder's .FIN files are missing."
      : "";
  return {
    files,
    maps,
    art: artFiles.every((f) => files.has(f)),
    problem,
  };
}
