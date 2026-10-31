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

func TestPortraitLightArtilleryBindingsAndExports(t *testing.T) {
	path, manifest, source := portraitTestFixture(t, 1024, 1024)
	manifest = bytes.Replace(manifest, []byte(`"family":"infantry"`), []byte(`"family":"lightArtillery"`), 1)
	unitTestWrite(t, path, manifest)
	out := filepath.Join(t.TempDir(), "artillery")
	if err := run([]string{"-portrait", path, "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	var report portraitReport
	if err := json.Unmarshal(unitTestRead(t, filepath.Join(out, "report.json")), &report); err != nil {
		t.Fatal(err)
	}
	if report.Family != "lightArtillery" || report.Profile != "portrait-light-artillery-v1" || !reflect.DeepEqual(report.Classes, []string{"CP_lArtillery", "EC_lArtillery"}) {
		t.Fatalf("wrong family bindings: %+v", report)
	}
	if !bytes.Equal(source, unitTestRead(t, filepath.Join(out, "lightArtillery-source.png"))) || !bytes.Equal(manifest, unitTestRead(t, filepath.Join(out, "manifest.json"))) {
		t.Fatal("source or manifest snapshot changed")
	}
	gallery := string(unitTestRead(t, filepath.Join(out, "index.html")))
	if strings.Contains(gallery, "infantry") || !strings.Contains(gallery, "CP_lArtillery") || !strings.Contains(gallery, "EC_lArtillery") {
		t.Fatal("gallery has stale or missing class bindings")
	}
	for _, output := range report.Outputs {
		if !strings.HasPrefix(output.Path, "lightArtillery-") || !strings.Contains(gallery, `href="`+output.Path+`"`) {
			t.Fatalf("wrong output/gallery path: %s", output.Path)
		}
		im := unitTestLoad(t, filepath.Join(out, output.Path))
		if im.Bounds().Dx() != output.Dimensions.Width || im.Bounds().Dy() != output.Dimensions.Height {
			t.Fatal("incorrect output dimensions")
		}
	}
}

func TestPortraitEliteInfantryBindingsAndExports(t *testing.T) {
	path, manifest, source := portraitTestFixture(t, 1024, 1024)
	manifest = bytes.Replace(manifest, []byte(`"family":"infantry"`), []byte(`"family":"eliteInfantry"`), 1)
	unitTestWrite(t, path, manifest)
	out := filepath.Join(t.TempDir(), "artillery")
	if err := run([]string{"-portrait", path, "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	var report portraitReport
	if err := json.Unmarshal(unitTestRead(t, filepath.Join(out, "report.json")), &report); err != nil {
		t.Fatal(err)
	}
	if report.Family != "eliteInfantry" || report.Profile != "portrait-elite-infantry-v1" || !reflect.DeepEqual(report.Classes, []string{"CP_eliteInf", "EC_eliteInf"}) {
		t.Fatalf("wrong elite family bindings: %+v", report)
	}
	if !bytes.Equal(source, unitTestRead(t, filepath.Join(out, "eliteInfantry-source.png"))) || !bytes.Equal(manifest, unitTestRead(t, filepath.Join(out, "manifest.json"))) {
		t.Fatal("source or manifest snapshot changed")
	}
	gallery := string(unitTestRead(t, filepath.Join(out, "index.html")))
	if strings.Contains(gallery, "CP_infantry") || !strings.Contains(gallery, "CP_eliteInf") || !strings.Contains(gallery, "EC_eliteInf") {
		t.Fatal("gallery has stale or missing class bindings")
	}
	for _, output := range report.Outputs {
		if !strings.HasPrefix(output.Path, "eliteInfantry-") || !strings.Contains(gallery, `href="`+output.Path+`"`) {
			t.Fatalf("wrong output/gallery path: %s", output.Path)
		}
		im := unitTestLoad(t, filepath.Join(out, output.Path))
		if im.Bounds().Dx() != output.Dimensions.Width || im.Bounds().Dy() != output.Dimensions.Height {
			t.Fatal("incorrect output dimensions")
		}
	}
}

func TestCompoundCanonicalManifestPaths(t *testing.T) {
	if err := compoundSafeRelative("source/portraits/reference-v1/light-artillery-01.png"); err != nil {
		t.Fatalf("canonical nested source rejected: %v", err)
	}
	for _, path := range []string{"", "../outside.png", "source/../outside.png", "./source.png", "source//image.png", `source\image.png`, "original/image.png", "source/PILOT/image.png", "/absolute.png"} {
		if err := compoundSafeRelative(path); err == nil {
			t.Errorf("unsafe or noncanonical path accepted: %q", path)
		}
	}
}
