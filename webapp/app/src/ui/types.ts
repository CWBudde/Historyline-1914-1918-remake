import type { Design } from "../original/design.ts";
import type { Signal } from "@preact/signals";
import type { MapBattle } from "../mapBattle.ts";
import type { SpriteFamily } from "../unitFrame.ts";
import type { UiAssets } from "./assets.ts";
import type { Hit } from "./orderMenu.ts";
import type { MenuView } from "../battle/runtime.ts";
import type { Layout } from "../storage/store.ts";

export interface BattleUIProps {
  battle: MapBattle;
  assets: UiAssets;
  families: Readonly<Record<string, SpriteFamily>>;
  revision: Signal<number>;
  scale: Signal<number>;
  reduced: Signal<boolean>;
  scenes: Signal<boolean>;
  painted: Signal<boolean>;
  /** The map design drawn now; F8 and onDesign cycle it. */
  design?: Signal<Design>;
  /** The order menu round a unit (ui/orderMenu.ts), null while closed. */
  menu: Signal<MenuView | null>;
  /** Whether the battle's menu board (settings, surrender, leave) is open. */
  menuOpen: Signal<boolean>;
  /** The level's best score, for the situation, where it is known. */
  best?: number;
  settingsBusy?: boolean;
  settingsNote?: string;
  embedded?: boolean;
  showScene?: boolean;
  readOnly?: boolean;
  /** The layout setting (auto, single or original). */
  layout?: Layout;
  onLayout?(value: Layout): void;
  onScale(value: number): void;
  onReduced(value: boolean): void;
  onScenes(value: boolean): void;
  onDesign(): void;
  onLeave(): void;
  onFullscreen(): void;
  /** onMenu runs a menu slot (a plate or the centre's cancel). */
  onMenu(hit: Hit): void;
  /** onMenuFocus lights the plate that has the keyboard focus (null: none). */
  onMenuFocus(hit: Hit): void;
  /** The order strip's global orders: end phase (F1), the strategic map (O), information (I). */
  onPhase(): void;
  onMap(): void;
  onInfo(): void;

  onSceneHost(element: HTMLDivElement | null): void;
  onStrategyHost(element: HTMLDivElement | null): void;
}
