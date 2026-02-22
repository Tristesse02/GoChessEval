package main

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/notnil/chess"
)

// --- Request / Response types ---

type CompareMoveRequest struct {
	FEN   string   `json:"fen"`
	Moves []string `json:"moves"`
}

type MoveResult struct {
	Move         string `json:"move"`
	Evaluation   string `json:"evaluation"`
	BestResponse string `json:"bestResponse"`
}

type CompareMoveResponse struct {
	Results []MoveResult `json:"results"`
}

// --- Thread safety ---

// stockfishMu serializes calls to callStockfish.
// The C++ wrapper uses a static string buffer internally,
// which means it is NOT safe to call from multiple goroutines at the same time.
// The mutex ensures only one goroutine talks to Stockfish at a time.
var stockfishMu sync.Mutex

// --- Helper: apply a UCI move to a FEN string ---

// applyMove takes a FEN position and a UCI move (e.g. "e2e4") and returns
// the FEN of the resulting position after the move is played.
func applyMove(fen, uciMove string) (string, error) {
	fenOpt, err := chess.FEN(fen)
	if err != nil {
		return "", err
	}
	game := chess.NewGame(fenOpt)

	move, err := chess.UCINotation{}.Decode(game.Position(), uciMove)
	if err != nil {
		return "", err
	}

	if err := game.Move(move); err != nil {
		return "", err
	}

	return game.Position().String(), nil
}

// --- Handler ---

// compareMovesHandler evaluates multiple candidate moves from a given position
// simultaneously using goroutines, then returns all evaluations ranked in the
// same order as the input moves.
func compareMovesHandler(c *gin.Context) {
	var req CompareMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if len(req.Moves) == 0 {
		c.JSON(400, gin.H{"error": "No moves provided"})
		return
	}

	// Pre-allocate a results slice the same length as the input moves.
	// Each goroutine writes to its own index — no mutex needed for this slice.
	results := make([]MoveResult, len(req.Moves))

	// WaitGroup lets the handler wait until every goroutine has finished.
	var wg sync.WaitGroup

	for i, move := range req.Moves {
		wg.Add(1)

		// Launch one goroutine per candidate move.
		// We pass idx and uciMove as arguments to avoid closure capture issues.
		go func(idx int, uciMove string) {
			defer wg.Done() // signal this goroutine is done when it returns

			// Step 1: apply the candidate move to get the opponent's position
			newFEN, err := applyMove(req.FEN, uciMove)
			if err != nil {
				results[idx] = MoveResult{
					Move:         uciMove,
					Evaluation:   "invalid move",
					BestResponse: "unknown",
				}
				return
			}

			// Step 2: ask Stockfish to evaluate the resulting position.
			// Lock the mutex so only one goroutine calls the C++ wrapper at a time.
			stockfishMu.Lock()
			eval, bestResponse := callStockfish(newFEN)
			stockfishMu.Unlock()

			results[idx] = MoveResult{
				Move:         uciMove,
				Evaluation:   eval,
				BestResponse: bestResponse,
			}
		}(i, move)
	}

	// Block until all goroutines have written their results
	wg.Wait()

	c.JSON(200, CompareMoveResponse{Results: results})
}
