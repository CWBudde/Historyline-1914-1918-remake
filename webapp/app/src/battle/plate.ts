/**
 * The battle scene's plate, painted once per size as internal/dospanel paintBattle paints it: brushed steel over
 * the whole canvas, a raised plate per side with the photo's recess, the blank photo, the riveted name plate and
 * the hex of dark ground in a recessed steel rim, and brass rivets along the seam between the halves.
 */
import type { Rect } from "../camera.ts";
import {
  Painter,
  hexOf,
  toneBack,
  toneBlank,
  tonePlate,
  toneWell,
} from "../split/painter.ts";
import type { BattleLayout } from "./layout.ts";

/** around calls fill for the parts of outer outside inner (inner lies within outer): above, below, left, right. */
function around(outer: Rect, inner: Rect, fill: (r: Rect) => void): void {
  const top = inner.y - outer.y,
    bottom = outer.y + outer.height - (inner.y + inner.height);
  fill({ x: outer.x, y: outer.y, width: outer.width, height: top });
  fill({
    x: outer.x,
    y: inner.y + inner.height,
    width: outer.width,
    height: bottom,
  });
  fill({
    x: outer.x,
    y: inner.y,
    width: inner.x - outer.x,
    height: inner.height,
  });
  fill({
    x: inner.x + inner.width,
    y: inner.y,
    width: outer.x + outer.width - (inner.x + inner.width),
    height: inner.height,
  });
}

/**
 * paintBattle paints layout l (in the painter's device pixels) onto p. Each pixel's steel is painted once: what a
 * later part covers is left out of the part below it, as the plate is painted where the scene opens.
 */
export function paintBattle(p: Painter, l: BattleLayout): void {
  const [top, bottom] = l.sides,
    block: Rect = {
      x: Math.min(top.half.x, bottom.half.x),
      y: top.half.y,
      width:
        Math.max(
          top.half.x + top.half.width,
          bottom.half.x + bottom.half.width,
        ) - Math.min(top.half.x, bottom.half.x),
      height: bottom.half.y + bottom.half.height - top.half.y,
    };
  // The panel behind the plates: round the block and in the seam between the halves.
  around({ x: 0, y: 0, width: p.w, height: p.h }, block, (r) => {
    p.steel(r, toneBack);
  });
  const seam = top.half.y + top.half.height;
  p.steel(
    { x: block.x, y: seam, width: block.width, height: bottom.half.y - seam },
    toneBack,
  );
  for (const s of l.sides) {
    // The raised plate round the photo's recess, the recess round the photo, the photo's blank plate.
    around(s.half, s.photoWell, (r) => {
      p.steel(r, tonePlate);
    });
    p.bevel(s.half, tonePlate, true);
    around(s.photoWell, s.photo, (r) => {
      p.steel(r, toneWell);
    });
    p.bevel(s.photoWell, toneBack, false);
    p.plate(s.photo, toneBlank);
    p.namePlate(s.name);
    p.field(...hexOf(s.hex));
  }
  const rad = Math.max(1.2, 1.3 * l.scale);
  for (const r of l.rivets) p.rivet(r.x, r.y, rad);
}
