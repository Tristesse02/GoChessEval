#include "stockfish_wrapper.h"
#include <iostream>
#include <string>
// C++ version of stdio.h: standard input/output library
#include <cstdio>
// C++ version of stdlib.h: general utilities library
// It provides general-purpose utility functions, such as
// memory management, random no generation, process control.
// Often included when working with system-level functions
// like popen() or system()
#include <cstdlib>

extern "C" const char *evaluate_position(const char *fen)
{
	static std::string output;					   // ensure output string persists across function calls
	FILE *stockfish = popen("stockfish.exe", "w"); // Launch Stockfish; open pipe to read and write to process

	if (!stockfish)
	{
		perror("popen failed"); // Print detailed error
		return "error";
	}
	if (stockfish)
	{
		fprintf(stockfish, "uci\n");				  // Initialize the Universal Chess Interface (UCI) mode
		fprintf(stockfish, "position fen %s\n", fen); // Tells Stockfish to evaulate the position represented by the FEN string
		fprintf(stockfish, "go depth 15\n");		  // Instructs Stockfish to search the game tree to a depth of 15 moves

		char buffer[1024];								 // Buffer to store the output from Stockfish
		while (fgets(buffer, sizeof(buffer), stockfish)) // Read the output by Stockfish line by line using fgets
		{
			std::cout << "vl bn oi";
			std::string line(buffer);
			// std::cout << line;
			if (line.find("bestmove") != std::string::npos)
			{
				output = line;
				break;
			}
		}
		pclose(stockfish); // close the pipe
	}
	return output.c_str(); // return the result
}

int main()
{
	const char *fen = "rnbqkb1r/pp2pppp/2p2n2/3p4/2PP4/2N2N2/PP2PPPP/R1BQKB1R w KQkq - 0 1";
	const char *result = evaluate_position(fen);
	printf("%p\n", result);
	std::cout << "Stockfish result: " << result << std::endl;
	return 0;
}