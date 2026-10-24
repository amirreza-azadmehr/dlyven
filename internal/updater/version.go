package updater

import (
	"strconv"
	"strings"
)

// Newer reports whether version a is newer than b. Versions look like "1.2.3", with an
// optional "v" prefix and "-suffix". A suffixed build (1.2.3-pr.4) is older than 1.2.3,
// the way semver orders pre-releases. Unparseable versions are never newer.
func Newer(a, b string) bool {
	ac, apre, aok := parseVersion(a)
	bc, bpre, bok := parseVersion(b)
	if !aok || !bok {
		return false
	}
	for i := range ac {
		if ac[i] != bc[i] {
			return ac[i] > bc[i]
		}
	}
	return apre == "" && bpre != ""
}

func parseVersion(v string) (core [3]int, pre string, ok bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	v, pre, _ = strings.Cut(v, "-")
	parts := strings.Split(v, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return core, "", false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return core, "", false
		}
		core[i] = n
	}
	return core, pre, true
}
