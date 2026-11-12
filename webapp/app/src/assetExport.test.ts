import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { gunzipSync } from "node:zlib";
import { expect, it } from "vitest";
import {
  bind,
  mustTables,
  profilesOf,
  scenario,
  play,
  level,
  hexgrid,
} from "@historyline/engine";
import { repo, listDir } from "../../engine/src/testing/files.ts";
import { AssetLoader } from "./assets.ts";
import { fileRef, record, type FileRef } from "./assetManifest.ts";
import { parseBakedLayers } from "./mapScene.ts";

const dir = join(repo, "build/webapp/data");
const fetcher: typeof fetch = async (url) => {
  const path = new URL(
    url instanceof Request ? url.url : url.toString(),
  ).pathname.replace("/data/", "");
  const bytes = await readFile(join(dir, path));
  return new Response(bytes);
};

it("exports all 156 setups as sessions identical to Go's initial saves", async () => {
  const loader = await AssetLoader.open(
    "https://example.test/data/index.json",
    { fetch: fetcher, cache: null },
  );
  const months = listDir("levels")
    .filter((m) => /^19\d\d-\d\d$/.test(m))
    .sort();
  expect(Object.keys(loader.index.levels).sort()).toEqual(months);
  const profiles = profilesOf(
    mustTables(await loader.rule("full.json")),
    mustTables(await loader.rule("demo.json")),
    await loader.rule("extension.json"),
  );
  const data = bind.bindingData(
    await loader.rule("units.json"),
    await loader.rule("records.json"),
  );
  const terrain = new Set(
    Object.keys(record(await loader.rule("terrain.json"))),
  );
  for (const month of months) {
    const files = await loader.levelFiles(month);
    const file = scenario.decode(
      new Map(
        ["map.json", "cp.json", "ec.json", "2p.json"].map((name) => [
          name,
          files.get(name),
        ]),
      ),
      terrain,
    );
    file.month = month; // as the campaign loads it: the extension follows the month
    for (const mode of ["cp", "ec", "2p"] as const) {
      const lv = scenario.level(file, mode);
      const session = play.newSession(lv, data, {
        profile: profiles.full,
        ...(profiles.extended === undefined
          ? {}
          : { extended: profiles.extended }),
        computer: [true, true],
      });
      const golden = gunzipSync(
        await readFile(
          join(repo, `build/webapp/golden/${month}-${mode}.jsonl.gz`),
        ),
      )
        .toString("utf8")
        .split("\n", 1)[0];
      if (golden === undefined) throw new Error("empty Go trace");
      const expected = record(JSON.parse(golden) as unknown);
      expect(
        new TextDecoder().decode(play.save(session)),
        `${month}:${mode}`,
      ).toBe(expected.save);
      const baked = parseBakedLayers(await loader.baked(month, mode));
      expect(baked.scene.grid).toEqual(level.grid(lv));
      for (const [key, densities] of baked.backgrounds) {
        const c = play.coordFromKey(key),
          cell = baked.scene.cells[c.line]?.[c.col];
        expect(hexgrid.contains(baked.scene.grid, c), key).toBe(true);
        if (cell === undefined) throw new Error(`missing baked cell ${key}`);
        const cluster = baked.scene.clusters[cell.cluster - 1];
        const background =
          cluster !== undefined && cluster.stamp !== ""
            ? cell.under
            : cell.background;
        expect([...densities.keys()].sort()).toEqual([1, 2, 4]);
        for (const [density, image] of densities) {
          expect(image.id, `${month}:${mode}:${key}`).toBe(background);
          expect(image.density).toBe(density);
        }
      }
    }
  }
}, 120_000);

it("pins every exported catalogue, tile, baked raster and font to existing bytes", async () => {
  const files: FileRef[] = [];
  const checked = new Set<string>();
  const collect = (root: unknown) => {
    const queue: unknown[] = [root];
    while (queue.length > 0) {
      const value = queue.pop();
      if (value === null || typeof value !== "object") continue;
      if (Array.isArray(value)) {
        queue.push(...(value as unknown[]));
        continue;
      }
      const obj = record(value);
      if (
        typeof obj.path === "string" &&
        typeof obj.sha256 === "string" &&
        typeof obj.bytes === "number"
      ) {
        const ref = fileRef(obj);
        if (!checked.has(ref.path)) {
          checked.add(ref.path);
          files.push(ref);
        }
      } else queue.push(...Object.values(obj));
    }
  };
  collect(
    JSON.parse(await readFile(join(dir, "index.json"), "utf8")) as unknown,
  );
  for (let start = 0; start < files.length;) {
    // Bounded parallel reads keep the complete hash audit practical on NTFS.
    const batch = files.slice(start, start + 32);
    start += batch.length;
    await Promise.all(
      batch.map(async (ref) => {
        const bytes = await readFile(join(dir, ref.path));
        expect(bytes.length, ref.path).toBe(ref.bytes);
        expect(createHash("sha256").update(bytes).digest("hex"), ref.path).toBe(
          ref.sha256,
        );
        if (ref.path.endsWith(".json"))
          collect(JSON.parse(bytes.toString("utf8")) as unknown);
      }),
    );
  }
  expect(checked.size).toBeGreaterThan(1000);
}, 120_000);
