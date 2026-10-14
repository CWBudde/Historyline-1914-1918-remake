#import "../template.typ": *

#let units = yaml("units.yaml")

#let sections = (
  ground: "Ground Troops",
  tanks: "Tanks",
  air: "Airplanes",
  navy: "Ships",
)

#let terrain = (
  ("deep-water", "Deep water"),
  ("shallow-water", "Shallow water"),
  ("plain", "Plain, road, bridge"),
  ("forest", "Forest"),
  ("mountains", "Mountains, narrow bridges"),
  ("trenches", "Trenches, wire, barriers"),
  ("tracks", "Tracks"),
)

#let tactical-rows = (
  ("air", "Attack against airborne targets"),
  ("water", "Attack against targets on water"),
  ("ground", "Attack against targets on ground"),
  ("defence", "Defence value"),
  ("speed", "Speed"),
  ("weight", "Weight"),
  ("group", "Strength of group"),
  ("cost", "Production costs"),
)

#let spec-box(spec) = block(
  width: 100%,
  inset: 0.7em,
  fill: paper-tint,
  radius: 2pt,
  {
    set text(size: 7.5pt)
    set par(justify: false, leading: 0.5em, spacing: 0.55em)
    text(size: 7pt, weight: "bold", tracking: 0.12em, fill: accent)[SPECIFICATION]
    if "model" in spec {
      parbreak()
      emph(spec.model)
    }
    for item in spec.items {
      parbreak()
      [*#item.key:* #item.value]
    }
  },
)

#let terrain-strip(impassable) = table(
  columns: (auto, 1.1em),
  inset: (x: 0.35em, y: 0.22em),
  stroke: none,
  fill: (_, y) => if calc.odd(y) { paper-tint },
  align: (right, center),
  ..terrain
    .map(((key, label)) => (
      text(size: 6.5pt, fill: muted, label),
      if key in impassable { text(fill: accent, weight: "bold", size: 8pt)[✘] } else {
        text(fill: rule-colour, size: 7pt)[○]
      },
    ))
    .flatten(),
)

#let data-sheet(u) = block(
  width: 100%,
  breakable: false,
  stroke: (top: 1.2pt + ink, bottom: 1.2pt + ink),
  inset: (y: 0.6em),
  {
    align(center, text(font: blackletter, size: 12pt)[Tactical Data])
    v(-0.2em)
    grid(
      columns: (1fr, auto, auto),
      column-gutter: 0.9em,
      align: (left + horizon, center + horizon, right + horizon),
      table(
        columns: (1fr, auto),
        inset: (x: 0.35em, y: 0.24em),
        stroke: none,
        fill: (_, y) => if calc.odd(y) { paper-tint },
        align: (left, right),
        ..tactical-rows
          .map(((key, label)) => (
            text(size: 7.5pt, label),
            text(font: sans, size: 7.5pt, weight: "bold", str(u.tactical.at(key))),
          ))
          .flatten(),
      ),
      stack(
        spacing: 0.6em,
        badge(u.faction, size: 1.3em),
        ..u.sprites.map(s => box(
          stroke: 0.4pt + rule-colour,
          inset: 2pt,
          picture(s, width: 1.25cm, height: 1.25cm, pixelated: true),
        )),
      ),
      terrain-strip(u.impassable),
    )
  },
)

#let photos(u) = {
  let n = u.photos.len()
  if n == 0 { return }
  let height = if n > 1 { 3.6cm } else { 4.6cm }
  align(center, grid(
    columns: n,
    column-gutter: 0.8em,
    ..u.photos.map(p => figure(
      picture(p.file, height: height),
      caption: if p.caption != "" { p.caption },
      kind: image,
      supplement: none,
      outlined: false,
    )),
  ))
}

#let unit-sheet(u) = {
  pagebreak(weak: true)
  heading(level: 3, outlined: true, bookmarked: true)[#u.name]
  if "subtitle" in u {
    v(-0.6em)
    align(center, text(style: "italic", fill: muted, u.subtitle))
  }
  photos(u)
  set text(size: 9pt)
  set par(leading: 0.55em)
  if "spec" in u {
    grid(
      columns: (1fr, 0.95fr),
      column-gutter: 1em,
      u.text, spec-box(u.spec),
    )
  } else {
    u.text
  }
  v(1fr)
  data-sheet(u)
}

// Section divider: the chapter heading sits alone, centred, on its own page.
#let section-page(title) = {
  pagebreak(weak: true)
  v(1fr)
  show heading.where(level: 2): it => align(center)[
    #ornament
    #v(0.4em)
    #text(font: blackletter, size: 28pt, weight: "regular", fill: accent, it.body)
    #v(0.2em)
    #ornament
  ]
  heading(level: 2, title)
  v(1.3fr)
}

#for (key, title) in sections {
  section-page(title)
  show heading.where(level: 3): it => align(center, text(
    font: blackletter,
    size: 18pt,
    weight: "regular",
    fill: ink,
    it.body,
  ))
  for u in units.filter(u => u.section == key) {
    unit-sheet(u)
  }
}
