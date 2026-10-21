// Historyline 1914-1918 remake — art book. One page per art family from art/index.json
// (`just art-index`), the style guide art/style.md and the PULSE reference design.
// Build with `just artbook` (build/artbook/Historyline_Art.pdf).
//
// Inputs: base (the art tree, default the repository root), thumbs (directory of the
// master and sample thumbnails written by `hlart thumbs`; default: the files themselves), pulse
// (directory of the PULSE map written by `hlart pulse`; default: no map chapter), style=false
// skips art/style.md, reference=false skips the PULSE reference chapter.

#import "/docs/manual/template.typ": accent, ink, muted, paper-tint, rule-colour, serif, sans, blackletter, ornament, note
#import "markdown.typ": md, reflow, inline

#let repo = "https://github.com/CWBudde/Historyline-1914-1918-remake/"
#let blob(path) = link(repo + "blob/main/" + path, raw(path))
#let tree(path) = link(repo + "tree/main/" + path, raw(path + "/"))

#let base = sys.inputs.at("base", default: "")
#let index = json(base + "/art/index.json")
#let thumbs = sys.inputs.at("thumbs", default: "")
#let with-style = sys.inputs.at("style", default: "true") == "true"
#let with-reference = sys.inputs.at("reference", default: "true") == "true"
#let pulse = sys.inputs.at("pulse", default: "")

#let kinds = (
  ("terrain", "Terrain"),
  ("unit", "Units"),
  ("portrait", "Portraits"),
  ("interface", "Interface"),
  ("screen", "Screens"),
)

#let art-label(id) = label("art:" + id)
#let known = index.assets.map(a => a.id)

#set document(title: "Historyline 1914–1918 remake — Art book", author: "Historyline remake")
#set text(font: serif, size: 9.5pt, fill: ink, lang: "en", region: "gb")
#set par(justify: true, leading: 0.6em, spacing: 0.9em)
#set table(stroke: 0.4pt + rule-colour, inset: (x: 0.45em, y: 0.35em))
#show table: set par(justify: false)
#set list(marker: text(fill: accent)[▸])
#show link: set text(fill: accent)
#show raw: set text(size: 8pt)

#set page(
  paper: "a4",
  margin: (x: 1.8cm, top: 2cm, bottom: 1.8cm),
  header: context {
    let page = here().page()
    let chapters = query(heading.where(level: 1)).filter(h => h.location().page() <= page)
    if chapters.len() == 0 or chapters.last().location().page() == page { return }
    set text(size: 7.5pt, fill: muted, tracking: 0.04em)
    [Historyline 1914–1918 · Art book #h(1fr) #smallcaps(chapters.last().body)]
    v(-0.5em)
    line(length: 100%, stroke: 0.4pt + rule-colour)
  },
  footer: context {
    set text(size: 8pt, fill: muted)
    align(center)[#ornament #h(0.5em) #counter(page).display() #h(0.5em) #ornament]
  },
)

#show heading.where(level: 1): it => {
  pagebreak(weak: true)
  v(1.2em)
  align(center)[
    #text(font: blackletter, size: 24pt, weight: "regular", fill: accent, it.body)
    #v(-0.6em)
    #line(length: 30%, stroke: 0.6pt + accent)
  ]
  v(1em)
}
#show heading.where(level: 2): it => block(above: 0em, below: 0.5em)[
  #text(font: blackletter, size: 17pt, weight: "regular", fill: ink, it.body)
]
#show heading.where(level: 3): it => block(above: 1.1em, below: 0.5em, sticky: true)[
  #text(size: 10pt, weight: "bold", fill: accent, smallcaps(it.body))
]
#show heading.where(level: 4): it => block(above: 0.9em, below: 0.4em, sticky: true, text(weight: "bold", it.body))

#let tag(body, fill: muted) = box(
  inset: (x: 0.4em, y: 0.15em), radius: 2pt, fill: fill,
  text(font: sans, size: 7pt, fill: white, weight: "bold", upper(body)),
)
#let status-tag(s) = tag(s, fill: if s == "final" { accent } else { muted })
#let small(body) = text(size: 7.5pt, fill: muted, body)
#let ids(xs) = if xs.len() == 0 { small[—] } else { xs.map(x => raw(x)).join(", ") }

#let picture(path, width: 100%, height: auto) = image(path, width: width, height: height, fit: "contain")
#let file(path) = if thumbs == "" { base + "/" + path } else { thumbs + "/" + path }

// Title page
#page(header: none, footer: none, margin: 2.5cm)[
  #set align(center)
  #v(2fr)
  #text(font: blackletter, size: 50pt, fill: ink)[Historyline]
  #v(-1.2em)
  #text(size: 18pt, tracking: 0.35em, fill: accent)[1914–1918]
  #v(1.5em)
  #line(length: 50%, stroke: 0.8pt + rule-colour)
  #v(0.8em)
  #text(size: 15pt, style: "italic")[Art book]
  #v(0.3em)
  #text(size: 10pt, fill: muted)[Masters, prompts and runtime art of the remake]
  #v(3fr)
  #set text(size: 8.5pt, fill: muted)
  #let n(kind) = index.assets.filter(a => a.kind == kind).len()
  #kinds.map(((k, t)) => [#n(k) #lower(t)]).join[ · ] \
  #index.assets.filter(a => a.status == "final").len() final,
  #index.assets.filter(a => a.status == "candidate").len() candidate \
  Generated from #raw("art/index.json") · #link(repo)[#repo.trim("/", at: end).split("//").last()]
]

// Contents
#page(header: none)[
  #align(center, text(font: blackletter, size: 22pt)[Contents])
  #v(0.8em)
  #set text(size: 9pt)
  #show outline.entry.where(level: 1): it => {
    v(0.8em)
    text(weight: "bold", fill: accent, it)
  }
  #outline(title: none, depth: 2, indent: 1.2em)
]

// Style guide
= Style
#if with-style {
  let src = read("/art/style.md")
  md(src)
} else {
  note[The style guide lives in #blob("art/style.md").]
}

// One page per art family.
#let section(title, src) = if src != "" {
  heading(level: 4, outlined: false, title)
  md(src)
}

#let asset-page(a, first: false) = {
  if not first { pagebreak(weak: true) }
  [#heading(level: 2, a.title)#art-label(a.id)]
  block(below: 0.9em)[
    #status-tag(a.status) #h(0.4em) #raw(a.id) #h(1fr) #small[#a.runtime.count runtime files
      #if a.runtime.densities.len() > 0 [· densities #a.runtime.densities.map(d => str(d) + "×").join(", ")]]
  ]

  // The first master beside the look and the repository paths, further masters below.
  let master(m, height) = figure(
    box(fill: paper-tint, inset: 3pt, picture(file(m.path), height: height)),
    caption: small[#raw(m.path.split("/").last()) · #m.width × #m.height],
    supplement: none, numbering: none,
  )
  grid(
    columns: (55%, 1fr),
    column-gutter: 1.2em,
    if a.masters.len() == 0 {
      box(width: 100%, height: 3cm, fill: paper-tint, stroke: (paint: rule-colour, dash: "dashed"),
        align(center + horizon, small[no master]))
    } else {
      master(a.masters.first(), 7cm)
    },
    {
      heading(level: 4, outlined: false)[Look]
      md(a.look)
      heading(level: 4, outlined: false)[Runtime IDs]
      text(size: 8pt, ids(a.runtime_ids))
      heading(level: 4, outlined: false)[Files]
      set text(size: 7.5pt)
      [Prompt: #blob(a.prompt) \ Masters: #tree(a.dir) \ Runtime: #tree(a.assets_dir)]
    },
  )
  if a.masters.len() > 1 {
    heading(level: 4, outlined: false)[More masters]
    grid(columns: (1fr,) * 5, gutter: 0.5em, ..a.masters.slice(1).map(m => master(m, 2.4cm)))
  }

  if a.runtime.samples.len() > 0 {
    heading(level: 4, outlined: false)[Runtime samples]
    let cols = calc.min(4, a.runtime.samples.len())
    grid(
      columns: (1fr,) * cols,
      gutter: 0.5em,
      ..a.runtime.samples.map(s => figure(
        box(fill: paper-tint, inset: 2pt, picture(file(s), height: if a.kind == "screen" { 4.5cm } else { 2.6cm })),
        caption: small(raw(s.trim(a.assets_dir + "/", at: start))),
        supplement: none, numbering: none,
      )),
    )
  }

  heading(level: 4, outlined: false)[Base prompt]
  block(width: 100%, inset: 0.7em, fill: paper-tint, stroke: (left: 2pt + accent),
    text(size: 8.5pt, raw(reflow(a.base_prompt), block: true)))

  if a.variants.len() > 0 {
    heading(level: 4, outlined: false)[Variants]
    table(
      columns: (auto, auto, 1fr),
      fill: (_, y) => if y == 0 { paper-tint },
      table.header[*Variant*][*Runtime IDs*][*Prompt delta*],
      ..a.variants.map(v => (inline(v.variant), text(size: 7.5pt, ids(v.runtime_ids)), md(v.delta))).flatten(),
    )
  }
  section([Edges & joins], a.edges)
  section([Post], a.post)
  if a.avoid.len() > 0 {
    heading(level: 4, outlined: false)[Avoid]
    list(..a.avoid.map(x => md(x)))
  }
}

#for (kind, title) in kinds {
  let assets = index.assets.filter(a => a.kind == kind)
  if assets.len() > 0 {
    heading(level: 1, title)
    for (i, a) in assets.enumerate() { asset-page(a, first: i == 0) }
  }
}
#if index.assets.len() == 0 {
  heading(level: 1)[Families]
  note[No art family has a #raw("prompt.md") yet; run #raw("just art-index") once they exist.]
}

// The PULSE level drawn with the runtime art (`hlart pulse`).
#if pulse != "" {
  let info = json(pulse + "/pulse.json")
  // A fractional height only resolves in the page flow, not inside a figure.
  let plate(file, caption) = {
    block(height: 1fr, width: 100%, below: 0.5em, align(center, image(pulse + "/" + file, height: 100%, fit: "contain")))
    align(center, small(caption))
  }
  heading(level: 1)[PULSE map]
  [
    Level #blob(info.level) drawn with the current runtime art at #(info.density)× (#(info.width * info.density) ×
    #(info.height * info.density) pixels, #info.hexes hexes), layered as in the game: context backgrounds,
    infrastructure, then the #info.units starting units facing north in their owner's colours. Fog, owner
    bars and interface are left out. Compositor gaps: #info.gaps, issues: #info.issues.
    Compare with the target look in the PULSE reference chapter.
  ]
  plate("pulse-units.png")[With the starting units]
  pagebreak()
  plate("pulse-ground.png")[Terrain and infrastructure only]
}

// PULSE reference design
#if with-reference {
  let ref = json("/art/reference/pulse/reference.json")
  heading(level: 1)[PULSE reference]
  [
    The reference painting of level #blob(ref.level) fixes the look of each terrain family. Every
    family below names its runtime terrain types or infrastructure prefixes and the exemplar cells
    (#raw("r<row>c<column>")) that show it best. The painting itself stays local
    (#raw(ref.image.path), #ref.image.width × #ref.image.height); the data is #blob("art/reference/pulse/reference.json"),
    checked by #tree("tools/reference-design").
  ]
  heading(level: 3)[Families]
  table(
    columns: (22%, 24%, 1fr),
    fill: (_, y) => if y == 0 { paper-tint },
    table.header[*Family*][*Runtime*][*Exemplars and notes*],
    ..ref.families.map(f => {
      let key = f.art.trim("art/", at: start)
      let runtime = if f.kind == "infra" { f.infra_prefixes.map(p => p + "*") } else { f.runtime_ids }
      (
        [*#f.id* #small(f.kind) \ #text(size: 7pt, if key in known { link(art-label(key), raw(key)) } else { tree(f.art) })],
        text(size: 7.5pt, ids(runtime)),
        [#text(size: 7.5pt, f.exemplars.map(raw).join(", ")) #if "notes" in f [\ #text(size: 8pt, f.notes)]],
      )
    }).flatten(),
  )
  heading(level: 3)[Features]
  for feat in ref.features {
    block(breakable: false, below: 0.8em)[
      *#feat.title* #h(0.3em) #small(raw(feat.id)) \
      #text(size: 7.5pt, fill: muted, feat.cells.map(raw).join(", ")) \
      #feat.notes
    ]
  }
}
