// Package search provides a local, offline semantic search pipeline:
// MD/HTML/JSON → chunk → embed → store (vector + text) → CLI search.
//
// Every component is an interface so the pipeline can be swapped (e.g. a
// different embedder or vector store) without touching the indexer.
package search

import (
	"context"
)

// Document is the parsed output of one source file: raw content plus
// provenance metadata used for content-addressed storage and deduplication.
type Document struct {
	ID      string
	Path    string
	Content string
	Meta    DocumentMeta
}

// DocumentMeta carries provenance for a parsed document.
type DocumentMeta struct {
	Source       string // e.g. "md", "html", "json"
	ModifiedTime int64  // unix nano
	Hash         string // SHA256 of raw content
}

// Chunk is one slice of a document, produced by a Chunker.
type Chunk struct {
	ID         string
	DocID      string
	ParentID   string // document id this chunk came from
	Content    string
	TokenCount int
	Section    string // heading path, e.g. "## Overview > ### Usage"
	Meta       map[string]string
}

// Embedding is a vector plus the chunk id it represents.
type Embedding struct {
	ChunkID string
	Value   []float32
}

// Result is one search hit returned by a VectorStore or TextIndex.
type Result struct {
	ChunkID  string
	DocID    string
	Content  string
	Score    float32
	Metadata map[string]string
}

// Parser turns a file on disk into Documents.
type Parser interface {
	Parse(ctx context.Context, path string) ([]Document, error)
	Supported() []string // extensions this parser handles, e.g. ".md"
}

// Chunker splits a Document into Chunk slices.
type Chunker interface {
	Chunk(doc Document) []Chunk
}

// Embedder turns text into vectors.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dimensions() int
}

// VectorStore is a vector similarity index.
type VectorStore interface {
	Add(ctx context.Context, chunks []Chunk, vectors [][]float32) error
	Search(ctx context.Context, query []float32, topK int) ([]Result, error)
	Delete(ctx context.Context, chunkIDs []string) error
	Count() int
}

// TextIndex is a lexical (BM25) index.
type TextIndex interface {
	Add(ctx context.Context, chunks []Chunk) error
	Search(ctx context.Context, query string, topK int) ([]Result, error)
	Delete(ctx context.Context, chunkIDs []string) error
	Count() int
}
