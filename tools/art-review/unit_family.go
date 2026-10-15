package main

import (
	"fmt"
	"strings"
)

// The bounded pilot supports exactly these two runtime families. Owner palette
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
	default:
		return unitFamily{}, fmt.Errorf("unit family %q must be infantry or lightArtillery", name)
	}
}
