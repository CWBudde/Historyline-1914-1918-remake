// Go's encoding/json as the saves need it (gojson.ts): the writer against what json.Marshal and strconv.Quote
// write (recorded from go1.26), and the reader's syntax errors and Go-semantics decoding. save.test.ts checks the
// whole of LoadGame against Go's results (testing/goLoadCases.ts).

import { describe, expect, it } from "vitest";
import {
  type JSONNode,
  JSONSyntaxError,
  UnmarshalTypeError,
  encodeInt,
  encodeIntMap,
  encodeObject,
  encodeString,
  goArray,
  goBool,
  goInt,
  goIntMap,
  goPointer,
  goSlice,
  goString,
  goStruct,
  parse,
  quote,
  unmarshal,
} from "./gojson.ts";

const utf8 = (s: string): Uint8Array => new TextEncoder().encode(s);

describe("encodeString", () => {
  it("escapes as json.Marshal does (HTML-safe, \\b and \\f, U+2028/9)", () => {
    expect(
      encodeString('a<b>&c\b\f\n\r\t\u0001\u007f"\\/\u2028\u2029\u00e9z\u001f'),
    ).toBe(
      '"a\\u003cb\\u003e\\u0026c\\b\\f\\n\\r\\t\\u0001\u007f\\"\\\\/\\u2028\\u2029\u00e9z\\u001f"',
    );
    expect(encodeString("\u{1F600}\u4e2d")).toBe('"\u{1F600}\u4e2d"');
    expect(encodeString("")).toBe('""');
  });
  it("writes a lone surrogate as Go writes invalid UTF-8", () => {
    expect(encodeString("x\ud800y\udc00")).toBe('"x\\ufffdy\\ufffd"');
  });
});

describe("encodeInt", () => {
  it("writes decimal integers", () => {
    expect([0, -0, 7, -12, 2 ** 53 - 1].map(encodeInt)).toStrictEqual([
      "0",
      "0",
      "7",
      "-12",
      "9007199254740991",
    ]);
  });
  it("refuses what is no Go int", () => {
    expect(() => encodeInt(1.5)).toThrow(TypeError);
    expect(() => encodeInt(2 ** 53)).toThrow(TypeError);
  });
});

describe("encodeIntMap", () => {
  it("sorts the keys by their decimal string, as Go sorts map keys", () => {
    const m = new Map([10, 2, 1, -1, 0, 240, 3].map((k) => [k, k] as const));
    expect(encodeIntMap(m, encodeInt)).toBe(
      '{"-1":-1,"0":0,"1":1,"10":10,"2":2,"240":240,"3":3}',
    );
    expect(encodeIntMap(new Map(), encodeInt)).toBe("{}");
  });
});

describe("encodeObject", () => {
  it("writes the fields in order, compact", () => {
    expect(
      encodeObject([
        ["b", "1"],
        ["a<", "[]"],
      ]),
    ).toBe('{"b":1,"a\\u003c":[]}');
  });
});

describe("quote", () => {
  it("is strconv.Quote (%q)", () => {
    expect(
      quote(
        "\u{E0001}\u00ad\u{1F600}\u0007\u000b\u0000\u0600\u200b\ufffd\u{10FFFF}'",
      ),
    ).toBe(
      '"\\U000e0001\\u00ad\u{1F600}\\a\\v\\x00\\u0600\\u200b\ufffd\\U0010ffff\'"',
    );
    expect(quote('x"y\n<\u0001\u00e9\u00a0\u007f')).toBe(
      '"x\\"y\\n<\\x01\u00e9\\u00a0\\x7f"',
    );
  });
});

describe("parse", () => {
  it("keeps number literals, key order and duplicate keys", () => {
    expect(
      parse(utf8(' {"b":-0.5e-3,"a":[true,false,null],"b":"x"} ')),
    ).toStrictEqual({
      kind: "object",
      entries: [
        ["b", { kind: "number", literal: "-0.5e-3" }],
        [
          "a",
          {
            kind: "array",
            items: [
              { kind: "bool", value: true },
              { kind: "bool", value: false },
              { kind: "null" },
            ],
          },
        ],
        ["b", { kind: "string", value: "x" }],
      ],
    } satisfies JSONNode);
  });
  it("unquotes as Go does: escapes, surrogate pairs, invalid UTF-8 byte by byte", () => {
    const s = (b: Uint8Array): string => {
      const n = parse(b);
      if (n.kind !== "string") {
        throw new Error("not a string");
      }
      return n.value;
    };
    expect(s(utf8('"\\ud83d\\ude00\\u00e9\\/\\b\\f\\n\\r\\t\\"\\\\"'))).toBe(
      '\u{1F600}\u00e9/\b\f\n\r\t"\\',
    );
    expect(s(utf8('"\\ud800\\u0041\\udc00\\ud800"'))).toBe(
      "\ufffdA\ufffd\ufffd",
    );
    // E2 82 (a truncated sequence), z, FF, ED A0 80 (an encoded surrogate): a U+FFFD per invalid byte.
    expect(
      s(new Uint8Array([0x22, 0xe2, 0x82, 0x7a, 0xff, 0xed, 0xa0, 0x80, 0x22])),
    ).toBe("\ufffd\ufffdz\ufffd\ufffd\ufffd\ufffd");
    expect(s(utf8('"\u4e2d\u{1F600}"'))).toBe("\u4e2d\u{1F600}");
  });
  it.each([
    ["[".repeat(10001), "invalid character '[' exceeded max depth"],
    ["[".repeat(10000) + "{", "invalid character '{' exceeded max depth"],
    ['{"a":"\\', "invalid character ' ' in string escape code"],
    [
      '{"a":"\\u12',
      "invalid character ' ' in \\u hexadecimal character escape",
    ],
    ['{"a":1.', "invalid character ' ' after decimal point in numeric literal"],
    ['{"a":t', "invalid character ' ' in literal true (expecting 'r')"],
  ])("refuses %#", (input, msg) => {
    expect(() => parse(utf8(input))).toThrow(new JSONSyntaxError(msg));
  });
  it("reads deep nesting up to Go's limit", () => {
    expect(parse(utf8("[".repeat(10000) + "]".repeat(10000))).kind).toBe(
      "array",
    );
  });
});

interface Inner {
  n: number;
  list: number[] | null;
}

interface Outer {
  name: string;
  on: boolean;
  pair: number[];
  inner: Inner;
  ptr: Inner | null;
  byKey: Map<number, Inner> | null;
}

const innerCodec = goStruct<Inner>("inner", "p.inner", {
  n: ["n", goInt],
  list: ["list", goSlice(goInt)],
});

const outerCodec = goStruct<Outer>("outer", "p.outer", {
  name: ["name", goString("string")],
  on: ["on", goBool],
  pair: ["pair", goArray(goInt, 2)],
  inner: ["inner", innerCodec],
  ptr: ["ptr", goPointer(innerCodec)],
  byKey: ["byKey", goIntMap(innerCodec)],
});

describe("unmarshal", () => {
  const read = (s: string): Outer => unmarshal(utf8(s), outerCodec);

  it("gives zero values for absent fields and ignores unknown ones", () => {
    expect(read('{"x":[1.5,"y"]}')).toStrictEqual({
      name: "",
      on: false,
      pair: [0, 0],
      inner: { n: 0, list: null },
      ptr: null,
      byKey: null,
    });
  });
  it("matches keys case-insensitively, the last one winning", () => {
    expect(read('{"NAME":"a","Name":"b","nAmE":"c","ON":true}')).toMatchObject({
      name: "c",
      on: true,
    });
  });
  it("decodes into what is there: null keeps a value, a map merges, a pointer's struct merges", () => {
    const o = read(
      '{"pair":[1,2],"pair":[null],"inner":{"n":3,"list":[1,2]},"inner":null,"inner":{"list":[null,null,4]},' +
        '"ptr":{"n":5},"ptr":{"list":[]},"byKey":{"1":{"n":1}},"byKey":{"+01":{"list":[7]},"2":null}}',
    );
    expect(o.pair).toStrictEqual([1, 0]);
    expect(o.inner).toStrictEqual({ n: 3, list: [1, 2, 4] });
    expect(o.ptr).toStrictEqual({ n: 5, list: [] });
    expect(o.byKey).toStrictEqual(
      new Map([
        [1, { n: 0, list: [7] }],
        [2, { n: 0, list: null }],
      ]),
    );
    expect(
      read('{"ptr":{"n":1},"ptr":null,"byKey":{},"byKey":null}'),
    ).toMatchObject({ ptr: null, byKey: null });
  });
  it("drops array elements past the length unchecked and zeroes the missing ones", () => {
    expect(read('{"pair":[1,2,"x",{}]}').pair).toStrictEqual([1, 2]);
    expect(read('{"pair":[1,2],"pair":[3]}').pair).toStrictEqual([3, 0]);
  });
  it.each([
    [
      '{"inner":{"n":1.0}}',
      "json: cannot unmarshal number 1.0 into Go struct field inner.inner.n of type int",
    ],
    [
      '{"byKey":{"1":{"list":["x"]}}}',
      "json: cannot unmarshal string into Go struct field inner.byKey.list of type int",
    ],
    [
      '{"byKey":{" 1":{}}}',
      "json: cannot unmarshal number  1 into Go struct field outer.byKey of type int",
    ],
    [
      '{"ptr":[]}',
      "json: cannot unmarshal array into Go struct field outer.ptr of type p.inner",
    ],
    [
      '{"pair":{}}',
      "json: cannot unmarshal object into Go struct field outer.pair of type [2]int",
    ],
    ["[]", "json: cannot unmarshal array into Go value of type p.outer"],
  ])("refuses %s", (input, msg) => {
    expect(() => read(input)).toThrow(new UnmarshalTypeError(msg));
  });
  it("refuses integers past what the port holds exactly (Go's int64 holds them)", () => {
    expect(() => read('{"inner":{"n":9007199254740992}}')).toThrow(
      /outside the port's integer range/,
    );
    expect(read('{"inner":{"n":-9007199254740991}}').inner.n).toBe(
      -9007199254740991,
    );
  });
});
