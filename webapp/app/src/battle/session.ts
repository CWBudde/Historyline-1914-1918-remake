import {
  bind,
  fromOriginal,
  mustTables,
  play,
  profilesOf,
  scenario,
} from "@historyline/engine";
import type { AssetLoader } from "../assets.ts";
import { record } from "../assetManifest.ts";
import type { SpriteFamily } from "../unitFrame.ts";
import type { Mission } from "../shell/campaign.ts";

export interface BattleData {
  binding: bind.BindingData;
  profiles: ReturnType<typeof profilesOf>;
  families: Readonly<Record<string, SpriteFamily>>;
}
export async function loadBattleData(
  loader: Pick<AssetLoader, "rule">,
): Promise<BattleData> {
  const [units, records, full, demo, sprites, extension] = await Promise.all(
    [
      "units.json",
      "records.json",
      "full.json",
      "demo.json",
      "sprites.json",
      "extension.json",
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
    profiles: profilesOf(mustTables(full), mustTables(demo), extension),
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
    ...(data.profiles.extended === undefined
      ? {}
      : { extended: data.profiles.extended }),
    limit: options.limit,
    computer,
  });
}

/**
 * missionSession is a mission's battle: a shipped setup through newBattleSession, an imported original map bound
 * exactly (fromOriginal) on its view, as the desktop's -level original/NN did.
 */
export function missionSession(
  mission: Pick<Mission, "file" | "mode" | "original">,
  side: number,
  options: { limit: number; computer: boolean },
  data: BattleData,
  saved?: Uint8Array | string,
): play.Session {
  const o = mission.original;
  if (o === undefined)
    return newBattleSession(
      mission.file,
      mission.mode,
      side,
      options,
      data,
      saved,
    );
  if (side !== 0 && side !== 1) throw Error("Invalid battle side.");
  const board = fromOriginal(o.map, data.binding, data.profiles.full),
    view = structuredClone(o.view);
  if (saved !== undefined)
    return play.loadBoard(board, view, data.binding, saved, data.profiles);
  const computer: [boolean, boolean] = [false, false];
  computer[1 - side] = options.computer;
  return play.newBoard(board, view, data.binding, {
    profile: data.profiles.full,
    limit: options.limit,
    computer,
  });
}
