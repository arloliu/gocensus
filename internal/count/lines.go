package count

import (
	"bytes"
	"go/scanner"
	"go/token"
)

func rawLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := bytes.Count(content, []byte{'\n'})
	if content[len(content)-1] != '\n' {
		lines++
	}
	return lines
}

func effectiveLines(content []byte) int {
	fileSet := token.NewFileSet()
	file := fileSet.AddFile("", fileSet.Base(), len(content))
	var scan scanner.Scanner
	scan.Init(file, content, nil, 0)

	lines := map[int]struct{}{}
	for {
		pos, tok, _ := scan.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON {
			continue
		}
		lines[fileSet.Position(pos).Line] = struct{}{}
	}
	return len(lines)
}
