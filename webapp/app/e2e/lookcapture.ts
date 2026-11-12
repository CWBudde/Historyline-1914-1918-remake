// Look comparison captures: the web edition's screens at desktop and phone sizes beside the desktop game's
// (cmd/historyline, driven under Xvfb), into build/review/look-compare/ with one index page.
//
//   cd webapp/app && node e2e/lookcapture.ts [--web] [--go] [--index]
//        [--parts flow,attack,tiles] [--sizes 1920x1080,2560x1440,3840x2160,phone]
//        [--layouts responsive,original] [--levels 1914-08,1914-10,1915-08]
//        [--browser chromium|firefox|webkit] [--soft] [--port 4191]
//
// Without --web/--go/--index all three run. The web part serves build/webapp/site itself (as e2e/server.ts
// does) unless E2E_BASE_URL names a running server; it seeds the app's settings (layout, painted) into its
// IndexedDB before the app starts. The Go part runs bin/historyline (go build when it is missing) on its
// own Xvfb display, sized by xdotool, and drives it with the keyboard; it needs Xvfb, xdotool and
// ImageMagick. Parts: flow = the menus and the opening of 1914-08:cp as Germany; attack = a game forked
// from the level's winning recording after nine resolutions (hl llmplay fork), CP in its attack phase, in
// both front ends; tiles = the remake's tile design (web painted off, desktop -backdrop off) on --levels
// at the default, the closest and the farthest zoom. The web runs on the machine's GPU (Chromium with
// ANGLE on Vulkan); --soft keeps Chromium's software WebGL (SwiftShader) and writes into <layout>-soft, as
// software rendering loads the art so slowly that settled screenshots can still miss tiles and units.
import { execFileSync, spawn, type ChildProcess } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { createServer, type Server } from "node:http";
import { join, resolve } from "node:path";
import {
  chromium,
  firefox,
  webkit,
  devices,
  type Browser,
  type BrowserContext,
  type BrowserContextOptions,
  type Page,
} from "@playwright/test";

const repo = resolve(import.meta.dirname, "../../..");
const out = join(repo, "build/review/look-compare");
const scratch = join(repo, "build/scratch/look-compare");

interface Size {
  name: string;
  width: number; // device pixels
  height: number;
  dpr: number; // the web's; the desktop game draws device pixels
  phone: boolean;
}
const allSizes: Size[] = [
  { name: "1920x1080", width: 1920, height: 1080, dpr: 1, phone: false },
  { name: "2560x1440", width: 2560, height: 1440, dpr: 1, phone: false },
  // A 4K monitor as desktops set it up: 1920×1080 CSS pixels at DPR 2.
  { name: "3840x2160", width: 3840, height: 2160, dpr: 2, phone: false },
  // Pixel 7 portrait: 412×915 CSS pixels at DPR 2.625.
  { name: "phone", width: 1082, height: 2402, dpr: 2.625, phone: true },
];

/** Every screen of the comparison, in play order, with what it shows. */
const flowScreens: [string, string][] = [
  ["menu", "Main menu"],
  ["side", "Side choice"],
  ["news", "Newspaper page"],
  ["brief", "Briefing (new at the front)"],
  ["loading", "Loading the battle"],
  ["battle", "Battle, idle (cursor on the HQ)"],
  ["cursor", "Cursor on an own unit (supply car)"],
  ["move", "Unit selected, move range"],
  ["contents", "Contents of a container (the supply car)"],
  ["unitinfo", "Unit / type information"],
  ["strategic", "Strategic map / overview"],
  ["gameinfo", "Game information (info on the HQ)"],
  ["prompt", "Phase change prompt"],
  ["phase", "Phase change in progress (0.9 s, then every 1.5 s)"],
  ["outcome", "Outcome after surrender"],
  ["attack-idle", "Mid-game (forked save): attack phase"],
  ["attack", "Attack: artillery selected, targets"],
  ["battle-scene", "Battle animation (every 1.2 s)"],
];
const tileLevels: Record<string, [string, string]> = {
  "1914-08": ["PULSE", "summer, river, moat and forest"],
  "1914-10": ["CIVIL", "winter"],
  "1915-08": ["ORKAN", "summer coast, sea and harbour"],
  "1916-02": ["PARTS", "winter coast, sea and harbour"],
};
const tileViews: [string, string][] = [
  ["idle", "default zoom"],
  ["near", "closest zoom (7 wheel steps in)"],
  ["far", "farthest zoom (14 wheel steps out)"],
];

const args = process.argv.slice(2);
function flag(name: string): boolean {
  return args.includes(`--${name}`);
}
function option(name: string, fallback: string): string[] {
  const i = args.indexOf(`--${name}`);
  return (i >= 0 ? (args[i + 1] ?? fallback) : fallback)
    .split(",")
    .filter((s) => s !== "");
}
const runAll = !flag("web") && !flag("go") && !flag("index");
const parts = option("parts", "flow,attack,tiles");
const sizeNames = option("sizes", allSizes.map((s) => s.name).join(","));
const sizes = allSizes.filter((s) => sizeNames.includes(s.name));
const layouts = option("layouts", "responsive,original");
const levels = option("levels", "1914-08,1914-10,1915-08");
const browserName = option("browser", "chromium")[0] ?? "chromium";
const port = Number(option("port", "4191")[0]);

function sleep(ms: number): Promise<void> {
  return new Promise((done) => setTimeout(done, ms));
}
function log(text: string): void {
  console.log(`[look] ${text}`);
}
function first<T>(list: readonly T[]): T {
  const v = list[0];
  if (v === undefined) throw Error("empty list");
  return v;
}

// ---------------------------------------------------------------- the cursor's key steps (internal/gesture Neighbor)

type Cell = [number, number]; // line, column
type Step = "Up" | "Down" | "Left" | "Right";
function neighbour([l, c]: Cell, step: Step): Cell {
  const odd = (l & 1) === 1;
  switch (step) {
    case "Up":
      return [l - 2, c];
    case "Down":
      return [l + 2, c];
    case "Left":
      return odd ? [l + 1, c] : [l - 1, c - 1];
    case "Right":
      return odd ? [l + 1, c + 1] : [l - 1, c];
  }
}
/** The shortest key sequence from one cell to another. */
function keyPath(from: Cell, to: Cell): Step[] {
  const key = (c: Cell) => `${c[0]},${c[1]}`;
  const seen = new Map<string, { prev: Cell; step: Step } | null>([
    [key(from), null],
  ]);
  const queue: Cell[] = [from];
  while (queue.length > 0) {
    const cell = queue.shift();
    if (cell === undefined) break;
    if (key(cell) === key(to)) {
      const steps: Step[] = [];
      let at = cell;
      for (let link = seen.get(key(at)); link; link = seen.get(key(at))) {
        steps.unshift(link.step);
        at = link.prev;
      }
      return steps;
    }
    for (const step of ["Up", "Down", "Left", "Right"] as const) {
      const next = neighbour(cell, step);
      if (next[0] < -2 || next[0] > 80 || next[1] < -2 || next[1] > 40)
        continue;
      if (seen.has(key(next))) continue;
      seen.set(key(next), { prev: cell, step });
      queue.push(next);
    }
  }
  throw Error(`no key path from ${key(from)} to ${key(to)}`);
}

// On the forked save, in the board's own 0-based cells (hl llmplay prints them one higher): the CP HQ's
// entrance, the medium artillery u13 and its target, the EC cavalry it destroys.
const forkHQ: Cell = [5, 0];
const forkArtillery: Cell = [16, 3];
const forkTarget: Cell = [16, 2];

// The forked mid-game save both sides load.
function forkedSave(): string {
  const dir = join(scratch, "fork9"),
    save = join(dir, "save.json");
  if (!existsSync(save)) {
    mkdirSync(scratch, { recursive: true });
    execFileSync(
      join(repo, "bin/hl"),
      [
        "llmplay",
        "fork",
        "-dir",
        "levels/1914-08/replays/cp.jsonl",
        "-at",
        "9",
        "-out",
        dir,
      ],
      { cwd: repo, stdio: "inherit" },
    );
  }
  return save;
}

// ---------------------------------------------------------------- web

const contentTypes: Record<string, string> = {
  ".css": "text/css",
  ".html": "text/html; charset=utf-8",
  ".jpg": "image/jpeg",
  ".js": "text/javascript",
  ".json": "application/json",
  ".png": "image/png",
  ".webp": "image/webp",
  ".webmanifest": "application/manifest+json",
  ".woff2": "font/woff2",
};
function serveSite(): Server {
  const root = join(repo, "build/webapp/site");
  return createServer((request, response) => {
    const path = new URL(request.url ?? "/", "http://localhost").pathname;
    if (!path.startsWith("/hl/") || path.includes("..")) {
      response.writeHead(404).end();
      return;
    }
    const file =
      path === "/hl/"
        ? join(root, "index.html")
        : join(root, decodeURIComponent(path.slice(4)));
    try {
      const body = readFileSync(file);
      const dot = file.lastIndexOf(".");
      response.writeHead(200, {
        "Content-Type":
          contentTypes[dot >= 0 ? file.slice(dot) : ""] ??
          "application/octet-stream",
      });
      response.end(body);
    } catch {
      response.writeHead(404).end();
    }
  }).listen(port, "localhost");
}

/** Settings go in before the app opens its database, in the same upgrade that makes its stores. */
function seedSettings(settings: Record<string, unknown>): string {
  return `(() => {
    const request = indexedDB.open("historyline", 1);
    request.onupgradeneeded = () => {
      const db = request.result;
      db.createObjectStore("saves", { keyPath: "slot" });
      db.createObjectStore("prefs", { keyPath: "key" });
      db.createObjectStore("scores", { keyPath: "level" });
      request.transaction.objectStore("prefs").put({ key: "settings", value: ${JSON.stringify(settings)} });
    };
    request.onsuccess = () => request.result.close();
  })();`;
}

/** Waits until two screenshots in a row are equal (the map and its art have settled), at most limit ms. */
async function settle(page: Page, limit = 20_000): Promise<number> {
  const start = Date.now();
  let last: Buffer | null = null;
  while (Date.now() - start < limit) {
    await sleep(700);
    const shot = await page.screenshot({
      type: "jpeg",
      quality: 40,
      scale: "css",
    });
    if (last !== null && shot.equals(last)) return Date.now() - start;
    last = shot;
  }
  return -1;
}
async function press(page: Page, key: string, wait = 350): Promise<void> {
  await page.keyboard.press(key);
  await sleep(wait);
}
/** Hold fire (Space) and push a direction: the gesture of the original. */
async function gesture(page: Page, key: string): Promise<void> {
  await page.keyboard.down(" ");
  await sleep(250);
  await page.keyboard.press(key);
  await sleep(250);
  await page.keyboard.up(" ");
  await sleep(700);
}

class WebRun {
  readonly dir: string;
  readonly shots: Record<string, string> = {};
  readonly errors: string[] = [];
  readonly metrics: Record<string, unknown> = {};
  readonly browser: Browser;
  readonly size: Size;
  readonly layout: string;
  constructor(browser: Browser, size: Size, layout: string) {
    this.browser = browser;
    this.size = size;
    this.layout = layout;
    this.dir = join(
      out,
      "web",
      size.name,
      flag("soft") ? `${layout}-soft` : layout,
    );
    mkdirSync(this.dir, { recursive: true });
  }
  get baseURL(): string {
    return process.env["E2E_BASE_URL"] ?? `http://localhost:${port}/hl/`;
  }
  async open(
    settings: Record<string, unknown>,
  ): Promise<{ context: BrowserContext; page: Page }> {
    const base: BrowserContextOptions = this.size.phone
      ? { ...devices["Pixel 7"] }
      : {
          viewport: {
            width: this.size.width / this.size.dpr,
            height: this.size.height / this.size.dpr,
          },
          deviceScaleFactor: this.size.dpr,
        };
    const context = await this.browser.newContext({
      ...base,
      reducedMotion: "no-preference",
    });
    await context.addInitScript(
      seedSettings({ layout: this.layout, ...settings }),
    );
    const page = await context.newPage();
    page.on("pageerror", (e) => this.errors.push(`pageerror: ${e.message}`));
    page.on("console", (m) => {
      if (m.type() === "error") this.errors.push(`console: ${m.text()}`);
    });
    return { context, page };
  }
  async shoot(page: Page, name: string, note = ""): Promise<void> {
    await page.screenshot({
      path: join(this.dir, `${name}.png`),
      scale: "device",
    });
    this.shots[name] = note;
    log(
      `web ${this.size.name} ${this.layout} ${name}${note ? ` (${note})` : ""}`,
    );
  }
  async step(name: string, work: () => Promise<void>): Promise<void> {
    try {
      await work();
    } catch (error) {
      this.errors.push(
        `${name}: ${error instanceof Error ? (error.message.split("\n")[0] ?? "") : String(error)}`,
      );
      log(`web ${this.size.name} ${this.layout} ${name} failed`);
    }
  }
  async metricsOf(page: Page, key: string): Promise<void> {
    this.metrics[key] = await page.evaluate(() => ({
      dpr: window.devicePixelRatio,
      inner: [window.innerWidth, window.innerHeight],
      canvases: [...document.querySelectorAll("canvas")].map((c) => ({
        cls: c.className,
        css: [c.clientWidth, c.clientHeight],
        backing: [c.width, c.height],
      })),
      font: getComputedStyle(document.body).fontFamily,
    }));
  }
  save(): void {
    const file = join(this.dir, "run.json");
    const old = existsSync(file)
      ? (JSON.parse(readFileSync(file, "utf8")) as {
          shots?: Record<string, string>;
          metrics?: Record<string, unknown>;
        })
      : {};
    writeFileSync(
      file,
      JSON.stringify(
        {
          size: this.size,
          layout: this.layout,
          browser: browserName,
          gpu: !flag("soft"),
          shots: { ...old.shots, ...this.shots },
          errors: this.errors,
          metrics: { ...old.metrics, ...this.metrics },
        },
        null,
        2,
      ),
    );
  }
}

const ready = (page: Page) => page.locator(".battle-ready").first();
const button = (page: Page, name: string | RegExp, exact = false) =>
  page.getByRole("button", { name, exact });

/** Reads the cursor cell from the dock's terrain heading ("Terrain line, col"), present in both layouts. */
async function cursorCell(page: Page): Promise<Cell> {
  const text = await page
    .locator(".battle-terrain h3 small")
    .first()
    .textContent();
  const m = /(\d+),\s*(\d+)/.exec(text ?? "");
  if (m === null) throw Error(`no cursor cell in "${text ?? ""}"`);
  return [Number(m[1]), Number(m[2])];
}
async function walk(page: Page, steps: Step[]): Promise<void> {
  for (const s of steps) await press(page, `Arrow${s}`, 220);
}

async function webBattleStart(
  run: WebRun,
  page: Page,
  code: string,
): Promise<void> {
  await page.goto(run.baseURL);
  await button(page, "Start").waitFor({ timeout: 180_000 });
  if (code !== "PULSE") {
    await button(page, /^Code/).click();
    await page.getByRole("textbox", { name: "Map code" }).fill(code);
    await button(page, "Confirm code").click();
    await button(page, "Start").waitFor();
  }
  await button(page, "Start").click();
  await button(page, /^Germany/).click();
  await button(page, "Continue").click();
  await button(page, "To battle").click();
  await ready(page).waitFor({ state: "attached", timeout: 300_000 });
}

async function webFlow(run: WebRun): Promise<void> {
  const { context, page } = await run.open({});
  await run.step("menu", async () => {
    await page.goto(run.baseURL);
    await button(page, "Start").waitFor({ timeout: 180_000 });
    await settle(page, 8_000);
    await run.shoot(page, "menu");
  });
  await run.step("side", async () => {
    await button(page, "Start").click();
    await button(page, /^Germany/).waitFor();
    await settle(page, 5_000);
    await run.shoot(page, "side");
  });
  await run.step("news", async () => {
    await button(page, /^Germany/).click();
    await button(page, "Continue").waitFor();
    await settle(page, 5_000);
    await run.shoot(page, "news");
  });
  await run.step("brief", async () => {
    await button(page, "Continue").click();
    await button(page, "To battle").waitFor();
    await settle(page, 5_000);
    await run.shoot(page, "brief");
  });
  await run.step("battle", async () => {
    await button(page, "To battle").click();
    await sleep(600);
    await run.shoot(page, "loading");
    const t0 = Date.now();
    await ready(page).waitFor({ state: "attached", timeout: 300_000 });
    const readyMs = Date.now() - t0;
    // What a player sees the moment the battle is ready, before the unit art has arrived.
    await run.shoot(page, "battle-at-ready", "at .battle-ready, no settling");
    await page.waitForLoadState("networkidle").catch(() => undefined);
    const settled = await settle(page, 60_000);
    run.metrics["battleReadyMs"] = readyMs;
    run.metrics["battleSettledAfterReadyMs"] =
      settled < 0 ? -1 : Date.now() - t0 - readyMs;
    await run.shoot(
      page,
      "battle",
      settled < 0 ? "still changing after 60 s" : "",
    );
    await run.metricsOf(page, "battle");
  });
  // The cursor starts on the HQ in both front ends: three key steps right and one down is the supply car
  // beside the castle on PULSE (the desktop part uses the same keys).
  await run.step("cursor", async () => {
    await walk(page, ["Right", "Right", "Right", "Down"]);
    await settle(page, 5_000);
    await run.shoot(page, "cursor");
  });
  await run.step("move", async () => {
    await press(page, " ", 800);
    await settle(page, 5_000);
    await run.shoot(page, "move");
    await press(page, "Escape");
  });
  await run.step("contents", async () => {
    await gesture(page, "ArrowLeft");
    let note = "gesture left (open) on the supply car";
    if (!(await page.locator(".battle-contents").first().isVisible())) {
      await button(page, "Contents", true).first().click();
      note = "the gesture opened nothing; the dock's Contents button";
    }
    await settle(page, 5_000);
    await run.shoot(page, "contents", note);
    await press(page, "Escape");
    await press(page, "Escape");
  });
  await run.step("unitinfo", async () => {
    await gesture(page, "ArrowDown");
    await settle(page, 5_000);
    await run.shoot(page, "unitinfo", "gesture down (info) on the supply car");
    await press(page, "Escape");
  });
  await run.step("strategic", async () => {
    await gesture(page, "ArrowRight");
    await settle(page, 5_000);
    await run.shoot(page, "strategic", "gesture right (map)");
    await press(page, "Escape");
  });
  await run.step("gameinfo", async () => {
    await walk(page, ["Up", "Left", "Left", "Left"]);
    await gesture(page, "ArrowDown");
    let note = "gesture down (info) on the HQ";
    if (!(await page.getByText("Battle information").first().isVisible())) {
      await button(page, "Info", true).first().click();
      note = "the gesture opened nothing; the dock's Info button";
    }
    await settle(page, 5_000);
    await run.shoot(page, "gameinfo", note);
    await press(page, "Escape");
  });
  await run.step("prompt", async () => {
    await press(page, "F1", 800);
    await settle(page, 5_000);
    await run.shoot(page, "prompt");
  });
  await run.step("phase", async () => {
    const confirm = button(page, /^Confirm/);
    if (await confirm.isVisible()) await confirm.click();
    else await press(page, "F1", 0);
    await sleep(900);
    await run.shoot(page, "phase", "0.9 s after confirming");
    for (let i = 1; i <= 3; i++) {
      await sleep(1500);
      await run.shoot(
        page,
        `phase-${i}`,
        `${0.9 + 1.5 * i} s after confirming`,
      );
    }
  });
  await run.step("outcome", async () => {
    await page
      .waitForFunction(
        () => {
          const b = document.querySelector<HTMLButtonElement>(".battle-ready");
          return b !== null && !b.disabled;
        },
        undefined,
        { timeout: 180_000 },
      )
      .catch(() => undefined);
    const skip = page.locator(".battle-host-close");
    if (await skip.isVisible()) await skip.click().catch(() => undefined);
    await press(page, "F2", 800);
    const confirm = button(page, /^Confirm|^Surrender|^Yes/);
    if (await confirm.first().isVisible())
      await confirm
        .first()
        .click()
        .catch(() => undefined);
    else await press(page, "F2", 800);
    const ended = page.locator(".battle-ended button");
    await ended.waitFor({ timeout: 60_000 });
    await ended.click();
    await settle(page, 8_000);
    await run.shoot(page, "outcome");
  });
  await context.close();
}

async function webAttack(run: WebRun): Promise<void> {
  const { context, page } = await run.open({});
  await run.step("attack-idle", async () => {
    await page.goto(run.baseURL);
    await button(page, "Extras").click({ timeout: 180_000 });
    await button(page, "Load game").click();
    const envelope = JSON.stringify({
      map: "1914-08:cp",
      side: 0,
      data: readFileSync(forkedSave()).toString("base64"),
    });
    await page.getByLabel("Import slot 1").setInputFiles({
      name: "fork9.json",
      mimeType: "application/json",
      buffer: Buffer.from(envelope),
    });
    await sleep(800);
    await button(page, "Load", true).first().click();
    await ready(page).waitFor({ state: "attached", timeout: 300_000 });
    await page.waitForLoadState("networkidle").catch(() => undefined);
    await settle(page, 60_000);
    await run.shoot(page, "attack-idle");
  });
  await run.step("attack", async () => {
    const start = await cursorCell(page);
    run.metrics["forkCursorStart"] = start;
    await walk(page, keyPath(start, forkArtillery));
    await press(page, " ", 800);
    await settle(page, 5_000);
    await run.shoot(
      page,
      "attack",
      `cursor from ${start.join(",")} to ${forkArtillery.join(",")}, space`,
    );
  });
  await run.step("battle-scene", async () => {
    await walk(page, keyPath(forkArtillery, forkTarget));
    await press(page, " ", 800);
    await press(page, "F1", 800);
    const confirm = button(page, /^Confirm/);
    if (await confirm.isVisible()) await confirm.click();
    for (let i = 0; i < 8; i++) {
      await sleep(1200);
      await run.shoot(page, i === 0 ? "battle-scene" : `battle-scene-${i}`);
    }
  });
  await context.close();
}

/** Scrolls the mouse wheel over the map's centre (the human panel's in the split layout). */
async function wheel(page: Page, steps: number): Promise<void> {
  const box = page.viewportSize();
  if (box === null) return;
  const split = await page.locator(".split-screen").count();
  await page.mouse.move(
    split > 0 ? box.width / 4 : box.width * 0.4,
    box.height / 2,
  );
  for (let i = 0; i < Math.abs(steps); i++) {
    await page.mouse.wheel(0, steps < 0 ? -100 : 100);
    await sleep(150);
  }
}

async function webTiles(run: WebRun): Promise<void> {
  for (const month of levels) {
    const level = tileLevels[month];
    if (level === undefined) continue;
    const { context, page } = await run.open({ painted: false });
    await run.step(`tiles-${month}`, async () => {
      await webBattleStart(run, page, level[0]);
      await page.waitForLoadState("networkidle").catch(() => undefined);
      await settle(page, 60_000);
      await run.shoot(page, `tiles-${month}-idle`, `${level[0]}, ${level[1]}`);
      await wheel(page, -7);
      await settle(page, 30_000);
      await run.shoot(page, `tiles-${month}-near`);
      await wheel(page, 14);
      await settle(page, 30_000);
      await run.shoot(page, `tiles-${month}-far`);
    });
    await context.close();
  }
}

async function captureWeb(): Promise<void> {
  const server = process.env["E2E_BASE_URL"] === undefined ? serveSite() : null;
  const gpu = !flag("soft")
    ? ["--enable-gpu", "--ignore-gpu-blocklist", "--use-angle=vulkan"]
    : [];
  const engine =
    browserName === "firefox"
      ? firefox
      : browserName === "webkit"
        ? webkit
        : chromium;
  const browser = await engine.launch({
    args: browserName === "chromium" ? gpu : [],
  });
  try {
    for (const size of sizes)
      for (const layout of layouts) {
        const run = new WebRun(browser, size, layout);
        if (parts.includes("flow")) await webFlow(run);
        if (parts.includes("attack")) await webAttack(run);
        if (parts.includes("tiles")) await webTiles(run);
        run.save();
      }
  } finally {
    await browser.close();
    server?.close();
  }
}

// ---------------------------------------------------------------- go (the desktop game under Xvfb)

const display = option("display", ":93")[0] ?? ":93";
function x(command: string, ...rest: string[]): string {
  return execFileSync(command, rest, {
    env: { ...process.env, DISPLAY: display },
    encoding: "utf8",
  });
}
// The arrows repeat after 18 frames held (cmd/historyline/cursor.go repeatDelay): keep presses short, and step the
// cursor only at the first (fastest) size, where frames come at 60 per second even under software GL.
async function goKey(key: string, wait = 600): Promise<void> {
  x("xdotool", "keydown", key);
  await sleep(150);
  x("xdotool", "keyup", key);
  await sleep(wait);
}
async function goWalk(steps: Step[]): Promise<void> {
  for (const s of steps) await goKey(s, 300);
}
async function goGesture(key: string): Promise<void> {
  x("xdotool", "keydown", "space");
  await sleep(600);
  x("xdotool", "keydown", key);
  await sleep(400);
  x("xdotool", "keyup", key);
  await sleep(500);
  x("xdotool", "keyup", "space");
  await sleep(1500);
}
function goWindow(): string {
  return x("xdotool", "search", "--name", "Historyline").split("\n")[0] ?? "";
}
function goFingerprint(): string {
  return x(
    "sh",
    "-c",
    `import -window ${goWindow()} -resize 48x27! -depth 8 rgb:- | md5sum`,
  );
}
/** Keys go to the window by explicit focus; the pointer waits outside it, off every button and hex. */
function parkPointer(): void {
  x("xdotool", "windowfocus", goWindow());
  x("xdotool", "mousemove", "3839", "2559");
}
async function goResize(size: Size): Promise<void> {
  // The game sets its own window size when it maps the window: repeat until the size holds.
  for (let i = 0; i < 10; i++) {
    const w = goWindow();
    x("xdotool", "windowmove", w, "0", "0");
    x("xdotool", "windowsize", w, String(size.width), String(size.height));
    parkPointer();
    await sleep(3000);
    const geometry = x("xdotool", "getwindowgeometry", w);
    if (geometry.includes(`${size.width}x${size.height}`)) return;
  }
  log(`go: the window would not take ${size.name}`);
}
/** Mouse wheel over the human (left) panel: button 4 zooms in, 5 out. */
async function goWheel(size: Size, steps: number): Promise<void> {
  x(
    "xdotool",
    "mousemove",
    String(Math.round(size.width / 4)),
    String(Math.round(size.height / 2)),
  );
  for (let i = 0; i < Math.abs(steps); i++) {
    x("xdotool", "click", steps < 0 ? "4" : "5");
    await sleep(250);
  }
  parkPointer();
  await sleep(2500);
}

const goShots: Record<string, Record<string, string>> = {};
function goShoot(size: Size, name: string, note = ""): void {
  const dir = join(out, "go", size.name);
  mkdirSync(dir, { recursive: true });
  x("import", "-window", goWindow(), join(dir, `${name}.png`));
  (goShots[size.name] ??= {})[name] = note;
  log(`go ${size.name} ${name}${note ? ` (${note})` : ""}`);
}
/** One screen at every size: the game lays itself out anew for each window size. */
async function goEverySize(name: string, note = ""): Promise<void> {
  for (const size of sizes) {
    await goResize(size);
    goShoot(size, name, note);
  }
  // Back to the first size: the key steps that follow see the same panels every time.
  if (sizes.length > 1) await goResize(first(sizes));
}

function startGo(extra: string[]): ChildProcess {
  const bin = join(repo, "bin/historyline");
  if (!existsSync(bin))
    execFileSync("go", ["build", "-o", "bin/", "./cmd/historyline"], {
      cwd: repo,
      stdio: "inherit",
    });
  return spawn(
    "systemd-run",
    [
      "--user",
      "--scope",
      "-p",
      "MemoryMax=8G",
      "-p",
      "MemorySwapMax=0",
      bin,
      ...extra,
    ],
    { cwd: repo, env: { ...process.env, DISPLAY: display }, stdio: "ignore" },
  );
}
async function waitWindow(): Promise<void> {
  for (let i = 0; i < 240; i++) {
    try {
      if (goWindow() !== "") return;
    } catch {
      /* not up yet */
    }
    await sleep(500);
  }
  throw Error("the desktop game's window did not open");
}
/** Starts the game straight in a battle at one window size and waits until its first frame has settled. A direct
 * start draws the remake's tiles; the painted design needs -backdrop with the level's painting. */
async function goBattle(size: Size, extra: string[]): Promise<ChildProcess> {
  const game = startGo(extra);
  await waitWindow();
  await goResize(size);
  // Loaded once the picture holds still for two looks after at least 30 s (the start is black, then the map).
  const start = Date.now();
  let last = "";
  for (;;) {
    await sleep(5000);
    const now = goFingerprint();
    if (now === last && Date.now() - start > 30_000) break;
    if (Date.now() - start > 20 * 60_000) {
      log("go: the battle did not settle in 20 minutes");
      break;
    }
    last = now;
  }
  await goResize(size);
  return game;
}

/** The menu pages at every size, up to the loading page (the battle screens get a game per size). */
async function goMenus(): Promise<void> {
  const game = startGo([]);
  try {
    await waitWindow();
    await sleep(5000);
    parkPointer();
    await goEverySize("menu");
    await goKey("Return", 2000);
    await goEverySize("side");
    await goKey("Return", 2000);
    await goEverySize("news");
    await goKey("Return", 2000);
    await goEverySize("brief");
    await goKey("Right", 300);
    await goKey("Return", 3000);
    goShoot(first(sizes), "loading");
  } finally {
    game.kill("SIGTERM");
    await sleep(3000);
  }
}

/** The battle screens of the opening: the keys at the first size, each screen at every size. */
async function goOpening(): Promise<void> {
  const game = await goBattle(first(sizes), [
    "-level",
    "1914-08:cp",
    "-backdrop",
    "levels/1914-08/backdrop.jpg",
  ]);
  try {
    await goEverySize("battle");
    await goWalk(["Right", "Right", "Right", "Down"]);
    await goEverySize("cursor");
    await goKey("space", 1500);
    await goEverySize("move");
    await goKey("Escape");
    await goGesture("Left");
    await goEverySize("contents", "gesture left (open) on the supply car");
    await goKey("Escape");
    await goKey("Escape");
    await goGesture("Down");
    await goEverySize("unitinfo", "gesture down (info) on the supply car");
    await goKey("Escape");
    await goGesture("Right");
    await goEverySize("strategic", "gesture right (map)");
    await goKey("Escape");
    await goWalk(["Up", "Left", "Left", "Left"]);
    await goGesture("Down");
    await goEverySize("gameinfo", "gesture down (info) on the HQ");
    await goKey("Escape");
    await goKey("F1", 1500);
    await goEverySize("prompt");
    const size = first(sizes);
    await goKey("F1", 900);
    goShoot(size, "phase", "0.9 s after confirming");
    for (let i = 1; i <= 3; i++) {
      await sleep(1500);
      goShoot(size, `phase-${i}`, `${0.9 + 1.5 * i} s after confirming`);
    }
    await sleep(90_000); // the computer's orders play out
    await goKey("F2", 1000);
    await goKey("F2", 3000);
    await sleep(10_000);
    await goEverySize("outcome", "after surrendering (F2 twice)");
  } finally {
    game.kill("SIGTERM");
    await sleep(3000);
  }
}

async function goAttack(): Promise<void> {
  const game = await goBattle(first(sizes), [
    "-level",
    "1914-08:cp",
    "-backdrop",
    "levels/1914-08/backdrop.jpg",
    "-snapshot",
    forkedSave(),
  ]);
  try {
    await goEverySize("attack-idle");
    await goWalk(keyPath(forkHQ, forkArtillery));
    await goKey("space", 1500);
    await goEverySize(
      "attack",
      `cursor from the HQ ${forkHQ.join(",")} to ${forkArtillery.join(",")}, space`,
    );
    const size = first(sizes);
    await goWalk(keyPath(forkArtillery, forkTarget));
    await goKey("space", 1200);
    await goKey("F1", 1200);
    await goKey("F1", 0);
    for (let i = 0; i < 8; i++) {
      await sleep(1200);
      goShoot(size, i === 0 ? "battle-scene" : `battle-scene-${i}`);
    }
  } finally {
    game.kill("SIGTERM");
    await sleep(3000);
  }
}

/** The tile design: a game per level and size, as one game through every size outgrows 6 GB at the phone size. */
async function goTiles(): Promise<void> {
  for (const month of levels) {
    if (tileLevels[month] === undefined) continue;
    for (const size of sizes) {
      const game = await goBattle(size, [
        "-backdrop",
        "off",
        "-level",
        `${month}:cp`,
      ]);
      try {
        goShoot(size, `tiles-${month}-idle`);
        await goWheel(size, -7);
        goShoot(size, `tiles-${month}-near`);
        await goWheel(size, 14);
        goShoot(size, `tiles-${month}-far`);
      } finally {
        game.kill("SIGTERM");
        await sleep(3000);
      }
    }
  }
}

async function captureGo(): Promise<void> {
  const xvfb = spawn(
    "Xvfb",
    [display, "-screen", "0", "3840x2560x24", "-nolisten", "tcp"],
    { stdio: "ignore" },
  );
  await sleep(1500);
  try {
    // The desktop part also takes "menus" or "opening" alone (flow is both).
    if (parts.includes("flow") || parts.includes("menus")) await goMenus();
    if (parts.includes("flow") || parts.includes("opening")) await goOpening();
    if (parts.includes("attack")) await goAttack();
    if (parts.includes("tiles")) await goTiles();
  } finally {
    xvfb.kill("SIGTERM");
    const file = join(out, "go", "live.json");
    const old = existsSync(file)
      ? (JSON.parse(readFileSync(file, "utf8")) as Record<
          string,
          Record<string, string>
        >)
      : {};
    for (const [k, v] of Object.entries(goShots)) old[k] = { ...old[k], ...v };
    writeFileSync(file, JSON.stringify(old, null, 2));
  }
}

// ---------------------------------------------------------------- index

function exists(path: string): boolean {
  return existsSync(join(out, path));
}
function writeIndex(): void {
  const columns: { title: string; path: (size: string) => string }[] = [
    { title: "Desktop game (Go)", path: (s) => `go/${s}` },
    {
      title: "Web, default (responsive) layout",
      path: (s) => `web/${s}/responsive`,
    },
    { title: "Web, original split layout", path: (s) => `web/${s}/original` },
  ];
  const sections: [string, string][] = [
    ...flowScreens,
    ...Object.entries(tileLevels).flatMap(([month, [code, what]]) =>
      tileViews.map(([view, zoom]): [string, string] => [
        `tiles-${month}-${view}`,
        `Tiles: ${month} ${code} (${what}), ${zoom}`,
      ]),
    ),
  ];
  const rows: string[] = [];
  for (const size of allSizes) {
    if (!columns.some((c) => exists(c.path(size.name)))) continue;
    rows.push(
      `<h2 id="${size.name}">${size.name}: ${size.width}×${size.height} device px (web at DPR ${size.dpr})</h2>`,
    );
    for (const [screen, title] of sections) {
      const cells = columns.map((c) => {
        const variants = [
          "",
          "-at-ready",
          "-1",
          "-2",
          "-3",
          "-4",
          "-5",
          "-6",
          "-7",
        ]
          .map((suffix) => `${c.path(size.name)}/${screen}${suffix}.png`)
          .filter(exists);
        return `<figure><figcaption>${c.title}</figcaption>${
          variants.length === 0
            ? `<div class="none">not captured</div>`
            : variants
                .map(
                  (v) =>
                    `<a href="${v}"><img loading="lazy" src="${v}" alt="${title}" title="${v}"></a>`,
                )
                .join("")
        }</figure>`;
      });
      if (cells.every((c) => c.includes("not captured"))) continue;
      rows.push(
        `<section id="${size.name}-${screen}"><h3>${title} <code>${screen}</code></h3><div class="pair">${cells.join("")}</div></section>`,
      );
    }
  }
  let staged = "";
  if (exists("go/captures.json")) {
    const frames = (
      JSON.parse(readFileSync(join(out, "go/captures.json"), "utf8")) as {
        frames: { file: string; view: string }[];
      }
    ).frames;
    staged = `<h2 id="staged">Desktop game, staged frames (-capture)</h2><p><code>bin/historyline -level 1914-08:cp -terrain build/packs/terrain -capture build/review/look-compare/go</code>: the split screen's states at 1920×1080 and the opening from 1024×768 to 3840×2160.</p><div class="grid">${frames
      .map(
        (f) =>
          `<figure><figcaption>${f.file}</figcaption><a href="go/${f.file}"><img loading="lazy" src="go/${f.file}" alt="${f.view}"></a></figure>`,
      )
      .join("")}</div>`;
  }
  const html = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Look comparison</title>
<style>
:root { color-scheme: dark; --bg: #15181b; --fg: #e6e1d6; --muted: #9a958a; --line: #2c3136; }
body { margin: 0; padding: 16px 24px 64px; background: var(--bg); color: var(--fg); font: 14px/1.4 system-ui, sans-serif; }
nav a { color: #d8b56a; margin-right: 12px; }
h2 { margin-top: 40px; border-bottom: 1px solid var(--line); padding-bottom: 6px; }
h3 { font-size: 14px; margin: 24px 0 8px; color: var(--muted); font-weight: 600; }
.pair { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; align-items: start; }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(360px, 1fr)); gap: 12px; }
figure { margin: 0; }
figcaption { font-size: 12px; color: var(--muted); margin-bottom: 4px; }
img { width: 100%; height: auto; display: block; border: 1px solid var(--line); margin-bottom: 4px; }
.none { padding: 24px; border: 1px dashed var(--line); color: var(--muted); text-align: center; }
@media (max-width: 900px) { .pair { grid-template-columns: 1fr; } }
</style></head><body>
<h1>Historyline: desktop game vs web edition</h1>
<p>Written by <code>webapp/app/e2e/lookcapture.ts</code>; click a picture for its full size. Gap list and work list: <a href="SPEC.md">SPEC.md</a>.</p>
<nav>${allSizes.map((s) => `<a href="#${s.name}">${s.name}</a>`).join("")}<a href="#staged">staged</a></nav>
${rows.join("\n")}
${staged}
</body></html>`;
  writeFileSync(join(out, "index.html"), html);
  log(`index: ${join(out, "index.html")}`);
}

// ---------------------------------------------------------------- main

mkdirSync(out, { recursive: true });
if (runAll || flag("go")) await captureGo();
if (runAll || flag("web")) await captureWeb();
writeIndex();
