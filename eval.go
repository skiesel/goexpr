package goexpr

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
)

type FunctionEvaluator func([]float64)(float64, error)

// Evaluate evalues an expression, given a scope.
func Evaluate(parsed *Expression, scope map[string]float64, functions map[string]FunctionEvaluator) (float64, error) {

	result, err := evaluate(parsed.Ast, scope, functions)

	if err != nil {
		return 0, err
	}

	return result, nil
}

func evaluate(node ast.Node, scope map[string]float64, functions map[string]FunctionEvaluator) (value float64, err error) {

	switch node.(type) {

	case *ast.Ident:
		value, err = evaluateIdent(node.(*ast.Ident), scope)

	case *ast.BinaryExpr:
		value, err = evaluateBinary(node.(*ast.BinaryExpr), scope, functions)

	case *ast.ParenExpr:
		value, err = evaluate(node.(*ast.ParenExpr).X, scope, functions)

	case *ast.BasicLit:
		value, err = strconv.ParseFloat(node.(*ast.BasicLit).Value, 64)

	case *ast.CallExpr:
		argValues := make([]float64, len(node.(*ast.CallExpr).Args))
		for i, arg := range node.(*ast.CallExpr).Args {
			argValues[i], err = evaluate(arg, scope, functions)
			if err != nil {
				return value, err
			}
		}

		value, err = evaluateCallExpr(node.(*ast.CallExpr), functions, argValues)

	default:
		err = fmt.Errorf("unsupported node %+v (type %+v)", node, reflect.TypeOf(node))
	}

	return value, err
}

func evaluateIdent(node *ast.Ident, scope map[string]float64) (float64, error) {

	value, found := scope[node.Name]

	if !found {
		return 0, fmt.Errorf("no value for %s in scope %v", node.Name, scope)
	}

	return value, nil
}

func evaluateBinary(node *ast.BinaryExpr, scope map[string]float64, functions map[string]FunctionEvaluator) (float64, error) {

	lValue, err := evaluate(node.X, scope, functions)

	if err != nil {
		return 0, err
	}

	rValue, err := evaluate(node.Y, scope, functions)

	if err != nil {
		return 0, err
	}

	var value float64

	switch node.Op {
	case token.ADD:
		value = lValue + rValue
	case token.SUB:
		value = lValue - rValue
	case token.MUL:
		value = lValue * rValue
	case token.QUO:
		value = lValue / rValue
	default:
		err = fmt.Errorf("unsupported binary operation: %s", node.Op)
	}

	return value, err
}

func evaluateCallExpr(node *ast.CallExpr, functions map[string]FunctionEvaluator, argValues []float64) (float64, error) {
	functionName := node.Fun.(*ast.Ident).Name
	function, found := functions[functionName]

	if !found {
		return 0, fmt.Errorf("no function for %s in functions %v", functionName, functions)
	}

	return function(argValues)
}