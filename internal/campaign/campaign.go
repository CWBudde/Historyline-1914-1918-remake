// Package campaign holds the campaign's calendar: one level per month from
// August 1914 to July 1918 (levels/monthly). The original's campaigns ran
// through 24 periods of two months, and its maps now take the first month
// of each period. The news texts are newly written one-paragraph summaries
// of the events, shown between missions as in the original: a month's own
// page where one is written, else its period's.
package campaign

import "fmt"

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
		"on the Marne the advance is stopped.",
	}},
	{"October – November 1914", "The Race to the Sea", []string{
		"Both armies reach for the Channel",
		"ports. At Ypres the last open ground",
		"closes: a trench line now runs from",
		"the sea to the Swiss border.",
	}},
	{"December 1914 – January 1915", "A Line of Trenches", []string{
		"Soldiers spend Christmas in flooded",
		"trenches; here and there the guns fall",
		"silent for a day. The generals plan",
		"how to break a front that cannot bend.",
	}},
	{"February – March 1915", "War at Sea Declared", []string{
		"Germany declares the waters around",
		"Britain a war zone for its U-boats.",
		"British and French ships force the",
		"Dardanelles to open a road to Russia.",
	}},
	{"April – May 1915", "Gas on the Western Front", []string{
		"At Ypres a green cloud rolls over the",
		"trenches: chlorine gas, a new weapon.",
		"Off Ireland a U-boat sinks the liner",
		"Lusitania with heavy loss of life.",
	}},
	{"June – July 1915", "Italy Joins the Entente", []string{
		"Italy declares war on Austria-Hungary",
		"and attacks along the Isonzo. In the",
		"east, German and Austrian armies drive",
		"the Russians out of Galicia and Poland.",
	}},
	{"August – September 1915", "The Great Retreat", []string{
		"Warsaw falls. Russia's armies retreat",
		"hundreds of miles, burning what they",
		"leave. In France, autumn offensives in",
		"Artois and Champagne gain little ground.",
	}},
	{"October – November 1915", "A New Front at Salonika", []string{
		"Allied troops land at Salonika to aid",
		"Serbia — too late: with Bulgaria beside",
		"the Central Powers, Serbia is overrun",
		"and its army flees over the mountains.",
	}},
	{"December 1915 – January 1916", "Gallipoli Abandoned", []string{
		"The last Allied troops slip away from",
		"Gallipoli by night. In Britain, single",
		"men are called to the colours as the",
		"volunteer armies reach their limit.",
	}},
	{"February – March 1916", "Verdun", []string{
		"Germany attacks the fortress ring of",
		"Verdun to bleed France white. 'They",
		"shall not pass,' answers Paris. The",
		"battle becomes the longest of the war.",
	}},
	{"April – May 1916", "Rebellion in Dublin", []string{
		"Irish republicans seize buildings in",
		"Dublin at Easter; the rising is crushed",
		"within a week. In Mesopotamia a British",
		"army surrenders at Kut al-Amara.",
	}},
	{"June – July 1916", "Jutland and the Somme", []string{
		"The grand fleets clash off Jutland;",
		"both claim victory, neither risks a",
		"second meeting. On the Somme the great",
		"offensive opens; Russia strikes in Galicia.",
	}},
	{"August – September 1916", "Romania Enters the War", []string{
		"Romania joins the Entente and marches",
		"into Transylvania. On the Somme, a new",
		"machine crawls over the trenches: the",
		"tank goes into battle for the first time.",
	}},
	{"October – November 1916", "An Emperor Dies", []string{
		"Franz Joseph, emperor for 68 years,",
		"dies in Vienna; Karl succeeds him. The",
		"Somme battle drowns in November mud,",
		"having advanced a few miles at most.",
	}},
	{"December 1916 – January 1917", "Peace Notes, No Peace", []string{
		"Berlin offers negotiations; the Entente",
		"rejects the note as a trap. General",
		"Nivelle promises France a breakthrough",
		"within forty-eight hours of spring.",
	}},
	{"February – March 1917", "Revolution in Petrograd", []string{
		"Bread riots become revolution: the Tsar",
		"abdicates and a provisional government",
		"takes power. Germany answers at sea with",
		"unrestricted U-boat warfare.",
	}},
	{"April – May 1917", "America Declares War", []string{
		"The United States declares war on",
		"Germany. In France the Nivelle offensive",
		"fails on the Aisne; whole divisions",
		"refuse to attack again.",
	}},
	{"June – July 1917", "The Third Battle of Ypres", []string{
		"British guns fire four million shells",
		"before the infantry advances in Flanders.",
		"In Russia, Kerensky's summer offensive",
		"begins — the last his army can make.",
	}},
	{"August – September 1917", "Mud", []string{
		"Rain turns the Ypres battlefield into a",
		"swamp in which men and guns sink. German",
		"troops take Riga with new storm tactics",
		"that will soon be tried in the west.",
	}},
	{"October – November 1917", "Caporetto", []string{
		"German and Austrian divisions shatter",
		"the Italian front at Caporetto; Italy",
		"falls back to the Piave. In Petrograd the",
		"Bolsheviks seize power in Russia.",
	}},
	{"December 1917 – January 1918", "Russia Leaves the War", []string{
		"An armistice stills the Eastern Front;",
		"at Brest-Litovsk the Bolsheviks haggle",
		"over peace. Every German division freed",
		"in the east is a division for France.",
	}},
	{"February – March 1918", "The Kaiser's Battle", []string{
		"The Treaty of Brest-Litovsk ends the",
		"war in the east. On 21 March the greatest",
		"German offensive of the war bursts on",
		"the Somme: the race against America begins.",
	}},
	{"April – May 1918", "The Lys Offensive", []string{
		"A second storm breaks in Flanders and",
		"is held before Hazebrouck. British",
		"raiders block Zeebrugge harbour. American",
		"troops now land in France by the thousands.",
	}},
	{"June – July 1918", "The Turning of the Tide", []string{
		"Germany's last offensives reach the",
		"Marne and fail. On 18 July the Allies,",
		"with tanks and American divisions, strike",
		"back: the initiative changes hands for good.",
	}},
}

// Months are the months with a news page of their own, by "YYYY-MM".
var Months = map[string]Period{
	"1914-09": {"September 1914", "The Miracle of the Marne", []string{
		"Paris is saved: on the Marne the German",
		"advance breaks. Kluck's flank, struck",
		"from Paris, gives way, and the invaders",
		"fall back to the Aisne to dig in.",
	}},
	"1914-11": {"November 1914", "The Wood at Nonne Bosschen", []string{
		"At Ypres the Prussian Guard storms out of",
		"Polygon Wood against a thin British line.",
		"The woods change hands, the gun line is",
		"threatened, and then the line holds.",
	}},
}

// News is the news page of a month ("1914-08"): its own page, else its
// period's; false outside August 1914 – July 1918.
func News(month string) (Period, bool) {
	if p, ok := Months[month]; ok {
		return p, true
	}
	i := PeriodOf(month)
	if i < 0 {
		return Period{}, false
	}
	return Periods[i], true
}

// PeriodOf is the index into Periods of a month ("1914-08"); -1 outside
// August 1914 – July 1918 or for a malformed month.
func PeriodOf(month string) int {
	var y, m int
	if n, err := fmt.Sscanf(month, "%4d-%2d", &y, &m); n != 2 || err != nil || m < 1 || m > 12 || fmt.Sprintf("%04d-%02d", y, m) != month {
		return -1
	}
	k := (y-1914)*12 + m - 8 // months after August 1914
	if k < 0 || k >= 2*len(Periods) {
		return -1
	}
	return k / 2
}
