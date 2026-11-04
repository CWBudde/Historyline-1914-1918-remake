import { tmpdir } from "node:os";
import { join } from "node:path";

// Vite's and Vitest's cache stay off the project drive: $HISTORYLINE_VITE_CACHE or the system temp folder.
export function cacheDir(name: string): string {
  return join(
    process.env["HISTORYLINE_VITE_CACHE"] ?? join(tmpdir(), "historyline-vite"),
    name,
  );
}
