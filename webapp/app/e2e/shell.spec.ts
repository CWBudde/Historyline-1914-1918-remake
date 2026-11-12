import { devices, expect, test, type Page } from "@playwright/test";
import { openMenu } from "./battle.ts";

/** The CSS pixels of one reference pixel of the desktop game's 640×400 menus, as shell.css sets --u. */
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
/** Waits until the steel tile and the rivet head made from the art are on the column. */
async function steel(page: Page): Promise<string> {
  const sheet = page.locator(".shell-sheet");
  await expect
    .poll(
      () =>
        sheet.evaluate((el) => {
          const image = getComputedStyle(el).backgroundImage;
          return /url\("blob:.*url\("blob:/.test(image) ? image : "";
        }),
      { timeout: 30_000 },
    )
    .not.toBe("");
  return sheet.evaluate((el) => getComputedStyle(el).backgroundImage);
}

test.describe("the menus in the desktop game's steel", () => {
  test("a riveted steel column with steel buttons, focus and help", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await openMenu(page);
    const u = await unit(page);
    expect(u).toBeCloseTo(2.7, 1);
    const images = await steel(page);
    // Four rivet heads and the graded tile.
    expect(images.match(/url\(/g)).toHaveLength(5);
    const column = await page.locator(".shell-sheet").boundingBox();
    expect(column?.x).toBe(0);
    expect(column?.width).toBeCloseTo(304 * u, 0);
    expect(column?.height).toBeCloseTo(1080, 0);

    const start = page.getByRole("button", { name: "Start" }),
      code = page.getByRole("button", { name: /^Code/ });
    await expect(start).toBeFocused();
    const box = await start.boundingBox();
    expect(box?.width).toBeCloseTo(216 * u, 0);
    expect(box?.height).toBeGreaterThanOrEqual(30 * u - 1);
    // The focused button is amber in brass lines, the others cream.
    const colour = (b: typeof start) =>
      b.evaluate((el) => getComputedStyle(el).color);
    expect(await colour(start)).toBe("rgb(240, 184, 88)");
    expect(await colour(code)).toBe("rgb(251, 227, 199)");
    expect(
      await start.evaluate((el) => getComputedStyle(el).textTransform),
    ).toBe("uppercase");
    await expect(page.locator(".shell-help")).toContainText("Begin the battle");
    await page.keyboard.press("ArrowDown");
    await expect(code).toBeFocused();
    expect(await colour(code)).toBe("rgb(240, 184, 88)");
    await expect(page.locator(".shell-help")).toContainText("Type the code");

    // Exit stands at the painting's foot, right.
    const exit = await page.getByRole("button", { name: "Exit" }).boundingBox();
    expect((exit?.x ?? 0) + (exit?.width ?? 0)).toBeCloseTo(1920 - 24 * u, 0);
    expect((exit?.y ?? 0) + (exit?.height ?? 0)).toBeCloseTo(1080 - 24 * u, 0);

    // The country choice: a name plate and two riveted plates.
    await start.click();
    await expect(page.locator(".shell-name-plate")).toContainText(
      "SELECT COUNTRY",
    );
    const cards = page.locator(".shell-country");
    await expect(cards).toHaveCount(2);
    const card = await cards.first().boundingBox();
    expect(card?.height).toBeGreaterThanOrEqual(238 * u - 1);
  });

  test("the front page on newsprint and the new units in wells on hull plating", async ({
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
    await expect(page.locator(".shell-backdrop.shell-hull")).toHaveCount(1);
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
    test("the painting in a band above a full-width column, touch-sized buttons", async ({
      page,
    }) => {
      await openMenu(page);
      await steel(page);
      const view = page.viewportSize();
      if (view === null) throw Error("no viewport");
      const band = await page.locator(".shell-backdrop").boundingBox(),
        column = await page.locator(".shell-sheet").boundingBox();
      // The whole painting's width at its own aspect, at most two fifths of the height.
      expect(band?.height).toBeCloseTo(
        Math.min(view.width * 0.625, view.height * 0.4),
        0,
      );
      expect(column?.width).toBeCloseTo(view.width, 0);
      expect(column?.y).toBeCloseTo(band?.height ?? 0, 0);
      expect(
        await page.evaluate(
          () =>
            document.querySelector(".shell-ui")?.scrollWidth ===
            document.querySelector(".shell-ui")?.clientWidth,
        ),
      ).toBe(true);
      for (const name of [
        "Start",
        "Game values",
        "Control",
        "Extras",
        "Exit",
      ]) {
        const box = await page.getByRole("button", { name }).boundingBox();
        expect(box?.height, name).toBeGreaterThanOrEqual(44);
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
