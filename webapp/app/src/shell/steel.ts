/**
 * The menus' steel, as the desktop game paints it (internal/dospanel/menu.go, paint.go, surface.go): one palette of
 * olive-grey steel from dark to light, brown name plates with brass edges, brass rivets. The painted surfaces
 * (surface/steel, surface/rivets of the screens catalog) give only their texture: each texel's ratio to the tile's
 * mean modulates the palette's colour, so the plates keep its values.
 *
 * The page draws its plates with CSS: a colour of the ramp under the graded tile, blended with `overlay`. For a
 * colour c at most half bright, overlay gives 2·s·c, so a tile texel s = ½(1 + gain·d) yields c·(1 + gain·d), which
 * is texture.Modulate. Every steel tone then shares one graded tile.
 */

export type RGB = readonly [number, number, number];
export interface Pixels {
  width: number;
  height: number;
  /** RGBA, not premultiplied. */
  data: Uint8ClampedArray;
}

export const steelDark: RGB = [0x24, 0x24, 0x1c],
  steelLight: RGB = [0xb2, 0xb2, 0xa6],
  plateBrown: RGB = [0x42, 0x34, 0x26],
  plateEdge: RGB = [0x1c, 0x14, 0x0c],
  brassLight: RGB = [0xe0, 0xbc, 0x6c],
  brass: RGB = [0xa8, 0x80, 0x3c],
  brassDark: RGB = [0x58, 0x40, 0x1c];

/** How strongly the steel painting shows on steel, and on the brown name plates. */
export const steelGain = 0.6,
  brownGain = 0.9;

export function mix(a: RGB, b: RGB, t: number): RGB {
  const k = Math.max(0, Math.min(1, t));
  return [
    Math.round(a[0] + (b[0] - a[0]) * k),
    Math.round(a[1] + (b[1] - a[1]) * k),
    Math.round(a[2] + (b[2] - a[2]) * k),
  ];
}
/** ramp is the steel colour at t from dark (0) to light (1). */
export function ramp(t: number): RGB {
  return mix(steelDark, steelLight, t);
}
export function css(c: RGB): string {
  return `rgb(${c[0]} ${c[1]} ${c[2]})`;
}

/**
 * gradeTile is the overlay tile of a seamless painting: per texel and channel ½(1 + gain·(v/mean − 1)), the
 * deviation of texture.NewTile scaled for the overlay blend.
 */
export function gradeTile(src: Pixels, gain: number): Pixels {
  const n = src.width * src.height,
    sum = [0, 0, 0];
  for (let i = 0; i < n; i++)
    for (let ch = 0; ch < 3; ch++)
      sum[ch] = (sum[ch] ?? 0) + (src.data[4 * i + ch] ?? 0);
  const mean = sum.map((s) => s / Math.max(1, n));
  const out = new Uint8ClampedArray(4 * n);
  for (let i = 0; i < n; i++) {
    for (let ch = 0; ch < 3; ch++) {
      const m = mean[ch] ?? 0,
        d = m > 0 ? (src.data[4 * i + ch] ?? 0) / m - 1 : 0;
      out[4 * i + ch] = Math.round(127.5 * (1 + gain * d));
    }
    out[4 * i + 3] = 255;
  }
  return { width: src.width, height: src.height, data: out };
}

/** A rivet head's centre and radius in the sheet's texels. */
export interface RivetHead {
  cx: number;
  cy: number;
  r: number;
}
/** The rivet sheet's grid: rivetCols×rivetRows round heads lit from the top left, one in each cell. */
const rivetCols = 4,
  rivetRows = 2;

/**
 * rivetHeads measures the heads of a rivet sheet as the desktop game does: in each cell the centre and the radius
 * of the texels at least half covered, and the heads' mean colour.
 */
export function rivetHeads(sheet: Pixels): {
  heads: RivetHead[];
  mean: RGB;
} {
  const heads: RivetHead[] = [],
    sum = [0, 0, 0];
  let total = 0;
  for (let row = 0; row < rivetRows; row++)
    for (let col = 0; col < rivetCols; col++) {
      const x0 = Math.floor((col * sheet.width) / rivetCols),
        x1 = Math.floor(((col + 1) * sheet.width) / rivetCols),
        y0 = Math.floor((row * sheet.height) / rivetRows),
        y1 = Math.floor(((row + 1) * sheet.height) / rivetRows);
      let sx = 0,
        sy = 0,
        k = 0;
      for (let y = y0; y < y1; y++)
        for (let x = x0; x < x1; x++) {
          const i = 4 * (y * sheet.width + x);
          if ((sheet.data[i + 3] ?? 0) > 127) {
            sx += x + 0.5;
            sy += y + 0.5;
            k++;
            for (let ch = 0; ch < 3; ch++)
              sum[ch] = (sum[ch] ?? 0) + (sheet.data[i + ch] ?? 0);
          }
        }
      if (k >= 4) {
        heads.push({ cx: sx / k, cy: sy / k, r: Math.sqrt(k / Math.PI) });
        total += k;
      }
    }
  const mean = sum.map((s) => Math.max(1, s / Math.max(1, total)));
  return { heads, mean: [mean[0] ?? 0, mean[1] ?? 0, mean[2] ?? 0] };
}

function luma(c: RGB): number {
  return 0.299 * c[0] + 0.587 * c[1] + 0.114 * c[2];
}

/**
 * paintRivet sets a brass rivet of radius rad around (cx, cy) into dst, with the dark crescent it casts to the
 * bottom right (dospanel painter.rivet): the head's own texture, evenly brightened so that the heads' mean is as
 * light as brass. Without a head it shades a dome from the palette.
 */
export function paintRivet(
  dst: Pixels,
  cx: number,
  cy: number,
  rad: number,
  sheet: Pixels | null,
  head: RivetHead | null,
  mean: RGB,
): void {
  const k = luma(brass) / luma(mean),
    shade = ramp(0.1);
  const blend = (x: number, y: number, c: RGB, a: number) => {
    if (x < 0 || y < 0 || x >= dst.width || y >= dst.height || a <= 0) return;
    const i = 4 * (y * dst.width + x),
      da = (dst.data[i + 3] ?? 0) / 255,
      oa = a + da * (1 - a);
    for (let ch = 0; ch < 3; ch++)
      dst.data[i + ch] =
        oa > 0
          ? ((c[ch] ?? 0) * a + (dst.data[i + ch] ?? 0) * da * (1 - a)) / oa
          : 0;
    dst.data[i + 3] = 255 * oa;
  };
  const sample = (u: number, v: number): RGB | null => {
    if (sheet === null || head === null) return null;
    // A box filter over the texels one output pixel covers.
    const r = head.r * 0.96,
      foot = r / rad,
      n = Math.min(8, Math.max(1, Math.ceil(foot))),
      c = [0, 0, 0];
    let a = 0;
    for (let j = 0; j < n; j++)
      for (let i = 0; i < n; i++) {
        const x = Math.round(
            head.cx + u * r + foot * ((i + 0.5) / n - 0.5) - 0.5,
          ),
          y = Math.round(head.cy + v * r + foot * ((j + 0.5) / n - 0.5) - 0.5);
        const tx = Math.min(sheet.width - 1, Math.max(0, x)),
          ty = Math.min(sheet.height - 1, Math.max(0, y)),
          t = 4 * (ty * sheet.width + tx),
          w = Math.max(0, (sheet.data[t + 3] ?? 0) / 255 - 0.5) * 2;
        for (let ch = 0; ch < 3; ch++)
          c[ch] = (c[ch] ?? 0) + w * (sheet.data[t + ch] ?? 0);
        a += w;
      }
    if (a <= 0) return mean;
    return [
      Math.min(255, ((c[0] ?? 0) / a) * k),
      Math.min(255, ((c[1] ?? 0) / a) * k),
      Math.min(255, ((c[2] ?? 0) / a) * k),
    ];
  };
  for (let y = Math.floor(cy - rad - 2); y <= cy + rad * 1.5 + 2; y++)
    for (let x = Math.floor(cx - rad - 2); x <= cx + rad * 1.5 + 2; x++) {
      const dx = x + 0.5 - cx,
        dy = y + 0.5 - cy,
        d = Math.hypot(dx, dy),
        sd = Math.hypot(dx - rad * 0.4, dy - rad * 0.4);
      if (sd < rad + 0.5 && d >= rad)
        blend(x, y, shade, 0.5 * Math.min(1, rad + 0.5 - sd));
      const cov = Math.min(1, rad + 0.5 - d);
      if (cov <= 0) continue;
      const painted = sample(dx / rad, dy / rad);
      if (painted !== null) {
        blend(x, y, painted, cov);
        continue;
      }
      const nx = dx / rad,
        ny = dy / rad,
        l =
          -0.5 * nx -
          0.5 * ny +
          0.707 * Math.sqrt(Math.max(0, 1 - nx * nx - ny * ny));
      blend(
        x,
        y,
        l > 0.5
          ? mix(brass, brassLight, (l - 0.5) * 2)
          : mix(brassDark, brass, l * 2),
        cov,
      );
    }
}

/** The rivet image's geometry: a size×size square with the head of radius rivetRadius centred at rivetCentre. */
export const rivetSize = 40,
  rivetRadius = 14,
  rivetCentre = 16;

/** rivetImage is one rivet head with its shadow on a clear ground, for the plates' corners. */
export function rivetImage(sheet: Pixels | null, index = 0): Pixels {
  const out: Pixels = {
    width: rivetSize,
    height: rivetSize,
    data: new Uint8ClampedArray(4 * rivetSize * rivetSize),
  };
  const measured = sheet === null ? null : rivetHeads(sheet);
  const heads = measured?.heads ?? [],
    head = heads.length === 0 ? null : (heads[index % heads.length] ?? null);
  paintRivet(
    out,
    rivetCentre,
    rivetCentre,
    rivetRadius,
    head === null ? null : sheet,
    head,
    measured?.mean ?? brass,
  );
  return out;
}

/**
 * seamTile is the hull plating's seams (dospanel menuBackground): one tile of `step` reference pixels at `k` texels
 * each, a dark and a lit line along its top and left edge and a rivet where they cross, on a clear ground. Repeated,
 * it seams the background every step.
 */
export function seamTile(sheet: Pixels | null, step = 160, k = 4): Pixels {
  const size = step * k,
    out: Pixels = {
      width: size,
      height: size,
      data: new Uint8ClampedArray(4 * size * size),
    };
  const dark = ramp(0.1),
    lit = ramp(0.42),
    lines = Math.round(2 * k) + 2;
  const set = (x: number, y: number, c: RGB) => {
    const i = 4 * (y * size + x);
    out.data[i] = c[0];
    out.data[i + 1] = c[1];
    out.data[i + 2] = c[2];
    out.data[i + 3] = 255;
  };
  for (let i = 0; i < lines; i++) {
    const c = i > k ? lit : dark;
    for (let t = 0; t < size; t++) {
      set(i, t, c);
      set(t, i, c);
    }
  }
  const measured = sheet === null ? null : rivetHeads(sheet),
    head = measured?.heads[3] ?? null,
    rad = Math.max(1.5, 1.7 * k);
  // The crossing's rivet, in every corner, so the tile wraps.
  for (const [ox, oy] of [
    [0, 0],
    [size, 0],
    [0, size],
    [size, size],
  ] as const)
    paintRivet(
      out,
      ox + k,
      oy + k,
      rad,
      head === null ? null : sheet,
      head,
      measured?.mean ?? brass,
    );
  return out;
}
