package boolexpr

import (
	"strconv"
	"strings"

	parsec "github.com/prataprc/goparsec"
)

type (
	Sym        string
	Comparison struct {
		Left  parsec.ParsecNode
		Right parsec.ParsecNode
		Op    parsec.ParsecNode
	}
	LogicalExpr []parsec.ParsecNode
	Group       struct {
		Children parsec.ParsecNode
	}
	And []parsec.ParsecNode
	Or  []parsec.ParsecNode
	Exp []parsec.ParsecNode
)

var (
	literal = parsec.OrdChoice(
		func(pn []parsec.ParsecNode) parsec.ParsecNode {
			switch p := pn[0].(type) {
			case *parsec.Terminal:
				switch p.Name {
				case "SYMBOL":
					return Sym(p.Value)
				case "INT":
					v, err := strconv.ParseInt(p.Value, 10, 64)
					if err != nil {
						return p
					}
					return float64(v)
				case "FLOAT":
					v, err := strconv.ParseFloat(p.Value, 64)
					if err != nil {
						return p
					}
					return float64(v)
				default:
					return p
				}
			case string:
				return strings.Trim(p, `"`)
			default:
				return p
			}
		},
		parsec.Float(),
		parsec.Int(),
		parsec.String(),
		parsec.Token(`[a-zA-Z\-_]+`, "SYMBOL"),
	)

	opers = parsec.OrdChoice(
		func(pn []parsec.ParsecNode) parsec.ParsecNode {
			return pn[0].(*parsec.Terminal).Value
		},
		parsec.Atom("=", "EQ"),
		parsec.Atom(">=", "GTE"),
		parsec.Atom("<=", "LTE"),
		parsec.Atom(">", "GT"),
		parsec.Atom("<", "LT"),
		parsec.Atom("!=", "NEQ"),
	)

	comparison = parsec.And(
		func(pn []parsec.ParsecNode) parsec.ParsecNode {
			return Comparison{
				Left:  pn[0],
				Op:    pn[1],
				Right: pn[2],
			}
		},
		literal,
		opers,
		literal,
	)

	logicalOps = parsec.OrdChoice(
		func(pn []parsec.ParsecNode) parsec.ParsecNode {
			return pn[0].(*parsec.Terminal).Value
		},
		parsec.Atom("and", "AND"),
		parsec.Atom("or", "OR"),
	)

	logicalExprs = parsec.And(
		func(pn []parsec.ParsecNode) parsec.ParsecNode {
			children := LogicalExpr{pn[0]}
			for _, n := range pn[1:] {
				nc, ok := n.([]parsec.ParsecNode)
				if ok {
					children = append(children, nc...)
				}
			}
			return children
		},
		comparison,
		parsec.Maybe(
			func(pn []parsec.ParsecNode) parsec.ParsecNode {
				return pn[0]
			},
			parsec.Many(
				func(pn []parsec.ParsecNode) parsec.ParsecNode {
					children := []parsec.ParsecNode{}
					for _, n := range pn {
						nc, ok := n.([]parsec.ParsecNode)
						if ok {
							children = append(children, nc...)
						}
					}
					return children
				},
				parsec.And(
					nil,
					logicalOps,
					comparison,
				),
			),
		),
	)

	group = parsec.And(
		func(pn []parsec.ParsecNode) parsec.ParsecNode {
			return Group{
				Children: pn[1],
			}
		},
		parsec.Atom("(", "OPENGRP"),
		logicalExprs,
		parsec.Atom(")", "CLOSEGRP"),
	)

	exprOrGroup = parsec.OrdChoice(
		func(pn []parsec.ParsecNode) parsec.ParsecNode {
			return pn
		},
		group,
		logicalExprs,
	)

	Parser = parsec.And(
		func(pn []parsec.ParsecNode) parsec.ParsecNode {
			children := Exp{pn[0]}
			for _, n := range pn[1:] {
				nc, ok := n.([]parsec.ParsecNode)
				if ok {
					children = append(children, nc...)
				}
			}
			return Exp(children)
		},
		exprOrGroup,
		parsec.Maybe(
			func(pn []parsec.ParsecNode) parsec.ParsecNode {
				return pn[0]
			},
			parsec.Many(
				func(pn []parsec.ParsecNode) parsec.ParsecNode {
					children := []parsec.ParsecNode{}
					for _, n := range pn {
						nc, ok := n.([]parsec.ParsecNode)
						if ok {
							children = append(children, nc...)
						}
					}
					return children
				},
				parsec.And(
					nil,
					logicalOps,
					exprOrGroup,
				),
			),
		),
	)
)
