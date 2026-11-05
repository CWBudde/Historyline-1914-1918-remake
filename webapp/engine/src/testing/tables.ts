// Test-only: both builds' tables and profiles, read from internal/dosrules/{full,demo}.json as Go embeds them
// (Go's FullTables, DemoTables, Full, Demo and Profiles).

import {
  type Build,
  type Profile,
  type Tables,
  demoProfile,
  fullProfile,
  mustTables,
} from "../profile.ts";
import type { MoveTables } from "../move.ts";
import { readJSON } from "./files.ts";

/** FullTables are the tables of the full game, the build of record. */
export const FullTables: Tables = mustTables(
  readJSON("internal/dosrules/full.json"),
);

/** DemoTables are the tables of HL14-18.EXE of the free DOS demo. */
export const DemoTables: Tables = mustTables(
  readJSON("internal/dosrules/demo.json"),
);

/** Full is profile dos-0.1 on the full build. */
export const Full: Profile = fullProfile(FullTables);

/** Demo is profile dos-0.1 on the demo build. */
export const Demo: Profile = demoProfile(DemoTables);

/** Profiles are the profiles of the identified builds by build name, in a fixed order. */
export const Profiles: readonly (readonly [Build, Profile])[] = [
  ["demo", Demo],
  ["full", Full],
];

/** Builds are the profiles by build name, the profiles loadGame knows (Go's Full and Demo). */
export const Builds: Readonly<Record<Build, Profile>> = {
  full: Full,
  demo: Demo,
};

/** cInfantry and cPanzer are the infantry and panzer types the scene tests place (cycle_test.go). */
export const cInfantry = 18;
export const cPanzer = 47;
/** cSupply is a ground transporter (cycle_test.go). */
export const cSupply = 22;
/** cTransportRec is the container record the scenes' transporters take (cycle_test.go). */
export const cTransportRec = 3;
/** cBuilder is the construction unit (energy_test.go). */
export const cBuilder = 16;
/** cSapper is the sapper (sapper_test.go). */
export const cSapper = 20;

/** moveTables is motion_test.go moveTables: the profile's movement tables, which must hold all 51 types. */
export function moveTables(p: Profile): MoveTables {
  const m = p.tables?.move ?? null;
  if (m === null || m.types.length !== 51 || m.terrains.length === 0) {
    throw new Error(`${p.tables?.build ?? ""}: no movement tables`);
  }
  return m;
}
