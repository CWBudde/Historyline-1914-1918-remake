// Reading the files the player picks or drops, in the browser: only the names the import can use are read.
import type { Picked } from "./files.ts";

const wanted =
  /(^|\/)([0-9A-Z_]{1,8}\.(FIN|SHP|COM)|CODES\.DAT|00\.PAL|(PARTS|PARTW|UNIT)\.(DAT|LIB))$/i;

async function read(path: string, file: File): Promise<Picked> {
  return { path, bytes: new Uint8Array(await file.arrayBuffer()) };
}

/** fromFileList reads a file input's files, with their folder paths where the input picked a folder. */
export async function fromFileList(
  list: FileList | readonly File[],
): Promise<Picked[]> {
  const files = Array.from(list).map((f) => ({
    path: f.webkitRelativePath || f.name,
    file: f,
  }));
  return Promise.all(
    files.filter((f) => wanted.test(f.path)).map((f) => read(f.path, f.file)),
  );
}

function entries(dir: FileSystemDirectoryEntry): Promise<FileSystemEntry[]> {
  const reader = dir.createReader(),
    out: FileSystemEntry[] = [];
  return new Promise((resolve, reject) => {
    const next = () => {
      reader.readEntries((batch) => {
        if (batch.length === 0) resolve(out);
        else {
          out.push(...batch);
          next();
        }
      }, reject);
    };
    next();
  });
}
async function walk(
  entry: FileSystemEntry,
  out: Picked[],
  depth = 0,
): Promise<void> {
  if (entry.isFile) {
    if (!wanted.test(entry.fullPath)) return;
    const file = await new Promise<File>((resolve, reject) => {
      (entry as FileSystemFileEntry).file(resolve, reject);
    });
    out.push(await read(entry.fullPath, file));
  } else if (entry.isDirectory && depth < 6)
    for (const child of await entries(entry as FileSystemDirectoryEntry))
      await walk(child, out, depth + 1);
}

/** fromDrop reads dropped files and folders. */
export async function fromDrop(data: DataTransfer): Promise<Picked[]> {
  const roots = Array.from(data.items)
    .map((item) => item.webkitGetAsEntry())
    .filter((e): e is FileSystemEntry => e !== null);
  if (roots.length === 0) return fromFileList(data.files);
  const out: Picked[] = [];
  for (const root of roots) await walk(root, out);
  return out;
}
