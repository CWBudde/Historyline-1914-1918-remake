import { expect, test } from "@playwright/test";
import {
  battleReady,
  expectMapDrawn,
  openMenu,
  startFirstBattle,
} from "./battle.ts";

// A human who only ends each phase: the real UI and the production bundle, with no test hooks.
test("a level plays through to its outcome page", async ({ page }) => {
  test.setTimeout(30 * 60_000);
  const failures: string[] = [];
  page.on("pageerror", (error) => failures.push(error.message));

  await openMenu(page);
  await startFirstBattle(page);
  await expectMapDrawn(page);

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
    `ended after ${phases} phases: ${await page.locator(".battle-ended").innerText()}`,
  );
  await ended.click();
  await expect(
    page.getByRole("heading", { name: /VICTORY|DEFEAT|DRAW/ }),
  ).toBeVisible();
  expect(failures).toEqual([]);
});
