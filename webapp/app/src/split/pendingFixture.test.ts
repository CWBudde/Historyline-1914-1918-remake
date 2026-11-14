import { readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { expect, it } from "vitest";
import { play } from "@historyline/engine";
import { Builds } from "../../../engine/src/testing/tables.ts";
import { repo } from "../../../engine/src/testing/files.ts";
import { playData, shippedLevel } from "../../../engine/src/testing/play.ts";
import {
  decodeEnvelope,
  decodePendingBefore,
  encodeEnvelope,
  parseSave,
  serializeSave,
  type SaveEnvelope,
} from "../storage/store.ts";
import { LocalMatch } from "./match.ts";
import { goEvents, playEvents, restorePending } from "./pending.ts";

// Two fixtures carry a hot-seat save with a pending playback between the editions: the desktop game writes
// pending-desktop.json (cmd/historyline/pending_fixture_test.go), this test writes pending-web.json
// (HL_UPDATE_PENDING=1 rewrites it), and each edition resumes the other's. Both hold August 1914's two-player
// battle after a resolution with a battle that the left human saw and the right one has still to see, saved
// behind the right human's curtain.
const desktopFile = join(repo, "cmd/historyline/testdata/pending-desktop.json"),
  webFile = join(repo, "webapp/app/src/split/testdata/pending-web.json");

const load = (bytes: Uint8Array) =>
  play.load(shippedLevel("1914-08", "2p"), playData, bytes, Builds);

function readFixture(file = desktopFile): SaveEnvelope {
  const saved = parseSave(readFileSync(file, "utf8"));
  expect(saved).toMatchObject({ map: "1914-08:2p", side: 1, mode: "hotseat" });
  expect(saved.pending?.side).toBe(1);
  return saved;
}

it("reads the desktop's pending events as the engine resolves the game before them", () => {
  const saved = readFixture();
  if (saved.pending === undefined) throw Error("no pending playback");
  const p = restorePending(saved.pending, load);
  if (p === null) throw Error("the desktop's pending playback does not read");
  const before = load(p.before);
  expect(play.ready(before, 0) && play.ready(before, 1)).toBe(true);
  expect(p.view).toEqual(play.view(before, 1));
  const events = play.resolve(before);
  expect(events.some((e) => e.kind === play.Battle)).toBe(true);
  expect(p.events).toEqual(events);
  expect(playEvents(goEvents(events))).toEqual(events);
});

it.each([
  ["desktop's", desktopFile],
  ["web edition's", webFile],
])(
  "resumes the %s save with the right human's playback at its takeover",
  (_, file) => {
    const saved = readFixture(file);
    if (saved.pending === undefined || saved.mode === undefined)
      throw Error("no pending playback");
    const p = restorePending(saved.pending, load);
    if (p === null) throw Error("the desktop's pending playback does not read");
    const session = load(decodeEnvelope(saved)),
      match = new LocalMatch(session, saved.mode, saved.side);
    const right = match.sides[1];
    // The panels apply the scene setting before the runtime restores the playback.
    right.playback.setScenes(true);
    match.restore(p);
    // Until the takeover the right map shows the game before the resolution and keeps its battle.
    expect(match.handover).toBe(1);
    expect(right.view()).toEqual(p.view);
    expect(right.busy()).toBe(true);
    right.tick(100_000);
    expect(right.view()).toEqual(p.view);
    expect(match.savedPending()?.side).toBe(1);
    match.takeOver();
    expect(match.savedPending()).toBeNull();
    expect(right.playback.scene()?.battle).toEqual(
      p.events.find((e) => e.kind === play.Battle)?.battle,
    );
    while (right.busy()) right.tick(100_000);
    expect(right.view()).toEqual(play.view(session, 1));
  },
);

it("saves the same battle as the desktop game, the fixture the desktop resumes", async () => {
  const saved = readFixture();
  if (saved.pending === undefined) throw Error("no pending playback");
  // The left human takes over at the battle's eve, sees the resolution and confirms no orders.
  const session = load(decodePendingBefore(saved.pending)),
    match = new LocalMatch(session, "hotseat", 0);
  match.takeOver();
  const left = match.sides[0];
  while (left.busy()) left.tick(100_000);
  await left.phase();
  expect(match.handover).toBe(1);
  const pending = match.savedPending();
  expect(pending?.side).toBe(1);
  const web = encodeEnvelope(
    "1914-08:2p",
    match.saveSide(),
    play.save(session),
    {
      mode: "hotseat",
      pending,
    },
  );
  // The same game and the same game before as the desktop's save.
  expect(web.side).toBe(saved.side);
  expect(web.data).toBe(saved.data);
  expect(web.pending?.before).toBe(saved.pending.before);
  const text = serializeSave(web) + "\n";
  if (process.env["HL_UPDATE_PENDING"] === "1") writeFileSync(webFile, text);
  expect(readFileSync(webFile, "utf8")).toBe(text);
});
