// The three map designs F8 cycles through: the level's painted backdrop where it has one, the remake's terrain
// tiles, and the original game's tiles and units where the player imported their files.
export type Design = "painted" | "tiles" | "original";
export interface DesignSettings {
  painted: boolean;
  originalArt: boolean;
}
export interface Available {
  painted: boolean;
  original: boolean;
  /** An original map: without a painting it shows the original art by default. */
  originalMap: boolean;
}

/** design is the design a battle draws with these settings and what it has. */
export function design(s: DesignSettings, a: Available): Design {
  if (s.originalArt && a.original) return "original";
  if (s.painted && a.painted) return "painted";
  if (s.painted && a.originalMap && a.original) return "original";
  return "tiles";
}

/** nextDesign is the settings change of F8: the next available design after the current one. */
export function nextDesign(s: DesignSettings, a: Available): DesignSettings {
  const order: Design[] = ["painted", "tiles", "original"];
  const current = design(s, a);
  for (let step = 1; step <= order.length; step++) {
    const next = order[(order.indexOf(current) + step) % order.length];
    if (next === "painted" && a.painted)
      return { painted: true, originalArt: false };
    if (next === "tiles") return { painted: false, originalArt: false };
    if (next === "original" && a.original)
      return { painted: s.painted, originalArt: true };
  }
  return { painted: false, originalArt: false };
}

export const designNames: Readonly<Record<Design, string>> = {
  painted: "Painted",
  tiles: "Tiles",
  original: "Original",
};
