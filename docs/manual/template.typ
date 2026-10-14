// Shared styles for the reconstructed Historyline 1914-1918 manual.

#let accent = rgb("#7a2a1c")
#let ink = rgb("#222222")
#let muted = rgb("#6b6358")
#let paper-tint = rgb("#f3eee4")
#let rule-colour = rgb("#b9ab94")

#let serif = ("Libertinus Serif", "DejaVu Serif")
#let sans = ("Noto Sans", "DejaVu Sans")
#let blackletter = ("UnifrakturMaguntia", "Libertinus Serif")

// Original artwork is extracted from the local PDF at build time (`just manual-images`).
// Without `--input images=true` every picture becomes a labelled placeholder.
#let images-enabled = sys.inputs.at("images", default: "false") == "true"
#let image-dir = "/build/manual/img/"

#let picture(file, width: auto, height: auto, pixelated: false) = {
  if width == auto and height == auto { width = 100% }
  if images-enabled {
    image(image-dir + file, width: width, height: height, fit: "contain", scaling: if pixelated { "pixelated" } else { auto })
  } else {
    box(
      width: if width == auto { height } else { width },
      height: if height != auto { height } else if type(width) == length { width } else { 2.6cm },
      fill: paper-tint,
      stroke: (paint: rule-colour, dash: "dashed"),
      align(center + horizon, text(font: sans, size: 6.5pt, fill: muted, file)),
    )
  }
}

#let fig(file, caption: none, width: 100%) = figure(
  picture(file, width: width),
  caption: caption,
  kind: image,
  supplement: none,
)

// A control symbol with its explanation next to it.
#let symbol-item(file, name, body) = block(breakable: false, grid(
  columns: (2.1cm, 1fr),
  column-gutter: 0.8em,
  align: (center + top, left + top),
  {
    set par(justify: false)
    picture(file, width: 1.9cm)
    v(-0.4em)
    text(size: 7pt, style: "italic", fill: muted, name)
  },
  body,
))

// Text with a picture beside it (replaces the original's wrapped layouts).
#let beside(file, caption: none, width: 45%, body) = grid(
  columns: (1fr, width),
  column-gutter: 1em,
  body,
  fig(file, caption: caption),
)

// Faction marks drawn after the original data sheets: an iron cross for the Central European Powers,
// a roundel for the Allies.
#let iron-cross(size) = box(width: size, height: size, {
  let c = size / 2
  let t = size * 0.3
  let n = size * 0.08
  place(polygon(
    fill: ink,
    (c - t, 0pt), (c + t, 0pt), (c + n, c - n),
    (size, c - t), (size, c + t), (c + n, c + n),
    (c + t, size), (c - t, size), (c - n, c + n),
    (0pt, c + t), (0pt, c - t), (c - n, c - n),
  ))
})

#let roundel(size) = box(width: size, height: size, align(center + horizon, circle(
  radius: size / 2, stroke: 0.5pt + ink, fill: rgb("#8a8a8a"),
  align(center + horizon, circle(radius: size * 0.3, fill: white, align(center + horizon, circle(radius: size * 0.1, fill: ink)))),
)))

#let badge(faction, size: 0.9em) = box(baseline: 15%, {
  if faction in ("allies", "both") { roundel(size) }
  if faction == "both" { h(0.15em) }
  if faction in ("central", "both") { iron-cross(size) }
})

// "p. 12"-style cross reference to a labelled (unnumbered) heading.
#let see(target) = context link(target)[p.~#counter(page).at(locate(target)).first()]

#let ornament = text(fill: accent, size: 8pt)[◆]

#let note(body) = block(
  width: 100%,
  inset: (x: 0.9em, y: 0.7em),
  fill: paper-tint,
  stroke: (left: 2pt + accent),
  text(size: 9pt, body),
)

// Two-column key/value table used for keyboard and scoring lists.
#let keytable(..rows) = table(
  columns: (auto, 1fr),
  stroke: none,
  inset: (x: 0.5em, y: 0.32em),
  fill: (_, y) => if calc.odd(y) { paper-tint },
  ..rows.pos().flatten(),
)

#let part-title(title, target, subtitle: none) = {
  pagebreak(weak: true, to: "odd")
  set page(header: none, footer: none)
  v(1fr)
  align(center)[
    #text(font: blackletter, size: 30pt, fill: ink)[Historyline]
    #v(-0.6em)
    #text(font: serif, size: 13pt, tracking: 0.3em, fill: muted)[1914–1918]
    #v(1.4em)
    #line(length: 40%, stroke: 0.6pt + rule-colour)
    #v(0.6em)
    #heading(level: 1)[#title]#target
    #if subtitle != none { text(style: "italic", fill: muted, subtitle) }
  ]
  v(1.4fr)
  pagebreak()
}

#let manual(body) = {
  set document(
    title: "Historyline 1914–1918 — Instructions and Description of Units",
    author: "Blue Byte (1992); reconstructed for the Historyline remake",
  )
  set text(font: serif, size: 10pt, fill: ink, lang: "en", region: "gb", hyphenate: true)
  set par(justify: true, leading: 0.62em, spacing: 0.95em)

  set page(
    paper: "a5",
    margin: (inside: 1.9cm, outside: 1.6cm, top: 2cm, bottom: 2cm),
    header: context {
      let page = here().page()
      let parts = query(heading.where(level: 1).before(here()))
      let chapters = query(heading.where(level: 2).before(here()))
      if parts.len() == 0 { return }
      // Chapter-opening pages carry their own title.
      if query(heading.where(level: 2)).any(h => h.location().page() == page) { return }
      let part = parts.last().body
      let chapter = if chapters.len() > 0 and chapters.last().location().page() >= parts.last().location().page() {
        chapters.last().body
      }
      set text(size: 7.5pt, fill: muted, tracking: 0.04em)
      if calc.even(page) [#smallcaps(part) #h(1fr) #chapter] else [#chapter #h(1fr) #smallcaps(part)]
      v(-0.5em)
      line(length: 100%, stroke: 0.4pt + rule-colour)
    },
    footer: context {
      set text(size: 8pt, fill: muted)
      align(center)[#ornament #h(0.5em) #counter(page).display() #h(0.5em) #ornament]
    },
  )

  // Level 1: the two booklets (Instructions, Description of Units).
  show heading.where(level: 1): it => text(font: blackletter, size: 24pt, weight: "regular", fill: accent, it.body)

  // Level 2: chapters, each on a fresh page like the original booklet.
  show heading.where(level: 2): it => {
    pagebreak(weak: true)
    v(0.6em)
    align(center)[
      #text(font: blackletter, size: 19pt, weight: "regular", fill: ink, it.body)
      #v(-0.7em)
      #line(length: 30%, stroke: 0.6pt + accent)
    ]
    v(0.8em)
  }

  show heading.where(level: 3): it => block(above: 1.3em, below: 0.6em, sticky: true)[
    #text(size: 11pt, weight: "bold", fill: accent, smallcaps(it.body))
  ]
  show heading.where(level: 4): it => block(above: 1em, below: 0.45em, sticky: true)[
    #text(size: 10pt, weight: "bold", it.body)
  ]

  show figure.caption: it => {
    set par(justify: false)
    text(size: 7.5pt, style: "italic", fill: muted, it.body)
  }
  set figure(gap: 0.45em)
  show figure: set block(above: 1em, below: 1em)

  set enum(indent: 0.2em, body-indent: 0.6em, spacing: 0.75em)
  set list(indent: 0.3em, marker: text(fill: accent)[▸])
  set terms(separator: [ — ], hanging-indent: 1.2em, spacing: 0.9em)
  show terms.item: it => par(hanging-indent: 1.2em)[*#it.term* — #it.description]
  set table(stroke: 0.4pt + rule-colour)

  body
}
