import { expect, test, type Page } from "@playwright/test";
import { battleReady, openMenu, seed, startFirstBattle } from "./battle.ts";

// The order menu round a unit and the order strip, in the real UI of the production bundle.

/** ownUnit hovers over the map until the hovered cell holds an own unit (its card shows) and returns that point. */
async function ownUnit(
  page: Page,
  surface: string,
  card: string,
  from = 30,
): Promise<{ x: number; y: number }> {
  const box = await page.locator(surface).first().boundingBox();
  if (box === null) throw Error("no map");
  for (let y = box.y + from; y < box.y + box.height; y += 36)
    for (let x = box.x + 30; x < box.x + box.width; x += 48) {
      await page.mouse.move(x, y);
      if ((await page.locator(card).count()) > 0) return { x, y };
    }
  throw Error("no own unit on the map");
}

const plates = (page: Page) => page.locator(".order-menu .order-plate");

test.describe("the order menu", () => {
  test.skip(
    ({ browserName, isMobile }) => browserName !== "chromium" || isMobile,
    "One engine is enough for the menu's logic; the phone layout has its own test below.",
  );

  test("single map: plates on the neighbours, wedges, caption, keys and the strip", async ({
    page,
  }) => {
    test.setTimeout(5 * 60_000);
    const failures: string[] = [];
    page.on("pageerror", (error) => failures.push(error.message));
    await page.addInitScript(seed({ layout: "single", scenes: false }));
    await openMenu(page);
    await startFirstBattle(page);

    // The strip: Info, Map and End phase as nameplates with their keys; End phase is the end-of-phase control.
    const strip = page.locator(".battle-top .order-strip");
    await expect(strip.locator(".order-nameplate")).toHaveText([
      "IInfo",
      "OMap",
      "F1End phase",
    ]);
    await expect(battleReady(page)).toHaveAttribute("aria-label", "End phase");

    const unit = await ownUnit(
      page,
      "#stage canvas",
      ".battle-dock .battle-unit-card",
      // A unit clear of the opening's upper edge: its rosette stays centred on it and a click beyond its Move
      // plate still lands on the map.
      230,
    );
    // Right-click opens the menu: a Move plate straight above the unit, absent orders leave their cells free.
    await page.mouse.click(unit.x, unit.y, { button: "right" });
    const menu = page.getByRole("menu", { name: "Orders" });
    await expect(menu).toBeVisible();
    const move = menu.getByRole("menuitem", { name: "Move" });
    await expect(move).toHaveAttribute("data-slot", "n");
    const count = await plates(page).count();
    expect(count).toBeGreaterThanOrEqual(1);
    expect(count).toBeLessThanOrEqual(3);
    const box = await move.boundingBox();
    if (box === null) throw Error("no plate");
    // A plate is at least 72 CSS px wide with a mouse and the cell's shape (so at least 60 px tall).
    expect(box.width).toBeGreaterThanOrEqual(72 - 0.5);
    expect(box.height).toBeGreaterThanOrEqual(48);
    expect(box.y + box.height / 2).toBeLessThan(unit.y);

    // Hovering the wedge lights the plate and names it on the caption, with its key.
    await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
    await expect(move).toHaveAttribute("data-lit", "true");
    await expect(page.locator(".order-caption")).toContainText("Move");
    await expect(page.locator(".order-caption kbd")).toHaveText("W");

    // Escape closes it; W runs Move from the keyboard once it is open again.
    await page.keyboard.press("Escape");
    await expect(menu).toHaveCount(0);
    await page.mouse.click(unit.x, unit.y, { button: "right" });
    await expect(menu).toBeVisible();
    await page.keyboard.press("w");
    await expect(menu).toHaveCount(0);
    await expect(page.locator(".battle-status")).toContainText(
      "pick where to move",
    );

    // A click on the selected unit opens its menu; the centre cancels the selection with it.
    await page.mouse.click(unit.x, unit.y);
    await expect(menu).toBeVisible();
    await page.mouse.click(unit.x, unit.y);
    await expect(menu).toHaveCount(0);

    // The wedge, not the plate: a click beyond the plate's own outline but up from the unit runs Move.
    await page.mouse.click(unit.x, unit.y, { button: "right" });
    await expect(menu).toBeVisible();
    // From the rosette's centre (the unit's cell), a neighbour's distance up is the Move plate's centre.
    const centre = await menu
        .getByRole("menuitem", { name: "Cancel" })
        .boundingBox(),
      up = await move.boundingBox();
    if (centre === null || up === null) throw Error("no menu");
    const cx = centre.x + centre.width / 2,
      cy = centre.y + centre.height / 2,
      step = cy - (up.y + up.height / 2);
    await page.mouse.click(cx + step * 0.3, cy - step * 1.6);

    await expect(menu).toHaveCount(0);
    await expect(page.locator(".battle-status")).toContainText(
      "pick where to move",
    );
    await page.keyboard.press("Escape");

    // Enter on the selected unit opens the menu from the keyboard, its first plate focused; arrows walk the focus.
    await page.mouse.click(unit.x, unit.y);
    await page.keyboard.press("Enter");
    await expect(menu).toBeVisible();
    await expect(move).toBeFocused();
    await page.keyboard.press("ArrowDown");
    await expect(menu.getByRole("menuitem", { name: "Cancel" })).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(menu).toHaveCount(0);

    // End phase from the strip asks first; it is never on a plate round a unit.
    await battleReady(page).click();
    await expect(
      page.getByRole("button", { name: "Confirm (F1)" }),
    ).toBeVisible();
    await page.getByRole("button", { name: "Back to orders" }).click();
    expect(failures).toEqual([]);
  });

  test("an order not possible now stays as a stripped plate that names why", async ({
    page,
  }) => {
    test.setTimeout(5 * 60_000);
    await page.addInitScript(seed({ layout: "single", scenes: false }));
    await openMenu(page);
    await startFirstBattle(page);
    const box = await page.locator("#stage canvas").first().boundingBox();
    if (box === null) throw Error("no map");
    // Find a unit with a disabled order (the first battle's supply car has an empty hold).
    for (let y = box.y + 30; y < box.y + box.height; y += 36)
      for (let x = box.x + 30; x < box.x + box.width; x += 48) {
        await page.mouse.move(x, y);
        if (
          (await page.locator(".battle-dock .battle-unit-card").count()) === 0
        )
          continue;
        await page.mouse.click(x, y, { button: "right" });
        const disabled = page.locator('.order-plate[aria-disabled="true"]');
        if ((await disabled.count()) === 0) {
          // Escape with nothing open would open the game information.
          if ((await plates(page).count()) > 0)
            await page.keyboard.press("Escape");
          continue;
        }
        const name = (await disabled.first().getAttribute("aria-label")) ?? "";
        expect(name).toMatch(/^[A-Z][a-z ]+: ./);
        const r = await disabled.first().boundingBox();
        if (r === null) throw Error("no plate");
        await page.mouse.click(r.x + r.width / 2, r.y + r.height / 2);
        // It does not run: the menu stays, the plate lit, its reason on the caption and in the status.
        await expect(page.getByRole("menu", { name: "Orders" })).toBeVisible();
        await expect(disabled.first()).toHaveAttribute("data-lit", "true");
        await expect(page.locator(".order-caption-reason")).not.toBeEmpty();
        return;
      }
    throw Error("no unit with an order not possible now");
  });

  test("split screen: the menu in the player's panel and the strip on its status strip", async ({
    page,
  }) => {
    test.setTimeout(5 * 60_000);
    await page.setViewportSize({ width: 1600, height: 900 });
    await page.addInitScript(seed({ layout: "original", scenes: false }));
    await openMenu(page);
    await startFirstBattle(page);
    const strip = page.locator(".split-strip");
    await expect(strip).toHaveCount(1);
    await expect(strip.locator(".order-nameplate")).toHaveCount(3);
    // The status well is a status now, not a button that ends the phase.
    await expect(page.locator('.split-well[data-side="0"]')).toHaveAttribute(
      "role",
      "status",
    );
    const unit = await ownUnit(
      page,
      '.split-input[data-side="0"]',
      '.split-panel[data-side="0"] .battle-unit-card',
    );
    await page.mouse.click(unit.x, unit.y, { button: "right" });
    const menu = page.getByRole("menu", { name: "Orders" });
    await expect(menu).toBeVisible();
    await expect(menu.getByRole("menuitem", { name: "Move" })).toHaveAttribute(
      "data-slot",
      "n",
    );
    // A flick: press on the unit, drag up and release runs Move without a click on a plate.
    await page.keyboard.press("Escape");
    await expect(menu).toHaveCount(0);
    await page.mouse.move(unit.x, unit.y);
    await page.mouse.down();
    await page.mouse.move(unit.x, unit.y - 20, { steps: 3 });
    await page.mouse.move(unit.x + 2, unit.y - 50, { steps: 3 });
    await page.mouse.up();
    await expect(menu).toHaveCount(0);
    await expect(
      page.locator('.split-panel[data-side="0"] .battle-status'),
    ).toContainText("pick where to move");
  });
});

test.describe("the order menu on a phone", () => {
  test.use({
    hasTouch: true,
    isMobile: true,
    viewport: { width: 412, height: 915 },
  });
  test.skip(
    ({ browserName }) => browserName !== "chromium",
    "Touch input is emulated in Chromium.",
  );
  test("a tap on an own unit opens touch-sized plates; the strip lies along the bottom edge", async ({
    page,
  }) => {
    test.setTimeout(5 * 60_000);
    await page.addInitScript(seed({ scenes: false }));
    await openMenu(page);
    await startFirstBattle(page);
    // Upright, the global orders are touch-sized plates along the plate's bottom edge, under the thumb.
    const strip = page.locator(".battle-dock .order-strip");
    const bottom = await strip.boundingBox();
    expect((bottom?.y ?? 0) + (bottom?.height ?? 0)).toBeGreaterThan(915 - 80);
    for (const plate of await strip.locator(".order-nameplate").all())
      expect((await plate.boundingBox())?.height ?? 0).toBeGreaterThanOrEqual(
        44,
      );
    const box = await page.locator("#stage canvas").first().boundingBox();
    if (box === null) throw Error("no map");
    for (let y = box.y + 30; y < box.y + box.height; y += 28)
      for (let x = box.x + 25; x < box.x + box.width; x += 34) {
        await page.touchscreen.tap(x, y);
        if ((await plates(page).count()) === 0) {
          if ((await page.locator(".battle-contents").count()) > 0)
            await page.keyboard.press("Escape");
          continue;
        }
        const r = await plates(page).first().boundingBox();
        expect(r?.width ?? 0).toBeGreaterThanOrEqual(58 - 0.5);
        expect(r?.height ?? 0).toBeGreaterThanOrEqual(48);
        await expect(page.locator(".order-menu")).toHaveAttribute(
          "data-touch",
          "true",
        );
        return;
      }
    throw Error("no own unit on the phone's map");
  });
});
