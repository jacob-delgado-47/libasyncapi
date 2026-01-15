// Copyright 2026 Princess Beef Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

// Package visitor provides a visitor pattern for traversing AsyncAPI 3.0 documents.
// It operates on high-level models from doc.Model(), using pre-order traversal
// (visit node before children) with stack-based cycle detection for schemas.
//
// Node types passed to Visit include high-level asyncapi types, high-level base
// types from libopenapi (SchemaProxy, Schema, Contact, License), and *low.Reference
// for unresolved references (channel servers, operation channels/messages, etc.).
// Use a type switch in Visit to handle specific node types.
//
// To stop traversal early without error, return ErrStopTraversal from Visit.
package visitor

import (
	"context"
	"errors"

	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
)

// ErrStopTraversal can be returned from Visit to stop traversal cleanly.
// This allows visitors to find a specific node and stop without returning an error.
var ErrStopTraversal = errors.New("stop traversal")

// Visitor visits AsyncAPI document nodes during traversal.
// Visit is called in pre-order (before descending into children).
type Visitor interface {
	// Visit is called for each node in pre-order traversal.
	// Return an error to stop traversal.
	Visit(ctx context.Context, node any) error
}

// SchemaVisitor extends Visitor with schema lifecycle events.
// EnterSchema fires before resolution, LeaveSchema fires after (guaranteed via defer).
type SchemaVisitor interface {
	Visitor

	// EnterSchema is called before resolving and walking a SchemaProxy.
	// Return an error to skip this schema and its children.
	EnterSchema(ctx context.Context, proxy *highbase.SchemaProxy) error

	// LeaveSchema is called after walking a schema (guaranteed via defer).
	// The err parameter contains any error from walking children.
	// The schema parameter may be nil if resolution failed.
	LeaveSchema(ctx context.Context, proxy *highbase.SchemaProxy, schema *highbase.Schema, err error)

	// SkipCircularRef is called when a circular reference is detected.
	// The ref is the reference string that would cause infinite recursion.
	SkipCircularRef(ctx context.Context, proxy *highbase.SchemaProxy, ref string) error
}

// PolymorphicVisitor extends SchemaVisitor with allOf/oneOf/anyOf events.
// Leave callbacks are guaranteed to fire via defer, even on errors.
type PolymorphicVisitor interface {
	SchemaVisitor

	// EnterAllOf is called before walking allOf schemas.
	// count is the number of schemas in the allOf array.
	EnterAllOf(ctx context.Context, schema *highbase.Schema, count int) error
	// LeaveAllOf is called after walking allOf schemas (guaranteed via defer).
	LeaveAllOf(ctx context.Context, schema *highbase.Schema, err error)

	// EnterOneOf is called before walking oneOf schemas.
	EnterOneOf(ctx context.Context, schema *highbase.Schema, count int) error
	// LeaveOneOf is called after walking oneOf schemas (guaranteed via defer).
	LeaveOneOf(ctx context.Context, schema *highbase.Schema, err error)

	// EnterAnyOf is called before walking anyOf schemas.
	EnterAnyOf(ctx context.Context, schema *highbase.Schema, count int) error
	// LeaveAnyOf is called after walking anyOf schemas (guaranteed via defer).
	LeaveAnyOf(ctx context.Context, schema *highbase.Schema, err error)
}
