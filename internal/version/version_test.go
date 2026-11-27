package version

import "testing"

func TestCompareSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v0.1.0", "v0.2.0", -1},
		{"0.1.0", "v0.1.0", 0},
		{"v1.2.3", "v1.2.2", 1},
		{"v1.0.0-rc.1", "v1.0.0", -1},
		{"v1.0.0", "v1.0.0-rc.1", 1},
		{"v0.9.9", "v0.10.0", -1},
	}
	for _, c := range cases {
		if got := CompareSemver(c.a, c.b); got != c.want {
			t.Errorf("CompareSemver(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	if IsNewer("dev", "v0.2.0") {
		t.Error("dev build should not report IsNewer=true")
	}
	if IsNewer("build-20261127-abc1234", "v0.2.0") {
		t.Error("build-* should not report IsNewer=true")
	}
	if !IsNewer("v0.1.0", "v0.2.0") {
		t.Error("v0.1.0 -> v0.2.0 should be newer")
	}
	if IsNewer("v0.2.0", "v0.2.0") {
		t.Error("same version should not be newer")
	}
}
