import { Worker as Thread } from "node:worker_threads";
import { afterEach, expect, it } from "vitest";
import { Computer, type ComputerWorker } from "./computer.ts";
import type { ComputerRequest, PlanResponse } from "./computerProtocol.ts";
import { ai, play } from "@historyline/engine";
import { Full, Builds } from "../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../engine/src/testing/play.ts";

const threads: Thread[] = [];
afterEach(async () => {
  await Promise.all(threads.splice(0).map((t) => t.terminate()));
});
function computer(): Computer {
  const thread = new Thread(
    new URL("./testing/computerThread.ts", import.meta.url),
    { execArgv: ["--experimental-strip-types"] },
  );
  threads.push(thread);
  const port = {
    postMessage: (r: ComputerRequest) => {
      thread.postMessage(r);
    },
    addEventListener: (type: string, listener: (event: never) => void) => {
      if (type === "message")
        thread.on("message", (data: PlanResponse) => {
          (listener as (event: { data: PlanResponse }) => void)({ data });
        });
      else
        thread.on("error", (e: Error) => {
          (listener as (event: { message: string }) => void)({
            message: e.message,
          });
        });
    },
    terminate: () => {
      void thread.terminate();
    },
  } as ComputerWorker;
  return new Computer(port);
}
it("runs the browser worker entry off-thread, returns fair commands and progress in both phases", async () => {
  const c = computer(),
    s = play.newSession(shippedLevel("1914-08"), playData, {
      profile: Full,
      computer: [true, true],
    });
  for (let n = 0; n < 4; n++) {
    const saved = play.save(s),
      shadow = play.clone(s, Builds),
      progress: [number, number][] = [];
    const result = await c.commands(s, 1, (done, total) => {
      progress.push([done, total]);
    });
    expect(play.save(s)).toEqual(saved);
    const mark = play.mark(shadow),
      want = ai.plan({}, shadow, 1);
    expect(result).toEqual(play.commandsSince(shadow, mark));
    expect(progress[0]?.[0]).toBe(0);
    expect(progress.at(-1)?.[0]).toBe(progress.at(-1)?.[1]);
    expect(play.replay(s, result)).toEqual(want);
    ai.plan({}, s, 0);
    play.resolve(s);
  }
  c.dispose();
}, 30_000);
it("can plan while the other side gives orders, then replay onto the live session", async () => {
  const c = computer(),
    a = play.newSession(shippedLevel("1914-08"), playData, {
      profile: Full,
      computer: [true, true],
    }),
    b = play.clone(a, Builds);
  const pending = c.commands(a, 1);
  const computerEvents = ai.plan({}, b, 1);
  ai.plan({}, b, 0);
  ai.plan({}, a, 0);
  expect(play.replay(a, await pending)).toEqual(computerEvents);
  expect(play.save(a)).toEqual(play.save(b));
  c.dispose();
});
it("reports worker refusals and rejects in-flight and future work on disposal", async () => {
  const c = computer(),
    s = play.newSession(shippedLevel("1914-08"), playData, { profile: Full });
  await expect(c.commands(s, 2)).rejects.toThrow("dosai: side 2");
  const pending = c.commands(s, 0),
    check = expect(pending).rejects.toThrow("Computer stopped");
  c.dispose();
  await check;
  await expect(c.commands(s, 0)).rejects.toThrow("Computer stopped");
});
it("runs PlayOut off-thread and returns the complete replayable command cycle", async () => {
  const c = computer(),
    s = play.newSession(shippedLevel("1914-08"), playData, {
      profile: Full,
      computer: [true, true],
    }),
    shadow = play.clone(s, Builds),
    before = play.save(s);
  const pending = c.playOut(s, 4);
  ai.playOut(shadow, [{}, {}], 4);
  const result = await pending;
  expect(play.save(s)).toEqual(before);
  expect(result).toEqual(play.commandsSince(shadow, 0));
  expect(result.filter((c) => c.kind === "resolve")).toHaveLength(4);
  play.replay(s, result);
  expect(play.save(s)).toEqual(play.save(shadow));
  c.dispose();
}, 30_000);
