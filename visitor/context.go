// Copyright 2026 Princess Beef Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package visitor

import (
	"context"
	"fmt"
	"strings"
)

// ctxKey is a private typed key to avoid context collisions.
type ctxKey int

const (
	depthKey ctxKey = iota
	pathKey
	stackKey
	parentKey
)

// Depth returns the current traversal depth from context.
// Returns 0 if not set.
func Depth(ctx context.Context) int {
	if v := ctx.Value(depthKey); v != nil {
		return v.(int)
	}
	return 0
}

// WithDepth returns a context with the given depth.
func WithDepth(ctx context.Context, depth int) context.Context {
	return context.WithValue(ctx, depthKey, depth)
}

// Path returns the current JSON Pointer path (RFC 6901) from context.
// Root is empty string "". First segment creates "/segment".
// Example: "/channels/userSignup/messages/signupMessage"
func Path(ctx context.Context) string {
	if v := ctx.Value(pathKey); v != nil {
		return v.(string)
	}
	return "" // Root is empty string, not "/"
}

// WithPath returns a context with the given path.
func WithPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, pathKey, path)
}

// AppendPath appends a segment to the current path and returns a new context.
// Handles empty root correctly: "" + "info" = "/info"
// Escapes ~ as ~0 and / as ~1 per RFC 6901.
// Guards against empty segments to avoid trailing slashes.
func AppendPath(ctx context.Context, segment string) context.Context {
	if segment == "" {
		return ctx // No-op for empty segment, avoids trailing "/"
	}
	escaped := escapeJSONPointer(segment)
	current := Path(ctx)
	// Always prefix with / - works for both empty root and existing paths
	return WithPath(ctx, current+"/"+escaped)
}

// AppendIndex appends an array index to the current path.
func AppendIndex(ctx context.Context, index int) context.Context {
	current := Path(ctx)
	return WithPath(ctx, fmt.Sprintf("%s/%d", current, index))
}

// escapeJSONPointer escapes a path segment per RFC 6901.
// ~ becomes ~0, / becomes ~1 (order matters: ~ first)
func escapeJSONPointer(segment string) string {
	// fast path: no escaping needed (common case)
	if !strings.ContainsAny(segment, "~/") {
		return segment
	}
	segment = strings.ReplaceAll(segment, "~", "~0")
	segment = strings.ReplaceAll(segment, "/", "~1")
	return segment
}

// withChild returns a context with incremented depth and updated parent.
// This is a convenience helper for the common pattern in walk functions.
func withChild(ctx context.Context, node any) context.Context {
	return WithParent(WithDepth(ctx, Depth(ctx)+1), node)
}

// schemaKey uniquely identifies a SchemaProxy for cycle detection.
// It combines pointer address and reference string to handle both
// inline schemas and $ref schemas correctly.
type schemaKey struct {
	ptr uintptr // Pointer address of SchemaProxy
	ref string  // Reference string (empty for inline schemas)
}

// Stack returns the current schema recursion stack from context.
// This is a stack, not a global visited set - entries are removed after LeaveSchema.
// Returns nil if not set.
func Stack(ctx context.Context) map[schemaKey]bool {
	if v := ctx.Value(stackKey); v != nil {
		return v.(map[schemaKey]bool)
	}
	return nil
}

// WithStack returns a context with the given schema stack.
func WithStack(ctx context.Context, stack map[schemaKey]bool) context.Context {
	return context.WithValue(ctx, stackKey, stack)
}

// Parent returns the parent node from context.
// Returns nil if not set or at root.
func Parent(ctx context.Context) any {
	return ctx.Value(parentKey)
}

// WithParent returns a context with the given parent node.
func WithParent(ctx context.Context, parent any) context.Context {
	return context.WithValue(ctx, parentKey, parent)
}
