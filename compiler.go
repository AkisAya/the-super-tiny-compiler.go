package main

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Token struct {
	Name  string
	Value string
}

var blankPattern, _ = regexp.Compile(`\s`)
var numberPattern, _ = regexp.Compile(`[0-9]`)
var letterPattern, _ = regexp.Compile(`[a-zA-Z]`)

func isBlank(char string) bool {
	return blankPattern.MatchString(char)
}

func isNumber(char string) bool {
	return numberPattern.MatchString(char)
}

func isLetter(char string) bool {
	return letterPattern.MatchString(char)
}

func tokenize(input string) []Token{
	cur := 0
	var tokens []Token

	for cur < len(input) {
		char := string(input[cur])

		if char == "(" {
			tokens = append(tokens, Token{"paren", "("})
			cur++
			continue
		}

		if char == ")" {
			tokens = append(tokens, Token{"paren", ")"})
			cur++
			continue
		}

		if isBlank(char) {
			cur++
			continue
		}

		if isNumber(char) {
			value := ""
			for isNumber(char) {
				value += char
				cur++
				char = string(input[cur])
			}
			tokens = append(tokens, Token{"number", value})
			continue
		}

		if isLetter(char) {
			value := ""
			for isLetter(char) {
				value += char
				cur++
				char = string(input[cur])
			}
			tokens = append(tokens, Token{"method", value})
			continue
		}

		panic("unknown character: " + char)
	}
	return tokens
}

// LNode struct for Lisp Node,
type LNode struct {
	Kind     string
	Value    string
	Children []LNode
}


var parserIdx = 0
func parser(tokens []Token) LNode {

	ast := LNode{Kind: "Program", Children: []LNode{}}


	for parserIdx < len(tokens) {
		ast.Children = append(ast.Children, walk(tokens))
	}

	// reset
	parserIdx = 0
	return ast
}

func walk(tokens []Token) LNode {
	token := tokens[parserIdx]

	if token.Name == "number" {
		parserIdx++
		return LNode{Kind: "NumberLiteral", Value: token.Value}
	}

	// look for CallExpression
	if token.Name == "paren" && token.Value == "(" {
		// skip "("
		parserIdx++
		token = tokens[parserIdx]
		// token after "(" should be the method
		node := LNode{Kind: "CallExpression", Value: token.Value, Children: []LNode{}}

		// evaluated to the next token
		parserIdx++
		token = tokens[parserIdx]

		// loop till the end of a CallExpression, indicated by ")"
		for token.Name != "paren" ||
			(token.Name == "paren" && token.Value == "(") {
			node.Children = append(node.Children, walk(tokens))
			token = tokens[parserIdx]
		}
		// skip last token ")"
		parserIdx++

		return node
	}
	tokenStr, _ := json.Marshal(token)
	panic("invalid token: " + string(tokenStr))
}

// CNode struct for C AST
type CNode struct {
	Kind string
	Name string
	Body []CNode
	// must be pointer
	Expression *CNode
	Callee *CNode
	Arguments []CNode
}







func main() {
	res := parser(tokenize("(add 1 (subtract 2 3))"))
	res1, _ := json.Marshal(res)
	fmt.Println(string(res1))
}
