import { expect, it } from "vitest";
import { play, hexgrid } from "@historyline/engine";
import { Full } from "../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../engine/src/testing/play.ts";
import { unitFrames } from "./unitFrame.ts";

const sprites = {
  CP_infantry: { family: "infantry" },
  EC_infantry: { family: "infantry" },
  CP_destroyer: { family: "destroyer", long: "destroyer-long" },
};
function base(): play.UnitView {
  const s = play.newSession(shippedLevel("1914-08"), playData, {
    profile: Full,
  });
  const u = [...play.view(s, 0).units.values()][0];
  if (u === undefined) throw new Error("missing fixture unit");
  return u;
}

it("uses owner palettes, weakened families and own-side spent art", () => {
  const v: play.View = {
    side: 0,
    buildings: new Map(),
    units: new Map<string, play.UnitView>([
      [
        "0,0",
        {
          ...base(),
          index: 1,
          id: "CP_infantry",
          size: 4,
          side: 0,
          owner: 2,
          moved: true,
          at: { line: 0, col: 0 },
          head: null,
          rear: null,
        },
      ],
      [
        "0,1",
        {
          ...base(),
          index: 2,
          id: "EC_infantry",
          size: 6,
          side: 1,
          owner: 1,
          moved: true,
          at: { line: 0, col: 1 },
          head: null,
          rear: null,
        },
      ],
    ]),
  };
  const out = unitFrames(
    v,
    sprites,
    new Map([
      [1, "ne"],
      [2, "sw"],
    ]),
  );
  expect(out.map((u) => u.id)).toEqual([
    "unit/infantry-4/ec/ne-spent",
    "unit/infantry/cp/sw-normal",
  ]);
});

it("draws a linked ship once at its midpoint facing rear to head", () => {
  const head = { line: 0, col: 0 },
    rear = { line: 1, col: 0 },
    u = {
      ...base(),
      index: 1,
      id: "CP_destroyer",
      side: 0,
      owner: 1,
      at: head,
      head: null,
      rear,
    };
  const view: play.View = {
    side: 0,
    buildings: new Map(),
    units: new Map<string, play.UnitView>([
      ["0,0", u],
      ["1,0", { ...u, index: 2, at: rear, head, rear: null }],
    ]),
  };
  const out = unitFrames(view, sprites, new Map());
  expect(out).toHaveLength(1);
  expect(out[0]?.id).toBe("unit/destroyer-long/cp/nw-normal");
  const [hx, hy] = hexgrid.center(head),
    [rx, ry] = hexgrid.center(rear);
  expect(out[0]?.at).toEqual({ x: (hx + rx) / 2, y: (hy + ry) / 2 });
  expect(out[0]?.drawAt).toEqual(rear);
});

it("keeps the other side's moved position concealed until phase resolution", () => {
  const s = play.newSession(shippedLevel("1914-08"), playData, {
    profile: Full,
  });
  play.confirm(s, 0);
  play.confirm(s, 1);
  play.resolve(s);
  const before = play.view(s, 0);
  let moving: play.UnitView | undefined;
  let to: hexgrid.Coord | undefined;
  for (const u of play.view(s, 1).units.values()) {
    if (u.side !== 1) continue;
    const dest = [...play.destinations(s, 1, u.index)]
      .map(play.coordFromKey)
      .find((c) => c.line !== u.at.line || c.col !== u.at.col);
    if (dest !== undefined) {
      moving = u;
      to = dest;
      break;
    }
  }
  if (moving === undefined || to === undefined)
    throw new Error("fixture has no enemy move");
  play.move(s, 1, moving.index, to);
  const after = play.view(s, 0);
  const frames = unitFrames(
    after,
    Object.fromEntries(
      [...playData.units].map(([id]) => [id, { family: "fixture" }]),
    ),
    new Map(),
  );
  const frame = frames.find((f) => f.unit.index === moving.index);
  const old = [...before.units.values()].find((u) => u.index === moving.index);
  if (old === undefined) throw new Error("missing remembered enemy");
  const [x, y] = hexgrid.center(old.at);
  expect(frame?.at).toEqual({ x, y });
});
