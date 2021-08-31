package parser

import (
	"fmt"
	"interpreter-go/ast"
	"interpreter-go/lexer"
	"interpreter-go/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // < or >
	SUM         // +
	PRODUCT     // *
	PREFIX      // -x or !x
	CALL        // myFunc()
)

type Parser struct {
	lexer           *lexer.Lexer
	curToken        token.Token
	peekToken       token.Token
	errors          []string
	prefrixParseFns map[token.TokenType]prefrixParseFn
	infixParseFns   map[token.TokenType]infixParseFn
}

type (
	prefrixParseFn func() ast.Expression
	infixParseFn   func(ast.Expression) ast.Expression
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
}

func New(lexer *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:           lexer,
		errors:          []string{},
		prefrixParseFns: make(map[token.TokenType]prefrixParseFn),
		infixParseFns:   make(map[token.TokenType]infixParseFn),
	}

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (parser *Parser) Errors() []string {
	return parser.errors
}

func (parser *Parser) peekError(t token.TokenType) {
	message := fmt.Sprintf("Expected next token to be %s, got %s instead", t, parser.peekToken.Type)

	parser.errors = append(parser.errors, message)
}

func (parser *Parser) nextToken() {
	parser.curToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
}

func (parser *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for parser.curToken.Type != token.EOF {
		statement := parser.parseStatement()

		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}

		parser.nextToken()
	}

	return program
}

func (parser *Parser) parseStatement() ast.Statement {
	switch parser.curToken.Type {
	case token.LET:
		return parser.parseLetStatement()
	case token.RETURN:
		return parser.parseReturnStatement()
	default:
		return parser.parseExpressionStatement()
	}
}

func (parser *Parser) parseLetStatement() *ast.LetStatement {
	statement := &ast.LetStatement{Token: parser.curToken}

	if !parser.expectPeek(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{Token: parser.curToken, Value: parser.curToken.Literal}

	if !parser.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: skipping the expression until we encounter a semicolon
	for !parser.curTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}

	return statement
}

func (parser *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: parser.curToken}

	parser.nextToken()

	for !parser.curTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}

	return statement
}

func (parser *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: parser.curToken}

	stmt.Expression = parser.parseExpression(LOWEST)

	if parser.peekTokenIs(token.SEMICOLON) {
		parser.nextToken()
	}

	return stmt
}

func (parser *Parser) parseExpression(precedence int) ast.Expression {
	prefix := parser.prefrixParseFns[parser.curToken.Type]

	if prefix == nil {
		parser.noPrefixParseFnError(parser.curToken.Type)
		return nil
	}

	leftExpression := prefix()

	for !parser.peekTokenIs(token.SEMICOLON) && precedence < parser.peekPrecedence() {
		infix := parser.infixParseFns[parser.peekToken.Type]
		if infix == nil {
			return leftExpression
		}

		parser.nextToken()
		leftExpression = infix(leftExpression)
	}

	return leftExpression
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefrixParseFn) {
	p.prefrixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)

		p.errors = append(p.errors, msg)

		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if precedence, ok := precedences[p.peekToken.Type]; ok {
		return precedence
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if precedence, ok := precedences[p.curToken.Type]; ok {
		return precedence
	}

	return LOWEST
}
