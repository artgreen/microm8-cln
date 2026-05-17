package ui

// Glyph rune-string constants that are referenced from untagged code
// (ui.go) and must therefore be available under every build tag. The
// remaining Symbol* constants (CTRL/SHIFT/Option/Alt/ENTER/Backspace)
// stay in menu_m8.go because they are only consumed by that file's
// !nox-tagged menu hints.
//
// These values are pure data — Unicode codepoints assigned by the
// bundled font — so there is no variant-specific behavior to fork.
const (
	SymbolOff          = string(rune(256))
	SymbolOn           = string(rune(257))
	SymbolSliderHandle = string(rune(1154))
	SymbolSliderMark   = string(rune(1105))
)
