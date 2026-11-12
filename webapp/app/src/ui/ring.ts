/**
 * The orders and the desktop's four key gestures (internal/gesture ForMap, ForContents): what holding fire and
 * pushing a direction runs. The pointer's order menu round a unit is orderMenu.ts.
 */
export type Action =
  | "cancel"
  | "move"
  | "attack"
  | "mode"
  | "open"
  | "map"
  | "info"
  | "exit"
  | "produce"
  | "repair";
export type Direction = "centre" | "up" | "left" | "right" | "down";
export type Placement = Record<Direction, Action>;
export function placement(
  attack: boolean,
  ownContainer: boolean,
  contents = false,
): Placement {
  return contents
    ? {
        centre: "cancel",
        up: "exit",
        left: "produce",
        right: "cancel",
        down: "repair",
      }
    : {
        centre: "cancel",
        up: attack ? "attack" : "move",
        left: ownContainer ? "open" : "mode",
        right: "map",
        down: "info",
      };
}
export function ringActions(
  p: Placement,
  allowed: (a: Action) => boolean,
): Action[] {
  return [p.up, p.right, p.down, p.left].filter(
    (a) => a !== "cancel" && allowed(a),
  );
}
