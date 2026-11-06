import { expect, it } from "vitest";
import { play, phaseMovement } from "@historyline/engine";
import { Full, Builds } from "../../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../../engine/src/testing/play.ts";
import { MapBattle } from "../mapBattle.ts";

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
  return { b, session, release };
}
it("pauses pending computer application and playback while the leave page is open", async () => {
  const { b, session, release } = pendingBattle(),
    pending = b.phase();
  b.setPaused(true);
  const before = play.save(session);
  release();
  await new Promise((resolve) => setTimeout(resolve, 0));
  expect(play.save(session)).toEqual(before);
  expect(b.busy()).toBe(true);
  b.tick(10000);
  expect(play.phase(session, 0)).toBe(phaseMovement);
  b.setPaused(false);
  await pending;
  expect(play.phase(session, 0)).not.toBe(phaseMovement);
});
it("disposing a battle releases its pause and discards a late computer result", async () => {
  const { b, session, release } = pendingBattle(),
    pending = b.phase();
  b.setPaused(true);
  const before = play.save(session);
  b.dispose();
  release();
  await pending;
  expect(play.save(session)).toEqual(before);
  expect(b.busy()).toBe(false);
  await b.phase();
  b.requestSurrender();
  expect(play.save(session)).toEqual(before);
});
