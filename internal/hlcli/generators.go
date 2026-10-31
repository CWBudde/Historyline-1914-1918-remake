package hlcli

import (
	"context"
	"flag"
	"fmt"
	"io"
)

type generator struct {
	name, tag, test, pkg, summary string
}

// These authoring hooks remain in their tagged test packages. The fixed test
// names prevent a generator invocation from running unrelated asset writers.
var generators = []generator{
	{"airfield-tiles", "campaigngfx", "TestGenerateAirfieldTiles", "art-review", "Install airfield terrain"},
	{"bridge-decks", "bridgegfx", "TestGenerateBridgeDecks", "terrain-stamps", "Install bridge decks"},
	{"campaign-contexts", "campaigngfx", "TestGenerateCampaignContexts", "art-review", "Install campaign terrain contexts"},
	{"campaign-shores", "campaigngfx", "TestGenerateCampaignShores", "art-review", "Install campaign shoreline terrain"},
	{"campaign-trenches", "campaigngfx", "TestGenerateCampaignTrenches", "art-review", "Install campaign trench terrain"},
	{"compound-shed-masters", "campaigngfx", "TestGenerateCompoundShedMasters", "art-review", "Write compound shed masters"},
	{"compound-tiles", "campaigngfx", "TestGenerateCompoundTiles", "art-review", "Install compound terrain"},
	{"grass-detail-candidate", "grassgfx", "TestGenerateGrassDetailCandidate", "art-review", "Write a grass detail candidate"},
	{"grass-scale-proof", "grassgfx", "TestGenerateGrassScaleProof", "art-review", "Write the grass scale proof"},
	{"grass-variations", "grassgfx", "TestGenerateGrassVariations", "art-review", "Install selected grass variations"},
	{"harbour-tiles", "campaigngfx", "TestGenerateHarbourTiles", "art-review", "Install harbour terrain"},
	{"lattice-ground-contexts", "campaigngfx", "TestRegenerateLatticeGroundContexts", "art-review", "Regenerate and install lattice-periodic ground contexts"},
	{"long-ships", "campaigngfx", "TestGenerateLongShips", "art-review", "Install long ship sprites"},
	{"quay-tiles", "campaigngfx", "TestGenerateQuayTiles", "art-review", "Install quay terrain"},
	{"retained-context-bundle", "retainedgfx", "TestGenerateRetainedContextBundle", "art-review", "Supplement an ignored bundle from retained context profiles"},
	{"river-bank-gravel-crop", "riverbankgfx", "TestGenerateRiverBankGravelCrop", "art-review", "Write river bank gravel crops"},
	{"ruin-tiles", "campaigngfx", "TestGenerateRuinTiles", "art-review", "Install ruin terrain"},
	{"sea-kit", "campaigngfx", "TestGenerateSeaKit", "art-review", "Install sea kit items"},
	{"sea-tiles", "campaigngfx", "TestGenerateSeaTiles", "art-review", "Install open sea terrain"},
	{"surf", "campaigngfx", "TestGenerateSurf", "art-review", "Install surf terrain"},
	{"water-rocks", "campaigngfx", "TestGenerateWaterRocks", "art-review", "Install rocks in water"},
	{"winter-paths-and-rails", "campaigngfx", "TestGenerateWinterPathsAndRails", "art-review", "Install winter paths and rails"},
	{"winter-roads", "campaigngfx", "TestGenerateWinterRoads", "art-review", "Install winter roads"},
	{"winter-streams", "campaigngfx", "TestGenerateWinterStreams", "art-review", "Install winter streams"},
	{"winter-trenches", "campaigngfx", "TestGenerateWinterTrenches", "art-review", "Install winter trenches"},
	{"winter-water", "campaigngfx", "TestGenerateWinterWater", "art-review", "Install winter water terrain"},
}

func generatorCommands() command {
	var children []command
	for _, g := range generators {
		children = append(children, leaf(g.name, g.summary, g.run))
	}
	return group("generate", "Explicit repository authoring hooks (some replace art and assets; requires Go)", children...)
}

func (g generator) run(ctx context.Context, args []string, in io.Reader, out, errs io.Writer) error {
	f := flag.NewFlagSet("hl export generate "+g.name, flag.ContinueOnError)
	f.SetOutput(errs)
	var family, manifest, profiles, bundle string
	if g.tag == "retainedgfx" {
		f.StringVar(&family, "context-family", "", "ground, forest, or hills (required)")
		f.StringVar(&manifest, "context-manifest", "", "current pinned input manifest (required)")
		f.StringVar(&profiles, "context-profiles", "", "retained context profiles file (required)")
		f.StringVar(&bundle, "context-bundle", "", "fresh ignored CLI bundle to supplement (required)")
	}
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	if g.tag == "retainedgfx" && (family == "" || manifest == "" || profiles == "" || bundle == "") {
		return fmt.Errorf("all four context flags are required")
	}
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	testArgs := []string{"test", "-timeout", "40m", "-tags", g.tag, "./tools/" + g.pkg, "-run", "^" + g.test + "$", "-count=1", "-v"}
	if g.tag == "retainedgfx" {
		testArgs = append(testArgs, "-args", "-context-family", family, "-context-manifest", manifest, "-context-profiles", profiles, "-context-bundle", bundle)
	}
	return runProcess(ctx, root, "go", testArgs, in, out, errs)
}
