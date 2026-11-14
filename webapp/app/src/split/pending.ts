import { play } from "@historyline/engine";
import { decodePendingBefore, type SavePending } from "../storage/store.ts";

/**
 * A hot-seat resolution one human has not seen yet, as a save keeps it (internal/saves.Pending): the side, the
 * game just before the resolution and its events. The events are written in Go's JSON of internal/dosplay
 * events, the encoding the Go golden traces compare against the engine (engine/src/playGolden.test.ts): Go's
 * field names (ID and HQ in capitals), the rules events' tagged fields in lower case, nil lists as null and
 * the rules events' zero fields left out. The desktop game writes the same, so each edition plays the other's.
 */
export interface PendingPlayback {
  side: 0 | 1;
  /** The game just before the resolution: play.save's bytes. */
  before: Uint8Array;
  /** The side's view of that game, shown until the playback has run. */
  view: play.View;
  events: play.Event[];
}

// The keys whose values Go writes with its lower-case JSON tags: a rules event and an end.
const tagged = new Set(["raw", "end"]);
// Within a rules event, the values written with Go's field names again.
const untagged = new Set(["outcome", "messages"]);
// Go's nil lists, which the engine keeps empty: an event's and its rules event's (a unit's rear is a hex or null).
const eventLists = ["path", "rear", "hexes"],
  rawLists = ["route", "messages", "cells"];

function goName(key: string): string {
  return key === "id"
    ? "ID"
    : key === "hq"
      ? "HQ"
      : key.charAt(0).toUpperCase() + key.slice(1);
}

function engineName(key: string): string {
  return key === "ID"
    ? "id"
    : key === "HQ"
      ? "hq"
      : key.charAt(0).toLowerCase() + key.slice(1);
}

function toGo(value: unknown, inTag: boolean): unknown {
  if (Array.isArray(value)) return value.map((v) => toGo(v, inTag));
  if (typeof value !== "object" || value === null) return value;
  return Object.fromEntries(
    Object.entries(value).map(([key, v]) => {
      const nested = inTag ? !untagged.has(key) : tagged.has(key);
      return [inTag ? key : goName(key), toGo(v, nested)];
    }),
  );
}

function fromGo(value: unknown): unknown {
  if (Array.isArray(value)) return value.map((v) => fromGo(v));
  if (typeof value !== "object" || value === null) return value;
  return Object.fromEntries(
    Object.entries(value).map(([k, v]) => [engineName(k), fromGo(v)]),
  );
}

/** emptyLists makes the null values of o's keys empty lists. */
function emptyLists(o: Record<string, unknown>, keys: readonly string[]) {
  for (const key of keys) if (o[key] === null) o[key] = [];
}

/** goEvents writes session events in Go's JSON of dosplay events. */
export function goEvents(
  events: readonly play.Event[],
): Record<string, unknown>[] {
  return events.map((e) => toGo(e, false) as Record<string, unknown>);
}

function record(v: unknown, what: string): Record<string, unknown> {
  if (typeof v !== "object" || v === null || Array.isArray(v))
    throw Error(`pending playback: ${what} is no object`);
  return v as Record<string, unknown>;
}

function integer(v: unknown, what: string): void {
  if (!Number.isSafeInteger(v))
    throw Error(`pending playback: ${what} is no integer`);
}

function coord(v: unknown, what: string): void {
  const c = record(v, what);
  integer(c["line"], `${what} line`);
  integer(c["col"], `${what} col`);
}

function coords(v: unknown, what: string): void {
  if (!Array.isArray(v)) throw Error(`pending playback: ${what} is no list`);
  for (const c of v) coord(c, what);
}

/** playEvents reads events in Go's JSON of dosplay events; it throws on a malformed one. */
export function playEvents(values: readonly unknown[]): play.Event[] {
  return values.map((value, i) => {
    const e = record(fromGo(value), `event ${i}`);
    integer(e["kind"], "kind");
    integer(e["side"] ?? 0, "side");
    integer(e["unit"] ?? 0, "unit");
    for (const key of ["from", "to"]) coord(e[key] ?? { line: 0, col: 0 }, key);
    for (const key of ["path", "rear", "hexes"]) coords(e[key] ?? [], key);
    if (e["battle"] !== undefined && e["battle"] !== null) {
      const b = record(e["battle"], "battle");
      for (const side of ["attacker", "defender"]) {
        const f = record(b[side], side);
        record(f["unit"], `${side} unit`);
        coord(record(f["unit"], `${side} unit`)["at"], `${side} unit at`);
      }
    }
    emptyLists(e, eventLists);
    const raw = record(e["raw"] ?? {}, "raw");
    emptyLists(raw, rawLists);
    if (raw["outcome"] !== undefined && raw["outcome"] !== null)
      emptyLists(record(raw["outcome"], "outcome"), ["destroyed"]);
    return play.newEvent({
      ...(e as Partial<play.Event>),
      raw: { ...play.zeroRawEvent(), ...(raw as Partial<play.RawEvent>) },
    });
  });
}

/**
 * restorePending reads a saved pending playback: its game through load, the side's view of it and the events.
 * One that does not read is null; the battle goes on without its playback.
 */
export function restorePending(
  saved: SavePending,
  load: (before: Uint8Array) => play.Session,
): PendingPlayback | null {
  try {
    const before = decodePendingBefore(saved),
      events = playEvents(saved.events);
    return {
      side: saved.side,
      before,
      view: play.view(load(before), saved.side),
      events,
    };
  } catch {
    return null;
  }
}
