package digest

import (
	"bufio"
	"io"
	"strings"
)

// LineReader は1行分の文字配列を返すインターフェースです。
// csv.Reader もこれを満たします。
type LineReader interface {
	Read() ([]string, error)
}

// rawReader は単純な文字列分割を行うリーダーです。
type rawReader struct {
	scanner   *bufio.Scanner
	separator string
}

func newRawReader(r io.Reader, sep rune) *rawReader {
	return &rawReader{
		scanner:   bufio.NewScanner(r),
		separator: string(sep),
	}
}

func (r *rawReader) Read() ([]string, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	line := r.scanner.Text()
	return strings.Split(line, r.separator), nil
}

// getNextNLines の引数を *csv.Reader から LineReader インターフェースに変更します
func getNextNLines(reader LineReader) ([][]string, bool, error) {
	lines := make([][]string, bufferSize)

	lineCount := 0
	eofReached := false
	for ; lineCount < bufferSize; lineCount++ {
		line, err := reader.Read()
		lines[lineCount] = line
		if err != nil {
			if err == io.EOF {
				eofReached = true
				break
			}

			return nil, true, err
		}
	}

	return lines[:lineCount], eofReached, nil
}
