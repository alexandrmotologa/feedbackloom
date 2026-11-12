package cluster

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"strings"
	"unicode"
)

// Vectorizer defines the interface for converting text into dense float32 embeddings.
type Vectorizer interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	Dimension() int
}

// LocalSemanticVectorizer is a deterministic, offline subword n-gram vectorizer
// with term frequency weighting and L2 normalization, producing 128-dimensional dense vectors.
// It requires zero external network calls or API keys, ensuring 100% offline self-hosted operation.
type LocalSemanticVectorizer struct {
	dim int
}

// NewLocalSemanticVectorizer creates a vectorizer with the given dimension (default 128).
func NewLocalSemanticVectorizer(dim int) *LocalSemanticVectorizer {
	if dim <= 0 {
		dim = 128
	}
	return &LocalSemanticVectorizer{dim: dim}
}

func (v *LocalSemanticVectorizer) Dimension() int {
	return v.dim
}

// Embed generates a dense L2-normalized vector for the input text.
func (v *LocalSemanticVectorizer) Embed(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, v.dim)
	clean := strings.ToLower(text)

	// Tokenize into words
	words := strings.FieldsFunc(clean, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	if len(words) == 0 {
		return vec, nil
	}

	stopWords := map[string]bool{
		"the": true, "and": true, "is": true, "in": true, "it": true,
		"of": true, "to": true, "a": true, "our": true, "we": true,
		"for": true, "on": true, "with": true, "as": true, "this": true,
		"that": true, "at": true, "by": true, "from": true, "be": true,
	}

	// 1. Word-level hashing & weighting
	for _, w := range words {
		if stopWords[w] || len(w) < 2 {
			continue
		}

		// Domain feature keywords receive elevated semantic weight
		weight := float32(1.0)
		switch {
		case strings.Contains(w, "sso"), strings.Contains(w, "saml"), strings.Contains(w, "okta"), strings.Contains(w, "azure"):
			weight = 3.5
		case strings.Contains(w, "audit"), strings.Contains(w, "siem"), strings.Contains(w, "soc2"), strings.Contains(w, "log"):
			weight = 3.5
		case strings.Contains(w, "rbac"), strings.Contains(w, "role"), strings.Contains(w, "permission"), strings.Contains(w, "scope"):
			weight = 3.5
		case strings.Contains(w, "export"), strings.Contains(w, "bulk"), strings.Contains(w, "csv"), strings.Contains(w, "api"):
			weight = 3.0
		}

		idx := hashToBucket(w, v.dim)
		sign := hashSign(w)
		vec[idx] += weight * sign

		// 2. Character 3-grams for morphology and typo tolerance
		runes := []rune(w)
		if len(runes) >= 3 {
			for i := 0; i <= len(runes)-3; i++ {
				ngram := string(runes[i : i+3])
				nIdx := hashToBucket(ngram, v.dim)
				nSign := hashSign(ngram)
				vec[nIdx] += 0.4 * weight * nSign
			}
		}
	}

	// L2 normalization: ||v||_2 = 1.0
	var sumSquares float64
	for _, val := range vec {
		sumSquares += float64(val * val)
	}
	norm := math.Sqrt(sumSquares)
	if norm > 1e-9 {
		invNorm := float32(1.0 / norm)
		for i := range vec {
			vec[i] *= invNorm
		}
	}

	return vec, nil
}

func hashToBucket(s string, dim int) int {
	h := sha256.Sum256([]byte(s))
	val := binary.LittleEndian.Uint32(h[:4])
	return int(val % uint32(dim))
}

func hashSign(s string) float32 {
	h := sha256.Sum256([]byte(s + "_sign"))
	if h[0]%2 == 0 {
		return 1.0
	}
	return -1.0
}

// CosineDistance calculates 1 - CosineSimilarity between two normalized float vectors.
func CosineDistance(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 1.0
	}
	var dot float64
	for i := range a {
		dot += float64(a[i] * b[i])
	}
	dist := 1.0 - dot
	if dist < 0.0 {
		return 0.0
	}
	if dist > 2.0 {
		return 2.0
	}
	return dist
}

// EuclideanDistance computes L2 distance between two vectors.
func EuclideanDistance(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 1.0
	}
	var sum float64
	for i := range a {
		diff := float64(a[i] - b[i])
		sum += diff * diff
	}
	return math.Sqrt(sum)
}
