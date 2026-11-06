/** Native unitsprite dressing on straight RGBA pixels, independent of Canvas. */
export type Rim = "dark" | "light";
export interface RGBARaster {
  width: number;
  height: number;
  data: Uint8ClampedArray<ArrayBuffer>;
}

export function rimFor(terrain: string, infra: boolean): Rim {
  if (
    !infra &&
    [
      "forest",
      "forest_w",
      "water",
      "waterDeep",
      "waterSea",
      "waterStones",
    ].includes(terrain)
  )
    return "light";
  return "dark";
}

/** Texel border; add margin / density to the original logical anchor. */
export function dressMargin(density: number): number {
  return density * Math.ceil(6 / Math.SQRT2 + 3 * 0.75 + 1);
}

interface Layer {
  r: number;
  g: number;
  b: number;
  a: number;
}
function over(l: Layer, r: number, g: number, b: number, alpha: number): Layer {
  const a = alpha + l.a * (1 - alpha);
  if (a === 0) return l;
  return {
    r: (r * alpha + l.r * l.a * (1 - alpha)) / a,
    g: (g * alpha + l.g * l.a * (1 - alpha)) / a,
    b: (b * alpha + l.b * l.a * (1 - alpha)) / a,
    a,
  };
}

function dilate(
  alpha: Float64Array,
  width: number,
  height: number,
  radius: number,
): Float64Array {
  const rows = [alpha];
  let previous = alpha;
  for (let k = 1; k <= radius; k++) {
    const next = new Float64Array(alpha.length);
    for (let y = 0; y < height; y++)
      for (let x = 0; x < width; x++) {
        const i = y * width + x;
        let value = previous[i] ?? 0;
        if (x >= k) value = Math.max(value, alpha[i - k] ?? 0);
        if (x + k < width) value = Math.max(value, alpha[i + k] ?? 0);
        next[i] = value;
      }
    rows.push(next);
    previous = next;
  }
  const out = new Float64Array(alpha.length);
  for (let dy = -radius; dy <= radius; dy++) {
    const row = rows[Math.floor(Math.sqrt(radius * radius - dy * dy))];
    if (row === undefined) continue;
    for (let y = Math.max(0, -dy); y < Math.min(height, height - dy); y++)
      for (let x = 0; x < width; x++) {
        const i = y * width + x;
        out[i] = Math.max(out[i] ?? 0, row[(y + dy) * width + x] ?? 0);
      }
  }
  return out;
}

function blur(
  alpha: Float64Array,
  width: number,
  height: number,
  sigma: number,
): Float64Array {
  const radius = Math.ceil(3 * sigma),
    kernel = new Float64Array(2 * radius + 1);
  let sum = 0;
  for (let i = 0; i < kernel.length; i++) {
    const v = Math.exp(-((i - radius) * (i - radius)) / (2 * sigma * sigma));
    kernel[i] = v;
    sum += v;
  }
  for (let i = 0; i < kernel.length; i++) kernel[i] = (kernel[i] ?? 0) / sum;
  const pass = (input: Float64Array, dx: number, dy: number): Float64Array => {
    const out = new Float64Array(input.length);
    for (let y = 0; y < height; y++)
      for (let x = 0; x < width; x++) {
        let value = 0;
        for (let k = 0; k < kernel.length; k++) {
          const nx = x + (k - radius) * dx,
            ny = y + (k - radius) * dy;
          if (nx >= 0 && ny >= 0 && nx < width && ny < height)
            value += (kernel[k] ?? 0) * (input[ny * width + nx] ?? 0);
        }
        out[y * width + x] = value;
      }
    return out;
  };
  return pass(pass(alpha, 1, 0), 0, 1);
}

/** Same disc rim, Gaussian contact shadow and straight-alpha order as Go Dress. */
export function dress(
  src: RGBARaster,
  density: number,
  flying: boolean,
  rim: Rim,
): RGBARaster {
  const margin = dressMargin(density),
    width = src.width + 2 * margin,
    height = src.height + 2 * margin;
  const sprite = new Uint8ClampedArray(width * height * 4);
  let minX = width,
    minY = height,
    maxX = 0,
    maxY = 0;
  for (let y = 0; y < src.height; y++) {
    sprite.set(
      src.data.subarray(y * src.width * 4, (y + 1) * src.width * 4),
      ((y + margin) * width + margin) * 4,
    );
    for (let x = 0; x < src.width; x++) {
      if ((src.data[(y * src.width + x) * 4 + 3] ?? 0) !== 0) {
        minX = Math.min(minX, x + margin);
        minY = Math.min(minY, y + margin);
        maxX = Math.max(maxX, x + margin + 1);
        maxY = Math.max(maxY, y + margin + 1);
      }
    }
  }
  if (minX >= maxX || minY >= maxY) return { width, height, data: sprite };
  const shift = Math.round(((flying ? 6 : 2) / Math.SQRT2) * density);
  const sigma = 0.75 * density,
    radius = Math.round(1.5 * density),
    reach = radius + Math.ceil(3 * sigma);
  const left = Math.max(0, minX - reach),
    top = Math.max(0, minY - reach);
  const right = Math.min(width, maxX + reach + shift),
    bottom = Math.min(height, maxY + reach + shift);
  const w = right - left,
    h = bottom - top,
    alpha = new Float64Array(w * h);
  for (let y = 0; y < h; y++)
    for (let x = 0; x < w; x++)
      alpha[y * w + x] =
        (sprite[((top + y) * width + left + x) * 4 + 3] ?? 0) / 255;
  const outline = dilate(alpha, w, h, radius),
    shadow = blur(alpha, w, h, sigma);
  const out = new Uint8ClampedArray(sprite.length);
  for (let y = 0; y < h; y++)
    for (let x = 0; x < w; x++) {
      const i = y * w + x,
        p = ((top + y) * width + left + x) * 4;
      const sa =
        x >= shift && y >= shift
          ? 0.45 * (shadow[(y - shift) * w + x - shift] ?? 0)
          : 0;
      let c = over({ r: 0, g: 0, b: 0, a: 0 }, 18, 26, 16, sa);
      c =
        rim === "light"
          ? over(c, 196, 190, 162, outline[i] ?? 0)
          : over(c, 16, 18, 12, outline[i] ?? 0);
      c = over(
        c,
        sprite[p] ?? 0,
        sprite[p + 1] ?? 0,
        sprite[p + 2] ?? 0,
        alpha[i] ?? 0,
      );
      out[p] = Math.round(c.r);
      out[p + 1] = Math.round(c.g);
      out[p + 2] = Math.round(c.b);
      out[p + 3] = Math.round(c.a * 255);
    }
  return { width, height, data: out };
}

/** Native terrain edge padding, returned as straight RGBA for ImageData.
 * Includes Go's 8-bit premultiplication round trip; grown opaque pixels keep
 * the native right/left/down/up/diagonal neighbour preference.
 */
export function padded(
  src: RGBARaster,
  pad: number,
  bleed: boolean,
): RGBARaster {
  const width = src.width + 2 * pad,
    height = src.height + 2 * pad,
    data = new Uint8ClampedArray(width * height * 4);
  for (let y = 0; y < src.height; y++)
    for (let x = 0; x < src.width; x++) {
      const from = (y * src.width + x) * 4,
        to = ((y + pad) * width + x + pad) * 4,
        a = src.data[from + 3] ?? 0;
      for (let c = 0; c < 3; c++)
        data[to + c] = Math.floor(
          ((src.data[from + c] ?? 0) * a * 257) / (255 * 256),
        );
      data[to + 3] = a;
    }
  if (bleed) {
    const offsets = [
      [1, 0],
      [-1, 0],
      [0, 1],
      [0, -1],
      [1, 1],
      [-1, -1],
      [1, -1],
      [-1, 1],
    ] as const;
    for (let step = 0; step < pad; step++) {
      const old = data.slice();
      for (let y = 0; y < height; y++)
        for (let x = 0; x < width; x++) {
          const i = (y * width + x) * 4;
          if (old[i + 3] === 255) continue;
          for (const [dx, dy] of offsets) {
            const nx = x + dx,
              ny = y + dy,
              j = (ny * width + nx) * 4;
            if (
              nx >= 0 &&
              ny >= 0 &&
              nx < width &&
              ny < height &&
              old[j + 3] === 255
            ) {
              data.set(old.subarray(j, j + 4), i);
              break;
            }
          }
        }
    }
  }
  for (let i = 0; i < data.length; i += 4) {
    const a = data[i + 3] ?? 0;
    if (a === 0 || a === 255) continue;
    for (let c = 0; c < 3; c++)
      data[i + c] = Math.floor(((data[i + c] ?? 0) * 65535) / (a * 256));
  }
  return { width, height, data };
}
