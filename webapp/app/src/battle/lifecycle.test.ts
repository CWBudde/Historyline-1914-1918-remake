import { expect, it } from "vitest";
import { play, phaseMovement } from "@historyline/engine";
import { Full, Builds } from "../../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../../engine/src/testing/play.ts";
import { MapBattle } from "../mapBattle.ts";

/** pendingBattle is a battle whose computer is planning, its answer held until release. */
function pendingBattle() {
  const session = play.newSession(shippedLevel("1914-08"), playData, {
    profile: Full,
    computer: [false, true],
  });
  let release: () => void = () => {};
  const gate = new Promise<void>((r) => {
    release = r;
  });
  const b = new MapBattle(session, 0, {
    commands: async (s, side) => {
      const copy = play.clone(s, Builds),
        mark = play.mark(copy);
      await gate;
      play.confirm(copy, side);
      return play.commandsSince(copy, mark);
    },
  });
  b.tick(16);
  return { b, session, release };
}
const flush = () => new Promise((resolve) => setTimeout(resolve, 0));

it("pauses the computer's carrying-out and playback while the leave page is open", async () => {
  const { b, session, release } = pendingBattle();
  await b.phase();
  b.setPaused(true);
  const before = play.save(session);
  release();
  await flush();
  b.tick(10000);
  expect(play.save(session)).toEqual(before);
  expect(b.computerState()).toBe("planning");
  expect(play.phase(session, 0)).toBe(phaseMovement);
  b.setPaused(false);
  for (let n = 0; n < 200 && b.computerState() !== null; n++) b.tick(16);
  expect(play.phase(session, 0)).not.toBe(phaseMovement);
});
it("disposing a battle releases its pause and discards a late computer result", async () => {
  const { b, session, release } = pendingBattle();
  await b.phase();
  b.setPaused(true);
  const before = play.save(session);
  b.dispose();
  release();
  await flush();
  b.tick(10000);
  expect(play.save(session)).toEqual(before);
  expect(b.busy()).toBe(false);
  expect(b.computerState()).toBeNull();
  await b.phase();
  b.requestSurrender();
  expect(play.save(session)).toEqual(before);
});
