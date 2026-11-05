// Test-only: inputs to Go's LoadGame and what it gave (go1.26, encoding/json), recorded once from the Go
// engine: the error message, or the bytes Save writes for the loaded game. They pin the port's JSON reading
// (syntax errors, type errors with Go's field paths, case-insensitive keys, duplicate keys, nulls, arrays past
// their length, invalid UTF-8) and its refusals.

/** GoLoadCase is one input (in as UTF-8, hex as raw bytes) and Go's result: err, or out, the re-saved game. */
export interface GoLoadCase {
  in?: string;
  hex?: string;
  err?: string;
  out?: string;
}

export const goLoadCases: readonly GoLoadCase[] = [
  { in: "", err: "dosrules: load: unexpected end of JSON input" },
  {
    in: '{"schema":3}x',
    err: "dosrules: load: invalid character 'x' after top-level value",
  },
  {
    in: '{"schema":}',
    err: "dosrules: load: invalid character '}' looking for beginning of value",
  },
  {
    hex: "efbbbf7b7d",
    err: "dosrules: load: invalid character '\u00ef' looking for beginning of value",
  },
  {
    in: '{"build":"a\\u00"}',
    err: "dosrules: load: invalid character '\"' in \\u hexadecimal character escape",
  },
  {
    in: '{"schema":01}',
    err: "dosrules: load: invalid character '1' after object key:value pair",
  },
  {
    in: "[1,]",
    err: "dosrules: load: invalid character ']' looking for beginning of value",
  },
  {
    in: '{"a" 1}',
    err: "dosrules: load: invalid character '1' after object key",
  },
  {
    in: "{1:2}",
    err: "dosrules: load: invalid character '1' looking for beginning of object key string",
  },
  {
    in: '{"a":tru}',
    err: "dosrules: load: invalid character '}' in literal true (expecting 'e')",
  },
  {
    in: '{"a":trux}',
    err: "dosrules: load: invalid character 'x' in literal true (expecting 'e')",
  },
  {
    in: '{"a":-x}',
    err: "dosrules: load: invalid character 'x' in numeric literal",
  },
  {
    in: '{"a":1.}',
    err: "dosrules: load: invalid character '}' after decimal point in numeric literal",
  },
  {
    in: '{"a":1e}',
    err: "dosrules: load: invalid character '}' in exponent of numeric literal",
  },
  {
    in: '{"a":1e+}',
    err: "dosrules: load: invalid character '}' in exponent of numeric literal",
  },
  {
    in: '{"a":"\\q"}',
    err: "dosrules: load: invalid character 'q' in string escape code",
  },
  {
    in: '{"a":"\u0001"}',
    err: "dosrules: load: invalid character '\\x01' in string literal",
  },
  {
    in: '{"a":"\n"}',
    err: "dosrules: load: invalid character '\\n' in string literal",
  },
  {
    in: '{"a":[1 2]}',
    err: "dosrules: load: invalid character '2' after array element",
  },
  {
    in: "nul",
    err: "dosrules: load: invalid character ' ' in literal null (expecting 'l')",
  },
  {
    in: "nulx",
    err: "dosrules: load: invalid character 'x' in literal null (expecting 'l')",
  },
  {
    hex: "80",
    err: "dosrules: load: invalid character '\\u0080' looking for beginning of value",
  },
  {
    hex: "7f",
    err: "dosrules: load: invalid character '\\x7f' looking for beginning of value",
  },
  {
    in: "'",
    err: "dosrules: load: invalid character '\\'' looking for beginning of value",
  },
  {
    in: '{"a":1}}',
    err: "dosrules: load: invalid character '}' after top-level value",
  },
  {
    in: '{"a":falsy}',
    err: "dosrules: load: invalid character 'y' in literal false (expecting 'e')",
  },
  {
    in: "\u00a0",
    err: "dosrules: load: invalid character '\u00c2' looking for beginning of value",
  },
  { in: '{"a"', err: "dosrules: load: unexpected end of JSON input" },
  { in: '{"a":"', err: "dosrules: load: unexpected end of JSON input" },
  { in: "{", err: "dosrules: load: unexpected end of JSON input" },
  { in: "[", err: "dosrules: load: unexpected end of JSON input" },
  {
    in: '{"a":1,}',
    err: "dosrules: load: invalid character '}' looking for beginning of object key string",
  },
  {
    in: '{"a":01.5}',
    err: "dosrules: load: invalid character '1' after object key:value pair",
  },
  {
    in: '{"a":1.5e3x}',
    err: "dosrules: load: invalid character 'x' after object key:value pair",
  },
  { in: '{"a":"\\uD800"}', err: "dosrules: load: schema 0, want 3" },
  {
    in: '{"a":"\\u12G4"}',
    err: "dosrules: load: invalid character 'G' in \\u hexadecimal character escape",
  },
  { in: "-", err: "dosrules: load: invalid character ' ' in numeric literal" },
  { in: "  ", err: "dosrules: load: unexpected end of JSON input" },
  {
    in: '{"schema":1.5}',
    err: "dosrules: load: json: cannot unmarshal number 1.5 into Go struct field savedGame.schema of type int",
  },
  {
    in: '{"schema":1e2}',
    err: "dosrules: load: json: cannot unmarshal number 1e2 into Go struct field savedGame.schema of type int",
  },
  {
    in: '{"schema":"3"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.schema of type int",
  },
  {
    in: '{"schema":true}',
    err: "dosrules: load: json: cannot unmarshal bool into Go struct field savedGame.schema of type int",
  },
  {
    in: '{"schema":99999999999999999999}',
    err: "dosrules: load: json: cannot unmarshal number 99999999999999999999 into Go struct field savedGame.schema of type int",
  },
  {
    in: '{"schema":-9223372036854775809}',
    err: "dosrules: load: json: cannot unmarshal number -9223372036854775809 into Go struct field savedGame.schema of type int",
  },
  {
    in: '{"schema":[]}',
    err: "dosrules: load: json: cannot unmarshal array into Go struct field savedGame.schema of type int",
  },
  {
    in: '{"schema":{}}',
    err: "dosrules: load: json: cannot unmarshal object into Go struct field savedGame.schema of type int",
  },
  {
    in: '{"build":5}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field savedGame.build of type string",
  },
  {
    in: '{"build":[]}',
    err: "dosrules: load: json: cannot unmarshal array into Go struct field savedGame.build of type string",
  },
  {
    in: '{"winter":1}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field savedGame.winter of type bool",
  },
  {
    in: '{"winter":"true"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.winter of type bool",
  },
  {
    in: '{"maps":{}}',
    err: "dosrules: load: json: cannot unmarshal object into Go struct field savedGame.maps of type [2][]int",
  },
  {
    in: '{"maps":[[1,"a"]]}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.maps of type int",
  },
  {
    in: '{"maps":[1]}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field savedGame.maps of type []int",
  },
  {
    in: '{"maps":"x"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.maps of type [2][]int",
  },
  {
    in: '{"units":[]}',
    err: "dosrules: load: json: cannot unmarshal array into Go struct field savedGame.units of type map[int]dosrules.savedUnit",
  },
  {
    in: '{"units":{"x":{}}}',
    err: "dosrules: load: json: cannot unmarshal number x into Go struct field savedGame.units of type int",
  },
  {
    in: '{"units":{"x":{"type":"a"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.type of type int",
  },
  {
    in: '{"units":{"0":5}}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field savedGame.units of type dosrules.savedUnit",
  },
  {
    in: '{"units":{"0":{"pos":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.pos of type [2]int",
  },
  {
    in: '{"units":{"0":{"pos":["x"]}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.pos of type int",
  },
  {
    in: '{"units":{"1.5":{}}}',
    err: "dosrules: load: json: cannot unmarshal number 1.5 into Go struct field savedGame.units of type int",
  },
  {
    in: '{"units":{"99999999999999999999":{}}}',
    err: "dosrules: load: json: cannot unmarshal number 99999999999999999999 into Go struct field savedGame.units of type int",
  },
  {
    in: '{"holds":{"1":{"slots":["a"]}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field Hold.holds.slots of type int",
  },
  {
    in: '{"holds":{"1":{"flags":"a"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field Hold.holds.flags of type int",
  },
  {
    in: '{"holds":{"1":{"cells":{}}}}',
    err: "dosrules: load: json: cannot unmarshal object into Go struct field Hold.holds.cells of type [4]int",
  },
  {
    in: '{"holds":true}',
    err: "dosrules: load: json: cannot unmarshal bool into Go struct field savedGame.holds of type map[int]dosrules.Hold",
  },
  {
    in: '{"players":[1]}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field savedGame.players of type dosrules.savedPlayer",
  },
  {
    in: '{"players":{}}',
    err: "dosrules: load: json: cannot unmarshal object into Go struct field savedGame.players of type [2]dosrules.savedPlayer",
  },
  {
    in: '{"players":[{"orders":[1.5]}]}',
    err: "dosrules: load: json: cannot unmarshal number 1.5 into Go struct field savedPlayer.players.orders of type int",
  },
  {
    in: '{"players":[{"ready":"x"}]}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedPlayer.players.ready of type bool",
  },
  {
    in: '{"players":[{"computer":0}]}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field savedPlayer.players.computer of type bool",
  },
  {
    in: '{"players":[{"orders":{}}]}',
    err: "dosrules: load: json: cannot unmarshal object into Go struct field savedPlayer.players.orders of type []int",
  },
  {
    in: '{"players":[{"phase":"1"}]}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedPlayer.players.phase of type int",
  },
  {
    in: '{"ended":5}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field savedGame.ended of type dosrules.End",
  },
  {
    in: '{"ended":{"won":[1]}}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field End.ended.won of type dosrules.EndReason",
  },
  {
    in: '{"ended":{"over":1}}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field End.ended.over of type bool",
  },
  {
    in: '{"ended":{"won":"x"}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field End.ended.won of type [2]dosrules.EndReason",
  },
  {
    in: '{"ended":{"winner":"x"}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field End.ended.winner of type int",
  },
  {
    in: '{"ended":{"score":true}}',
    err: "dosrules: load: json: cannot unmarshal bool into Go struct field End.ended.score of type int",
  },
  {
    in: '{"ended":[]}',
    err: "dosrules: load: json: cannot unmarshal array into Go struct field savedGame.ended of type dosrules.End",
  },
  {
    in: '{"energy":["x"]}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.energy of type int",
  },
  {
    in: '{"energy":{}}',
    err: "dosrules: load: json: cannot unmarshal object into Go struct field savedGame.energy of type [2]int",
  },
  {
    in: '{"buildings":{"1":{"builds":["x"]}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field Building.buildings.builds of type int",
  },
  {
    in: '{"buildings":{"1":{"credits":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field Building.buildings.credits of type int",
  },
  {
    in: '{"buildings":[]}',
    err: "dosrules: load: json: cannot unmarshal array into Go struct field savedGame.buildings of type map[int]dosrules.Building",
  },
  {
    in: '{"buildings":{"y":{}}}',
    err: "dosrules: load: json: cannot unmarshal number y into Go struct field savedGame.buildings of type int",
  },
  {
    in: '{"buildings":{"1":5}}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field savedGame.buildings of type dosrules.Building",
  },
  {
    in: "[]",
    err: "dosrules: load: json: cannot unmarshal array into Go value of type dosrules.savedGame",
  },
  {
    in: "5",
    err: "dosrules: load: json: cannot unmarshal number into Go value of type dosrules.savedGame",
  },
  {
    in: '"x"',
    err: "dosrules: load: json: cannot unmarshal string into Go value of type dosrules.savedGame",
  },
  {
    in: "true",
    err: "dosrules: load: json: cannot unmarshal bool into Go value of type dosrules.savedGame",
  },
  {
    in: '{"SCHEMA":"x"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.schema of type int",
  },
  {
    in: '{"turn":"x"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.turn of type int",
  },
  {
    in: '{"limit":"x"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.limit of type int",
  },
  {
    in: '{"concealed":"x"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.concealed of type bool",
  },
  {
    in: '{"width":"x"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.width of type int",
  },
  {
    in: '{"height":[]}',
    err: "dosrules: load: json: cannot unmarshal array into Go struct field savedGame.height of type int",
  },
  {
    in: '{"turn":"x","schema":"y"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.turn of type int",
  },
  {
    in: '{"units":{"1":{"type":"a"},"0":{"budget":"b"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.type of type int",
  },
  {
    in: '{"units":{"0":{"budget":"b","exp":"c"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.budget of type int",
  },
  {
    in: '{"units":{"x":{"type":5},"y":{"type":"z"}}}',
    err: "dosrules: load: json: cannot unmarshal number x into Go struct field savedGame.units of type int",
  },
  {
    in: '{"units":{"0":{"budget":1.5}}}',
    err: "dosrules: load: json: cannot unmarshal number 1.5 into Go struct field savedUnit.units.budget of type int",
  },
  {
    in: '{"units":{"0":{"exp":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.exp of type int",
  },
  {
    in: '{"units":{"0":{"size":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.size of type int",
  },
  {
    in: '{"units":{"0":{"alt":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.alt of type int",
  },
  {
    in: '{"units":{"0":{"flags4":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.flags4 of type int",
  },
  {
    in: '{"units":{"0":{"flags6":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.flags6 of type int",
  },
  {
    in: '{"units":{"0":{"serial":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.serial of type int",
  },
  {
    in: '{"units":{"0":{"hold":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.hold of type int",
  },
  {
    in: '{"units":{"0":{"facing":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.facing of type [2]int",
  },
  {
    in: '{"units":{"0":{"target":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.target of type int",
  },
  {
    in: '{"units":{"0":{"action":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.action of type [7]int",
  },
  {
    in: '{"units":{"0":{"action":["x"]}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedUnit.units.action of type int",
  },
  {
    in: '{"holds":{"1":{"occupant":"x"}}}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field Hold.holds.occupant of type int",
  },
  {
    in: '{"holds":{"x":{}}}',
    err: "dosrules: load: json: cannot unmarshal number x into Go struct field savedGame.holds of type int",
  },
  {
    in: '{"holds":{"1":5}}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field savedGame.holds of type dosrules.Hold",
  },
  {
    in: '{"maps":[["x"]],"schema":"y"}',
    err: "dosrules: load: json: cannot unmarshal string into Go struct field savedGame.maps of type int",
  },
  {
    in: '{"ended":{"won":["",{}]}}',
    err: "dosrules: load: json: cannot unmarshal object into Go struct field End.ended.won of type dosrules.EndReason",
  },
  {
    in: '{"Ended":{"Won":["",5]}}',
    err: "dosrules: load: json: cannot unmarshal number into Go struct field End.ended.won of type dosrules.EndReason",
  },
  { in: "null", err: "dosrules: load: schema 0, want 3" },
  { in: " \t\r\n{} ", err: "dosrules: load: schema 0, want 3" },
  {
    in: '{"schema":3,"build":"x\\"y\\n<\\u0001\\u00e9\\u00a0\\u007f"}',
    err: 'dosrules: load: unknown tables "x\\"y\\n<\\x01\u00e9\\u00a0\\x7f"',
  },
  {
    in: '{"schema":2,"build":"83fc4e37","winter":true}',
    err: 'dosrules: load: tables "83fc4e37" have no winter terrain',
  },
  {
    in: '{"schema":3,"build":"fe4fa446"}',
    err: "dosrules: load: 0\u00d70 map",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":0,"height":5}',
    err: "dosrules: load: 0\u00d75 map",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":-1,"height":-2}',
    err: "dosrules: load: -1\u00d7-2 map",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"units":{"241":{}}}',
    err: "dosrules: load: unit 241 of type 0",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"units":{"-1":{}}}',
    err: "dosrules: load: unit -1 of type 0",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"units":{"0":{"type":-1}}}',
    err: "dosrules: load: unit 0 of type -1",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"units":{"0":{"type":51}}}',
    err: "dosrules: load: unit 0 of type 51",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"units":{"0":{"type":50}}}',
    err: "dosrules: load: map 0 has 0 entries for 1 cells",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[1,2],[1]]}',
    err: "dosrules: load: map 0 shows unit 2 at 0",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[1,2,3],[1]]}',
    err: "dosrules: load: map 0 has 3 entries for 1 cells",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[1,255],[1]]}',
    err: "dosrules: load: map 1 has 1 entries for 1 cells",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[1,256],[1,2]]}',
    err: "dosrules: load: map 0 shows unit 256 at 0",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[1,-1],[1,2]]}',
    err: "dosrules: load: map 0 shows unit -1 at 0",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[1,240],[1,2]]}',
    err: "dosrules: load: map 0 shows unit 240 at 0",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[1,241],[1,0]]}',
    err: "dosrules: load: map 1 shows unit 0 at 0",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[1,255],[1,255]],"players":[{"orders":[0]}]}',
    err: "dosrules: load: side 0 orders unit 0",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[1,255],[1,255]],"players":[{},{"orders":[3]}]}',
    err: "dosrules: load: side 1 orders unit 3",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]]}',
    err: "dosrules: load: dosrules: phases 0/0",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":2},{"phase":2}]}',
    err: "dosrules: load: dosrules: phases 2/2",
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":1}]}',
    err: "dosrules: load: dosrules: phases 1/1",
  },
  { in: '{"schema":4}', err: "dosrules: load: schema 4, want 3" },
  { in: '{"schema":1,"build":"zz"}', err: "dosrules: load: schema 1, want 3" },
  {
    in: '{"schema":3,"build":"FE4FA446"}',
    err: 'dosrules: load: unknown tables "FE4FA446"',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}]}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"winter":true}',
    out: '{"schema":3,"build":"fe4fa446","winter":true,"width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"winter":false,"concealed":false,"ended":null,"buildings":{}}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"holds":null,"units":null}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"ended":{}}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"ended":{"over":false,"winner":0,"won":["",""],"score":0},"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"ended":{"over":true,"winner":-1,"won":["<&>\\u2028\\u2029\\ud800\\u0041\\udc00","\\u00e9\\u0001\\b\\f\\n\\r\\t\\"\\\\\\/\\u007f"],"score":5}}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"ended":{"over":true,"winner":-1,"won":["\\u003c\\u0026\\u003e\\u2028\\u2029\ufffdA\ufffd","\u00e9\\u0001\\b\\f\\n\\r\\t\\"\\\\/"],"score":5},"energy":[0,0]}',
  },
  {
    hex: "7b22736368656d61223a332c226275696c64223a226665346661343436222c227769647468223a312c22686569676874223a312c226d617073223a5b5b302c3235355d2c5b302c3235355d5d2c22706c6179657273223a5b7b227068617365223a317d2c7b227068617365223a327d5d2c22656e646564223a7b22776f6e223a5b22e2827affeda080222c2259225d7d7d",
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"ended":{"over":false,"winner":0,"won":["\ufffd\ufffdz\ufffd\ufffd\ufffd\ufffd","Y"],"score":0},"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"ended":{"won":["\\ud83d\\ude00\u00e9\u4e2d","\\udc00\\ud800"]}}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"ended":{"over":false,"winner":0,"won":["\u{1F600}\u00e9\u4e2d","\ufffd\ufffd"],"score":0},"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"Schema":3,"WIDTH":1,"\u017fchema":3,"Concealed":true,"\u212aeep":1,"limit":4,"LIMIT":16,"turn":-0,"Turn":7}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":7,"limit":16,"concealed":true,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"extra":1.5e300,"more":[1,"x",{}],"maps":[[0,255],[0,255],[1,2,3]],"energy":[1,2,3],"x":{"y":null}}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[1,2]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"energy":[5]}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[5,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"units":{"+0":{"type":1},"00":{"type":2,"pos":[1,2,3],"facing":[4],"action":[1,2,3,4,5,6,7,8,"x"],"unknown":true}}}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{"0":{"budget":0,"exp":0,"size":0,"alt":0,"flags4":0,"flags6":0,"type":2,"serial":0,"hold":0,"pos":[1,2],"facing":[4,0],"target":0,"action":[1,2,3,4,5,6,7]}},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"units":{"0":null},"maps":[[0,0],[0,255]]}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,0],[0,255]],"units":{"0":{"budget":0,"exp":0,"size":0,"alt":0,"flags4":0,"flags6":0,"type":0,"serial":0,"hold":0,"pos":[0,0],"facing":[0,0],"target":0,"action":[0,0,0,0,0,0,0]}},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"units":{"10":{},"2":{},"1":{},"0":{},"240":{}}}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{"0":{"budget":0,"exp":0,"size":0,"alt":0,"flags4":0,"flags6":0,"type":0,"serial":0,"hold":0,"pos":[0,0],"facing":[0,0],"target":0,"action":[0,0,0,0,0,0,0]},"1":{"budget":0,"exp":0,"size":0,"alt":0,"flags4":0,"flags6":0,"type":0,"serial":0,"hold":0,"pos":[0,0],"facing":[0,0],"target":0,"action":[0,0,0,0,0,0,0]},"10":{"budget":0,"exp":0,"size":0,"alt":0,"flags4":0,"flags6":0,"type":0,"serial":0,"hold":0,"pos":[0,0],"facing":[0,0],"target":0,"action":[0,0,0,0,0,0,0]},"2":{"budget":0,"exp":0,"size":0,"alt":0,"flags4":0,"flags6":0,"type":0,"serial":0,"hold":0,"pos":[0,0],"facing":[0,0],"target":0,"action":[0,0,0,0,0,0,0]},"240":{"budget":0,"exp":0,"size":0,"alt":0,"flags4":0,"flags6":0,"type":0,"serial":0,"hold":0,"pos":[0,0],"facing":[0,0],"target":0,"action":[0,0,0,0,0,0,0]}},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"holds":{"64":{"slots":[1],"cells":[2],"flags":3,"occupant":4},"65":{"flags":32768},"0":{"slots":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15]},"256":null,"-3":{},"100":{},"128":{}}}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{"-3":{"slots":[0,0,0,0,0,0,0,0,0,0,0,0,0,0],"cells":[0,0,0,0],"flags":0,"occupant":0},"0":{"slots":[1,2,3,4,5,6,7,8,9,10,11,12,13,14],"cells":[0,0,0,0],"flags":0,"occupant":0},"100":{"slots":[0,0,0,0,0,0,0,0,0,0,0,0,0,0],"cells":[0,0,0,0],"flags":8,"occupant":0},"128":{"slots":[0,0,0,0,0,0,0,0,0,0,0,0,0,0],"cells":[0,0,0,0],"flags":8,"occupant":0},"256":{"slots":[0,0,0,0,0,0,0,0,0,0,0,0,0,0],"cells":[0,0,0,0],"flags":8,"occupant":0},"64":{"slots":[1,0,0,0,0,0,0,0,0,0,0,0,0,0],"cells":[2,0,0,0],"flags":3,"occupant":4},"65":{"slots":[0,0,0,0,0,0,0,0,0,0,0,0,0,0],"cells":[0,0,0,0],"flags":32768,"occupant":0}},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"buildings":{"64":{"credits":7,"builds":[]},"3":{"builds":[22,18]},"10":{},"-1":null}}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0],"buildings":{"-1":{"credits":0},"10":{"credits":0},"3":{"credits":0,"builds":[22,18]},"64":{"credits":7}}}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"players":[{"phase":1,"orders":[],"ready":false},{"phase":2,"computer":true,"ready":true}]}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2,"ready":true,"computer":true}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"players":[{"phase":1},{"phase":2}],"turn":12,"limit":99,"concealed":true,"energy":[-5,1000000]}',
    out: '{"schema":3,"build":"fe4fa446","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":12,"limit":99,"concealed":true,"energy":[-5,1000000]}',
  },
  {
    in: '{"schema":2,"build":"83fc4e37","width":2,"height":1,"maps":[[0,255,0,255],[0,255,0,255]],"players":[{"phase":2},{"phase":1}]}',
    out: '{"schema":3,"build":"83fc4e37","width":2,"height":1,"maps":[[0,255,0,255],[0,255,0,255]],"units":{},"holds":{},"players":[{"phase":2},{"phase":1}],"turn":0,"limit":0,"energy":[0,0]}',
  },
  {
    in: '{"build":"fe4fa446","schema":3,"height":1,"width":1,"players":[{"phase":1},{"phase":2}],"maps":[[0,255],[0,255]],"build":"83fc4e37"}',
    out: '{"schema":3,"build":"83fc4e37","width":1,"height":1,"maps":[[0,255],[0,255]],"units":{},"holds":{},"players":[{"phase":1},{"phase":2}],"turn":0,"limit":0,"energy":[0,0]}',
  },
];
