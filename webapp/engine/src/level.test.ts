import { describe, expect, it } from "vitest";
import { grid, marshal, parse } from "./level.ts";
import { readBytes } from "./testing/files.ts";

describe("legacy levels", () => {
  it.each(["1.0", "1e0", "-0.0"])(
    "rejects non-integer JSON notation %s like Go",
    (literal) => {
      expect(() =>
        parse(
          `{"mapName":"m","mapNo":${literal},"player":"p","tiles":{"line1":[["grass","",""]]}}`,
        ),
      ).toThrow();
      expect(() =>
        parse(
          `{"mapName":"m","mapNo":0,"player":"p","tiles":{"line1":[["hqN","","",[${literal}]]]}}`,
        ),
      ).toThrow();
    },
  );
  it("reads a null content owner as Go's zero int", () => {
    expect(
      parse(
        '{"mapName":"m","mapNo":0,"player":"p","tiles":{"line1":[["hqN","","",[null]]]}}',
      ).lines[0]?.[0]?.content?.owner,
    ).toBe(0);
  });
  it("writes a programmatically empty map with Go's whitespace", () => {
    expect(
      marshal({ mapName: "m", mapNo: 0, player: "p", lines: [], extra: null }),
    ).toBe(
      '{\n\t"mapName" :\t"m",\n\t"mapNo" :\t0,\n\t"player" :\t"p",\n\t"tiles" :\t{\n\t\t\t\t}\n}\n\n\n\n\n\n\n',
    );
  });
  it.each([
    "internal/level/testdata/pulse.json",
    "art/reference/pulse/level.json",
  ])("%s parses and writes the Go bytes", (path) => {
    const raw = new TextDecoder().decode(readBytes(path));
    const level = parse(raw);
    expect(marshal(level)).toBe(raw);
    expect(grid(level)).toEqual({ width: 8, lines: 47 });
  });

  const header = { mapName: "m", mapNo: 1, player: "p" };
  const hex = ["grass", "", ""];
  it("accepts string map numbers and preserves infrastructure and content forms", () => {
    const level = parse(
      JSON.stringify({
        ...header,
        mapNo: "7",
        tiles: {
          line1: [
            ["hqN", [], "", [0, []]],
            ["grass", "road236", "", [1]],
          ],
        },
      }),
    );
    expect(level.mapNo).toBe(7);
    expect(parse(marshal(level))).toEqual(level);
    expect(level.lines[0]?.[0]).toEqual({
      terrain: "hqN",
      infra: { ids: [], list: true },
      unit: "",
      content: { owner: 0, units: [] },
    });
    expect(level.lines[0]?.[1]?.content?.units).toBeNull();
  });

  it.each([
    { tiles: {} },
    { tiles: { line2: [hex] } },
    { tiles: { line1: [hex, hex], line2: [hex, hex] } },
    { tiles: { line1: [hex, hex], line2: [hex], line3: [hex] } },
    { tiles: { line1: [["grass", ""]] } },
    { tiles: { line1: [[1, "", ""]] } },
    { tiles: { line1: [["grass", 5, ""]] } },
    { tiles: { line1: [["grass", "", "", []]] } },
    { tiles: { line1: [["grass", "", "", [0, [], 1]]] } },
    { tiles: { line1: [hex] }, extra: 1 },
    { tiles: { line1: [hex] }, mapNo: "x" },
  ])("refuses malformed level %#", (bad) => {
    expect(() => parse(JSON.stringify({ ...header, ...bad }))).toThrow();
  });
  it("refuses missing headers and trailing objects", () => {
    expect(() => parse('{"tiles":{"line1":[["grass","",""]]}}')).toThrow();
    expect(() =>
      parse(JSON.stringify({ ...header, tiles: { line1: [hex] } }) + " {}"),
    ).toThrow();
  });
  it.each(['a"b', "a'b", "a\\b", "a\nb", "a\u007fb"])(
    "refuses unrepresentable strings %j",
    (unit) => {
      const level = parse(
        JSON.stringify({ ...header, tiles: { line1: [["grass", "", unit]] } }),
      );
      expect(() => marshal(level)).toThrow();
    },
  );
});
