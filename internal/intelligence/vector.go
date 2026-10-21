package intelligence

import (
	"math"
	"strings"
)

type Vector []float64

func ComputeCosineSimilarity(a, b Vector) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func GenerateMockEmbedding(skillText string) Vector {
	lower := strings.ToLower(skillText)
	v := Vector{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8}

	if strings.Contains(lower, "go") || strings.Contains(lower, "golang") {
		v[0] = 0.95
		v[1] = 0.88
	}
	if strings.Contains(lower, "rust") || strings.Contains(lower, "vector") {
		v[2] = 0.96
		v[3] = 0.91
	}
	if strings.Contains(lower, "react") || strings.Contains(lower, "ui") || strings.Contains(lower, "gpu") {
		v[4] = 0.92
		v[5] = 0.89
	}

	var norm float64
	for _, val := range v {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	for i := range v {
		v[i] = v[i] / norm
	}

	return v
}
