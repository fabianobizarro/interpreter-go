package repl

import (
	"bufio"
	"fmt"
	"interpreter-go/lexer"
	"interpreter-go/token"
	"io"
)

const PROMPT = ">>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		lexer := lexer.New(line)

		for tk := lexer.NextToken(); tk.Type != token.EOF; tk = lexer.NextToken() {
			fmt.Fprintf(out, "%+v\n", tk)
		}
	}
}
