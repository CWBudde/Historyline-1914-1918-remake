// Command historyline is the Go port of the Historyline 1914-1918 remake.
//
// For now it shows a shipped level as flat-colour placeholder hexes with scrolling,
// zoom and hex picking (PLAN.md 2.4). Real graphics follow with the runtime asset
// packs (2.3).
package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"image/color"
	"log"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scores"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	startWidth  = 1800
	startHeight = 1000
	scrollSpeed = 12 // default map scroll speed, screen pixels per frame
)

// Pixel outline of a hex around its centre (flat top, 96×80).
var hexOutline = [6][2]float32{{-48, 0}, {-25, -40}, {25, -40}, {48, 0}, {25, 40}, {-25, 40}}

type game struct {
	lvl         *level.Level
	reveal      bool // dev: every unit where its own side sees it
	hexNumbers  bool // dev: print line,col on every hex
	grid        hexgrid.Grid
	cam         *camera.Camera
	hover       hexgrid.Coord
	hoverOnMap  bool
	view        view // device and logical screen sizes, UI scale
	w, h        int  // logical HUD size (view.LogW, view.LogH)
	hud         hudLayout
	hudImg      *ebiten.Image // logical HUD canvas, magnified onto the screen
	terrainFill map[string]color.RGBA
	dos         *dosControl // the session's controller; nil only in tests of the bare map
	art         *artPacks   // unit sprite packs (-units); nil draws placeholder dots
	screens     *screenArt  // the screen catalog (-screens): portraits, badges, effects; nil shows plain ones
	finale      struct {    // the finale's art and the catalog it was cut from (taking.go)
		from *screenArt
		art  *dospanel.FinaleArt
	}
	surrender struct { // the surrender scene's art and its catalog (taking.go)
		from *screenArt
		art  *dospanel.SurrenderArt
	}
	scores       *scores.Store          // the best scores; nil keeps none
	artErr       string                 // last sprite failure, shown in the status line
	terrain      *terrainArt            // composed terrain tiles (-terrain); nil draws coloured hexes
	terrainErr   string                 // last terrain tile failure, shown in the status line
	hdUnits      <-chan *artPacks       // full-quality unit art while it loads (-units-hd, without the menus)
	hdTerrain    <-chan *terrainArt     // full-quality terrain while it loads (-terrain-hd)
	records      []int                  // the terrain records as the sappers' work played left them, nil before (trench.go)
	backdrop     *backdrop              // the painted ground drawn in place of the terrain (-backdrop); nil draws the terrain
	backdropView *ebiten.Image          // the view of the backdrop the seen hexes are greyed from
	reworkedHex  map[hexgrid.Coord]bool // the hexes the sappers changed (trench.go)
	scroll       float64                // map scroll speed, screen pixels per frame

	split    *split // the split screen; nil only in tests of the bare map
	drawing  bool   // a dos panel is being drawn: drawSide's map
	drawSide int
	tileTint map[string]color.RGBA // average colour of a terrain tile, by asset ID
	orig     *origArt              // the original tiles and units (F8, origtiles.go); nil until first asked for
}

// newMapGame is the map view of lvl without a rules engine.
func newMapGame(lvl *level.Level, reveal bool) *game {
	g := lvl.Grid()
	ww, wh := g.Bounds()
	world := camera.Size{W: ww, H: wh}
	gm := &game{
		lvl: lvl, reveal: reveal, grid: g,
		cam:         camera.New(world, camera.Size{W: startWidth, H: startHeight}),
		terrainFill: map[string]color.RGBA{},
		scroll:      scrollSpeed,
	}
	gm.applyLayout(startWidth, startHeight, 1)
	return gm
}

// applyLayout adapts the camera, HUD and split screen to a window of outW × outH
// device-independent pixels at display scale dpi. The camera works in device
// pixels with the UI scale as its display factor, so the map keeps its apparent
// size and the view centre stays put.
func (g *game) applyLayout(outW, outH int, dpi float64) {
	v := newView(outW, outH, dpi)
	if v == g.view {
		return
	}
	g.view = v
	g.w, g.h = v.LogW, v.LogH
	g.hud = layoutHUD(v.LogW, v.LogH)
	g.cam.Resize(camera.Size{W: float64(v.PhysW), H: float64(v.PhysH)})
	g.cam.SetScale(float64(v.UI))
	g.layoutSplit()
}

// fogOf is what the shown panel knows about c: without fog (classic, or
// -reveal) every hex is visible; under fog a hex out of its eyes' sight is
// remembered terrain.
func (g *game) fogOf(c hexgrid.Coord) state.Fog {
	if g.reveal {
		return state.Visible
	}
	sight := g.dos.sightOf(g.dos.eyes(g.shownSide()))
	if sight == nil || sight[c] {
		return state.Visible
	}
	return state.Seen
}

// fogShade applies what the player knows to a terrain colour. Unseen hexes show
// nothing of the terrain and remembered ones are dimmed and grey.
func fogShade(f state.Fog, c color.RGBA) color.RGBA {
	switch f {
	case state.Visible:
		return c
	case state.Seen:
		grey := uint8((int(c.R) + int(c.G) + int(c.B)) / 3 / 2)
		return color.RGBA{R: grey, G: grey, B: grey + 8, A: 255}
	}
	return color.RGBA{R: 8, G: 8, B: 10, A: 255}
}

// walking reports whether the unit standing on at is still being animated
// there. A building or carrier the unit walks into stays drawn.
func (g *game) walking(at hexgrid.Coord) bool { return g.dosWalking(at) }

// unitShown reports whether a unit is drawn on c: one stands there on the
// shown side's map and is not walking.
func (g *game) unitShown(c hexgrid.Coord) bool {
	_, ok := g.dos.unitsOf(g.shownSide(), g.reveal)[c]
	return ok && !g.walking(c)
}

// ownerShown is the owner of the building on c as the shown side's map has
// it, None when there is no building.
func (g *game) ownerShown(c hexgrid.Coord) state.Player {
	return g.dos.buildingsOf(g.shownSide())[c].Owner
}

// ownerColour marks the two sides' buildings and unit plates in their
// minimap colours, as the original does: the German side blue, the Entente
// yellow.
func ownerColour(p state.Player) color.RGBA {
	return sideShade(p).Lit
}

func (g *game) fill(terrain string) color.RGBA {
	if c, ok := g.terrainFill[terrain]; ok {
		return c
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(terrain))
	v := hash.Sum32()
	c := color.RGBA{R: 70 + uint8(v>>16)%120, G: 90 + uint8(v>>8)%120, B: 60 + uint8(v)%100, A: 255}
	g.terrainFill[terrain] = c
	return c
}

func (g *game) Update() error {
	if u, ok := arrived(&g.hdUnits); ok {
		g.art.release()
		g.art = u
	}
	if t, ok := arrived(&g.hdTerrain); ok {
		g.terrain.release()
		g.terrain = t
		if g.records != nil {
			g.reworkTerrain()
		}
	}
	anyInput := len(inpututil.AppendJustPressedKeys(nil)) > 0 || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if g.dos.scene() && anyInput {
		g.dos.SkipScene() // as the battle panel, any key or click skips a scene
	}
	g.dos.Tick()
	if g.dos.rework {
		g.dos.rework = false
		g.reworkTerrain()
	}
	g.updateSplit()
	// Space is the fire key of the cursor (key set A).
	pressed := func(k ebiten.Key) bool { return k != ebiten.KeySpace && inpututil.IsKeyJustPressed(k) }
	for _, a := range actionsFor(pressed) {
		if !g.applyAction(a) {
			return ebiten.Termination
		}
	}
	// The wheel steps the zoom levels and the cursor takes the mouse and the
	// keys.
	_, wheel := ebiten.Wheel()
	mx, my := ebiten.CursorPosition()
	g.wheelSplit(wheel, mx, my)
	g.updateCursor()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(mapGround)
	g.drawSplit(screen)
	g.drawWells(screen)
	g.drawFinale(screen)
	g.drawSurrender(screen)

	// Everything below is HUD, drawn in logical pixels and magnified by the UI scale.
	if g.hudImg == nil || g.hudImg.Bounds().Dx() != g.w || g.hudImg.Bounds().Dy() != g.h {
		g.hudImg = ebiten.NewImage(g.w, g.h)
	}
	hud := g.hudImg
	hud.Clear()
	if g.hexNumbers {
		g.drawSplitHexNumbers(hud)
	}
	g.drawOrderNumbers(hud)
	g.drawBanner(hud)
	g.drawInfo(hud)
	if dev := g.devIndicators(); dev != "" {
		ebitenutil.DebugPrintAt(hud, dev, g.hud.DevRight-textWidth(dev), 0)
	}
	g.drawStatus(hud)
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(g.view.UI), float64(g.view.UI))
	screen.DrawImage(hud, &op)
}

// drawMapLayer draws the map in the world rectangle (vx, vy, vw, vh) through
// g.cam on dst, in device pixels. The selection, orders, walker and hover
// outline belong to the human side's map only (human). The edge of an original
// map is not drawn (playGrid), so its standing items cannot reach into view.
func (g *game) drawMapLayer(dst *ebiten.Image, vx, vy, vw, vh float64, human bool) {
	// A hex just outside the view counts while its standing items may reach in.
	const mx, my = hexgrid.HexWidth/2 + mapcompose.OverhangX, hexgrid.HexHeight/2 + mapcompose.OverhangY
	play := g.playGrid()
	var inView []hexgrid.Coord
	for l, line := range g.lvl.Lines {
		for c := range line {
			coord := hexgrid.Coord{Line: l, Col: c}
			cx, cy := hexgrid.Center(coord)
			if !play.Contains(coord) || cx+mx < vx || cx-mx > vx+vw || cy+my < vy || cy-my > vy+vh {
				continue
			}
			inView = append(inView, coord)
		}
	}
	// Terrain, then owner marks, then units and their markers, each in its own pass so that a later
	// hex never covers what an earlier one drew beyond its outline.
	switch {
	case g.drawOrigTiles(dst, inView):
	case g.backdrop != nil:
		g.drawBackdrop(dst, vx, vy, vw, vh, inView)
	case g.terrain != nil:
		g.drawTerrain(dst, inView)
	default:
		for _, coord := range inView {
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(coord))
			hx := g.lvl.Lines[coord.Line][coord.Col]
			g.drawHex(dst, float32(sx), float32(sy), fogShade(g.fogOf(coord), g.fill(hx.Terrain)), false)
		}
	}
	for _, coord := range inView {
		if owner := g.ownerShown(coord); owner != state.None {
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(coord))
			e := float32(g.cam.Effective())
			vector.FillRect(dst, float32(sx)-20*e, float32(sy)-34*e, 40*e, 6*e, fogShade(g.fogOf(coord), ownerColour(owner)), false)
		}
	}
	for _, coord := range inView {
		if g.unitShown(coord) {
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(coord))
			g.drawDosUnit(dst, coord, sx, sy)
		}
	}
	g.drawUnitMarks(dst, inView)
	if w := g.dos.walkOf(g.shownSide()); !w.Done() && g.fogOf(w.Hex()) == state.Visible {
		wx, wy := w.Pos()
		sx, sy := g.cam.WorldToScreen(wx, wy)
		face, _ := unitsprite.Toward(w.Heading())
		if !g.drawWalker(dst, w, w.Hex(), sx, sy, face) {
			vector.FillCircle(dst, float32(sx), float32(sy), float32(14*g.cam.Effective()), color.RGBA{230, 230, 230, 255}, true)
		}
	}
	if human {
		g.darken(dst, inView) // the cursor is drawn over the frame (drawCursor)
		g.drawOrders(dst)
	}
	// Hit marks show on both sides' maps: both see the battle.
	g.drawMarks(dst)
}

// drawStatus prints the art failures at the top left of the HUD: the status
// wells carry the state and the message.
func (g *game) drawStatus(dst *ebiten.Image) { ebitenutil.DebugPrint(dst, g.splitErrors()) }

// devIndicators names the active development flags so a screenshot can never be
// mistaken for normal play.
func (g *game) devIndicators() string {
	var on []string
	if g.reveal {
		on = append(on, "REVEAL")
	}
	if g.hexNumbers {
		on = append(on, "HEXNUMBERS")
	}
	if g.art != nil && g.art.candidate || g.terrain != nil && g.terrain.candidate {
		on = append(on, "CANDIDATE-ART")
	}
	if g.terrain != nil {
		if n := len(g.terrain.scene.Gaps); n > 0 {
			on = append(on, fmt.Sprintf("CONTEXT-GAP %d", n))
		}
		if n := len(g.terrain.scene.Issues); n > 0 {
			on = append(on, fmt.Sprintf("COMPOUND-ISSUE %d", n))
		}
	}
	if len(on) == 0 {
		return ""
	}
	return "DEV: " + strings.Join(on, " ")
}

func (g *game) drawHex(dst *ebiten.Image, cx, cy float32, fill color.RGBA, outline bool) {
	z, ui := float32(g.cam.Effective()), float32(g.view.UI)
	var p vector.Path
	for i, v := range hexOutline {
		x, y := cx+v[0]*z, cy+v[1]*z
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.Close()
	if outline {
		vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: 3 * ui}, pathOptions(color.RGBA{255, 240, 80, 255}))
		return
	}
	vector.FillPath(dst, &p, &vector.FillOptions{}, pathOptions(fill))
	vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: ui}, pathOptions(color.RGBA{0, 0, 0, 90}))
}

func pathOptions(c color.RGBA) *vector.DrawPathOptions {
	var o vector.DrawPathOptions
	o.AntiAlias = true
	o.ColorScale.ScaleWithColor(c)
	return &o
}

// Layout returns the screen in device pixels, so the map is drawn at full
// resolution; the HUD is magnified separately (see view). The split screen
// keeps the window large enough for its smallest panels.
func (g *game) Layout(outsideW, outsideH int) (int, int) {
	dpi := ebiten.Monitor().DeviceScaleFactor()
	g.applyLayout(outsideW, outsideH, dpi)
	if g.split != nil {
		limitWindow(splitMinWindow(dpi))
	} else {
		limitWindow(0, 0)
	}
	return g.view.PhysW, g.view.PhysH
}

// loadLevel loads name: a setup of a scenario file below dir
// ("monthly/1914-08:ec", see loadScenario), or a level file of the
// legacy format by its path ("art/reference/pulse/level.json").
func loadLevel(dir, name string) (*level.Level, error) {
	for prefix, emb := range scenarioDirs {
		if strings.HasPrefix(name, prefix) {
			return loadScenario(dir, emb, name)
		}
	}
	if !strings.HasSuffix(name, ".json") {
		return nil, fmt.Errorf("level %q: not a scenario setup (monthly/…, skirmish/…), original/NN or a .json file", name)
	}
	return level.Load(name)
}

// matchConfig is everything a battle is started from: by the menu, or directly
// by -level or -player.
type matchConfig struct {
	lvl        *level.Level
	board      *dosbind.Board // an original map's exact binding (loadMap); nil binds lvl
	set        *data.Set
	player     state.Player // the human side, whose view is shown
	rules      string       // rulesClassic (default) or rulesModern
	opponent   string       // ai or pass
	limit      int          // orders per phase, 0 for no limit
	reveal     bool
	hexNumbers bool
	reduced    bool
	scroll     float64        // map scroll speed, screen pixels per frame
	units      *artPacks      // shared by every battle; nil draws placeholder dots
	screens    *screenArt     // the screen catalog for the portraits; nil shows blank photos
	scores     *scores.Store  // the best scores, opened once; nil keeps none
	saves      *saves.Store   // the single saved battle; nil keeps none
	saved      []byte         // a saved battle to resume; nil starts fresh
	terrain    string         // terrain catalog directory, composed per level; "" draws coloured hexes
	terrainHD  string         // full-quality terrain catalog loaded in the background of a battle; "" for none
	candidates bool           // allow candidate terrain packs
	backdrop   *backdropImage // the painted ground of every level (-backdrop); nil takes the level's own (levelBackdrop)
	noBackdrop bool           // draw the terrain even where the level has a backdrop (-backdrop off)
	levels     string         // the directory of the level files (-levels)
}

// The rule sets of -rules. Both play the original order cycle on a dosplay
// session; modern extends it.
const (
	rulesClassic = "classic" // the original rules profile (default)
	rulesModern  = "modern"  // the original rules with extensions
)

// rulesOptions sets the session options of a rule set on o.
func rulesOptions(rules string, o dosplay.Options) dosplay.Options {
	switch rules {
	case rulesClassic: // the original rules profile as it is
	case rulesModern:
		o.Fog = true
	}
	return o
}

// newMatch starts the battle of c, counting the art it fetches in m.
func newMatch(c matchConfig, m *meter) (*game, error) {
	if c.rules != rulesClassic && c.rules != rulesModern {
		return nil, fmt.Errorf("unknown -rules %q (want %s or %s)", c.rules, rulesClassic, rulesModern)
	}
	o := rulesOptions(c.rules, dosplay.Options{Limit: c.limit})
	var planner dosPlanner = passPlanner{}
	if c.opponent == "ai" {
		o.Computer[1-dosbind.Side(c.player)] = true
		planner = computerPlanner()
	}
	var s *dosplay.Session
	var err error
	switch {
	case c.saved != nil && c.board != nil:
		s, err = dosplay.LoadBoard(c.board, c.lvl, c.set, c.saved)
	case c.saved != nil:
		s, err = dosplay.Load(c.lvl, c.set, c.saved)
	case c.board != nil:
		s, err = dosplay.NewBoard(c.board, c.lvl, c.set, o)
	default:
		s, err = dosplay.New(c.lvl, c.set, o)
	}
	if err != nil {
		return nil, err
	}
	if c.saved != nil {
		// The save carries Computer flags and the limit; the planner
		// follows the other side's flag.
		planner = passPlanner{}
		if s.Game.Players[1-s.SideOf(c.player)].Computer {
			planner = computerPlanner()
		}
	}
	gm := newDosGame(c.lvl, s, s.SideOf(c.player), planner, c.reveal)
	gm.hexNumbers = c.hexNumbers
	gm.dos.SetReducedMotion(c.reduced)
	if c.scroll > 0 {
		gm.scroll = c.scroll
	}
	gm.art, gm.screens, gm.scores = c.units, c.screens, c.scores
	bd := c.backdrop
	if bd == nil && !c.noBackdrop && c.lvl.Extra != nil && c.lvl.Extra.Backdrop != "" {
		var err error
		if bd, err = levelBackdrop(c.levels, c.lvl.Extra.Backdrop); err != nil {
			return nil, err
		}
	}
	if bd != nil {
		var err error
		if gm.backdrop, err = bd.on(c.lvl.Grid()); err != nil {
			return nil, err
		}
	}
	if c.terrain != "" {
		var err error
		if gm.terrain, err = loadTerrainArt(c.terrain, c.candidates, c.lvl, m); err != nil {
			return nil, err
		}
		if c.terrainHD != "" {
			gm.hdTerrain = background("full-quality terrain", func() (*terrainArt, error) {
				return loadTerrainArt(c.terrainHD, c.candidates, c.lvl, &meter{pause: pagePauser()})
			})
		}
	}
	return gm, nil
}

func main() {
	name := flag.String("level", "monthly/1914-08:cp", "a scenario setup (monthly/1914-08:ec, skirmish/01-track:2p), original/NN for a map of the local original files, or the path of a level file of the legacy format (art/reference/pulse/level.json); skips the menu")
	dir := flag.String("levels", "levels", "directory with the level files")
	side := flag.String("player", "", "side the player plays and whose view is shown: central, entente (default: the level's side); skips the menu")
	rules := flag.String("rules", rulesClassic, "rule set: classic (default: the original rules) or modern (the original rules with extensions)")
	reveal := flag.Bool("reveal", false, "dev: show the whole map, ignoring fog of war (dos: every unit where its own side sees it)")
	hexNumbers := flag.Bool("hexnumbers", false, "dev: print line,col on every hex")
	reduced := flag.Bool("reduced-motion", false, "walk instantly and show short static effects (toggle in game with M)")
	units := flag.String("units", "", "unit sprite catalog built by hlasset build-units (default: placeholder dots)")
	terrain := flag.String("terrain", "", "terrain tile catalog built by hlasset build-terrain (default: coloured hexes)")
	screens := flag.String("screens", "", "screen catalog built by hlasset build-screens (default: plain menu panels)")
	opponent := flag.String("opponent", "ai", "who plays the other side: ai (computer opponent) or pass (passes every turn)")
	unitsHD := flag.String("units-hd", "", "web: full-quality unit catalog, loaded in the background when the -units catalog downloaded fast")
	terrainHD := flag.String("terrain-hd", "", "web: full-quality terrain catalog, loaded in the background of a battle when the -units catalog downloaded fast")
	screensHD := flag.String("screens-hd", "", "web: full-quality screen catalog, loaded in the background when the -units catalog downloaded fast")
	candidates := flag.Bool("allow-candidates", false, "allow unit, terrain and screen packs that are review candidates, not approved art")
	capture := flag.String("capture", "", "dev: render review frames of the battle screen to a directory under build/ and exit (requires -units and -terrain)")
	captureMap := flag.Bool("capture-map", false, "with -capture: render nine 720p/1080p/4K frames of the whole map instead (overview, detail, own view)")
	captureLook := flag.Bool("capture-look", false, "with -capture: render the whole map revealed and in the own view at exact 0.5x, 1x and 1.6x effective zoom at 1080p/4K instead")
	backdropPath := flag.String("backdrop", "", "a PNG or JPEG painting of the whole level's ground, drawn in place of the terrain tiles of every level, without a hex grid (must match the level's world aspect); off draws the tiles where a level has its own painting")
	flag.Parse()
	if (*captureLook || *captureMap) && *capture == "" {
		log.Fatal("-capture-look and -capture-map require -capture")
	}
	direct := *capture != ""
	flag.Visit(func(f *flag.Flag) { direct = direct || f.Name == "level" || f.Name == "player" })

	set, err := data.Load()
	if err != nil {
		log.Fatal(err)
	}
	if *opponent != "ai" && *opponent != "pass" {
		log.Fatalf("unknown -opponent %q (want ai or pass)", *opponent)
	}
	if *rules != rulesClassic && *rules != rulesModern {
		log.Fatalf("unknown -rules %q (want %s or %s)", *rules, rulesClassic, rulesModern)
	}
	c := matchConfig{
		set: set, rules: *rules, opponent: *opponent, reveal: *reveal, hexNumbers: *hexNumbers,
		reduced: *reduced, terrain: *terrain, candidates: *candidates, scores: scores.Open(), saves: saves.Open(), levels: *dir,
	}
	// The web page shows how much of the art has arrived until the game starts.
	boot := &meter{fetched: pageFetched}
	if *units != "" {
		if c.units, err = loadUnitArt(*units, *candidates, boot); err != nil {
			log.Fatal(err)
		}
	}
	// Only a fast connection loads the full-quality art as well.
	hd := hdDirs{units: *unitsHD, screens: *screensHD, terrain: *terrainHD}
	if !boot.fast() {
		hd = hdDirs{}
	}
	c.terrainHD = hd.terrain
	ebiten.SetWindowSize(startWidth, startHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Historyline 1914-1918")
	switch *backdropPath {
	case "":
	case "off":
		c.noBackdrop = true
	default:
		if c.backdrop, err = loadBackdrop(*backdropPath); err != nil {
			log.Fatal(err)
		}
	}
	if *screens != "" {
		if c.screens, err = loadScreenArt(*screens, *candidates, boot); err != nil {
			log.Fatal(err)
		}
		dospanel.SetSurfaces(c.screens.surfaces())
		dosscreen.SetArt(c.screens.frameArt())
	}

	if !direct {
		campaign, err := loadCampaign(*dir, set)
		if err != nil {
			log.Fatal(err)
		}
		pageReady()
		a := newApp(c, campaign, c.screens)
		a.hd = loadHDArt(hd, *candidates)
		if err := ebiten.RunGame(a); err != nil {
			log.Fatal(err)
		}
		return
	}

	if c.lvl, c.board, err = loadMap(*dir, *name, set); err != nil {
		log.Fatal(err)
	}
	if c.player, err = state.PlayerByName(c.lvl.Player); err != nil {
		log.Fatal(err)
	}
	switch *side {
	case "":
	case "central":
		c.player = state.CentralPowers
	case "entente":
		c.player = state.Entente
	default:
		log.Fatalf("unknown -player %q (want central or entente)", *side)
	}
	gm, err := newMatch(c, boot)
	if err != nil {
		log.Fatal(err)
	}
	if *capture != "" {
		if gm.art == nil || gm.terrain == nil {
			log.Fatal("-capture requires both -units and -terrain")
		}
		frames := splitFramePlan()
		switch {
		case *captureLook:
			frames = reviewFramePlan(true)
		case *captureMap:
			frames = reviewFramePlan(false)
		}
		if err := writeReviewFrames(gm, *capture, frames); err != nil {
			log.Fatal(err)
		}
		return
	}
	if hd.units != "" {
		gm.hdUnits = background("full-quality units", func() (*artPacks, error) {
			return loadUnitArt(hd.units, *candidates, &meter{pause: pagePauser()})
		})
	}
	pageReady()
	if err := ebiten.RunGame(gm); err != nil {
		log.Fatal(err)
	}
}
