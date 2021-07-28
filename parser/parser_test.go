package parser

import (
	"interpreter-go/ast"
	"interpreter-go/lexer"
	"testing"
)

func TestLetStatement(t *testing.T) {
	input := `let x = 5;
		let y = 10;
		let foobar = 838383;`

	lexer := lexer.New(input)
	parser := New(lexer)

	program := parser.ParseProgram()

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("ParseProgram() returned %d statements, expected 3", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {

		statement := program.Statements[i]
		if !testLetStatement(t, statement, tt.expectedIdentifier) {
			return
		}

	}

}

func testLetStatement(t *testing.T, statement ast.Statement, name string) bool {

	if statement.TokenLiteral() != "let" {
		t.Errorf("%d. statement.TokenLiteral() is %s, expected %s", 0, statement.TokenLiteral(), "let")
		return false
	}

	letStatement, ok := statement.(*ast.LetStatement)
	if !ok {
		t.Errorf("%d. statement is %T, expected *ast.LetStatement", 0, statement)
		return false
	}

	if letStatement.Name.Value != name {
		t.Errorf("%d. letStatement.Name.Value is %s, expected %s", 0, letStatement.Name.Value, name)
		return false
	}

	if letStatement.TokenLiteral() != name {
		t.Errorf("%d. letStatement.TokenLiteral() is %s, expected %s", 0, letStatement.TokenLiteral(), name)
		return false
	}

	return true

}
