import { expect, it } from "vitest";
import { ai, hexgrid, play, phaseMovement } from "@historyline/engine";
import { Full, Builds } from "../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../engine/src/testing/play.ts";
import { Computer, type ComputerWorker } from "./computer.ts";
import type { PlanRequest, PlanResponse } from "./computerProtocol.ts";
import {
  MapBattle,
  actFastSpeed,
  actRestMs,
  computerStage,
  planMinMs,
} from "./mapBattle.ts";

/**
 * FakeWorker stands in for computer.worker.ts: it keeps a snapshot of each request, as postMessage does, and
 * answers when the test says so, planning on that snapshot.
 */
class FakeWorker implements ComputerWorker {
  readonly requests: PlanRequest[] = [];
  private listener: ((event: { data: PlanResponse }) => void) | null = null;
  postMessage(request: PlanRequest): void {
    this.requests.push(structuredClone(request));
  }
  addEventListener(type: "message" | "error", listener: never): void {
    if (type === "message") this.listener = listener;
  }
  terminate(): void {}
  /** answer plans the oldest open request on its snapshot and returns the commands, or fails it with error. */
  answer(error?: string): play.Command[] {
    const r = this.requests.shift();
    if (r === undefined || this.listener === null)
      throw Error("no open request");
    if (error !== undefined) {
      this.listener({ data: { kind: "error", id: r.id, message: error } });
      return [];
    }
    this.listener({
      data: { kind: "progress", id: r.id, done: 1, total: 4 },
    });
    const commands = ai.commands({}, r.session, r.side);
    this.listener({ data: { kind: "result", id: r.id, commands } });
    return commands;
  }
}
/**
 * fixture is a battle on 1914-08 with the player on side player and the computer on the other, both panels ticked
 * as the split screen does. The month opens with side 0 moving and side 1 attacking.
 */
function fixture(
  o: { player?: 0 | 1; shown?: boolean; session?: play.Session } = {},
) {
  const player = o.player ?? 1,
    computer = 1 - player,
    s =
      o.session ??
      play.newSession(shippedLevel("1914-08"), playData, {
        profile: Full,
        computer: [computer === 0, computer === 1],
      }),
    worker = new FakeWorker(),
    b = new MapBattle(s, player, new Computer(worker)),
    c = new MapBattle(s, computer);
  b.stage = computerStage(c, () => o.shown ?? true);
  let elapsed = 0;
  /** tick runs both panels one frame. */
  const tick = (ms = 16) => {
    b.tick(ms);
    c.tick(ms);
    elapsed += ms;
  };
  /** run ticks frames until done holds or limit ms passed and reports the ms it took. */
  const run = (done: () => boolean, limit = 600_000) => {
    const from = elapsed;
    while (!done() && elapsed - from < limit) tick();
    return elapsed - from;
  };
  return { s, player, computer, worker, b, c, tick, run };
}
/** flush lets the worker's answer reach the battle. */
const flush = () => new Promise((resolve) => setTimeout(resolve, 0));
/** seen is what a view shows its side: every unit and building, the order flag of its own units only. */
const seen = (v: play.View) => ({
  ...v,
  units: new Map(
    [...v.units].map(([k, u]) => [
      k,
      u.side === v.side ? u : { ...u, moved: false },
    ]),
  ),
});
const moves = (cmds: readonly play.Command[]) =>
  cmds.filter((c) => c.kind === "move" || c.kind === "exit").length;

/** movable is a unit of side's with a destination, other than the units in but. */
function movable(s: play.Session, side: number, but: number[] = []) {
  for (const u of play.view(s, side).units.values()) {
    if (u.side !== side || u.head !== null || but.includes(u.index)) continue;
    try {
      const to = play.coords(play.destinations(s, side, u.index));
      if (to.length > 0) return { u, to };
    } catch {
      /* Units without a legal movement order are skipped. */
    }
  }
  throw Error("no movable unit");
}

it("plans as soon as the phase opens, on a snapshot, while the player keeps giving orders", () => {
  const f = fixture({ player: 0 }),
    mark = play.mark(f.s);
  expect(f.worker.requests).toHaveLength(0);
  f.tick();
  expect(f.worker.requests).toHaveLength(1);
  expect(f.worker.requests[0]?.side).toBe(f.computer);
  expect(f.b.computerState()).toBe("planning");
  expect(f.b.progress).not.toBeNull();
  expect(f.b.busy()).toBe(false);
  expect(play.ready(f.s, f.player)).toBe(false);
  // The player gives an order while the computer plans; the planner's snapshot does not see it.
  const { u, to } = movable(f.s, f.player);
  f.b.click(u.at);
  f.b.click(to[0] ?? u.at);
  expect(play.mark(f.s)).toBe(mark + 1);
  expect(play.mark(f.worker.requests[0]?.session ?? f.s)).toBe(mark);
  // Ticking on does not start another run.
  f.run(() => false, 2000);
  expect(f.worker.requests).toHaveLength(1);
});

it("carries out the commands one at a time, each move resting and walking on the computer's panel only", async () => {
  const f = fixture();
  f.tick();
  const planned = f.worker.answer();
  expect(moves(planned)).toBeGreaterThan(1);
  await flush();
  const mark = play.mark(f.s),
    fair = structuredClone(play.view(f.s, f.player)),
    player = structuredClone(f.b.view());
  // The planning shows for a moment before the first command lands.
  f.run(() => false, planMinMs - 32);
  expect(play.mark(f.s)).toBe(mark);
  f.run(() => play.mark(f.s) > mark);
  expect(f.b.computerState()).toBe("acting");
  expect(f.b.progress).toBeNull();
  expect(play.mark(f.s) - mark).toBeLessThan(planned.length);
  // The first move rests on its first hex, then walks.
  const walker = f.c.playback.moving();
  expect(walker).not.toBeNull();
  const start = walker?.at;
  f.run(() => false, actRestMs - 32);
  expect(f.c.playback.moving()?.at).toEqual(start);
  f.run(() => false, 64);
  expect(f.c.playback.moving()?.at).not.toEqual(start);
  expect(f.b.playback.moving()).toBeNull();
  // One command after another: the log grows in steps, and only the computer's panel walks.
  const marks = new Set<number>([play.mark(f.s)]);
  let walks = 0,
    walking = true;
  f.run(() => {
    marks.add(play.mark(f.s));
    const now = f.c.playback.walking();
    if (now && !walking) walks++;
    walking = now;
    expect(f.b.playback.moving()).toBeNull();
    return f.b.computerState() === null;
  });
  expect(marks.size).toBeGreaterThan(2);
  expect(walks).toBeGreaterThan(0);
  expect(play.ready(f.s, f.computer)).toBe(true);
  expect(play.commandsSince(f.s, mark)).toEqual(planned);
  // Fairness: the player's map does not change from the computer's moves before the phase change. (Only an enemy
  // unit's order flag follows the live unit, as in Go's dosplay; the panel shows that flag for own units only.)
  expect(seen(play.view(f.s, f.player))).toEqual(seen(fair));
  expect(seen(f.b.view())).toEqual(seen(player));
  expect(play.phase(f.s, f.computer)).toBe(phaseMovement);
});

it("carries out the rest three times as fast once the player confirms, resolves when both are ready and plans again", async () => {
  const carry = async (confirm: boolean) => {
    const f = fixture();
    f.tick();
    f.worker.answer();
    await flush();
    f.run(() => f.b.computerState() === "acting");
    if (confirm) {
      await f.b.phase();
      expect(play.ready(f.s, f.player)).toBe(true);
      expect(f.worker.requests).toHaveLength(0);
    }
    let resolved = 0;
    f.b.onResolution = () => {
      resolved++;
    };
    const ms = f.run(() => f.b.computerState() === null);
    return { f, ms, resolved };
  };
  const slow = await carry(false),
    fast = await carry(true);
  expect(slow.resolved).toBe(0);
  expect(play.phase(slow.f.s, slow.f.computer)).toBe(phaseMovement);
  expect(fast.resolved).toBe(1);
  expect(play.phase(fast.f.s, fast.f.computer)).not.toBe(phaseMovement);
  expect(fast.ms).toBeLessThan(slow.ms / 2);
  expect(fast.ms).toBeGreaterThan(slow.ms / (actFastSpeed + 1));
  expect(fast.f.c.playback.pace).toBe(1);
  // The next phase's planning waits for the resolution's playback to end.
  const f = fast.f;
  expect(f.worker.requests).toHaveLength(0);
  f.run(() => f.worker.requests.length > 0);
  expect(f.b.playback.resolving()).toBe(false);
  expect(f.b.computerState()).toBe("planning");
  // Once the computer has confirmed, the player's confirm resolves at once.
  f.worker.answer();
  await flush();
  f.run(() => f.b.computerState() === null);
  expect(play.ready(f.s, f.computer)).toBe(true);
  const phase = play.phase(f.s, f.player);
  await f.b.phase();
  expect(play.phase(f.s, f.player)).not.toBe(phase);
  expect(f.worker.requests).toHaveLength(0);
});

it("carries out every command at once where the computer's panel is not shown", async () => {
  const f = fixture({ shown: false });
  f.tick();
  const planned = f.worker.answer();
  await flush();
  const mark = play.mark(f.s);
  f.run(() => f.b.computerState() !== "planning");
  expect(f.b.computerState()).toBeNull();
  expect(play.commandsSince(f.s, mark)).toEqual(planned);
  expect(f.c.playback.walking()).toBe(false);
  expect(play.ready(f.s, f.computer)).toBe(true);
});

it("plans again on the live session when a command fails there, and confirms when that fails too", async () => {
  const f = fixture();
  f.tick();
  const planned = f.worker.answer(),
    first = planned.find((c) => c.kind === "move");
  if (first === undefined) throw Error("no planned move");
  await flush();
  // The live session changes in between: the unit has its order already, so the command fails there.
  play.replay(f.s, [first]);
  f.run(() => f.worker.requests.length > 0);
  expect(f.b.computerState()).toBe("planning");
  expect(play.ready(f.s, f.computer)).toBe(false);
  // The retry plans on the session as it is now and is carried out.
  const mark = play.mark(f.s),
    again = f.worker.answer();
  expect(again).not.toContainEqual(first);
  await flush();
  f.run(() => f.b.computerState() === null);
  expect(play.commandsSince(f.s, mark)).toEqual(again);
  expect(play.ready(f.s, f.computer)).toBe(true);

  // A worker that fails twice leaves the computer's orders as they are: it confirms, and the battle goes on.
  const g = fixture();
  g.tick();
  g.worker.answer("worker unavailable");
  await flush();
  g.run(() => g.worker.requests.length > 0);
  g.worker.answer("worker unavailable");
  await flush();
  g.run(() => g.b.computerState() === null);
  expect(play.ready(g.s, g.computer)).toBe(true);
  expect(g.b.status).toContain("worker unavailable");
  await g.b.phase();
  expect(play.phase(g.s, g.computer)).not.toBe(phaseMovement);
});

it("lands nothing while paused or after dispose, during planning and carrying-out", async () => {
  // Paused while planning: the answer waits for the resume.
  const f = fixture();
  f.tick();
  f.b.setPaused(true);
  f.c.setPaused(true);
  f.worker.answer();
  await flush();
  const before = play.save(f.s);
  f.run(() => false, 10_000);
  expect(play.save(f.s)).toEqual(before);
  f.b.setPaused(false);
  f.c.setPaused(false);
  f.run(() => f.b.computerState() === "acting");
  expect(play.save(f.s)).not.toEqual(before);
  // Paused while carrying out: no further command.
  f.b.setPaused(true);
  f.c.setPaused(true);
  const paused = play.save(f.s);
  f.run(() => false, 10_000);
  expect(play.save(f.s)).toEqual(paused);
  // Disposed while carrying out: nothing more lands.
  f.b.dispose();
  f.c.dispose();
  f.run(() => false, 10_000);
  expect(play.save(f.s)).toEqual(paused);
  expect(play.ready(f.s, f.computer)).toBe(false);

  // Disposed while planning: a late answer is dropped.
  const g = fixture();
  g.tick();
  g.b.dispose();
  const fresh = play.save(g.s);
  g.worker.answer();
  await flush();
  g.run(() => false, 10_000);
  expect(play.save(g.s)).toEqual(fresh);
  expect(g.b.computerState()).toBeNull();
});

it("starts planning again in a loaded battle the computer has not confirmed, and resolves it", async () => {
  const s = play.newSession(shippedLevel("1914-08"), playData, {
    profile: Full,
    computer: [true, false],
  });
  play.confirm(s, 1);
  const f = fixture({ session: play.clone(s, Builds) });
  expect(play.ready(f.s, f.player)).toBe(true);
  f.tick();
  expect(f.b.computerState()).toBe("planning");
  f.worker.answer();
  await flush();
  f.run(() => f.b.computerState() === null);
  expect(play.phase(f.s, f.computer)).not.toBe(phaseMovement);
  // A battle the computer has confirmed in plans nothing.
  const done = play.clone(f.s, Builds);
  play.confirm(done, f.computer);
  const g = fixture({ session: done });
  g.run(() => false, 2000);
  expect(g.worker.requests).toHaveLength(0);
  expect(g.b.computerState()).toBeNull();
});

it("rests a computer move on its first hex at the panel's pace, and only turns it without a rest", () => {
  const s = play.newSession(shippedLevel("1914-08"), playData, {
      profile: Full,
    }),
    c = new MapBattle(s, 0),
    { u, to } = movable(s, 0),
    far = to
      .map((at) => ({ at, n: play.route(s, 0, u.index, at).length }))
      .sort((a, b) => b.n - a.n)[0]?.at;
  if (far === undefined) throw Error("no destination");
  expect(c.act(play.move(s, 0, u.index, far), actRestMs)).toBe(true);
  const first = c.playback.moving()?.at,
    [x, y] = hexgrid.center(u.at);
  expect(first).toEqual({ x, y });
  c.playback.pace = actFastSpeed;
  c.tick(actRestMs / actFastSpeed - 1);
  expect(c.playback.moving()?.at).toEqual(first);
  c.tick(2);
  expect(c.playback.moving()?.at).not.toEqual(first);
  c.tick(100_000);
  expect(c.playback.walking()).toBe(false);
  // Without a rest (its panel hidden) the move only turns the unit.
  const other = movable(s, 0, [u.index]);
  c.playback.directions.delete(other.u.index);
  expect(
    c.act(play.move(s, 0, other.u.index, other.to[0] ?? other.u.at), null),
  ).toBe(false);
  expect(c.playback.directions.has(other.u.index)).toBe(true);
});
