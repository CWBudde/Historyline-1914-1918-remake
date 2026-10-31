package hlcli

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finextract"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlart"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlasset"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlbackdrop"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlcampaign"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlgallery"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlkit"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlllplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlmaterial"
	artreview "github.com/CWBudde/Historyline-1914-1918-remake/tools/art-review"
	candidatereview "github.com/CWBudde/Historyline-1914-1918-remake/tools/candidate-review"
	interfaceart "github.com/CWBudde/Historyline-1914-1918-remake/tools/interface-art"
	referencedesign "github.com/CWBudde/Historyline-1914-1918-remake/tools/reference-design"
	terrainstamps "github.com/CWBudde/Historyline-1914-1918-remake/tools/terrain-stamps"
	unitpanel "github.com/CWBudde/Historyline-1914-1918-remake/tools/unit-panel"
	unitsquad "github.com/CWBudde/Historyline-1914-1918-remake/tools/unit-squad"
	unittopview "github.com/CWBudde/Historyline-1914-1918-remake/tools/unit-topview"
)

func commands() command {
	return group("hl", "Historyline developer tools",
		group("art", "Art catalogue, thumbnails and map renders", actions("index check thumbs pulse map", hlart.Run,
			"Write the art catalogue", "Check prompts, masters, runtime files and catalogue", "Write catalogue thumbnails", "Render the PULSE reference map", "Render a level and detail crop")...),
		group("assets", "Build and validate runtime art packs", actions("build build-units build-terrain build-screens packs lite terrain-gaps validate", hlasset.Run,
			"Build an interface pack", "Build unit packs", "Build terrain packs", "Build screen packs", "Update changed runtime packs", "Compress packs for the web", "Report missing terrain clusters", "Validate a runtime pack")...),
		group("export", "Export art from pinned sources",
			leaf("materials", "Export reusable periodic world materials", tool(hlmaterial.Run)),
			leaf("kits", "Prepare world kits; -install writes runtime cells", tool(hlkit.Run)),
			leaf("terrain-stamps", "Export pinned terrain cut-outs", tool(terrainstamps.Run)),
			leaf("unit-topview", "Export and install top-view unit families", tool(unittopview.Run)),
			leaf("unit-squad", "Compose a squad atlas", tool(unitsquad.Run)),
			leaf("unit-panel", "Export the unit inspector panel", tool(unitpanel.Run)),
			leaf("interface", "Export interface, weapon and effect candidates", tool(interfaceart.Run)),
			generatorCommands()),
		group("review", "Compare and review art in its game context",
			leaf("art", "Export art comparisons and manifest-driven candidate bundles", tool(artreview.Run)),
			leaf("reference", "Check or render the PULSE reference design", tool(referencedesign.Run)),
			leaf("candidates", "Render raw candidates beside installed art", tool(candidatereview.Run)),
			leaf("gallery", "Build the comparison desk, composer and PULSE study", tool(hlgallery.Run))),
		group("levels", "Extract and prepare campaign levels",
			leaf("extract", "Extract local original maps as level JSON", tool(finextract.Run)),
			leaf("campaign", "Write campaign folders from original maps (overwrites setups)", tool(hlcampaign.Run)),
			group("backdrop", "Encode or check painted backdrop JPEGs", actions("jpeg check", hlbackdrop.Run,
				"Encode the named levels' JPEGs from lossless sources", "Check all backdrop sources against their JPEGs")...)),
		group("llmplay", "Play, record and review games with an external agent", actions("new prompt step run replay render review gif", hlllplay.Run,
			"Start a recorded game", "Print the current phase prompt", "Apply one atomic batch of orders", "Play a recorded game through a driver", "Verify every replay state hash", "Render replay frames and video", "Write the game review and counted lessons", "Encode recorded frames as a GIF")...),
		group("preview", "Native graphics previews", leaf("effects", "Preview unit movement and effects (requires a graphics display)", previewEffects)),
		leaf("serve", "Serve the local web demo", webServer),
		leaf("web", "Build the web demo into build/web (-port serves it)", webBuild),
		leaf("artbook", "Render the art book PDF into build/artbook (needs typst)", artbook),
		leaf("manual", "Build the reconstructed manual PDF (needs typst; -images, -watch)", manual),
	)
}
