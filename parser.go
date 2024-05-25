package boolexpr

import (
	"strconv"

	parsec "github.com/prataprc/goparsec"
)

type SYM string
type Expr []parsec.ParsecNode
type Exprs []parsec.ParsecNode
type Group parsec.ParsecNode

var (
	symbol  = parsec.Token(`[a-z]+`, "SYMBOL")
	literal = parsec.OrdChoice(func(pn []parsec.ParsecNode) parsec.ParsecNode {
		switch p := pn[0].(type) {
		case *parsec.Terminal:
			switch p.Name {
			case "SYMBOL":
				return SYM(p.Value)
			case "INT":
				v, err := strconv.ParseInt(p.Value, 10, 64)
				if err != nil {
					return v
				}
				return v
			default:
				return p
			}
		default:
			return p
		}
	}, parsec.Float(), parsec.Int(), parsec.String(), symbol)

	and        = parsec.Atom("and", "AND")
	or         = parsec.Atom("or", "OR")
	logicalOps = parsec.OrdChoice(func(pn []parsec.ParsecNode) parsec.ParsecNode {
		return pn[0]
	}, and, or)

	eq    = parsec.Atom("=", "EQ")
	gt    = parsec.Atom(">", "GT")
	lt    = parsec.Atom("<", "LT")
	gte   = parsec.Atom(">=", "GTE")
	lte   = parsec.Atom("<=", "LTE")
	neq   = parsec.Atom("!=", "NEQ")
	opers = parsec.OrdChoice(func(pn []parsec.ParsecNode) parsec.ParsecNode {
		return pn[0]
	}, eq, gt, lt, gte, lte, neq)

	comparison = parsec.And(func(pn []parsec.ParsecNode) parsec.ParsecNode {
		return Expr(pn)
	}, literal, opers, literal)

	logicalExprs = parsec.Kleene(func(pn []parsec.ParsecNode) parsec.ParsecNode {
		return Exprs(pn)
	}, comparison, logicalOps)

	opengroup  = parsec.Atom("(", "OPENGRP")
	closegroup = parsec.Atom(")", "CLOSEGRP")
	group      = parsec.And(func(pn []parsec.ParsecNode) parsec.ParsecNode {
		return Group(pn[1])
	}, opengroup, logicalExprs, closegroup)

	exprOrGroup = parsec.OrdChoice(func(pn []parsec.ParsecNode) parsec.ParsecNode {
		return pn
	}, group, logicalExprs)

	Parser = parsec.Kleene(nil, exprOrGroup, logicalOps)
)
