import { expect, test, type Page } from "@playwright/test";
import { openMenu, seed, startFirstBattle } from "./battle.ts";

// The battle screen in the menus' worn army metal: steel housing, brass plates, stencilled status, no pixel font
// with the painted design. The production bundle, the first battle.

/** The old battle UI's blue-grey gradients and sheets, none of which may be left. */
const oldBlue = [
  "rgb(53, 75, 85)",
  "rgb(27, 43, 52)",
  "rgb(52, 73, 82)",
  "rgb(23, 40, 48)",
  "rgb(42, 62, 71)",
  "rgb(26, 44, 53)",
  "rgb(28, 45, 53)",
  "rgb(53, 73, 81)",
  "rgb(25, 42, 51)",
];

/** blueElements lists the battle page's elements still painted in the old blue. */
async function blueElements(page: Page): Promise<string[]> {
  return page.evaluate((colours) => {
    const out: string[] = [];
    for (const element of document.querySelectorAll<HTMLElement>(
      ".battle-page *",
    )) {
      const style = getComputedStyle(element),
        paint = `${style.backgroundImage} ${style.backgroundColor}`;
      if (colours.some((colour) => paint.includes(colour)))
        out.push(`${element.tagName}.${element.className}`);
    }
    return out;
  }, oldBlue);
}

test.describe("the battle screen's armour", () => {
  test.skip(
    ({ browserName, isMobile }) => browserName !== "chromium" || isMobile,
    "One engine checks the layout; the phone has its own test below.",
  );

  test("split screen: a strip under each window, the menu board, steel and stencil only", async ({
    page,
  }) => {
    test.setTimeout(5 * 60_000);
    const failures: string[] = [];
    page.on("pageerror", (error) => failures.push(error.message));
    await page.setViewportSize({ width: 1600, height: 900 });
    await page.addInitScript(seed({ scenes: false }));
    await openMenu(page);
    await startFirstBattle(page);

    // The painted design: the steel housing, no pixel-font wells or overlays.
    await expect(page.locator(".split-screen")).toHaveAttribute(
      "data-hud",
      "steel",
    );
    await expect(
      page.locator(".split-well-canvas, .split-overlay"),
    ).toHaveCount(0);
    expect(
      await page
        .locator('.split-well[data-side="0"]')
        .evaluate((e) => getComputedStyle(e).fontFamily),
    ).toContain("Stardos Stencil");

    // A strip under each window, as wide as it.
    for (const side of [0, 1]) {
      const panel = await page
          .locator(`.split-panel[data-side="${side}"]`)
          .boundingBox(),
        strip = await page
          .locator(`.split-steel-strip[data-side="${side}"]`)
          .boundingBox();
      if (panel === null || strip === null) throw Error("no panel or strip");
      expect(strip.y).toBeGreaterThanOrEqual(panel.y + panel.height);
      expect(strip.y + strip.height).toBeLessThanOrEqual(900);
      expect(Math.abs(strip.x - panel.x)).toBeLessThan(1);
      expect(Math.abs(strip.width - panel.width)).toBeLessThan(1);
    }
    // The player's strip carries the plates; every plate lies inside its strip.
    const own = page.locator('.split-steel-strip[data-side="0"]'),
      ownBox = await own.boundingBox();
    if (ownBox === null) throw Error("no strip");
    const plates = own.locator(".hl-nameplate");
    await expect(plates).toHaveCount(4);
    for (const plate of await plates.all()) {
      const box = await plate.boundingBox();
      if (box === null) throw Error("no plate");
      expect(box.x).toBeGreaterThanOrEqual(ownBox.x);
      expect(box.x + box.width).toBeLessThanOrEqual(ownBox.x + ownBox.width);
    }
    await expect(
      page.locator('.split-steel-strip[data-side="1"] .hl-nameplate'),
    ).toHaveCount(0);

    // Menu opens its board over the player's window, its choices as plates; Escape closes it.
    const menu = page.locator("#battle-menu-0");
    await expect(menu).toBeHidden();
    await own.getByRole("button", { name: "Menu" }).click();
    await expect(menu).toBeVisible();
    await expect(
      menu.getByRole("button", { name: /^UI scale: 100%$/ }),
    ).toBeFocused();
    for (const name of [
      /^Layout: /,
      /^Motion: /,
      /^Battle scenes: /,
      /^Map design: Painted$/,
      "Fullscreen",
      "Keys and touch",
      "Surrender",
      "Leave battle",
    ])
      await expect(menu.getByRole("button", { name })).toBeVisible();
    const panel = await page
        .locator('.split-panel[data-side="0"]')
        .boundingBox(),
      board = await menu.boundingBox();
    if (panel === null || board === null) throw Error("no board");
    expect(board.x).toBeGreaterThanOrEqual(panel.x);
    expect(board.x + board.width).toBeLessThanOrEqual(panel.x + panel.width);
    expect(await page.locator("select, input[type=checkbox]").count()).toBe(0);
    await page.keyboard.press("Escape");
    await expect(menu).toBeHidden();

    // Info opens the situation on a steel board in the player's window.
    await page.keyboard.press("Escape");
    await expect(
      page.locator('.split-panel[data-side="0"] .battle-board-game'),
    ).toBeVisible();
    await expect(
      page.getByRole("heading", { name: "Situation" }),
    ).toBeVisible();
    expect(await blueElements(page)).toEqual([]);
    await page.keyboard.press("Escape");
    expect(failures).toEqual([]);
  });

  test("single map: the strip under the opening, the side board and steel only", async ({
    page,
  }) => {
    test.setTimeout(5 * 60_000);
    await page.setViewportSize({ width: 1000, height: 800 });
    await page.addInitScript(seed({ scenes: false }));
    await openMenu(page);
    await startFirstBattle(page);
    const map = await page.locator("#stage canvas").first().boundingBox(),
      bar = await page.locator(".battle-bar").boundingBox();
    if (map === null || bar === null) throw Error("no map or strip");
    // The opening is framed by plate on every side, the strip under it.
    expect(map.x).toBeGreaterThan(4);
    expect(map.y).toBeGreaterThan(4);
    expect(map.x + map.width).toBeLessThan(1000 - 4);
    expect(bar.y).toBeGreaterThanOrEqual(map.y + map.height);
    expect(bar.y + bar.height).toBeLessThanOrEqual(800);
    await expect(page.locator(".battle-bar .order-strip")).toBeVisible();
    await expect(page.locator(".battle-dock")).toBeVisible();
    await page.getByRole("button", { name: "Menu" }).click();
    await expect(page.locator("#battle-menu-0")).toBeVisible();
    // The side board gives way to what opens.
    await expect(page.locator(".battle-dock")).toBeHidden();
    expect(await blueElements(page)).toEqual([]);
    await page.keyboard.press("Escape");
    await expect(page.locator("#battle-menu-0")).toBeHidden();
  });
});

test.describe("the battle screen upright", () => {
  test.use({
    hasTouch: true,
    isMobile: true,
    viewport: { width: 390, height: 844 },
  });
  test.skip(
    ({ browserName }) => browserName !== "chromium",
    "Touch input is emulated in Chromium.",
  );
  test("the strip at the top, the plate below the map with touch-sized plates", async ({
    page,
  }) => {
    test.setTimeout(5 * 60_000);
    await page.addInitScript(seed({ scenes: false }));
    await openMenu(page);
    await startFirstBattle(page);
    const top = await page.locator(".battle-bar").boundingBox(),
      map = await page.locator("#stage canvas").first().boundingBox(),
      dock = await page.locator(".battle-dock").boundingBox();
    if (top === null || map === null || dock === null) throw Error("no layout");
    expect(top.y).toBeLessThan(4);
    expect(map.y).toBeGreaterThanOrEqual(top.y + top.height - 1);
    expect(dock.y).toBeGreaterThanOrEqual(map.y + map.height - 1);
    expect(dock.y + dock.height).toBeLessThanOrEqual(844 + 1);
    const menu = page.locator(".battle-bar").getByRole("button", {
      name: "Menu",
    });
    expect((await menu.boundingBox())?.height ?? 0).toBeGreaterThanOrEqual(44);
    await menu.tap();
    await expect(page.locator("#battle-menu-0")).toBeVisible();
    expect(await blueElements(page)).toEqual([]);
  });
});
