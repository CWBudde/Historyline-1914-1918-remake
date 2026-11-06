import { defineConfig, devices } from "@playwright/test";

const port = process.env["E2E_PORT"] ?? "4173";

export default defineConfig({
  testDir: ".",
  outputDir: "../../../build/review/webapp-e2e/results",
  workers: 1,
  timeout: 5 * 60_000,
  reporter: [["list"]],
  use: {
    baseURL: `http://localhost:${port}/hl/`,
    reducedMotion: "reduce",
    trace: "retain-on-failure",
  },
  webServer: {
    command: "node e2e/server.ts",
    cwd: "..",
    url: `http://localhost:${port}/hl/`,
    reuseExistingServer: process.env["CI"] === undefined,
  },
  projects: [
    { name: "chromium", use: devices["Desktop Chrome"] },
    { name: "firefox", use: devices["Desktop Firefox"] },
    { name: "webkit", use: devices["Desktop Safari"] },
    { name: "android", use: devices["Pixel 7"] },
    { name: "ios", use: devices["iPhone 15"] },
  ],
});
