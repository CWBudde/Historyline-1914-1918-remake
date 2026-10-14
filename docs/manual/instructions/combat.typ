#import "../template.typ": *

== Combat Position

Although a large number of tactical measures play an important role in Historyline 1914–1918, the most
important aspect is the outcome of battles between opposing units. The successful or unsuccessful outcome
of a battle depends largely on the combat position. In the following explanation, the unit carrying out
the action is referred to as the "attacker", whose target is the "defender".

The combat position is important both to the attacker and the defender. Basically, the position of the
attacker will have an effect on his attack values, and that of the defender will affect his defence values.

=== Pincer Movement

For the attacker, this means that every additional unit positioned around the defender which would also be
capable of attacking him is added to the attack values. Such a position is described as a _pincer
movement_. Thus, the pincer movement increases the attack value of the attacking unit. A unit which is
positioned directly behind the defender doubles the attack capability of the attacker; units at the sides
increase the value by a smaller amount. However, pincer movements have no effect on the attacking strength
of long-range weapons. The pincer movement will only be effective if the attack proceeds from one hexagon
to the next; it would be unfair to add 50% or 100% to the already high attack values of units such as the
heavy artillery.

Remember that a pincer movement will not transform a weak unit into some sort of wonder weapon; a transport
unit, for example, will seldom be able to inflict significant damage on a tank. Even when you have
completely surrounded the defender, you will need to use a strong unit to carry out the attack.

=== Blocking

How can the defender avoid being hemmed in? For one thing, it will not be easy for the attacker to move his
units near to or behind his opponent. Also, every additional one of the defender's units which is
positioned next to both the defender and the attacker increases his defence value. This tactic is referred
to as _blocking_. Like the pincer movement, blocking cannot prevent attacks with long-range weapons. The
basic defence value of the defender is again increased: each "blocking" unit increases the defence
capacity by one quarter.

=== Distance

The various artillery weapons played a very important role in the First World War. However, in general the
guns' accuracy decreased as their distance from the target increased, and this is also true in Historyline
1914–1918. Experience will show you the optimum distance from the target.

== Experience

The quality of a unit is dependent not only on its fire power, range and armour, but also on the unit's
battle experience. When one of your units is involved in a battle and inflicts damage on the enemy, the
unit will receive one experience point. If your unit manages to destroy a whole enemy unit completely, you
will receive two additional points. The battle experience is shown in the status line. The more experience
points a unit has gathered, the more dangerous it is to the enemy, as experience has a decisive effect on
the outcome of a battle. In addition to the "accuracy" of the unit being enhanced, its "skill" in avoiding
enemy fire is also increased.

Try to ensure that all your units gain sufficient experience! In the early stages of the game you should
not attack heavily armoured enemy units with units of your own which possess less fire power, as your unit
will not be able to destroy a single vehicle. If you are attacking one of the enemy's units which has
already been weakened, you should attack with a unit which is capable of destroying the enemy unit
completely. Your unit will then immediately receive two experience points.

== Unit Information

#beside("img-045-026.jpg", caption: [Unit information], width: 40%)[
  When you select the question mark above a unit, you will access an information display which gives you
  details of all the units. Under the large picture of the unit is some important technical information.
  Reading from the top, this shows:

  + Fire power and range against airborne targets.
  + Fire power and range against ground targets.
  + Fire power and range against naval targets.
  + Armour of unit.
  + Range of unit over ideal terrain.
  + Weight of unit (vital when loading on a transporter).
  + Maximum group strength.
]

You cannot obtain any information about your opponents' units.

== General Information <general-information>

#beside("img-046-027.jpg", caption: [An example of some general information], width: 40%)[
  To obtain precise information about the current state of play, select the question mark when the cursor
  is _not_ above a unit or building. You will then obtain a large amount of numerical information, as
  follows:

  + *Round* gives the number of previous command cycles which have already been used in the current game.
  + *Level* is the program's internal level number.
  + *Mode* shows which mode your command centre is currently in. If it shows "Move", you can order your
    troops to move; otherwise you can plan your attacks and other manoeuvres.
]

#[
#set enum(start: 4)
+ *Best* is the abbreviation for "Best Score", which shows the highest score achieved on this map.
+ *Present* is the current score.
+ The large information box provides information about the relative position of the players. You can see
  how many units, depots and factories belong to the two players and which can still be captured. Units,
  depots or factories which appear in the third column of the box do not belong to either player, but may
  be captured.
+ *Total energy* tells you how much energy is currently available to build and repair units.
+ *Energy credits* shows the amount of energy which will be credited to your account in each command cycle.
+ *Moves* shows the number of moves which you have already made.
+ *Limit* gives you information about the limitation on moves set in the main menu.
]

== Scoring

After playing for a long time, you need to know more than simply who won or lost; so each of the maps has a
list of the highest-scoring winners. If you want to play to gain the highest score, you must of course
understand the scoring system.

The defence values of all the units on the map added together are the starting point from which the score
is calculated. The current score can be seen in General Information. As soon as a unit disappears, whether
friend or foe, the score decreases. This means that your final score will be reduced when you destroy enemy
troops, but if you win the game by totally destroying all the enemy's units, you will gain extra points to
compensate. In addition, the options contained in the Game Values menu also have a positive effect on the
final score.

The points which can be gained are:

#align(center, block(width: 75%, keytable(
  [All units destroyed], [+500 points],
  [With concealed inventories], [+100 points],
  [Limit of 4 moves], [Points × 4],
  [Limit of 8 moves], [Points × 3],
  [Limit of 16 moves], [Points × 2],
  [*Maximum possible score*], [*32,500 points*],
)))

== General Strategy

Two important factors determine whether you win or lose: one is your "global strategy", the other is your
behaviour on the battle ground itself.

"Global strategy" means the sequence of capturing buildings. To determine how this is to be done, you must
evaluate when and with what losses the buildings can be captured. These factors will be affected by the
contents and the tactical position of the building. If, for example, a well-equipped factory can be
occupied by you before the enemy, and you will incur acceptable losses, you should capture it first. In
evaluating depots you should consider not only the units they contain but also their energy supply.

Despite the importance of buildings, success in small-scale battles is always decisive to the outcome of
the game. Your superiority is dependent on the position of your units. Both the offensive and defensive
capabilities of your units can be greatly enhanced by making use of geographical features. Some basic rules
are:

+ Avoid roads and flat, open countryside. Such terrain increases your offensive power to some extent, but
  it also greatly reduces your defensive capability.
+ Always try to have difficult terrain near the battle line and at your rear, so as to avoid being hemmed
  in.
+ Buildings which can be used to repair units should always be close at hand.
+ At the front line you should always have heavy defensive weapons or units which can defend themselves
  against an enemy attack. You should always ensure that anti-aircraft guns are available when being
  attacked from the air.
+ Your second line should comprise units with long-range weapons. If you do not have enough units of this
  sort, use units which have a long range or can travel over difficult terrain.
+ Your third line should comprise long-range weapons (e.g. artillery or armed trains).
+ You should always keep some long-range, large-capacity transport vehicles in reserve. If it becomes
  clear that your units cannot hold out, you can then quickly withdraw with the units you have left.
