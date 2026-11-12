import { createHash } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { defineConfig, type Plugin } from "vite";
import { cacheDir } from "../cachedir.ts";

const outDir = "../../build/webapp/site";

/** Fills the service worker's file list and cache version from the finished bundle. */
function precacheManifest(): Plugin {
  return {
    name: "historyline-precache",
    apply: "build",
    async writeBundle(options, bundle) {
      const built = Object.keys(bundle).filter((file) => file !== "sw.js");
      const files = [
        "./",
        "index.json",
        "manifest.webmanifest",
        "icon-192.png",
        "icon-512.png",
        "icon-maskable-512.png",
        "apple-touch-icon.png",
        ...built.filter((file) => file !== "index.html"),
      ].sort();
      const version = createHash("sha256")
        .update(files.join("\n"))
        .digest("hex")
        .slice(0, 16);
      const path = join(options.dir ?? outDir, "sw.js");
      const source = await readFile(path, "utf8");
      if (
        !source.includes("__HISTORYLINE_PRECACHE__") ||
        !source.includes("__HISTORYLINE_VERSION__")
      )
        throw new Error("sw.js lost its precache placeholders");
      await writeFile(
        path,
        source
          .replace(
            // The minifier may change the literal's quotes.
            /["'`]__HISTORYLINE_PRECACHE__["'`]/,
            () => JSON.stringify(JSON.stringify(files)),
          )
          .replace("__HISTORYLINE_VERSION__", version),
      );
    },
  };
}

export default defineConfig({
  // Relative URLs: the site works from any path, such as the GitHub Pages project path.
  base: "./",
  cacheDir: cacheDir("app"),
  publicDir: "../../build/webapp/data",
  plugins: [precacheManifest()],
  // Hash-named outputs are immutable; watching 50,000 generated files exhausts
  // the dev server heap. Regenerate assets before starting the server.
  server: { watch: { ignored: ["**/build/webapp/data/blobs/**"] } },
  build: {
    outDir,
    emptyOutDir: true,
    target: "es2022",
    rolldownOptions: {
      input: { main: "index.html", sw: "src/sw.ts" },
      output: {
        entryFileNames: (chunk) =>
          chunk.name === "sw" ? "sw.js" : "assets/[name]-[hash].js",
      },
    },
  },
});
