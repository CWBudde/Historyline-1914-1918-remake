// Serves the production site as GitHub Pages does: below a project path and gzipped where Pages gzips.
import { createReadStream, statSync } from "node:fs";
import { createServer } from "node:http";
import { extname, join, normalize, resolve } from "node:path";
import { createGzip } from "node:zlib";

const root = resolve(import.meta.dirname, "../../../build/webapp/site");
const prefix = process.env["E2E_BASE_PATH"] ?? "/hl/";
const port = Number(process.env["E2E_PORT"] ?? "4173");
const types: Record<string, string> = {
  ".css": "text/css",
  ".html": "text/html; charset=utf-8",
  ".jpg": "image/jpeg",
  ".js": "text/javascript",
  ".json": "application/json",
  ".png": "image/png",
  ".txt": "text/plain",
  ".webp": "image/webp",
  ".woff2": "font/woff2",
};
const compressed = new Set([".css", ".html", ".js", ".json", ".txt"]);

createServer((request, response) => {
  const path = new URL(request.url ?? "/", "http://localhost").pathname;
  if (!path.startsWith(prefix)) {
    response.writeHead(404).end();
    return;
  }
  const relative = normalize(decodeURIComponent(path.slice(prefix.length)));
  let file = join(root, relative === "." ? "index.html" : relative);
  if (!file.startsWith(root)) {
    response.writeHead(403).end();
    return;
  }
  try {
    if (statSync(file).isDirectory()) file = join(file, "index.html");
    statSync(file);
  } catch {
    response.writeHead(404).end();
    return;
  }
  const extension = extname(file);
  const gzip =
    compressed.has(extension) &&
    (request.headers["accept-encoding"] ?? "").includes("gzip");
  response.writeHead(200, {
    "Cache-Control": "max-age=600",
    "Content-Type": types[extension] ?? "application/octet-stream",
    ...(gzip ? { "Content-Encoding": "gzip" } : {}),
  });
  const body = createReadStream(file);
  if (gzip) body.pipe(createGzip()).pipe(response);
  else body.pipe(response);
}).listen(port, "localhost");
