import { readdirSync, readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { isFixtureRecord } from "./index.ts";

// The Go engine's fixtures and rule tables are read in place; the TS port is checked against the same records.
const dosrules = fileURLToPath(
  new URL("../../../internal/dosrules/", import.meta.url),
);
const kinds = [
  "casualties",
  "containers",
  "engagements",
  "movements",
  "orders",
  "steps",
  "visibility",
];

function readJSON(path: string): unknown {
  return JSON.parse(readFileSync(dosrules + path, "utf8"));
}

// The header checks of dosrules_test.go checkHeader: schema 1, the profile name, an emulated routine of the
// tables' build, ids unique per file.
describe.each(["demo", "full"])("dosrules fixtures, %s build", (build) => {
  const tables = readJSON(`${build}.json`) as { build: string };

  it("holds exactly the known kinds", () => {
    const files = readdirSync(`${dosrules}testdata/${build}`).filter((f) =>
      f.endsWith(".json"),
    );
    expect(files.map((f) => f.slice(0, -".json".length)).sort()).toEqual(kinds);
  });

  it.each(kinds)("%s records carry the fixture header", (kind) => {
    const records = readJSON(`testdata/${build}/${kind}.json`);
    expect(Array.isArray(records)).toBe(true);
    const list = records as unknown[];
    expect(list.length).toBeGreaterThan(0);
    const ids = new Set<string>();
    for (const r of list) {
      if (!isFixtureRecord(r)) {
        expect.fail(`not a fixture record: ${JSON.stringify(r).slice(0, 120)}`);
      }
      expect(r.schema).toBe(1);
      expect(r.profile).toBe("dos-0.1");
      expect(r.source.kind).toBe("emulated-routine");
      expect(r.source.build).toBe(tables.build);
      expect(ids.has(r.id), r.id).toBe(false);
      ids.add(r.id);
    }
  });
});
