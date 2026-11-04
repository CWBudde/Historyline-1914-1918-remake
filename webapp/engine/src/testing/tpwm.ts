// Test-only: the Turbo Packer ("TPWM") unpacker of internal/tpwm for the DOS files (little-endian size), which
// strength.test.ts uses to read the local original unit table.

const magic = [0x54, 0x50, 0x57, 0x4d]; // "TPWM"

/** maxUnpacked bounds the declared output size. */
export const maxUnpacked = 32 * 1024 * 1024;

/** hasMagic reports whether b starts with a TPWM header. */
export function hasMagic(b: Uint8Array): boolean {
  return b.length >= 4 && magic.every((v, i) => b[i] === v);
}

/**
 * unpack decodes a TPWM stream whose size field is little-endian (Go's tpwm.Unpack with binary.LittleEndian).
 * The stream must be consumed exactly. Data without the TPWM magic is returned unchanged.
 */
export function unpack(b: Uint8Array): Uint8Array {
  if (!hasMagic(b)) {
    return b;
  }
  if (b.length < 8) {
    throw new Error("truncated TPWM header");
  }
  const byte = (i: number): number => {
    const v = b[i];
    if (v === undefined) {
      throw new RangeError(`index ${i} past ${b.length}`);
    }
    return v;
  };
  const size =
    (byte(4) | (byte(5) << 8) | (byte(6) << 16) | (byte(7) << 24)) >>> 0;
  if (size <= 0 || size > maxUnpacked) {
    throw new Error(`invalid TPWM size ${size}`);
  }
  const out: number[] = [];
  let pos = 8;
  while (out.length < size) {
    if (pos >= b.length) {
      throw new Error("truncated TPWM flag");
    }
    const flags = byte(pos);
    pos++;
    for (let bit = 7; bit >= 0 && out.length < size; bit--) {
      if ((flags & (1 << bit)) === 0) {
        if (pos >= b.length) {
          throw new Error("truncated TPWM literal");
        }
        out.push(byte(pos));
        pos++;
        continue;
      }
      if (pos + 2 > b.length) {
        throw new Error("truncated TPWM reference");
      }
      const hi = byte(pos);
      const lo = byte(pos + 1);
      pos += 2;
      const distance = ((hi & 0xf0) << 4) | lo;
      if (distance === 0 || distance > out.length) {
        throw new Error("invalid TPWM distance");
      }
      for (let n = Math.min((hi & 15) + 3, size - out.length); n > 0; n--) {
        out.push(out[out.length - distance] ?? 0);
      }
    }
  }
  if (pos !== b.length) {
    throw new Error("trailing bytes after TPWM stream");
  }
  return Uint8Array.from(out);
}
