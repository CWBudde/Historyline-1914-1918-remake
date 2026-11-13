import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { expect, it } from "vitest";
import type { End } from "@historyline/engine";
import {
  defaultSettings,
  encodeEnvelope,
  parseSave,
  serializeSave,
  type SaveSlot,
  type ScoreEntry,
  type Settings,
} from "../storage/store.ts";
import { repo } from "../../../engine/src/testing/files.ts";
import { AssetLoader } from "../assets.ts";
import {
  loadCampaign,
  type Campaign,
  type Mission,
  type UnitCard,
} from "./campaign.ts";
import {
  ShellController,
  type BattleStart,
  type ShellHooks,
  type ShellStore,
} from "./controller.ts";

function deferred() {
  let resolve!: () => void, reject!: (error: Error) => void;
  const promise = new Promise<void>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}
function catalog(): Campaign {
  const missions: Mission[] = [];
  for (const side of [0, 1] as const) {
    const mode = side === 0 ? "cp" : "ec";
    for (const [index, month] of ["1914-08", "1914-09", "1918-11"].entries()) {
      const cards: UnitCard[] = Array.from(
        { length: index === 0 ? 6 : 0 },
        (_, i) => ({
          id: `${mode}_${i}`,
          family: "infantry",
          name: "Infantry",
          country: "Germany",
          armour: 20,
          speed: 4,
          dos: i,
          attack: "Ground 30",
          reach: "Range 1",
          note: "Infantry",
        }),
      );
      missions.push({
        key: `${month}:${mode}`,
        month,
        mode,
        side,
        code:
          index === 0
            ? side === 0
              ? "PULSE"
              : "BATLE"
            : index === 1
              ? side === 0
                ? "KLUCK"
                : "HISTORY"
              : side === 0
                ? "WATER"
                : "END",
        name: "Field",
        season: "summer",
        briefing: "Orders",
        next:
          index === 2 ? null : `${index === 0 ? "1914-09" : "1918-11"}:${mode}`,
        news: {
          label: "August 1914",
          headline: "War",
          caption: "Infantry",
          deck: "Europe",
          body: ["Report"],
          briefs: ["News"],
        },
        issue: index + 1,
        newUnits: cards,
        file: {
          month,
          name: "Field",
          season: "summer",
          seasons: new Map(),
          backdrop: "",
          map: [],
          setups: new Map(),
        },
      });
    }
  }
  const byKey = new Map(missions.map((m) => [m.key, m]));
  for (const mission of missions.filter((m) => m.mode === "cp")) {
    const local: Mission = {
      ...mission,
      key: `${mission.month}:2p`,
      mode: "2p",
      next: null,
      newUnits: [],
    };
    byKey.set(local.key, local);
  }
  return {
    missions,
    byKey,
    first(side) {
      const m = missions.find((m) => m.side === side);
      if (!m) throw Error("no side");
      return m;
    },
    findCode(code) {
      return missions.find((m) => m.code === code.toUpperCase()) ?? null;
    },
    forSide(key, side) {
      const current = byKey.get(key);
      if (!current) throw Error("no mission");
      if (current.mode === "2p") return current;
      return (
        byKey.get(`${current.month}:${side === 0 ? "cp" : "ec"}`) ?? current
      );
    },
  };
}
function fixture(available = true) {
  const campaign = catalog();
  const slots = new Map<string, SaveSlot>();
  const scores: ScoreEntry[] = [];
  const starts: { request: BattleStart; signal: AbortSignal }[] = [];
  const pauses: boolean[] = [];
  const applied: Settings[] = [];
  const persisted: Settings[] = [];
  const state = {
    discarded: 0,
    validated: 0,
    writes: 0,
    canSave: true,
    fail: "",
    startGate: null as ReturnType<typeof deferred> | null,
    writeGate: null as ReturnType<typeof deferred> | null,
    rejectValidation: false,
  };
  const envelope = encodeEnvelope(
    "1914-08:cp",
    0,
    new TextEncoder().encode('{"schema":1}'),
  );
  const fail = () => {
    if (state.fail) throw Error(state.fail);
  };
  const store: ShellStore = {
    listSaves: () => {
      fail();
      return Promise.resolve([...slots.values()]);
    },
    getSave: (slot) => {
      fail();
      return Promise.resolve(slots.get(slot));
    },
    putSave: async (slot) => {
      state.writes++;
      if (state.writeGate) await state.writeGate.promise;
      fail();
      slots.set(slot.slot, slot);
    },
    deleteSave: (slot) => {
      fail();
      slots.delete(slot);
      return Promise.resolve();
    },
    saveSettings: async (settings) => {
      if (state.writeGate) await state.writeGate.promise;
      fail();
      persisted.push(settings);
    },
    listScores: () => {
      fail();
      return Promise.resolve([...scores]);
    },
    recordScore: (level, score) => {
      fail();
      scores.push({ level, score });
      return Promise.resolve(true);
    },
  };
  const hooks: ShellHooks = {
    start: (request, signal) => {
      starts.push({ request, signal });
      return state.startGate?.promise ?? Promise.resolve();
    },
    pause: (value) => {
      pauses.push(value);
    },
    discard: () => {
      state.discarded++;
    },
    capture: () => envelope,
    canSave: () => state.canSave,
    validate: (env) => {
      state.validated++;
      if (state.rejectValidation) throw Error("game does not fit map");
      const mission = campaign.byKey.get(env.map);
      if (!mission) throw Error("map missing");
      return mission;
    },
    settings: (settings) => {
      applied.push(settings);
    },
  };
  const model = new ShellController(
    campaign,
    available ? store : null,
    defaultSettings(),
    hooks,
  );
  return {
    model,
    campaign,
    slots,
    scores,
    starts,
    pauses,
    applied,
    persisted,
    state,
    envelope,
    store,
    hooks,
  };
}
async function battle(f: ReturnType<typeof fixture>, side: 0 | 1 = 0) {
  f.model.start();
  await f.model.chooseSide(side);
  await f.model.continueNews();
  await f.model.toBattle();
  expect(f.model.page).toBe("battle");
}
const won: End = { over: true, winner: 0, won: ["hq", ""], score: 750 };

it.each(["simultaneous", "hotseat"] as const)(
  "starts the selected month's local setup directly through news for %s humans",
  async (playMode) => {
    const f = fixture();
    f.model.submitCode("HISTORY");
    f.model.unitPage = 3;
    await f.model.updateSettings({ playMode, layout: "original" });
    expect(f.persisted[0]).toMatchObject({ playMode, layout: "original" });
    f.model.start();
    expect(f.model.mission.key).toBe("1914-09:2p");
    expect(f.model.side).toBe(0);
    expect(f.model.unitPage).toBe(0);
    expect(f.model.page).toBe("news");
    await f.model.chooseSide(1);
    expect(f.model.mission.key).toBe("1914-09:2p");
    await f.model.continueNews();
    expect(f.model.page).toBe("battle");
    expect(f.starts[0]?.request).toMatchObject({
      mission: { key: "1914-09:2p" },
      side: 0,
    });
  },
);

it("backs out of local news to the menu and skips absent news without showing country or cards", async () => {
  const f = fixture();
  await f.model.updateSettings({ playMode: "hotseat" });
  f.model.start();
  f.model.back();
  expect(f.model.page).toBe("main");
  const mission = f.campaign.byKey.get("1914-08:2p");
  if (!mission) throw Error("missing local fixture");
  mission.news = null;
  f.model.start();
  await new Promise((resolve) => setTimeout(resolve, 0));
  expect(f.model.page).toBe("battle");
  expect(f.starts).toHaveLength(1);
});

it("restores local native saves with their saved human side and returns to that side's campaign for a new single game", async () => {
  const f = fixture();
  const saved = { ...f.envelope, map: "1914-09:2p", side: 1 as const };
  await f.model.importSlot("2", serializeSave(saved));
  await f.model.loadSlot("2");
  expect(f.model.page).toBe("battle");
  expect(f.model.side).toBe(1);
  expect(f.starts[0]?.request).toMatchObject({
    mission: { key: "1914-09:2p" },
    side: 1,
    saved,
  });
  // The host captures the active local side; the shell keeps its native bytes.
  f.hooks.capture = () => saved;
  f.model.requestLeave();
  await f.model.saveAndLeave("1");
  expect(f.slots.get("1")?.envelope).toEqual(saved);
  await f.model.updateSettings({ playMode: "single" });
  f.model.start();
  expect(f.model.page).toBe("country");
  expect(f.model.mission.key).toBe("1914-09:ec");
  expect(f.model.side).toBe(1);
  await f.model.chooseSide(0);
  expect(f.model.mission.key).toBe("1914-09:cp");
});

it.each([0, 1] as const)(
  "shows side %s winning a local battle without recording or advancing a campaign win",
  async (winner) => {
    const f = fixture();
    await f.model.updateSettings({ playMode: "simultaneous" });
    f.model.start();
    await f.model.continueNews();
    // Preferences for the next battle cannot change this local outcome.
    await f.model.updateSettings({ playMode: "single", opponent: "passive" });
    expect(f.model.activeMission?.key).toBe("1914-08:2p");
    await f.model.endBattle({
      ...won,
      winner,
      won: winner === 0 ? ["hq", ""] : ["", "hq"],
    });
    f.model.showOutcome();
    expect(f.model.page).toBe("outcome");
    expect(f.model.result).toMatchObject({ won: true, draw: false, next: "" });
    expect(f.model.result?.text).toContain(
      winner === 0 ? "Central Powers" : "Entente Cordial",
    );
    expect(f.model.mission.key).toBe("1914-08:2p");
    expect(f.scores).toEqual([]);
  },
);

it("follows native code, country, news, card pagination and Escape transitions", async () => {
  const { model } = fixture();
  let updates = 0;
  model.onChange = () => {
    updates++;
  };
  model.open("code");
  model.typed = "b-a t!le";
  model.submitCode();
  expect(model.mission.code).toBe("BATLE");
  expect(model.page).toBe("main");
  const selected = model.mission;
  model.submitCode("unknown");
  expect(model.mission).toBe(selected);
  expect(model.note).toContain("Unknown code");
  model.start();
  expect(model.page).toBe("country");
  expect(model.side).toBe(1);
  await model.chooseSide(0);
  expect(model.mission.code).toBe("PULSE");
  expect(model.page).toBe("news");
  model.back();
  expect(model.page).toBe("country");
  await model.chooseSide(0);
  await model.continueNews();
  expect(model.page).toBe("units");
  expect(model.cards.map((c) => c.id)).toEqual([
    "cp_0",
    "cp_1",
    "cp_2",
    "cp_3",
  ]);
  model.nextUnits();
  model.nextUnits();
  expect(model.unitPage).toBe(1);
  expect(model.cards.map((c) => c.id)).toEqual(["cp_4", "cp_5"]);
  model.back();
  expect(model.unitPage).toBe(0);
  model.back();
  expect(model.page).toBe("news");
  model.back();
  model.back();
  expect(model.page).toBe("main");
  for (const page of [
    "code",
    "values",
    "control",
    "extras",
    "country",
    "outcome",
  ] as const) {
    model.open(page);
    model.back();
    expect(model.page).toBe("main");
  }
  for (const page of ["saves", "scores", "offline"] as const) {
    model.open(page);
    model.back();
    expect(model.page).toBe("extras");
  }
  expect(updates).toBeGreaterThan(0);
});

it("skips missing news and empty card pages", async () => {
  const f = fixture();
  const mission = f.campaign.byKey.get("1914-09:cp");
  if (mission === undefined) throw Error("missing fixture mission");
  f.model.mission = mission;
  f.model.mission.news = null;
  f.model.start();
  await f.model.chooseSide(0);
  expect(f.model.page).toBe("battle");
  expect(f.starts[0]?.request.mission.key).toBe("1914-09:cp");
  const second = fixture();
  second.model.mission.news = null;
  second.model.start();
  await second.model.chooseSide(0);
  expect(second.model.page).toBe("units");
  second.model.back();
  expect(second.model.page).toBe("country");
});

it("snapshots loading requests, cancels stale starts and refuses duplicate starts", async () => {
  const f = fixture();
  const gate = deferred();
  f.state.startGate = gate;
  const pending = f.model.toBattle();
  expect(f.model.page).toBe("loading");
  f.model.loading(0.5, "Map");
  expect(f.model.loadingProgress).toBe(0.5);
  expect(f.model.loadingText).toBe("Map");
  await f.model.toBattle();
  expect(f.starts).toHaveLength(1);
  f.model.back();
  expect(f.starts[0]?.signal.aborted).toBe(true);
  expect(f.model.page).toBe("main");
  f.model.submitCode("BATLE");
  f.state.startGate = null;
  await f.model.toBattle();
  expect(f.model.activeMission?.code).toBe("BATLE");
  gate.resolve();
  await pending;
  expect(f.model.activeMission?.code).toBe("BATLE");
  expect(f.model.page).toBe("battle");
  expect(f.starts[0]?.request.mission.code).toBe("PULSE");
});

it("keeps failed loads retryable and ignores obsolete rejections", async () => {
  const f = fixture();
  f.state.startGate = deferred();
  const first = f.model.toBattle();
  f.state.startGate.reject(Error("image missing"));
  await first;
  expect(f.model.page).toBe("loading");
  expect(f.model.note).toContain("image missing");
  expect(f.model.activeMission).toBeNull();
  f.state.startGate = null;
  await f.model.toBattle();
  expect(f.model.page).toBe("battle");
  f.model.requestLeave();
  f.model.leave();
  const gate = deferred();
  f.state.startGate = gate;
  const stale = f.model.toBattle();
  f.model.cancelLoading();
  gate.reject(Error("old load failed"));
  await stale;
  expect(f.model.page).toBe("main");
  expect(f.model.note).toBe("");
});

it("preserves a paused battle when saving fails and leaves only after transaction success", async () => {
  const f = fixture();
  await battle(f);
  const active = f.model.activeMission;
  f.model.requestLeave();
  expect(f.pauses).toEqual([true]);
  expect(f.model.page).toBe("leave");
  f.state.fail = "quota exceeded";
  await f.model.saveAndLeave("1");
  expect(f.model.page).toBe("leave");
  expect(f.model.activeMission).toBe(active);
  expect(f.state.discarded).toBe(0);
  expect(f.pauses).toEqual([true]);
  expect(f.model.note).toContain("quota");
  f.state.fail = "";
  f.state.writeGate = deferred();
  const saving = f.model.saveAndLeave("1");
  expect(f.model.busy).toBe(true);
  expect(f.model.canSave()).toBe(false);
  f.model.stay();
  f.model.leave();
  f.model.back();
  expect(f.model.page).toBe("leave");
  await f.model.saveAndLeave("2");
  expect(f.state.writes).toBe(2);
  f.state.writeGate.resolve();
  await saving;
  expect(f.model.page).toBe("main");
  expect(f.model.activeMission).toBeNull();
  expect(f.state.discarded).toBe(1);
  expect(f.slots.get("1")).toMatchObject({
    slot: "1",
    name: "1914-08 · PULSE",
    envelope: f.envelope,
  });
  expect(f.slots.has("2")).toBe(false);
});

it("resumes the same battle after Stay and discards explicitly on Leave", async () => {
  const f = fixture();
  await battle(f);
  const active = f.model.activeMission;
  f.model.requestLeave();
  f.model.back();
  expect(f.pauses).toEqual([true, false]);
  expect(f.model.activeMission).toBe(active);
  expect(f.model.page).toBe("battle");
  f.model.requestLeave();
  f.model.leave();
  expect(f.state.discarded).toBe(1);
  expect(f.model.page).toBe("main");
  expect(f.model.activeMission).toBeNull();
});

it("validates imports before overwriting and exports the exact native envelope", async () => {
  const f = fixture();
  const old: SaveSlot = {
    slot: "1",
    name: "old",
    savedAt: 1,
    envelope: f.envelope,
  };
  f.slots.set("1", old);
  await f.model.importSlot("1", "broken");
  expect(f.state.writes).toBe(0);
  expect(f.slots.get("1")).toBe(old);
  f.state.rejectValidation = true;
  await f.model.importSlot("1", serializeSave(f.envelope));
  expect(f.state.validated).toBe(1);
  expect(f.state.writes).toBe(0);
  expect(f.slots.get("1")).toBe(old);
  f.state.rejectValidation = false;
  await f.model.importSlot("1", serializeSave(f.envelope));
  expect(f.state.writes).toBe(1);
  expect(f.model.slots).toHaveLength(1);
  const exported = await f.model.exportSlot("1");
  expect(exported).not.toBeNull();
  if (exported === null) throw Error("missing exported save");
  expect(parseSave(exported)).toEqual(f.envelope);
  await f.model.deleteSlot("1");
  expect(f.slots.size).toBe(0);
  expect(f.model.slots).toEqual([]);
  await f.model.importSlot("4", serializeSave(f.envelope));
  expect(f.state.writes).toBe(1);
  expect(f.model.note).toContain("slot");
});

it("restores the original setup and saved human side only after validating it", async () => {
  const f = fixture();
  const saved = { ...f.envelope, side: 1 as const };
  f.slots.set("2", {
    slot: "2",
    savedAt: 1,
    name: "restored",
    envelope: saved,
  });
  f.state.rejectValidation = true;
  await f.model.loadSlot("2");
  expect(f.model.page).toBe("main");
  expect(f.starts).toEqual([]);
  expect(f.model.mission.code).toBe("PULSE");
  f.state.rejectValidation = false;
  await f.model.loadSlot("2");
  expect(f.model.page).toBe("battle");
  expect(f.model.side).toBe(1);
  expect(f.model.activeMission?.key).toBe("1914-08:cp");
  expect(f.starts[0]?.request).toMatchObject({
    mission: { key: "1914-08:cp" },
    side: 1,
    saved,
  });
  await f.model.endBattle({ ...won, winner: 1, won: ["", "surrender"] });
  f.model.showOutcome();
  expect(f.model.mission.key).toBe("1914-09:ec");
  expect(f.model.result?.next).toBe("HISTORY");
});

it("persists validated settings before applying and serializes overlapping operations", async () => {
  const f = fixture();
  const before = f.model.settings;
  f.state.writeGate = deferred();
  const update = f.model.updateSettings({ scroll: "fast", limit: 10 });
  expect(f.model.settings).toBe(before);
  expect(f.applied).toEqual([]);
  expect(f.model.busy).toBe(true);
  await f.model.updateSettings({ scenes: false });
  f.state.writeGate.resolve();
  await update;
  expect(f.model.settings).toMatchObject({
    scroll: "fast",
    limit: 10,
    scenes: true,
  });
  expect(f.applied).toHaveLength(1);
  const kept = f.model.settings;
  f.state.fail = "storage unavailable";
  await f.model.updateSettings({ painted: false });
  expect(f.model.settings).toBe(kept);
  expect(f.applied).toHaveLength(1);
  expect(f.model.note).toContain("storage");
  await f.model.updateSettings({ uiScale: 9 });
  expect(f.model.settings).toBe(kept);
  const local = fixture(false);
  await local.model.updateSettings({ scenes: false });
  expect(local.model.settings.scenes).toBe(false);
  expect(local.model.note).toContain("persist");
  expect(local.model.storageAvailable()).toBe(false);
});

it("shows storage failures and never clears the previously loaded lists", async () => {
  const f = fixture();
  f.slots.set("1", {
    slot: "1",
    name: "old",
    savedAt: 1,
    envelope: f.envelope,
  });
  await f.model.openStorage("saves");
  expect(f.model.page).toBe("saves");
  expect(f.model.slots).toHaveLength(1);
  f.state.fail = "database damaged";
  await f.model.refreshStorage();
  expect(f.model.slots).toHaveLength(1);
  expect(f.model.note).toContain("damaged");
  expect(await f.model.exportSlot("1")).toBeNull();
  expect(f.model.note).toContain("damaged");
  const unavailable = fixture(false);
  await unavailable.model.openStorage("scores");
  expect(unavailable.model.note).toContain("unavailable");
  await unavailable.model.loadSlot("1");
  expect(unavailable.starts).toEqual([]);
});

it("records human wins exactly once, waits for explicit outcome and advances only once", async () => {
  const f = fixture();
  await battle(f);
  await f.model.endBattle(won);
  await f.model.endBattle(won);
  expect(f.scores).toEqual([{ level: "PULSE", score: 750 }]);
  expect(f.model.page).toBe("battle");
  expect(f.model.mission.code).toBe("PULSE");
  expect(f.state.discarded).toBe(0);
  expect(f.model.result).toMatchObject({
    won: true,
    draw: false,
    text: "Central Powers hold the enemy HQ",
    score: 750,
  });
  f.model.showOutcome();
  f.model.showOutcome();
  expect(f.model.page).toBe("outcome");
  expect(f.model.mission.code).toBe("KLUCK");
  expect(f.model.result?.next).toBe("KLUCK");
  expect(f.state.discarded).toBe(1);
  f.model.continueOutcome();
  expect(f.model.page).toBe("main");
  await f.model.endBattle(won);
  expect(f.scores).toHaveLength(1);
});
it.each([
  {
    end: {
      over: true,
      winner: 1,
      won: ["", "annihilation"],
      score: 500,
    } satisfies End,
    text: "Entente Cordial destroyed every enemy unit",
    draw: false,
  },
  {
    end: {
      over: true,
      winner: -1,
      won: ["annihilation", "annihilation"],
      score: 0,
    } satisfies End,
    text: "Neither side has won.",
    draw: true,
  },
])(
  "does not score or advance defeat/draw: $text",
  async ({ end, text, draw }) => {
    const f = fixture();
    await battle(f);
    await f.model.endBattle(end);
    f.model.showOutcome();
    expect(f.model.result).toMatchObject({ won: false, draw, text, next: "" });
    expect(f.model.mission.code).toBe("PULSE");
    expect(f.scores).toEqual([]);
  },
);
it("handles final-month victory and score failure without losing the outcome", async () => {
  const f = fixture();
  f.model.submitCode("WATER");
  await battle(f);
  f.state.fail = "score not kept";
  await f.model.endBattle({ ...won, won: ["surrender", ""] });
  expect(f.model.note).toContain("score not kept");
  expect(f.model.page).toBe("battle");
  f.model.showOutcome();
  expect(f.model.result).toMatchObject({
    won: true,
    next: "",
    text: "Central Powers accepted the surrender",
  });
  expect(f.model.mission.code).toBe("WATER");
  expect(f.model.note).toContain("score not kept");
});

it("retries a failed restoration with its original saved data", async () => {
  const f = fixture();
  f.slots.set("1", {
    slot: "1",
    name: "saved",
    savedAt: 1,
    envelope: f.envelope,
  });
  f.state.startGate = deferred();
  const pending = f.model.loadSlot("1");
  // Storage validation finishes before the host begins restoring.
  await new Promise((resolve) => setTimeout(resolve, 0));
  f.state.startGate.reject(Error("textures unavailable"));
  await pending;
  expect(f.model.page).toBe("loading");
  f.state.startGate = null;
  await f.model.toBattle();
  expect(f.starts.map((start) => start.request.saved)).toEqual([
    f.envelope,
    f.envelope,
  ]);
  expect(f.model.page).toBe("battle");
});

it("does not overwrite a later loading error with a previous battle's delayed score failure", async () => {
  const f = fixture();
  await battle(f);
  const score = deferred();
  f.store.recordScore = async () => {
    await score.promise;
    return true;
  };
  const recording = f.model.endBattle(won);
  f.model.showOutcome();
  f.model.continueOutcome();
  const load = deferred();
  f.state.startGate = load;
  const starting = f.model.toBattle();
  load.reject(Error("new map unavailable"));
  await starting;
  score.reject(Error("old score failed"));
  await recording;
  expect(f.model.note).toBe("new map unavailable");
});

it("advances representative points in the shipped 52-month chains, including the final missions", async () => {
  const root = join(repo, "build/webapp/data");
  const fetcher: typeof fetch = async (url) => {
    const pathname = new URL(
      url instanceof Request ? url.url : url.toString(),
    ).pathname.replace("/data/", "");
    return new Response(await readFile(join(root, pathname)));
  };
  const campaign = await loadCampaign(
    await AssetLoader.open("https://example.test/data/index.json", {
      fetch: fetcher,
      cache: null,
    }),
  );
  for (const [key, next, code] of [
    ["1914-08:cp", "1914-09:cp", "KLUCK"],
    ["1914-08:ec", "1914-09:ec", "MARNE"],
    ["1914-12:cp", "1915-01:cp", "CROUY"],
    ["1918-10:cp", "1918-11:cp", "WATER"],
    ["1918-10:ec", "1918-11:ec", "TRUCE"],
    ["1918-11:cp", "1918-11:cp", ""],
    ["1918-11:ec", "1918-11:ec", ""],
  ] as const) {
    const base = fixture();
    const model = new ShellController(campaign, base.store, defaultSettings(), {
      start: () => Promise.resolve(),
      pause: () => undefined,
      discard: () => undefined,
      capture: () => base.envelope,
      canSave: () => true,
      validate: () => campaign.first(0),
      settings: () => undefined,
    });
    const mission = campaign.byKey.get(key);
    if (!mission) throw Error("missing campaign mission");
    model.mission = mission;
    model.side = mission.side;
    await model.toBattle();
    await model.endBattle({
      ...won,
      winner: mission.side,
      won: mission.side === 0 ? ["hq", ""] : ["", "hq"],
    });
    expect(model.mission.key).toBe(key);
    model.showOutcome();
    expect(model.mission.key).toBe(next);
    expect(model.result?.next).toBe(code);
  }
});

it("publishes the target page before a pause hook notifies the host", async () => {
  const f = fixture();
  await battle(f);
  const observed: [boolean, string][] = [];
  f.hooks.pause = (paused) => {
    observed.push([paused, f.model.page]);
  };
  f.model.requestLeave();
  f.model.stay();
  expect(observed).toEqual([
    [true, "leave"],
    [false, "battle"],
  ]);
});
