//go:build dev

package main

// devBuild is true under `wails dev` (which builds with -tags dev): the updater must
// never replace a development binary.
const devBuild = true
