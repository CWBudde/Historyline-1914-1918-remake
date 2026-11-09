// Package dosrules (internal/dosrules) is the pure resolver of the original DOS rules (profile dos-0.1): the
// casualty routine, the pincer and blocking support, the battle distance, the experience awards and the order
// handler that composes them; the movement queries, the move order and its steps, loading, unloading and
// exits, and the phase change that brings both sides' views into line, for two rule tables: Full (the full
// game, the build of record) and Demo. It holds only the numeric fields the rules read; the fixtures under
// internal/dosrules/testdata/ pin its results.
//
// Go embeds full.json and demo.json; the engine imports no files, so the caller passes their parsed JSON in
// (mustTables, then fullProfile and demoProfile, or profilesOf for both).

import {
  array,
  arrayField,
  int,
  intField,
  object,
  present,
  stringField,
} from "./decode.ts";
import type { MoveTables, MoveTerrain, MoveType } from "./move.ts";

/** Type is the part of a UNIT.DAT type record that combat and production read. */
export interface Type {
  /** +2. */
  armour: number;
  /** +3, full group strength. */
  group: number;
  /** +6 word: classes 0x3c, long range 0x40, never returns fire 0x80. */
  targetFlags: number;
  /** +8, encoded (one more than shown). */
  airRange: number;
  /** +9, encoded. */
  surfaceRange: number;
  /** +10. */
  airPower: number;
  /** +11. */
  groundPower: number;
  /** +12. */
  waterPower: number;
  /** +0x0d word: target classes it attacks. */
  attackMask: number;
  /** +0x0f word: bit 0 holds fire after a move, bit 2 counts the alternate strength. */
  strengthFlags: number;
  /** +0x3f, the production price; 255: never sold. */
  cost: number;
}

/**
 * alternate reports whether the unit's strength is its alternate byte. No shipped type enables it; synthetic
 * casualty fixtures cover the branch.
 */
export function alternate(t: Type): boolean {
  return (t.strengthFlags & 4) !== 0;
}

/**
 * setsUp reports whether the unit cannot attack in the action phase after it moved (medium, heavy and train
 * artillery): it must set up first.
 */
export function setsUp(t: Type): boolean {
  return (t.strengthFlags & 1) !== 0;
}

/** Terms are the signed attack and defence terms of one terrain class. */
export interface Terms {
  attack: number;
  defence: number;
}

/**
 * Tables are the numbers one build supplies: its 51 type records and the terms of the nine terrain classes,
 * and the fields of the type and terrain records the movement queries read (null where the build's movement
 * code is not located).
 */
export interface Tables {
  /** SHA-256 prefix of the executable. */
  build: string;
  types: Type[];
  /** Always nine, one per terrain class. */
  terms: Terms[];
  move: MoveTables | null;
}

/** The terrain classes, the length of Tables.terms (Go's [9]Terms). */
export const terrainClasses = 9;

/**
 * Profile is a rules profile bound to one build's tables. Builds that turn out to differ in a rule get a field
 * here, not a hidden constant. Go copies it by value; nothing here changes one.
 */
export interface Profile {
  readonly name: string;
  readonly tables: Tables | null;
  /**
   * TransposedDistance passes (row, column) to the hex distance, as the casualty routine does: one of the six
   * adjacent directions then counts as distance 2 and ranged distances are off by up to two. Without it the
   * distance is called with (column, row), as the movement code does, and is the layout distance.
   */
  readonly transposedDistance: boolean;
  /**
   * Builders are the construction units, whose live +0x0a holds the builder charges. A repair costs Repair
   * energy, BuilderRepair for a construction unit.
   */
  readonly builders: readonly number[];
  readonly repair: number;
  readonly builderRepair: number;
  /** Season(true) chose the winter terrain records (unexported in Go). */
  readonly winter: boolean;
}

/** newProfile is a Profile literal: the fields not given are Go's zero values. */
export function newProfile(fields: Partial<Profile>): Profile {
  return {
    name: "",
    tables: null,
    transposedDistance: false,
    builders: [],
    repair: 0,
    builderRepair: 0,
    winter: false,
    ...fields,
  };
}

/** tablesOf is p.Tables, which Go dereferences without a check: a profile without tables panics. */
export function tablesOf(p: Profile): Tables {
  if (p.tables === null) {
    throw new TypeError(
      "runtime error: invalid memory address or nil pointer dereference",
    );
  }
  return p.tables;
}

/** fullProfile is Full, profile dos-0.1 on the full build, as executed: the build of record. */
export function fullProfile(tables: Tables): Profile {
  return newProfile({
    name: "dos-0.1",
    tables,
    transposedDistance: true,
    builders: [16],
    repair: 15,
    builderRepair: 50,
  });
}

/**
 * demoProfile is Demo, profile dos-0.1 on the demo build, as executed. Every recorded case gives what the full
 * build gives; a rule found to differ becomes a Profile field set here.
 */
export function demoProfile(tables: Tables): Profile {
  return newProfile({
    name: "dos-0.1",
    tables,
    transposedDistance: true,
    builders: [16],
    repair: 15,
    builderRepair: 50,
  });
}

/** Builds are the identified builds by name (the directories under testdata/). */
export type Build = "full" | "demo";

/**
 * Profiles are the profiles a save may name: the identified builds and, where the caller has the extension's
 * tables, the extended profile (Go's Full, Demo and Extended).
 */
export type Profiles = Readonly<Record<Build, Profile>> & {
  readonly extended?: Profile;
};

/**
 * profilesOf is Profiles: the profiles of the identified builds by build name, and the extended profile on the
 * full tables when the parsed extension.json is given.
 */
export function profilesOf(
  full: Tables,
  demo: Tables,
  extension?: unknown,
): Profiles {
  const p = { full: fullProfile(full), demo: demoProfile(demo) };
  return extension === undefined
    ? p
    : { ...p, extended: extendedProfile(extendTables(full, extension)) };
}

/**
 * extendTables is base with the type records of the extension (the parsed extension.json) appended from index
 * 51: the original records keep their indices. The build is the base's with the extension's suffix (Go's
 * ExtendedTables).
 */
export function extendTables(base: Tables, extension: unknown): Tables {
  const x = mustTables(extension);
  const move =
    base.move !== null && x.move !== null
      ? { ...base.move, types: [...base.move.types, ...x.move.types] }
      : base.move;
  return {
    ...base,
    build: base.build + x.build,
    types: [...base.types, ...x.types],
    move,
  };
}

/** extendedProfile is Extended: the full profile on the extended tables, the original rules plus the new units' (gas.ts). */
export function extendedProfile(tables: Tables): Profile {
  return newProfile({
    name: "dos-0.1+units1",
    tables,
    transposedDistance: true,
    builders: [16],
    repair: 15,
    builderRepair: 50,
  });
}

/**
 * mustTables reads one build's tables from the parsed JSON of full.json or demo.json, as Go's json.Unmarshal
 * fills Tables: absent fields are zero, unknown fields are ignored, the terms are [attack, defence] pairs. It
 * throws a DecodeError where the JSON does not fit.
 */
export function mustTables(json: unknown): Tables {
  const o = object(json, "tables");
  const terms: Terms[] = [];
  for (const [i, v] of (arrayField(o, "terms", "tables") ?? []).entries()) {
    if (i < terrainClasses) {
      // Go fills a [9]Terms and drops elements past it.
      terms.push(decodeTerms(v, `tables.terms[${i}]`));
    }
  }
  while (terms.length < terrainClasses) {
    terms.push({ attack: 0, defence: 0 });
  }
  return {
    build: stringField(o, "build", "tables"),
    types: (arrayField(o, "types", "tables") ?? []).map((v, i) =>
      decodeType(v, `tables.types[${i}]`),
    ),
    terms,
    move: present(o, "move") ? decodeMoveTables(o["move"]) : null,
  };
}

/** The keys of a Type in JSON, for a strict read (decodeType with strict). */
const typeKeys = [
  "armour",
  "group",
  "targetFlags",
  "airRange",
  "surfaceRange",
  "airPower",
  "groundPower",
  "waterPower",
  "attackMask",
  "strengthFlags",
  "cost",
] as const;

/** decodeType reads a type record; strict refuses unknown fields (Go's DisallowUnknownFields). */
export function decodeType(v: unknown, what: string, strict = false): Type {
  const o = object(v, what, strict ? typeKeys : undefined);
  const n = (k: (typeof typeKeys)[number]): number => intField(o, k, what);
  return {
    armour: n("armour"),
    group: n("group"),
    targetFlags: n("targetFlags"),
    airRange: n("airRange"),
    surfaceRange: n("surfaceRange"),
    airPower: n("airPower"),
    groundPower: n("groundPower"),
    waterPower: n("waterPower"),
    attackMask: n("attackMask"),
    strengthFlags: n("strengthFlags"),
    cost: n("cost"),
  };
}

// Terms.UnmarshalJSON reads the terms as an [attack, defence] pair (a [2]int: null and missing elements are
// zero, elements past two are dropped).
function decodeTerms(v: unknown, what: string): Terms {
  if (v === null) {
    return { attack: 0, defence: 0 };
  }
  const p = array(v, what);
  const elem = (i: number): number => {
    const x = p[i];
    return x === undefined || x === null ? 0 : int(x, `${what}[${i}]`);
  };
  return { attack: elem(0), defence: elem(1) };
}

function decodeMoveTables(v: unknown): MoveTables {
  const what = "tables.move";
  const o = object(v, what);
  const winter = arrayField(o, "winter", what);
  return {
    types: (arrayField(o, "types", what) ?? []).map((t, i) =>
      decodeMoveType(t, `${what}.types[${i}]`),
    ),
    terrains: (arrayField(o, "terrains", what) ?? []).map((t, i) =>
      decodeMoveTerrain(t, `${what}.terrains[${i}]`),
    ),
    winter:
      winter === null
        ? null
        : winter.map((t, i) => decodeMoveTerrain(t, `${what}.winter[${i}]`)),
  };
}

function decodeMoveType(v: unknown, what: string): MoveType {
  const o = object(v, what);
  const n = (k: keyof MoveType): number => intField(o, k, what);
  return {
    budget: n("budget"),
    speed: n("speed"),
    mask: n("mask"),
    class: n("class"),
    flags: n("flags"),
    distanceWeight: n("distanceWeight"),
    costWeight: n("costWeight"),
    capacity: n("capacity"),
    weight: n("weight"),
    carries: n("carries"),
  };
}

function decodeMoveTerrain(v: unknown, what: string): MoveTerrain {
  const o = object(v, what);
  const n = (k: keyof MoveTerrain): number => intField(o, k, what);
  return {
    flags: n("flags"),
    mask: n("mask"),
    cost: n("cost"),
    airCost: n("airCost"),
    class: n("class"),
  };
}
