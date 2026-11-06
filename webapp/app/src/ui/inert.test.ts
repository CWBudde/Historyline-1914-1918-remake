import { expect, it } from "vitest";
import { makeInert } from "./inert.ts";

it("keeps the stage inert until overlapping leave and handover dialogs both release it", () => {
  for (const order of [
    [0, 1],
    [1, 0],
  ] as const) {
    const node = { inert: false } as HTMLElement,
      releases = [makeInert(node), makeInert(node)] as const;
    expect(node.inert).toBe(true);
    releases[order[0]]();
    expect(node.inert).toBe(true);
    releases[order[1]]();
    expect(node.inert).toBe(false);
    releases[order[0]]();
    expect(node.inert).toBe(false);
  }
});
it("preserves an existing inert state and isolates different document controls", () => {
  const a = { inert: true } as HTMLElement,
    b = { inert: false } as HTMLElement,
    releaseA = makeInert(a),
    releaseB = makeInert(b);
  releaseA();
  expect(a.inert).toBe(true);
  expect(b.inert).toBe(true);
  releaseB();
  expect(b.inert).toBe(false);
});
