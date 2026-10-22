package assets

import (
	"fmt"
	"reflect"
	"testing"
	"testing/fstest"
)

func groupFiles(groups ...string) fstest.MapFS {
	f := fstest.MapFS{}
	for _, d := range []int{2, 4} {
		for _, g := range groups {
			f[fmt.Sprintf("%dx/%s/manifest.json", d, g)] = &fstest.MapFile{Data: []byte("{}")}
		}
	}
	return f
}

func TestPackGroups(t *testing.T) {
	for _, tc := range []struct {
		name   string
		groups []string
		want   []string
	}{
		{"legacy", []string{"backgrounds", "standing-stamps"}, []string{"backgrounds", "standing-stamps"}},
		{"parts", []string{"backgrounds-3", "standing-stamps-2", "backgrounds", "backgrounds-2", "standing-stamps"}, []string{"backgrounds", "backgrounds-2", "backgrounds-3", "standing-stamps", "standing-stamps-2"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PackGroups(groupFiles(tc.groups...), []string{"backgrounds"}, []string{"standing", "standing-stamps"}, []int{2, 4})
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("groups %v, %v; want %v", got, err, tc.want)
			}
		})
	}
}

func TestPackGroupsRejectsPartialCatalogues(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(fstest.MapFS)
	}{
		{"missing-density", func(f fstest.MapFS) { delete(f, "4x/backgrounds/manifest.json") }},
		{"unpaired-part", func(f fstest.MapFS) { f["2x/backgrounds-2/manifest.json"] = &fstest.MapFile{} }},
		{"missing-manifest", func(f fstest.MapFS) {
			f["2x/backgrounds-2/image.png"] = &fstest.MapFile{}
			f["4x/backgrounds-2/image.png"] = &fstest.MapFile{}
		}},
		{"gap", func(f fstest.MapFS) {
			for p, v := range groupFiles("backgrounds-3") {
				f[p] = v
			}
		}},
		{"optional-without-base", func(f fstest.MapFS) {
			for p, v := range groupFiles("standing-2") {
				f[p] = v
			}
		}},
		{"unknown", func(f fstest.MapFS) {
			for p, v := range groupFiles("other") {
				f[p] = v
			}
		}},
		{"leading-zero", func(f fstest.MapFS) {
			for p, v := range groupFiles("backgrounds-02") {
				f[p] = v
			}
		}},
		{"part-one", func(f fstest.MapFS) {
			for p, v := range groupFiles("backgrounds-1") {
				f[p] = v
			}
		}},
		{"enormous-part", func(f fstest.MapFS) {
			for p, v := range groupFiles("backgrounds-9223372036854775807") {
				f[p] = v
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := groupFiles("backgrounds")
			tc.change(f)
			if got, err := PackGroups(f, []string{"backgrounds"}, []string{"standing"}, []int{2, 4}); err == nil {
				t.Fatalf("accepted %v", got)
			}
		})
	}
}

func TestPackGroupsSingleDensity(t *testing.T) {
	f := groupFiles("backgrounds", "backgrounds-2")
	for p := range f {
		if p[:2] == "4x" {
			delete(f, p)
		}
	}
	got, err := PackGroups(f, []string{"backgrounds"}, nil, []int{2})
	if err != nil || len(got) != 2 {
		t.Fatalf("browser groups %v: %v", got, err)
	}
}
