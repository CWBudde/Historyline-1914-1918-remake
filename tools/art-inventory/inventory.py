#!/usr/bin/env python3
"""Reconcile curated art metadata with static loader, header and level evidence.

No game imports, image libraries, or network access. See README.md for boundaries.
"""
import argparse
import ast
import csv
import hashlib
import io
import json
import struct
from collections import defaultdict
from pathlib import Path

FIELDS = ["path", "kind", "width", "height", "runtime_ids", "loaded_by",
          "used_in_levels", "status", "provenance", "evidence", "notes",
          "loader_scopes", "sha256"]


def scan_source(source, filename):
    """Extract literal pygame image/font loads using actual AST dictionary context."""
    loads = []

    def walk(node, target="", keys=(), scope=()):
        if isinstance(node, (ast.ClassDef, ast.FunctionDef, ast.AsyncFunctionDef)):
            for child in ast.iter_child_nodes(node):
                walk(child, scope=scope + (node.name,))
            return
        if isinstance(node, (ast.Assign, ast.AnnAssign)):
            assigned = node.targets[0] if isinstance(node, ast.Assign) else node.target
            walk(node.value, ast.unparse(assigned), (), scope)
            return
        if isinstance(node, ast.Dict):
            for key, value in zip(node.keys, node.values):
                if not isinstance(key, ast.Constant) or not isinstance(key.value, str):
                    walk(value, target, keys, scope)
                    continue
                walk(value, target, keys + (key.value,), scope)
            return
        if isinstance(node, ast.Call):
            name = ast.unparse(node.func)
            if name in ("pygame.image.load", "pygame.font.Font"):
                if not node.args or not isinstance(node.args[0], ast.Constant) or not isinstance(node.args[0].value, str):
                    raise ValueError(f"dynamic asset path at {filename}:{node.lineno}")
                if target in ("bgTiles", "infraIcons"):
                    runtime_id = keys[0]
                elif target in ("unitsParameters", "weaponsParameters"):
                    runtime_id = ".".join(keys)
                else:
                    runtime_id = target or f"{'.'.join(scope)}:image@{node.lineno}"
                loads.append({"path": node.args[0].value, "runtime_id": runtime_id,
                              "location": f"{filename}:{node.lineno}", "scope": ".".join(scope),
                              "kind": "font" if name.endswith("Font") else "image"})
                return
            if name.endswith(".append"):
                target = name.removesuffix(".append")
        for child in ast.iter_child_nodes(node):
            walk(child, target, keys, scope)

    walk(ast.parse(source, filename=filename))
    return loads


def dimensions(path):
    """Read stored canvas dimensions; these do not certify original-game resolution."""
    data = path.read_bytes()[:64]
    if data.startswith(b"\x89PNG\r\n\x1a\n") and len(data) >= 24:
        return struct.unpack(">II", data[16:24])
    if data.startswith((b"GIF87a", b"GIF89a")) and len(data) >= 10:
        return struct.unpack("<HH", data[6:10])
    if data.startswith(b"gimp xcf ") and b"\0" in data:
        offset = data.index(b"\0") + 1
        if len(data) >= offset + 8:
            return struct.unpack(">II", data[offset:offset + 8])
    raise ValueError(f"unsupported or truncated image header: {path}")


def data_tables(source):
    """Read just the literal unit weapon links and terrain keys, never evaluate code."""
    units = {}
    modifiers = set()
    for node in ast.parse(source).body:
        if not isinstance(node, ast.Assign) or not isinstance(node.targets[0], ast.Name):
            continue
        name = node.targets[0].id
        if name == "bgTilesModifiers":
            modifiers = {key.value for key in node.value.keys}
        if name == "unitsParameters":
            for key, value in zip(node.value.keys, node.value.values):
                if isinstance(value, ast.Dict):
                    fields = {field.value: entry for field, entry in zip(value.keys, value.values)}
                    units[key.value] = {"weapons": ast.literal_eval(fields["weapons"])}
    return units, modifiers


def level_usage(levels, units):
    """Map IDs to levels, including stacked overlays and units inside containers."""
    usage = defaultdict(set)
    for path in sorted(levels.glob("*.json")):
        level = json.loads(path.read_text())
        for line in level["tiles"].values():
            for tile in line:
                usage[tile[0]].add(path.name)
                infrastructure = tile[1] if isinstance(tile[1], list) else [tile[1]]
                for runtime_id in infrastructure:
                    if runtime_id:
                        usage[runtime_id].add(path.name)
                unit_ids = [tile[2]] if tile[2] else []
                if len(tile) > 3 and len(tile[3]) > 1:
                    unit_ids.extend(tile[3][1])
                for runtime_id in unit_ids:
                    usage[runtime_id].add(path.name)
                    for weapon in units.get(runtime_id, {}).get("weapons", []):
                        if weapon:
                            usage[weapon].add(path.name)
    return usage


def asset_kind(path):
    if path.endswith(".xcf"):
        return "layered-source"
    prefixes = {"gfx/hexTypes/": "terrain", "gfx/infrastructure/": "infrastructure",
                "gfx/units/icons/": "unit-icons", "gfx/units/pictures/": "unit-pictures",
                "gfx/weapons/": "weapon", "gfx/menuIcons/": "ui-icon"}
    return next((kind for prefix, kind in prefixes.items() if path.startswith(prefix)), "ui")


def render_inventory(root, curated, levels=None):
    """root is the Python reference project (sources and gfx/); levels defaults to root/levels."""
    loads = []
    # Deliberately exclude private original archives and art tooling from the runtime scan.
    # The tree has no other Python; tools and testdata generators live outside the root.
    sources = sorted(root.glob("*.py"))
    for path in sources:
        loads.extend(scan_source(path.read_text(), path.relative_to(root).as_posix()))
    source = root / "hlrData.py"
    units, modifiers = data_tables(source.read_text()) if source.exists() else ({}, set())
    usage = level_usage(levels or root / "levels", units)
    by_path = defaultdict(list)
    for load in loads:
        by_path[load["path"]].append(load)
    metadata = {row["path"]: row for row in curated if row["path"]}
    paths = {path.relative_to(root).as_posix() for path in (root / "gfx").rglob("*") if path.is_file()}
    paths.update(by_path)
    rows = []
    for relative in sorted(paths):
        path = root / relative
        row = dict(metadata.get(relative, {}))
        row.update(path=relative, kind=row.get("kind") or asset_kind(relative))
        asset_loads = by_path[relative]
        ids = sorted({load["runtime_id"] for load in asset_loads})
        row["runtime_ids"] = " ".join(ids)
        row["loaded_by"] = " ".join(sorted({load["location"] for load in asset_loads}))
        row["loader_scopes"] = " ".join(sorted({load["scope"] for load in asset_loads if load["scope"]}))
        levels = set()
        for load in asset_loads:
            levels.update(usage.get(load["runtime_id"].split(".")[0], set()))
        # These class assets depend on the corresponding container terrain, not a unit table.
        container_tiles = {"hq.png": ("hqN", "hqN_w"),
                           "storage.png": ("cmpN", "cmpN_w"),
                           "factory.png": ("factoryE", "factoryE_w")}
        if relative.startswith("gfx/units/pictures/") and path.name in container_tiles:
            for runtime_id in container_tiles[path.name]:
                levels.update(usage.get(runtime_id, set()))
        row["used_in_levels"] = " ".join(sorted(levels))
        row["width"], row["height"] = dimensions(path) if path.exists() and path.suffix.lower() in (".png", ".gif", ".xcf") else ("", "")
        row["sha256"] = hashlib.sha256(path.read_bytes()).hexdigest() if path.exists() else ""
        row["provenance"] = row.get("provenance") or "unknown"
        row["evidence"] = row.get("evidence") or "no source/provenance evidence recorded"
        # Derived findings must disappear when their underlying condition changes.
        notes = [note for note in row.get("notes", "").split("; ") if note and
                 not note.startswith(("runtime id differs from filename:", "loaded but not placed in", "runtime ID maps to filename:"))]
        alias_notes = []
        if row["kind"] in ("terrain", "infrastructure") and asset_loads and not any(load["scope"] for load in asset_loads):
            stem = path.stem.removeprefix("hex_").replace("_winter", "_w")
            alias_notes = [f"runtime ID maps to filename: {runtime_id} -> {path.name}" for runtime_id in ids if runtime_id != stem and not runtime_id.startswith("self.")]
        if not path.exists():
            row["status"] = "external-dependency" if relative == "freesansbold.ttf" else "missing-file"
            row["kind"] = "font" if relative == "freesansbold.ttf" else row["kind"]
        elif asset_loads:
            row["status"] = "alias" if alias_notes else "loaded"
        elif row.get("status") not in ("source", "experimental", "reference-only"):
            row["status"] = "unused"
        notes.extend(alias_notes)
        if asset_loads and row["kind"] in ("terrain", "infrastructure") and not levels:
            notes.append("loaded but not placed in any shipped level (including backup)")
        row["notes"] = "; ".join(dict.fromkeys(notes))
        rows.append(row)

    # Keep curated screen requirements and planned content gaps; verify their source evidence
    # in ui-inventory.md. Missing modifier definitions are derived rather than hand-counted.
    tile_ids = {load["runtime_id"] for load in loads if load["path"].startswith("gfx/hexTypes/")}
    synthetic = {row["runtime_ids"]: dict(row) for row in curated if not row["path"] and
                 not (row["kind"] == "terrain" and row["runtime_ids"] in tile_ids)}
    for runtime_id in sorted(modifiers - tile_ids):
        synthetic.setdefault(runtime_id, dict(path="", kind="terrain", runtime_ids=runtime_id,
                                             status="missing", provenance="n/a", evidence="hlrData.py bgTilesModifiers",
                                             notes="terrain modifier defined but no tile image loaded"))
    for runtime_id, row in sorted(synthetic.items()):
        row["used_in_levels"] = " ".join(sorted(usage.get(runtime_id, set())))
        rows.append(row)
    output = io.StringIO(newline="")
    writer = csv.DictWriter(output, fieldnames=FIELDS, lineterminator="\n", extrasaction="raise")
    writer.writeheader()
    writer.writerows(rows)
    return output.getvalue(), len(loads), len(paths)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[2])
    parser.add_argument("--python-root", type=Path, help="Python reference project; default <root>/legacy/python")
    parser.add_argument("--check", action="store_true", help="report drift without writing")
    args = parser.parse_args()
    inventory = args.root / "docs/art/asset-inventory.csv"
    with inventory.open(newline="") as handle:
        curated = list(csv.DictReader(handle))
    generated, loads, paths = render_inventory(args.python_root or args.root / "legacy/python", curated, args.root / "levels")
    if args.check:
        if generated != inventory.read_text():
            print("asset inventory is stale; regenerate with tools/art-inventory/inventory.py")
            return 1
        print(f"asset inventory current: {paths} asset/dependency paths, {loads} loader calls")
        return 0
    inventory.write_text(generated)
    print(f"wrote {inventory}: {paths} asset/dependency paths, {loads} loader calls")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
