package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Token struct {
	Name  string `json:",omitempty"`
	Value string `json:",omitempty"`
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

// LNode struct for Lisp AST
type LNode struct {
	Kind     string  `json:",omitempty"`
	Value    string  `json:",omitempty"`
	Children []LNode `json:",omitempty"`
}


func parser(tokens []Token) LNode {
	ast := LNode{Kind: "Program", Children: []LNode{}}

	idx := 0
	var node LNode
	for idx < len(tokens) {
		node, idx = walk(tokens, idx)
		ast.Children = append(ast.Children, node)
	}
	return ast
}
// iterate over tokens and return constructed Node and idx to the next token
func walk(tokens []Token, idx int) (LNode, int) {
	token := tokens[idx]

	if token.Name == "number" {
		idx++
		return LNode{Kind: "NumberLiteral", Value: token.Value}, idx
	}

	// look for CallExpression
	if token.Name == "paren" && token.Value == "(" {
		// skip "("
		idx++
		token = tokens[idx]
		// token after "(" should be the method
		node := LNode{Kind: "CallExpression", Value: token.Value, Children: []LNode{}}

		// evaluated to the next token
		idx++
		token = tokens[idx]

		// loop till the end of a CallExpression, indicated by ")"
		var nestedNode LNode
		for token.Name != "paren" ||
			(token.Name == "paren" && token.Value == "(") {
			nestedNode, idx = walk(tokens, idx)
			node.Children = append(node.Children, nestedNode)
			token = tokens[idx]
		}
		// skip last token ")"
		idx++
		return node, idx
	}
	tokenStr, _ := json.Marshal(token)
	panic("invalid token: " + string(tokenStr))
}

// CNode struct for C AST
type CNode struct {
	Kind string          `json:",omitempty"`
	Name string          `json:",omitempty"`
	Value string         `json:",omitempty"`
	Body []*CNode         `json:",omitempty"`
	// must be pointer
	Expression *CNode    `json:",omitempty"`
	Callee *CNode        `json:",omitempty"`
	Arguments []*CNode   `json:",omitempty"`
}



type Visitor interface {
	Enter(lNode LNode, parent LNode, cNodeParent *CNode) *CNode
	Leave(lNode LNode, parent LNode, cNodeParent *CNode) *CNode
}

type ProgramVisitor struct {}
type CallExpressionVisitor struct {}
type NumberLiteralVisitor struct {}

func (program *ProgramVisitor) Enter(lNode LNode, parent LNode, cNodeParent *CNode) *CNode {
	*cNodeParent = CNode{
		Kind: "Program",
		Body: []*CNode{},
	}
	return cNodeParent
}

func (program *ProgramVisitor) Leave(lNode LNode, parent LNode, cNodeParent *CNode)  *CNode {
	// do nothing
	return nil
}

func (callExpression *CallExpressionVisitor) Enter(lNode LNode, parent LNode, cNodeParent *CNode) *CNode {
	expression := CNode{
		Kind: "CallExpression",
		Callee: &CNode{
			Kind: "Identifier",
			Name: lNode.Value,
		},
		Arguments: []*CNode{},
	}

	if parent.Kind != "CallExpression" {
		expressionStatement := CNode{
			Kind: "ExpressionStatement",
			Expression: &expression,
		}
		cNodeParent.Body = append(cNodeParent.Body, &expressionStatement)
	} else {
		cNodeParent.Arguments = append(cNodeParent.Arguments, &expression)
	}
	return &expression
}

func (callExpression *CallExpressionVisitor) Leave(lNode LNode, parent LNode, cNodeParent *CNode) *CNode {
	// do nothing
	return nil
}

func (numberLiteral *NumberLiteralVisitor) Enter(lNode LNode, parent LNode, cNodeParent *CNode) *CNode {
	cNodeParent.Arguments = append(cNodeParent.Arguments, &CNode{Kind: "NumberLiteral", Value: lNode.Value})
	return nil
}

func (numberLiteral *NumberLiteralVisitor) Leave(lNode LNode, parent LNode, cNodeParent *CNode) *CNode {
	// do nothing
	return nil
}


func traverser(ast LNode, visitors map[string]Visitor) CNode {
	cNode := &CNode{}
	traverseLNode(ast, LNode{}, cNode, visitors)
	return *cNode
}

func traverseLNode(lNode LNode, parent LNode, cNodeParent *CNode, visitors map[string]Visitor)  {

	visitorFn := visitors[lNode.Kind]

	cNode := visitorFn.Enter(lNode, parent, cNodeParent)

	switch lNode.Kind {
	case "Program", "CallExpression":
		for _, child := range lNode.Children {
			traverseLNode(child, lNode, cNode, visitors)
		}

	}

	visitorFn.Leave(lNode, parent, cNodeParent)
}


func transform(ast LNode) CNode {

	visitorMap := make(map[string]Visitor)
	visitorMap["Program"] = &ProgramVisitor{}
	visitorMap["CallExpression"] = &CallExpressionVisitor{}
	visitorMap["NumberLiteral"] = &NumberLiteralVisitor{}

	return traverser(ast, visitorMap)

}

func CCodeGenerator(node CNode) string {

	switch node.Kind {
	case "Program":
		var exprStmts []string
		for _, exprStmtNode := range node.Body {
			exprStmts = append(exprStmts, CCodeGenerator(*exprStmtNode))
		}
		return strings.Join(exprStmts, "\n")
	case "ExpressionStatement":
		return CCodeGenerator(*node.Expression) + ";"
	case "CallExpression":
		var args []string
		for _, arg := range node.Arguments {
			args = append(args, CCodeGenerator(*arg))
		}
		return fmt.Sprintf("%s(%s)", node.Callee.Name, strings.Join(args, ", "))
	case "NumberLiteral":
		return node.Value
	}
	panic("Unexpected Node Type: " + node.Kind)
}

func compileListToC(input string) string {
	tokens := tokenize(input)
	ast := parser(tokens)
	cAst := transform(ast)
	return CCodeGenerator(cAst)
}


func main() {
	input := "(add 1 (subtract 2 3))"
	fmt.Printf("Input:\n%s\nOutput:\n%s\n", input, compileListToC(input))
	input2 := "(add 1 2)\n(subtract 3 4)"
	fmt.Printf("Input:\n%s\nOutput:\n%s\n", input2, compileListToC(input2))

}
