import { expect, type Page } from "@playwright/test";

/** Opens the menu and walks through the first campaign month's pages into its battle. */
export async function openMenu(page: Page): Promise<void> {
  await page.goto("./");
  await expect(page.getByRole("button", { name: "Start" })).toBeVisible({
    timeout: 120_000,
  });
}
export async function startFirstBattle(page: Page): Promise<void> {
  await page.getByRole("button", { name: "Start" }).click();
  await page.getByRole("button", { name: /^Germany/ }).click();
  await page.getByRole("button", { name: "Continue" }).click();
  await page.getByRole("button", { name: "To battle" }).click();
  await expect(battleReady(page)).toBeVisible({ timeout: 120_000 });
}
export function battleReady(page: Page) {
  return page.locator(".battle-ready").first();
}
/** A blank or single-colour canvas compresses to far less than a drawn map. */
export async function expectMapDrawn(page: Page): Promise<void> {
  await expect
    .poll(
      async () =>
        (await page.locator("#stage canvas").first().screenshot()).byteLength,
      { timeout: 60_000 },
    )
    .toBeGreaterThan(100_000);
}
