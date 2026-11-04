// The movement queries of segment 0b06 (demo addresses) work on a byte per cell in arrays of 64 cells per row
// (index row*64+column, 0x1104 bytes); the map itself is two bytes per cell in rows of Width cells (terrain
// record, unit index, 0xf1 and up empty). Column and row are cell % Width and cell / Width; odd columns sit
// half a cell lower.
//
// L1 holds only the table records the profile reads (MoveType, MoveTerrain, MoveTables); the queries follow.

/** MoveType holds the fields of a type record the movement queries read. */
export interface MoveType {
  /** Byte 0, copied into live byte 0 by the constructor. */
  budget: number;
  /** Byte 1, the displayed speed (not read here). */
  speed: number;
  /** Byte 5, ANDed with the terrain mask. */
  mask: number;
  /** Word +0x0d, copied into live +4 (0x40: a ship two cells long). */
  class: number;
  /** Word +0x11, copied into live +6 (0x10: a flat cost of 2). */
  flags: number;
  /** The path score weighs the distance to the goal by +0x15 (signed byte). */
  distanceWeight: number;
  /** The path score weighs the entry cost by +0x16 (signed byte). */
  costWeight: number;
  /** A transporter's capacity (+0x40), which the end-cell check compares (UNIT.DAT byte 64). */
  capacity: number;
  /** Every type's weight (+0x41, UNIT.DAT byte 65). */
  weight: number;
  /** The classes (container flags 0xfc0) a transporter takes, word +0x13. */
  carries: number;
}

/** MoveTerrain is a 6-byte terrain record as the movement queries read it. */
export interface MoveTerrain {
  /** Word 0: 0x4000 closed to air units; 0x400/0x100/0x40 buildings. */
  flags: number;
  /** Byte 2. */
  mask: number;
  /** Byte 3. */
  cost: number;
  /** Byte 4, the cost for units with live +4 bit 0x10. */
  airCost: number;
  /** Byte 5, the combat class the order handler reads (Battlefield.Terrain). */
  class: number;
}

/** MoveTables are a build's type and terrain records. */
export interface MoveTables {
  types: MoveType[];
  /** The summer table (WW1GNDS.DAT). */
  terrains: MoveTerrain[];
  /**
   * The winter table (WW1GNDW.DAT) where the build has one (Go's nil: null). In the full build it differs
   * from the summer table in the cost only.
   */
  winter: MoveTerrain[] | null;
}
