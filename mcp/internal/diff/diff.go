// Package diff provides deterministic patch creation and application for
// the per-file diff tree system.
//
// It uses github.com/sergi/go-diff (Google's diff-match-patch port) to
// produce serialisable patch text that can be stored in DiffBlob payloads
// and later applied to reconstruct candidate file content.
//
// Patch format: diff-match-patch patch text (deterministic, invertible).
package diff

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const Format = "dmp_patch"

// Engine wraps the diff-match-patch algorithm with a production-shaped API.
type Engine struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

// NewEngine returns a ready-to-use Engine.
func NewEngine() *Engine {
	dmp := diffmatchpatch.New()
	dmp.PatchMargin = 4
	return &Engine{dmp: dmp}
}

// CreatePatch computes a serialised patch from oldContent to newContent.
// The result can be persisted in a DiffBlob.Payload.
func (e *Engine) CreatePatch(oldContent, newContent string) string {
	diffs := e.dmp.DiffMain(oldContent, newContent, true)
	diffs = e.dmp.DiffCleanupEfficiency(diffs)
	patches := e.dmp.PatchMake(oldContent, diffs)
	return e.dmp.PatchToText(patches)
}

// ApplyPatch applies a serialised patch to baseContent and returns the
// resulting text. Returns an error if any hunk fails to apply.
func (e *Engine) ApplyPatch(baseContent, patchText string) (string, error) {
	patches, err := e.dmp.PatchFromText(patchText)
	if err != nil {
		return "", fmt.Errorf("parse patch text: %w", err)
	}
	if len(patches) == 0 {
		return baseContent, nil
	}

	result, applied := e.dmp.PatchApply(patches, baseContent)
	for i, ok := range applied {
		if !ok {
			return "", fmt.Errorf("hunk %d/%d failed to apply", i+1, len(applied))
		}
	}
	return result, nil
}

// ContentHash returns a hex-encoded SHA-256 hash of content.
func ContentHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

// HasChanges returns true if the two contents differ.
func (e *Engine) HasChanges(oldContent, newContent string) bool {
	return oldContent != newContent
}
