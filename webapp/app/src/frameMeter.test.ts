import { expect, it } from "vitest";
import { FrameMeter, frameMeasure } from "./frameMeter.ts";

it("reports one complete CPU frame and clears retained timing entries", () => {
  let time = 2;
  const measures: unknown[] = [],
    cleared: string[] = [];
  const meter = new FrameMeter({
    now: () => time,
    measure: (name, span) => {
      measures.push({ name, ...span });
    },
    clearMeasures: (name) => {
      cleared.push(name);
    },
  });
  meter.end();
  expect(measures).toEqual([]);
  meter.begin();
  time = 9;
  meter.end();
  meter.end();
  expect(measures).toEqual([{ name: frameMeasure, start: 2, end: 9 }]);
  expect(cleared).toEqual([frameMeasure]);
  time = 20;
  meter.begin();
  time = 23;
  meter.end();
  expect(measures).toHaveLength(2);
});
