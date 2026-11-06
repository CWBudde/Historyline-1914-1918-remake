export {};

// The service worker keeps the app itself. Art and level files stay in AssetStore's
// content cache, which verifies them by hash; this worker never touches blobs/.
interface WorkerScope {
  location: Location;
  caches: CacheStorage;
  addEventListener(
    type: "install" | "activate",
    listener: (event: { waitUntil(work: Promise<unknown>): void }) => void,
  ): void;
  addEventListener(
    type: "fetch",
    listener: (event: {
      request: Request;
      respondWith(response: Promise<Response>): void;
    }) => void,
  ): void;
}
const worker = self as unknown as WorkerScope;
// Written by the build, see precacheManifest in vite.config.ts.
const files = JSON.parse("__HISTORYLINE_PRECACHE__") as string[];
const cacheName = "historyline-shell-__HISTORYLINE_VERSION__";
const page = new URL("./", worker.location.href).href;

worker.addEventListener("install", (event) => {
  event.waitUntil(
    worker.caches.open(cacheName).then((cache) => cache.addAll(files)),
  );
});
worker.addEventListener("activate", (event) => {
  event.waitUntil(
    worker.caches
      .keys()
      .then((names) =>
        Promise.all(
          names
            .filter(
              (n) => n.startsWith("historyline-shell-") && n !== cacheName,
            )
            .map((n) => worker.caches.delete(n)),
        ),
      ),
  );
});
async function fresh(request: Request, fallback: string): Promise<Response> {
  const cache = await worker.caches.open(cacheName);
  try {
    const response = await fetch(request);
    if (response.ok) await cache.put(fallback, response.clone());
    return response;
  } catch (error) {
    const stored = await cache.match(fallback);
    if (stored === undefined) throw error;
    return stored;
  }
}
async function stored(request: Request): Promise<Response> {
  const cache = await worker.caches.open(cacheName);
  return (await cache.match(request)) ?? fetch(request);
}
worker.addEventListener("fetch", (event) => {
  const { request } = event,
    url = new URL(request.url);
  if (
    request.method !== "GET" ||
    url.origin !== worker.location.origin ||
    url.href.startsWith(new URL("blobs/", page).href)
  )
    return;
  if (request.mode === "navigate") event.respondWith(fresh(request, page));
  else if (url.href === new URL("index.json", page).href)
    event.respondWith(fresh(request, url.href));
  else event.respondWith(stored(request));
});
