package cli

import (
	"io"
	"os"
)

func isTerminal(stream io.Writer) bool {
	file, ok := stream.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func isInteractiveTerminal(input io.Reader, output io.Writer) bool {
	inputFile, inputOK := input.(*os.File)
	if !inputOK || !isTerminal(output) {
		return false
	}
	info, err := inputFile.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
