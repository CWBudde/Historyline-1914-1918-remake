// Twin of internal/dosrules/gate_move_test.go: TestMovementGate, the R2 gate on each recorded build. Like the
// R1 gate (gate.test.ts) every clause reads the recorded expectations, never the resolver's own results, and
// checks them against the rule the spec states; the replays (motion.test.ts) show that the resolver
// reproduces them. The inputs are the fixtures' own set-up and the profile's tables.

import { readFileSync, readdirSync } from "node:fs";
import { describe, expect, it } from "vitest";
import { type Hold, emptyHold, holdBytes, slotFree } from "./container.ts";
import { at, attempt, errorsIs } from "./goruntime.ts";
import { div, int8, rem } from "./int.ts";
import {
  type ContainerFixture,
  type FixtureRecord,
  type MovementFixture,
  type StepExpected,
  type StepFixture,
  containerMovement,
  fixtureMovement,
  ints,
  movementQuery,
} from "./motionFixtures.ts";
import {
  type MoveTables,
  type Movement,
  liveLen,
  liveRecords,
  moveSeason,
  neighbours64,
  newMovement,
  occupants,
  owns,
  path,
  pieceAt,
  reach,
} from "./move.ts";
import type { Profile } from "./profile.ts";
import { execute } from "./route.ts";
import { ErrUndefined } from "./support.ts";
import { load } from "./testing/fixtures.ts";
import {
  decodeContainerFixture,
  decodeMovementFixture,
  decodeStepFixture,
} from "./testing/motionFixtures.ts";
import { Full, Profiles, moveTables } from "./testing/tables.ts";

// A gate clause counts the behaviours it sees and collects the violations it finds (Go's t.Errorf).
interface Gate {
  seen: Map<string, number>;
  count: (name: string) => void;
  errors: string[];
}

function gate(): Gate {
  const seen = new Map<string, number>();
  return {
    seen,
    count: (name) => seen.set(name, (seen.get(name) ?? 0) + 1),
    errors: [],
  };
}

// need is board_test.go need: no violation, and every named behaviour seen at least once.
function need(g: Gate, ...keys: string[]): void {
  expect(g.errors).toStrictEqual([]);
  for (const k of keys) {
    expect(g.seen.get(k) ?? 0, `no case showed "${k}"`).toBeGreaterThan(0);
  }
}

// cellsOf are the cells of a map's neighbours64 list; on a map one cell wide or high the edge clips
// (else-ifs) leak past the map, and those are left out.
function cellsOf(g: Movement, cell: number): number[] {
  const out: number[] = [];
  for (const n of neighbours64(
    rem(cell, g.width),
    div(cell, g.width),
    g.width,
    g.height,
  )) {
    if (n >= 0 && rem(n, 64) < g.width && div(n, 64) < g.height) {
      out.push(div(n, 64) * g.width + rem(n, 64));
    }
  }
  return out;
}

// closed reports whether the cell's terrain is closed to the moving unit: the masks share no bit, or the
// flags hit the caller's block mask.
function closed(tab: MoveTables, g: Movement, cell: number): boolean {
  const tr = at(tab.terrains, at(g.terrain, cell));
  const typ = at(tab.types, pieceAt(g, g.unit).type);
  return (
    (tr.mask & typ.mask) === 0 ||
    (g.block !== 0xffff && (tr.flags & g.block) !== 0)
  );
}

function season(p: Profile, f: MovementFixture): MoveTables {
  const tab = moveSeason(moveTables(p), f.winter ?? false);
  if (tab === null) {
    throw new Error(`${f.id}: a winter case without a winter table`);
  }
  return tab;
}

function classOf(id: string): string {
  const cut = id.indexOf("/");
  return cut < 0 ? id : id.slice(0, cut);
}

// Costs (*Raw budget vs 2 × speed*): on a map of one terrain with only the mover, each neighbour of the start
// keeps the budget minus the entry cost: terrain byte 3 (of the season's table), byte 4 for an air unit, 2 for
// live +6 bit 0x10, and 80 for closed terrain. The budget of the terrain cases is type byte 0, which is not
// 2 × speed for every type.
function gateCosts(p: Profile, ms: readonly MovementFixture[]): void {
  const v = gate();
  for (const f of ms) {
    if (f.budget >= 128) {
      // The budget is a signed byte in the cost buffer.
      for (const [c, r] of f.expected.remaining.entries()) {
        if (r >= 0) {
          v.errors.push(`${f.id}: budget ${f.budget} reaches cell ${c}`);
        }
      }
      v.count("budget >= 128 reaches nothing");
      continue;
    }
    if (f.cells.length !== 0 || f.units.length !== 1 || f.unseen.length !== 0) {
      continue;
    }
    const tab = season(p, f);
    const g = movementQuery(f);
    const u = pieceAt(g, g.unit);
    const typ = at(tab.types, u.type);
    const tr = at(tab.terrains, f.fill);
    const class_ = classOf(f.id);
    if (class_ === "terrain" || class_ === "winter") {
      if (f.budget !== typ.budget) {
        v.errors.push(
          `${f.id}: budget ${f.budget}, type byte 0 is ${typ.budget}`,
        );
      }
      v.count(
        typ.budget === 2 * typ.speed
          ? "byte 0 is 2 × speed"
          : "byte 0 is not 2 × speed",
      );
    }
    let cost = tr.cost;
    let kind = "ground cost (byte 3)";
    if (closed(tab, g, u.cell)) {
      [cost, kind] = [80, "closed terrain (80)"];
    } else if ((u.flags6 & 0x10) !== 0) {
      [cost, kind] = [2, "flat cost 2"];
    } else if ((u.flags4 & 0x10) !== 0) {
      [cost, kind] = [tr.airCost, "air cost (byte 4)"];
    }
    let want = f.budget - cost;
    if (f.budget === 0 || want < 0) {
      want = -1;
    }
    for (const n of cellsOf(g, u.cell)) {
      const got = at(f.expected.remaining, n);
      if ((want >= 0 && got !== want) || (want < 0 && got >= 0)) {
        v.errors.push(
          `${f.id}: neighbour ${n} keeps ${got}, budget ${f.budget} cost ${cost}`,
        );
      }
    }
    v.count(kind);
    if (
      f.winter === true &&
      tr.cost !== at(moveTables(p).terrains, f.fill).cost
    ) {
      v.count("winter cost differs from summer");
    }
  }
  const keys = [
    "byte 0 is 2 × speed",
    "byte 0 is not 2 × speed",
    "budget >= 128 reaches nothing",
    "ground cost (byte 3)",
    "closed terrain (80)",
    "flat cost 2",
    "air cost (byte 4)",
  ];
  if (moveTables(p).winter !== null) {
    keys.push("winter cost differs from summer");
  }
  need(v, ...keys);
}

// enemy reports whether live flags +4 are another player's unit (a neutral one, bit 2, blocks its cell but
// casts no zone of control).
function enemy(flags: number, side: number): boolean {
  return (flags & 2) === 0 && !owns(flags, side);
}

// Endpoints (*Zone of control*): a destination is reached, is not the unit's own cell and holds no unit of the
// side's map; a cell next to an enemy is entered with 0 left, whatever it costs; own units are passed; a unit
// only on the other side's map is a free cell.
function gateEndpoints(p: Profile, ms: readonly MovementFixture[]): void {
  const v = gate();
  for (const f of ms) {
    const tab = season(p, f);
    const g = movementQuery(f);
    const x = f.expected;
    const start = pieceAt(g, g.unit).cell;
    const occ = occupants(g);
    for (const d of x.destinations ?? []) {
      if (d === start) {
        v.errors.push(`${f.id}: the unit's own cell is a destination`);
      } else if (at(x.remaining, d) < 0) {
        v.errors.push(`${f.id}: destination ${d} was not reached`);
      } else if (occ.has(d)) {
        v.errors.push(
          `${f.id}: destination ${d} holds unit ${occ.get(d) ?? 0}`,
        );
      }
    }
    for (const u of g.unseen.values()) {
      for (const d of x.destinations ?? []) {
        if (d === u.cell) {
          v.count("unseen unit's cell is a destination");
        }
      }
    }
    const zoc = new Set<number>();
    for (const [idx, u] of g.units) {
      if (idx === g.unit) {
        continue;
      }
      if (owns(u.flags4, g.side)) {
        if (at(x.remaining, u.cell) >= 0) {
          v.count("own unit passed");
        }
      } else if (enemy(u.flags4, g.side)) {
        for (const n of cellsOf(g, u.cell)) {
          zoc.add(n);
        }
      }
    }
    for (const n of zoc) {
      if (occ.has(n) || closed(tab, g, n)) {
        continue;
      }
      const r = at(x.remaining, n);
      if (r > 0) {
        v.errors.push(`${f.id}: cell ${n} next to an enemy keeps ${r}`);
      } else if (r === 0) {
        v.count("zone of control entered with 0 left");
      }
    }
  }
  need(
    v,
    "own unit passed",
    "zone of control entered with 0 left",
    "unseen unit's cell is a destination",
  );
}

// Domain: terrain the unit's mask does not share a bit with, or whose flags hit the block mask, is entered only
// from 80 budget up, and with the mask closed it is never a destination; an air unit never ends on terrain flag
// 0x4000.
function gateDomain(p: Profile, ms: readonly MovementFixture[]): void {
  const v = gate();
  for (const f of ms) {
    const tab = season(p, f);
    const g = movementQuery(f);
    const x = f.expected;
    const u = pieceAt(g, g.unit);
    const occ = occupants(g);
    const dest = new Set(x.destinations ?? []);
    for (let c = 0; c < g.width * g.height; c++) {
      if (occ.has(c) || !closed(tab, g, c)) {
        continue;
      }
      const reached = at(x.remaining, c) >= 0;
      if (reached && f.budget < 80) {
        v.errors.push(
          `${f.id}: closed cell ${c} reached on budget ${f.budget}`,
        );
      } else if (reached) {
        v.count("closed cell entered from budget 80");
      } else {
        v.count("closed cell not entered");
      }
      if (
        (at(tab.terrains, at(g.terrain, c)).mask &
          at(tab.types, u.type).mask) ===
        0
      ) {
        if (dest.has(c)) {
          v.errors.push(
            `${f.id}: cell ${c} outside the unit's mask is a destination`,
          );
        } else if (reached) {
          v.count("reached outside the mask, no destination");
        }
      }
    }
    if ((u.flags4 & 0x10) !== 0) {
      for (let c = 0; c < g.width * g.height; c++) {
        if ((at(tab.terrains, at(g.terrain, c)).flags & 0x4000) === 0) {
          continue;
        }
        if (dest.has(c)) {
          v.errors.push(`${f.id}: an air unit may end on flag 0x4000 at ${c}`);
        } else if (at(x.remaining, c) >= 0) {
          v.count("air unit reaches flag 0x4000, no destination");
        }
      }
    }
  }
  need(
    v,
    "closed cell entered from budget 80",
    "closed cell not entered",
    "reached outside the mask, no destination",
    "air unit reaches flag 0x4000, no destination",
  );
}

// codes are the message codes of a recorded outcome.
function codes(x: StepExpected): Set<number> {
  return new Set((x.messages ?? []).map((m) => m[0]));
}

// taken are a side's taken slots of a container record.
function taken(h: Hold, side: number): number[] {
  return h.slots.slice(7 * side, 7 * side + 7).filter((s) => s <= slotFree);
}

function holdOf(g: Movement, idx: number): Hold {
  return g.holds.get(idx) ?? emptyHold();
}

function bytesOf(r: FixtureRecord): number[] {
  return r.bytes ?? [];
}

// Containers: a load into an own transporter is refused with 0x3a when the unit's live +6 shares none of the
// record's classes (+0x19 & 0xfc0) or the transporter has moved, with 0x3b when the side's seven slots are
// taken, a carrying unit's cargo needs more free slots than there are less one, or the weights exceed the
// transporter type's capacity; otherwise the unit takes the side's first free slot. A building never refuses
// with 0x3a or for weight.
function gateContainers(p: Profile, cs: readonly ContainerFixture[]): void {
  const v = gate();
  const tab = moveTables(p);
  for (const f of cs) {
    if (f.order.hold >= 0) {
      continue;
    }
    const g = containerMovement(f);
    const x = f.expected;
    const mover = pieceAt(g, g.unit);
    const side = g.side;
    const goal = f.order.goal;
    const got = codes(x);
    if ((x.order & 1) === 0 && (x.messages ?? []).length === 0) {
      // Refused before the end-cell check: no reach bit on the goal.
      v.count("goal not reached");
      continue;
    }
    let idx = occupants(g).get(goal);
    if (idx === undefined) {
      if (
        (at(tab.terrains, at(g.terrain, goal)).flags &
          (0x400 | 0x100 | 0x40)) !==
        0
      ) {
        if (got.has(0x3a) || got.has(0x40)) {
          v.errors.push(
            `${f.id}: a building refused with ${JSON.stringify(x.messages)}`,
          );
        }
        if (got.has(0x3b)) {
          v.count("building full");
        } else if ((x.order & 1) !== 0) {
          v.count("building entered");
        }
      }
      continue;
    }
    let tr = pieceAt(g, idx);
    if ((tr.flags4 & 0x40) !== 0 && (tr.flags4 & 0x80) === 0) {
      idx--;
      tr = pieceAt(g, idx);
    }
    if (!owns(tr.flags4, side) || (tr.flags4 & 0x1000) === 0) {
      continue;
    }
    const h = holdOf(g, int8(tr.hold));
    const in_ = taken(h, side);
    let weight = at(tab.types, mover.type).weight;
    for (const u of in_) {
      weight += at(tab.types, pieceAt(g, u).type).weight;
    }
    let fits = true;
    if ((mover.flags4 & 0x1000) !== 0) {
      const cargo = taken(holdOf(g, int8(mover.hold)), side);
      for (const u of cargo) {
        weight += at(tab.types, pieceAt(g, u).type).weight;
      }
      fits = 7 - in_.length >= cargo.length + 1;
      v.count("carrying transporter");
    }
    const capacity = at(tab.types, tr.type).capacity;
    let want = 0;
    if ((mover.flags6 & h.flags & 0xfc0) === 0) {
      want = 0x3a;
      v.count("0x3a: no shared class");
    } else if ((tr.flags4 & 0x200) !== 0) {
      want = 0x3a;
      v.count("0x3a: transporter moved");
    } else if (in_.length === 7) {
      want = 0x3b;
      v.count("0x3b: seven slots taken");
    } else if (weight > capacity || !fits) {
      want = 0x3b;
      v.count("0x3b: over capacity");
    } else if (weight === capacity) {
      v.count("loaded at capacity");
    } else {
      v.count("loaded");
    }
    const recorded = `recorded order ${x.order} messages ${JSON.stringify(x.messages)}`;
    if (want !== 0) {
      if (!got.has(want) || (x.order & 1) !== 0) {
        v.errors.push(`${f.id}: ${recorded}, the rule gives ${want}`);
      }
      continue;
    }
    if (got.has(0x3a) || got.has(0x3b) || (x.order & 1) === 0) {
      v.errors.push(`${f.id}: ${recorded}, the rule loads`);
      continue;
    }
    // The unit is in the side's first free slot of the record.
    let slot = in_.length;
    const free = h.slots
      .slice(7 * side, 7 * side + 7)
      .findIndex((s) => s > slotFree);
    if (free >= 0) {
      slot = free;
    }
    const loaded = (x.holds ?? []).some(
      (r) =>
        r.index === int8(tr.hold) && bytesOf(r)[7 * side + slot] === g.unit,
    );
    if (!loaded) {
      v.errors.push(
        `${f.id}: unit ${g.unit} not in slot ${slot} of record ${tr.hold}`,
      );
    }
  }
  need(
    v,
    "loaded",
    "loaded at capacity",
    "carrying transporter",
    "0x3a: no shared class",
    "0x3a: transporter moved",
    "0x3b: seven slots taken",
    "0x3b: over capacity",
    "building full",
    "building entered",
  );
}

// Exits: the exit's reach search starts from budget 4 when the unit's live +6 has bit 0x10, else 8, whatever
// live byte 0 holds.
function gateExits(cs: readonly ContainerFixture[]): void {
  const v = gate();
  for (const f of cs) {
    if (f.order.hold < 0) {
      continue;
    }
    const g = containerMovement(f);
    const u = pieceAt(g, g.unit);
    const want = (u.flags6 & 0x10) !== 0 ? 4 : 8;
    let top = -1;
    for (const r of f.expected.remaining ?? []) {
      top = Math.max(top, r[1]);
    }
    if (top !== want) {
      v.errors.push(
        `${f.id}: the exit reach starts from ${top}, the rule gives ${want}`,
      );
    }
    v.count(`exit budget ${want}`);
    if (f.budget !== want) {
      v.count("live byte 0 not the exit budget");
    }
  }
  need(v, "exit budget 4", "exit budget 8", "live byte 0 not the exit budget");
}

// No consumables: no move, load or exit changes a unit's live byte 0, its experience, group size or alternate
// strength, so there is no fuel, ammo or resupply; a unit leaves a container slot only by the exit or the
// hand-over of a carrying transporter moving into a building.
function gateConsumables(
  ss: readonly StepFixture[],
  cs: readonly ContainerFixture[],
): void {
  const v = gate();
  const check = (
    id: string,
    g: Movement,
    live: readonly FixtureRecord[] | null,
  ): void => {
    const before = liveRecords(g);
    for (const r of live ?? []) {
      const b = before.get(r.index) ?? new Uint8Array(liveLen);
      for (const k of [0, 1, 2, 3]) {
        if (bytesOf(r)[k] !== at(b, k)) {
          v.errors.push(
            `${id}: unit ${r.index} live byte ${k} changed from ${at(b, k)} to ${bytesOf(r)[k] ?? 0}`,
          );
        }
      }
      v.count("live records kept");
    }
  };
  for (const f of ss) {
    const g = fixtureMovement(f);
    check(f.id, g, f.expected.live);
    // A moving transporter rewrites its record's cells; its slots stay.
    const before = ints(holdBytes(g).get(0) ?? new Uint8Array(0)).slice(0, 14);
    if (
      JSON.stringify((f.expected.container ?? []).slice(0, 14)) !==
      JSON.stringify(before)
    ) {
      v.errors.push(`${f.id}: a move changed the slots of container record 0`);
    }
  }
  for (const f of cs) {
    const g = containerMovement(f);
    check(f.id, g, f.expected.live);
    const exit = f.order.hold >= 0;
    const handOver =
      f.order.hold < 0 && (pieceAt(g, g.unit).flags4 & 0x1000) !== 0;
    for (const r of f.expected.holds ?? []) {
      const h = holdOf(g, r.index);
      for (const [i, s] of h.slots.entries()) {
        if (s > slotFree || bytesOf(r)[i] === s) {
          continue;
        }
        if (
          exit &&
          r.index === f.order.hold &&
          i === 7 * g.side + f.order.slot
        ) {
          v.count("exit freed its slot");
        } else if (handOver) {
          v.count("hand-over moved cargo");
        } else {
          v.errors.push(
            `${f.id}: unit ${s} left slot ${i} of record ${r.index}`,
          );
        }
      }
    }
  }
  need(v, "live records kept", "exit freed its slot", "hand-over moved cargo");
}

describe.each(Profiles)("movement gate, %s build", (build, p) => {
  const ms = load(build, "movements.json", decodeMovementFixture);
  const ss = load(build, "steps.json", decodeStepFixture);
  const cs = load(build, "containers.json", decodeContainerFixture);
  it("costs", () => {
    gateCosts(p, ms);
  });
  it("endpoints", () => {
    gateEndpoints(p, ms);
  });
  it("domain", () => {
    gateDomain(p, ms);
  });
  it("containers", () => {
    gateContainers(p, cs);
  });
  it("exits", () => {
    gateExits(cs);
  });
  it("no consumables", () => {
    gateConsumables(ss, cs);
  });
});

describe("movement gate", () => {
  // No foreign rules: the engine's code imports only its own modules, so neither the app nor another package's
  // rules can enter it. (Go's clause also checks that Live names the 26 bytes of the original record; Live is
  // board.go's, ported with board.ts.)
  it("no foreign rules", () => {
    const dir = new URL("./", import.meta.url);
    const files = readdirSync(dir).filter(
      (f) => f.endsWith(".ts") && !f.endsWith(".test.ts"),
    );
    expect(files.length).toBeGreaterThan(0);
    const foreign: string[] = [];
    for (const name of files) {
      const src = readFileSync(new URL(name, dir), "utf8");
      for (const m of src.matchAll(
        /^\s*(?:import|export)\b[^;]*?from\s+"([^"]+)"/gm,
      )) {
        const spec = m[1] ?? "";
        if (!spec.startsWith("./")) {
          foreign.push(`${name} imports ${spec}`);
        }
      }
    }
    expect(foreign).toStrictEqual([]);
  });

  // Gated off: what the original leaves to memory that is not its own stays undefined in the profile. A path
  // search that runs dry reads past the reach-bit array (the demo's image then refused with 0x48, the full
  // build's walked), and on a map one row high the path search reads past the map; both give ErrUndefined.
  // Go runs them through Profile.Move and Profile.Route (board.go, L4): Move is Execute, Route the path search
  // after the reach search.
  it("gated off", () => {
    const tab = moveTables(Full);
    const infantry = 0;
    const open = 121;
    const deep = 47;
    const typ = at(tab.types, infantry);
    const piece = {
      type: infantry,
      cell: 40,
      flags4: typ.class,
      flags6: typ.flags,
      hold: 0,
    };
    const islands = newMovement({
      width: 9,
      height: 9,
      terrain: Array.from({ length: 81 }, () => deep),
      units: new Map([[0, piece]]),
      unit: 0,
      budget: typ.budget,
      block: 3,
    });
    islands.terrain[40] = open;
    islands.terrain[0] = open;
    const [, moveErr] = attempt(() =>
      execute(tab, islands, { goal: 0, retarget: -1, stale: [0] }),
    );
    expect(
      errorsIs(moveErr, ErrUndefined),
      `a dry path search gives ${String(moveErr)}`,
    ).toBe(true);

    const row = newMovement({
      width: 9,
      height: 1,
      terrain: Array.from({ length: 9 }, () => open),
      units: new Map([[0, { ...piece, cell: 4 }]]),
      unit: 0,
      budget: typ.budget,
      block: 3,
    });
    const [, routeErr] = attempt(() => {
      const [buf, bits] = reach(tab, row);
      return path(tab, row, buf, bits, 0);
    });
    expect(
      errorsIs(routeErr, ErrUndefined),
      `a route on a map one row high gives ${String(routeErr)}`,
    ).toBe(true);
  });
});
