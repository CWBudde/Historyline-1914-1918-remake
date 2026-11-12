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
