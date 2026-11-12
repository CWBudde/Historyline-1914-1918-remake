// Test-only: the local original game files (original/extracted/dos-full, git-ignored), keyed as the import keys them
// ("MAP/05.FIN", "CODES.DAT"); null without them. Only numbers derived from them enter the assertions.
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { repo } from "./files.ts";

export const originalDir = `${repo}original/extracted/dos-full/`;

export function originalFiles(): Map<string, Uint8Array> | null {
  if (!existsSync(`${originalDir}CODES.DAT`)) return null;
  const files = new Map<string, Uint8Array>();
  const add = (path: string) => {
    files.set(
      path.toUpperCase(),
      new Uint8Array(readFileSync(originalDir + path)),
    );
  };
  add("CODES.DAT");
  add("00.PAL");
  for (const name of readdirSync(`${originalDir}MAP`)) add(`MAP/${name}`);
  for (const name of readdirSync(`${originalDir}LIB`))
    if (/^(PARTS|PARTW|UNIT)\.(DAT|LIB)$/i.test(name)) add(`LIB/${name}`);
  return files;
}
