package component

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var metaRow = regexp.MustCompile(`\[([a-zA-Z]+)\]: # \(([^)]{0,})\)`)

func checkMeta(meta string, order []string) error {
	rows := strings.Split(meta, "\n")
	if len(rows) != len(order) {
		return ErrInvalid
	}
	for i := range order {
		m := metaRow.FindStringSubmatch(rows[i])
		if len(m) != 3 || m[1] != order[i] {
			return ErrInvalid
		}
	}
	return nil
}

type args []interface{}

func setMeta(meta string, order []string, pointers args) error {
	rows := strings.Split(strings.TrimSpace(meta), "\n")
	if len(rows) != len(order) {
		return ErrInvalid
	}
	for i, p := range pointers {
		v := metaRow.FindStringSubmatch(rows[i])[2]
		switch pointer := p.(type) {
		case *string:
			*pointer = v
		case *bool:
			*pointer = v == "true"
		case *int:
			n, err := strconv.Atoi(v)
			if err != nil {
				return ErrInvalid
			}
			*pointer = n
		case *float64:
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return ErrInvalid
			}
			*pointer = n
		default:
			panic(fmt.Sprintf("Unknown type: %T", pointer))
		}
	}
	return nil
}

func getMeta(order []string, values args) string {
	b := bytes.NewBuffer(nil)
	for i := range values {
		if i > 0 {
			b.WriteRune('\n')
		}
		fmt.Fprintf(b, "[%s]: # (%v)", order[i], values[i])
	}
	return b.String()
}
