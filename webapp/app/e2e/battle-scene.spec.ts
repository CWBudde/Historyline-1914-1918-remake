import { mkdirSync } from "node:fs";
import { join } from "node:path";
import { expect, test, type Page } from "@playwright/test";
import { battleReady, canEndPhase, openMenu, seed } from "./battle.ts";

// The battle scene fills its well: the steel plate reaches every edge of the host, no blank band above or below,
// from a phone held upright to the split screen. The production bundle; the first battle, ending phases until the
// computer's side fights. BATTLE_SCENE_SHOTS=<dir> also saves the scene mid-battle (BATTLE_SCENE_PREFIX names
// the files).

const shots = process.env["BATTLE_SCENE_SHOTS"],
  prefix = process.env["BATTLE_SCENE_PREFIX"] ?? "";

/**
 * untilScene ends phases (confirming each) until a battle scene opens. The page's clock is paused first and runs
 * on in steps of 50 ms, so the scene is caught within a step of opening, even a reduced one.
 */
async function untilScene(page: Page): Promise<void> {
  const skip = page.locator(".battle-host-close:visible"),
    confirm = page.getByRole("button", { name: "Confirm (F1)" });
  // The page's clock runs a little ahead of its Date.now(): a pause just after it can lie in its past.
  await page.clock.pauseAt((await page.evaluate(() => Date.now())) + 1000);
  for (;;) {
    if (await skip.count()) break;
    if (await confirm.isVisible())
      await confirm.click({ timeout: 2_000 }).catch(() => undefined);
    else if (await canEndPhase(page))
      await battleReady(page)
        .click({ timeout: 2_000 })
        .catch(() => undefined);
    else await page.clock.runFor(50);
  }
  // Two frames to draw the scene.
  await page.clock.runFor(34);
}

/** canvasBox is the scene's canvas and its host. */
async function boxes(page: Page) {
  const host = page.locator(".battle-scene-host:visible").first(),
    canvas = host.locator("canvas");
  const h = await host.boundingBox(),
    c = await canvas.boundingBox();
  if (h === null || c === null) throw Error("no scene host or canvas");
  return { host, canvas, h, c };
}

/** bandRows is the share of the canvas's rows, from the top and from the bottom, that are one flat colour. */
async function flatRows(page: Page): Promise<{ top: number; bottom: number }> {
  const { canvas } = await boxes(page);
  const png = await canvas.screenshot();
  return page.evaluate(async (data) => {
    const img = new Image();
    img.src = `data:image/png;base64,${data}`;
    await img.decode();
    const c = document.createElement("canvas");
    c.width = img.width;
    c.height = img.height;
    const g = c.getContext("2d");
    if (g === null) throw Error("no 2d context");
    g.drawImage(img, 0, 0);
    const px = g.getImageData(0, 0, c.width, c.height).data;
    const flat = (y: number) => {
      let min = 255,
        max = 0;
      for (let x = Math.ceil(c.width * 0.05); x < c.width * 0.95; x++) {
        const i = 4 * (y * c.width + x),
          v = (px[i] ?? 0) + (px[i + 1] ?? 0) + (px[i + 2] ?? 0);
        min = Math.min(min, v);
        max = Math.max(max, v);
      }
      return max - min < 12;
    };
    let top = 0,
      bottom = 0;
    while (top < c.height && flat(top)) top++;
    while (bottom < c.height && flat(c.height - 1 - bottom)) bottom++;
    return { top: top / c.height, bottom: bottom / c.height };
  }, png.toString("base64"));
}

/** checkFills checks that the canvas covers its host and draws to every edge. */
async function checkFills(page: Page, name: string): Promise<void> {
  const { h, c } = await boxes(page);
  expect(Math.abs(c.x - h.x)).toBeLessThan(2);
  expect(Math.abs(c.y - h.y)).toBeLessThan(2);
  expect(Math.abs(c.width - h.width)).toBeLessThan(2);
  expect(Math.abs(c.height - h.height)).toBeLessThan(2);
  const band = await flatRows(page);
  expect(band.top, `${name}: flat rows at the top`).toBeLessThan(0.02);
  expect(band.bottom, `${name}: flat rows at the bottom`).toBeLessThan(0.02);
}

/** shoot saves the scene's frames as the battle plays, a fifth of a second apart, until it closes: the page and the
 * host alone. */
async function shoot(page: Page, name: string): Promise<void> {
  if (shots === undefined) return;
  mkdirSync(shots, { recursive: true });
  const skip = page.locator(".battle-host-close:visible"),
    host = page.locator(".battle-scene-host:visible").first();
  for (let i = 0; i < 16 && (await skip.count()) > 0; i++) {
    const n = String(i).padStart(2, "0");
    await page.screenshot({ path: join(shots, `${prefix}${name}-${n}.png`) });
    await host
      .screenshot({
        path: join(shots, `${prefix}${name}-scene-${n}.png`),
        timeout: 2_000,
      })
      .catch(() => undefined);
    await page.clock.runFor(200);
  }
}

const sizes = [
  { name: "390x844", width: 390, height: 844, mobile: true, layout: "auto" },
  { name: "360x640", width: 360, height: 640, mobile: true, layout: "auto" },
  {
    name: "1000x800-single",
    width: 1000,
    height: 800,
    mobile: false,
    layout: "single",
  },
  {
    name: "1920x1080-split",
    width: 1920,
    height: 1080,
    mobile: false,
    layout: "auto",
  },
] as const;

for (const size of sizes)
  test.describe(`the battle scene at ${size.name}`, () => {
    test.use({
      viewport: { width: size.width, height: size.height },
      isMobile: size.mobile,
      hasTouch: size.mobile,
      deviceScaleFactor: size.mobile ? 2 : 1,
      reducedMotion: "no-preference",
    });
    test.skip(
      ({ browserName }) => browserName !== "chromium",
      "One engine checks the layout at each size.",
    );
    test("the scene fills its well", async ({ page }) => {
      test.setTimeout(10 * 60_000);
      const failures: string[] = [];
      page.on("pageerror", (error) => failures.push(error.message));
      await page.clock.install();
      await page.addInitScript(
        seed({ scenes: true, reduced: false, layout: size.layout }),
      );
      await openMenu(page);
      await page.getByRole("button", { name: "Start" }).click();
      await page.getByRole("button", { name: /^Germany/ }).click();
      await page.getByRole("button", { name: "Continue" }).click();
      await page.getByRole("button", { name: "To battle" }).click();
      await expect(battleReady(page)).toBeVisible({ timeout: 120_000 });
      await untilScene(page);
      // The frames are saved even when the check fails, to see why.
      const fills = await checkFills(page, size.name).then(
        () => null,
        (e: unknown) => (e instanceof Error ? e : new Error(String(e))),
      );
      await shoot(page, size.name);
      if (fills !== null) throw fills;
      expect(failures).toEqual([]);
    });
  });

test.describe("the battle scene with reduced motion", () => {
  test.use({
    viewport: { width: 390, height: 844 },
    isMobile: true,
    hasTouch: true,
    deviceScaleFactor: 2,
  });
  test.skip(
    ({ browserName }) => browserName !== "chromium",
    "One engine checks the reduced scene.",
  );
  test("shows the end at once, fills its well and closes on Skip", async ({
    page,
  }) => {
    test.setTimeout(10 * 60_000);
    await page.clock.install();
    await page.addInitScript(seed({ scenes: true, reduced: true }));
    await openMenu(page);
    await page.getByRole("button", { name: "Start" }).click();
    await page.getByRole("button", { name: /^Germany/ }).click();
    await page.getByRole("button", { name: "Continue" }).click();
    await page.getByRole("button", { name: "To battle" }).click();
    await expect(battleReady(page)).toBeVisible({ timeout: 120_000 });
    await untilScene(page);
    await checkFills(page, "reduced");
    if (shots !== undefined) {
      mkdirSync(shots, { recursive: true });
      await page
        .locator(".battle-scene-host:visible")
        .first()
        .screenshot({
          path: join(shots, `${prefix}390x844-reduced-scene.png`),
        });
    }
    await page.clock.resume();
    await page.locator(".battle-host-close:visible").first().click();
    await expect(page.locator(".battle-scene-host")).toHaveCount(0);
  });
});
