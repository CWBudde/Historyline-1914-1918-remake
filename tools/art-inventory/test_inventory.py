"""Regression tests for asset discovery, without importing the game or pygame."""
import csv
import importlib.util
import struct
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

MODULE = Path(__file__).with_name("inventory.py")


class InventoryTests(unittest.TestCase):
    def load_module(self):
        self.assertTrue(MODULE.exists(), "the inventory generator is not implemented")
        spec = importlib.util.spec_from_file_location("art_inventory", MODULE)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        return module

    def test_loader_context_keeps_weapon_ids_and_shared_unit_ids(self):
        # A nearest-text-key scan incorrectly gives the weapon EC_sappers.picture.
        module = self.load_module()
        source = '''
import pygame
unitsParameters = {
 "CP_sappers": {"icon": pygame.image.load("gfx/shared.png")},
 "EC_sappers": {"icon": pygame.image.load("gfx/shared.png")}}
weaponsParameters = {"bayonet": {"picture": pygame.image.load("gfx/bayonet.png")}}
# pygame.image.load("gfx/commented.png")
class GUI:
 def __init__(self):
  self.cursorGfx = pygame.image.load("gfx/cursor.png")
  self.buttons = [pygame.image.load("gfx/normal.png"), pygame.image.load("gfx/focus.png")]
  self.font = pygame.font.Font("freesansbold.ttf", 32)
'''
        loads = module.scan_source(source, "sample.py")
        mapping = [(load["path"], load["runtime_id"]) for load in loads]
        self.assertEqual(mapping, [
            ("gfx/shared.png", "CP_sappers.icon"),
            ("gfx/shared.png", "EC_sappers.icon"),
            ("gfx/bayonet.png", "bayonet.picture"),
            ("gfx/cursor.png", "self.cursorGfx"),
            ("gfx/normal.png", "self.buttons"),
            ("gfx/focus.png", "self.buttons"),
            ("freesansbold.ttf", "self.font"),
        ])
        self.assertEqual(loads[3]["scope"], "GUI.__init__")

    def test_dynamic_image_path_is_reported_instead_of_silently_omitted(self):
        module = self.load_module()
        with self.assertRaisesRegex(ValueError, "dynamic.*sample.py:1"):
            module.scan_source('pygame.image.load("gfx/" + filename)', "sample.py")

    def test_unrelated_numeric_dictionary_does_not_prevent_loader_scan(self):
        module = self.load_module()
        source = 'labels = {1: "Germany", 2: "France"}\nicon = pygame.image.load("gfx/icon.png")'
        self.assertEqual(module.scan_source(source, "main.py")[0]["path"], "gfx/icon.png")

    def test_level_usage_includes_stacked_infra_container_units_and_weapons(self):
        # Omitting nested container entries loses the only use of heavy artillery.
        module = self.load_module()
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "levels").mkdir()
            (root / "levels" / "level1_cp.json").write_text('''{
              "tiles": {"line1": [
                ["cmpN", ["stream25", "rBridge36"], "CP_infantry", [1, ["CP_hArtillery"]]],
                ["grass", "road14", ""]]}}''')
            units = {"CP_infantry": {"weapons": ["bayonet", None]},
                     "CP_hArtillery": {"weapons": ["morser", None]}}
            usage = module.level_usage(root / "levels", units)
            for key in ("cmpN", "stream25", "rBridge36", "CP_infantry",
                        "CP_hArtillery", "bayonet", "morser", "road14"):
                self.assertEqual(usage[key], {"level1_cp.json"}, key)

    def test_dimensions_use_file_headers_for_png_gif_and_xcf(self):
        module = self.load_module()
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            fixtures = [
                ("tile.png", b"\x89PNG\r\n\x1a\n" + b"\0" * 8 + struct.pack(">II", 96, 80), (96, 80)),
                ("effect.gif", b"GIF89a" + struct.pack("<HH", 192, 96), (192, 96)),
                ("master.xcf", b"gimp xcf v011\0" + struct.pack(">II", 259, 300), (259, 300)),
            ]
            for name, content, expected in fixtures:
                path = root / name
                path.write_bytes(content)
                self.assertEqual(module.dimensions(path), expected)
            bad = root / "broken.png"
            bad.write_bytes(b"not an image")
            with self.assertRaises(ValueError):
                module.dimensions(bad)

    def test_check_detects_stale_csv_without_writing_and_regeneration_is_stable(self):
        module = self.load_module()
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "gfx").mkdir()
            (root / "gfx" / "grass.png").write_bytes(
                b"\x89PNG\r\n\x1a\n" + b"\0" * 8 + struct.pack(">II", 96, 80))
            (root / "hlrData.py").write_text(
                'bgTiles = {"grass": pygame.image.load("gfx/grass.png")}\n')
            inventory = root / "docs" / "art" / "asset-inventory.csv"
            inventory.parent.mkdir(parents=True)
            with inventory.open("w", newline="") as handle:
                writer = csv.DictWriter(handle, fieldnames=module.FIELDS)
                writer.writeheader()
                writer.writerow({"path": "gfx/grass.png", "kind": "terrain", "width": "1",
                                 "height": "1", "provenance": "unknown", "evidence": "repository history"})
            before = inventory.read_bytes()
            args = [sys.executable, str(MODULE), "--root", str(root), "--python-root", str(root)]
            check = subprocess.run(args + ["--check"], capture_output=True, text=True)
            self.assertEqual(check.returncode, 1, check.stderr)
            self.assertEqual(inventory.read_bytes(), before)
            regenerate = subprocess.run(args, capture_output=True, text=True)
            self.assertEqual(regenerate.returncode, 0, regenerate.stderr)
            first = inventory.read_bytes()
            regenerate = subprocess.run(args, capture_output=True, text=True)
            self.assertEqual(regenerate.returncode, 0, regenerate.stderr)
            self.assertEqual(inventory.read_bytes(), first)
            check = subprocess.run(args + ["--check"], capture_output=True, text=True)
            self.assertEqual(check.returncode, 0, check.stderr)
            row = next(csv.DictReader(inventory.read_text().splitlines()))
            self.assertEqual((row["runtime_ids"], row["width"], row["height"]), ("grass", "96", "80"))

    def test_local_overlay_variable_is_not_classified_as_a_level_id_alias(self):
        module = self.load_module()
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            path = root / "gfx" / "hexTypes" / "depotOwnership.png"
            path.parent.mkdir(parents=True)
            path.write_bytes(b"\x89PNG\r\n\x1a\n" + b"\0" * 8 + struct.pack(">II", 35, 18))
            (root / "gui.py").write_text(
                'def recolour():\n colImage = pygame.image.load("gfx/hexTypes/depotOwnership.png")\n')
            content, _, _ = module.render_inventory(root, [])
            row = next(csv.DictReader(content.splitlines()))
            self.assertEqual(row["status"], "loaded")
            self.assertEqual(row["runtime_ids"], "colImage")


if __name__ == "__main__":
    unittest.main()
