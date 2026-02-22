## About
A simple backend build with Go and C++ to evaluate the chess position (with the aid of our beloved chess engine Stockfish)

Fork, Pipe, Execlp, ...

## Run
To run the program, navigate to the folder and start server using `go run .`\
To test result of many possibilities moves, for Window try using this template: `curl.exe -X POST http://localhost:8080/compare-moves -H "Content-Type: application/json" -d "{\"fen\": \"r1bqkb1r/pp3ppp/2nppn2/8/3NP3/2N1B3/PPP2PPP/R2QKB1R w KQkq - 0 1\", \"moves\": [\"d4f5\", \"d4b5\", \"e4e5\"]}"
`
