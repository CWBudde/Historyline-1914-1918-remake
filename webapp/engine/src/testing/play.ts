import { object } from "../decode.ts";
import { bindingData } from "../bind.ts";
import { decode, level, type Mode } from "../scenario.ts";
import { readJSON } from "./files.ts";

export const playData = bindingData(
  readJSON("internal/data/units.json"),
  readJSON("internal/dosbind/records.json"),
);
const terrain = new Set(
  Object.keys(object(readJSON("internal/data/terrain.json"), "terrain")),
);
export function shippedLevel(month: string, mode: Mode = "cp") {
  return level(
    decode(
      new Map(
        ["map.json", "cp.json", "ec.json", "2p.json"].map((name) => [
          name,
          readJSON(`levels/${month}/${name}`),
        ]),
      ),
      terrain,
    ),
    mode,
  );
}

// Convert Go's public field names and nil slices to the port's value representation.
export function goValue(v: unknown, key = ""): unknown {
  if (
    v === null &&
    [
      "path",
      "rear",
      "hexes",
      "route",
      "messages",
      "cells",
      "destroyed",
      "units",
    ].includes(key)
  )
    return [];
  if (Array.isArray(v)) return v.map((x: unknown) => goValue(x));
  if (typeof v !== "object" || v === null) return v;
  return Object.fromEntries(
    Object.entries(v).map(([k, x]) => {
      const name =
        k === "ID"
          ? "id"
          : k === "HQ"
            ? "hq"
            : k.charAt(0).toLowerCase() + k.slice(1);
      return [name, goValue(x, name)];
    }),
  );
}
