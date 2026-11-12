// The original game's maps as missions, from the files the player imported: each map bound exactly (the engine's
// fromOriginal, at the battle's start) and drawn from its view (origview). A single-player map is offered for
// either side against the computer, a two-player map also to two humans; a win leads to the map the original's
// CODES.DAT names next.
import {
  type bind,
  finmap,
  level,
  moveTablesOf,
  origmaps,
  origview,
  type profilesOf,
  scenario,
} from "@historyline/engine";
import type { Mission } from "../shell/campaign.ts";

export interface OriginalTables {
  looks: origview.Looks;
  terrains: ReadonlySet<string>;
}
export interface OriginalData {
  binding: bind.BindingData;
  profiles: ReturnType<typeof profilesOf>;
}
/** OriginalSource is what an original mission binds and draws. */
export interface OriginalSource {
  stem: string;
  map: finmap.Level;
  view: level.Level;
  /**
   * backdrop is the shipped month made from this map (the same code word, size and winter or not), whose painting
   * the painted design shows; the month follows the map as hlcampaign wrote it, hand edits aside.
   */
  backdrop?: string;
}

export const prefix = "original/";

/** loadOriginalTables reads the look table and the terrain names the view needs. */
export async function loadOriginalTables(loader: {
  rule(name: string): Promise<unknown>;
}): Promise<OriginalTables> {
  const [looks, terrain] = await Promise.all([
    loader.rule("looks.json"),
    loader.rule("terrain.json"),
  ]);
  if (typeof terrain !== "object" || terrain === null)
    throw new Error("missing terrain table");
  return {
    looks: origview.parseLooks(looks),
    terrains: new Set(Object.keys(terrain)),
  };
}

const sideNames = ["The Central Powers", "The Entente"] as const;

/** scenarioFile is the view as a scenario folder, which the menus read; the battle binds the map itself. */
function scenarioFile(
  stem: string,
  view: level.Level,
  winter: boolean,
  modes: readonly scenario.Mode[],
  code: string,
): scenario.File {
  const units: scenario.Unit[] = [],
    buildings: scenario.Building[] = [];
  const map = view.lines.map((line, l) =>
    line.map((h, c) => {
      if (h.unit !== "") units.push({ at: { line: l, col: c }, id: h.unit });
      if (h.content !== null)
        buildings.push({
          at: { line: l, col: c },
          owner: h.content.owner,
          credits: 0,
          units: h.content.units === null ? null : [...h.content.units],
        });
      return { ...h, unit: "", content: null };
    }),
  );
  return {
    month: prefix + stem,
    name: `Original ${stem}`,
    season: winter ? "winter" : "summer",
    seasons: new Map(),
    backdrop: "",
    map,
    setups: new Map(
      modes.map((m, i) => [
        m,
        {
          from: "",
          code: i === 0 ? code : "",
          units,
          buildings,
          builds: [],
          computerBuilds: [],
        },
      ]),
    ),
  };
}

export interface OriginalCatalog {
  missions: Mission[];
  /** Maps that could not be offered, with the reason. */
  skipped: string[];
}

/** originalMissions are the imported maps' missions, in map order: cp, ec and, for two-player maps, 2p. */
export function originalMissions(
  files: ReadonlyMap<string, Uint8Array>,
  data: OriginalData,
  tables: OriginalTables,
  shipped: readonly Pick<Mission, "code" | "month" | "file">[] = [],
): OriginalCatalog {
  const maps = origmaps.read(files, moveTablesOf(data.profiles.full));
  const stems = new Set(maps.map((m) => m.source.map));
  const missions: Mission[] = [],
    skipped: string[] = [];
  for (const map of maps) {
    const stem = map.source.map,
      m = map.mission;
    // A map without a CODES.DAT record (STATMAP) is no mission of the original either.
    if (m === undefined) continue;
    let view: level.Level;
    try {
      view = origview.campaign(
        map,
        data.binding,
        tables.looks,
        tables.terrains,
      );
    } catch (e) {
      skipped.push(`${stem}: ${e instanceof Error ? e.message : String(e)}`);
      continue;
    }
    const modes: scenario.Mode[] =
      m.players === 2 ? ["cp", "ec", "2p"] : ["cp", "ec"];
    const file = scenarioFile(stem, view, m.winter, modes, m.code);
    const next =
      m.next === null ? null : String(m.next).padStart(stem.length, "0");
    const month = shipped.find(
      (s) =>
        s.code === m.code &&
        (s.file.season === "winter") === m.winter &&
        s.file.map.length === view.lines.length &&
        s.file.map[0]?.length === view.lines[0]?.length,
    )?.month;
    const source: OriginalSource = {
      stem,
      map,
      view,
      ...(month === undefined ? {} : { backdrop: month }),
    };
    const season = m.winter ? "winter" : "summer";
    for (const mode of modes) {
      const side = mode === "ec" ? 1 : 0;
      missions.push({
        key: `${prefix}${stem}:${mode}`,
        month: prefix + stem,
        mode,
        side,
        code: m.code,
        name: `Original ${stem}`,
        season,
        briefing:
          mode === "2p"
            ? `Original map ${stem} (${m.code}), ${season} — two human commanders; capture the enemy HQ.`
            : `Original map ${stem} (${m.code}), ${season} — command ${sideNames[side]}${m.players === 2 ? "" : m.humanSides[0] === side ? "" : " (the original's computer side)"}; capture the enemy HQ.`,
        next:
          mode !== "2p" && next !== null && stems.has(next) && !m.campaignEnd
            ? `${prefix}${next}:${mode}`
            : null,
        news: null,
        issue: 0,
        newUnits: [],
        file,
        original: source,
      });
    }
  }
  return { missions, skipped };
}

/** isOriginal reports whether a mission key or month names an original map. */
export function isOriginal(key: string): boolean {
  return key.startsWith(prefix);
}
