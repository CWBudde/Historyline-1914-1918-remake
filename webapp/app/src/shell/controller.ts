import type { End } from "@historyline/engine";
import {
  parseSave,
  serializeSave,
  validateSettings,
  type BrowserStore,
  type SaveEnvelope,
  type SaveSlot,
  type ScoreEntry,
  type Settings,
} from "../storage/store.ts";
import type { Campaign, Mission, UnitCard } from "./campaign.ts";

export type ShellPage =
  | "main"
  | "code"
  | "values"
  | "control"
  | "extras"
  | "country"
  | "news"
  | "units"
  | "loading"
  | "battle"
  | "leave"
  | "outcome"
  | "scores"
  | "saves"
  | "closed";
export interface BattleStart {
  mission: Mission;
  side: 0 | 1;
  saved?: SaveEnvelope;
}
export interface Outcome {
  won: boolean;
  draw: boolean;
  text: string;
  score: number;
  next: string;
}
export const saveSlots = ["1", "2", "3"] as const;
export interface ShellHooks {
  start(request: BattleStart, signal: AbortSignal): Promise<void>;
  pause(paused: boolean): void;
  discard(): void;
  capture(): SaveEnvelope;
  canSave(): boolean;
  validate(envelope: SaveEnvelope): Mission;
  settings(settings: Settings): void;
}
export type ShellStore = Pick<
  BrowserStore,
  | "listSaves"
  | "getSave"
  | "putSave"
  | "deleteSave"
  | "saveSettings"
  | "listScores"
  | "recordScore"
>;
function message(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}
function validSlot(slot: string): void {
  if (!saveSlots.some((value) => value === slot))
    throw Error("Invalid save slot.");
}

/** Native menu flow; the host owns rendering, battle resources and engine validation. */
export class ShellController {
  page: ShellPage = "main";
  mission: Mission;
  side: 0 | 1 = 0;
  unitPage = 0;
  typed = "";
  note = "";
  busy = false;
  loadingProgress = 0;
  loadingText = "";
  slots: SaveSlot[] = [];
  scores: ScoreEntry[] = [];
  result: Outcome | null = null;
  activeMission: Mission | null = null;
  onChange: (() => void) | null = null;
  private loadAbort: AbortController | null = null;
  private loadGeneration = 0;
  private loadPending = false;
  private pendingSave: SaveEnvelope | undefined;
  private activeSide: 0 | 1 = 0;
  private battleGeneration = 0;
  private ended = false;
  private pendingNext: Mission | null = null;

  constructor(
    readonly campaign: Campaign,
    readonly store: ShellStore | null,
    public settings: Settings,
    private readonly hooks: ShellHooks,
  ) {
    this.mission = campaign.first(0);
    this.settings = validateSettings(settings);
  }
  get cards(): UnitCard[] {
    return this.mission.newUnits.slice(
      this.unitPage * 4,
      this.unitPage * 4 + 4,
    );
  }
  canSave(): boolean {
    return (
      this.activeMission !== null &&
      !this.busy &&
      !this.ended &&
      this.hooks.canSave()
    );
  }
  storageAvailable(): boolean {
    return this.store !== null;
  }
  private changed(): void {
    this.onChange?.();
  }
  private show(page: ShellPage): void {
    this.page = page;
    this.note = "";
    this.changed();
  }
  open(page: ShellPage): void {
    if (this.busy || this.loadPending || this.activeMission !== null) return;
    if (page === "code") this.typed = "";
    this.show(page);
  }
  start(): void {
    if (this.busy || this.page !== "main") return;
    this.pendingSave = undefined;
    this.side = this.mission.side;
    this.show("country");
  }
  submitCode(input = this.typed): void {
    if (this.busy || this.loadPending || this.activeMission !== null) return;
    this.typed = input
      .replace(/[^a-z0-9]/gi, "")
      .toUpperCase()
      .slice(0, 12);
    const mission = this.campaign.findCode(this.typed);
    if (mission !== null) this.mission = mission;
    this.pendingSave = undefined;
    this.page = "main";
    this.note =
      mission === null
        ? `Unknown code ${this.typed}. The map stays ${this.mission.code}.`
        : `Map ${mission.code} selected.`;
    this.changed();
  }
  async chooseSide(side: 0 | 1): Promise<void> {
    if (this.busy || this.page !== "country") return;
    this.side = side;
    this.mission = this.campaign.forSide(this.mission.key, side);
    this.unitPage = 0;
    if (this.mission.news !== null) this.show("news");
    else await this.afterNews();
  }
  async continueNews(): Promise<void> {
    if (this.page === "news" && !this.busy) await this.afterNews();
  }
  private async afterNews(): Promise<void> {
    this.unitPage = 0;
    if (this.mission.newUnits.length > 0) this.show("units");
    else await this.toBattle();
  }
  nextUnits(): void {
    if (this.page !== "units" || this.busy) return;
    if ((this.unitPage + 1) * 4 < this.mission.newUnits.length) {
      this.unitPage++;
      this.changed();
    }
  }
  async toBattle(): Promise<void> {
    if (this.busy || this.loadPending || this.activeMission !== null) return;
    await this.begin(this.pendingSave);
  }
  private async begin(saved?: SaveEnvelope): Promise<void> {
    this.battleGeneration++;
    const request: BattleStart = {
      mission: this.mission,
      side: this.side,
      ...(saved === undefined ? {} : { saved }),
    };
    const generation = ++this.loadGeneration;
    const abort = new AbortController();
    this.loadAbort = abort;
    this.loadPending = true;
    this.result = null;
    this.pendingNext = null;
    this.loadingProgress = 0;
    this.loadingText = "Loading battle…";
    this.show("loading");
    try {
      await this.hooks.start(request, abort.signal);
      if (generation !== this.loadGeneration || abort.signal.aborted) return;
      this.activeMission = request.mission;
      this.activeSide = request.side;
      this.ended = false;
      this.pendingSave = undefined;
      this.loadingProgress = 1;
      this.show("battle");
    } catch (error) {
      if (generation === this.loadGeneration && !abort.signal.aborted) {
        this.note = message(error);
        this.changed();
      }
    } finally {
      if (generation === this.loadGeneration) {
        this.loadPending = false;
        this.changed();
      }
    }
  }
  back(): void {
    if (this.busy) return;
    switch (this.page) {
      case "loading":
        this.cancelLoading();
        break;
      case "leave":
        this.stay();
        break;
      case "scores":
      case "saves":
        this.show("extras");
        break;
      case "news":
        this.show("country");
        break;
      case "units":
        if (this.unitPage > 0) {
          this.unitPage--;
          this.changed();
        } else this.show(this.mission.news !== null ? "news" : "country");
        break;
      case "code":
      case "values":
      case "control":
      case "extras":
      case "country":
      case "outcome":
      case "closed":
        this.show("main");
        break;
      default:
        break;
    }
  }
  cancelLoading(): void {
    if (this.page !== "loading" || this.busy) return;
    this.loadGeneration++;
    this.loadAbort?.abort();
    this.loadAbort = null;
    this.loadPending = false;
    this.pendingSave = undefined;
    this.show("main");
  }
  loading(progress: number, text: string): void {
    if (this.page !== "loading" || !this.loadPending) return;
    this.loadingProgress = Number.isFinite(progress)
      ? Math.max(0, Math.min(1, progress))
      : 0;
    this.loadingText = text;
    this.changed();
  }
  requestLeave(): void {
    if (this.page !== "battle" || this.ended || this.busy) return;
    this.page = "leave";
    this.hooks.pause(true);
    this.show("leave");
  }
  stay(): void {
    if (this.page !== "leave" || this.busy) return;
    this.page = "battle";
    this.hooks.pause(false);
    this.show("battle");
  }
  leave(): void {
    if (this.page !== "leave" || this.busy) return;
    this.hooks.discard();
    this.activeMission = null;
    this.battleGeneration++;
    this.pendingNext = null;
    this.result = null;
    this.show("main");
  }
  async openStorage(page: "saves" | "scores"): Promise<void> {
    if (this.busy || this.loadPending || this.activeMission !== null) return;
    this.show(page);
    await this.refreshStorage();
  }
  private async stored<T>(
    work: (store: ShellStore) => Promise<T>,
  ): Promise<T | null> {
    if (this.busy) return null;
    if (this.store === null) {
      this.note = "Browser storage is unavailable.";
      this.changed();
      return null;
    }
    this.busy = true;
    this.note = "";
    this.changed();
    try {
      return await work(this.store);
    } catch (error) {
      this.note = message(error);
      return null;
    } finally {
      this.busy = false;
      this.changed();
    }
  }
  async refreshStorage(): Promise<void> {
    await this.stored(async (store) => {
      const [slots, scores] = await Promise.all([
        store.listSaves(),
        store.listScores(),
      ]);
      this.slots = slots;
      this.scores = scores;
    });
  }
  async saveAndLeave(slot: string): Promise<void> {
    if (this.page !== "leave" || !this.canSave()) return;
    const saved = await this.stored(async (store) => {
      validSlot(slot);
      const mission = this.activeMission;
      if (mission === null) throw Error("No active battle.");
      const envelope = parseSave(serializeSave(this.hooks.capture()));
      this.hooks.validate(envelope);
      const value: SaveSlot = {
        slot,
        name: `${mission.month} · ${mission.code}`,
        savedAt: Date.now(),
        envelope,
      };
      await store.putSave(value);
      this.slots = [
        ...this.slots.filter((old) => old.slot !== slot),
        value,
      ].sort((a, b) => a.slot.localeCompare(b.slot));
      return true;
    });
    if (saved) {
      this.leave();
      this.note = "Battle saved. Load Game in Extras continues it.";
      this.changed();
    }
  }
  async loadSlot(slot: string): Promise<void> {
    if (this.activeMission !== null || this.loadPending) return;
    const loaded = await this.stored(async (store) => {
      validSlot(slot);
      const saved = await store.getSave(slot);
      if (saved === undefined) throw Error("No saved battle in this slot.");
      const envelope = parseSave(serializeSave(saved.envelope));
      const mission = this.hooks.validate(envelope);
      return { mission, envelope };
    });
    if (loaded === null) return;
    this.mission = loaded.mission;
    this.side = loaded.envelope.side;
    this.pendingSave = loaded.envelope;
    await this.toBattle();
  }
  async importSlot(slot: string, text: string): Promise<void> {
    await this.stored(async (store) => {
      validSlot(slot);
      const envelope = parseSave(text);
      const mission = this.hooks.validate(envelope);
      await store.putSave({
        slot,
        name: `${mission.month} · ${mission.code}`,
        savedAt: Date.now(),
        envelope,
      });
      this.slots = await store.listSaves();
    });
  }
  async exportSlot(slot: string): Promise<string | null> {
    return this.stored(async (store) => {
      validSlot(slot);
      const saved = await store.getSave(slot);
      if (saved === undefined) throw Error("No saved battle in this slot.");
      return serializeSave(saved.envelope);
    });
  }
  async deleteSlot(slot: string): Promise<void> {
    await this.stored(async (store) => {
      validSlot(slot);
      await store.deleteSave(slot);
      this.slots = this.slots.filter((saved) => saved.slot !== slot);
    });
  }
  async updateSettings(patch: Partial<Settings>): Promise<void> {
    if (this.busy) return;
    let next: Settings;
    try {
      next = validateSettings({ ...this.settings, ...patch });
    } catch (error) {
      this.note = message(error);
      this.changed();
      return;
    }
    if (this.store === null) {
      this.settings = next;
      this.hooks.settings(next);
      this.note =
        "Settings apply for this session; they cannot persist because browser storage is unavailable.";
      this.changed();
      return;
    }
    await this.stored(async (store) => {
      await store.saveSettings(next);
      this.settings = next;
      this.hooks.settings(next);
    });
  }
  async endBattle(end: End): Promise<void> {
    const played = this.activeMission;
    if (!end.over || played === null || this.ended) return;
    this.ended = true;
    const generation = this.battleGeneration;
    const draw = end.winner !== 0 && end.winner !== 1;
    const won = !draw && end.winner === this.activeSide;
    const reason =
      end.won[end.winner] === "hq"
        ? "hold the enemy HQ"
        : end.won[end.winner] === "surrender"
          ? "accepted the surrender"
          : "destroyed every enemy unit";
    this.result = {
      won,
      draw,
      text: draw
        ? "Neither side has won."
        : `${end.winner === 0 ? "Central Powers" : "Entente Cordial"} ${reason}`,
      score: end.score,
      next: "",
    };
    const next = won
      ? this.campaign.forSide(played.key, this.activeSide).next
      : null;
    this.pendingNext =
      next === null ? null : (this.campaign.byKey.get(next) ?? null);
    this.changed();
    if (!won) return;
    if (this.store === null) {
      this.note =
        "The best score cannot be kept: browser storage is unavailable.";
      this.changed();
      return;
    }
    try {
      await this.store.recordScore(played.code.toUpperCase(), end.score);
    } catch (error) {
      if (generation === this.battleGeneration) {
        this.note = `The best score was not kept: ${message(error)}`;
        this.changed();
      }
    }
  }
  showOutcome(): void {
    if (this.activeMission === null || this.result === null || !this.ended)
      return;
    this.hooks.discard();
    this.activeMission = null;
    if (this.result.won && this.pendingNext !== null) {
      this.mission = this.pendingNext;
      this.side = this.mission.side;
      this.result.next = this.mission.code;
    }
    this.pendingNext = null;
    this.page = "outcome";
    this.changed();
  }
  continueOutcome(): void {
    if (this.page === "outcome" && !this.busy) this.show("main");
  }
}
