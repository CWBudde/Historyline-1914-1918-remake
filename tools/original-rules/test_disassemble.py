import struct
import unittest

import disassemble as research


class TPWMTests(unittest.TestCase):
    def test_literal_and_overlapping_reference(self):
        packed = b"TPWM" + struct.pack("<I", 9) + b"\x10ABC\x03\x03"
        self.assertEqual(research.unpack(packed), b"ABCABCABC")

    def test_final_reference_is_clipped_to_declared_size(self):
        packed = b"TPWM" + struct.pack("<I", 4) + b"\x10ABC\x0f\x03"
        self.assertEqual(research.unpack(packed), b"ABCA")

    def test_rejects_bad_bounds_and_trailing_data(self):
        cases = [b"TPWM", b"TPWM" + bytes(4),
                 b"TPWM" + struct.pack("<I", 2) + b"\x00A",
                 b"TPWM" + struct.pack("<I", 3) + b"\x80\x00\x01",
                 b"TPWM" + struct.pack("<I", 3) + b"\x80\x00",
                 b"TPWM" + struct.pack("<I", 1) + b"\x00Ax",
                 b"TPWM" + struct.pack(">I", 3) + b"\x00ABC"]
        for packed in cases:
            with self.subTest(packed=packed), self.assertRaises(ValueError):
                research.unpack(packed)

    def test_uncompressed_data(self):
        self.assertEqual(research.unpack(b"plain"), b"plain")


class MZTests(unittest.TestCase):
    def fixture(self):
        data = bytearray(80)
        struct.pack_into("<14H", data, 0, 0x5A4D, 80, 1, 1, 4, 0, 0xFFFF,
                         0, 0x80, 0, 0, 0, 28, 0)
        struct.pack_into("<HH", data, 28, 2, 0)
        struct.pack_into("<H", data, 66, 0x1234)
        return data

    def test_relocations_and_overlay_are_separate_from_load_image(self):
        data = self.fixture() + b"overlay"
        header, image, relocs = research.parse_mz(data)
        self.assertEqual(header["header_bytes"], 64)
        self.assertEqual(header["trailing_file_bytes"], 7)
        self.assertEqual(len(image), 16)
        self.assertEqual(relocs, [{"segment": 0, "offset": 2, "image_offset": 2,
                                   "segment_value": 0x1234}])

    def test_rejects_truncated_image_and_invalid_relocations(self):
        cases = [self.fixture()[:-1], b"MZ", bytes(80)]
        data = self.fixture()
        struct.pack_into("<HH", data, 28, 15, 0)
        cases.append(data)
        data = self.fixture()
        struct.pack_into("<H", data, 24, 63)
        cases.append(data)
        for data in cases:
            with self.subTest(data=data), self.assertRaises(ValueError):
                research.parse_mz(data)


class RecordTests(unittest.TestCase):
    def test_preserves_unknown_bytes_and_decodes_offsets(self):
        record = bytearray(69)
        record[:4] = bytes([12, 4, 20, 6])
        record[27:31] = b"TYPE"
        record[44:48] = b"NAME"
        record[61:66] = bytes([65, 66, 90, 15, 4])
        rows = research.units(bytes(record) * 51)
        self.assertEqual(len(rows), 51)
        self.assertEqual(rows[50]["offset"], 50 * 69)
        self.assertEqual(rows[0]["name"], "NAME")
        self.assertEqual(rows[0]["cost_raw"], 90)
        self.assertEqual(rows[0]["capacity_raw"], 15)
        self.assertEqual(rows[0]["weight_raw"], 4)
        self.assertEqual(bytes.fromhex(rows[0]["raw_hex"]), record)

    def test_terrain_stride_and_little_endian_flags(self):
        rows = research.terrain(bytes([8, 128, 16, 6, 2, 8]) * 175)
        self.assertEqual(rows[174]["offset"], 174 * 6)
        self.assertEqual(rows[0]["flags_raw"], 0x8008)
        self.assertEqual(rows[0]["movement_cost_raw"], 6)
        self.assertEqual(rows[0]["combat_class"], 8)

    def test_rejects_other_record_counts(self):
        with self.assertRaises(ValueError):
            research.units(bytes(69))
        with self.assertRaises(ValueError):
            research.terrain(bytes(6))


if __name__ == "__main__":
    unittest.main()
