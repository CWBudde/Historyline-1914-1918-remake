import { expect, test } from "@playwright/test";
import { openMenu, startFirstBattle } from "./battle.ts";

const budgetMs = 16.7;

// Needs a real GPU, which CI lacks: E2E_FRAMES=1 pnpm --dir webapp run e2e --project=chromium --headed frames
test("panning and zooming the map stays within the frame budget", async ({
  page,
}) => {
  test.skip(
    process.env["E2E_FRAMES"] === undefined,
    "Set E2E_FRAMES=1 on a machine with a GPU.",
  );
  await openMenu(page);
  await startFirstBattle(page);
  await page.waitForLoadState("networkidle");
  const box = page.viewportSize();
  if (box === null) throw Error("No viewport.");
  const x = box.width / 2,
    y = box.height / 2;
  const sampling = page.evaluate(
    () =>
      new Promise<number[]>((resolve) => {
        const deltas: number[] = [];
        let last = performance.now();
        const frame = (now: number) => {
          deltas.push(now - last);
          last = now;
          if (deltas.length < 600) requestAnimationFrame(frame);
          else resolve(deltas.slice(30));
        };
        requestAnimationFrame(frame);
      }),
  );
  const state = { running: true };
  void sampling.finally(() => {
    state.running = false;
  });
  await page.mouse.move(x, y);
  for (let step = 0; state.running; step++) {
    await page.mouse.wheel(0, step % 40 < 20 ? -60 : 60);
    await page.mouse.down();
    await page.mouse.move(
      x + 120 * Math.sin(step / 5),
      y + 80 * Math.cos(step / 5),
    );
    await page.mouse.up();
  }
  const deltas = (await sampling).sort((a, b) => a - b);
  const p95 = deltas[Math.floor(deltas.length * 0.95)] ?? Infinity;
  console.log(
    `frames: ${deltas.length}, p95 ${p95.toFixed(1)} ms, median ${(deltas[deltas.length >> 1] ?? 0).toFixed(1)} ms`,
  );
  expect(p95).toBeLessThanOrEqual(budgetMs);
});
