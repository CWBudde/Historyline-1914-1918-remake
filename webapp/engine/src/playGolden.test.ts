import { readFileSync } from "node:fs";
import { join } from "node:path";
import { gunzipSync } from "node:zlib";
import { describe, expect, it } from "vitest";
import * as play from "./play.ts";
import * as ai from "./ai.ts";
import { array, object, int, stringField } from "./decode.ts";
import { type Event as RuleEvent, newEvent } from "./cycle.ts";
import { Extended, Full } from "./testing/tables.ts";
import { repo, listDir } from "./testing/files.ts";
import { goValue, playData, shippedLevel } from "./testing/play.ts";

const dir =
  process.env["HL_WEBAPP_GOLDEN"] ?? join(repo, "build/webapp/golden");
const index = object(
  JSON.parse(readFileSync(join(dir, "index.json"), "utf8")),
  "index",
);
const names = array(index["traces"], "traces").map((s) => String(s));

describe.each(["replay", "plan"] as const)("Go %s goldens", (operation) => {
  it("cover every shipped setup at the Go AI's playout cap", () => {
    expect(index["schema"]).toBe(1);
    expect(index["resolutions"]).toBe(300);
    expect(names).toHaveLength(156);
    expect(new Set(names).size).toBe(156);
    expect(names).toEqual(
      listDir("levels")
        .filter((m) => /^19\d\d-\d\d$/.test(m))
        .sort()
        .flatMap((m) =>
          ["cp", "ec", "2p"].map((mode) => `${m}-${mode}.jsonl.gz`),
        ),
    );
  });
  it.each(names)(
    "%s: commands match session events and exact saved bytes",
    (name) => {
      const records = gunzipSync(readFileSync(join(dir, name)))
        .toString("utf8")
        .trimEnd()
        .split("\n");
      const header = object(JSON.parse(records.shift() ?? "null"), "header");
      const id = stringField(header, "name", "header");
      expect(name).toBe(id.replace(":", "-") + ".jsonl.gz");
      const [month, mode] = id.split(":");
      if (
        month === undefined ||
        (mode !== "cp" && mode !== "ec" && mode !== "2p")
      )
        throw new Error(`bad setup ${id}`);
      const s = play.newSession(shippedLevel(month, mode), playData, {
        profile: Full,
        extended: Extended,
        computer: [true, true],
      });
      expect(new TextDecoder().decode(play.save(s))).toBe(header["save"]);
      expect(records.length).toBeGreaterThan(0);
      let n = 0;
      for (const record of records) {
        const frame = object(JSON.parse(record), "frame");
        const cmds = array(frame["commands"], "commands").map(
          (v): play.Command => {
            const c = object(v, "command");
            const kind = stringField(
              c,
              "kind",
              "command",
            ) as play.Command["kind"];
            return {
              kind,
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
            : [...ai.plan({}, s, 0), ...ai.plan({}, s, 1), ...play.resolve(s)];
        const normalize = (c: play.Command) => ({
          kind: c.kind,
          side: c.side,
          cell: c.cell,
          unit: c.unit ?? 0,
          hold: c.hold ?? 0,
          slot: c.slot ?? 0,
          type: c.type ?? 0,
        });
        expect(
          play.commandsSince(s, mark).map(normalize),
          `${id}, resolution ${n}: commands`,
        ).toEqual(cmds.map(normalize));
        const expected = array(goValue(frame["events"]), "events").map((v) => {
          const ev = object(v, "event");
          return {
            ...ev,
            raw: newEvent(object(ev["raw"], "raw") as unknown as RuleEvent),
          };
        });
        expect(goValue(events), `${id}, resolution ${n}: events`).toEqual(
          expected,
        );
        expect(
          new TextDecoder().decode(play.save(s)),
          `${id}, resolution ${n}: save`,
        ).toBe(frame["save"]);
        n++;
      }
      expect(n).toBeLessThanOrEqual(300);
      if (n < 300)
        expect(
          play.ended(s),
          `${id}: trace stopped before the cap without an outcome`,
        ).not.toBeNull();
    },
    600_000,
  );
});
