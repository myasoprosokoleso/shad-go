//go:build !solution

package main

import (
	"fmt"
	"strconv"
	"strings"
)

type stackFunc func() error

type Evaluator struct {
	stack []int
	words map[string][]stackFunc
}

// NewEvaluator creates evaluator.
func NewEvaluator() *Evaluator {
	e := Evaluator{}
	e.words = map[string][]stackFunc{
		"+":    {func() error { return e.processArithmetic("+") }},
		"-":    {func() error { return e.processArithmetic("-") }},
		"*":    {func() error { return e.processArithmetic("*") }},
		"/":    {func() error { return e.processArithmetic("/") }},
		"dup":  {e.processDup},
		"over": {e.processOver},
		"drop": {e.processDrop},
		"swap": {e.processSwap},
	}

	return &e
}

func (e *Evaluator) appendNumber(num int) error {
	e.stack = append(e.stack, num)
	return nil
}

func (e *Evaluator) checkOperands(operator string, required int) error {
	if len(e.stack) < required {
		return fmt.Errorf("required %d operands for %s, but got %d", required, operator, len(e.stack))
	}
	return nil
}

func (e *Evaluator) processArithmetic(operator string) error {
	if err := e.checkOperands(operator, 2); err != nil {
		return err
	}

	b := e.stack[len(e.stack)-1]
	a := e.stack[len(e.stack)-2]
	var res int
	switch operator {
	case "+":
		res = a + b
	case "-":
		res = a - b
	case "*":
		res = a * b
	default:
		if b == 0 {
			return fmt.Errorf("division by zero")
		}
		res = a / b
	}

	e.stack = e.stack[:len(e.stack)-2]
	e.stack = append(e.stack, res)
	return nil
}

func (e *Evaluator) processDup() error {
	if err := e.checkOperands("dup", 1); err != nil {
		return err
	}

	e.stack = append(e.stack, e.stack[len(e.stack)-1])
	return nil
}

func (e *Evaluator) processOver() error {
	if err := e.checkOperands("over", 2); err != nil {
		return err
	}

	e.stack = append(e.stack, e.stack[len(e.stack)-2])
	return nil
}

func (e *Evaluator) processDrop() error {
	if err := e.checkOperands("drop", 1); err != nil {
		return err
	}

	e.stack = e.stack[:len(e.stack)-1]
	return nil
}

func (e *Evaluator) processSwap() error {
	if err := e.checkOperands("swap", 2); err != nil {
		return err
	}

	e.stack[len(e.stack)-1], e.stack[len(e.stack)-2] = e.stack[len(e.stack)-2], e.stack[len(e.stack)-1]
	return nil
}

func (e *Evaluator) parseInput(input []string) ([]stackFunc, error) {
	var parsed []stackFunc

	for i := 0; i < len(input); {
		if definition, ok := e.words[input[i]]; ok {
			parsed = append(parsed, definition...)
			i++
			continue
		}

		if input[i] != ":" {
			num, err := strconv.Atoi(input[i])
			if err != nil {
				return nil, fmt.Errorf("unknown word: %s", input[i])
			}

			f := func() error { return e.appendNumber(num) }
			parsed = append(parsed, f)
			i++
			continue
		}

		word := input[i+1]
		if _, err := strconv.Atoi(word); err == nil {
			return nil, fmt.Errorf("word cannot be a number: %s", word)
		}

		wordStart := i + 2
		wordEnd := wordStart
		for input[wordEnd] != ";" {
			wordEnd++
		}

		definition, err := e.parseInput(input[wordStart:wordEnd])
		if err != nil {
			return nil, err
		}
		e.words[word] = definition
		i = wordEnd + 1
	}

	return parsed, nil
}

// Process evaluates sequence of words or definition.
//
// Returns resulting stack state and an error.
func (e *Evaluator) Process(row string) ([]int, error) {
	input := strings.Fields(strings.ToLower(row))

	definition, err := e.parseInput(input)
	if err != nil {
		return nil, err
	}

	for _, f := range definition {
		if err := f(); err != nil {
			return nil, err
		}
	}

	return e.stack, nil
}
