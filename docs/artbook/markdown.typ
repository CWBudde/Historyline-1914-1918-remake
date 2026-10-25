// A small Markdown reader for the art book: headings, paragraphs, bullet and numbered
// lists, pipe tables, code fences, and inline `code`, **bold**, _emphasis_ and [links](url).

#let inline-pattern = regex(
  "`[^`]+`|\*\*[^*]+\*\*|\b_[^_]+_\b|\[([^\]]+)\]\(([^)\s]+)\)",
)

#let inline(s) = {
  let out = []
  let pos = 0
  for m in s.matches(inline-pattern) {
    out += s.slice(pos, m.start)
    let t = m.text
    if t.starts-with("`") {
      out += raw(t.slice(1, -1))
    } else if t.starts-with("**") {
      out += strong(t.slice(2, -2))
    } else if t.starts-with("_") {
      out += emph(t.slice(1, -1))
    } else {
      let (body, url) = m.captures
      out += if url.starts-with("http") { link(url, body) } else {
        [#body (#raw(url))]
      }
    }
    pos = m.end
  }
  out + s.slice(pos)
}

// reflow joins hard-wrapped lines; blank lines stay paragraph breaks. Prompts are
// prose wrapped for the editor, so the book lets Typst wrap them instead.
#let reflow(s) = (
  s
    .split(regex("\n\s*\n"))
    .map(p => p.split("\n").map(l => l.trim()).join(" "))
    .join("\n\n")
)

#let bullet = regex("^([-*+]|\d+[.)])\s+")

#let special(t) = (
  t.starts-with("```")
    or t.starts-with("#")
    or t.starts-with("|")
    or t.starts-with(">")
    or t.match(bullet) != none
)

#let cells(row) = row.trim("|").split("|").map(c => c.trim())

#let md-table(rows) = {
  let body = rows.filter(r => cells(r).any(c => c.trim(regex("[-: ]")) != ""))
  if body.len() == 0 { return }
  let head = cells(body.at(0))
  let n = head.len()
  table(
    columns: n,
    fill: (_, y) => if y == 0 { rgb("#f3eee4") },
    table.header(..head.map(c => strong(inline(c)))),
    ..body
      .slice(1)
      .map(r => {
        let cs = cells(r)
        range(n).map(i => text(size: 8pt, inline(cs.at(i, default: ""))))
      })
      .flatten(),
  )
}

// md renders Markdown text. Headings become unnumbered headings from level `level`
// down; a leading level-1 heading (the document title) is dropped.
#let md(src, level: 3) = {
  let lines = src.replace("\r", "").split("\n")
  let i = 0
  let first = true
  while i < lines.len() {
    let t = lines.at(i).trim()
    if t == "" {
      i += 1
      continue
    }
    if t.starts-with("```") {
      let body = ()
      i += 1
      while i < lines.len() and not lines.at(i).trim().starts-with("```") {
        body.push(lines.at(i))
        i += 1
      }
      i += 1
      let lang = t.trim("`").trim()
      let code = body.join("\n")
      block(width: 100%, inset: 0.6em, fill: rgb("#f3eee4"), if lang
        in ("", "text") { raw(reflow(code), block: true) } else {
        raw(code, lang: lang, block: true)
      })
    } else if t.starts-with("#") {
      let hashes = t.position(" ")
      if hashes == none { hashes = t.len() }
      if not (first and hashes == 1) {
        heading(level: calc.min(6, level + hashes - 1), outlined: false, inline(
          t.slice(hashes).trim(),
        ))
      }
      i += 1
    } else if t.starts-with("|") {
      let rows = ()
      while i < lines.len() and lines.at(i).trim().starts-with("|") {
        rows.push(lines.at(i).trim())
        i += 1
      }
      md-table(rows)
    } else if t.starts-with(">") {
      let body = ()
      while i < lines.len() and lines.at(i).trim().starts-with(">") {
        body.push(lines.at(i).trim().trim(">", at: start).trim())
        i += 1
      }
      block(inset: (left: 0.8em), stroke: (left: 1.5pt + rgb("#b9ab94")), emph(
        inline(body.join(" ")),
      ))
    } else if t.match(bullet) != none {
      let numbered = not "-*+".contains(t.first())
      let items = ()
      while i < lines.len() and lines.at(i).trim() != "" {
        let u = lines.at(i).trim()
        let m = u.match(bullet)
        if m != none {
          items.push(u.slice(m.end))
        } else if special(u) {
          break
        } else {
          items.at(-1) = items.at(-1) + " " + u
        }
        i += 1
      }
      if numbered { enum(..items.map(inline)) } else {
        list(..items.map(inline))
      }
    } else {
      let body = ()
      while (
        i < lines.len()
          and lines.at(i).trim() != ""
          and (body.len() == 0 or not special(lines.at(i).trim()))
      ) {
        body.push(lines.at(i).trim())
        i += 1
      }
      par(inline(body.join(" ")))
    }
    first = false
  }
}
