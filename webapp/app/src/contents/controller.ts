import { hexgrid, level, play, phaseMovement } from "@historyline/engine";

export interface ContentsBox {
  at: hexgrid.Coord;
  hold: number;
  slot: number;
  producing: boolean;
  pick: number;
}
export interface ExitAim {
  hold: number;
  slot: number;
  unit: play.UnitView;
  from: hexgrid.Coord;
  destinations: Set<string>;
}
export interface BuildChoice extends play.BuildOption {
  reason: string | null;
}

/** Native refusal text as a clause, matching dosControl.reason. */
function reason(error: unknown): string {
  const text = (error instanceof Error ? error.message : String(error)).replace(
    /\.$/,
    "",
  );
  return text.charAt(0).toLowerCase() + text.slice(1);
}
function refused(check: () => void): string | null {
  try {
    check();
    return null;
  } catch (error) {
    return reason(error);
  }
}

/** A side's contents cursor; all rules and energy changes belong to play. */
export class ContentsController {
  box: ContentsBox | null = null;
  message = "";

  constructor(
    readonly session: play.Session,
    readonly side: number,
    private readonly blocked: () => boolean = () => false,
  ) {
    if (side !== 0 && side !== 1) throw new Error("Invalid contents side.");
  }

  open(at: hexgrid.Coord): boolean {
    const view = play.view(this.session, this.side);
    let carrier = view.units.get(level.coordKey(at));
    if (carrier?.head !== null && carrier?.head !== undefined)
      carrier = view.units.get(level.coordKey(carrier.head));
    const from =
      carrier?.side === this.side && carrier.carrier
        ? carrier.at
        : (play.entrance(this.session, at) ?? at);
    const content = play.contents(this.session, this.side, from);
    if (content === null) return false;
    if (
      content.slots.length === 0 &&
      play.buildable(this.session, this.side, content.hold).length === 0
    ) {
      this.message = "nothing stored here";
      return false;
    }
    this.box = {
      at: { ...from },
      hold: content.hold,
      slot: content.slots[0]?.slot ?? 0,
      producing: false,
      pick: 0,
    };
    this.message = `${this.name()}: ${content.slots.length} ${content.slots.length === 1 ? "unit" : "units"}`;
    return true;
  }

  /** Cancel returns from production to slots; context changes dismiss all. */
  close(all = false): void {
    if (!all && this.box?.producing === true) this.box.producing = false;
    else this.box = null;
    this.message = "";
  }

  select(slot: number): void {
    if (!Number.isInteger(slot) || slot < 0 || slot >= play.Slots)
      this.fail("no such slot");
    if (this.box !== null) this.box.slot = slot;
  }

  step(direction: "centre" | "up" | "down" | "left" | "right"): void {
    const b = this.box;
    if (b === null) return;
    if (b.producing) {
      this.selectBuild(
        b.pick + (direction === "down" ? 1 : direction === "up" ? -1 : 0),
      );
      return;
    }
    const next =
      direction === "left"
        ? b.slot - 1
        : direction === "right"
          ? b.slot + 1
          : direction === "up" && b.slot % 2 === 0
            ? b.slot + 1
            : direction === "down" && b.slot % 2 === 1
              ? b.slot - 1
              : b.slot;
    if (next >= 0 && next < play.Slots) b.slot = next;
  }

  selectBuild(pick: number): void {
    if (this.box !== null)
      this.box.pick = Math.min(
        Math.max(0, pick),
        Math.max(0, this.options().length - 1),
      );
  }

  slots(): play.SlotView[] {
    if (this.box === null) return [];
    const content = play.contentSlots(this.session, this.side, this.box.at);
    return content?.hold === this.box.hold ? content.slots : [];
  }

  unit(): play.UnitView | null {
    return this.box === null
      ? null
      : (this.slots()[this.box.slot]?.unit ?? null);
  }

  name(): string {
    if (this.box === null) return "contents";
    const view = play.view(this.session, this.side),
      key = level.coordKey(this.box.at);
    const carrier = view.units.get(key);
    if (
      carrier?.side === this.side &&
      carrier.carrier &&
      carrier.hold === this.box.hold
    )
      return carrier.name;
    const b = view.buildings.get(key);
    if (b?.hold !== this.box.hold || b.owner !== play.faction(this.side))
      return "contents";
    return b.kind === play.HQ
      ? "headquarters"
      : b.kind === play.Depot
        ? "depot"
        : "factory";
  }

  resources(): { stored: number; total: number } {
    const stored =
      this.box !== null && this.slots().length === play.Slots
        ? (this.session.game.buildings.get(this.box.hold)?.credits ?? 0)
        : 0;
    return { stored, total: play.energy(this.session, this.side) };
  }

  options(): BuildChoice[] {
    if (this.box === null || this.slots().length !== play.Slots) return [];
    const hold = this.box.hold,
      guard = this.guard();
    return play.buildable(this.session, this.side, hold).map((option) => ({
      ...option,
      reason:
        guard ??
        refused(() => {
          play.canBuild(this.session, this.side, hold, option.type);
        }),
    }));
  }

  actions(): {
    leave: string | null;
    repair: string | null;
    produce: string | null;
  } {
    const b = this.box;
    const common = b === null ? "no contents open" : this.guard();
    if (b === null || common !== null)
      return { leave: common, repair: common, produce: common };
    const u = this.unit(),
      options = this.options();
    return {
      leave: u === null ? "empty slot" : this.cannotLeave(b, u),
      repair:
        u === null
          ? "empty slot"
          : refused(() => {
              play.canRepair(this.session, this.side, b.hold, b.slot);
            }),
      produce: options.some((o) => o.reason === null)
        ? null
        : (options.at(-1)?.reason ?? "nothing to build"),
    };
  }

  showProduction(): void {
    if (this.box === null) return;
    if (this.options().length === 0) {
      this.message = "nothing to build";
      return;
    }
    this.box.producing = true;
    this.box.pick = Math.max(
      0,
      this.options().findIndex((o) => o.reason === null),
    );
    this.message = "produce: pick a type";
  }

  repair(): play.Event[] {
    const box = this.mutableBox(),
      name = this.unit()?.name ?? "unit";
    try {
      const events = play.repair(this.session, this.side, box.hold, box.slot);
      this.message = `${name} repaired - ${play.energy(this.session, this.side)} energy left`;
      return events;
    } catch (error) {
      this.message = `${name}: ${reason(error)}`;
      throw error;
    }
  }

  build(type: number): play.Event[] {
    const box = this.mutableBox();
    const name =
      play
        .buildable(this.session, this.side, box.hold)
        .find((o) => o.type === type)?.name ?? "unit";
    try {
      const events = play.build(this.session, this.side, box.hold, type);
      box.producing = false;
      for (const event of events)
        if (event.kind === play.Built)
          box.slot = play.unit(this.session, event.unit).slot;
      this.message = `${name} built - ${play.energy(this.session, this.side)} energy left`;
      return events;
    } catch (error) {
      this.message = `${name}: ${reason(error)}`;
      throw error;
    }
  }

  prepareExit(): ExitAim {
    const box = this.mutableBox(),
      unit = this.unit();
    if (unit === null) this.fail("empty slot");
    if (play.phase(this.session, this.side) !== phaseMovement)
      this.fail("leaves in the movement phase", unit.name);
    let destinations: Set<string>;
    try {
      destinations = play.exitDestinations(
        this.session,
        this.side,
        box.hold,
        box.slot,
      );
    } catch (error) {
      this.message = `${unit.name}: ${reason(error)}`;
      throw error;
    }
    if (destinations.size === 0) this.fail("nowhere to go", unit.name);
    const aim = {
      hold: box.hold,
      slot: box.slot,
      unit,
      from: { ...box.at },
      destinations,
    };
    this.box = null;
    this.message = `${unit.name} leaves: pick where to go`;
    return aim;
  }

  private cannotLeave(box: ContentsBox, unit: play.UnitView): string | null {
    if (play.phase(this.session, this.side) !== phaseMovement)
      return "leaves in the movement phase";
    return refused(() => {
      if (
        play.exitDestinations(this.session, this.side, box.hold, unit.slot)
          .size === 0
      )
        throw new Error("nowhere to go");
    });
  }

  private guard(): string | null {
    if (play.ended(this.session) !== null) return "the battle is over";
    if (this.blocked()) return "the orders are being carried out";
    if (play.ready(this.session, this.side))
      return "orders confirmed: F1 opens them again";
    return null;
  }

  private mutableBox(): ContentsBox {
    const why = this.guard();
    if (why !== null) this.fail(why);
    if (this.box === null) this.fail("no contents open");
    return this.box;
  }

  private fail(why: string, name?: string): never {
    this.message = name === undefined ? why : `${name}: ${why}`;
    throw new Error(why);
  }
}
