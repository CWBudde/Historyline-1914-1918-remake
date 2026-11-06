import { defineConfig } from "vite";
import { cacheDir } from "../cachedir.ts";

export default defineConfig({
  // Relative URLs: the site works from any path, such as /next/ on GitHub Pages.
  base: "./",
  cacheDir: cacheDir("app"),
  publicDir: "../../build/webapp/data",
  // Hash-named outputs are immutable; watching 50,000 generated files exhausts
  // the dev server heap. Regenerate assets before starting the server.
  server: { watch: { ignored: ["**/build/webapp/data/blobs/**"] } },
  build: {
    outDir: "../../build/webapp/site",
    emptyOutDir: true,
    target: "es2022",
  },
});
