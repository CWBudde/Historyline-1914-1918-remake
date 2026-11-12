import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test, type Page } from "@playwright/test";
import {
  battleReady,
  expectMapDrawn,
  openMenu,
  startFirstBattle,
} from "./battle.ts";

function code(month: string): string {
  const path = resolve(import.meta.dirname, `../../../levels/${month}/cp.json`);
  return (JSON.parse(readFileSync(path, "utf8")) as { code: string }).code;
}
async function battleFromCode(page: Page, month: string): Promise<void> {
  await page.getByRole("button", { name: /^Code/ }).click();
  await page.getByRole("textbox").fill(code(month));
  await page.keyboard.press("Enter");
  await page.getByRole("button", { name: "Start" }).click();
  await page.getByRole("button", { name: /^Germany/ }).click();
  // A month shows its news and new units only when it has any.
  const unavailable = page.getByText(/not available offline yet/);
  for (;;) {
    if (
      (await battleReady(page).isVisible()) ||
      (await unavailable.isVisible())
    )
      return;
    for (const name of ["Continue", "To battle"]) {
      const button = page.getByRole("button", { name });
      if (await button.isVisible()) await button.click({ timeout: 2_000 });
    }
    await page.waitForTimeout(200);
  }
}
/**
 * The keep starts 15 s after a battle starts and reads one file at a time.
 * Without a GPU the map's frames starve it for over 20 minutes, so no more are
 * drawn until the caller's reload.
 */
async function waitForKeep(page: Page): Promise<void> {
  await page.evaluate(() => {
    window.requestAnimationFrame = () => 0;
  });
  await page.waitForTimeout(16_000);
  let last = -1;
  for (;;) {
    const count = await page.evaluate(async () => {
      const cache = await caches.open("historyline-content-v1");
      return (await cache.keys()).length;
    });
    if (count === last) return;
    last = count;
    await page.waitForTimeout(5_000);
  }
}

/** A month's backdrop tile hashes, read online through the app's index and the month's pyramid file. */
async function tileHashes(page: Page, month: string): Promise<string[]> {
  return page.evaluate(async (wanted) => {
    interface Pyramid {
      mips: { tiles: { file: { sha256: string } }[] }[];
    }
    const index = (await (await fetch("index.json")).json()) as {
      levels: Record<
        string,
        { backdrop?: Pyramid; backdropFile?: { path: string } }
      >;
    };
    const level = index.levels[wanted];
    const pyramid =
      level?.backdropFile === undefined
        ? level?.backdrop
        : ((await (await fetch(level.backdropFile.path)).json()) as Pyramid);
    return (pyramid?.mips ?? []).flatMap((mip) =>
      mip.tiles.map((tile) => tile.file.sha256),
    );
  }, month);
}

/** How many of the given backdrop tiles are missing from the content cache. */
async function missingTiles(page: Page, hashes: string[]): Promise<number> {
  return page.evaluate(async (wanted) => {
    const cache = await caches.open("historyline-content-v1");
    let missing = 0;
    for (const hash of wanted) {
      const key = new URL(`/__historyline_content__/${hash}`, location.origin)
        .href;
      if ((await cache.match(key)) === undefined) missing++;
    }
    return missing;
  }, hashes);
}

test("played and upcoming months work offline, others say why not", async ({
  page,
  context,
}) => {
  test.setTimeout(20 * 60_000);
  await openMenu(page);
  await page.evaluate(async () => {
    await navigator.serviceWorker.ready;
  });
  const tiles = new Map<string, string[]>();
  for (const month of ["1914-08", "1914-09", "1914-10", "1918-11"])
    tiles.set(month, await tileHashes(page, month));
  await startFirstBattle(page);
  await waitForKeep(page);

  await context.setOffline(true);
  await page.reload();
  await expect(page.getByRole("button", { name: "Start" })).toBeVisible({
    timeout: 60_000,
  });

  for (const month of ["1914-08", "1914-09", "1914-10"]) {
    const hashes = tiles.get(month) ?? [];
    expect(hashes.length, month).toBeGreaterThan(0);
    expect(await missingTiles(page, hashes), month).toBe(0);
  }
  const lastMonth = tiles.get("1918-11") ?? [];
  expect(lastMonth.length).toBeGreaterThan(0);
  expect(await missingTiles(page, lastMonth)).toBe(lastMonth.length);

  // The played month, without any report of missing full-quality art.
  await startFirstBattle(page);
  await expectMapDrawn(page);
  await expect(page.locator("body")).not.toContainText(
    /asset fetch|Failed to fetch/,
  );
  await battleReady(page).click();
  await expect(page.locator(".battle-reason")).toHaveCount(0);

  // The next month of the chain was kept in the background.
  await page.reload();
  await expect(page.getByRole("button", { name: "Start" })).toBeVisible({
    timeout: 60_000,
  });
  await battleFromCode(page, "1914-09");
  await expect(battleReady(page)).toBeVisible({ timeout: 60_000 });
  await expectMapDrawn(page);

  // The war's last month was not.
  await page.reload();
  await expect(page.getByRole("button", { name: "Start" })).toBeVisible({
    timeout: 60_000,
  });
  await battleFromCode(page, "1918-11");
  await expect(page.getByText(/not available offline yet/)).toBeVisible({
    timeout: 60_000,
  });
});

/** startFirstBattle with a longer wait: without a GPU a cold start from the cache is slow. */
async function startSlowBattle(page: Page): Promise<void> {
  await page.getByRole("button", { name: "Start" }).click();
  await page.getByRole("button", { name: /Germany/ }).click();
  await page.getByRole("button", { name: "Continue" }).click();
  await page.getByRole("button", { name: "To battle" }).click();
  await expect(battleReady(page)).toBeVisible({ timeout: 480_000 });
}

async function openOfflineMaps(page: Page): Promise<void> {
  await page.getByRole("button", { name: "Extras" }).click();
  await page.getByRole("button", { name: /^Offline maps/ }).click();
  await expect(
    page.getByRole("heading", { name: "Selected map" }),
  ).toBeVisible();
}
const downloadButton = (page: Page) =>
  page.getByRole("button", { name: /^Download map · / });
async function offlineRecords(page: Page): Promise<number> {
  return page.evaluate(async () => {
    const cache = await caches.open("historyline-offline-v1");
    return (await cache.keys()).length;
  });
}
async function contentFiles(page: Page): Promise<number> {
  return page.evaluate(async () => {
    const cache = await caches.open("historyline-content-v1");
    return (await cache.keys()).length;
  });
}

test("a map is downloaded from Extras, played offline and removed", async ({
  page,
  context,
}) => {
  test.setTimeout(20 * 60_000);
  await openMenu(page);
  await page.evaluate(async () => {
    await navigator.serviceWorker.ready;
  });
  const month = "1914-08";
  const tiles = await tileHashes(page, month);
  expect(tiles.length).toBeGreaterThan(0);
  expect(await missingTiles(page, tiles)).toBe(tiles.length);

  await openOfflineMaps(page);
  const offer = await downloadButton(page).innerText({ timeout: 120_000 });
  const megabytes = Number(/([\d.]+) MB/.exec(offer)?.[1]);
  expect(megabytes).toBeGreaterThan(0);
  console.log(`offline month ${month}: ${megabytes} MB to download`);
  await downloadButton(page).click();
  await expect(page.locator("progress")).toBeVisible();
  await expect(page.getByRole("button", { name: "Cancel" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Downloaded" })).toBeVisible({
    timeout: 15 * 60_000,
  });
  expect(await missingTiles(page, tiles)).toBe(0);
  await expect(page.getByText(/Remove all/)).toBeVisible();
  expect(await offlineRecords(page)).toBe(1);
  console.log(
    `offline storage: ${await page.getByText(/Browser storage|Storage use/).innerText()}`,
  );

  await context.setOffline(true);
  await page.reload();
  await expect(page.getByRole("button", { name: "Start" })).toBeVisible({
    timeout: 60_000,
  });
  await startSlowBattle(page);
  await expectMapDrawn(page);
  await expect(page.locator("body")).not.toContainText(
    /asset fetch|Failed to fetch/,
  );

  // Back in the menu, the month is still listed, and removing it deletes its files.
  await page.reload();
  await expect(page.getByRole("button", { name: "Start" })).toBeVisible({
    timeout: 60_000,
  });
  await openOfflineMaps(page);
  await expect(page.getByRole("button", { name: "Downloaded" })).toBeVisible();
  const before = await contentFiles(page);
  await page.getByRole("button", { name: "Remove", exact: true }).click();
  await expect(downloadButton(page)).toBeVisible();
  expect(await missingTiles(page, tiles)).toBe(tiles.length);
  expect(await offlineRecords(page)).toBe(0);
  expect(await contentFiles(page)).toBeLessThan(before);
  // The shared art went with the last kept month; what remains is art the pages loaded on their own.
  console.log(
    `content cache: ${before} files before the removal, ${await contentFiles(page)} after`,
  );
});

test("cancelling a download keeps nothing", async ({ page }) => {
  test.setTimeout(10 * 60_000);
  await openMenu(page);
  const tiles = await tileHashes(page, "1914-08");
  await openOfflineMaps(page);
  await downloadButton(page).click({ timeout: 120_000 });
  // Cancel with files already stored.
  await expect
    .poll(async () => (await contentFiles(page)) > 3, { timeout: 60_000 })
    .toBe(true);
  await page.getByRole("button", { name: "Cancel" }).click();
  await expect(downloadButton(page)).toBeVisible();
  await expect(page.getByRole("button", { name: "Downloaded" })).toHaveCount(0);
  // The cleanup of the arrived files follows the cancel.
  await expect
    .poll(async () => await missingTiles(page, tiles), { timeout: 30_000 })
    .toBe(tiles.length);
  expect(await offlineRecords(page)).toBe(0);
  await expect(page.getByText("No maps are kept")).toBeVisible();
});
