package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type connectionPoint [2]float64

func (p connectionPoint) add(q connectionPoint) connectionPoint {
	return connectionPoint{p[0] + q[0], p[1] + q[1]}
}

func (p connectionPoint) mul(s float64) connectionPoint         { return connectionPoint{p[0] * s, p[1] * s} }
func (p connectionPoint) sub(q connectionPoint) connectionPoint { return p.add(q.mul(-1)) }
func (p connectionPoint) length() float64                       { return math.Hypot(p[0], p[1]) }

type connectionGeometry struct {
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Polygon   []connectionPoint `json:"polygon"`
	Neighbors []connectionPoint `json:"neighbors"`
	Ports     []connectionPoint `json:"ports"`
}

func sharedConnectionGeometry() connectionGeometry {
	return connectionGeometry{96, 80, []connectionPoint{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}, []connectionPoint{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}, []connectionPoint{{83.5, 20}, {83.5, 60}, {48, 80}, {12.5, 60}, {12.5, 20}, {48, 0}}}
}

type connectionMaterial struct {
	Source       string `json:"source"`
	SourceSHA256 string `json:"source_sha256"`
}
type connectionVariant struct {
	ID              string `json:"id"`
	Ports           []int  `json:"ports"`
	ClosedRing      bool   `json:"closed_ring,omitempty"`
	BlindSpurs      []int  `json:"blind_spurs,omitempty"`
	DugoutDirection int    `json:"dugout_direction,omitempty"`
}
type connectionFamily struct {
	BankProfile    string              `json:"bank_profile,omitempty"`
	ID             string              `json:"id"`
	Width          float64             `json:"width"`
	Border         float64             `json:"border"`
	Material       string              `json:"material"`
	BorderMaterial string              `json:"border_material"`
	Variants       []connectionVariant `json:"variants"`
	Markings       string              `json:"markings,omitempty"`
	Verge          string              `json:"verge,omitempty"`
	Season         string              `json:"season,omitempty"`
}
type connectionManifest struct {
	SchemaVersion int                           `json:"schema_version"`
	Geometry      connectionGeometry            `json:"geometry"`
	Materials     map[string]connectionMaterial `json:"materials"`
	Families      []connectionFamily            `json:"families"`
}
type connectionMaterialReport struct {
	connectionMaterial
	Dimensions     dimensions `json:"dimensions"`
	RetainedSource string     `json:"retained_source"`
}
type connectionVariantReport struct {
	connectionVariant
	Family         string       `json:"family"`
	Width          float64      `json:"width"`
	Border         float64      `json:"border"`
	Material       string       `json:"material"`
	BorderMaterial string       `json:"border_material"`
	Outputs        []packOutput `json:"outputs"`
	Markings       string       `json:"markings,omitempty"`
	Verge          string       `json:"verge,omitempty"`
}
type connectionCell struct {
	Origin  connectionPoint `json:"origin"`
	Variant string          `json:"variant,omitempty"`
	Ports   []int           `json:"ports"`
}
type connectionScene struct {
	ID               string           `json:"id"`
	CenterLineParity int              `json:"center_line_parity"`
	Cells            []connectionCell `json:"cells"`
	Output           packOutput       `json:"output"`
	Output4x         packOutput       `json:"output_4x"`
}
type connectionReport struct {
	SchemaVersion    int                                 `json:"schema_version"`
	ManifestSHA256   string                              `json:"manifest_sha256"`
	Geometry         connectionGeometry                  `json:"geometry"`
	Materials        map[string]connectionMaterialReport `json:"materials"`
	Variants         []connectionVariantReport           `json:"variants"`
	Scenes           []connectionScene                   `json:"scenes"`
	Method           string                              `json:"method"`
	TextureMethod    string                              `json:"texture_method"`
	StreamInterior   string                              `json:"stream_interior"`
	PathInterior     string                              `json:"path_interior,omitempty"`
	RailProfile      *connectionRailProfile              `json:"rail_profile,omitempty"`
	TrenchProfile    *connectionTrenchProfile            `json:"trench_profile,omitempty"`
	CompositionScope string                              `json:"composition_scope"`
	Ownership        string                              `json:"ownership"`
	Acceptance       string                              `json:"acceptance"`
	Outputs          []packOutput                        `json:"outputs"`
	seasonReport
}
type connectionArtifact struct {
	name string
	data []byte
}
type connectionInputs struct {
	manifest      connectionManifest
	materials     map[string]*image.NRGBA
	sources       map[string][]byte
	hash          string
	grassSampling bool
}

// decodeConnectionJSON rejects duplicate keys as well as unknown schema fields.
func decodeConnectionJSON(data []byte, dst any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	var value func() error
	value = func() error {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				k, err := d.Token()
				if err != nil {
					return err
				}
				key, ok := k.(string)
				if !ok || seen[key] {
					return errors.New("duplicate JSON key")
				}
				seen[key] = true
				if err := value(); err != nil {
					return err
				}
			}
		case '[':
			for d.More() {
				if err := value(); err != nil {
					return err
				}
			}
		default:
			return errors.New("unexpected JSON delimiter")
		}
		_, err = d.Token()
		return err
	}
	if err := value(); err != nil {
		return err
	}
	if _, err := d.Token(); !errors.Is(err, io.EOF) {
		return errors.New("requires exactly one JSON object")
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	return d.Decode(dst)
}

func loadConnectionInputs(path string) (connectionInputs, error) {
	in := connectionInputs{materials: map[string]*image.NRGBA{}, sources: map[string][]byte{}, grassSampling: true}
	b, err := os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodeConnectionJSON(b, &in.manifest); err != nil {
		return in, fmt.Errorf("connected manifest: %w", err)
	}
	if err = validateConnectionTrenchFields(b); err != nil {
		return in, err
	}
	if err = validateConnectionRoadFields(b); err != nil {
		return in, err
	}
	m := &in.manifest
	if m.SchemaVersion != 1 || !reflect.DeepEqual(m.Geometry, sharedConnectionGeometry()) {
		return in, errors.New("connected requires schema_version 1 and the fixed authorized 96x80 connection geometry")
	}
	if len(m.Families) < 1 || len(m.Families) > 5 {
		return in, errors.New("connected requires one to five unique road/stream/path/rail/trench families")
	}
	families := map[string]bool{}
	gravelBank := false
	for _, f := range m.Families {
		if (f.ID != "road" && f.ID != "stream" && f.ID != "path" && f.ID != "rail" && f.ID != "trench") || families[f.ID] {
			return in, errors.New("family must be unique road, stream, path, rail or trench")
		}
		families[f.ID] = true
		gravelBank = gravelBank || f.ID == "stream" && f.BorderMaterial == "bank"
	}
	materialIDs := []string{"grass", "road", "water", "earth"}
	if gravelBank {
		materialIDs = append(materialIDs, "bank")
	}
	if families["path"] {
		materialIDs = append(materialIDs, "path")
	}
	if families["rail"] {
		materialIDs = append(materialIDs, "rail", "timber", "steel")
	}
	if families["trench"] {
		materialIDs = append(materialIDs, "trench")
		if !families["rail"] {
			materialIDs = append(materialIDs, "timber")
		}
	}
	if len(m.Materials) != len(materialIDs) {
		return in, errors.New("connected requires grass/road/water/earth materials, optional bank for gravel-bank stream, plus path for path, rail/timber/steel for rail and trench/timber for trench (shared timber once)")
	}
	sum := sha256.Sum256(b)
	in.hash = hex.EncodeToString(sum[:])
	for _, id := range materialIDs {
		material, ok := m.Materials[id]
		if !ok {
			return in, fmt.Errorf("missing material %s", id)
		}
		if material, _, err = seasonRef(sourceBase(path), material); err != nil {
			return in, fmt.Errorf("material %s: %w", id, err)
		}
		m.Materials[id] = material
		want, err := hex.DecodeString(material.SourceSHA256)
		if err != nil || len(want) != 32 {
			return in, fmt.Errorf("material %s requires SHA256", id)
		}
		if material.Source == "" || filepath.IsAbs(material.Source) {
			return in, fmt.Errorf("material %s source must be relative to the source root", id)
		}
		source := filepath.Join(sourceBase(path), material.Source)
		resolved, err := filepath.EvalSymlinks(source)
		if err != nil {
			return in, err
		}
		absolute, err := filepath.Abs(resolved)
		if err != nil {
			return in, err
		}
		for _, p := range []string{source, absolute} {
			for _, part := range strings.Split(filepath.ToSlash(p), "/") {
				if strings.EqualFold(part, "original") || strings.EqualFold(part, "pilot") {
					return in, errors.New("original and pilot sources are excluded from connected mode")
				}
			}
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			return in, err
		}
		sum := sha256.Sum256(data)
		if !bytes.Equal(sum[:], want) {
			return in, fmt.Errorf("material %s source SHA256 mismatch", id)
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return in, err
		}
		if cfg.Width < 384 || cfg.Height < 320 || int64(cfg.Width)*int64(cfg.Height) > maxInputPixels {
			return in, fmt.Errorf("material %s must be at least 384x320 and at most 64 megapixels", id)
		}
		decoded, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return in, err
		}
		im := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
		for y := 0; y < cfg.Height; y++ {
			for x := 0; x < cfg.Width; x++ {
				p := color.NRGBAModel.Convert(decoded.At(x, y)).(color.NRGBA)
				if p.A != 255 {
					return in, fmt.Errorf("material %s must be opaque", id)
				}
				im.SetNRGBA(x, y, p)
			}
		}
		in.materials[id] = im
		in.sources[id] = data
	}
	safe := regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	ids := map[string]bool{}
	outputNames := map[string]bool{"grass-2x.png": true, "grass-4x.png": true, "contact-sheet.png": true}
	count := 0
	for _, f := range m.Families {
		expected := f.ID
		if f.BankProfile != "" && (f.BankProfile != authoredBankProfile || f.ID != "stream" || f.BorderMaterial != "bank") {
			return in, errors.New("bank_profile requires stream with bank border and authored-rocks-v1")
		}
		if f.ID == "stream" {
			expected = "water"
		}
		if f.Material != expected || f.BorderMaterial != "earth" && (f.ID != "stream" || f.BorderMaterial != "bank") {
			return in, errors.New("road/stream/path/rail/trench must use road/water/path/rail/trench and earth border materials; stream alone may use bank")
		}
		if math.IsNaN(f.Width) || math.IsInf(f.Width, 0) || math.IsNaN(f.Border) || math.IsInf(f.Border, 0) || f.Width < 2 || f.Width > 28 || f.Border < 0 || f.Border > 6 || f.Width/2+f.Border > 20 {
			return in, errors.New("profile requires width 2..28, border 0..6, outer radius <=20")
		}
		if f.ID == "rail" && f.Width < 8 {
			return in, errors.New("rail profile requires core width >=8 world px for its rails and sleepers")
		}
		if f.ID == "trench" && (f.Width < 4 || f.Width > 10 || f.Border < 1.5 || f.Border > 5) {
			return in, errors.New("trench profile requires width 4..10 and border 1.5..5 world px")
		}
		if err := validateConnectionSeason(f); err != nil {
			return in, err
		}
		if len(f.Variants) == 0 {
			return in, errors.New("family requires variants")
		}
		for _, v := range f.Variants {
			count++
			if !safe.MatchString(v.ID) || ids[v.ID] || v.ID == "grass" || v.ID == "contact-sheet" {
				return in, fmt.Errorf("unsafe or duplicate variant id %q", v.ID)
			}
			ids[v.ID] = true
			for _, suffix := range []string{"-2x.png", "-4x.png", "-grass-2x.png", "-patch-parity0.png", "-patch-parity1.png", "-patch-parity0-4x.png", "-patch-parity1-4x.png"} {
				name := v.ID + suffix
				if outputNames[name] {
					return in, fmt.Errorf("derived output name collision: %s", name)
				}
				outputNames[name] = true
			}
			if err := validateConnectionTrenchVariant(f.ID, v); err != nil {
				return in, err
			}
			if f.ID != "trench" && (len(v.Ports) < 2 || len(v.Ports) > 6) {
				return in, errors.New("variants require 2..6 explicit ports")
			}
			ports := map[int]bool{}
			for _, p := range v.Ports {
				if p < 1 || p > 6 || ports[p] {
					return in, errors.New("ports must be unique integers 1..6")
				}
				ports[p] = true
			}
			if f.ID == "rail" && !validConnectionRailPorts(v.Ports) {
				return in, errors.New("rail requires two ports or a 3/6 through spine with one branch")
			}
		}
	}
	if count > 64 {
		return in, errors.New("connected mode is bounded to 64 variants")
	}
	return in, nil
}

func runConnected(path, out string, log io.Writer) (resultErr error) {
	if _, err := os.Lstat(out); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return errors.New("connected output directory already exists")
		}
		return err
	}
	in, err := loadConnectionInputs(path)
	if err != nil {
		return err
	}
	if err = seasonConnectionVariants(&in.manifest); err != nil {
		return err
	}
	report := connectionReport{
		SchemaVersion: 1, seasonReport: currentSeasonReport(), ManifestSHA256: in.hash, Geometry: in.manifest.Geometry, Materials: map[string]connectionMaterialReport{},
		Method:           "Candidate overlay. Unclipped new feature and grass render at 4x with four WORLD pixels of overscan on every side (416x352). Feature/bank edges use analytic signed-distance AA, 0.25 world-pixel ramp; distances round to 1e-9 world px for reciprocal half-alpha stability. Existing alpha-aware Lanczos3 reduces the UNMASKED overscan 416x352 -> 208x176. Crop 4-world-pixel margins AFTER filtering to 384x320 and 192x160, then apply complementary hard half-open polygon ownership at final pixel centers. No per-sprite polygon-edge AA; a future complete-map compositor may AA its outer boundary once. Each port has >=8 straight world pixels and extends beyond cutout by outer radius+4; cubic connectors flatten at 64 steps; junctions merge by minimum distance. No legacy footprint or pixels. Optional road markings=dashed: 0.9-world-pixel light centre dashes and 0.6-world-pixel continuous edge lines one world pixel inside each pavement edge, eight-world-pixel nominal dash length and twelve-world-pixel nominal period; eight-world-pixel port phases are reciprocal, curved interiors fit the phase, junction centres remain unmarked. Narrow dusty shoulders blend the earth material with RGB(177,168,145) at weight.35.",
		TextureMethod:    "New opaque source images only. Non-grass materials retain bilinear mirror-periodic coordinates, periods (142,80), and two-coset averaging. All independent cells use world origin phase (0,0). No original-game pixels. " + grassSamplingProfile,
		StreamInterior:   "Stream-only organic interior. Let S(t)=clamp(t,0,1)^2*(3-2*clamp(t,0,1)), E=S((minimum signed polygon-edge distance-8)/8), A=2pi(x/142+y/80), B=2pi(x/142-y/80), using world x/y. Warp local route lookup by E*(1.2*sin(A),1.2*sin(B)); core radius=width/2+0.8*E*sin(A+0.4)*cos(B). Thus the 8-world-pixel edge zone is unchanged and the envelope reaches full effect at 16 px. For junctions, smoothly union core SDF with a central disc of radius 0.68*width at warped coordinates: h=max(3-abs(core-disc),0)/3, merged=min(core,disc)-3*h*h/4, core=core+E*(merged-core). Border=max(0,border-0.8*E+0.3*E*sin(B+0.7)). Pocket=S((sin(A)*sin(B)-0.15)/0.65); bank opacity=1-0.65*E*pocket; final alpha=outer_AA*(core_AA+(1-core_AA)*bank_opacity). Cores stay opaque; translucent earth banks reveal underlying terrain with no baked-in grass. Road arithmetic and protected stream profiles are unchanged.",
		CompositionScope: "Tested composition scope: pixel-aligned 2x/4x export sprites with integer runtime lattice translations and ordinary source-over blits. Other camera zoom should assemble the scene at asset density and resample the completed scene once. Independently filtered or fractional-position sprite zoom is not validated; no runtime zoom implementation is included.",
		Ownership:        "Binary final-pixel polygon ownership: cross(edge,point)>0; on an edge include dy<0 or dy==0 && dx>0 (top-left rule), exclude the reciprocal edge. Ground and overlay exports share this new polygon. Zero RGBA outside ownership. Gallery patches decode exported PNG bytes and assemble ground then overlays using ordinary image.Draw Over at integer runtime cell translations, at BOTH 2x and 4x. They do not rerender world geometry or conceal sprite assembly gaps.",
		Acceptance:       "Candidate only. Actual painted joins, banks, curves, texture repetition and runtime appearance require visual review; no seamlessness or acceptance claim.",
	}
	report.Method += " Marked-road verge only: add 0.65*E*(0.65*sin(3A+2B)+0.35*sin(7A-5B)) world pixels to the configured shoulder, where E=S((minimum polygon edge distance-8)/8), A=2pi(x/142+y/80), B=2pi(x/142-y/80). Pavement width and road markings are unchanged; the first eight world pixels at every port retain fixed shoulder width."
	for _, f := range in.manifest.Families {
		if f.Markings == "ruts" && f.ID == "road" {
			report.Method += fmt.Sprintf(" Road markings=ruts replace the dashes and edge lines: two slush ruts %.1f world px either side of every course's centre line blend the surface toward RGB(%d,%d,%d) at weight %.2f, full within %.1f world px and fading by smoothstep over a further %.1f.", connectionRutOffset, connectionRutColor.R, connectionRutColor.G, connectionRutColor.B, connectionRutStrength, connectionRutCore, connectionRutFade)
		}
		switch {
		case f.ID == "path" && f.Markings == "ruts" && f.Verge == "":
			offset, core, fade := connectionPathRutProfile(f.Width)
			report.Method += fmt.Sprintf(" Path markings=ruts: two faint cart ruts %.1f world px either side of the centre line (full within %.1f, fading over %.1f) blend the surface toward the earth border material at weight %.2f; the earth verge is unchanged.", offset, core, fade, connectionPathSummerRutStrength)
		case f.Verge == "snow" && f.ID == "path":
			offset, core, fade := connectionPathRutProfile(f.Width)
			report.Method += fmt.Sprintf(" Path markings=ruts, verge=snow: two cart ruts %.1f world px either side of the centre line (full within %.1f, fading over %.1f) and the earth border blended toward the sampled grass (winter ground) material at weight %.2f.", offset, core, fade, connectionSnowVergeMix)
		case f.Verge == "snow" && f.ID == "rail":
			report.Method += fmt.Sprintf(" Rail verge=snow: the ballast blends toward the sampled grass (winter ground) material at weight %.2f between the sleepers, rising by smoothstep to %.2f on the shoulders beyond the sleeper ends and on the earth border; sleepers blend toward it at %.2f; steel darkens toward RGB(%d,%d,%d) at weight %.2f.", connectionRailSnowBetween, connectionSnowVergeMix, connectionRailSnowDust, connectionRailWinterSteel.R, connectionRailWinterSteel.G, connectionRailWinterSteel.B, connectionRailWinterSteelMix)
		case f.Verge == "snow":
			report.Method += fmt.Sprintf(" Road verge=snow: the shoulders blend the earth material with RGB(%d,%d,%d) at weight %.2f instead of the dusty tint.", connectionSnowShoulder.R, connectionSnowShoulder.G, connectionSnowShoulder.B, connectionSnowShoulderMix)
		}
	}
	if in.materials["bank"] != nil {
		report.StreamInterior += " Opt-in stream border_material=bank overrides the earth-bank treatment: pinned bank material is retained independently of earth. Keep the same organic core geometry and protected port width; border=max(0,border+1.1*E*sin(3A+2B)), bank opacity=1-0.2*E*pocket, giving a 2.4..4.6 world-pixel interior bank for configured border3.5. Legacy earth-stream border2.5 remains a separate unchanged profile; optional bank-stream border2.5 remains supported. The river-bank material may use an unchanged-pixel crop of its parent master to make loose stones larger; other lakeside variants retain their full master. Water color blends toward RGB(84,135,139) at weight0.24 near the bank. Dark wet lip is bank blended toward RGB(35,50,43) at weight0.72 over signed core distance -0.45..1.05. Water-side current flecks and broken foam use only lattice-periodic A/B, a narrow Gaussian centered1.2 world pixels inside water (width0.45), crest=S((sin(6A+4B+0.8sin(3B))-0.15)/0.6), and foam RGB(215,231,226) at maximumweight0.72. Full pale bank replaces the wet lip by distance1.05 outside water. Signed color distance rounds1e-9 for reciprocal stability; no baked grass, no lakeside mask changes."
	}
	for _, f := range in.manifest.Families {
		report.Method += connectionSeasonDescription(f)
	}
	for _, f := range in.manifest.Families {
		if f.BankProfile == authoredBankProfile {
			report.StreamInterior += authoredBankDescription
			report.TextureMethod += authoredBankDescription
		}
	}
	var artifacts []connectionArtifact
	exports := map[string]map[int]*image.NRGBA{}
	pngPixels := map[string]*image.NRGBA{}
	addPNG := func(name string, im *image.NRGBA) (packOutput, error) {
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			return packOutput{}, err
		}
		decoded, err := png.Decode(bytes.NewReader(b.Bytes()))
		if err != nil {
			return packOutput{}, err
		}
		pixels := image.NewNRGBA(im.Bounds())
		draw.Draw(pixels, pixels.Bounds(), decoded, image.Point{}, draw.Src)
		pngPixels[name] = pixels
		sum := sha256.Sum256(b.Bytes())
		o := packOutput{name, hex.EncodeToString(sum[:]), dimensions{im.Bounds().Dx(), im.Bounds().Dy()}}
		artifacts = append(artifacts, connectionArtifact{name, b.Bytes()})
		report.Outputs = append(report.Outputs, o)
		return o, nil
	}
	for _, id := range []string{"grass", "road", "water", "earth", "bank", "path", "rail", "timber", "steel", "trench"} {
		mat := in.materials[id]
		if mat == nil {
			continue
		}
		width := 0.
		for _, f := range in.manifest.Families {
			if f.ID == id {
				width = f.Width
			}
		}
		if id == "rail" {
			profile := connectionRailProfileFor(width)
			report.RailProfile = &profile
		}
		if id == "trench" {
			profile := fixedConnectionTrenchProfile()
			report.TrenchProfile = &profile
		}
		if id == "path" {
			report.PathInterior = "Path-only organic interior. Let S(t)=clamp(t,0,1)^2*(3-2*clamp(t,0,1)), E=S((minimum signed polygon-edge distance-8)/8), A=2pi(x/142+y/80), B=2pi(x/142-y/80), using world x/y. Warp local route lookup by E*(0.55*sin(A),0.55*sin(B)), displacement length <=0.55*sqrt(2)*E world px. Wear=0.2*sin(5*A+2*B+0.7)*sin(3*B-A); core radius=width/2+E*(0.4*sin(A+0.4)*cos(B)+wear). For 3+-port junctions smoothly union core SDF with a central disc of radius 0.65*width at warped coordinates: h=max(1.5-abs(core-disc),0)/1.5, merged=min(core,disc)-1.5*h*h/4, core=core+E*(merged-core). Border=border*(1-0.45*E+0.25*E*sin(B+0.7)). Pocket=S((sin(2*A+0.3)*cos(3*B-0.4)+0.15)/0.9); bank opacity=1-0.75*E*pocket; final alpha=outer_AA*(core_AA+(1-core_AA)*bank_opacity). The 8-world-pixel edge zone is unchanged; full variation begins at 16 px. Cores stay opaque; irregular translucent earth verges reveal underlying terrain without baked-in grass. Road and stream arithmetic is unchanged."
			if k := connectionPathScale(width); k != 1 {
				report.PathInterior += fmt.Sprintf(" This path is %g world px wide: the warp, radius and wear amplitudes and the junction union span are those above scaled by width/%g = %.3g, and its cart ruts lie %.3g world px either side of the centre line.", width, connectionPathBaseWidth, k, connectionPathRutOffset*k)
			}
		}
		name := id + "-material-source.png"
		artifacts = append(artifacts, connectionArtifact{name, in.sources[id]})
		report.Materials[id] = connectionMaterialReport{in.manifest.Materials[id], dimensions{mat.Bounds().Dx(), mat.Bounds().Dy()}, name}
	}
	grass, err := exportConnectionTiles(in, nil, nil)
	if err != nil {
		return err
	}
	if _, err = addPNG("grass-4x.png", grass[4]); err != nil {
		return err
	}
	if _, err = addPNG("grass-2x.png", grass[2]); err != nil {
		return err
	}
	exports["grass"] = map[int]*image.NRGBA{2: pngPixels["grass-2x.png"], 4: pngPixels["grass-4x.png"]}
	var panels []*image.NRGBA
	for _, f := range in.manifest.Families {
		for _, v := range f.Variants {
			tiles, err := exportConnectionTiles(in, &f, &v)
			if err != nil {
				return err
			}
			four, two := tiles[4], tiles[2]
			vr := connectionVariantReport{connectionVariant: v, Family: f.ID, Width: f.Width, Border: f.Border, Material: f.Material, BorderMaterial: f.BorderMaterial, Markings: f.Markings, Verge: f.Verge}
			for _, a := range []struct {
				name string
				im   *image.NRGBA
			}{{v.ID + "-4x.png", four}, {v.ID + "-2x.png", two}} {
				o, err := addPNG(a.name, a.im)
				if err != nil {
					return err
				}
				vr.Outputs = append(vr.Outputs, o)
			}
			exports[v.ID] = map[int]*image.NRGBA{2: pngPixels[v.ID+"-2x.png"], 4: pngPixels[v.ID+"-4x.png"]}
			composite := image.NewNRGBA(two.Bounds())
			draw.Draw(composite, composite.Bounds(), exports["grass"][2], image.Point{}, draw.Over)
			draw.Draw(composite, composite.Bounds(), exports[v.ID][2], image.Point{}, draw.Over)
			if _, err = addPNG(v.ID+"-grass-2x.png", composite); err != nil {
				return err
			}
			panels = append(panels, two, composite)
			report.Variants = append(report.Variants, vr)
			for parity := 0; parity < 2; parity++ {
				scene, err := buildConnectionScene(in.manifest.Geometry, f, v, parity)
				if err != nil {
					return err
				}
				report.Scenes = append(report.Scenes, scene)
			}
		}
	}
	for i := range report.Scenes {
		scene := &report.Scenes[i]
		for _, scale := range []int{2, 4} {
			name := scene.ID + ".png"
			if scale == 4 {
				name = scene.ID + "-4x.png"
			}
			o, err := addPNG(name, renderConnectionScene(exports, *scene, scale))
			if err != nil {
				return err
			}
			if scale == 2 {
				scene.Output = o
			} else {
				scene.Output4x = o
			}
		}
	}
	if _, err = addPNG("contact-sheet.png", connectionContactSheet(panels)); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	artifacts = append(artifacts, connectionArtifact{"report.json", append(data, '\n')})
	tmpl, err := template.New("connected").Parse(connectionGallery)
	if err != nil {
		return err
	}
	var html bytes.Buffer
	if err = tmpl.Execute(&html, report); err != nil {
		return err
	}
	artifacts = append(artifacts, connectionArtifact{"index.html", html.Bytes()})
	if err = os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err = os.Mkdir(out, 0o755); err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, os.RemoveAll(out))
		}
	}()
	for _, a := range artifacts {
		if err = saveExclusive(filepath.Join(out, a.name), func(w io.Writer) error { _, err := w.Write(a.data); return err }); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(log, "Wrote %d connected candidates and %d parity patches to %s; visual acceptance pending.\n", len(report.Variants), len(report.Scenes), out)
	return err
}

const connectionGallery = `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Connected material candidates</title><style>body{background:#242923;color:#eee;font:16px system-ui;margin:2rem}a{color:#c7e6ad}section{display:flex;flex-wrap:wrap;gap:1rem}img{max-width:100%;height:auto}.overlay{background:repeating-conic-gradient(#777 0% 25%,#555 0% 50%) 0/16px 16px}article{border-top:1px solid #888;padding:1rem 0}</style><h1>Connected material candidates</h1><p>{{.Acceptance}}</p><p>{{.Ownership}}</p><p>{{.Method}}</p><p>{{.StreamInterior}}</p>{{if .PathInterior}}<p>{{.PathInterior}}</p>{{end}}{{if .RailProfile}}<p>{{.RailProfile.Method}} {{.RailProfile.Phase}} {{.RailProfile.Turnouts}}</p>{{end}}<p>{{.CompositionScope}}</p><p><a href="report.json">Geometry, provenance, topology and output hashes</a></p><img src="contact-sheet.png" alt="Overlay and grass composite pairs">{{range .Variants}}<article><h2>{{.ID}} · {{.Family}} · ports {{.Ports}}</h2><p>Core {{.Width}} world px; earth border {{.Border}} world px per side</p><section><a href="{{.ID}}-4x.png"><img class="overlay" src="{{.ID}}-2x.png" alt="{{.ID}} transparent overlay"></a><img src="{{.ID}}-grass-2x.png" alt="{{.ID}} on grass"></section></article>{{end}}<h2>Seven-cell patches · both line parities</h2>{{range .Scenes}}<article><h3>{{.ID}}</h3><a href="{{.Output4x.Path}}"><img src="{{.Output.Path}}" alt="{{.ID}}"></a><p>2× PNG assembly preview; open image for 4× assembly.</p><p>All interior declared ports are reciprocal. Outer patch continuations leave the patch.</p></article>{{end}}<h2>Full new material sources</h2><section>{{range $id,$m:=.Materials}}<figure><a href="{{$m.RetainedSource}}"><img style="width:192px" src="{{$m.RetainedSource}}" alt="{{$id}} new material"></a><figcaption>{{$id}} · {{$m.Dimensions.Width}}×{{$m.Dimensions.Height}}</figcaption></figure>{{end}}</section></html>`

// seasonConnectionVariants suffixes the variant IDs, and so the runtime files,
// of a seasonal export; winter families have no autumn or spring form.
func seasonConnectionVariants(m *connectionManifest) error {
	if !seasonActive() {
		return nil
	}
	for i := range m.Families {
		f := &m.Families[i]
		for j := range f.Variants {
			if f.Season == "winter" || f.Verge == "snow" || strings.HasSuffix(f.Variants[j].ID, "_w") {
				return errSeasonWinter
			}
			f.Variants[j].ID = seasonID(f.Variants[j].ID)
		}
	}
	return nil
}

// Keep sorting deterministic when selecting equivalent neighbor variants.
func sortedConnectionVariants(vs []connectionVariant) []connectionVariant {
	out := append([]connectionVariant(nil), vs...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
