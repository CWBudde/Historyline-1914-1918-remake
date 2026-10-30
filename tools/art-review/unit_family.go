package main

import (
	"fmt"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// The exporter supports the explicit runtime families below. Owner palette
// selection and all image geometry/filtering remain shared between them.
type unitFamily struct {
	name            string
	classIDs        []string
	paletteSelector string
	profile         string
	acceptance      string
	title           string
	subject         string
}

func lookupUnitFamily(name string) (unitFamily, error) {
	if title, ok := catalogueSubjects[name]; ok {
		return unitFamily{
			name: name, title: title, subject: title, classIDs: unitsprite.ClassIDs(name),
			paletteSelector: "Owner palette is independent of class ID/country, including captured units.",
			profile:         "unit-" + name + "-v1: six authored square panels row-major N,NW,SW,S,SE,NE. " + unitAlgorithmProfile,
			acceptance:      "Selected catalogue artwork with 2x/4x exports and gameplay class bindings; final visual acceptance remains open.",
		}, nil
	}
	switch name {
	case "infantry":
		return unitFamily{
			name: name, classIDs: []string{"CP_infantry", "EC_infantry"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured infantry; no eliteinfantry alias.",
			profile:         unitProfile, acceptance: unitAcceptance, title: "Infantry", subject: "infantry",
		}, nil
	case "lightArtillery":
		return unitFamily{
			name: name, classIDs: []string{"CP_lArtillery", "EC_lArtillery"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured light artillery; no medium/heavy artillery aliases.",
			profile:         "unit-light-artillery-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_lArtillery and EC_lArtillery independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("infantry", "light artillery", "physical boot pivot", "physical gun/crew pivot", "residual cloth hue", "residual paint/cloth hue").Replace(unitAcceptance),
			title:           "Light artillery", subject: "light artillery",
		}, nil
	case "eliteInfantry":
		return unitFamily{
			name: name, classIDs: []string{"CP_eliteInf", "EC_eliteInf"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured elite infantry; no ordinary infantry alias.",
			profile:         "unit-elite-infantry-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_eliteInf and EC_eliteInf independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.ReplaceAll(unitAcceptance, "PROVISIONAL infantry", "PROVISIONAL elite infantry machine-gun team"),
			title:           "Elite infantry", subject: "elite infantry machine-gun team",
		}, nil
	case "cavalry":
		return unitFamily{
			name: name, classIDs: []string{"CP_cavalry", "EC_cavalry"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured cavalry; no other family aliases.",
			profile:         "unit-cavalry-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_cavalry and EC_cavalry independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL cavalry", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Cavalry", subject: "cavalry",
		}, nil
	case "mediumArtillery":
		return unitFamily{
			name: name, classIDs: []string{"CP_mArtillery", "EC_mArtillery"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured medium artillery; no other family aliases.",
			profile:         "unit-medium-artillery-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_mArtillery and EC_mArtillery independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL medium artillery", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Medium artillery", subject: "medium artillery",
		}, nil
	case "heavyArtillery":
		return unitFamily{
			name: name, classIDs: []string{"CP_hArtillery", "EC_hArtillery"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured heavy artillery; no other family aliases.",
			profile:         "unit-heavy-artillery-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_hArtillery and EC_hArtillery independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL heavy artillery", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Heavy artillery", subject: "heavy artillery",
		}, nil
	case "supplyCar":
		return unitFamily{
			name: name, classIDs: []string{"CP_supplyCar", "EC_supplyCar"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured supply car; no other family aliases.",
			profile:         "unit-supply-car-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_supplyCar and EC_supplyCar independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL supply car", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Supply car", subject: "supply car",
		}, nil
	case "bunker":
		return unitFamily{
			name: name, classIDs: []string{"CP_bunker", "EC_bunker"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured bunker; no other family aliases.",
			profile:         "unit-bunker-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_bunker and EC_bunker independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL bunker", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Bunker", subject: "bunker",
		}, nil
	case "armoredCar":
		return unitFamily{
			name: name, classIDs: unitsprite.ClassIDs("armored-car"),
			paletteSelector: "Owner palette is independent of class ID/country, including captured armoured car; no other family aliases.",
			profile:         "unit-armored-car-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_armCar and EC_armCar independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL armoured car", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Armoured car", subject: "armoured car",
		}, nil
	case "charron":
		return unitFamily{
			name: name, classIDs: []string{"EC_charron"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured charron; no other family aliases.",
			profile:         "unit-charron-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup EC_charron independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL charron", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Charron", subject: "charron",
		}, nil
	case "patrolBoat":
		return unitFamily{
			name: name, classIDs: []string{"CP_patrolBoat", "EC_patrolBoat"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured patrol boat; no other family aliases.",
			profile:         "unit-patrol-boat-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_patrolBoat and EC_patrolBoat independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL patrol boat", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Patrol boat", subject: "patrol boat",
		}, nil
	case "transportTrain":
		return unitFamily{
			name: name, classIDs: unitsprite.ClassIDs("transport-train"),
			paletteSelector: "Owner palette is independent of class ID/country, including captured transport train; no other family aliases.",
			profile:         "unit-transport-train-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_transTrain and EC_transTrain independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL transport train", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Transport train", subject: "transport train",
		}, nil
	case "armoredTrain":
		return unitFamily{
			name: name, classIDs: []string{"CP_armTrain", "EC_armTrain"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured armoured train; no other family aliases.",
			profile:         "unit-armored-train-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_armTrain and EC_armTrain independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL armoured train", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Armoured train", subject: "armoured train",
		}, nil
	case "trainArtillery":
		return unitFamily{
			name: name, classIDs: []string{"CP_trainArt", "EC_trainArt"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured artillery train; no other family aliases.",
			profile:         "unit-train-artillery-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_trainArt and EC_trainArt independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL artillery train", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Artillery train", subject: "artillery train",
		}, nil
	case "sappers":
		return unitFamily{
			name: name, classIDs: []string{"CP_sappers", "EC_sappers"},
			paletteSelector: "Owner palette is independent of class ID/country, including captured sappers; no other family aliases.",
			profile:         "unit-sappers-v1: six authored square panels row-major N,NW,SW,S,SE,NE; class family lookup CP_sappers and EC_sappers independently selects owner cp/ec/neutral. " + unitAlgorithmProfile,
			acceptance:      strings.NewReplacer("PROVISIONAL infantry", "PROVISIONAL sappers", "physical boot pivot", "physical subject pivot").Replace(unitAcceptance),
			title:           "Sappers", subject: "sappers",
		}, nil
	default:
		return unitFamily{}, fmt.Errorf("unit family %q must name a canonical runtime graphics family", name)
	}
}

var catalogueSubjects = map[string]string{
	"anti-tank": "Anti-tank gun", "static-aa": "Static anti-aircraft gun", "mobile-aa": "Mobile anti-aircraft gun",
	"a7v": "A7V", "st-chamond": "St-Chamond", "renault-ft17": "Renault FT17", "mark-i": "Mark I", "mark-iv": "Mark IV",
	"torpedo-boat": "Torpedo boat", "destroyer": "Destroyer", "battleship": "Battleship", "transport-ship": "Transport ship", "submarine": "Submarine",
	"static-balloon": "Observation balloon", "fokker-e-i": "Fokker E.I", "fokker-e-iii": "Fokker E.III",
	"albatros-d-iii": "Albatros D.III", "fokker-dr-i": "Fokker Dr.I", "fokker-d-vii": "Fokker D.VII", "junkers-j4": "Junkers J4",
	"gotha-g-vb": "Gotha G.Vb", "zeppelin-staaken-r-vi": "Zeppelin-Staaken R.VI", "dh2": "DH2",
	"morane-saulnier-n": "Morane-Saulnier N", "nieuport-xvii": "Nieuport XVII", "spad-vii": "SPAD VII",
	"se5a": "SE5a", "spad-xiii": "SPAD XIII", "sopwith-camel": "Sopwith Camel", "voisin-iii": "Voisin III",
	"handley-page-0400": "Handley Page 0/400", "dh4": "DH4", "construction-unit": "Construction unit",
	"infantry-1": "Infantry, one man", "infantry-2": "Infantry, two men", "infantry-3": "Infantry, three men", "infantry-4": "Infantry, four men", "infantry-5": "Infantry, five men", "elite-infantry-1": "Elite infantry, strength 1-3",
}
