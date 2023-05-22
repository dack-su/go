// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gover

import (
	"sort"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// IsToolchain reports whether the module path corresponds to the
// virtual, non-downloadable module tracking go or toolchain directives in the go.mod file.
//
// Note that IsToolchain only matches "go" and "toolchain", not the
// real, downloadable module "golang.org/toolchain" containing toolchain files.
//
//	IsToolchain("go") = true
//	IsToolchain("toolchain") = true
//	IsToolchain("golang.org/x/tools") = false
//	IsToolchain("golang.org/toolchain") = false
func IsToolchain(path string) bool {
	return path == "go" || path == "toolchain"
}

// ModCompare returns the result of comparing the versions x and y
// for the module with the given path.
// The path is necessary because the "go" and "toolchain" modules
// use a different version syntax and semantics (gover, this package)
// than most modules (semver).
func ModCompare(path string, x, y string) int {
	if path == "go" {
		return Compare(x, y)
	}
	if path == "toolchain" {
		return Compare(untoolchain(x), untoolchain(y))
	}
	return semver.Compare(x, y)
}

// ModSort is like module.Sort but understands the "go" and "toolchain"
// modules and their version ordering.
func ModSort(list []module.Version) {
	sort.Slice(list, func(i, j int) bool {
		mi := list[i]
		mj := list[j]
		if mi.Path != mj.Path {
			return mi.Path < mj.Path
		}
		// To help go.sum formatting, allow version/file.
		// Compare semver prefix by semver rules,
		// file by string order.
		vi := mi.Version
		vj := mj.Version
		var fi, fj string
		if k := strings.Index(vi, "/"); k >= 0 {
			vi, fi = vi[:k], vi[k:]
		}
		if k := strings.Index(vj, "/"); k >= 0 {
			vj, fj = vj[:k], vj[k:]
		}
		if vi != vj {
			return ModCompare(mi.Path, vi, vj) < 0
		}
		return fi < fj
	})
}

// ModIsValid reports whether vers is a valid version syntax for the module with the given path.
func ModIsValid(path, vers string) bool {
	if IsToolchain(path) {
		return parse(vers) != (version{})
	}
	return semver.IsValid(vers)
}

// untoolchain converts a toolchain name like "go1.2.3" to a Go version like "1.2.3".
// It also converts "anything-go1.2.3" (for example, "gccgo-go1.2.3") to "1.2.3".
func untoolchain(x string) string {
	if strings.HasPrefix(x, "go1") {
		return x[len("go"):]
	}
	if i := strings.Index(x, "-go1"); i >= 0 {
		return x[i+len("-go"):]
	}
	return x
}
