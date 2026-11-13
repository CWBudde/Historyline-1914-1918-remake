import { devices, expect, test, type Page } from "@playwright/test";
import { openMenu } from "./battle.ts";

/** The CSS pixels of one unit of the shell's layout, as shell.css sets --u. */
async function unit(page: Page): Promise<number> {
  return page.evaluate(() => {
    const probe = document.createElement("div");
    probe.style.width = "calc(100 * var(--u))";
    document.querySelector(".shell-ui")?.append(probe);
    const u = probe.getBoundingClientRect().width / 100;
    probe.remove();
    return u;
  });
}
/** Waits until the plate, the rivets and both nameplate states made from the armour's art are on the page. */
async function armour(page: Page): Promise<void> {
  await expect
    .poll(
      () =>
        page
          .locator(".shell-ui")
          .evaluate((el) =>
            ["--hl-plate", "--hl-nameplate", "--hl-nameplate-polished"].every(
              (name) =>
                el.style.getPropertyValue(name).startsWith('url("blob:'),
            ),
          ),
      { timeout: 30_000 },
    )
    .toBe(true);
}
/** The painted face of a nameplate: its border image. */
function face(page: Page, name: string | RegExp) {
  return page
    .getByRole("button", { name })
    .evaluate((el) => getComputedStyle(el).borderImageSource);
}

test.describe("the menus on the armour plate", () => {
  test("the start screen: stencilled title, brass nameplates, the embrasure and the hint strip", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await openMenu(page);
    const u = await unit(page);
    expect(u).toBeCloseTo(2.7, 1);
    await armour(page);
    await expect(
      page.getByRole("img", { name: "Historyline 1914–1918" }),
    ).toBeVisible();
    await expect(page.getByRole("img", { name: "Germany" })).toBeVisible();
    await expect(page.getByRole("img", { name: "France" })).toBeVisible();

    // Five nameplates in a column at the left; a browser page has no Exit.
    const names = [
      "Start",
      /^Code/,
      "Game values",
      "Control",
      "Extras",
    ] as const;
    await expect(page.getByRole("button", { name: "Exit" })).toHaveCount(0);
    let last = 0;
    for (const name of names) {
      const box = await page.getByRole("button", { name }).boundingBox();
      expect(box?.x).toBeCloseTo((1920 - 712 * u) / 2 + 40 * u, 0);
      expect(box?.width).toBeCloseTo(158 * u, 0);
      expect(box?.height).toBeCloseTo(28 * u, 0);
      expect(box?.y ?? 0).toBeGreaterThan(last);
      last = box?.y ?? 0;
    }
    // The selected plate is rubbed bright, the others dull; the lettering is engraved in dark ink.
    const start = page.getByRole("button", { name: "Start" }),
      code = page.getByRole("button", { name: /^Code/ });
    await expect(start).toBeFocused();
    const dull = await face(page, /^Code/),
      bright = await face(page, "Start");
    expect(dull).toMatch(/^url\("blob:/);
    expect(bright).toMatch(/^url\("blob:/);
    expect(bright).not.toBe(dull);
    expect(await start.evaluate((el) => getComputedStyle(el).color)).toBe(
      "rgb(27, 21, 11)",
    );
    expect(
      await start.evaluate((el) => getComputedStyle(el).fontFamily),
    ).toMatch(/^"Zilla Slab"/);
    await expect(page.locator(".shell-help")).toContainText("Begin the battle");
    await page.keyboard.press("ArrowDown");
    await expect(code).toBeFocused();
    expect(await face(page, /^Code/)).toBe(bright);
    expect(await face(page, "Start")).toBe(dull);
    await expect(page.locator(".shell-help")).toContainText("Type the code");

    // The embrasure at the right shows the menu painting in its opening; the hint strip lies above it.
    const frame = await page.locator(".shell-embrasure").boundingBox(),
      view = await page.locator(".shell-embrasure-view").boundingBox(),
      help = await page.locator(".shell-help").boundingBox();
    expect(frame?.width).toBeCloseTo(470 * u, 0);
    expect(frame?.height).toBeCloseTo((470 * u * 967) / 1494, 0);
    expect(view?.width ?? 0).toBeGreaterThan(0.85 * (frame?.width ?? 0));
    expect((help?.y ?? 0) + (help?.height ?? 0)).toBeLessThan(frame?.y ?? 0);
    await expect(
      page.locator(".shell-embrasure-view img.shell-scene"),
    ).toBeVisible();

    // The country choice: a stencilled heading and the two sides' plaques, named on brass plates.
    await page.getByRole("button", { name: "Start" }).click();
    await expect(
      page.getByRole("heading", { name: "SELECT COUNTRY" }),
    ).toBeVisible();
    const cards = page.locator(".shell-country");
    await expect(cards).toHaveCount(2);
    await expect(cards.locator(".shell-nameplate")).toHaveCount(2);
    const card = await cards.first().boundingBox();
    expect(card?.height).toBeGreaterThanOrEqual(238 * u - 1);
  });

  test("the front page on newsprint and the new units in wells on the plate", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await openMenu(page);
    const u = await unit(page);
    await page.getByRole("button", { name: "Start" }).click();
    await page.getByRole("button", { name: /^Germany/ }).click();

    const sheet = page.locator(".shell-page-news .shell-sheet");
    await expect
      .poll(
        () => sheet.evaluate((el) => getComputedStyle(el).backgroundImage),
        { timeout: 30_000 },
      )
      .toMatch(/^url\("blob:/);
    const paper = await sheet.boundingBox();
    expect(paper?.width).toBeCloseTo(616 * u, -0.5);
    // Within a pixel and a half: --u is a fraction of a pixel. The picture is 240 by 160; Continue stands under the orders in its column.
    const photo = await page.locator(".shell-news-photo").boundingBox();
    expect(photo?.width).toBeCloseTo(240 * u, -0.5);
    expect(photo?.height).toBeCloseTo(160 * u, -0.5);
    const go = page.getByRole("button", { name: "Continue" });
    await expect(
      page.locator(".shell-news-picture").getByRole("button", {
        name: "Continue",
      }),
    ).toBeVisible();
    const orders = await page.locator(".shell-orders").boundingBox(),
      button = await go.boundingBox();
    expect(button?.y ?? 0).toBeGreaterThan((orders?.y ?? 0) + 1);
    expect((button?.y ?? 0) + (button?.height ?? 0)).toBeLessThanOrEqual(
      (paper?.y ?? 0) + (paper?.height ?? 0),
    );
    expect(
      await page
        .locator(".shell-news-copy")
        .evaluate((el) => getComputedStyle(el).columnCount),
    ).toBe("2");

    await go.click();
    await expect(page.getByRole("button", { name: "To battle" })).toBeVisible();
    await armour(page);
    const wells = page.locator(".shell-unit.shell-well");
    expect(await wells.count()).toBeGreaterThan(0);
    expect(await wells.count()).toBeLessThanOrEqual(4);
    const portrait = await page
      .locator(".shell-portrait")
      .first()
      .boundingBox();
    expect(portrait?.width).toBeCloseTo(69 * u, -0.5);
    expect(portrait?.height).toBeCloseTo(62 * u, -0.5);
    // The buttons beside the title, as in the desktop game.
    const title = await page
        .getByRole("heading", { name: "NEW AT THE FRONT" })
        .boundingBox(),
      battle = await page
        .getByRole("button", { name: "To battle" })
        .boundingBox();
    expect(battle?.y ?? 0).toBeLessThan((title?.y ?? 0) + (title?.height ?? 0));
    expect(battle?.x ?? 0).toBeGreaterThan(960);
  });

  test.describe("on a phone held upright", () => {
    const pixel = devices["Pixel 7"];
    // The device's browser type would force a worker of its own inside a describe group.
    test.use({
      viewport: pixel.viewport,
      deviceScaleFactor: pixel.deviceScaleFactor,
      isMobile: pixel.isMobile,
      hasTouch: pixel.hasTouch,
      userAgent: pixel.userAgent,
    });
    test("the title at the top, an embrasure strip, a column of touch-sized nameplates", async ({
      page,
    }) => {
      await openMenu(page);
      await armour(page);
      const view = page.viewportSize();
      if (view === null) throw Error("no viewport");
      const title = await page
          .getByRole("img", { name: "Historyline 1914–1918" })
          .boundingBox(),
        frame = await page.locator(".shell-embrasure").boundingBox(),
        start = await page.getByRole("button", { name: "Start" }).boundingBox();
      expect(title?.y ?? 99).toBeLessThan(40);
      expect(frame?.y ?? 0).toBeGreaterThan(
        (title?.y ?? 0) + (title?.height ?? 0),
      );
      expect(frame?.width ?? 0).toBeGreaterThan(view.width * 0.7);
      expect(start?.y ?? 0).toBeGreaterThan(
        (frame?.y ?? 0) + (frame?.height ?? 0),
      );
      expect(
        await page.evaluate(
          () =>
            document.querySelector(".shell-ui")?.scrollWidth ===
            document.querySelector(".shell-ui")?.clientWidth,
        ),
      ).toBe(true);
      for (const name of ["Start", "Game values", "Control", "Extras"]) {
        const box = await page.getByRole("button", { name }).boundingBox();
        expect(box?.height, name).toBeGreaterThanOrEqual(48);
        expect(box?.x ?? 0, name).toBeGreaterThanOrEqual(0);
        expect((box?.x ?? 0) + (box?.width ?? 0), name).toBeLessThanOrEqual(
          view.width,
        );
      }
      await page.getByRole("button", { name: "Game values" }).tap();
      await expect(
        page.getByRole("heading", { name: "GAME VALUES" }),
      ).toBeVisible();
      await page.getByRole("button", { name: "Back" }).tap();
      await expect(page.getByRole("button", { name: "Start" })).toBeVisible();
    });
  });
});
