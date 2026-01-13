// Copyright 2026 Princess Beef Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package asyncapi

import (
	"context"
	"crypto/sha256"

	"github.com/pb33f/libopenapi/datamodel/low"
	"github.com/pb33f/libopenapi/datamodel/low/base"
	"github.com/pb33f/libopenapi/index"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"go.yaml.in/yaml/v4"
)

// Info represents a low-level AsyncAPI 3.0 Info object.
//
// The Info object provides metadata about the API. The metadata can be used by clients if needed,
// and can be presented in editing or documentation generation tools.
//
// Unlike OpenAPI Info, AsyncAPI Info includes tags and externalDocs fields, and these can
// contain references ($ref).
//
//	https://www.asyncapi.com/docs/reference/specification/v3.0.0#infoObject
type Info struct {
	Title          low.NodeReference[string]
	Version        low.NodeReference[string]
	Description    low.NodeReference[string]
	TermsOfService low.NodeReference[string]
	Contact        low.NodeReference[*base.Contact]
	License        low.NodeReference[*base.License]
	Tags           low.NodeReference[*orderedmap.Map[low.KeyReference[string], low.ValueReference[*Tag]]]
	ExternalDocs   low.NodeReference[*ExternalDoc]
	Extensions     *orderedmap.Map[low.KeyReference[string], low.ValueReference[*yaml.Node]]
	KeyNode        *yaml.Node
	RootNode       *yaml.Node
	idx            *index.SpecIndex
	ctx            context.Context
	*low.Reference
	low.NodeMap
}

// GetRootNode returns the root yaml node of the Info object.
func (i *Info) GetRootNode() *yaml.Node {
	return i.RootNode
}

// GetKeyNode returns the key yaml node of the Info object.
func (i *Info) GetKeyNode() *yaml.Node {
	return i.KeyNode
}

// GetExtensions returns all extensions for Info.
func (i *Info) GetExtensions() *orderedmap.Map[low.KeyReference[string], low.ValueReference[*yaml.Node]] {
	return i.Extensions
}

// FindExtension attempts to locate an extension with the supplied key.
func (i *Info) FindExtension(ext string) *low.ValueReference[*yaml.Node] {
	return low.FindItemInOrderedMap(ext, i.Extensions)
}

// GetIndex returns the index.SpecIndex instance attached to the Info object.
func (i *Info) GetIndex() *index.SpecIndex {
	return i.idx
}

// GetContext returns the context.Context instance used when building the Info object.
func (i *Info) GetContext() context.Context {
	return i.ctx
}

// Build extracts the Info object from the supplied root node.
func (i *Info) Build(ctx context.Context, keyNode, root *yaml.Node, idx *index.SpecIndex) error {
	i.KeyNode = keyNode
	root = utils.NodeAlias(root)
	i.RootNode = root
	utils.CheckForMergeNodes(root)
	i.Reference = new(low.Reference)
	i.Nodes = low.ExtractNodes(ctx, root)
	i.Extensions = low.ExtractExtensions(root)
	i.idx = idx
	i.ctx = ctx

	contact, _ := low.ExtractObject[*base.Contact](ctx, ContactLabel, root, idx)
	i.Contact = contact

	lic, _ := low.ExtractObject[*base.License](ctx, LicenseLabel, root, idx)
	i.License = lic

	extDocs, _ := low.ExtractObject[*ExternalDoc](ctx, ExternalDocsLabel, root, idx)
	i.ExternalDocs = extDocs

	tags, tLabel, tValue, err := low.ExtractArray[*Tag](ctx, TagsLabel, root, idx)
	if err != nil {
		return err
	}
	if tags != nil {
		tagMap := orderedmap.New[low.KeyReference[string], low.ValueReference[*Tag]]()
		for _, tag := range tags {
			key := low.KeyReference[string]{
				Value:   tag.Value.Name.Value,
				KeyNode: tag.ValueNode,
			}
			tagMap.Set(key, tag)
		}
		i.Tags = low.NodeReference[*orderedmap.Map[low.KeyReference[string], low.ValueReference[*Tag]]]{
			Value:     tagMap,
			KeyNode:   tLabel,
			ValueNode: tValue,
		}
	}

	return nil
}

// Hash returns a consistent SHA256 Hash of the Info object.
func (i *Info) Hash() [32]byte {
	sb := low.GetStringBuilder()
	defer low.PutStringBuilder(sb)

	if !i.Title.IsEmpty() {
		sb.WriteString(i.Title.Value)
		sb.WriteByte('|')
	}
	if !i.Version.IsEmpty() {
		sb.WriteString(i.Version.Value)
		sb.WriteByte('|')
	}
	if !i.Description.IsEmpty() {
		sb.WriteString(i.Description.Value)
		sb.WriteByte('|')
	}
	if !i.TermsOfService.IsEmpty() {
		sb.WriteString(i.TermsOfService.Value)
		sb.WriteByte('|')
	}
	if !i.Contact.IsEmpty() {
		sb.WriteString(low.GenerateHashString(i.Contact.Value))
		sb.WriteByte('|')
	}
	if !i.License.IsEmpty() {
		sb.WriteString(low.GenerateHashString(i.License.Value))
		sb.WriteByte('|')
	}
	if !i.ExternalDocs.IsEmpty() {
		sb.WriteString(low.GenerateHashString(i.ExternalDocs.Value))
		sb.WriteByte('|')
	}
	if i.Tags.Value != nil {
		for v := range orderedmap.SortAlpha(i.Tags.Value).ValuesFromOldest() {
			sb.WriteString(low.GenerateHashString(v.Value))
			sb.WriteByte('|')
		}
	}
	for _, ext := range low.HashExtensions(i.Extensions) {
		sb.WriteString(ext)
		sb.WriteByte('|')
	}
	return sha256.Sum256([]byte(sb.String()))
}
