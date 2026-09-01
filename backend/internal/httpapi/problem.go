package httpapi

import (
	"encoding/json"
	"net/http"
)

// ProblemJSONMediaType is the established Daily News error media type.
const ProblemJSONMediaType = "application/problem+json"

// Problem is the Flutter-compatible Problem Details payload.
type Problem struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
}

// ValidationProblem converts strict request decoding failure into the API contract.
func ValidationProblem(_ error) Problem {
	return Problem{Title: "Validation failed", Status: http.StatusUnprocessableEntity}
}

// WriteProblem writes the FastAPI-compatible Problem Details wrapper.
func WriteProblem(writer http.ResponseWriter, problem Problem) {
	writer.Header().Set("Content-Type", ProblemJSONMediaType)
	writer.WriteHeader(problem.Status)
	_ = json.NewEncoder(writer).Encode(struct {
		Detail Problem `json:"detail"`
	}{Detail: problem})
}
