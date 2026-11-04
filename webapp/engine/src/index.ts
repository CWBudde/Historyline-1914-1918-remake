// The web edition's rules engine: a TypeScript port of the original rules profile (internal/dosrules and the
// packages around it), held to the Go engine by its fixtures under internal/dosrules/testdata.

export const engineVersion = "0.0.0";

export {
  isFixtureRecord,
  type FixtureRecord,
  type FixtureSource,
} from "./fixture.ts";
