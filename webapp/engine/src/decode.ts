// Reading parsed JSON into the engine's records the way Go's encoding/json fills structs: an absent or null
// field is the zero value, a field of the wrong kind is an error, and a strict reader (Go's
// DisallowUnknownFields) refuses keys the record does not know.

import { GoError } from "./goruntime.ts";

export type JSONObject = Readonly<Record<string, unknown>>;

/** DecodeError is a JSON value that does not fit the record read from it. */
export class DecodeError extends GoError {}

/** object reads v as an object; with known, every key must be one of them. */
export function object(
  v: unknown,
  what: string,
  known?: readonly string[],
): JSONObject {
  if (typeof v !== "object" || v === null || Array.isArray(v)) {
    throw new DecodeError(`${what}: not an object`);
  }
  const o = v as JSONObject;
  if (known !== undefined) {
    for (const k of Object.keys(o)) {
      if (!known.includes(k)) {
        throw new DecodeError(`${what}: unknown field "${k}"`);
      }
    }
  }
  return o;
}

/** int reads v as an integer. */
export function int(v: unknown, what: string): number {
  if (typeof v !== "number" || !Number.isSafeInteger(v)) {
    throw new DecodeError(`${what}: not an integer`);
  }
  return v + 0; // no -0
}

/** array reads v as an array. */
export function array(v: unknown, what: string): readonly unknown[] {
  if (!Array.isArray(v)) {
    throw new DecodeError(`${what}: not an array`);
  }
  return v as readonly unknown[];
}

/** present reports whether o has the field key with a value other than null. */
export function present(o: JSONObject, key: string): boolean {
  return o[key] !== undefined && o[key] !== null;
}

/** intField is o[key] as an integer, 0 when absent. */
export function intField(o: JSONObject, key: string, what: string): number {
  return present(o, key) ? int(o[key], `${what}.${key}`) : 0;
}

/** boolField is o[key] as a boolean, false when absent. */
export function boolField(o: JSONObject, key: string, what: string): boolean {
  if (!present(o, key)) {
    return false;
  }
  const v = o[key];
  if (typeof v !== "boolean") {
    throw new DecodeError(`${what}.${key}: not a boolean`);
  }
  return v;
}

/** stringField is o[key] as a string, "" when absent. */
export function stringField(o: JSONObject, key: string, what: string): string {
  if (!present(o, key)) {
    return "";
  }
  const v = o[key];
  if (typeof v !== "string") {
    throw new DecodeError(`${what}.${key}: not a string`);
  }
  return v;
}

/** arrayField is o[key] as an array, null (Go's nil slice) when absent. */
export function arrayField(
  o: JSONObject,
  key: string,
  what: string,
): readonly unknown[] | null {
  return present(o, key) ? array(o[key], `${what}.${key}`) : null;
}
