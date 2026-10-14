#import "../template.typ": *

== Bibliography and List of Illustrations <bibliography>

#let sources = (
  A: [_Enzyklopädie der Flugzeuge_, Weltbild Verlag GmbH, Augsburg],
  B: [_Kampfflugzeuge_, Buch und Zeit Verlagsgesellschaft mbH, Cologne],
  C: [_Luftkämpfe_, Buch und Zeit Verlagsgesellschaft mbH, Cologne],
  D: [_Waffen im Einsatz_, John Jahr Verlag KG, Hamburg],
  E: [_Illustrierte Geschichte des Weltkrieges_, Union Deutsche Verlagsgesellschaft in Stuttgart, Berlin,
    Leipzig, Vienna],
  F: [_Chronik 1914_ -- _Chronik 1918_, Chronik Verlag, Dortmund],
  G: [_Illustrierte Geschichte des Ersten Weltkrieges_, Nebel Verlag GmbH & Co. KG, Munich],
  H: [_Das große Buch der Schiffstypen_, Dudszus & Köpke, Pietsch-Verlag],
  I: [_Panzer_, Verlagsunion Erich Pabel -- Arthur Moewig KG, Rastatt],
  J: [_Mig Megvalósul Gyönyörü Képességünk A Rend_, Móra Ferenc Verlag, Budapest],
  K: [_Az Elsö Világháború_, Ernö Gundos, Móra Ferenc Verlag, Budapest],
)

#table(
  columns: (auto, 1fr),
  stroke: none,
  inset: (x: 0.4em, y: 0.3em),
  ..sources.pairs().map(((k, v)) => (text(weight: "bold", fill: accent, k), text(size: 9pt, v))).flatten(),
)

=== Illustrations

#let illustrations = (
  ("Infantry", "G"), ("Elite Infantry", "E"), ("Cavalry", "E"), ("Anti-Tank Guns, PAK", "E"),
  ("Sappers", "E"), ("Supply Depot, Construction Unit", "E"), ("Transport of Supplies", "E"),
  ("Supply Train", "F"), ("Static Anti-aircraft Emplacement", "E"), ("Mobile Anti-aircraft Emplacement", "F"),
  ("Light Field Artillery", "F"), ("Medium Field Artillery", "E"), ("Heavy Field Artillery", "E"),
  ("Train-mounted Artillery", "G"), ("Armoured Train", "J"), ("Bunker", "K"), ("Static Balloon", "E"),
  ("German Armoured Car", "E"), ("A7V Tank", "I"), ("Charron Armoured Car", "I"), ("St. Chamond Tank (16)", "F"),
  ("Renault FT 17 Light Tank", "I"), ("Battle Tank Mark I", "I"), ("Battle Tank Mark IV", "F"),
  ("Fokker E.I", "A"), ("Fokker E.III", "A"), ("Albatross D.III", "A"), ("Fokker Dr.I", "A"),
  ("Fokker D.VII", "A"), ("Junkers J 4-10", "C"), ("Gotha Bomber", "B"), ("Zeppelin Staaken", "D"),
  ("D.H.2", "A"), ("Morane-Saulnier N", "C"), ("Nieuport XVII", "C"), ("Spad VII", "B"), ("S.E.5a", "A"),
  ("Spad XIII", "B"), ("Sopwith Camel", "A"), ("Voisin III (LA)", "C"), ("D.H.4", "C"),
  ("Handley Page 0/400", "A"), ("Patrol Boat", "E"), ("Torpedo Boat", "H"), ("Torpedo Boat Destroyer", "H"),
  ("Battleship (Ship of the Line)", "H"), ("Transport Ship", "H"), ("Submarine", "H"),
)

#columns(2, gutter: 1.2em, {
  set text(size: 8pt)
  set par(justify: false)
  for (name, src) in illustrations {
    block(spacing: 0.42em)[#name #box(width: 1fr, repeat[.#h(2pt)]) #text(weight: "bold", fill: accent, src)]
  }
})

#v(1em)
We would like to thank Mr. Hans W. Cada for placing at our disposal his collection of original issues of the
_Illustrierte Geschichte des Weltkrieges_ of 1914–1918.
