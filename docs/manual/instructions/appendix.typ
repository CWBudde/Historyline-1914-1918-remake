#import "../template.typ": *

== Appendix

=== Use of the Keyboard

#grid(
  columns: (1fr, 1fr),
  column-gutter: 1.2em,
  [
    #heading(level: 4, outlined: false)[Cursor control: Player A]
    #keytable(
      [Cursor up], [Joystick up],
      [Cursor down], [Joystick down],
      [Cursor left], [Joystick left],
      [Cursor right], [Joystick right],
      [Space bar], [Fire button],
    )
  ],
  [
    #heading(level: 4, outlined: false)[Cursor control: Player B]
    #keytable(
      [D], [Joystick up],
      [C], [Joystick down],
      [X], [Joystick left],
      [V], [Joystick right],
      [Ctrl], [Fire button],
    )
  ],
)

The following additional options can be accessed by using the function keys:

#align(center, block(width: 75%, keytable(
  [F1], [Confirmation of mode change],
  [F2], [Music on/off],
  [F3], [Sounds on/off],
  [F4], [Displays program version number],
  [F5], [Joystick 1 on/off],
  [F6], [Joystick 2 on/off],
  [F7], [Mouse on/off],
  [F8], [Mouse speed],
  [F9], [Battle on/off],
)))

Keys F5/F6 can be used specifically to switch a joystick on or off, if the program's automatic joystick
recognition does not function (the cursor cannot be steered). This may happen if your hardware does not
keep within certain tolerances.

=== Two-Player Maps <two-player-maps>

The passwords to the two-player maps contained in Historyline 1914–1918 are:

#align(center, table(
  columns: 6,
  inset: (x: 0.7em, y: 0.4em),
  align: center,
  stroke: 0.4pt + rule-colour,
  ..(
    "TRACK", "HUSAR", "BEAST", "PLATE", "LIGHT", "SCROL",
    "VIRUS", "BISON", "DRUCK", "TROLL", "UBOOT", "DROID",
    "GRAND", "ROYAL", "WATER", "SKILL", "SKULL", "AUDIO",
    "SPELL", "CAMEL", "FLAGS", "STORY", "SCOUT", "GREEN",
  ).map(p => text(font: sans, size: 8.5pt, tracking: 0.08em, p)),
))

=== Glossary

/ Action Mode: In this mode a player can only order his units to attack. He can also have units repaired
  and produced.
/ Battle: In Historyline 1914–1918 a battle is a clash of several enemy units which results in fighting.
/ Blocking: A defensive position which increases the defensive values of units during battle.
/ Cursor: Normally, the cursor shows the position on your computer at which the next letter appears. In
  Historyline 1914–1918 the cursor is the means by which all the functions are carried out. It is similar
  to a mouse pointer, but is controlled not by a mouse but by the keyboard or joystick.
/ Depot: A type of supply base where units can be repaired.
/ Factory: A building in which units can be repaired or produced.
/ Firing Range: The distance over which a unit can attack.
/ Headquarters: A large building for each player. Units can be repaired at the Headquarters.
/ Inventory: Indicates the contents of a building or transporter.
/ Movement Mode: In this mode a player can only order his units to move.
/ Pincer: An attacking manoeuvre which considerably increases the offensive values of the units in battle.
/ Range: The maximum distance which a unit can travel over ideal terrain.
/ Score: At the end of the game the winner is given a score which reflects the quality of his playing. The
  highest scores on each map are stored.
/ Shop: See Inventory.
/ Tactical Overview: The Tactical Overview shows a section of the whole map on which all the units and
  geographical features are displayed. Each of the six-sided figures represents an area with particular
  features.
/ Unit: A group of up to six identical vehicles which are shown collectively on the tactical map.

== Credits

#let role(title, ..names) = block(breakable: false, below: 0.9em)[
  #text(size: 7.5pt, fill: muted, tracking: 0.06em, upper(title)) \
  #names.pos().join(linebreak())
]

#set align(center)
#grid(
  columns: (1fr, 1fr),
  column-gutter: 1.5em,
  [
    #heading(level: 3, outlined: false)[Main Program]
    #role("Production", "Blue Byte")
    #role("Head of Production", "Lothar Schmitt")
    #role("Programming and Design", "Ralf J. Kraft", "Bernhard Ewers", "Thomas Häuser")
    #role("Graphics", "Janos Toth")
    #role("Layout and Animation", "Thorsten Knop")
    #role("Selection of Units", "Thomas Hertzler")
    #role("Level Design", "Janos Toth", "Christoph Werner", "Thorsten Knop")
    #role("Music Programming", "Oliver Koenig")
    #role("Music", "Haiko Ruttmann")
  ],
  [
    #heading(level: 3, outlined: false)[Opening and Closing Credits]
    #role("Script and Design", "Bernhard Ewers", "Thorsten Knop")
    #role("Program", "Bernhard Ewers")
    #role("Graphics", "Thorsten Knop", "Christoph Werner")
    #role("Music Sound Track", "Haiko Ruttmann")

    #heading(level: 3, outlined: false)[Intermediate Sequences]
    #role("Idea", "Lothar Schmitt")
    #role("Research and Texts", "Ulrich Albert Springmann", "Ralf J. Kraft")
    #role("Program", "Bernhard Ewers", "Oliver Koenig")
    #role("Graphics", "Thorsten Knop", "Christoph Werner")
    #role("Music Sound Track", "Haiko Ruttmann")
  ],
)

#heading(level: 3, outlined: false)[Documentation]
#grid(
  columns: (1fr, 1fr, 1fr, 1fr),
  column-gutter: 0.8em,
  role("Instructions", "Lothar Schmitt"),
  role("Weapons Manual", "Thomas Hertzler"),
  role("Graphics", "Thorsten Knop", "Janos Toth"),
  role("Translation", "Polylang"),
)

#heading(level: 3, outlined: false)[Special Thanks To]
Hans "MC" Ippisch, Stephan Friedl, Frank Hasselmans, Armin Gessert and everyone who helped us in the
development of the "Historyline 1914–1918" project.
