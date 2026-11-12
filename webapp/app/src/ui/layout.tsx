import type { Layout } from "../storage/store.ts";

/** The layout choices of the settings, in the order they are offered. */
export const layoutNames: readonly [Layout, string][] = [
  ["auto", "Automatic (split screen on wide windows)"],
  ["single", "Single map"],
  ["original", "Original split screen"],
];
/** layoutValue reads a select's value as a layout, null for anything else. */
export function layoutValue(value: string): Layout | null {
  return layoutNames.find(([layout]) => layout === value)?.[0] ?? null;
}
export function LayoutOptions() {
  return (
    <>
      {layoutNames.map(([value, name]) => (
        <option key={value} value={value}>
          {name}
        </option>
      ))}
    </>
  );
}
