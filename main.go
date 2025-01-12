package main

/*
#cgo CXXFLAGS: -std=c++11 -I.
#cgo LDFLAGS: -L. -lstockfish_wrapper -lstdc++
#include "stockfish_wrapper.h"
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unsafe"
)

type EvaluateRequest struct {
	FEN string `json:"fen"`
}

type EvaluateResponse struct {
	Evaluation string `json:"evaluation"`
	BestMove   string `json:"bestMove"`
}

func callStockfish(fen string) (string, string) {
	cFEN := C.CString(fen)
	defer C.free(unsafe.Pointer(cFEN))

	// Call the C++ function
	result := C.GoString(C.evaluate_position(cFEN))
	fmt.Println("Raw result from Stockfish:", result)

	// Initialize default values
	bestMove := "unknown"
	evaluation := "unknown"

	// Extract "bestmove" and "evaluation" using simple parsing
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				bestMove = parts[1] // The move after "bestmove"
			}
			break
		} else if strings.HasPrefix(line, "info") && strings.Contains(line, "score") {
			parts := strings.Fields(line)
			for i := 0; i < len(parts); i++ {
				if parts[i] == "score" && i+1 < len(parts) {
					evaluation = parts[i+1] + " " + parts[i+2] // e.g., "cp 28" or "mate 3"
					break
				}
			}
		}
	}

	return evaluation, bestMove
}

func evaluateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	evaluation, bestMove := callStockfish(req.FEN)

	response := EvaluateResponse{
		Evaluation: evaluation,
		BestMove:   bestMove,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/evaluate", evaluateHandler)
	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
