import { campaign, scenario } from "@historyline/engine";
import type { AssetLoader } from "../assets.ts";
import { record } from "../assetManifest.ts";
import type { OriginalSource } from "../original/catalog.ts";

export type Article = campaign.Article;
export interface UnitCard {
  id: string;
  family: string;
  name: string;
  country: string;
  armour: number;
  speed: number;
  attack: string;
  reach: string;
  note: string;
  dos: number;
}
export interface Mission {
  key: string;
  month: string;
  mode: scenario.Mode;
  side: 0 | 1;
  code: string;
  name: string;
  season: scenario.Season;
  briefing: string;
  next: string | null;
  news: Article | null;
  issue: number;
  newUnits: UnitCard[];
  file: scenario.File;
  /** An original map from the player's imported files: the battle binds it exactly (original/catalog.ts). */
  original?: OriginalSource;
}
export interface Campaign {
  missions: Mission[];
  byKey: Map<string, Mission>;
  forSide(key: string, side: 0 | 1): Mission;
  findCode(code: string): Mission | null;
  first(side: 0 | 1): Mission;
  /** The imported original maps' missions (setOriginals), in map order. */
  originals?: Mission[];
}

/** setOriginals replaces the campaign's imported original maps. */
export function setOriginals(c: Campaign, missions: readonly Mission[]): void {
  for (const m of c.originals ?? []) c.byKey.delete(m.key);
  c.originals = [...missions];
  for (const m of missions) c.byKey.set(m.key, m);
}

function text(value: unknown): string {
  if (value == null) return "";
  if (typeof value !== "string") throw new Error("campaign: expected text");
  return value;
}
function integer(value: unknown): number {
  if (value == null) return 0;
  if (typeof value !== "number" || !Number.isSafeInteger(value))
    throw new Error("campaign: expected integer");
  return value;
}
function flag(value: unknown): boolean {
  if (value == null) return false;
  if (typeof value !== "boolean") throw new Error("campaign: expected boolean");
  return value;
}
function unitCards(
  rawUnits: unknown,
  rawSprites: unknown,
  rawNotes: unknown,
): Map<string, UnitCard> {
  const sprites = record(rawSprites);
  const notes = new Map(
    Object.entries(record(rawNotes)).map(([key, v]) => [key, text(v)]),
  );
  const families = new Map(
    Object.entries(sprites).map(([id, raw]) => [
      id,
      text(record(raw)["family"]),
    ]),
  );
  const result = new Map<string, UnitCard>();
  for (const [id, raw] of Object.entries(record(rawUnits))) {
    const u = record(raw),
      family = families.get(id);
    if (family === undefined || family === "")
      throw new Error(`campaign: no sprite family for ${id}`);
    const power = [0, 0, 0];
    let lo = Infinity,
      hi = 0;
    const slots: unknown = u["weapons"] ?? [];
    if (!Array.isArray(slots))
      throw new Error(`campaign: invalid weapons for ${id}`);
    for (const rawSlot of slots as unknown[]) {
      if (rawSlot == null) continue;
      const slot = record(rawSlot);
      if (slot["attack"] == null) continue;
      const at = record(slot["attack"]),
        p = integer(at["power"]);
      const hits = [flag(at["air"]), flag(at["ground"]), flag(at["water"])];
      const min = integer(at["rangeMin"]),
        max = integer(at["rangeMax"]);
      if (p === 0) continue;
      for (const [i, hit] of hits.entries())
        if (hit) power[i] = Math.max(power[i] ?? 0, p);
      lo = Math.min(lo, min);
      hi = Math.max(hi, max);
    }
    const parts = ["Air", "Ground", "Water"].flatMap((name, i) =>
      (power[i] ?? 0) > 0 ? [`${name} ${power[i] ?? 0}`] : [],
    );
    result.set(id, {
      id,
      family,
      name: text(u["name"]),
      country: text(u["country"]),
      armour: integer(u["armour"]),
      speed: integer(u["speed"]),
      dos: integer(u["dos"]),
      note: notes.get(family) ?? "",
      attack: parts.length === 0 ? "No weapons" : parts.join(" · "),
      reach:
        parts.length === 0
          ? ""
          : lo === hi
            ? `Range ${lo}`
            : `Range ${lo}–${hi}`,
    });
  }
  return result;
}

/** The native menu's CP chain, then EC chain, from the exported scenario JSON. */
export async function loadCampaign(
  loader: Pick<AssetLoader, "index" | "levelFiles" | "rule">,
): Promise<Campaign> {
  const [terrain, units, sprites, notes] = await Promise.all([
    loader.rule("terrain.json"),
    loader.rule("units.json"),
    loader.rule("sprites.json"),
    loader.rule("notes.json"),
  ]);
  const terrains = new Set(Object.keys(record(terrain)));
  const cards = unitCards(units, sprites, notes);
  const months = Object.keys(loader.index.levels).sort();
  const loaded = new Map<
    string,
    { file: scenario.File; news: Article | null }
  >();
  let cursor = 0;
  // Each levelFiles call already fetches its few JSON files together.
  await Promise.all(
    Array.from({ length: Math.min(4, months.length) }, async () => {
      while (cursor < months.length) {
        const month = months[cursor++];
        if (month === undefined) break;
        const files = await loader.levelFiles(month);
        const file = scenario.decode(
          new Map([...files].filter(([name]) => name !== "news.json")),
          terrains,
        );
        file.month = month;
        const news = files.has("news.json")
          ? campaign.parse(files.get("news.json"))
          : null;
        loaded.set(month, { file, news });
      }
    }),
  );
  const issues = new Map<string, number>();
  for (const month of months)
    if (loaded.get(month)?.news !== null) issues.set(month, issues.size + 1);
  const missions: Mission[] = [];
  for (const mode of ["cp", "ec"] as const) {
    const side = mode === "cp" ? 0 : 1,
      prefix = mode === "cp" ? "CP_" : "EC_";
    const seen = new Set<string>();
    let previous: Mission | null = null;
    for (const month of months) {
      const entry = loaded.get(month);
      if (entry === undefined)
        throw new Error(`campaign: missing month ${month}`);
      const { file, news } = entry,
        setup = scenario.setup(file, mode);
      if (setup === null || setup.code === "") continue;
      // Check the same setup-to-level invariants as the native catalog.
      scenario.level(file, mode);
      // The build list counts only where a factory stands to build from it.
      const factory = file.map.some((line) =>
        line.some((hex) => hex.terrain.startsWith("factory")),
      );
      const ids = new Set([
        ...setup.units.map((u) => u.id),
        ...setup.buildings.flatMap((b) => b.units ?? []),
        ...(factory ? setup.builds : []),
      ]);
      const newUnits: UnitCard[] = [];
      for (const id of ids) {
        const card = cards.get(id);
        if (!id.startsWith(prefix) || seen.has(id) || card === undefined)
          continue;
        seen.add(id);
        newUnits.push(card);
      }
      newUnits.sort(
        (a, b) => a.dos - b.dos || (a.id < b.id ? -1 : a.id > b.id ? 1 : 0),
      );
      const mission: Mission = {
        key: `${month}:${mode}`,
        month,
        mode,
        side,
        code: setup.code,
        name: file.name,
        season: file.season,
        briefing: `Map ${setup.code}, ${file.season} — command ${side === 0 ? "The Central Powers" : "The Entente"}; capture the enemy HQ.`,
        next: null,
        news,
        issue: issues.get(month) ?? 0,
        newUnits,
        file,
      };
      if (previous !== null) previous.next = mission.key;
      missions.push(mission);
      previous = mission;
    }
  }
  if (missions.length === 0) throw new Error("no campaign maps");
  const byKey = new Map(missions.map((mission) => [mission.key, mission]));
  for (const month of months) {
    const entry = loaded.get(month);
    if (entry === undefined) continue;
    const { file, news } = entry;
    const setup = scenario.setup(file, "2p");
    if (setup === null) continue;
    scenario.level(file, "2p");
    const code = setup.code || scenario.setup(file, "cp")?.code || "";
    const mission: Mission = {
      key: `${month}:2p`,
      month,
      mode: "2p",
      side: 0,
      code,
      name: file.name,
      season: file.season,
      briefing: `Map ${code}, ${file.season} — two human commanders; capture the enemy HQ.`,
      next: null,
      news,
      issue: issues.get(month) ?? 0,
      newUnits: [],
      file,
    };
    byKey.set(mission.key, mission);
  }
  return {
    missions,
    byKey,
    forSide(key, side) {
      const current = byKey.get(key);
      if (current === undefined) throw new Error(`campaign: no mission ${key}`);
      if (current.mode === "2p") return current;
      return (
        byKey.get(`${current.month}:${side === 0 ? "cp" : "ec"}`) ?? current
      );
    },
    findCode(code) {
      return (
        missions.find((m) => m.code === code.toUpperCase()) ??
        [...byKey.values()].find((m) => m.code === code.toUpperCase()) ??
        null
      );
    },
    first(side) {
      const mission = missions.find((m) => m.side === side);
      if (mission === undefined)
        throw new Error(`campaign: no missions for side ${side}`);
      return mission;
    },
  };
}
