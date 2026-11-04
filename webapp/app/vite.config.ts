import { defineConfig } from "vite";
import { cacheDir } from "../cachedir.ts";

export default defineConfig({
  // Relative URLs: the site works from any path, such as /next/ on GitHub Pages.
  base: "./",
  cacheDir: cacheDir("app"),
  build: {
    outDir: "../../build/webapp/site",
    emptyOutDir: true,
    target: "es2022",
  },
});
