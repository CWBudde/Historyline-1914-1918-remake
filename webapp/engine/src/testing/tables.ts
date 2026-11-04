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
