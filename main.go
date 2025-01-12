/*
#cgo CXXFLAGS: -std=c++11
#cgo LDFLAGS: -lstdc++
#include <stdlib.h>
#include "stockfish_wrapper.h"
*/
package main

import (
	"C"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	result := C.GoString(C.evaluate_position(cFEN))

	var bestMove, evaluation string
	n, err := fmt.Sscanf(result, "bestmove %s eval %s", &bestMove, &evaluation)
	if err != nil || n != 2 {
		bestMove, evaluation = "unknown", "unknown"
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
