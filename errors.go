// Copyright 2026 Princess Beef Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package libasyncapi

import (
	"errors"
	"fmt"
)

// ErrAsyncAPI2NotSupported is returned when attempting to parse an AsyncAPI 2.x specification.
// This library only supports AsyncAPI 3.0 and later.
var ErrAsyncAPI2NotSupported = errors.New("AsyncAPI 2.x is not supported; use AsyncAPI 3.0 or later")

// ErrNoAsyncAPIVersion is returned when the asyncapi version field is missing from the specification.
var ErrNoAsyncAPIVersion = errors.New("asyncapi version field is required")

// ErrInvalidAsyncAPIVersion is returned when the asyncapi version field is not a valid semver string.
var ErrInvalidAsyncAPIVersion = errors.New("asyncapi version is not a valid semver string")

// ErrInvalidYAML is returned when the specification is not valid YAML or JSON.
var ErrInvalidYAML = errors.New("specification is not valid YAML or JSON")

// ParseError represents an error that occurred during parsing with location information.
type ParseError struct {
	Message string
	Line    int
	Column  int
	Path    string
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s at line %d, column %d (%s)", e.Message, e.Line, e.Column, e.Path)
	}
	return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Line, e.Column)
}

// NewParseError creates a new ParseError with the given message and location.
func NewParseError(message string, line, column int, path string) *ParseError {
	return &ParseError{
		Message: message,
		Line:    line,
		Column:  column,
		Path:    path,
	}
}

// ValidationError represents a semantic validation error in the specification.
type ValidationError struct {
	Message string
	Path    string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s at %s", e.Message, e.Path)
	}
	return e.Message
}

// NewValidationError creates a new ValidationError.
func NewValidationError(message, path string) *ValidationError {
	return &ValidationError{
		Message: message,
		Path:    path,
	}
}

// ReferenceError represents an error resolving a $ref reference.
type ReferenceError struct {
	Reference string
	Message   string
	Line      int
	Column    int
}

// Error implements the error interface.
func (e *ReferenceError) Error() string {
	return fmt.Sprintf("reference error for '%s': %s at line %d, column %d",
		e.Reference, e.Message, e.Line, e.Column)
}

// NewReferenceError creates a new ReferenceError.
func NewReferenceError(ref, message string, line, column int) *ReferenceError {
	return &ReferenceError{
		Reference: ref,
		Message:   message,
		Line:      line,
		Column:    column,
	}
}

// CircularReferenceError represents a circular reference detected in the specification.
type CircularReferenceError struct {
	Path  string
	Chain []string
}

// Error implements the error interface.
func (e *CircularReferenceError) Error() string {
	return fmt.Sprintf("circular reference detected at %s", e.Path)
}

// NewCircularReferenceError creates a new CircularReferenceError.
func NewCircularReferenceError(path string, chain []string) *CircularReferenceError {
	return &CircularReferenceError{
		Path:  path,
		Chain: chain,
	}
}
