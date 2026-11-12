import { describe, expect, it } from "vitest";
import { installViewport } from "./viewport.ts";

class Target {
  readonly handlers = new Map<string, Set<() => void>>();
  addEventListener(name: string, fn: () => void): void {
    const set = this.handlers.get(name) ?? new Set<() => void>();
    set.add(fn);
    this.handlers.set(name, set);
  }
  removeEventListener(name: string, fn: () => void): void {
    this.handlers.get(name)?.delete(fn);
  }
  emit(name: string): void {
    for (const fn of this.handlers.get(name) ?? []) fn();
  }
  count(): number {
    let n = 0;
    for (const set of this.handlers.values()) n += set.size;
    return n;
  }
}

function setup(withVisual: boolean) {
  const props: Record<string, string> = {};
  const doc = {
    documentElement: {
      style: { setProperty: (k: string, v: string) => (props[k] = v) },
    },
  } as unknown as Document;
  const vv = Object.assign(new Target(), { height: 800, offsetTop: 0 });
  const win = Object.assign(new Target(), {
    innerHeight: 800,
    visualViewport: withVisual ? vv : null,
  });
  return { props, doc, vv, win: win as unknown as Window, winT: win };
}

describe("installViewport", () => {
  it("sets the properties on install and calls onChange", () => {
    const { props, doc, win } = setup(true);
    let calls = 0;
    installViewport(doc, win, () => calls++);
    expect(props).toEqual({
      "--app-height": "800px",
      "--app-top": "0px",
      "--keyboard": "0px",
    });
    expect(calls).toBe(1);
  });

  it("follows the visual viewport when the keyboard opens", () => {
    const { props, doc, win, vv } = setup(true);
    let calls = 0;
    installViewport(doc, win, () => calls++);
    vv.height = 500;
    vv.offsetTop = 20;
    vv.emit("resize");
    expect(props["--app-height"]).toBe("500px");
    expect(props["--app-top"]).toBe("20px");
    expect(props["--keyboard"]).toBe("280px");
    vv.height = 900;
    vv.offsetTop = 0;
    vv.emit("scroll");
    expect(props["--keyboard"]).toBe("0px");
    expect(calls).toBe(3);
  });

  it("falls back to innerHeight without a visual viewport", () => {
    const { props, doc, win, winT } = setup(false);
    installViewport(doc, win, () => {});
    expect(props["--app-height"]).toBe("800px");
    winT.innerHeight = 600;
    winT.emit("orientationchange");
    expect(props["--app-height"]).toBe("600px");
    expect(props["--keyboard"]).toBe("0px");
  });

  it("removes its listeners on dispose", () => {
    const { doc, win, vv, winT } = setup(true);
    const dispose = installViewport(doc, win, () => {});
    expect(vv.count()).toBe(2);
    expect(winT.count()).toBe(2);
    dispose();
    expect(vv.count()).toBe(0);
    expect(winT.count()).toBe(0);
  });
});
