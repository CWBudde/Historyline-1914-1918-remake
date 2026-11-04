import { defineConfig } from "vitest/config";
import { cacheDir } from "./cachedir.ts";

export default defineConfig({
  cacheDir: cacheDir("vitest"),
  test: {
    include: ["engine/src/**/*.test.ts", "app/src/**/*.test.ts"],
    environment: "node",
  },
});
