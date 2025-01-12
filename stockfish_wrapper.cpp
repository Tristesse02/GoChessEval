#include "stockfish_wrapper.h"
#include <windows.h>
#include <iostream>
#include <string>

extern "C"
{
	const char *evaluate_position(const char *fen)
	{
		static std::string output;
		output.clear(); // Clear the output string

		// Create pipes for input and output
		HANDLE hChildStdoutRd, hChildStdoutWr, hChildStdinRd, hChildStdinWr;
		SECURITY_ATTRIBUTES saAttr = {sizeof(SECURITY_ATTRIBUTES), NULL, TRUE};

		// Create pipes for STDOUT and STDIN
		if (!CreatePipe(&hChildStdoutRd, &hChildStdoutWr, &saAttr, 0))
		{
			perror("Stdout pipe creation failed\n");
			return "error";
		}
		if (!CreatePipe(&hChildStdinRd, &hChildStdinWr, &saAttr, 0))
		{
			perror("Stdin pipe creation failed\n");
			return "error";
		}

		// Ensure the write end of the pipes is not inherited
		SetHandleInformation(hChildStdoutRd, HANDLE_FLAG_INHERIT, 0);
		SetHandleInformation(hChildStdinWr, HANDLE_FLAG_INHERIT, 0);

		// Launch Stockfish as a child process
		PROCESS_INFORMATION piProcInfo;
		STARTUPINFOA siStartInfo = {sizeof(STARTUPINFOA)};
		siStartInfo.hStdError = hChildStdoutWr;
		siStartInfo.hStdOutput = hChildStdoutWr;
		siStartInfo.hStdInput = hChildStdinRd;
		siStartInfo.dwFlags |= STARTF_USESTDHANDLES;

		const char *command = "stockfish.exe"; // Ensure this is ANSI-compatible
		if (!CreateProcessA(NULL, (LPSTR)command, NULL, NULL, TRUE, 0, NULL, NULL, &siStartInfo, &piProcInfo))
		{
			perror("CreateProcess failed\n");
			return "error";
		}

		CloseHandle(hChildStdoutWr);
		CloseHandle(hChildStdinRd);

		// Write commands to Stockfish line by line
		std::string commands[] = {
			"uci\n",								   // Initialize Stockfish
			"isready\n",							   // Wait for "readyok"
			"ucinewgame\n",							   // Start a new game
			"position fen " + std::string(fen) + "\n", // Set the position using FEN
			"go depth 15\n"							   // Start calculating best move
		};

		for (const auto &cmd : commands)
		{
			DWORD written;
			if (!WriteFile(hChildStdinWr, cmd.c_str(), cmd.size(), &written, NULL))
			{
				perror("WriteFile failed\n");
				return "error";
			}
			std::cout << "Sending command: " << cmd << std::endl;

			// Add delays if necessary to ensure Stockfish processes commands
			Sleep(100);
		}
		CloseHandle(hChildStdinWr); // Close the input pipe after sending all commands

		// Read output from Stockfish
		char buffer[1024];
		DWORD bytesRead;
		while (ReadFile(hChildStdoutRd, buffer, sizeof(buffer) - 1, &bytesRead, NULL) && bytesRead > 0)
		{
			buffer[bytesRead] = '\0';
			output += buffer;
			std::cout << buffer;
			if (output.find("bestmove") != std::string::npos)
				break;
		}

		CloseHandle(hChildStdoutRd);
		CloseHandle(piProcInfo.hProcess);
		CloseHandle(piProcInfo.hThread);

		size_t pos = output.find("bestmove");
		if (pos != std::string::npos)
		{
			output = output.substr(pos); // Extract the line with "bestmove"
		}
		else
		{
			output = "no best move found";
		}

		return output.c_str();
	}
}

int main()
{
	const char *fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";
	const char *result = evaluate_position(fen);
	std::cout << "Minhdz result: " << result << std::endl;
	return 0;
}