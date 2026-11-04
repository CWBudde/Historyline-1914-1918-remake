// The envelope every record of internal/dosrules/testdata/{demo,full}/*.json shares (dosrules/fixtures.go).

export interface FixtureSource {
  kind: string;
  build: string;
  note?: string;
}

export interface FixtureRecord {
  schema: number;
  profile: string;
  id: string;
  source: FixtureSource;
}

function isObject(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

export function isFixtureRecord(v: unknown): v is FixtureRecord {
  if (!isObject(v) || !isObject(v["source"])) {
    return false;
  }
  const src = v["source"];
  return (
    typeof v["schema"] === "number" &&
    typeof v["profile"] === "string" &&
    typeof v["id"] === "string" &&
    typeof src["kind"] === "string" &&
    typeof src["build"] === "string" &&
    (src["note"] === undefined || typeof src["note"] === "string")
  );
}
