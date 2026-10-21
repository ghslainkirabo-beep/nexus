package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	"github.com/nexus/core/internal/intelligence"
)

type MatchRequest struct {
	TalentSkills   string `json:"talent_skills"`
	OpportunityReq string `json:"opportunity_requirements"`
}

type MatchResponse struct {
	Status            string  `json:"status"`
	CosineSimilarity  float64 `json:"cosine_similarity"`
	MatchPercentage   float64 `json:"match_percentage"`
	ComputationTimeNs int64   `json:"computation_time_ns"`
	RecommendedAction string  `json:"recommended_action"`
}

func HandleSemanticMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startTime := time.Now()

	var req MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid match payload", http.StatusBadRequest)
		return
	}

	talentVec := intelligence.GenerateMockEmbedding(req.TalentSkills)
	oppVec := intelligence.GenerateMockEmbedding(req.OpportunityReq)

	similarity := intelligence.ComputeCosineSimilarity(talentVec, oppVec)
	matchPercentage := math.Round(similarity * 10000) / 100
	if matchPercentage < 0 {
		matchPercentage = 0
	}

	action := "STANDARD_POOL_ROUTING"
	if matchPercentage > 85.0 {
		action = "INSTANT_EXECUTIVE_DEPLOYMENT"
	} else if matchPercentage > 60.0 {
		action = "PRIORITY_VECTOR_STREAM"
	}

	response := MatchResponse{
		Status:            "OPTIMAL",
		CosineSimilarity:  similarity,
		MatchPercentage:   matchPercentage,
		ComputationTimeNs: time.Since(startTime).Nanoseconds(),
		RecommendedAction: action,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}