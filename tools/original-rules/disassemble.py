#!/usr/bin/env python3
"""Reproduce the local DOS rules investigation; original bytes stay in build/."""

import argparse
import csv
import hashlib
import json
from pathlib import Path
import re
import shutil
import struct
import subprocess


KNOWN_HL = "fe4fa44699ef36df0f1152786557e790651f44b836f9f6bffd3ca5dcc38a9b92"
MAX_UNPACKED = 32 * 1024 * 1024
# Segment, first instruction, exclusive end. These are manually traced windows,
# not an automatic partition into functions. Some contain only part of a function.
WINDOWS = {
    "combat": (0x2466, 0x000E, 0x0ABC),
    "combat_wrapper": (0x2453, 0x0009, 0x013C),
    "experience": (0x02DC, 0x0908, 0x0ABF),
    "pincer_and_blocking": (0x02DC, 0x262A, 0x29C9),
    "score": (0x02DC, 0x29C9, 0x2B1D),
    "repair": (0x0660, 0x3108, 0x3268),
    "unit_constructor": (0x15A4, 0x0009, 0x01AB),
    "unit_information": (0x272C, 0x0650, 0x0A0F),
    "hex_distance": (0x0E47, 0x2414, 0x24DB),
    "long_arithmetic": (0x0000, 0x03A1, 0x04CB),
}


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def unpack(data):
    """DOS TPWM: LE size, MSB first flags, overlapping 12-bit backreferences."""
    if not data.startswith(b"TPWM"):
        return data
    if len(data) < 8:
        raise ValueError("truncated TPWM header")
    size = int.from_bytes(data[4:8], "little")
    if not 0 < size <= MAX_UNPACKED:
        raise ValueError("invalid DOS TPWM size")
    out = bytearray()
    pos = 8
    while len(out) < size:
        if pos >= len(data):
            raise ValueError("truncated TPWM flag")
        flags = data[pos]
        pos += 1
        for bit in range(7, -1, -1):
            if len(out) == size:
                break
            if flags & (1 << bit):
                if pos + 2 > len(data):
                    raise ValueError("truncated TPWM reference")
                a, b = data[pos:pos + 2]
                pos += 2
                distance = ((a & 0xF0) << 4) | b
                if not 0 < distance <= len(out):
                    raise ValueError("invalid TPWM distance")
                for _ in range(min((a & 15) + 3, size - len(out))):
                    out.append(out[-distance])
            else:
                if pos >= len(data):
                    raise ValueError("truncated TPWM literal")
                out.append(data[pos])
                pos += 1
    if pos != len(data):
        raise ValueError("trailing bytes after DOS TPWM stream")
    return bytes(out)


def parse_mz(data):
    if len(data) < 28 or data[:2] != b"MZ":
        raise ValueError("not an MZ executable")
    values = struct.unpack_from("<14H", data)
    names = ("magic", "last_page_bytes", "pages", "relocation_count", "header_paragraphs",
             "min_extra_paragraphs", "max_extra_paragraphs", "ss", "sp", "checksum",
             "ip", "cs", "relocation_table_offset", "overlay")
    header = dict(zip(names, values))
    header_size = header["header_paragraphs"] * 16
    pages, last = header["pages"], header["last_page_bytes"]
    if pages == 0 or last > 511:
        raise ValueError("invalid MZ page count")
    declared = (pages - 1) * 512 + (last or 512)
    if not 28 <= header_size <= declared <= len(data):
        raise ValueError("invalid or truncated MZ image")
    offset = header["relocation_table_offset"]
    count = header["relocation_count"]
    if offset < 28 or offset + count * 4 > header_size:
        raise ValueError("invalid MZ relocation table")
    image = data[header_size:declared]
    relocations = []
    for i in range(count):
        off, seg = struct.unpack_from("<HH", data, offset + i * 4)
        linear = seg * 16 + off
        if linear + 2 > len(image):
            raise ValueError("relocation outside MZ image")
        target = struct.unpack_from("<H", image, linear)[0]
        relocations.append({"segment": seg, "offset": off, "image_offset": linear,
                            "segment_value": target})
    header["header_bytes"] = header_size
    header["image_bytes"] = len(image)
    header["trailing_file_bytes"] = len(data) - declared
    return header, image, relocations


def write_csv(path, rows):
    with path.open("w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=list(rows[0]))
        writer.writeheader()
        writer.writerows(rows)


def units(data):
    if len(data) != 51 * 69:
        raise ValueError("expected 51 DOS unit records of 69 bytes")
    rows = []
    for index in range(51):
        r = data[index * 69:(index + 1) * 69]
        rows.append({
            "index": index, "offset": index * 69,
            "category": r[27:44].split(b"\0")[0].decode("ascii"),
            # The long label occupies bytes 44..60; 61..62 are runtime IDs.
            "name": r[44:61].split(b"\0")[0].decode("ascii"),
            "movement_budget_raw": r[0], "displayed_speed": r[1],
            "armour": r[2], "max_group": r[3], "alternate_strength_raw": r[4],
            "movement_mask_raw": r[5], "target_flags_raw": r[6] | r[7] << 8,
            "air_range_encoded": r[8], "surface_range_encoded": r[9],
            "air_power": r[10], "ground_power": r[11], "water_power": r[12],
            "flags_13_raw": int.from_bytes(r[13:15], "little"),
            "flags_15_raw": int.from_bytes(r[15:17], "little"),
            "flags_17_raw": int.from_bytes(r[17:19], "little"),
            "cost_raw": r[63], "capacity_raw": r[64], "weight_raw": r[65],
            "raw_hex": r.hex(),
        })
    return rows


def terrain(data):
    if len(data) != 175 * 6:
        raise ValueError("expected 175 DOS terrain records of six bytes")
    return [{"index": i, "offset": i * 6,
             "flags_raw": int.from_bytes(data[i*6:i*6+2], "little"),
             "movement_mask_raw": data[i*6+2], "movement_cost_raw": data[i*6+3],
             "byte_4_raw": data[i*6+4], "combat_class": data[i*6+5],
             "raw_hex": data[i*6:i*6+6].hex()} for i in range(175)]


def disassemble(objdump, binary, output, start=0, end=None):
    command = [objdump, "-D", "-b", "binary", "-m", "i8086", "-M", "intel",
               f"--start-address={start}"]
    if end is not None:
        command.append(f"--stop-address={end}")
    command.append(str(binary))
    with output.open("w", encoding="utf-8") as f:
        subprocess.run(command, stdout=f, stderr=subprocess.PIPE, check=True, text=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path, default=Path("original/extracted/dos-full"))
    parser.add_argument("--out", type=Path, default=Path("build/original-rules/dos-full"))
    args = parser.parse_args()
    objdump = shutil.which("objdump")
    if not objdump:
        parser.error("GNU objdump with i8086 support is required")
    if args.out.exists():
        parser.error("output already exists; choose a fresh directory")
    source = args.source.resolve()
    output = args.out.resolve()
    # Refuse any output inside the source installation.
    if output == source or source in output.parents:
        parser.error("output must be outside the original installation")
    exe = (source / "HL.EXE").read_bytes()
    if sha256(exe) != KNOWN_HL:
        parser.error("HL.EXE differs from the investigated build; named offsets need retracing")
    header, image, relocations = parse_mz(exe)
    decoded = {name: unpack((source / name).read_bytes()) for name in
               ("UNIT.DAT", "WW1GNDS.DAT", "WW1GNDW.DAT", "NEWUNIT.DAT", "MNUDATA.DAT")}
    unit_rows = units(decoded["UNIT.DAT"])
    ground_rows = {name: terrain(decoded[name]) for name in ("WW1GNDS.DAT", "WW1GNDW.DAT")}
    output.mkdir(parents=True)
    (output / "segments").mkdir()
    (output / "windows").mkdir()
    image_path = output / "HL.image.bin"
    image_path.write_bytes(image)
    disassemble(objdump, image_path, output / "HL.image.linear.asm")
    code_end = 0x32E20  # First module data segment in this exact executable.
    segments = sorted({r["segment"] for r in relocations if r["segment"] * 16 < code_end})
    for i, seg in enumerate(segments):
        # Borland starts a new address segment before the preceding function's
        # final paragraph ends. Keep the 15-byte overlap; never truncate its tail.
        end = min((segments[i+1] * 16 if i+1 < len(segments) else code_end) + 15, code_end)
        binary = output / "segments" / f"{seg:04x}.bin"
        binary.write_bytes(image[seg*16:end])
        disassemble(objdump, binary, binary.with_suffix(".asm"))
    windows = []
    for name, (seg, start, end) in WINDOWS.items():
        binary = output / "windows" / f"{name}.bin"
        binary.write_bytes(image[seg*16:seg*16+end])
        disassemble(objdump, binary, binary.with_suffix(".asm"), start, end)
        windows.append({"name": name, "segment": f"{seg:04x}", "start": f"{start:04x}",
                        "end_exclusive": f"{end:04x}",
                        "file_offset": header["header_bytes"] + seg*16 + start})
    write_csv(output / "relocations.csv", relocations)
    # Only relocation-backed immediate far calls/jumps. Not a complete call graph:
    # near calls and indirect callbacks are deliberately excluded.
    calls = []
    for r in relocations:
        pos = r["image_offset"]
        if 3 <= pos < code_end and image[pos-3] in (0x9A, 0xEA):
            calls.append({"source_segment": f'{r["segment"]:04x}',
                          "source_offset": f'{r["offset"]-3:04x}',
                          "image_offset": pos-3,
                          "target_segment": f'{r["segment_value"]:04x}',
                          "target_offset": f'{int.from_bytes(image[pos-2:pos], "little"):04x}',
                          "kind": "call" if image[pos-3] == 0x9A else "jump"})
    write_csv(output / "far_calls.csv", calls)
    write_csv(output / "units.csv", unit_rows)
    for name, data in decoded.items():
        (output / (name + ".bin")).write_bytes(data)
    for name, rows in ground_rows.items():
        write_csv(output / (name + ".csv"), rows)
    strings = [{"file_offset": m.start(), "image_offset": m.start()-header["header_bytes"],
                "text": m.group().decode("ascii")}
               for m in re.finditer(rb"[\x20-\x7e]{4,}", exe)
               if m.start() >= header["header_bytes"]]
    write_csv(output / "strings.csv", strings)
    manifest = {"source": str(source), "hl_sha256": sha256(exe), "mz": header,
                "objdump_version": subprocess.check_output([objdump, "--version"], text=True).splitlines()[0],
                "segment_count": len(segments), "far_call_or_jump_count": len(calls),
                "windows": windows, "sources": {}}
    for name in ["HL.EXE", "HL14-18.EXE", *decoded]:
        raw = (source / name).read_bytes()
        entry = {"stored_bytes": len(raw), "sha256": sha256(raw)}
        if name in decoded:
            entry.update(unpacked_bytes=len(decoded[name]), unpacked_sha256=sha256(decoded[name]))
        if name.endswith(".EXE"):
            entry["mz"] = parse_mz(raw)[0]
        manifest["sources"][name] = entry
    (output / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Wrote {output}: {len(image)} image bytes, {len(segments)} segment listings, "
          f"{len(windows)} traced windows, {len(unit_rows)} units, 175 terrain records per season")


if __name__ == "__main__":
    main()
