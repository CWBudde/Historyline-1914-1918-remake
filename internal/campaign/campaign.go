// Package campaign holds the original campaign's calendar: each campaign
// (German, French, two-player) runs through the same 24 periods of two
// months, from August 1914 to July 1918, one mission per period. The news
// texts are newly written one-paragraph summaries of the period's events,
// shown between missions as in the original.
package campaign

// Period is one two-month layer of the war.
type Period struct {
	Label    string   // e.g. "August – September 1914"
	Headline string   // the news page's headline
	Lines    []string // the news page's short article lines
}

// Periods are the 24 layers in campaign order.
var Periods = []Period{
	{"August – September 1914", "War in Europe", []string{
		"Germany advances through Belgium",
		"towards Paris. In the east, Russia's",
		"armies are destroyed at Tannenberg;",
		"on the Marne the advance is stopped."}},
	{"October – November 1914", "The Race to the Sea", []string{
		"Both armies reach for the Channel",
		"ports. At Ypres the last open ground",
		"closes: a trench line now runs from",
		"the sea to the Swiss border."}},
	{"December 1914 – January 1915", "A Line of Trenches", []string{
		"Soldiers spend Christmas in flooded",
		"trenches; here and there the guns fall",
		"silent for a day. The generals plan",
		"how to break a front that cannot bend."}},
	{"February – March 1915", "War at Sea Declared", []string{
		"Germany declares the waters around",
		"Britain a war zone for its U-boats.",
		"British and French ships force the",
		"Dardanelles to open a road to Russia."}},
	{"April – May 1915", "Gas on the Western Front", []string{
		"At Ypres a green cloud rolls over the",
		"trenches: chlorine gas, a new weapon.",
		"Off Ireland a U-boat sinks the liner",
		"Lusitania with heavy loss of life."}},
	{"June – July 1915", "Italy Joins the Entente", []string{
		"Italy declares war on Austria-Hungary",
		"and attacks along the Isonzo. In the",
		"east, German and Austrian armies drive",
		"the Russians out of Galicia and Poland."}},
	{"August – September 1915", "The Great Retreat", []string{
		"Warsaw falls. Russia's armies retreat",
		"hundreds of miles, burning what they",
		"leave. In France, autumn offensives in",
		"Artois and Champagne gain little ground."}},
	{"October – November 1915", "A New Front at Salonika", []string{
		"Allied troops land at Salonika to aid",
		"Serbia — too late: with Bulgaria beside",
		"the Central Powers, Serbia is overrun",
		"and its army flees over the mountains."}},
	{"December 1915 – January 1916", "Gallipoli Abandoned", []string{
		"The last Allied troops slip away from",
		"Gallipoli by night. In Britain, single",
		"men are called to the colours as the",
		"volunteer armies reach their limit."}},
	{"February – March 1916", "Verdun", []string{
		"Germany attacks the fortress ring of",
		"Verdun to bleed France white. 'They",
		"shall not pass,' answers Paris. The",
		"battle becomes the longest of the war."}},
	{"April – May 1916", "Rebellion in Dublin", []string{
		"Irish republicans seize buildings in",
		"Dublin at Easter; the rising is crushed",
		"within a week. In Mesopotamia a British",
		"army surrenders at Kut al-Amara."}},
	{"June – July 1916", "Jutland and the Somme", []string{
		"The grand fleets clash off Jutland;",
		"both claim victory, neither risks a",
		"second meeting. On the Somme the great",
		"offensive opens; Russia strikes in Galicia."}},
	{"August – September 1916", "Romania Enters the War", []string{
		"Romania joins the Entente and marches",
		"into Transylvania. On the Somme, a new",
		"machine crawls over the trenches: the",
		"tank goes into battle for the first time."}},
	{"October – November 1916", "An Emperor Dies", []string{
		"Franz Joseph, emperor for 68 years,",
		"dies in Vienna; Karl succeeds him. The",
		"Somme battle drowns in November mud,",
		"having advanced a few miles at most."}},
	{"December 1916 – January 1917", "Peace Notes, No Peace", []string{
		"Berlin offers negotiations; the Entente",
		"rejects the note as a trap. General",
		"Nivelle promises France a breakthrough",
		"within forty-eight hours of spring."}},
	{"February – March 1917", "Revolution in Petrograd", []string{
		"Bread riots become revolution: the Tsar",
		"abdicates and a provisional government",
		"takes power. Germany answers at sea with",
		"unrestricted U-boat warfare."}},
	{"April – May 1917", "America Declares War", []string{
		"The United States declares war on",
		"Germany. In France the Nivelle offensive",
		"fails on the Aisne; whole divisions",
		"refuse to attack again."}},
	{"June – July 1917", "The Third Battle of Ypres", []string{
		"British guns fire four million shells",
		"before the infantry advances in Flanders.",
		"In Russia, Kerensky's summer offensive",
		"begins — the last his army can make."}},
	{"August – September 1917", "Mud", []string{
		"Rain turns the Ypres battlefield into a",
		"swamp in which men and guns sink. German",
		"troops take Riga with new storm tactics",
		"that will soon be tried in the west."}},
	{"October – November 1917", "Caporetto", []string{
		"German and Austrian divisions shatter",
		"the Italian front at Caporetto; Italy",
		"falls back to the Piave. In Petrograd the",
		"Bolsheviks seize power in Russia."}},
	{"December 1917 – January 1918", "Russia Leaves the War", []string{
		"An armistice stills the Eastern Front;",
		"at Brest-Litovsk the Bolsheviks haggle",
		"over peace. Every German division freed",
		"in the east is a division for France."}},
	{"February – March 1918", "The Kaiser's Battle", []string{
		"The Treaty of Brest-Litovsk ends the",
		"war in the east. On 21 March the greatest",
		"German offensive of the war bursts on",
		"the Somme: the race against America begins."}},
	{"April – May 1918", "The Lys Offensive", []string{
		"A second storm breaks in Flanders and",
		"is held before Hazebrouck. British",
		"raiders block Zeebrugge harbour. American",
		"troops now land in France by the thousands."}},
	{"June – July 1918", "The Turning of the Tide", []string{
		"Germany's last offensives reach the",
		"Marne and fail. On 18 July the Allies,",
		"with tanks and American divisions, strike",
		"back: the initiative changes hands for good."}},
}

// Months are the new monthly campaign's layers: one level per month,
// where the original campaign had one per two months (Periods). Only the
// first monthly level exists so far; the rest are added month by month.
var Months = []Period{
	{"September 1914", "The Miracle of the Marne", []string{
		"Paris is saved: on the Marne the German",
		"advance breaks. Kluck's flank, struck",
		"from Paris, gives way, and the invaders",
		"fall back to the Aisne to dig in."}},
}

// Month returns the monthly layer i, or false when there is none yet.
func Month(i int) (Period, bool) {
	if i < 0 || i >= len(Months) {
		return Period{}, false
	}
	return Months[i], true
}

// For returns the period of an original map: mapIndex counts the imported
// maps 00–71 in order, and each block of 24 (German, French, two-player)
// runs through Periods from the start. ok is false outside 0–71.
func For(mapIndex int) (Period, bool) {
	if mapIndex < 0 || mapIndex >= 3*len(Periods) {
		return Period{}, false
	}
	return Periods[mapIndex%len(Periods)], true
}

// Index is the period number (0–23) of an original map, or -1.
func Index(mapIndex int) int {
	if mapIndex < 0 || mapIndex >= 3*len(Periods) {
		return -1
	}
	return mapIndex % len(Periods)
}
