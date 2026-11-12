import { signal, type Signal } from "@preact/signals";
import { assetDensity, type AssetLoader } from "../assets.ts";

export interface UiImage {
  url: string | null;
  quality: "lite" | "full" | null;
  error: string | null;
}
const empty = (): UiImage => ({ url: null, quality: null, error: null });

/** The 640×400 logical size of a menu painting. */
const painting = { width: 640, height: 400 };
/**
 * paintingScale is the scale a menu painting is drawn at when it covers a window of CSS pixels: its art is chosen
 * for that size, as the desktop game draws its pages at density 2 once they are enlarged.
 */
export function paintingScale(
  width: number,
  height: number,
  uiScale: number,
): number {
  return Math.max(uiScale, width / painting.width, height / painting.height);
}

/** DOM images use the same verified progressive catalogs as the map. Object
 * URLs stay valid through a UI repaint and are released with the battle. */
export class UiAssets {
  private readonly images = new Map<string, Signal<UiImage>>();
  private readonly urls = new Set<string>();
  private readonly faces = new Set<FontFace>();
  private stopped = false;
  private readonly abort = new AbortController();
  constructor(
    private readonly loader: Pick<AssetLoader, "progressive">,
    private readonly dpr: () => number = () => window.devicePixelRatio,
  ) {}
  /** sharp brings lite art at the drawn density too, for paintings that fill the window. */
  image(
    catalog: string,
    id: string,
    scale = 1,
    sharp = false,
  ): Signal<UiImage> {
    const density = assetDensity(scale, this.dpr()),
      key = `${catalog}/${id}/${density}${sharp ? "/sharp" : ""}`;
    let state = this.images.get(key);
    if (state !== undefined) return state;
    state = signal<UiImage>(empty());
    if (this.stopped) return state;
    this.images.set(key, state);
    const target = state;
    void this.loader
      .progressive(
        catalog,
        [id],
        density,
        1,
        (image) => {
          if (this.stopped) return;
          const url = URL.createObjectURL(new Blob([image.bytes]));
          this.urls.add(url);
          target.value = { url, quality: image.quality, error: null };
        },
        this.abort.signal,
        sharp,
      )
      .catch((error: unknown) => {
        if (!this.stopped)
          target.value = {
            ...target.value,
            error: error instanceof Error ? error.message : String(error),
          };
      });
    return state;
  }
  async fonts(loader: AssetLoader): Promise<void> {
    const names: Record<
      string,
      { family: string; style?: string; weight?: string }
    > = {
      "NotoSans-Regular": { family: "HL Sans" },
      "NotoSans-SemiBold": { family: "HL Sans", weight: "600" },
      "NotoSerif-Regular": { family: "HL Serif" },
      "NotoSerif-Italic": { family: "HL Serif", style: "italic" },
      "NotoSerifDisplay-Bold": { family: "HL Display", weight: "700" },
    };
    await Promise.all(
      Object.entries(names).map(async ([name, font]) => {
        const ref = loader.index.fonts[name];
        if (ref === undefined) throw new Error(`missing font ${name}`);
        const face = new FontFace(font.family, await loader.store.get(ref), {
          ...font,
          display: "swap",
        });
        await face.load();
        if (!this.stopped) {
          document.fonts.add(face);
          this.faces.add(face);
        }
      }),
    );
  }
  dispose(): void {
    this.stopped = true;
    this.abort.abort();
    for (const url of this.urls) URL.revokeObjectURL(url);
    for (const face of this.faces) document.fonts.delete(face);
    for (const image of this.images.values()) image.value = empty();
    this.urls.clear();
    this.faces.clear();
    this.images.clear();
  }
}
