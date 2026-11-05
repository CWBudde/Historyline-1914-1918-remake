// Test-only file access: the Go engine's fixtures and rule tables are read in place, so the TS port is
// checked against the same records.

import { readFileSync, readdirSync } from "node:fs";
import { fileURLToPath } from "node:url";

/** repo is the repository root, with a trailing slash. */
export const repo = fileURLToPath(new URL("../../../../", import.meta.url));

/** dosrules is internal/dosrules/, with a trailing slash. */
export const dosrules = `${repo}internal/dosrules/`;

/** readJSON parses the JSON file at path below the repository root. */
export function readJSON(path: string): unknown {
  return JSON.parse(readFileSync(repo + path, "utf8"));
}

/** readBytes reads the file at path below the repository root. */
export function readBytes(path: string): Uint8Array {
  return new Uint8Array(readFileSync(repo + path));
}

/** listDir lists the directory at path below the repository root. */
export function listDir(path: string): string[] {
  return readdirSync(repo + path);
}
