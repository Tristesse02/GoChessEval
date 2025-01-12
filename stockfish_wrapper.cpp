#include <windows.h>
#include <iostream>
#include <string>

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

	// Write commands to Stockfish
	std::string stockfish_command = "uci\nposition fen " + std::string(fen) + "\ngo depth 15\n";
	DWORD written;
	WriteFile(hChildStdinWr, stockfish_command.c_str(), stockfish_command.size(), &written, NULL);
	CloseHandle(hChildStdinWr); // Close the write end after sending commands

	// Read output from Stockfish
	char buffer[1024];
	DWORD bytesRead;
	while (ReadFile(hChildStdoutRd, buffer, sizeof(buffer) - 1, &bytesRead, NULL) && bytesRead > 0)
	{
		buffer[bytesRead] = '\0';
		output += buffer;

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

int main()
{
	const char *fen = "r2qkb1r/ppp2ppp/2n1pn2/3p1bB1/2PP4/2N1PN2/PP3PPP/R2QKB1R w KQkq - 0 1";
	const char *result = evaluate_position(fen);
	std::cout << "Stockfish result: " << result << std::endl;
	return 0;
}
