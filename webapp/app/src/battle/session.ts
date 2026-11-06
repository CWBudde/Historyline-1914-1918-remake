import {
  bind,
  mustTables,
  play,
  profilesOf,
  scenario,
} from "@historyline/engine";
import type { AssetLoader } from "../assets.ts";
import { record } from "../assetManifest.ts";
import type { SpriteFamily } from "../unitFrame.ts";

export interface BattleData {
  binding: bind.BindingData;
  profiles: ReturnType<typeof profilesOf>;
  families: Readonly<Record<string, SpriteFamily>>;
}
export async function loadBattleData(
  loader: Pick<AssetLoader, "rule">,
): Promise<BattleData> {
  const [units, records, full, demo, sprites] = await Promise.all(
    [
      "units.json",
      "records.json",
      "full.json",
      "demo.json",
      "sprites.json",
    ].map((name) => loader.rule(name)),
  );
  const families: Record<string, SpriteFamily> = {};
  for (const [id, value] of Object.entries(record(sprites))) {
    const meta = record(value);
    if (
      typeof meta.family !== "string" ||
      (meta.long !== undefined && typeof meta.long !== "string")
    )
      throw Error(`invalid sprite family ${id}`);
    families[id] = {
      family: meta.family,
      ...(typeof meta.long === "string" ? { long: meta.long } : {}),
    };
  }
  return {
    binding: bind.bindingData(units, records),
    profiles: profilesOf(mustTables(full), mustTables(demo)),
    families,
  };
}
/** A restored game's options belong to its Go save, independent of menu defaults. */
export function newBattleSession(
  file: scenario.File,
  mode: scenario.Mode,
  side: number,
  options: { limit: number; computer: boolean },
  data: BattleData,
  saved?: Uint8Array | string,
): play.Session {
  if (side !== 0 && side !== 1) throw Error("Invalid battle side.");
  const lv = scenario.level(file, mode);
  if (saved !== undefined)
    return play.load(lv, data.binding, saved, data.profiles);
  const computer: [boolean, boolean] = [false, false];
  computer[1 - side] = options.computer;
  return play.newSession(lv, data.binding, {
    profile: data.profiles.full,
    limit: options.limit,
    computer,
  });
}
