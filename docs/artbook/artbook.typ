// Historyline 1914-1918 remake — art book. One page per art family from art/index.json
// (`just art-index`), the style guide art/style.md and the PULSE reference design.
// Build with `just artbook` (build/artbook/Historyline_Art.pdf).
//
// Inputs: base (the art tree, default the repository root), thumbs (directory of the
// optional master and sample thumbnails written by `hlart thumbs`; the normal book embeds native files), pulse
// (directory of the PULSE map written by `hlart pulse`; default: no map chapter),
// winter (directory of a winter level written by `hlart map`; default: no winter chapter),
// compare (directory written by `reference-design -compare`; default: no comparison), style=false
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
#let winter = sys.inputs.at("winter", default: "")
#let shore = sys.inputs.at("shore", default: "")
#let compare = sys.inputs.at("compare", default: "")

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
#let status-tag(s) = tag(s, fill: if s == "approved" { accent } else if s == "revised" { ink } else { muted })
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
  #("approved", "revised", "candidate").map(s => [#index.assets.filter(a => a.status == s).len() #s]).join[, ] \
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
#heading(level: 3)[Reading the artwork pages]
*Candidate* means generated artwork awaiting visual review. *Revised* means reworked against the
reference design. *Approved* means accepted. These are review states; each image keeps its own pixel
dimensions at every state.

*Source paintings* are the full-resolution master images. The primary master leads each family;
additional sources can be other materials, seasons or retained earlier drafts. *Game exports* are
the cut, scaled and composed images used by the renderer, shown at their highest available resolution.
*Appearance variants* describe terrain shapes, seasons, facings and other visual differences.

This edition embeds native source PNGs and game exports. The pixel dimensions printed with each
source describe the image in this PDF. PULSE uses the 4× runtime artwork; comparison images use PNG.
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

  // The first source painting beside the look and paths, further sources below.
  let master(m, height) = figure(
    box(fill: paper-tint, inset: 3pt, picture(file(m.path), height: height)),
    caption: small[#raw(m.path.split("/").last()) · #m.width × #m.height],
    supplement: none, numbering: none,
  )
  heading(level: 4, outlined: false)[Source paintings]
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
    heading(level: 4, outlined: false)[Additional master sources]
    grid(columns: (1fr,) * 5, gutter: 0.5em, ..a.masters.slice(1).map(m => master(m, 2.4cm)))
  }

  if a.runtime.samples.len() > 0 {
    heading(level: 4, outlined: false)[Game exports (highest available resolution)]
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
    heading(level: 4, outlined: false)[Appearance variants]
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
    #(info.height * info.density) pixels, #info.hexes hexes), layered as in the game: ground and cluster
    paintings, infrastructure, overhanging objects, then the #info.units starting units facing north in their owner's colours. Fog, owner
    bars and interface are left out. Compositor gaps: #info.gaps, issues: #info.issues.
    Compare with the target look in the PULSE reference chapter.
  ]
  plate("pulse-units.png")[With the starting units]
  pagebreak()
  plate("pulse-ground.png")[Terrain and infrastructure only]
}

#if winter != "" {
  let info = json(winter + "/map.json")
  let plate(file, caption) = {
    block(height: 1fr, width: 100%, below: 0.5em, align(center, image(winter + "/" + file, height: 100%, fit: "contain")))
    align(center, small(caption))
  }
  heading(level: 1)[Winter map — #info.name]
  [
    Level #blob(info.level) with its winter terrain and current shared infrastructure and units,
    drawn by the same compositor as PULSE at #(info.density)×
    (#(info.width * info.density) × #(info.height * info.density) pixels, #info.hexes hexes).
    The #info.units on-map starting units face north in their owner's colours.
    Compositor gaps: #info.gaps, issues: #info.issues. These are current candidate assets;
    winter artwork and seasonal infrastructure remain under review.
  ]
  plate("map-units.png")[#info.name — with the starting units]
  pagebreak()
  plate("map-ground.png")[#info.name — terrain and infrastructure only]
  pagebreak()
  heading(level: 2)[Winter terrain detail]
  [
    A central crop of the terrain rendering, with no resampling:
    #(info.detail.at(2) * info.density) × #(info.detail.at(3) * info.density) pixels.
    Snow, craters, roads, rail, trenches and the depot use the same exported tiles as the full map.
  ]
  plate("map-detail.png")[Native 4× detail — open the PNG to inspect individual tiles.]
}

#if shore != "" {
  heading(level: 1)[Reusable shoreline tiles]
  [
    Nine shore tiles above, six different compatible arrangements below. Every panel uses the same
    exported tile PNGs at their native 4× resolution. The six assemblies reuse the same central tile
    with different neighbours; land and water agree on shared edges and at three-tile corners.
    These are reusable terrain pieces. PULSE uses them too, without a lake-shaped painting.
    Shared borders retain their matching profile, while the interiors have asymmetric bends and
    rocky banks that widen and narrow.
    Submerged stones soften the water edge; moss and grass enter rock crevices along the meadow edge.
  ]
  figure(image(shore + "/shore-catalogue-4x.png", height: 21cm),
    caption: [All nine shores and six assemblies. Native PNG: 3144 × 3728 pixels.],
    supplement: none, numbering: none)
  pagebreak(weak: true)
  heading(level: 2)[Reusable river mouths]
  [
    A stream meeting a land edge opens a channel through the bank using a reusable mouth variant.
    Selection depends on the neighbouring stream's port, and the ordinary shore tile remains in use
    elsewhere. The supported set covers every fully-land edge of the nine shore shapes.
  ]
  figure(image(shore + "/shore-mouths-4x.png", width: 100%, height: 20cm),
    caption: [River-mouth pieces composed from the shared water, rock and meadow materials.],
    supplement: none, numbering: none)
}

// The reference painting beside the render (`reference-design -compare`, only with the local image).
#let art-status = (:)
#for a in index.assets { art-status.insert(a.id, a.status) }
#let family-art = (:)
#for f in json("/art/reference/pulse/reference.json").families { family-art.insert(f.id, f.art.trim("art/", at: start)) }
#let family-status(id) = {
  let key = family-art.at(id, default: "")
  if key in art-status { status-tag(art-status.at(key)) } else { small[no prompt] }
}

#if compare != "" {
  let cmp = json(compare + "/compare.json")
  let n(x, d: 1) = str(calc.round(x, digits: d))
  let outside(f) = if f.outside.len() == 0 { small[within] } else { text(fill: accent, weight: "bold", f.outside.join(", ")) }
  heading(level: 1)[Reference and render]
  [
    Left the reference painting registered onto the game grid, right #blob(cmp.level) drawn with the current
    runtime art. Metrics are measured at 1× over each family's exemplar hexes: mean CIELAB colour, contrast
    (standard deviation of L\*), chroma (mean C\*) and edges (share of pixels with a strong L\* gradient). The
    reference values are the targets in #blob("art/reference/pulse/targets.json"); the tolerance is in
    #blob("art/reference/pulse/reference.json").
  ]
  note[
    This is a candidate comparison. The map layout follows PULSE, while visual acceptance remains open.
    The PULSE lake uses reusable shore, water and water-stone tiles with shared materials.
    Compare its bank silhouettes, outlet and shoreline vegetation,
    canopy repetition, scattered meadow details, road and path edges, and the southern pair of craters. Higher pixel resolution
    preserves the authored detail; these composition differences still need review.
  ]
  table(
    columns: 8,
    align: (left, right, right, right, right, right, right, left),
    fill: (_, y) => if y == 0 { paper-tint },
    table.header[*Family*][*Status*][*ΔE*][*L\* a\* b\**][*Contrast*][*Chroma*][*Edges*][*Outside*],
    ..cmp.families.map(f => (
      text(size: 8pt, strong(f.id)),
      family-status(f.id),
      text(size: 8pt, n(f.delta_e)),
      text(size: 7pt)[#n(f.reference.l) #n(f.reference.a) #n(f.reference.b) \ #n(f.render.l) #n(f.render.a) #n(f.render.b)],
      text(size: 8pt)[#n(f.reference.contrast) → #n(f.render.contrast)],
      text(size: 8pt)[#n(f.reference.chroma) → #n(f.render.chroma)],
      text(size: 8pt)[#n(f.reference.edges, d: 2) → #n(f.render.edges, d: 2)],
      text(size: 8pt, outside(f)),
    )).flatten(),
  )
  heading(level: 3)[Full map]
  for b in cmp.bands {
    block(breakable: false, below: 0.6em, image(compare + "/" + b, width: 100%))
  }
  heading(level: 3)[Families]
  for f in cmp.families {
    block(breakable: false, below: 0.9em)[
      *#f.id* #h(0.4em) #family-status(f.id) #h(0.4em) #small[#f.exemplars.at(0), ΔE #n(f.delta_e)] #h(1fr) #outside(f)
      #image(compare + "/" + f.crops.at(0), width: 100%)
    ]
  }
  heading(level: 3)[Features]
  for ft in cmp.features {
    block(breakable: false, below: 0.9em)[
      *#ft.title* #h(0.3em) #small(raw(ft.id))
      #image(compare + "/" + ft.crop, width: 100%)
    ]
  }
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
        [*#f.id* #small(f.kind) #family-status(f.id) \ #text(size: 7pt, if key in known { link(art-label(key), raw(key)) } else { tree(f.art) })],
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

// Source catalogue follows the map comparison.
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
