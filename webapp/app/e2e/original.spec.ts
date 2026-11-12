import { existsSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test } from "@playwright/test";
import { battleReady, expectMapDrawn, openMenu } from "./battle.ts";

// The player's own copy of the original game (git-ignored original/, absent in CI): imported through the menu's
// folder picker, then original map 01 played to its outcome by a human who only ends each phase.
const folder = resolve(
  import.meta.dirname,
  "../../../original/extracted/dos-full",
);

test("an imported original map plays through to its outcome page", async ({
  page,
  browserName,
}) => {
  test.skip(
    !existsSync(resolve(folder, "CODES.DAT")),
    "no local original files",
  );
  test.skip(browserName !== "chromium", "folder upload is checked in Chromium");
  test.setTimeout(30 * 60_000);
  const failures: string[] = [];
  page.on("pageerror", (error) => failures.push(error.message));

  await openMenu(page);
  await page.getByRole("button", { name: "Extras" }).click();
  await page.getByRole("button", { name: "Original game" }).click();
  await page
    .getByLabel("Choose the original game's folder")
    .setInputFiles(folder);
  await expect(page.getByRole("status").first()).toContainText(
    "72 original maps imported, with the original tiles and units",
    { timeout: 60_000 },
  );
  await page.getByRole("button", { name: /^01 · / }).click();
  await expect(page.getByRole("status").first()).toContainText(
    "Original map 01",
  );
  await page.getByRole("button", { name: "Start" }).click();
  await page.getByRole("button", { name: /^Germany/ }).click();
  await expect(battleReady(page)).toBeVisible({ timeout: 120_000 });
  await expectMapDrawn(page);
  await expect(
    page.locator("button", { hasText: "Map design: Original" }),
  ).toHaveCount(1);

  const ended = page.locator(".battle-ended button");
  const confirm = page.getByRole("button", { name: "Confirm (F1)" });
  const skip = page.locator(".battle-host-close");
  const ready = battleReady(page);
  let phases = 0;
  while (!(await ended.isVisible())) {
    if (await skip.isVisible())
      await skip.click({ timeout: 2_000 }).catch(() => undefined);
    else if (await confirm.isVisible())
      await confirm.click({ timeout: 2_000 }).catch(() => undefined);
    else if (
      (await ready.isEnabled()) &&
      (await ready.innerText()).startsWith("Ready")
    ) {
      await ready.click({ timeout: 2_000 }).catch(() => undefined);
      phases++;
    } else await page.waitForTimeout(100);
  }
  console.log(
    `original/01 ended after ${phases} phases: ${await page.locator(".battle-ended").innerText()}`,
  );
  await ended.click();
  await expect(
    page.getByRole("heading", { name: /VICTORY|DEFEAT|DRAW/ }),
  ).toBeVisible();
  expect(failures).toEqual([]);
});
