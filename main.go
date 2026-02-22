package main

/*
#cgo CXXFLAGS: -std=c++11 -I.
#cgo LDFLAGS: -L. -lstockfish_wrapper -lstdc++
#include "stockfish_wrapper.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	"github.com/gin-gonic/gin"
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

	// Normalize Windows line endings
	result = strings.ReplaceAll(result, "\r\n", "\n")

	// Extract "bestmove" and "evaluation" using simple parsing
	// We keep updating evaluation so the last info line (deepest depth) wins
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				bestMove = parts[1]
			}
			break
		} else if strings.HasPrefix(line, "info") && strings.Contains(line, "score") {
			parts := strings.Fields(line)
			for i := 0; i < len(parts)-2; i++ {
				if parts[i] == "score" {
					evaluation = parts[i+1] + " " + parts[i+2]
					break
				}
			}
		}
	}

	return evaluation, bestMove
}

func evaluateHandler(c *gin.Context) {
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	evaluation, bestMove := callStockfish(req.FEN)

	c.JSON(200, EvaluateResponse{
		Evaluation: evaluation,
		BestMove:   bestMove,
	})
}

func main() {
	r := gin.Default()
	r.POST("/evaluate", evaluateHandler)
	log.Println("Server running on port 8080")
	r.Run(":8080")
}
