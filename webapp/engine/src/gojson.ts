// Go's encoding/json as the saves need it (no Go file of its own: save.go calls json.Marshal and json.Unmarshal).
//
// Writing: json.Marshal's bytes for ints, bools, strings, arrays, int-keyed maps and structs: compact, fields in
// declaration order (the caller drops omitempty fields), map keys sorted by their decimal string, strings with
// <, >, & escaped as \u003c etc., \b and \f short (go1.22 and later), U+2028/U+2029 escaped and invalid UTF-8 as
// \ufffd.
//
// Reading: json.Unmarshal into typed records (codecs). The bytes are checked as Go's scanner does, with its
// messages; strings are unquoted with Go's rules (each invalid UTF-8 byte a U+FFFD, a lone surrogate escape a
// U+FFFD); then the tree is decoded in input order into what the record holds: absent fields keep their zero
// value, unknown fields are skipped unchecked, keys match the json names case-insensitively (the last one wins),
// null keeps a value (but clears a pointer, map or slice), a repeated object merges into the map or struct
// there, array elements past a fixed array's length are skipped unchecked and missing ones zeroed. The first
// type error ends the read with Go's message (Go records it and reads on, but the result is the same error).
//
// Deviation: Go's int is 64 bits; the port holds integers exactly only up to 2^53, so an integer past that is
// refused with its own message where Go would read it.

import { GoError } from "./goruntime.ts";

/** JSONSyntaxError is Go's *json.SyntaxError: the bytes are not JSON. */
export class JSONSyntaxError extends GoError {}

/** UnmarshalTypeError is Go's *json.UnmarshalTypeError: a JSON value that does not fit the Go type. */
export class UnmarshalTypeError extends GoError {}

/** JSONNode is a parsed JSON value that keeps what Go reads: number literals, key order and repeated keys. */
export type JSONNode =
  | { readonly kind: "null" }
  | { readonly kind: "bool"; readonly value: boolean }
  | { readonly kind: "number"; readonly literal: string }
  | { readonly kind: "string"; readonly value: string }
  | { readonly kind: "array"; readonly items: JSONNode[] }
  | { readonly kind: "object"; readonly entries: [string, JSONNode][] };

// ---- writing ----

const hex = "0123456789abcdef";

/** encodeString is a string as json.Marshal writes it (with HTML escaping). */
export function encodeString(s: string): string {
  let out = '"';
  for (const ch of s) {
    const c = ch.codePointAt(0) ?? 0;
    if (c < 0x80) {
      switch (c) {
        case 0x22:
          out += '\\"';
          break;
        case 0x5c:
          out += "\\\\";
          break;
        case 0x08:
          out += "\\b";
          break;
        case 0x0c:
          out += "\\f";
          break;
        case 0x0a:
          out += "\\n";
          break;
        case 0x0d:
          out += "\\r";
          break;
        case 0x09:
          out += "\\t";
          break;
        default:
          if (c < 0x20 || c === 0x3c || c === 0x3e || c === 0x26) {
            out += `\\u00${hex.charAt(c >> 4)}${hex.charAt(c & 0xf)}`;
          } else {
            out += ch;
          }
      }
    } else if (c >= 0xd800 && c < 0xe000) {
      out += "\\ufffd"; // a lone surrogate: no UTF-8, as Go writes invalid bytes
    } else if (c === 0x2028 || c === 0x2029) {
      out += `\\u${c.toString(16)}`;
    } else {
      out += ch;
    }
  }
  return out + '"';
}

/** encodeInt is a Go int as json.Marshal writes it; it throws a TypeError for a number that is no exact integer. */
export function encodeInt(n: number): string {
  if (!Number.isSafeInteger(n)) {
    throw new TypeError(`gojson: ${String(n)} is no integer`);
  }
  return String(n);
}

/** encodeBool is a bool as json.Marshal writes it. */
export function encodeBool(b: boolean): string {
  return b ? "true" : "false";
}

/** encodeArray is a JSON array of encoded elements. */
export function encodeArray(items: readonly string[]): string {
  return `[${items.join(",")}]`;
}

/** encodeInts is a Go [n]int: n elements, 0 where a has none. */
export function encodeInts(a: readonly number[], n: number): string {
  return encodeArray(Array.from({ length: n }, (_, i) => encodeInt(a[i] ?? 0)));
}

/** encodeObject is a struct of encoded field values, in the order given. */
export function encodeObject(
  fields: readonly (readonly [string, string])[],
): string {
  return `{${fields.map(([k, v]) => `${encodeString(k)}:${v}`).join(",")}}`;
}

/** encodeIntMap is a Go map[int]T: its keys as strings, sorted as strings ("10" before "2"). */
export function encodeIntMap<T>(
  m: ReadonlyMap<number, T>,
  enc: (v: T) => string,
): string {
  const entries = [...m].map(([k, v]) => [encodeInt(k), v] as const);
  entries.sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0));
  return encodeObject(entries.map(([k, v]) => [k, enc(v)] as const));
}

// isPrint is Go's strconv.IsPrint, by the JS engine's Unicode tables: letters, marks, numbers, punctuation,
// symbols and the ASCII space.
const graphic = /^[\p{L}\p{M}\p{N}\p{P}\p{S}]$/u;
function isPrint(c: number): boolean {
  if (c < 0x80) {
    return c >= 0x20 && c < 0x7f;
  }
  return graphic.test(String.fromCodePoint(c));
}

// escapeRune is strconv's escape of a rune that is not printed as is.
function escapeRune(c: number): string {
  switch (c) {
    case 0x07:
      return "\\a";
    case 0x08:
      return "\\b";
    case 0x0c:
      return "\\f";
    case 0x0a:
      return "\\n";
    case 0x0d:
      return "\\r";
    case 0x09:
      return "\\t";
    case 0x0b:
      return "\\v";
  }
  if (c < 0x20 || c === 0x7f) {
    return `\\x${c.toString(16).padStart(2, "0")}`;
  }
  if (c >= 0xd800 && c < 0xe000) {
    c = 0xfffd; // not a valid rune
  }
  if (c < 0x10000) {
    return `\\u${c.toString(16).padStart(4, "0")}`;
  }
  return `\\U${c.toString(16).padStart(8, "0")}`;
}

/** quote is Go's strconv.Quote (the %q verb): s in double quotes with Go's escapes. */
export function quote(s: string): string {
  let out = '"';
  for (const ch of s) {
    const c = ch.codePointAt(0) ?? 0;
    if (c === 0x22 || c === 0x5c) {
      out += `\\${ch}`;
    } else if (isPrint(c) && !(c >= 0xd800 && c < 0xe000)) {
      out += ch;
    } else {
      out += escapeRune(c);
    }
  }
  return out + '"';
}

// ---- reading: the scanner ----

/** maxNestingDepth is Go's limit of nested arrays and objects. */
const maxNestingDepth = 10000;

// quoteChar is the scanner's quoting of the byte c in its messages.
function quoteChar(c: number): string {
  if (c === 0x27) {
    return "'\\''";
  }
  if (c === 0x22) {
    return "'\"'";
  }
  if (c === 0x5c) {
    return "'\\\\'";
  }
  return `'${isPrint(c) ? String.fromCodePoint(c) : escapeRune(c)}'`;
}

// The contexts where a space is wrong: at the end of the input, Go's scanner feeds a space and reports it
// rather than "unexpected end of JSON input".
const spaceRefused = new Set([
  "in numeric literal",
  "after decimal point in numeric literal",
  "in exponent of numeric literal",
  "in string escape code",
  "in \\u hexadecimal character escape",
  "in literal true (expecting 'r')",
  "in literal true (expecting 'u')",
  "in literal true (expecting 'e')",
  "in literal false (expecting 'a')",
  "in literal false (expecting 'l')",
  "in literal false (expecting 's')",
  "in literal false (expecting 'e')",
  "in literal null (expecting 'u')",
  "in literal null (expecting 'l')",
]);

const isSpace = (c: number | undefined): boolean =>
  c === 0x20 || c === 0x09 || c === 0x0a || c === 0x0d;
const isDigit = (c: number | undefined): boolean =>
  c !== undefined && c >= 0x30 && c <= 0x39;
const hexValue = (c: number | undefined): number => {
  if (c === undefined) {
    return -1;
  }
  if (c >= 0x30 && c <= 0x39) {
    return c - 0x30;
  }
  if (c >= 0x61 && c <= 0x66) {
    return c - 0x61 + 10;
  }
  if (c >= 0x41 && c <= 0x46) {
    return c - 0x41 + 10;
  }
  return -1;
};

type Frame =
  | { readonly kind: "array"; readonly items: JSONNode[] }
  | {
      readonly kind: "object";
      readonly entries: [string, JSONNode][];
      key: string;
    };

/**
 * parse reads data as Go's json.Unmarshal checks and scans it; it throws a JSONSyntaxError with Go's message
 * where data is no JSON.
 */
export function parse(data: Uint8Array): JSONNode {
  const b = data;
  const n = b.length;
  let i = 0;

  const fail = (context: string): never => {
    const c = b[i];
    if (c === undefined) {
      if (spaceRefused.has(context)) {
        throw new JSONSyntaxError(`invalid character ' ' ${context}`);
      }
      throw new JSONSyntaxError("unexpected end of JSON input");
    }
    throw new JSONSyntaxError(`invalid character ${quoteChar(c)} ${context}`);
  };
  const skipSpace = (): void => {
    while (isSpace(b[i])) {
      i++;
    }
  };
  // literal reads the rest of true, false or null after its first byte.
  const literal = (word: string): void => {
    i++;
    for (let k = 1; k < word.length; k++) {
      if (b[i] !== word.charCodeAt(k)) {
        fail(`in literal ${word} (expecting '${word.charAt(k)}')`);
      }
      i++;
    }
  };
  const number = (): JSONNode => {
    const start = i;
    if (b[i] === 0x2d) {
      i++;
      if (!isDigit(b[i])) {
        fail("in numeric literal");
      }
    }
    if (b[i] === 0x30) {
      i++;
    } else {
      while (isDigit(b[i])) {
        i++;
      }
    }
    if (b[i] === 0x2e) {
      i++;
      if (!isDigit(b[i])) {
        fail("after decimal point in numeric literal");
      }
      while (isDigit(b[i])) {
        i++;
      }
    }
    if (b[i] === 0x65 || b[i] === 0x45) {
      i++;
      if (b[i] === 0x2b || b[i] === 0x2d) {
        i++;
      }
      if (!isDigit(b[i])) {
        fail("in exponent of numeric literal");
      }
      while (isDigit(b[i])) {
        i++;
      }
    }
    let literal = "";
    for (let k = start; k < i; k++) {
      literal += String.fromCharCode(b[k] ?? 0);
    }
    return { kind: "number", literal };
  };
  // str scans a string from its opening quote and unquotes it.
  const str = (): string => {
    const start = ++i;
    for (;;) {
      const c = b[i];
      if (c === undefined) {
        fail("in string literal");
      } else if (c === 0x22) {
        break;
      } else if (c === 0x5c) {
        i++;
        const e = b[i];
        if (e === 0x75) {
          i++;
          for (let k = 0; k < 4; k++, i++) {
            if (hexValue(b[i]) < 0) {
              fail("in \\u hexadecimal character escape");
            }
          }
          continue;
        }
        if (e === undefined || !'"\\/bfnrt'.includes(String.fromCharCode(e))) {
          fail("in string escape code");
        }
        i++;
      } else if (c < 0x20) {
        fail("in string literal");
      } else {
        i++;
      }
    }
    const s = unquote(b.subarray(start, i));
    i++;
    return s;
  };
  const open = (): void => {
    if (stack.length + 1 > maxNestingDepth) {
      fail("exceeded max depth");
    }
    i++;
    skipSpace();
  };
  const key = (): string => {
    if (b[i] !== 0x22) {
      fail("looking for beginning of object key string");
    }
    const k = str();
    skipSpace();
    if (b[i] !== 0x3a) {
      fail("after object key");
    }
    i++;
    return k;
  };

  const stack: Frame[] = [];
  for (;;) {
    // A value starts at i.
    skipSpace();
    let value: JSONNode;
    const c = b[i];
    if (c === 0x7b) {
      open();
      if (b[i] === 0x7d) {
        i++;
        value = { kind: "object", entries: [] };
      } else {
        stack.push({ kind: "object", entries: [], key: key() });
        continue;
      }
    } else if (c === 0x5b) {
      open();
      if (b[i] === 0x5d) {
        i++;
        value = { kind: "array", items: [] };
      } else {
        stack.push({ kind: "array", items: [] });
        continue;
      }
    } else if (c === 0x22) {
      value = { kind: "string", value: str() };
    } else if (c === 0x2d || isDigit(c)) {
      value = number();
    } else if (c === 0x74) {
      literal("true");
      value = { kind: "bool", value: true };
    } else if (c === 0x66) {
      literal("false");
      value = { kind: "bool", value: false };
    } else if (c === 0x6e) {
      literal("null");
      value = { kind: "null" };
    } else {
      return fail("looking for beginning of value");
    }
    // The value is complete: hand it to the containers it closes.
    let next = false;
    while (!next) {
      const top = stack.at(-1);
      skipSpace();
      if (top === undefined) {
        if (i < n) {
          fail("after top-level value");
        }
        return value;
      }
      if (top.kind === "array") {
        top.items.push(value);
        if (b[i] === 0x2c) {
          i++;
          next = true;
        } else if (b[i] === 0x5d) {
          i++;
          stack.pop();
          value = top;
        } else {
          fail("after array element");
        }
      } else {
        top.entries.push([top.key, value]);
        if (b[i] === 0x2c) {
          i++;
          skipSpace();
          top.key = key();
          next = true;
        } else if (b[i] === 0x7d) {
          i++;
          stack.pop();
          value = { kind: "object", entries: top.entries };
        } else {
          fail("after object key:value pair");
        }
      }
    }
  }
}

// decodeRune is Go's utf8.DecodeRune: the rune at b[i] and its length; an invalid byte is U+FFFD of length 1.
function decodeRune(b: Uint8Array, i: number): [rune: number, size: number] {
  const c0 = b[i] ?? 0;
  const cont = (k: number, lo = 0x80, hi = 0xbf): number => {
    const c = b[i + k];
    return c !== undefined && c >= lo && c <= hi ? c & 0x3f : -1;
  };
  if (c0 < 0x80) {
    return [c0, 1];
  }
  if (c0 >= 0xc2 && c0 <= 0xdf) {
    const c1 = cont(1);
    if (c1 >= 0) {
      return [((c0 & 0x1f) << 6) | c1, 2];
    }
  } else if (c0 >= 0xe0 && c0 <= 0xef) {
    const c1 =
      c0 === 0xe0 ? cont(1, 0xa0) : c0 === 0xed ? cont(1, 0x80, 0x9f) : cont(1);
    const c2 = cont(2);
    if (c1 >= 0 && c2 >= 0) {
      return [((c0 & 0x0f) << 12) | (c1 << 6) | c2, 3];
    }
  } else if (c0 >= 0xf0 && c0 <= 0xf4) {
    const c1 =
      c0 === 0xf0 ? cont(1, 0x90) : c0 === 0xf4 ? cont(1, 0x80, 0x8f) : cont(1);
    const c2 = cont(2);
    const c3 = cont(3);
    if (c1 >= 0 && c2 >= 0 && c3 >= 0) {
      return [((c0 & 0x07) << 18) | (c1 << 12) | (c2 << 6) | c3, 4];
    }
  }
  return [0xfffd, 1];
}

// getu4 is the rune of the \uXXXX escape at b[i], -1 where there is none.
function getu4(b: Uint8Array, i: number): number {
  if (b[i] !== 0x5c || b[i + 1] !== 0x75) {
    return -1;
  }
  let r = 0;
  for (let k = 2; k < 6; k++) {
    const h = hexValue(b[i + k]);
    if (h < 0) {
      return -1;
    }
    r = r * 16 + h;
  }
  return r;
}

// unquote is Go's unquoteBytes on a scanned string's contents.
function unquote(b: Uint8Array): string {
  const runes: number[] = [];
  let i = 0;
  while (i < b.length) {
    const c = b[i] ?? 0;
    if (c === 0x5c) {
      const e = b[i + 1] ?? 0;
      if (e === 0x75) {
        let r = getu4(b, i);
        i += 6;
        if (r >= 0xd800 && r < 0xe000) {
          const r1 = getu4(b, i);
          if (r < 0xdc00 && r1 >= 0xdc00 && r1 < 0xe000) {
            runes.push(0x10000 + ((r - 0xd800) << 10) + (r1 - 0xdc00));
            i += 6;
            continue;
          }
          r = 0xfffd;
        }
        runes.push(r);
        continue;
      }
      const esc: Record<number, number> = {
        0x22: 0x22,
        0x5c: 0x5c,
        0x2f: 0x2f,
        0x62: 0x08,
        0x66: 0x0c,
        0x6e: 0x0a,
        0x72: 0x0d,
        0x74: 0x09,
      };
      runes.push(esc[e] ?? 0xfffd);
      i += 2;
    } else if (c < 0x80) {
      runes.push(c);
      i++;
    } else {
      const [r, size] = decodeRune(b, i);
      runes.push(r);
      i += size;
    }
  }
  let s = "";
  for (let k = 0; k < runes.length; k += 4096) {
    s += String.fromCodePoint(...runes.slice(k, k + 4096));
  }
  return s;
}

// ---- reading: decoding into typed records ----

/** DecodeContext is where a value sits, for Go's error messages: the innermost struct and the field path. */
export interface DecodeContext {
  readonly struct: string;
  readonly fields: readonly string[];
}

/** Codec reads one Go type: its name in Go's messages, its zero value and its decoding into a current value. */
export interface Codec<T> {
  readonly goType: string;
  zero(): T;
  decode(n: JSONNode, cur: T, ctx: DecodeContext): T;
}

const top: DecodeContext = { struct: "", fields: [] };

/** typeError is Go's UnmarshalTypeError for value (a JSON kind, or "number <literal>") into goType at ctx. */
function typeError(value: string, goType: string, ctx: DecodeContext): never {
  if (ctx.struct !== "" || ctx.fields.length > 0) {
    throw new UnmarshalTypeError(
      `json: cannot unmarshal ${value} into Go struct field ${ctx.struct}.${ctx.fields.join(".")} of type ${goType}`,
    );
  }
  throw new UnmarshalTypeError(
    `json: cannot unmarshal ${value} into Go value of type ${goType}`,
  );
}

const int64Min = -(2n ** 63n);
const int64Max = 2n ** 63n - 1n;

// parseInt is strconv.ParseInt(s, 10, 64) as the port holds it: the integer, or null where Go refuses s; it
// throws where Go reads a value past 2^53.
function parseInt64(s: string): number | null {
  if (!/^[+-]?[0-9]+$/.test(s)) {
    return null;
  }
  const v = BigInt(s);
  if (v < int64Min || v > int64Max) {
    return null;
  }
  const x = Number(v);
  if (!Number.isSafeInteger(x)) {
    throw new GoError(
      `json: number ${s} is outside the port's integer range (2^53)`,
    );
  }
  return x + 0; // no -0
}

/** goInt is Go's int. */
export const goInt: Codec<number> = {
  goType: "int",
  zero: () => 0,
  decode(n, cur, ctx) {
    switch (n.kind) {
      case "null":
        return cur;
      case "number": {
        const v = /^-?[0-9]+$/.test(n.literal) ? parseInt64(n.literal) : null;
        if (v === null) {
          return typeError(`number ${n.literal}`, "int", ctx);
        }
        return v;
      }
      default:
        return typeError(n.kind, "int", ctx);
    }
  },
};

/** goBool is Go's bool. */
export const goBool: Codec<boolean> = {
  goType: "bool",
  zero: () => false,
  decode(n, cur, ctx) {
    if (n.kind === "null") {
      return cur;
    }
    return n.kind === "bool" ? n.value : typeError(n.kind, "bool", ctx);
  },
};

/** goString is a Go string type by its name (string, or a named type such as dosrules.EndReason). */
export function goString(goType: string): Codec<string> {
  return {
    goType,
    zero: () => "",
    decode(n, cur, ctx) {
      if (n.kind === "null") {
        return cur;
      }
      return n.kind === "string" ? n.value : typeError(n.kind, goType, ctx);
    },
  };
}

/** goArray is Go's [length]T, held as an array of that length. */
export function goArray<T>(elem: Codec<T>, length: number): Codec<T[]> {
  const goType = `[${String(length)}]${elem.goType}`;
  const zero = (): T[] => Array.from({ length }, () => elem.zero());
  return {
    goType,
    zero,
    decode(n, cur, ctx) {
      if (n.kind === "null") {
        return cur;
      }
      if (n.kind !== "array") {
        return typeError(n.kind, goType, ctx);
      }
      const out = zero();
      for (let k = 0; k < length; k++) {
        const item = n.items[k];
        if (item !== undefined) {
          out[k] = elem.decode(item, cur[k] ?? elem.zero(), ctx);
        }
      }
      return out;
    },
  };
}

/**
 * goSlice is Go's []T, null for nil. Elements decode into the ones there (a null element keeps it); a new one
 * starts at zero (Go may show a stale element of the old backing array there, after a repeated key: not ported).
 */
export function goSlice<T>(elem: Codec<T>): Codec<T[] | null> {
  const goType = `[]${elem.goType}`;
  return {
    goType,
    zero: () => null,
    decode(n, cur, ctx) {
      if (n.kind === "null") {
        return null;
      }
      if (n.kind !== "array") {
        return typeError(n.kind, goType, ctx);
      }
      return n.items.map((item, k) =>
        elem.decode(item, cur?.[k] ?? elem.zero(), ctx),
      );
    },
  };
}

/** goIntMap is Go's map[int]T, null for nil; a key is read as strconv.ParseInt reads it, after its value. */
export function goIntMap<T>(elem: Codec<T>): Codec<Map<number, T> | null> {
  const goType = `map[int]${elem.goType}`;
  return {
    goType,
    zero: () => null,
    decode(n, cur, ctx) {
      if (n.kind === "null") {
        return null;
      }
      if (n.kind !== "object") {
        return typeError(n.kind, goType, ctx);
      }
      const m = new Map(cur);
      for (const [k, v] of n.entries) {
        const val = elem.decode(v, elem.zero(), ctx);
        const key = parseInt64(k);
        if (key === null) {
          return typeError(`number ${k}`, "int", ctx);
        }
        m.set(key, val);
      }
      return m;
    },
  };
}

/** goPointer is Go's *T, null for nil; its messages name T. */
export function goPointer<T>(elem: Codec<T>): Codec<T | null> {
  return {
    goType: elem.goType,
    zero: () => null,
    decode(n, cur, ctx) {
      if (n.kind === "null") {
        return null;
      }
      return elem.decode(n, cur ?? elem.zero(), ctx);
    },
  };
}

// foldName is Go's foldName for the json names here: ASCII upper case; U+017F (long s) and U+212A (Kelvin)
// fold to S and K, the only other runes whose fold matches an ASCII letter.
function foldName(s: string): string {
  return s.replace(/[a-z\u017f\u212a]/g, (c) =>
    c === "\u017f" ? "S" : c === "\u212a" ? "K" : c.toUpperCase(),
  );
}

/** StructFields are a struct's fields by record key: the json name and the field's codec, in Go's order. */
export type StructFields<T> = {
  readonly [K in keyof T]: readonly [json: string, codec: Codec<T[K]>];
};

/** goStruct is a Go struct named name (its type goType in messages) with the fields given. */
export function goStruct<T extends object>(
  name: string,
  goType: string,
  fields: StructFields<T>,
): Codec<T> {
  const keys = Object.keys(fields) as (keyof T & string)[];
  const byFold = new Map(keys.map((k) => [foldName(fields[k][0]), k] as const));
  const zero = (): T => {
    const o: Partial<T> = {};
    for (const k of keys) {
      o[k] = fields[k][1].zero();
    }
    return o as T;
  };
  return {
    goType,
    zero,
    decode(n, cur, ctx) {
      if (n.kind === "null") {
        return cur;
      }
      if (n.kind !== "object") {
        return typeError(n.kind, goType, ctx);
      }
      const out = { ...cur };
      for (const [k, v] of n.entries) {
        const f = byFold.get(foldName(k));
        if (f === undefined) {
          continue; // an unknown field, skipped unchecked
        }
        const [json, codec] = fields[f];
        out[f] = codec.decode(v, out[f], {
          struct: name,
          fields: [...ctx.fields, json],
        });
      }
      return out;
    },
  };
}

/** unmarshal is json.Unmarshal of data into a zero value of c; it throws a JSONSyntaxError or UnmarshalTypeError. */
export function unmarshal<T>(data: Uint8Array, c: Codec<T>): T {
  return c.decode(parse(data), c.zero(), top);
}
