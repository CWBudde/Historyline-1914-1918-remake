import { expect, it } from "vitest";
import { parseBakedLayers, terrainDraws } from "./mapScene.ts";

function fixture() {
  const cell = (Background: string, extra: Record<string, unknown> = {}) => ({
    Background,
    Overlays: null,
    Standing: null,
    Shadows: null,
    Scatter: null,
    Ground: null,
    Context: "",
    Cluster: 0,
    Under: "",
    ...extra,
  });
  return {
    schema: 1,
    scene: {
      Grid: { Width: 2, Lines: 2 },
      Cells: [
        [
          cell("tile/grass", {
            Standing: ["standing/tree"],
            Shadows: ["standing/shadow/tree"],
            Scatter: [{ ID: "standing/scatter/tree", X: 3, Y: 4 }],
          }),
          cell("tile/stone", {
            Cluster: 1,
            Under: "tile/grass",
            Overlays: ["tile/road"],
            Standing: ["standing/stone", "standing/road"],
            Shadows: ["standing/shadow/stone", "standing/shadow/road"],
          }),
        ],
        [
          cell("tile/stone", {
            Cluster: 1,
            Under: "tile/grass",
            Standing: ["standing/stone"],
          }),
        ],
      ],
      Clusters: [
        {
          Kind: "boulders",
          Cells: [
            { Line: 0, Col: 1 },
            { Line: 1, Col: 0 },
          ],
          Shape: "",
          Stamp: "stamp/boulders/shape/1",
          Standing: "standing/stamp/boulders/shape/1",
        },
      ],
      Gaps: null,
      Issues: null,
      Fallbacks: null,
    },
    seasons: {},
    water: {},
    blend: {},
  };
}

it("retains offscreen cluster anchors and draws global terrain passes", () => {
  const baked = parseBakedLayers(fixture());
  const draws = terrainDraws(baked.scene, [
    { line: 0, col: 0 },
    { line: 1, col: 0 },
  ]);
  const ids = draws.map((d) => d.id);
  expect(ids).toContain("stamp/boulders/shape/1");
  expect(ids).not.toContain("standing/stone");
  expect(ids).toContain("standing/road");
  const shadows = draws.filter((d) => d.pass === "shadow");
  const standing = draws.filter((d) => d.pass === "standing");
  expect(shadows.map((d) => d.id)).toContain("standing/shadow/road");
  expect(shadows.map((d) => d.id)).not.toContain("standing/shadow/stone");
  expect(Math.max(...shadows.map((d) => draws.indexOf(d)))).toBeLessThan(
    Math.min(...standing.map((d) => draws.indexOf(d))),
  );
  const scattered = standing.find((d) => d.id === "standing/scatter/tree");
  expect(scattered?.at).toEqual({ x: 51, y: 44 });
  expect(draws.filter((d) => d.id === "stamp/boulders/shape/1")).toHaveLength(
    1,
  );
});

it("rejects mismatched rows and keeps parsed scene data independent", () => {
  const raw = fixture();
  const baked = parseBakedLayers(raw);
  raw.scene.Cells[0]?.pop();
  expect(baked.scene.cells[0]).toHaveLength(2);
  expect(() => parseBakedLayers(raw)).toThrow(/row/);
});

it("preserves native water background pairs and rejects inconsistent geometry", () => {
  const sha = "a".repeat(64),
    file = { path: `blobs/${sha}.png`, sha256: sha, bytes: 16 },
    image = {
      id: "tile/grass",
      width: 256,
      height: 224,
      density: 2,
      logicalAnchor: [64, 56],
      lite: file,
      full: file,
    };
  const raw = { ...fixture(), backgrounds: { "0,0": { "2": image } } };
  const baked = parseBakedLayers(raw);
  expect(baked.backgrounds.get("0,0")?.get(2)).toEqual(image);
  expect(parseBakedLayers(fixture()).backgrounds.size).toBe(0);
  image.density = 4;
  expect(() => parseBakedLayers(raw)).toThrow(/density/);
});

it("uses baked season layers including empty passes and selected scatter, keeping main shadows", () => {
  const sha = "b".repeat(64),
    file = { path: `blobs/${sha}.png`, sha256: sha, bytes: 32 };
  const image = {
    id: "mixed/background/0,0",
    width: 128,
    height: 112,
    density: 1,
    logicalAnchor: [64, 56],
    lite: file,
    full: file,
  };
  const raw = {
    ...fixture(),
    seasonBakes: {
      layers: {
        "0,0": { background: { "1": image }, infrastructure: {}, standing: {} },
      },
      ground: { "0,0": [{ ID: "season-ground", X: 1, Y: 2 }] },
      standing: { "0,0": [{ ID: "season-crown", X: 3, Y: 4 }] },
    },
  };
  const baked = parseBakedLayers(raw),
    draws = terrainDraws(baked.scene, [{ line: 0, col: 0 }], baked.seasonBakes);
  expect(draws.filter((d) => d.pass === "background").map((d) => d.id)).toEqual(
    [image.id],
  );
  expect(draws.filter((d) => d.pass === "standing").map((d) => d.id)).toEqual([
    "season-crown",
  ]);
  expect(draws.filter((d) => d.pass === "stamp").map((d) => d.id)).toEqual([
    "season-ground",
  ]);
  expect(draws.filter((d) => d.pass === "shadow").map((d) => d.id)).toContain(
    "standing/shadow/tree",
  );
  expect(
    draws.filter((d) => d.pass === "shadow").map((d) => d.id),
  ).not.toContain("standing/shadow/season-crown");
});
