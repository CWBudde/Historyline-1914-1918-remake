// The monthly newspaper front pages (internal/campaign).
import { DecodeError, arrayField, object, stringField } from "./decode.ts";

export interface Article {
  label: string;
  headline: string;
  deck: string;
  caption: string;
  body: string[];
  briefs: string[];
}
export function month(file: string): string {
  const dir = file.slice(0, file.lastIndexOf("/") + 1),
    absolute = dir.startsWith("/"),
    parts: string[] = [];
  for (const p of dir.split("/")) {
    if (p === "" || p === ".") continue;
    if (p === ".." && parts.length > 0 && parts.at(-1) !== "..") parts.pop();
    else if (p !== ".." || !absolute) parts.push(p);
  }
  return parts.at(-1) ?? (absolute ? "/" : ".");
}
export function parse(raw: unknown): Article {
  const o = object(raw, "front page", [
    "label",
    "headline",
    "deck",
    "caption",
    "body",
    "briefs",
  ]);
  const paragraphs = (key: string): string[] => {
    const a = arrayField(o, key, "front page");
    if (a === null || a.length === 0)
      throw new DecodeError(`front page: ${key} is required`);
    return a.map((v) => {
      if (typeof v !== "string")
        throw new DecodeError(`front page: ${key} must contain strings`);
      return v;
    });
  };
  const a: Article = {
    label: stringField(o, "label", "front page"),
    headline: stringField(o, "headline", "front page"),
    deck: stringField(o, "deck", "front page"),
    caption: stringField(o, "caption", "front page"),
    body: paragraphs("body"),
    briefs: paragraphs("briefs"),
  };
  if (a.label === "" || a.headline === "" || a.deck === "" || a.caption === "")
    throw new DecodeError(
      "front page: label, headline, deck and caption are required",
    );
  return a;
}
