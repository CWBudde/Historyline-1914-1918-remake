package artreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestMixedAcceptsWideGravelStreamKeepsLegacyEarthProfile(t *testing.T) {
	for _, tc := range []struct {
		bank   string
		border float64
		valid  bool
	}{{"earth", 2.5, true}, {"bank", 2.5, true}, {"bank", 3.5, true}, {"earth", 3.5, false}, {"bank", 4.5, false}} {
		path, m := mixedFixture(t)
		component := filepath.Join(filepath.Dir(path), m.Manifests[0].Source)
		var source connectionManifest
		if err := json.Unmarshal(unitTestRead(t, component), &source); err != nil {
			t.Fatal(err)
		}
		for i := range source.Families {
			if source.Families[i].ID == "stream" {
				source.Families[i].Border = tc.border
				source.Families[i].BorderMaterial = tc.bank
			}
		}
		if tc.bank == "bank" {
			source.Materials["bank"] = source.Materials["earth"]
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
		if (err == nil) != tc.valid {
			t.Fatalf("%s border%g valid%v: %v", tc.bank, tc.border, tc.valid, err)
		}
		if tc.valid && (in.components["stream36"].family.Border != tc.border || in.components["stream36"].family.BorderMaterial != tc.bank) {
			t.Fatal("mixed import lost gravel profile")
		}
	}
}
