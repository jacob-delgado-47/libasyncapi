// Copyright 2026 Princess Beef Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package asyncapi

import (
	"context"
	"crypto/sha256"
	"errors"
	"sync"

	"github.com/pb33f/libopenapi/datamodel/low"
	"github.com/pb33f/libopenapi/index"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"go.yaml.in/yaml/v4"
)

// AsyncAPI represents a low-level AsyncAPI 3.0 Document object.
//
// This is the root document object of the AsyncAPI document.
//
//	https://www.asyncapi.com/docs/reference/specification/v3.0.0#A2SObject
type AsyncAPI struct {
	AsyncAPI           low.NodeReference[string]
	ID                 low.NodeReference[string]
	Info               low.NodeReference[*Info]
	Servers            low.NodeReference[*orderedmap.Map[low.KeyReference[string], low.ValueReference[*Server]]]
	DefaultContentType low.NodeReference[string]
	Channels           low.NodeReference[*orderedmap.Map[low.KeyReference[string], low.ValueReference[*Channel]]]
	Operations         low.NodeReference[*orderedmap.Map[low.KeyReference[string], low.ValueReference[*Operation]]]
	Components         low.NodeReference[*Components]
	Extensions         *orderedmap.Map[low.KeyReference[string], low.ValueReference[*yaml.Node]]
	KeyNode            *yaml.Node
	RootNode           *yaml.Node
	Index              *index.SpecIndex
	Rolodex            *index.Rolodex
	ctx                context.Context
	*low.Reference
	low.NodeMap
}

// GetRootNode returns the root yaml node of the AsyncAPI document.
func (a *AsyncAPI) GetRootNode() *yaml.Node {
	return a.RootNode
}

// GetKeyNode returns the key yaml node of the AsyncAPI document.
func (a *AsyncAPI) GetKeyNode() *yaml.Node {
	return a.KeyNode
}

// GetExtensions returns all extensions for AsyncAPI.
func (a *AsyncAPI) GetExtensions() *orderedmap.Map[low.KeyReference[string], low.ValueReference[*yaml.Node]] {
	return a.Extensions
}

// FindExtension attempts to locate an extension with the supplied key.
func (a *AsyncAPI) FindExtension(ext string) *low.ValueReference[*yaml.Node] {
	return low.FindItemInOrderedMap(ext, a.Extensions)
}

// GetIndex returns the index.SpecIndex instance attached to the AsyncAPI document.
func (a *AsyncAPI) GetIndex() *index.SpecIndex {
	return a.Index
}

// GetContext returns the context.Context instance used when building the AsyncAPI document.
func (a *AsyncAPI) GetContext() context.Context {
	return a.ctx
}

// Build extracts the AsyncAPI document from the supplied root node.
func (a *AsyncAPI) Build(ctx context.Context, keyNode, root *yaml.Node, idx *index.SpecIndex) error {
	a.KeyNode = keyNode
	root = utils.NodeAlias(root)
	a.RootNode = root
	utils.CheckForMergeNodes(root)
	a.Reference = new(low.Reference)
	a.Nodes = low.ExtractNodes(ctx, root)
	a.Extensions = low.ExtractExtensions(root)
	// Store extension nodes for line/column info preservation
	low.ExtractExtensionNodes(ctx, a.Extensions, a.Nodes)
	a.Index = idx
	a.ctx = ctx

	var wg sync.WaitGroup
	var errMu sync.Mutex
	var buildErrs []error

	captureErr := func(err error) {
		if err != nil {
			errMu.Lock()
			buildErrs = append(buildErrs, err)
			errMu.Unlock()
		}
	}

	// extract info
	wg.Add(1)
	go func() {
		defer wg.Done()
		info, err := low.ExtractObject[*Info](ctx, InfoLabel, root, idx)
		captureErr(err)
		a.Info = info
	}()

	// extract servers
	wg.Add(1)
	go func() {
		defer wg.Done()
		servers, sLabel, sValue, err := low.ExtractMap[*Server](ctx, ServersLabel, root, idx)
		captureErr(err)
		if servers != nil {
			a.Servers = low.NodeReference[*orderedmap.Map[low.KeyReference[string], low.ValueReference[*Server]]]{
				Value:     servers,
				KeyNode:   sLabel,
				ValueNode: sValue,
			}
		}
	}()

	// extract channels
	wg.Add(1)
	go func() {
		defer wg.Done()
		channels, cLabel, cValue, err := low.ExtractMap[*Channel](ctx, ChannelsLabel, root, idx)
		captureErr(err)
		if channels != nil {
			a.Channels = low.NodeReference[*orderedmap.Map[low.KeyReference[string], low.ValueReference[*Channel]]]{
				Value:     channels,
				KeyNode:   cLabel,
				ValueNode: cValue,
			}
		}
	}()

	// extract operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		ops, oLabel, oValue, err := low.ExtractMap[*Operation](ctx, OperationsLabel, root, idx)
		captureErr(err)
		if ops != nil {
			a.Operations = low.NodeReference[*orderedmap.Map[low.KeyReference[string], low.ValueReference[*Operation]]]{
				Value:     ops,
				KeyNode:   oLabel,
				ValueNode: oValue,
			}
		}
	}()

	// extract components
	wg.Add(1)
	go func() {
		defer wg.Done()
		comp, err := low.ExtractObject[*Components](ctx, ComponentsLabel, root, idx)
		captureErr(err)
		a.Components = comp
	}()

	wg.Wait()
	return errors.Join(buildErrs...)
}

// Hash returns a consistent SHA256 Hash of the AsyncAPI document.
func (a *AsyncAPI) Hash() [32]byte {
	sb := low.GetStringBuilder()
	defer low.PutStringBuilder(sb)

	if !a.AsyncAPI.IsEmpty() {
		sb.WriteString(a.AsyncAPI.Value)
		sb.WriteByte('|')
	}
	if !a.ID.IsEmpty() {
		sb.WriteString(a.ID.Value)
		sb.WriteByte('|')
	}
	if !a.Info.IsEmpty() {
		sb.WriteString(low.GenerateHashString(a.Info.Value))
		sb.WriteByte('|')
	}
	if !a.DefaultContentType.IsEmpty() {
		sb.WriteString(a.DefaultContentType.Value)
		sb.WriteByte('|')
	}
	if a.Servers.Value != nil {
		for v := range orderedmap.SortAlpha(a.Servers.Value).ValuesFromOldest() {
			sb.WriteString(low.GenerateHashString(v.Value))
			sb.WriteByte('|')
		}
	}
	if a.Channels.Value != nil {
		for v := range orderedmap.SortAlpha(a.Channels.Value).ValuesFromOldest() {
			sb.WriteString(low.GenerateHashString(v.Value))
			sb.WriteByte('|')
		}
	}
	if a.Operations.Value != nil {
		for v := range orderedmap.SortAlpha(a.Operations.Value).ValuesFromOldest() {
			sb.WriteString(low.GenerateHashString(v.Value))
			sb.WriteByte('|')
		}
	}
	if !a.Components.IsEmpty() {
		sb.WriteString(low.GenerateHashString(a.Components.Value))
		sb.WriteByte('|')
	}
	for _, ext := range low.HashExtensions(a.Extensions) {
		sb.WriteString(ext)
		sb.WriteByte('|')
	}
	return sha256.Sum256([]byte(sb.String()))
}
