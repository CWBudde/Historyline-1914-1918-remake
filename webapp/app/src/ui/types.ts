import type { Signal } from "@preact/signals";
import type { MapBattle } from "../mapBattle.ts";
import type { SpriteFamily } from "../unitFrame.ts";
import type { UiAssets } from "./assets.ts";
import type { RingLayout, Action } from "./ring.ts";

export interface BattleUIProps {
  battle: MapBattle;
  assets: UiAssets;
  families: Readonly<Record<string, SpriteFamily>>;
  revision: Signal<number>;
  scale: Signal<number>;
  reduced: Signal<boolean>;
  scenes: Signal<boolean>;
  painted: Signal<boolean>;
  ring: Signal<RingLayout | null>;
  settingsBusy?: boolean;
  settingsNote?: string;
  embedded?: boolean;
  showScene?: boolean;
  readOnly?: boolean;
  layout?: "responsive" | "original";
  onLayout?(value: "responsive" | "original"): void;
  onScale(value: number): void;
  onReduced(value: boolean): void;
  onScenes(value: boolean): void;
  onDesign(): void;
  onLeave(): void;
  onFullscreen(): void;
  onRingAction(action: Action): void;
  onCloseRing(): void;
  onSceneHost(element: HTMLDivElement | null): void;
  onStrategyHost(element: HTMLDivElement | null): void;
}
