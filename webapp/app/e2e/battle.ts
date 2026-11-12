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
/**
 * battleReady is the player's end-of-phase control in either layout: the single map's Ready button or, on the
 * split screen (the default on wide windows), the status well of the player's panel (the Central Powers').
 */
export function battleReady(page: Page) {
  return page
    .locator('.battle-ready:visible, .split-well[data-side="0"]:visible')
    .first();
}
/** canEndPhase reports whether the player's end-of-phase control offers Ready now. */
export async function canEndPhase(page: Page): Promise<boolean> {
  const ready = battleReady(page);
  if (!(await ready.isVisible()) || !(await ready.isEnabled())) return false;
  const label =
    (await ready.getAttribute("aria-label")) ?? (await ready.innerText());
  return /^Ready|· Ready$/.test(label.trim());
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
