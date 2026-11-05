import { engineVersion } from "@historyline/engine";
import { signal } from "@preact/signals";
import { Application, Graphics } from "pixi.js";
import { render } from "preact";
import { Computer } from "./computer.ts";

// Battle controllers use this worker for plans; its lifetime follows this document.
export const computer = new Computer();

// The viewport the canvas fills, in CSS pixels, and the device pixel ratio it renders at.
const viewport = signal({
  width: window.innerWidth,
  height: window.innerHeight,
  dpr: window.devicePixelRatio,
});

function Banner() {
  const { width, height, dpr } = viewport.value;
  return (
    <div class="banner">
      <h1>Historyline 1914-1918</h1>
      <p>
        Web edition in progress · engine {engineVersion} · {width}×{height} CSS
        px at {dpr}×
      </p>
    </div>
  );
}

// A placeholder map: flat-topped hexes in the game's layout, where line l±2 is the hex above or below and
// odd lines sit half a column to the right.
function drawHexes(g: Graphics, width: number, height: number): void {
  const r = 28;
  const h = (Math.sqrt(3) / 2) * r;
  g.clear();
  for (let line = 0; line * h <= height + 2 * h; line++) {
    const shift = line % 2 === 1 ? 1.5 * r : 0;
    for (let col = 0; col * 3 * r + shift <= width + 2 * r; col++) {
      const cx = col * 3 * r + shift;
      const cy = line * h;
      const points: number[] = [];
      for (let k = 0; k < 6; k++) {
        const a = (Math.PI / 3) * k;
        points.push(cx + r * Math.cos(a), cy + r * Math.sin(a));
      }
      g.poly(points);
    }
  }
  g.stroke({ width: 1, color: 0xc4a773, alpha: 0.35 });
}

async function start(): Promise<void> {
  const stage = document.getElementById("stage");
  const ui = document.getElementById("ui");
  if (stage === null || ui === null) {
    throw new Error("missing #stage or #ui");
  }
  const app = new Application();
  await app.init({
    resizeTo: window,
    resolution: window.devicePixelRatio,
    autoDensity: true,
    antialias: true,
    background: 0x162630,
  });
  stage.appendChild(app.canvas);

  const hexes = new Graphics();
  app.stage.addChild(hexes);
  const layout = (): void => {
    viewport.value = {
      width: window.innerWidth,
      height: window.innerHeight,
      dpr: window.devicePixelRatio,
    };
    app.renderer.resolution = window.devicePixelRatio;
    drawHexes(hexes, window.innerWidth, window.innerHeight);
  };
  layout();
  window.addEventListener("resize", layout);

  render(<Banner />, ui);
}

void start();
