# GoChessEval

A chess position evaluation backend built with **Go** and **C++**, powered by the [Stockfish](https://stockfishchess.org/) engine. Designed for high-level chess analysis — evaluate any position or compare candidate moves simultaneously using goroutines.

---

## Tech Stack

- **Go** — HTTP server and concurrency logic
- **Gin** — HTTP web framework
- **CGo** — bridges Go with the C++ Stockfish wrapper
- **C++** — native wrapper that communicates with Stockfish via pipes
- **Stockfish 17** — the underlying chess engine

---

## Endpoints

### `POST /evaluate`

Evaluates a single chess position.

**Request**
```json
{
  "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
}
```

**Response**
```json
{
  "evaluation": "cp 28",
  "bestMove": "e2e4"
}
```

- `evaluation` — centipawn score (e.g. `cp 28`) or forced mate (e.g. `mate 3`)
- `bestMove` — best move in UCI notation (e.g. `e2e4`)

---

### `POST /compare-moves`

Evaluates multiple candidate moves from a given position **simultaneously** using goroutines. Useful for opening preparation and critical position analysis — compare several ideas at once and see which holds up best against Stockfish's response.

**Request**
```json
{
  "fen": "r1bqkb1r/pp3ppp/2nppn2/8/3NP3/2N1B3/PPP2PPP/R2QKB1R w KQkq - 0 1",
  "moves": ["d4f5", "d4b5", "e4e5"]
}
```

**Response**
```json
{
  "results": [
    { "move": "d4f5", "evaluation": "cp 62", "bestResponse": "d6d5" },
    { "move": "d4b5", "evaluation": "cp 45", "bestResponse": "a7a6" },
    { "move": "e4e5", "evaluation": "cp 31", "bestResponse": "f6d5" }
  ]
}
```

- `moves` — candidate moves in UCI notation (e.g. `e2e4`)
- `bestResponse` — Stockfish's best reply to each candidate move

---

## Project Structure

```
chess-backend/
├── main.go                  # Server entry point, route registration
├── analyze.go               # /compare-moves handler and goroutine logic
├── stockfish_wrapper.cpp    # C++ wrapper that communicates with Stockfish via pipes
├── stockfish_wrapper.h      # Header for the C++ wrapper
├── stockfish_wrapper.dll    # Compiled C++ wrapper (Windows)
└── stockfish.exe            # Stockfish 17 binary
```

---

## Prerequisites

- [Go 1.23+](https://golang.org/dl/)
- [MinGW-W64](https://www.mingw-w64.org/) (GCC/G++ for Windows)
- Stockfish 17 binary placed as `stockfish.exe` in the project root

---

## Getting Started

**1. Clone the repository**
```bash
git clone https://github.com/Tristesse02/GoChessEval
cd GoChessEval
```

**2. Install dependencies**
```bash
go mod tidy
```

**3. Compile the C++ wrapper**
```bash
g++ -shared -o stockfish_wrapper.dll stockfish_wrapper.cpp -std=c++11 -lws2_32
```

**4. Run the server**
```bash
go run .
```

The server starts on `http://localhost:8080`.

---

## How Concurrency Works

The `/compare-moves` endpoint launches one goroutine per candidate move. Each goroutine:
1. Applies the candidate move to the FEN to produce the resulting position
2. Calls Stockfish to evaluate that position

Since the C++ wrapper uses a shared static buffer, a `sync.Mutex` serializes the Stockfish calls while goroutines handle everything else concurrently. A `sync.WaitGroup` ensures the handler waits for all evaluations before returning the response.