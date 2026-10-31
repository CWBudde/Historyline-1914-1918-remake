package artreview

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Exercise the actual CLI and saved images, including single-class families.
func TestPortraitRemainingFamilyBindingsAndExports(t *testing.T) {
	for _, tc := range []struct {
		family, profile string
		classes         []string
	}{
		{"cavalry", "portrait-cavalry-v1", []string{"CP_cavalry", "EC_cavalry"}},
		{"mediumArtillery", "portrait-medium-artillery-v1", []string{"CP_mArtillery", "EC_mArtillery"}},
		{"heavyArtillery", "portrait-heavy-artillery-v1", []string{"CP_hArtillery", "EC_hArtillery"}},
		{"supplyCar", "portrait-supply-car-v1", []string{"CP_supplyCar", "EC_supplyCar"}},
		{"bunker", "portrait-bunker-v1", []string{"CP_bunker", "EC_bunker"}},
		{"armoredCar", "portrait-armored-car-v1", []string{"CP_armCar", "EC_armCar"}},
		{"charron", "portrait-charron-v1", []string{"EC_charron"}},
		{"patrolBoat", "portrait-patrol-boat-v1", []string{"CP_patrolBoat", "EC_patrolBoat"}},
		{"transportTrain", "portrait-transport-train-v1", []string{"CP_transTrain", "EC_transTrain"}},
		{"armoredTrain", "portrait-armored-train-v1", []string{"CP_armTrain", "EC_armTrain"}},
		{"trainArtillery", "portrait-train-artillery-v1", []string{"CP_trainArt", "EC_trainArt"}},
		{"sappers", "portrait-sappers-v1", []string{"CP_sappers", "EC_sappers"}},
	} {
		t.Run(tc.family, func(t *testing.T) {
			path, manifest, source := portraitTestFixture(t, 1024, 1024)
			manifest = bytes.Replace(manifest, []byte(`"family":"infantry"`), []byte(`"family":"`+tc.family+`"`), 1)
			unitTestWrite(t, path, manifest)
			out := filepath.Join(t.TempDir(), "candidate")
			if err := run([]string{"-portrait", path, "-out", out}, io.Discard); err != nil {
				t.Fatal(err)
			}
			var report portraitReport
			if err := json.Unmarshal(unitTestRead(t, filepath.Join(out, "report.json")), &report); err != nil {
				t.Fatal(err)
			}
			if report.Family != tc.family || report.Profile != tc.profile || !reflect.DeepEqual(report.Classes, tc.classes) || len(report.Outputs) != 3 {
				t.Fatalf("incorrect bindings: %+v", report)
			}
			if !bytes.Equal(source, unitTestRead(t, filepath.Join(out, tc.family+"-source.png"))) || !bytes.Equal(manifest, unitTestRead(t, filepath.Join(out, "manifest.json"))) {
				t.Fatal("source/manifest snapshot changed")
			}
			gallery := string(unitTestRead(t, filepath.Join(out, "index.html")))
			for _, class := range tc.classes {
				if !strings.Contains(gallery, class) {
					t.Fatal("missing gallery binding", class)
				}
			}
			for _, output := range report.Outputs {
				if !strings.HasPrefix(output.Path, tc.family+"-") || !strings.Contains(gallery, `href="`+output.Path+`"`) {
					t.Fatal("wrong export path", output.Path)
				}
				im := unitTestLoad(t, filepath.Join(out, output.Path))
				if im.Bounds().Dx() != output.Dimensions.Width || im.Bounds().Dy() != output.Dimensions.Height {
					t.Fatal("incorrect saved dimensions")
				}
			}
		})
	}
}
