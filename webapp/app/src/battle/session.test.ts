import { expect, it } from "vitest";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { play, scenario } from "@historyline/engine";
import {
  extendedFromURL,
  loadBattleData,
  newBattleSession,
} from "./session.ts";

const files: Record<string, string> = {
  "units.json": "internal/data/units.json",
  "records.json": "internal/dosbind/records.json",
  "full.json": "internal/dosrules/full.json",
  "demo.json": "internal/dosrules/demo.json",
  "extension.json": "internal/dosrules/extension.json",
};
const rule = async (name: string): Promise<unknown> => {
  if (name === "sprites.json")
    return {
      CP_infantry: { family: "infantry" },
      EC_infantry: { family: "infantry" },
    };
  const path = files[name];
  if (path === undefined) throw Error(`unexpected rule ${name}`);
  return JSON.parse(
    await readFile(
      fileURLToPath(new URL(`../../../../${path}`, import.meta.url)),
      "utf8",
    ),
  ) as unknown;
};
async function shippedScenario(month: string) {
  const read = async (path: string): Promise<unknown> =>
    JSON.parse(
      await readFile(new URL(`../../../../${path}`, import.meta.url), "utf8"),
    ) as unknown;
  const files = await Promise.all(
    ["map.json", "cp.json", "ec.json", "2p.json"].map(
      async (name) => [name, await read(`levels/${month}/${name}`)] as const,
    ),
  );
  const terrains = await read("internal/data/terrain.json");
  if (typeof terrains !== "object" || terrains === null)
    throw Error("missing terrain");
  return scenario.decode(new Map(files), new Set(Object.keys(terrains)));
}
it("creates either campaign side with the selected opponent and native move limit", async () => {
  const data = await loadBattleData({ rule }),
    file = await shippedScenario("1914-08");
  for (const [mode, side] of [
    ["cp", 0],
    ["ec", 1],
  ] as const) {
    const s = newBattleSession(
      file,
      mode,
      side,
      { limit: 10, computer: true },
      data,
    );
    expect(s.game.limit).toBe(10);
    expect(s.game.players[side].computer).toBe(false);
    expect(s.game.players[1 - side]?.computer).toBe(true);
    expect(s.level.mapName).toBe(scenario.level(file, mode).mapName);
    expect(
      newBattleSession(
        file,
        mode,
        side,
        { limit: 0, computer: false },
        data,
      ).game.players.every((p) => !p.computer),
    ).toBe(true);
  }
});
it("restores exact Go-format state without applying new menu game values to it", async () => {
  const data = await loadBattleData({ rule }),
    file = await shippedScenario("1914-08"),
    s = newBattleSession(file, "ec", 1, { limit: 5, computer: false }, data);
  play.confirm(s, 1);
  const raw = play.save(s);
  const resumed = newBattleSession(
    file,
    "ec",
    1,
    { limit: 20, computer: true },
    data,
    raw,
  );
  expect(play.save(resumed)).toEqual(raw);
  expect(resumed.game.limit).toBe(5);
  expect(play.ready(resumed, 1)).toBe(true);
  expect(() =>
    newBattleSession(file, "ec", 2, { limit: 0, computer: false }, data),
  ).toThrow(/side/i);
  expect(() =>
    newBattleSession(file, "ec", 1, { limit: 0, computer: false }, data, "{}"),
  ).toThrow();
});
it("takes the extension units in on ?extended", async () => {
  expect(extendedFromURL("?extended")).toBe(true);
  expect(extendedFromURL("?extended=1")).toBe(true);
  expect(extendedFromURL("?extended=0")).toBe(false);
  expect(extendedFromURL("")).toBe(false);
  const data = await loadBattleData({ rule }),
    file = await shippedScenario("1918-02");
  expect(data.profiles.extended?.tables?.build).toBe(
    `${data.profiles.full.tables?.build ?? ""}+units1`,
  );
  const plain = newBattleSession(
    file,
    "cp",
    0,
    { limit: 0, computer: true },
    data,
  );
  expect(plain.game.profile).toBe(data.profiles.full);
  const s = newBattleSession(
    file,
    "cp",
    0,
    { limit: 0, computer: true },
    { ...data, extendedUnits: true },
  );
  expect(s.game.profile.tables).toBe(data.profiles.extended?.tables);
  expect(
    [...s.game.buildings.values()].some((b) => b.builds.includes(51)),
  ).toBe(true);
  const resumed = newBattleSession(
    file,
    "cp",
    0,
    { limit: 0, computer: true },
    data,
    play.save(s),
  );
  expect(resumed.game.profile.tables?.build).toBe(s.game.profile.tables?.build);
});
