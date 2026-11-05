import { expect, it } from "vitest";
import { month, parse } from "./campaign.ts";
import { listDir, readJSON } from "./testing/files.ts";

it("reads every shipped newspaper front page", () => {
  for (const dir of listDir("levels").filter((d) => /^19\d\d-\d\d$/.test(d))) {
    const file = `${dir}/news.json`;
    const raw = readJSON(`levels/${file}`);
    expect(parse(raw)).toEqual(raw);
    expect(month(file)).toBe(dir);
  }
});
const article = {
  label: "August 1914",
  headline: "War",
  deck: "News",
  caption: "Europe",
  body: ["Report"],
  briefs: ["Bulletin"],
};
it.each(["label", "headline", "deck", "caption", "body", "briefs"])(
  "requires %s",
  (field) => {
    expect(() =>
      parse({
        ...article,
        [field]: field === "body" || field === "briefs" ? [] : "",
      }),
    ).toThrow();
  },
);
it("refuses unknown fields and malformed paragraphs", () => {
  expect(() => parse({ ...article, extra: true })).toThrow();
  expect(() => parse({ ...article, body: [5] })).toThrow();
});
