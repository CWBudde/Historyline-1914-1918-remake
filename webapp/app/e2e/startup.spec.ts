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
  // Transfers by their end in wall-clock milliseconds: CDP times are monotonic seconds, mapped by the wall time
  // each request reports with its start. Full art follows the first playable frame at once, so only what ended
  // before its mark counts, not what arrives while the test notices it.
  const clock = { offset: Number.NaN };
  const finished: [number, number][] = [];
  cdp.on("Network.requestWillBeSent", (event) => {
    if (Number.isNaN(clock.offset))
      clock.offset = event.wallTime - event.timestamp;
  });
  cdp.on("Network.loadingFinished", (event) => {
    finished.push([event.timestamp, event.encodedDataLength]);
  });
  await openMenu(page);
  await startFirstBattle(page);
  await page.waitForFunction(
    () =>
      performance.getEntriesByName("historyline.first-playable").length === 1,
    undefined,
    { timeout: 120_000 },
  );
  const mark = await page.evaluate(
    () =>
      performance.timeOrigin +
      (performance.getEntriesByName("historyline.first-playable")[0]
        ?.startTime ?? 0),
  );
  expect(clock.offset).not.toBeNaN();
  const before = finished.filter(([at]) => (at + clock.offset) * 1000 <= mark),
    bytes = before.reduce((sum, [, size]) => sum + size, 0),
    requests = before.length;
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
