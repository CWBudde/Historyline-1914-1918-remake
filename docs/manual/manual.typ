// Historyline 1914-1918 — reconstruction of the original 1992 Blue Byte manual.
// Build with `just manual` (see docs/manual/README.md).

#import "template.typ": *

#show: manual

// Cover
#page(header: none, footer: none, margin: 2cm)[
  #set align(center)
  #v(2fr)
  #text(font: blackletter, size: 46pt, fill: ink)[Historyline]
  #v(-1.2em)
  #text(size: 18pt, tracking: 0.35em, fill: accent)[1914–1918]
  #v(1.5em)
  #line(length: 50%, stroke: 0.8pt + rule-colour)
  #v(0.8em)
  #text(size: 13pt, style: "italic")[Instructions]
  #v(0.2em)
  #text(size: 11pt, fill: muted)[and Description of Units]
  #v(3fr)
  #text(size: 8pt, fill: muted)[Blue Byte, 1992 · reconstructed edition]
]

// Colophon
#page(header: none, footer: none)[
  #set text(size: 8pt, fill: muted)
  #set par(justify: false)
  #v(1fr)
  *About this edition.* This is a typeset reconstruction of the English manual and weapons booklet that
  shipped with _Historyline 1914–1918_ by Blue Byte (1992). It was made for the Historyline remake. The text
  follows the original. Obvious typing and translation errors have been corrected, and the chapters on
  system requirements and on installing from floppy disk for MS-DOS and Amiga have been left out.

  The keyboard, joystick and menu descriptions refer to the original game.

  #v(0.6em)
  Software, manual, text, graphics and names © Blue Byte 1991/1992. Pictures are reproduced from the
  original manual and are not part of the remake's source distribution.
]

// Contents
#page(header: none)[
  #align(center, text(font: blackletter, size: 20pt)[Contents])
  #v(0.8em)
  #set text(size: 9pt)
  #show outline.entry.where(level: 1): it => {
    v(0.9em)
    text(weight: "bold", fill: accent, it)
  }
  #outline(title: none, depth: 2, indent: 1.2em)
]

#part-title([Instructions], <instructions>)

#include "instructions/introduction.typ"
#include "instructions/menus.typ"
#include "instructions/controls.typ"
#include "instructions/orders.typ"
#include "instructions/world.typ"
#include "instructions/combat.typ"
#include "instructions/appendix.typ"

#part-title([Description of Units], <description-of-units>, subtitle: [The Weapons Manual])

#include "weapons/intro.typ"
#include "weapons/units.typ"
#include "weapons/bibliography.typ"
