import { expect, test } from "@playwright/test";
import { openMenu, startFirstBattle } from "./battle.ts";

const maxTransfer = 8_000_000;

test("the first battle is playable within the transfer budget", async ({
  page,
  browserName,
  isMobile,
}) => {
  test.skip(
    browserName !== "chromium" || isMobile,
    "Transfer sizes come from Chromium's DevTools protocol at DPR 1.",
  );
  const cdp = await page.context().newCDPSession(page);
  await cdp.send("Network.enable");
  let bytes = 0,
    requests = 0;
  cdp.on("Network.loadingFinished", (event) => {
    bytes += event.encodedDataLength;
    requests++;
  });
  await openMenu(page);
  const menu = bytes;
  await startFirstBattle(page);
  // Let the map's visible lite art arrive; the offline keep waits 15 s and stays out of this.
  await page.waitForLoadState("networkidle");
  console.log(
    `first playable: ${(bytes / 1e6).toFixed(2)} MB in ${requests} requests (menu ${(menu / 1e6).toFixed(2)} MB)`,
  );
  expect(bytes).toBeLessThanOrEqual(maxTransfer);
});
