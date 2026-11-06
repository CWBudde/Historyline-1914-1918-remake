import { expect, it } from "vitest";
import { hexgrid, level } from "@historyline/engine";
import { reworkScene } from "./dynamicScene.ts";
import { terrainDraws, type MapScene, type SceneCell } from "./mapScene.ts";

const centre = { line: 2, col: 1 };
const adjacent = hexgrid.around(centre);
function hex(terrain = "grass", ids: string[] = []): level.Hex {
  return {
    terrain,
    infra: { ids, list: ids.length > 1 },
    unit: "",
    content: null,
  };
}
function cell(background = "tile/grass", overlays: string[] = []): SceneCell {
  return {
    background,
    overlays,
    standing: [],
    shadows: [],
    scatter: [],
    ground: [],
    cluster: 0,
    under: "",
  };
}
function fixture() {
  const original: level.Level = {
    mapName: "dynamic",
    mapNo: 0,
    player: "Central Powers",
    extra: null,
    lines: Array.from({ length: 5 }, (_, line) =>
      Array.from({ length: 3 - (line % 2) }, () => hex()),
    ),
  };
  const base: MapScene = {
    grid: { width: 3, lines: 5 },
    clusters: [],
    cells: original.lines.map((line) => line.map(() => cell())),
  };
  return { original, base };
}
function put(lv: level.Level, c: hexgrid.Coord, h: level.Hex) {
  const row = lv.lines[c.line];
  if (row === undefined) throw new Error("bad fixture coordinate");
  row[c.col] = h;
}
function at(s: MapScene, c = centre) {
  const result = s.cells[c.line]?.[c.col];
  if (result === undefined) throw new Error("missing scene hex");
  return result;
}

it.each([
  ["lone position", "trench0", [], "trench0"],
  [
    "position",
    "trench0",
    [
      [0, "trench14"],
      [3, "trench25"],
    ],
    "trench0-1",
  ],
  ["lone run", "trench14", [], "trench-s14"],
  ["end of line", "trench14", [[0, "trench14"]], "trench1-s4"],
  [
    "line",
    "trench14",
    [
      [0, "trench14"],
      [3, "trench124"],
    ],
    "trench14",
  ],
  [
    "branch",
    "trench25",
    [
      [0, "trench24"],
      [1, "trench25"],
      [4, "trench0"],
    ],
    "trench125",
  ],
  ["parallel", "trench14", [[1, "trench14"]], "trench-s14"],
] as const)("joins native trenches: %s", (_, piece, neighbours, want) => {
  const { base, original } = fixture();
  const edits = new Map([[level.coordKey(centre), hex("grass", [piece])]]);
  for (const [port, piece] of neighbours) {
    const c = adjacent[port];
    if (c === undefined) throw new Error("missing neighbour");
    edits.set(level.coordKey(c), hex("grass", [piece]));
  }
  const out = reworkScene(base, original, edits, () => true);
  expect(at(out.scene).overlays).toEqual([`tile/${want}`]);
});

it("refreshes a neighbouring trench when digging or filling changes its shared edge", () => {
  const { base, original } = fixture();
  const n = adjacent[0];
  if (n === undefined) throw new Error("missing neighbour");
  put(original, n, hex("grass", ["road14", "trench14"]));
  at(base, n).overlays = ["tile/road14", "tile/trench-s14"];
  const dug = reworkScene(
    base,
    original,
    new Map([[level.coordKey(centre), hex("grass", ["trench14"])]]),
    () => true,
  );
  expect(at(dug.scene, n).overlays).toEqual(["tile/road14", "tile/trench4-s1"]);
  expect(dug.changed.has(level.coordKey(n))).toBe(true);
  const filled = reworkScene(
    dug.scene,
    dug.level,
    new Map([[level.coordKey(centre), hex()]]),
    () => true,
  );
  expect(at(filled.scene, n).overlays).toEqual([
    "tile/road14",
    "tile/trench-s14",
  ]);
  expect(filled.changed.has(level.coordKey(n))).toBe(true);
});

it("keeps boundary ports and authored variants, with native missing-course fallbacks", () => {
  const { base, original } = fixture();
  const boundary = reworkScene(
    base,
    original,
    new Map([["0,0", hex("grass", ["trench6a"])]]),
    () => true,
  );
  expect(at(boundary.scene, { line: 0, col: 0 }).overlays).toEqual([
    "tile/trench6a",
  ]);
  const n = adjacent[1];
  if (n === undefined) throw new Error("missing neighbour");
  const edits = new Map([
    [level.coordKey(centre), hex("grass", ["trench2"])],
    [level.coordKey(n), hex("grass", ["trench5"])],
  ]);
  const fallback = reworkScene(
    base,
    original,
    edits,
    (id) => id === "tile/trench0-2" || id === "tile/trench5",
  );
  expect(at(fallback.scene).overlays).toEqual(["tile/trench0-2"]);
  const missing = reworkScene(base, original, edits, () => false);
  expect(at(missing.scene).overlays).toEqual(["tile/trench2"]);
});

it.each([
  ["summer", "grass", "", "tile/grass-detail-11"],
  ["autumn", "grass", "_au", "tile/grass-detail-11_au"],
  ["spring", "grass", "_sp", "tile/grass-detail-11_sp"],
  ["winter", "grass_w", "_w", "tile/grass_w"],
] as const)(
  "uses native seasonal forms and optional grass detail in %s",
  (season, terrain, suffix, background) => {
    const { base, original } = fixture();
    original.extra = {
      credits: new Map(),
      builds: [],
      backdrop: "",
      season,
      seasons: new Map(),
    };
    const edits = new Map([
      [level.coordKey(centre), hex(terrain, ["road14", "trench14"])],
    ]);
    const out = reworkScene(base, original, edits, () => true);
    expect(at(out.scene).background).toBe(background);
    expect(at(out.scene).overlays).toEqual([
      `tile/road14${suffix}`,
      `tile/trench-s14${suffix}`,
    ]);
    const fallback = reworkScene(
      base,
      original,
      edits,
      (id) => !/_(au|sp|w)$/.test(id),
    );
    expect(at(fallback.scene).overlays).toEqual([
      "tile/road14",
      "tile/trench-s14",
    ]);
  },
);

it("draws all four half-depot sites only when available and restores completed depot art", () => {
  const { base, original } = fixture();
  const pieces = {
    N: centre,
    W: adjacent[3],
    E: adjacent[1],
    S: { line: 4, col: 1 },
  };
  const edits = new Map<string, level.Hex>();
  for (const [piece, c] of Object.entries(pieces)) {
    if (c === undefined) throw new Error("missing depot piece");
    edits.set(level.coordKey(c), hex(`cmp${piece}`));
  }
  const half = reworkScene(base, original, edits, () => true);
  for (const [piece, c] of Object.entries(pieces)) {
    if (c === undefined) throw new Error("missing depot piece");
    expect(at(half.scene, c).background).toBe(`tile/site${piece}`);
  }
  const missing = reworkScene(
    base,
    original,
    edits,
    (id) => id !== "tile/siteW",
  );
  expect(at(missing.scene).background).toBe("tile/cmpN");
  const h = hex("cmpN");
  h.content = { owner: 1, units: null };
  edits.set(level.coordKey(centre), h);
  const full = reworkScene(base, original, edits, () => true);
  expect(at(full.scene).background).toBe("tile/cmpN");
  const resumed = reworkScene(
    half.scene,
    half.level,
    new Map([[level.coordKey(centre), h]]),
    () => true,
  );
  for (const [piece, c] of Object.entries(pieces)) {
    if (c === undefined) throw new Error("missing depot piece");
    expect(at(resumed.scene, c).background).toBe(`tile/cmp${piece}`);
    expect(resumed.changed.has(level.coordKey(c))).toBe(true);
  }
});

it.each([
  ["autumn", "", "_au"],
  ["spring", "", "_sp"],
  ["winter", "_w", "_w"],
] as const)(
  "uses all four site pieces in %s",
  (season, terrainSuffix, visualSuffix) => {
    const { base, original } = fixture();
    original.extra = {
      credits: new Map(),
      builds: [],
      backdrop: "",
      season,
      seasons: new Map(),
    };
    const pieces = new Map([
      [level.coordKey(centre), hex("cmpN" + terrainSuffix)],
      ["3,0", hex("cmpW" + terrainSuffix)],
      ["3,1", hex("cmpE" + terrainSuffix)],
      ["4,1", hex("cmpS" + terrainSuffix)],
    ]);
    const out = reworkScene(base, original, pieces, () => true);
    expect(at(out.scene).background).toBe("tile/siteN" + visualSuffix);
    expect(at(out.scene, { line: 3, col: 0 }).background).toBe(
      "tile/siteW" + visualSuffix,
    );
    expect(at(out.scene, { line: 3, col: 1 }).background).toBe(
      "tile/siteE" + visualSuffix,
    );
    expect(at(out.scene, { line: 4, col: 1 }).background).toBe(
      "tile/siteS" + visualSuffix,
    );
  },
);

it("keeps rules-season terrain while drawing a per-hex summer override", () => {
  const { base, original } = fixture();
  original.extra = {
    credits: new Map(),
    builds: [],
    backdrop: "",
    season: "winter",
    seasons: new Map([[level.coordKey(centre), "summer"]]),
  };
  const out = reworkScene(
    base,
    original,
    new Map([[level.coordKey(centre), hex("grass_w", ["trench14"])]]),
    () => true,
  );
  expect(at(out.scene).background).toBe("tile/grass-detail-11");
  expect(at(out.scene).overlays).toEqual(["tile/trench-s14"]);
  expect(out.level.lines[centre.line]?.[centre.col]?.terrain).toBe("grass_w");
});

it("invalidates a changed cluster without changing other art or either input", () => {
  const { base, original } = fixture();
  const other = adjacent[0];
  if (other === undefined) throw new Error("missing cluster member");
  base.clusters = [
    {
      kind: "boulders",
      cells: [other, centre],
      stamp: "stamp/boulders/pair/1",
      standing: "standing/stamp/boulders/pair/1",
    },
  ];
  for (const c of [other, centre]) {
    put(original, c, hex("stone"));
    Object.assign(at(base, c), {
      background: "context/stone/source",
      under: "tile/grass",
      cluster: 1,
    });
  }
  const before = structuredClone({ base, original });
  const edit = hex("grass", ["trench0"]);
  const out = reworkScene(
    base,
    original,
    new Map([[level.coordKey(centre), edit]]),
    () => true,
  );
  expect(out.scene.clusters[0]?.stamp).toBe("");
  expect(out.changed.has(level.coordKey(other))).toBe(true);
  expect(
    terrainDraws(out.scene, [other]).some(
      (draw) => draw.id === "context/stone/source",
    ),
  ).toBe(true);
  expect(at(out.scene, { line: 0, col: 0 })).toEqual(
    at(base, { line: 0, col: 0 }),
  );
  out.level.lines[centre.line]?.[centre.col]?.infra.ids.push("other");
  at(out.scene, other).overlays.push("other");
  expect({ base, original }).toEqual(before);
  expect(edit.infra.ids).toEqual(["trench0"]);
});

it("keeps visual art unchanged when only the legacy infrastructure list form differs", () => {
  const { base, original } = fixture();
  const h = hex();
  h.infra.list = true;
  const out = reworkScene(
    base,
    original,
    new Map([[level.coordKey(centre), h]]),
    () => true,
  );
  expect(out.changed.size).toBe(0);
  expect(out.scene).toEqual(base);
});
