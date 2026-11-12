/**
 * Keeps CSS custom properties on the root element in step with the visual
 * viewport, so the layout follows the on-screen keyboard and the browser bars
 * on phones: `--app-height`, `--app-top` and `--keyboard` (all CSS px).
 * Returns a disposer.
 */
export function installViewport(
  doc: Document,
  win: Window,
  onChange: () => void,
): () => void {
  const root = doc.documentElement;
  const vv = win.visualViewport;

  const update = (): void => {
    const height = vv ? vv.height : win.innerHeight;
    const top = vv ? vv.offsetTop : 0;
    const keyboard = Math.max(0, win.innerHeight - height - top);
    root.style.setProperty("--app-height", `${height}px`);
    root.style.setProperty("--app-top", `${top}px`);
    root.style.setProperty("--keyboard", `${keyboard}px`);
    onChange();
  };

  const listeners: [EventTarget, string][] = [
    [win, "resize"],
    [win, "orientationchange"],
  ];
  if (vv) listeners.push([vv, "resize"], [vv, "scroll"]);
  for (const [target, name] of listeners) target.addEventListener(name, update);
  update();

  return () => {
    for (const [target, name] of listeners)
      target.removeEventListener(name, update);
  };
}
