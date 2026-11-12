import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import {
  BrowserStore,
  decodeEnvelope,
  defaultSettings,
  encodeEnvelope,
  parseSave,
  serializeSave,
  splitByDefault,
  validateSettings,
} from "./store.ts";

const utf8 = (text: string) => new TextEncoder().encode(text);

describe("native save envelopes", () => {
  it("exports exactly the native Go envelope fields and preserves bytes", () => {
    const bytes = utf8('{"schema":3,"name":"ä / < > &"}\n');
    const envelope = encodeEnvelope("1914-08:ec", 1, bytes);
    expect(JSON.parse(serializeSave(envelope))).toStrictEqual({
      map: "1914-08:ec",
      side: 1,
      data: Buffer.from(bytes).toString("base64"),
    });
    expect(decodeEnvelope(parseSave(serializeSave(envelope)))).toEqual(bytes);
  });

  it("reads the native Go JSON indentation and ignores extra envelope fields", () => {
    const text =
      '{\n  "map": "1914-08",\n  "side": 0,\n  "data": "e30=",\n  "future": 4\n}';
    expect(parseSave(text)).toStrictEqual({
      map: "1914-08",
      side: 0,
      data: "e30=",
    });
    expect(new TextDecoder().decode(decodeEnvelope(parseSave(text)))).toBe(
      "{}",
    );
  });

  it.each(["demo-ended", "full-midround", "full-winter-containers"])(
    "round-trips the real Go save %s without changing a byte",
    (name) => {
      const bytes = readFileSync(
        new URL(
          `../../../../internal/dosrules/testdata/saves/${name}.json`,
          import.meta.url,
        ),
      );
      const envelope = encodeEnvelope("1914-08", 0, bytes);
      expect(envelope.data).toBe(bytes.toString("base64"));
      expect(decodeEnvelope(parseSave(serializeSave(envelope)))).toEqual(
        new Uint8Array(bytes),
      );
    },
  );

  it.each([
    "",
    "{",
    "null",
    "[]",
    "{}",
    '{"map":"1914-08","side":0}',
    '{"map":"1914-08","side":"0","data":"e30="}',
    '{"map":"1914-08","side":2,"data":"e30="}',
    '{"map":" ","side":0,"data":"e30="}',
    '{"map":"1914-08","side":0,"data":null}',
    '{"map":"1914-08","side":0,"data":""}',
  ])("refuses malformed envelope %s", (text) => {
    expect(() => parseSave(text)).toThrow();
  });

  it.each(["e30", "e30===", "e30=\n", "e3$=", "e31=", "_30="])(
    "refuses malformed or noncanonical base64 %s",
    (data) => {
      expect(() => decodeEnvelope({ map: "1914-08", side: 0, data })).toThrow();
    },
  );

  it.each(["null", "[]", "0", '"save"', "{", "{} trailing"])(
    "refuses a decoded payload that is not a JSON object: %s",
    (payload) => {
      expect(() => encodeEnvelope("1914-08", 0, utf8(payload))).toThrow();
      expect(() =>
        parseSave(
          JSON.stringify({
            map: "1914-08",
            side: 0,
            data: Buffer.from(payload).toString("base64"),
          }),
        ),
      ).toThrow();
    },
  );

  it("refuses invalid UTF-8 instead of replacing damaged bytes", () => {
    const bytes = new Uint8Array([123, 34, 120, 34, 58, 34, 255, 34, 125]);
    expect(() => encodeEnvelope("1914-08", 0, bytes)).toThrow();
  });

  it("bounds imports before decoding and bounds names and raw saves", () => {
    expect(() => parseSave(" ".repeat(12 * 1024 * 1024))).toThrow();
    expect(() => encodeEnvelope("x".repeat(129), 0, utf8("{}"))).toThrow();
    expect(() =>
      encodeEnvelope("1914-08", 0, new Uint8Array(8 * 1024 * 1024 + 1)),
    ).toThrow();
    expect(() =>
      decodeEnvelope({
        map: "1914-08",
        side: 0,
        data: "A".repeat(12 * 1024 * 1024),
      }),
    ).toThrow();
  });
});

describe("persistent settings", () => {
  it("uses native and existing browser defaults, respecting requested system motion", () => {
    expect(defaultSettings()).toStrictEqual({
      uiScale: 1,
      reduced: false,
      scenes: true,
      painted: true,
      originalArt: false,
      scroll: "normal",
      opponent: "computer",
      layout: "auto",
      playMode: "single",
      limit: 0,
    });
    expect(defaultSettings(true).reduced).toBe(true);
  });

  it("fills missing fields and ignores future settings without coercion", () => {
    expect(validateSettings({ scroll: "fast", future: 9 })).toStrictEqual({
      uiScale: 1,
      reduced: false,
      scenes: true,
      painted: true,
      originalArt: false,
      scroll: "fast",
      opponent: "computer",
      layout: "auto",
      playMode: "single",
      limit: 0,
    });
    expect(
      validateSettings({
        uiScale: 0.75,
        reduced: true,
        scenes: false,
        painted: false,
        scroll: "slow",
        opponent: "passive",
        layout: "original",
        playMode: "hotseat",
        limit: 20,
      }),
    ).toStrictEqual({
      uiScale: 0.75,
      reduced: true,
      scenes: false,
      painted: false,
      originalArt: false,
      scroll: "slow",
      opponent: "passive",
      layout: "original",
      playMode: "hotseat",
      limit: 20,
    });
  });

  it("reads the former responsive default as auto and keeps an explicit choice", () => {
    expect(validateSettings({ layout: "responsive" }).layout).toBe("auto");
    expect(validateSettings({ layout: "single" }).layout).toBe("single");
    expect(validateSettings({ layout: "auto" }).layout).toBe("auto");
  });

  it("splits landscape windows at least 1000 CSS pixels wide by default", () => {
    expect(splitByDefault(1920, 1080)).toBe(true);
    expect(splitByDefault(1280, 720)).toBe(true);
    expect(splitByDefault(1000, 700)).toBe(true);
    expect(splitByDefault(999, 600)).toBe(false);
    // A phone held either way and a tall window keep the single map.
    expect(splitByDefault(412, 915)).toBe(false);
    expect(splitByDefault(915, 412)).toBe(false);
    expect(splitByDefault(1200, 1600)).toBe(false);
  });

  it.each([
    null,
    [],
    "{}",
    { reduced: "false" },
    { scenes: 0 },
    { painted: null },
    { scroll: "turbo" },
    { opponent: "human" },
    { layout: "split" },
    { layout: undefined },
    { layout: false },
    { playMode: "local" },
    { playMode: undefined },
    { playMode: 2 },
    { limit: 4 },
    { limit: -5 },
    { uiScale: 0 },
    { uiScale: Infinity },
    { uiScale: 1.1 },
    { uiScale: "1" },
    { reduced: undefined },
  ])("surfaces corrupt settings %j", (value) => {
    expect(() => validateSettings(value)).toThrow();
  });
});

it("round-trips a two-human native envelope without adding presentation settings", () => {
  const saved = encodeEnvelope("1914-09:2p", 1, utf8('{"schema":3}'));
  expect(parseSave(serializeSave(saved))).toStrictEqual({
    map: "1914-09:2p",
    side: 1,
    data: "eyJzY2hlbWEiOjN9",
  });
  expect(validateSettings({ playMode: "simultaneous" }).playMode).toBe(
    "simultaneous",
  );
});

it("reports unavailable browser storage instead of pretending saves succeeded", async () => {
  await expect(BrowserStore.open()).rejects.toThrow(/IndexedDB/);
});
