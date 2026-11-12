import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { test, expect } from "@playwright/test";
import { build } from "vite";
import { cacheDir } from "../../cachedir.ts";
import type { AssetLoader } from "../src/assets.ts";
import type { SplitFrame } from "../src/split/frame.ts";
import type { layout } from "../src/split/geometry.ts";

interface Fixture {
  SplitFrame: typeof SplitFrame;
  layout: typeof layout;
}
interface FixtureWindow extends Window {
  __splitFrameFixture: Fixture;
  __splitFrameBaseline?: Fixture;
}

let fixture: string;
let baseline: string | undefined;

test.beforeAll(async () => {
  const frame = resolve(import.meta.dirname, "../src/split/frame.ts"),
    geometry = resolve(import.meta.dirname, "../src/split/geometry.ts");
  const result = await build({
    configFile: false,
    root: resolve(import.meta.dirname, ".."),
    cacheDir: cacheDir("split-frame"),
    publicDir: false,
    logLevel: "error",
    plugins: [
      {
        name: "split-frame-fixture",
        resolveId(id) {
          if (id === "virtual:split-frame") return "\0split-frame";
        },
        load(id) {
          if (id === "\0split-frame")
            return `import {SplitFrame} from ${JSON.stringify(frame)};
import {layout} from ${JSON.stringify(geometry)};
globalThis.__splitFrameFixture={SplitFrame,layout};`;
        },
      },
    ],
    build: {
      write: false,
      minify: false,
      target: "es2022",
      rolldownOptions: {
        input: "virtual:split-frame",
        output: { format: "es" },
      },
    },
  });
  if (Array.isArray(result) || !("output" in result))
    throw new Error("Missing split frame fixture bundle");
  const chunk = result.output.find((output) => output.type === "chunk");
  if (chunk === undefined) throw new Error("Missing split frame fixture code");
  fixture = chunk.code;
  // An optional, independently bundled pre-change frame verifies exact pixels
  // during optimization; ordinary runs compare reused and fresh real frames.
  const path = process.env["HISTORYLINE_SPLIT_FRAME_BASELINE"];
  if (path !== undefined)
    baseline = (await readFile(path, "utf8")).replaceAll(
      "__splitFrameFixture",
      "__splitFrameBaseline",
    );
});

test("reuses the slim frame material while retaining every pixel across pans and layouts", async ({
  page,
}) => {
  await page.route("**/__split-frame-fixture", (route) =>
    route.fulfill({ contentType: "text/html", body: "<!doctype html>" }),
  );
  await page.goto("./__split-frame-fixture");
  await page.addScriptTag({ type: "module", content: fixture });
  if (baseline !== undefined)
    await page.addScriptTag({ type: "module", content: baseline });
  const result = await page.evaluate(() => {
    const host = window as unknown as FixtureWindow,
      current = host.__splitFrameFixture,
      reference = host.__splitFrameBaseline ?? current,
      // paint() does not load assets; use the real deterministic materials.
      frame = new current.SplitFrame({} as AssetLoader, () => {}),
      created = new Set<HTMLCanvasElement>();
    // These methods are invoked with their owning context through call().
    // eslint-disable-next-line @typescript-eslint/unbound-method
    const originalFill = CanvasRenderingContext2D.prototype.fillRect;
    // eslint-disable-next-line @typescript-eslint/unbound-method
    const originalGet = HTMLCanvasElement.prototype.getContext;
    let recording = false,
      fullPatternFills = 0;
    HTMLCanvasElement.prototype.getContext = function (
      this: HTMLCanvasElement,
      type: string,
      options?: unknown,
    ) {
      created.add(this);
      return (
        originalGet as (
          this: HTMLCanvasElement,
          type: string,
          options?: unknown,
        ) => RenderingContext | null
      ).call(this, type, options);
    } as typeof HTMLCanvasElement.prototype.getContext;
    CanvasRenderingContext2D.prototype.fillRect = function (x, y, w, h) {
      if (
        recording &&
        this.fillStyle instanceof CanvasPattern &&
        x === 0 &&
        y === 0 &&
        w === this.canvas.width &&
        h === this.canvas.height
      )
        fullPatternFills++;
      originalFill.call(this, x, y, w, h);
    };
    const pixelsOf = (canvas: HTMLCanvasElement) => {
      const ctx = canvas.getContext("2d");
      if (ctx === null) throw new Error("Missing fixture canvas context");
      return ctx.getImageData(0, 0, canvas.width, canvas.height).data;
    };
    const grid = { width: 11, lines: 9 },
      differences: number[] = [],
      fills: number[] = [],
      movingMasks: number[] = [];
    try {
      for (const dpr of [1, 2, 3]) {
        for (const resized of [false, true]) {
          const screen = current.layout(
            (resized ? 440 : 640) * dpr,
            (resized ? 320 : 360) * dpr,
            dpr,
          );
          const before = fullPatternFills;
          let previous: Uint8ClampedArray | undefined;
          for (const [pair, row] of [
            [0, 0],
            [4, 3],
            [-4, -2],
          ] as const) {
            const views: Parameters<SplitFrame["paint"]>[1] = [
              screen.view(0, pair, row),
              screen.view(1, -pair, -row),
            ];
            recording = true;
            frame.paint(screen, views, grid, true, dpr);
            recording = false;
            const fresh = new reference.SplitFrame({} as AssetLoader, () => {});
            fresh.paint(screen, views, grid, true, dpr);
            const pixels = pixelsOf(frame.canvas),
              expected = pixelsOf(fresh.canvas);
            differences.push(
              pixels.reduce(
                (count, value, i) => count + Number(value !== expected[i]),
                0,
              ),
            );
            if (previous !== undefined)
              movingMasks.push(
                pixels.reduce(
                  (count, value, i) => count + Number(value !== previous?.[i]),
                  0,
                ),
              );
            previous = pixels;
            fresh.dispose();
          }
          fills.push(fullPatternFills - before);
          const views: Parameters<SplitFrame["paint"]>[1] = [
            screen.view(0, 2, 1),
            screen.view(1, 3, 2),
          ];
          for (const slim of [false, true]) {
            frame.paint(screen, views, grid, slim, dpr);
            const fresh = new reference.SplitFrame({} as AssetLoader, () => {});
            fresh.paint(screen, views, grid, slim, dpr);
            const pixels = pixelsOf(frame.canvas),
              expected = pixelsOf(fresh.canvas);
            differences.push(
              pixels.reduce(
                (count, value, i) => count + Number(value !== expected[i]),
                0,
              ),
            );
            fresh.dispose();
          }
        }
      }
    } finally {
      recording = false;
      frame.dispose();
      HTMLCanvasElement.prototype.getContext = originalGet;
      CanvasRenderingContext2D.prototype.fillRect = originalFill;
    }
    return {
      differences,
      fills,
      movingMasks,
      retainedCanvases: [...created].filter(
        (canvas) => canvas.width * canvas.height,
      ).length,
    };
  });
  expect(result.differences).toEqual(Array<number>(30).fill(0));
  expect(result.fills).toEqual([1, 1, 1, 1, 1, 1]);
  expect(result.movingMasks.every((count) => count > 0)).toBe(true);
  expect(result.retainedCanvases).toBe(0);
});
