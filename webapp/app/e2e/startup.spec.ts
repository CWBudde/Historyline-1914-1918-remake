import { expect, test } from "@playwright/test";
import { openMenu, startFirstBattle } from "./battle.ts";

// Lite art at the lowest density reaches the first playable frame; full art follows it.
const firstPlayableBytes = 5_000_000;

test("the first playable frame needs few megabytes, also at a phone's density", async ({
  page,
  browserName,
}, info) => {
  test.skip(
    browserName !== "chromium",
    "Transfer sizes come from Chromium's DevTools protocol.",
  );
  const cdp = await page.context().newCDPSession(page);
  await cdp.send("Network.enable");
  let bytes = 0,
    requests = 0,
    counting = true;
  cdp.on("Network.loadingFinished", (event) => {
    if (!counting) return;
    bytes += event.encodedDataLength;
    requests++;
  });
  await openMenu(page);
  await startFirstBattle(page);
  await page.waitForFunction(
    () =>
      performance.getEntriesByName("historyline.first-playable").length === 1,
    undefined,
    { timeout: 120_000 },
  );
  counting = false;
  console.log(
    `${info.project.name}: first playable ${(bytes / 1e6).toFixed(2)} MB in ${requests} requests`,
  );
  expect(bytes).toBeLessThanOrEqual(firstPlayableBytes);
});

test("the main loop and render stay within the frame budget", async ({
  page,
  browserName,
}, info) => {
  await page.addInitScript(() => {
    const probe = { collecting: false, frames: [] as number[] };
    (
      globalThis as typeof globalThis & { __frameBudget: typeof probe }
    ).__frameBudget = probe;
    new PerformanceObserver((list) => {
      if (probe.collecting)
        for (const entry of list.getEntries())
          if (entry.name === "historyline.main-frame")
            probe.frames.push(entry.duration);
    }).observe({ type: "measure" });
  });
  await openMenu(page);
  await startFirstBattle(page);
  const pan = async () => {
    for (const key of ["d", "s", "a", "w"]) {
      await page.keyboard.down(key);
      await page.waitForTimeout(500);
      await page.keyboard.up(key);
    }
  };
  await pan();
  await page.evaluate(() => {
    (
      globalThis as typeof globalThis & {
        __frameBudget: { collecting: boolean };
      }
    ).__frameBudget.collecting = true;
  });
  await pan();
  await pan();
  const frames = await page.evaluate(() => {
    const probe = (
      globalThis as typeof globalThis & {
        __frameBudget: { collecting: boolean; frames: number[] };
      }
    ).__frameBudget;
    probe.collecting = false;
    return probe.frames;
  });
  frames.sort((a, b) => a - b);
  const p95 = frames[Math.ceil(frames.length * 0.95) - 1] ?? Infinity;
  const touch = info.project.use.hasTouch === true;
  const limit = touch ? 1000 / 30 : 1000 / 60;
  console.log(
    `${info.project.name}: ${frames.length} frames, main loop and render p95 ${p95.toFixed(1)} ms (limit ${limit.toFixed(1)} ms)`,
  );
  expect(frames.length).toBeGreaterThan(30);
  // Linux WebKit renders in software: record its timings, enforce them on a GPU runner.
  if (
    browserName === "webkit" &&
    process.platform === "linux" &&
    process.env["HISTORYLINE_WEBKIT_GPU"] !== "1"
  )
    return;
  expect(p95).toBeLessThanOrEqual(limit);
});
