package artreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestMixedAcceptsExplicitMarkedRoadAndRejectsWrongWidth(t *testing.T) {
	for _, profile := range [][2]float64{{22, 1}, {22, 2}, {21, 1}, {22, 3}} {
		width, border := profile[0], profile[1]
		path, m := mixedFixture(t)
		component := filepath.Join(filepath.Dir(path), m.Manifests[0].Source)
		var source connectionManifest
		if err := json.Unmarshal(unitTestRead(t, component), &source); err != nil {
			t.Fatal(err)
		}
		for i := range source.Families {
			if source.Families[i].ID == "road" {
				source.Families[i].Width = width
				source.Families[i].Border = border
				source.Families[i].Markings = "dashed"
			}
		}
		data, err := json.Marshal(source)
		if err != nil {
			t.Fatal(err)
		}
		unitTestWrite(t, component, data)
		sum := sha256.Sum256(data)
		m.Manifests[0].SourceSHA256 = hex.EncodeToString(sum[:])
		data, err = json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		unitTestWrite(t, path, data)
		in, err := loadMixedInputs(path)
		if width == 22 && (border == 1 || border == 2) {
			if err != nil {
				t.Fatal(err)
			}
			if in.components["road36"].family.Markings != "dashed" {
				t.Fatal("mixed import lost road markings")
			}
		} else if err == nil {
			t.Fatal("incompatible marked road width accepted")
		}
	}
}
