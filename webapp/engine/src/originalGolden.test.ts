// Twin of the Go bindings of the original maps: `hl webapp original` writes, from the local original files,
// each map's finmap level, its exact board (dosbind.FromOriginal), its view (origview.Campaign) and an AI-vs-AI
// trace to build/webapp/original. The TS port reads the same local files and must match field for field, and
// replay and plan the traces to the same events and saved bytes. Skips without the files or the export.
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { gunzipSync } from "node:zlib";
import { describe, expect, it } from "vitest";
import * as play from "./play.ts";
import * as ai from "./ai.ts";
import * as origmaps from "./origmaps.ts";
import * as origview from "./origview.ts";
import { fromOriginal } from "./original.ts";
import { marshal } from "./level.ts";
import { moveTablesOf } from "./board.ts";
import { array, int, object, stringField } from "./decode.ts";
import { type Event as RuleEvent, newEvent } from "./cycle.ts";
import { Full } from "./testing/tables.ts";
import { readJSON, repo } from "./testing/files.ts";
import { goValue, playData } from "./testing/play.ts";
import { originalFiles } from "./testing/original.ts";

const dir =
  process.env["HL_WEBAPP_ORIGINAL"] ?? join(repo, "build/webapp/original");
const files = originalFiles();
const exported = existsSync(join(dir, "index.json"));
const index = exported
  ? object(JSON.parse(readFileSync(join(dir, "index.json"), "utf8")), "index")
  : {};
const stems = exported
  ? array(index["maps"], "maps").map((s) => String(s))
  : [];
const levels = files === null ? [] : origmaps.read(files, moveTablesOf(Full));
const looks = origview.parseLooks(readJSON("internal/origview/looks.json"));
const terrains = new Set(
  Object.keys(object(readJSON("internal/data/terrain.json"), "terrain")),
);
const ready = files !== null && exported;

function golden(stem: string): Record<string, unknown> {
  return object(
    JSON.parse(readFileSync(join(dir, `${stem}.json`), "utf8")),
    stem,
  );
}
function level(stem: string) {
  const l = levels.find((l) => l.source.map === stem);
  if (l === undefined) throw new Error(`no map ${stem}`);
  return l;
}

describe.skipIf(!ready)("Go original bindings", () => {
  it("cover every map of the local original files", () => {
    expect(index["schema"]).toBe(1);
    expect(stems).toEqual(levels.map((l) => l.source.map));
  });
  it.each(stems)("%s: level, board and view match Go", (stem) => {
    const g = golden(stem),
      l = level(stem);
    expect(JSON.parse(JSON.stringify(l)), "finmap level").toStrictEqual(
      g["level"],
    );
    const b = fromOriginal(l, playData, Full);
    const want = object(g["board"], "board");
    expect({
      width: b.width,
      height: b.height,
      winter: b.winter,
      terrain: b.terrain,
      units: [...b.units].map(([i, u]) => [i, u.type, u.owner, u.cell]),
      ids: Object.fromEntries([...b.ids].map(([i, id]) => [String(i), id])),
      holds: [...b.holds]
        .sort((x, y) => x[0] - y[0])
        .map(([i, h]) => ({ index: i, ...h })),
      buildings: [...b.buildings]
        .sort((x, y) => x[0] - y[0])
        .map(([i, bd]) => ({
          index: i,
          credits: bd.credits,
          builds: bd.builds,
          computerBuilds: bd.computerBuilds ?? [],
        })),
      parts: Object.fromEntries([...b.parts].map(([c, e]) => [String(c), e])),
    }).toStrictEqual(want);
    expect(
      marshal(origview.campaign(l, playData, looks, terrains)),
      "view",
    ).toBe(g["view"]);
  });
});

describe.skipIf(!ready).each(["replay", "plan"] as const)(
  "Go original %s traces",
  (operation) => {
    it.each(stems)(
      "%s: commands match session events and exact saved bytes",
      (stem) => {
        const records = gunzipSync(readFileSync(join(dir, `${stem}.jsonl.gz`)))
          .toString("utf8")
          .trimEnd()
          .split("\n");
        const header = object(JSON.parse(records.shift() ?? "null"), "header");
        expect(stringField(header, "name", "header")).toBe(`original/${stem}`);
        const l = level(stem);
        const s = play.newBoard(
          fromOriginal(l, playData, Full),
          origview.campaign(l, playData, looks, terrains),
          playData,
          { profile: Full, computer: [true, true] },
        );
        expect(new TextDecoder().decode(play.save(s))).toBe(header["save"]);
        let n = 0;
        const normalize = (c: play.Command) => ({
          kind: c.kind,
          side: c.side,
          cell: c.cell,
          unit: c.unit ?? 0,
          hold: c.hold ?? 0,
          slot: c.slot ?? 0,
          type: c.type ?? 0,
        });
        for (const record of records) {
          const frame = object(JSON.parse(record), "frame");
          const cmds = array(frame["commands"], "commands").map(
            (v): play.Command => {
              const c = object(v, "command");
              return {
                kind: stringField(c, "kind", "command") as play.Command["kind"],
                side: int(c["side"], "side"),
                unit: Number(c["unit"] ?? 0),
                hold: Number(c["hold"] ?? 0),
                slot: Number(c["slot"] ?? 0),
                cell: Number(c["cell"] ?? 0),
                type: Number(c["type"] ?? 0),
              };
            },
          );
          const mark = play.mark(s);
          const events =
            operation === "replay"
              ? play.replay(s, cmds)
              : [
                  ...ai.plan({}, s, 0),
                  ...ai.plan({}, s, 1),
                  ...play.resolve(s),
                ];
          expect(
            play.commandsSince(s, mark).map(normalize),
            `${stem}, resolution ${n}: commands`,
          ).toEqual(cmds.map(normalize));
          const expected = array(goValue(frame["events"]), "events").map(
            (v) => {
              const ev = object(v, "event");
              return {
                ...ev,
                raw: newEvent(object(ev["raw"], "raw") as unknown as RuleEvent),
              };
            },
          );
          expect(goValue(events), `${stem}, resolution ${n}: events`).toEqual(
            expected,
          );
          expect(
            new TextDecoder().decode(play.save(s)),
            `${stem}, resolution ${n}: save`,
          ).toBe(frame["save"]);
          n++;
        }
        const cap = Number(index["resolutions"]);
        if (n < cap)
          expect(
            play.ended(s),
            `${stem}: stopped without an outcome`,
          ).not.toBeNull();
      },
      600_000,
    );
  },
);
