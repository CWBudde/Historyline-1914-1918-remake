// Package campaign holds the campaign's calendar: one level per month from
// August 1914 to July 1918 (levels/monthly). Before each battle the player
// reads the month's front page of a neutral newspaper, newly written:
// the month's war in a few paragraphs, an illustration and a column of
// short items; and the page of the unit types new to their side (UnitNotes).
package campaign

// Article is the front page of a campaign month.
type Article struct {
	Label    string   // the issue's month: "August 1914"
	Headline string   // the page's headline
	Deck     string   // the line under the headline
	Caption  string   // under the illustration, art/screens/news/<month>.png
	Body     []string // the article's paragraphs
	Briefs   []string // the "In brief" column: news from elsewhere
}

// News is the front page of a month ("1914-08"); false for a month without
// one.
func News(month string) (Article, bool) {
	a, ok := Articles[month]
	return a, ok
}

// Articles are the front pages by month, "YYYY-MM".
var Articles = map[string]Article{
	"1914-08": {
		Label:    "August 1914",
		Headline: "War in Europe",
		Deck:     "Germany sweeps through Belgium; a Russian army is lost in the east",
		Caption:  "Our artist's sketch: German infantry pass a Belgian village as refugees look on",
		Body: []string{
			"The armies of Europe are on the march. On 4 August German troops cross the Belgian frontier, and by midnight Britain is at war. The ring of forts at Liège holds out for days until heavy siege howitzers batter it to rubble; on 20 August grey columns march through Brussels while families load what they can carry onto handcarts.",
			"Along the French frontier the clash is general. French offensives into Lorraine and the Ardennes break down with heavy losses, and at Mons on 23 August the British Expeditionary Force meets the German right wing before it too falls back. By the end of the month the Allied armies are retreating south, and the German columns press on towards Paris.",
			"In the east, Russia has moved faster than anyone expected, sending two armies into East Prussia. Near Tannenberg, Generals Hindenburg and Ludendorff encircle Samsonov's Second Army and take tens of thousands of prisoners. Each capital has a victory to celebrate and a front to fear for.",
		},
		Briefs: []string{
			"Serbia throws back the Austro-Hungarian invasion at Mount Cer",
			"Japan declares war on Germany and threatens Tsingtao",
			"Three German cruisers are sunk in a sea fight off Heligoland",
		},
	},
	"1914-09": {
		Label:    "September 1914",
		Headline: "The Miracle of the Marne",
		Deck:     "The German tide breaks east of Paris; both armies dig in on the Aisne",
		Caption:  "Paris taxicabs at dusk carry French reserves out to the battle",
		Body: []string{
			"The German armies have marched to within a few days of Paris, and the French government has left for Bordeaux. Then General von Kluck's First Army swings east of the capital, offering its flank. On 6 September General Joffre halts the long retreat and turns every army to attack along the Marne.",
			"From Paris, General Gallieni sends Maunoury's Sixth Army against that flank on the Ourcq. Reserves are rushed forward by every means, hundreds of the city's taxicabs among them. As Kluck turns to meet the blow, a gap opens between his army and its neighbour, and British and French troops push into it.",
			"On 9 September, fearing that its armies will be cut apart, the German command orders a retreat. The invaders fall back to the heights above the Aisne and dig in, and there the Allied pursuit stops before their machine guns. Paris is saved, but the war is not won: each side now reaches for the other's open northern flank.",
		},
		Briefs: []string{
			"Russians enter Lemberg; Austro-Hungarian armies fall back to the San",
			"U-9 sinks three British cruisers within an hour off the Dutch coast",
			"Moltke is replaced; Falkenhayn takes over the German high command",
		},
	},
	"1914-10": {
		Label:    "October 1914",
		Headline: "The Race to the Sea",
		Deck:     "Each army reaches for the other's flank until the Channel stops both",
		Caption:  "The Yser plain lies flooded by opened sluices; Belgian soldiers keep watch on a dike",
		Body: []string{
			"Since the Aisne, each army has tried to march round the other's northern flank, and each time it finds the enemy already there. Battles flare at Arras and near Lille, and the open flank creeps north towards the sea.",
			"Antwerp, Belgium's great fortress, falls on 10 October after its forts are pounded by heavy guns. King Albert's field army slips away along the coast and turns to stand on the little river Yser, on the last strip of Belgium still in its hands. For two weeks German divisions press the exhausted Belgians back towards the railway embankment between Nieuport and Dixmude.",
			"Then the sluices at Nieuport are opened. Sea water spreads slowly across the low fields until the plain before the embankment is a shallow lake, and the attack can go no further. The front settles behind the flood.",
			"Inland, around Ypres, British, French and German troops are now locked in a new and bitter battle for the last gap in the line.",
		},
		Briefs: []string{
			"Ottoman warships bombard Russian ports on the Black Sea",
			"The German drive on Warsaw is halted; the Ninth Army falls back",
			"The battleship Audacious strikes a mine and sinks off Ireland",
		},
	},
	"1914-11": {
		Label:    "November 1914",
		Headline: "The Wood at Nonne Bosschen",
		Deck:     "The Prussian Guard breaks into the wood, but the British line holds",
		Caption:  "Shattered woodland in November fog; troops advance among the stumps",
		Body: []string{
			"For three weeks the fighting around Ypres has not let up. German reserve corps, many of them young volunteers, have attacked again and again, and the thin British, French and Belgian lines have bent without breaking. On 11 November the German command makes one more attempt astride the Menin Road.",
			"After a heavy bombardment the Prussian Guard advances out of the morning mist. Part of it storms through the British trenches by Polygon Wood and into Nonne Bosschen, a small wood beyond them, with little but the British field guns ahead. The gunners fire at close range, and sappers fight as infantry.",
			"In the afternoon the Oxfordshire and Buckinghamshire Light Infantry counter-attack through the wood and drive the Guardsmen out. Rain and snow follow. By the third week of November both sides are spent, and the battle dies away.",
			"Ypres remains in Allied hands, inside a salient overlooked by German guns. The old British regular army is all but gone; the German volunteer regiments have suffered as terribly.",
		},
		Briefs: []string{
			"Off Coronel, von Spee's squadron sinks two British cruisers",
			"Britain and France declare war on the Ottoman Empire",
			"Tsingtao surrenders to Japanese and British forces",
		},
	},
	"1914-12": {
		Label:    "December 1914",
		Headline: "A Line of Trenches",
		Deck:     "From the sea to Switzerland, two armies dig in for the winter",
		Caption:  "Men of both armies meet in no man's land at Christmas; a fir tree stands on a parapet",
		Body: []string{
			"A single line of trenches now runs from the Belgian coast to the Swiss frontier. Rain fills them faster than pumps can empty them; men stand in water to the knee and build their sandbag walls higher. Both high commands plan how to break a front that offers no flank to turn.",
			"On Christmas Eve a hard frost settles over Flanders, and something no one ordered happens. In many places along the British line, German soldiers set small fir trees with candles on their parapets and sing carols. Voices answer from across the wire. By morning men climb out, meet in no man's land, shake hands and exchange tobacco, food and buttons.",
			"Officers arrange to bury the dead who have lain between the lines since autumn. In some places the quiet lasts a day, in others nearly a week. The Pope's appeal for a Christmas truce found no hearing among governments; for a little while, the soldiers make one of their own.",
			"Then the guns resume.",
		},
		Briefs: []string{
			"Off the Falklands, Sturdee's battlecruisers destroy von Spee's squadron",
			"German cruisers shell Scarborough, Hartlepool and Whitby",
			"Serbia retakes Belgrade and clears the invader from its soil",
		},
	},
	"1915-01": {
		Label:    "January 1915",
		Headline: "The Flood at Crouy",
		Deck:     "French gains north of Soissons are lost as the Aisne bursts its banks",
		Caption:  "The swollen Aisne and a broken bridge; soldiers wait on the bank in the rain",
		Body: []string{
			"North of Soissons the German line runs along the heights above the Aisne, close enough for its guns to command the town. On 8 January French troops attack Spur 132 near Crouy and, after hard fighting, take the crest and the trenches beyond it.",
			"Then the weather turns against them. Days of rain swell the Aisne until it bursts its banks, and the current sweeps away the light bridges behind the French line. Reinforcements and ammunition can no longer cross, and the men on the spur are left with the river at their backs.",
			"On 12 January General von Lochow's corps counter-attacks. The French are driven off the spur and back towards the flooded river, and by 14 January they are south of the Aisne again, having lost the ground they won and many men taken prisoner.",
			"Each side draws its lesson. In this war ground is worth only what can be supplied and held, and a river can decide a battle as surely as a gun.",
		},
		Briefs: []string{
			"Zeppelins bomb Great Yarmouth and King's Lynn in the first airship raid",
			"Off the Dogger Bank, British battlecruisers sink a German cruiser",
			"Enver Pasha's winter offensive in the Caucasus ends in disaster",
		},
	},
	"1915-02": {
		Label:    "February 1915",
		Headline: "War at Sea Declared",
		Deck:     "Berlin proclaims a U-boat war zone; Allied fleets bombard the Straits",
		Caption:  "Allied battleships in line bombard the forts at the mouth of the Dardanelles",
		Body: []string{
			"On 4 February Berlin declares the waters around Britain and Ireland a war zone. From 18 February, it warns, enemy merchant ships found there will be sunk, and neutral ships sail at their own risk. Washington replies that it will hold Germany to strict account for any American lives lost.",
			"Far to the south-east, the Allies try to open another road. On 19 February a British and French squadron under Admiral Carden begins to bombard the forts at the entrance to the Dardanelles. The battleships fire from long range and then close in; storms halt the work for days, but by the end of the month landing parties are going ashore to wreck the guns.",
			"If the fleet can force the Straits and reach Constantinople, London hopes, Turkey may be driven from the war and a sea road opened to Russia's Black Sea ports. The inner forts, the mobile batteries and the minefields of the Narrows still lie ahead.",
		},
		Briefs: []string{
			"A Turkish attack on the Suez Canal is beaten off",
			"Masuria: a Russian corps surrenders in the Augustow forest",
			"The First Canadian Division lands in France",
		},
	},
	"1915-04": {
		Label:    "April 1915",
		Headline: "Gas on the Western Front",
		Deck:     "A chlorine cloud opens a gap at Ypres; the liner Lusitania is sunk",
		Caption:  "A low cloud rolls over no man's land; soldiers press cloth pads to their faces",
		Body: []string{
			"Late in the afternoon of 22 April, on the northern side of the Ypres salient, German engineers open thousands of cylinders along their front. A yellow-green cloud drifts on the breeze towards the French lines held by territorial and Algerian troops. It is chlorine gas, used on this scale for the first time.",
			"Men choke and fall back, and within an hour a gap of several miles has opened. German infantry follow cautiously behind the cloud, but no great reserves stand ready. Canadian and British battalions are flung into the breach; two days later, at St Julien, the Canadians face the gas themselves, with soaked cloths pressed to their faces.",
			"The salient shrinks but holds. The Allies denounce the attack as a breach of the laws of war.",
			"On 7 May, off the south coast of Ireland, a U-boat torpedoes the Cunard liner Lusitania. She sinks in eighteen minutes; nearly 1,200 people are lost, more than a hundred of them Americans, and anger in the United States runs high.",
		},
		Briefs: []string{
			"Allied troops land on the Gallipoli peninsula under fire",
			"German and Austrian armies break the Russian front at Gorlice",
			"Asquith forms a coalition government in London",
		},
	},
	"1915-06": {
		Label:    "June 1915",
		Headline: "Italy Joins the Entente",
		Deck:     "Italy goes to war with Austria-Hungary and attacks on the Isonzo",
		Caption:  "Alpine troops haul a field gun up a mountain path above the Isonzo",
		Body: []string{
			"Italy is at war. On 23 May, after months of bargaining with both camps, Rome declares war on Austria-Hungary, its partner in the old Triple Alliance. With Germany it remains, for now, formally at peace. A new front opens along the mountains from the Swiss frontier to the Adriatic.",
			"General Cadorna, the Italian chief of staff, aims his main blow at the Isonzo valley, the road to Gorizia and Trieste. The Austro-Hungarian defenders, fewer in number but entrenched on the heights and the stony Carso plateau, wait above the river. On 23 June the first great Italian offensive opens; Alpine troops drag guns up mule paths, and the infantry climb into machine-gun fire.",
			"Vienna must now fight on three fronts, against Serbia, Russia and Italy. Yet in Galicia its fortunes turn: Mackensen's Austro-German armies retake Przemyśl on 3 June and Lemberg on 22 June, and the Russians fall back towards their own frontier.",
		},
		Briefs: []string{
			"A British airman destroys a Zeppelin in the air over Belgium",
			"Gallipoli: another Allied attack at Krithia gains little ground",
			"The French offensive in Artois ends with heavy losses",
		},
	},
	"1915-08": {
		Label:    "August 1915",
		Headline: "The Great Retreat",
		Deck:     "Warsaw falls and Russia's armies fall back across Poland",
		Caption:  "Carts and soldiers on a dusty eastern road; burning villages smoke on the horizon",
		Body: []string{
			"On 5 August German troops enter Warsaw. All summer the Austro-German armies have pressed the Russians from the south and west, and now the great salient of Russian Poland can no longer be held. The fortresses fall one after another: Kovno, Novogeorgievsk with its whole garrison, and on 26 August Brest-Litovsk.",
			"The Russian armies are short of shells and even of rifles, but they are not destroyed. They fall back eastwards, and as they go they burn crops, villages and bridges and drive the people before them. The roads east are choked with carts, cattle and refugees in their hundreds of thousands.",
			"For the Central Powers it is the greatest victory of the war so far. Yet every mile forward lengthens their supply lines over ruined roads, and Russia, though she has lost Poland, still stands. In Petrograd there is word that the Tsar means to take command of his armies himself.",
		},
		Briefs: []string{
			"Gallipoli: new British landings at Suvla Bay fail to break out",
			"Italy declares war on the Ottoman Empire",
			"A U-boat sinks the White Star liner Arabic off Ireland",
		},
	},
	"1915-10": {
		Label:    "October 1915",
		Headline: "A New Front at Salonika",
		Deck:     "Allied troops land in Greece as Serbia is attacked from two sides",
		Caption:  "Allied troops disembark at the Salonika quay from transports and boats",
		Body: []string{
			"On 5 October British and French troops begin to land at Salonika, in neutral Greece. They come at the invitation of Prime Minister Venizelos, who wants Greece to stand by Serbia; but King Constantine prefers neutrality, and Venizelos resigns the same day.",
			"Their goal is Serbia, now in mortal danger. On 6 October German and Austro-Hungarian armies under Mackensen cross the Danube and the Sava, and Belgrade falls within days. A week later Bulgaria joins the Central Powers and strikes from the east, cutting the railway up the Vardar by which the Allies hoped to come.",
			"General Sarrail's French divisions push north into Serbian Macedonia, but they are too few and too late. Caught between three armies, the Serbian army falls back through the mud of late autumn, fighting as it goes.",
			"In Salonika the camps spread over the hills, and the Allies weigh whether a front so far from home is a lifeline or a trap.",
		},
		Briefs: []string{
			"Nurse Edith Cavell is shot in Brussels for helping soldiers escape",
			"Italy opens a third offensive on the Isonzo",
			"Briand forms a new French government",
		},
	},
	"1915-12": {
		Label:    "December 1915",
		Headline: "Gallipoli Abandoned",
		Deck:     "Anzac and Suvla are emptied by night almost without loss",
		Caption:  "Boats slip away from a beach and its piers by night; ships wait offshore",
		Body: []string{
			"Eight months after the landings, the Allies give up the fight for the Dardanelles. Neither the fleet nor the armies have broken through, winter storms have wrecked piers and flooded trenches, and the troops are needed elsewhere. Lord Kitchener, having seen the peninsula for himself, agrees to evacuate.",
			"The withdrawal from Anzac and Suvla is a work of quiet cunning. Night after night men are taken off in silence while the trenches are made to look fully manned, and rifles rigged to fire on their own keep up a steady crackle. On the night of 19–20 December the last parties walk down to the beaches, and by dawn the boats are gone.",
			"Cape Helles is still held, but its turn is expected. For the Ottoman army, which has defended the peninsula with stubborn courage, it is a great victory; for the Allies, a costly failure. The road to Russia's Black Sea ports stays closed, and the divisions freed on both sides will soon be heard of again.",
		},
		Briefs: []string{
			"Haig succeeds Sir John French in command of the BEF",
			"Ottoman troops besiege Townshend's force at Kut al-Amara",
			"The Serbian army reaches the Adriatic after crossing the mountains",
		},
	},
	"1916-02": {
		Label:    "February 1916",
		Headline: "Verdun",
		Deck:     "Germany attacks the fortress city; Douaumont falls in four days",
		Caption:  "An endless convoy of lorries grinds along the road to Verdun in winter mud",
		Body: []string{
			"Early on 21 February more than a thousand German guns open fire on the French lines north of Verdun. For some nine hours the shells fall on woods and trenches; then the infantry of the Crown Prince's Fifth Army go forward over the snow. The attack has been held back for days by bad weather.",
			"In the Bois des Caures, Colonel Driant's chasseurs hold out for two days before they are overwhelmed. On 25 February a small German party walks almost unopposed into Fort Douaumont, the strongest fort of the ring. That night General Pétain is given command of the defence.",
			"His first concern is the road. Only one road and a narrow railway link Verdun with the rest of France, and along the road from Bar-le-Duc lorries now run nose to tail, day and night, bringing men, shells and food. Territorials shovel stone under the wheels to keep it open.",
			"The German attack slows but does not stop. On both banks of the Meuse the battle is only beginning.",
		},
		Briefs: []string{
			"Russian troops storm Erzurum in the Caucasus",
			"German forces in Cameroon cross into neutral Spanish Guinea",
			"Military service becomes compulsory for single men in Britain",
		},
	},
	"1916-04": {
		Label:    "April 1916",
		Headline: "Rebellion in Dublin",
		Deck:     "An Easter rising is put down in Dublin; Kut surrenders on the Tigris",
		Caption:  "The General Post Office in Dublin stands shelled and smoking",
		Body: []string{
			"On Easter Monday, 24 April, armed Irish Volunteers and men of the Citizen Army seize the General Post Office and other strongpoints in central Dublin. From the steps of the Post Office Patrick Pearse reads a proclamation of an Irish Republic. The city, taken by surprise on a holiday, is stunned.",
			"British troops pour in. Artillery and a gunboat on the Liffey shell rebel positions, and the fighting goes on street by street. By Friday the Post Office is ablaze and its garrison breaks out; on 29 April Pearse surrenders to spare further loss of life. Much of Sackville Street lies in ruins, and hundreds of civilians are among the dead and wounded.",
			"On the same day, far away on the Tigris, General Townshend surrenders at Kut al-Amara. Besieged since December, his British and Indian force has held out for nearly five months while relief columns tried and failed to break through. Hunger has done what assaults could not; some thirteen thousand men go into captivity.",
		},
		Briefs: []string{
			"Russian troops take Trebizond on the Black Sea",
			"Wilson warns Berlin over the U-boat war after the Sussex is torpedoed",
			"At Verdun the battle turns on the Mort Homme and Hill 304",
		},
	},
	"1916-06": {
		Label:    "June 1916",
		Headline: "Jutland and the Somme",
		Deck:     "The fleets meet in the North Sea as Russia strikes in Galicia",
		Caption:  "Our artist's sketch: battleships in line, their guns flashing through the North Sea haze",
		Body: []string{
			"The meeting of the great fleets comes at last off Jutland on 31 May. Beatty's battlecruisers find Hipper's, and behind each steams a battle fleet: Jellicoe's Grand Fleet and Scheer's High Seas Fleet. Twice the lines meet in haze and failing light, and twice Scheer turns away; in the night he slips behind the British and home.",
			"Britain loses three battlecruisers and more men, Germany fewer ships. Berlin speaks of victory, London of a fleet that still holds the sea. Neither seeks a second meeting, and the blockade stays.",
			"On 4 June General Brusilov's armies attack along a wide front in Galicia and Volhynia, without the usual long warning, and the Austro-Hungarian lines give way. Vienna must recall divisions from Italy, Berlin from the west.",
			"On the Somme a week-long bombardment ends on the morning of 1 July, and British and French infantry go over the top. The French advance; the British, meeting wire and machine guns the guns did not silence, suffer the heaviest single day's losses in their army's history.",
		},
		Briefs: []string{
			"Lord Kitchener is lost when HMS Hampshire sinks off Orkney on 5 June",
			"At Verdun, Fort Vaux surrenders on 7 June when its water runs out",
			"In the Hejaz, Sharif Hussein of Mecca rises against Ottoman rule",
		},
	},
	"1916-08": {
		Label:    "August 1916",
		Headline: "The Tank Goes into Battle",
		Deck:     "Armoured machines lead the attack at Flers; Romania joins the war",
		Caption:  "Our artist's sketch: a British tank crawls over the shell holes, infantry close behind",
		Body: []string{
			"On 27 August Romania declares war on Austria-Hungary and sends her armies through the Carpathian passes into Transylvania. Germany answers with a declaration of her own, and German and Bulgarian troops under Mackensen gather on the Danube to strike at the Dobruja.",
			"The news costs General von Falkenhayn his post. Field Marshal von Hindenburg becomes Chief of the General Staff with General Ludendorff at his side, and Falkenhayn goes to command against the Romanians.",
			"On the Somme the summer passes in costly struggles for woods and villages: Delville Wood, Guillemont, Ginchy. On 15 September the British attack between Flers and Courcelette with a weapon kept secret until that morning: armoured machines on caterpillar tracks, called tanks to hide their purpose.",
			"Many break down or stick in the craters. Those that go forward crush the wire and silence machine guns, and Flers falls with a tank leading the way down its street; the Canadians take Courcelette. The machines are slow and few, but a new kind of warfare has begun.",
		},
		Briefs: []string{
			"Italian troops take Gorizia on the Isonzo on 9 August",
			"Italy declares war on Germany on 28 August",
			"A German airship is shot down in flames at Cuffley, in sight of London",
		},
	},
	"1916-10": {
		Label:    "October 1916",
		Headline: "An Emperor Dies",
		Deck:     "Franz Joseph is dead after 68 years on the throne; Karl succeeds",
		Caption:  "Vienna mourns: the horse-drawn hearse of Franz Joseph passes through silent crowds",
		Body: []string{
			"On 24 October French infantry, advancing in fog behind a creeping barrage, retake Fort Douaumont at Verdun, lost in the first week of the German attack in February. Fort Vaux is reoccupied early in November. Ground that cost months to lose is won back in days.",
			"On the Somme the autumn rains turn the battlefield to mud. A last British effort takes Beaumont-Hamel on the Ancre in mid-November, and then the offensive stops. Since 1 July the Allies have advanced a few miles, at the cost of hundreds of thousands of men on both sides.",
			"In Romania, German and Austro-Hungarian troops force the mountain passes while Mackensen takes the port of Constanța. The Romanian army, driven out of Transylvania, falls back towards its capital.",
			"On 21 November Emperor Franz Joseph dies in Vienna at the age of 86. He came to the throne in 1848; few of his subjects remember another ruler. His great-nephew Karl succeeds him, a young man who is said to long for an end to the war.",
		},
		Briefs: []string{
			"President Wilson is re-elected on 7 November: 'he kept us out of war'",
			"Serbian and French troops retake Monastir in Macedonia on 19 November",
			"Austria's prime minister is shot dead in a Vienna restaurant on 21 October",
		},
	},
	"1916-12": {
		Label:    "December 1916",
		Headline: "Peace Notes, No Peace",
		Deck:     "Berlin offers talks, Washington asks for terms, the Entente refuses",
		Caption:  "Our artist's sketch: sentries keep watch in a snow-covered trench",
		Body: []string{
			"On 6 December German troops enter Bucharest. Most of Romania, with its wheat and oil, is now in the hands of the Central Powers; the Romanian army and government hold out in Moldavia.",
			"Six days later, on 12 December, Germany and her allies propose through neutral governments that peace talks begin. Their note names no terms. On 18 December President Wilson asks every belligerent to state the aims for which it fights.",
			"The Entente answers Berlin at the end of the month: an offer without terms, it says, is no offer but a manoeuvre. Its reply to Wilson, in January, demands the restoration of Belgium and Serbia and the evacuation of occupied lands. Neither camp expects the other to accept, and both prepare for a harder year.",
			"In France General Nivelle, whose methods retook the Verdun forts this autumn, replaces Joffre in command of the French armies. He promises a breakthrough in the spring, within forty-eight hours of its opening. In the trenches, sentries keep watch in the snow.",
		},
		Briefs: []string{
			"Lloyd George succeeds Asquith as British prime minister on 7 December",
			"A last French attack at Verdun on 15 December pushes the line back further",
			"Rasputin, confidant of the Tsar's family, is murdered in Petrograd",
		},
	},
	"1917-02": {
		Label:    "February 1917",
		Headline: "Revolution in Petrograd",
		Deck:     "The Tsar abdicates as Germany opens unrestricted U-boat warfare",
		Caption:  "Our artist's sketch: crowds and soldiers fill a snowy avenue in Petrograd",
		Body: []string{
			"On 1 February Germany declares that her U-boats will sink without warning any ship in the waters around Britain and France. Berlin wagers that Britain can be starved before America can matter; Washington breaks off relations on 3 February. On 1 March American papers print a German telegram proposing an alliance with Mexico should the United States enter the war.",
			"In Petrograd, queues for bread turn into strikes and marches in early March. Troops ordered to clear the streets refuse, and regiment after regiment joins the crowds. A Provisional Government forms from the Duma, while workers and soldiers elect a Soviet.",
			"On 15 March Tsar Nicholas II abdicates. The new government promises to fight on beside its allies, but whether the army will follow is asked in every capital.",
			"In France the Germans quietly leave the great bulge between Arras and Soissons for a new fortified line, the Hindenburg Line, destroying roads, bridges and villages behind them. The Allies' spring plans must be redrawn.",
		},
		Briefs: []string{
			"British troops enter Baghdad on 11 March",
			"Kut al-Amara, lost last April, is retaken by the British in February",
			"In East Africa, German forces under Lettow-Vorbeck fight on",
		},
	},
	"1917-04": {
		Label:    "April 1917",
		Headline: "America Declares War",
		Deck:     "The United States joins the war; on the Aisne, France's hopes fail",
		Caption:  "Our artist looks ahead: a troopship leaves harbour, soldiers crowding every rail",
		Body: []string{
			"On 2 April President Wilson asks Congress for a declaration of war against Germany, saying that the world must be made safe for democracy; on 6 April it is voted. America has a great fleet and greater factories, but only a small army. It will be many months before her divisions can stand in the line.",
			"On 9 April the British attack at Arras in snow and sleet. The four Canadian divisions, fighting together for the first time, storm Vimy Ridge, held by the Germans since 1914, and take the crest within hours.",
			"On 16 April General Nivelle opens his offensive on the Aisne, along the Chemin des Dames. The breakthrough promised within forty-eight hours does not come; the French gain ground at heavy cost.",
			"In May Nivelle is replaced. In one division after another French soldiers refuse to attack again, though they will hold their trenches. Their new commander promises leave, rest and better food, and to wait for the tanks and the Americans.",
		},
		Briefs: []string{
			"Lenin returns to Petrograd from exile in Switzerland on 16 April",
			"U-boats sink more shipping in April than in any month of the war so far",
			"General Pershing is named to command the American forces for France",
		},
	},
	"1917-06": {
		Label:    "June 1917",
		Headline: "The Mines of Messines",
		Deck:     "Nineteen mines open the battle for the ridge south of Ypres",
		Caption:  "Our artist's sketch: columns of earth rise along Messines Ridge at dawn",
		Body: []string{
			"At ten past three on the morning of 7 June, nineteen great mines explode under the German lines on the Messines Ridge, south of Ypres. For more than a year British, Australian and Canadian tunnellers have dug towards them under the ridge. The blast is said to be heard in London.",
			"General Plumer's Second Army, with British, Irish, Australian and New Zealand divisions, advances behind a heavy barrage and holds the crest within hours. German counter-attacks fail. It is a model of careful preparation.",
			"Then comes a pause. Haig's main Flanders offensive does not open until 31 July, after weeks of bombardment have wrecked the drainage of the low ground. On the first day the line moves forward on the left; in the afternoon the rain begins, and it does not stop for days.",
			"In Galicia the Russian offensive ordered by Kerensky opens on 1 July and gains ground for a few days, then collapses under the German counter-attack.",
		},
		Briefs: []string{
			"Gotha bombers raid London by daylight on 13 June",
			"The first American troops land in France at St Nazaire in late June",
			"The Reichstag votes for a peace without annexations on 19 July",
		},
	},
	"1917-08": {
		Label:    "August 1917",
		Headline: "Mud",
		Deck:     "Rain drowns the Flanders offensive; Riga falls to a new attack",
		Caption:  "A duckboard track winds across a flooded crater field among shattered trees",
		Body: []string{
			"August in Flanders is the wettest for years. The low ground east of Ypres, its drains destroyed by shellfire, becomes a field of water-filled craters joined by mud. Men move up along duckboard tracks; a slip from them can drown a laden man. Attacks at Langemarck in mid-August gain little.",
			"In September General Plumer takes over the main effort and changes the method: short steps of a mile or less behind enormous barrages, then a halt to hold the ground against counter-attack. In dry weather the Menin Road ridge and Polygon Wood fall in this way, on 20 and 26 September.",
			"In the east, on 1 September, General von Hutier's Eighth Army crosses the Dvina above Riga after a short, sudden bombardment of gas and high explosive. Its infantry push through the gaps in small groups rather than wait for the line to advance together. Riga falls on 3 September. Observers in both camps expect to see these methods again in the west.",
		},
		Briefs: []string{
			"Pope Benedict XV appeals to the warring powers for a just, lasting peace",
			"Italy's eleventh Isonzo battle gains ground on the Bainsizza plateau",
			"French troops retake the Mort-Homme at Verdun on 20 August",
		},
	},
	"1917-10": {
		Label:    "October 1917",
		Headline: "Caporetto",
		Deck:     "The Italian front breaks on the Isonzo; the Bolsheviks seize power",
		Caption:  "Our artist's sketch: Italian columns cross a mountain bridge in the rain, in retreat",
		Body: []string{
			"On 24 October, after a short bombardment with gas, German and Austro-Hungarian divisions attack on the upper Isonzo near Caporetto. In fog and rain they pass along the valleys while the Italians still hold the heights, and within days the whole front is coming apart. Hundreds of thousands are taken prisoner; the roads west fill with soldiers and refugees.",
			"The Italians fall back to the Tagliamento, then to the Piave, where the line holds in November. General Diaz replaces Cadorna, and British and French divisions are sent to Italy.",
			"On 7 November the Bolsheviks seize power in Petrograd. The next day Lenin's government calls on every nation at war for an immediate peace.",
			"In Flanders the Canadians take the ruins of Passchendaele on 6 November, and the offensive ends. On 20 November near Cambrai the British attack without a long bombardment, hundreds of tanks in the lead, and gain more in a day than in months at Ypres; church bells ring in Britain.",
		},
		Briefs: []string{
			"British forces take Beersheba on 31 October; Gaza falls a week later",
			"The Balfour Declaration backs a Jewish national home in Palestine",
			"Clemenceau, aged 76, becomes prime minister of France on 16 November",
		},
	},
	"1917-12": {
		Label:    "December 1917",
		Headline: "Russia Leaves the War",
		Deck:     "An armistice silences the Eastern Front; Germany looks west",
		Caption:  "Our artist's sketch: troop trains steam westward through the snow",
		Body: []string{
			"On 15 December, at Brest-Litovsk, Russian delegates sign an armistice with Germany and her allies, and peace talks open in the same town a week later. The Bolsheviks speak of a peace without annexations; the German generals speak of Poland, Lithuania and Courland, which their armies already hold.",
			"In January Trotsky arrives to lead the Russian delegation and draws out the talks, hoping that revolution will spread to the workers of Germany and Austria. Great strikes in Vienna and Berlin show that the hope is not empty, but they end without changing the war.",
			"Meanwhile the railways run west. Division after division leaves the Eastern Front for France and Belgium, where Germany will soon have more divisions than her enemies. The race is plain to both sides: Germany must strike before the Americans arrive in strength.",
			"On 8 January President Wilson sets out his Fourteen Points for the peace, among them open diplomacy, freedom of the seas and an association of nations.",
		},
		Briefs: []string{
			"General Allenby enters Jerusalem on foot on 11 December",
			"A munitions ship explodes in the harbour of Halifax, Nova Scotia",
			"Romania, alone in the east, signs an armistice on 9 December",
		},
	},
	"1918-02": {
		Label:    "February 1918",
		Headline: "The Kaiser's Battle",
		Deck:     "Peace in the east, and the greatest German offensive in the west",
		Caption:  "Our artist's sketch: storm troops move forward through the fog behind the barrage",
		Body: []string{
			"At Brest-Litovsk Trotsky declares in February that Russia will neither sign nor fight. On 18 February the German armies advance again, almost unopposed, and on 3 March the Bolsheviks sign. Russia gives up Poland, the Baltic lands, Finland and Ukraine.",
			"In the west the German army now has more divisions than its enemies. At 4.40 on the morning of 21 March some six thousand guns open fire on the British front around St Quentin. Five hours later, in thick fog, storm troops advance, passing strongpoints by and pushing deep into the defences.",
			"General Gough's Fifth Army is driven back across the old Somme battlefield. In days the Germans win more ground than either side has won in the west since 1914. On 26 March, at Doullens, the Allies give General Foch the coordination of their armies.",
			"Before Amiens the advance slows, its troops tired and far from their guns. Ludendorff has won ground, but not the decision he needs before the Americans come.",
		},
		Briefs: []string{
			"Ukraine signs a separate peace with the Central Powers on 9 February",
			"British troops take Jericho on 21 February",
			"From 23 March a giant gun shells Paris from some 75 miles away",
		},
	},
	"1918-04": {
		Label:    "April 1918",
		Headline: "Storm in Flanders",
		Deck:     "A second German blow falls on the Lys; British sailors raid Zeebrugge",
		Caption:  "Night at Zeebrugge: blockships at the mole under searchlights and star shells",
		Body: []string{
			"On 9 April a second German offensive breaks on the river Lys, south of Ypres, and overwhelms a Portuguese division in the line. Within days the British give up the Passchendaele ridge, won so dearly in the autumn, and the railway centre of Hazebrouck lies open.",
			"On 11 April Field Marshal Haig tells his troops that, with their backs to the wall, each must fight on to the end. The line holds before Hazebrouck. Mount Kemmel falls on 25 April, but by the end of the month the offensive is spent.",
			"On 14 April General Foch is named commander-in-chief of the Allied armies in France. American divisions now cross the Atlantic in growing numbers.",
			"On the night of 22–23 April British naval forces raid Zeebrugge, the outlet through which U-boats based at Bruges reach the sea. While the old cruiser Vindictive lands men on the mole, blockships steam in under fire and are sunk in the canal entrance. How long the channel stays closed remains to be seen.",
		},
		Briefs: []string{
			"The Royal Air Force is formed on 1 April from the army and navy air arms",
			"Manfred von Richthofen, the 'Red Baron', is killed in action on 21 April",
			"At Villers-Bretonneux tanks fight tanks for the first time, on 24 April",
		},
	},
	"1918-06": {
		Label:    "June 1918",
		Headline: "The Turning of the Tide",
		Deck:     "The last German offensive fails; on 18 July the Allies strike back",
		Caption:  "Our artist's sketch: small Renault tanks leave a forest edge through the wheat",
		Body: []string{
			"June finds German troops on the Marne once more, after their May offensive on the Aisne. At Belleau Wood American Marines fight through the month to clear the woods, and a further German thrust on the Matz is stopped.",
			"On 15 July Ludendorff attacks on both sides of Reims. East of the city General Gouraud's French army, warned in advance, has left its front line nearly empty, and the barrage falls on nothing. West of it German troops cross the Marne but can go no further.",
			"On 18 July the Allies strike back. With no long bombardment, General Mangin's Tenth Army advances out of the forest of Villers-Cotterêts against the flank of the salient, led by hundreds of tanks, many of them small new Renaults, with two American divisions in its line. The Germans must give up the Marne pocket.",
			"Since March, German armies have won more ground than in any year since 1914, and lost men they cannot replace. Now the initiative passes to Foch, and few expect it to return.",
		},
		Briefs: []string{
			"Austria-Hungary's offensive across the Piave fails in late June",
			"The former Tsar and his family are killed at Yekaterinburg in mid-July",
			"Influenza spreads through the armies and cities of both sides",
		},
	},
}
