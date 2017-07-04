package component

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type meta interface {
	order() []string
	pointers() args
	values() args
}

var metaRow = regexp.MustCompile(`\[([a-zA-Z]+)\]: # \(([^)]{0,})\)`)

func checkMeta(meta string, m meta) error {
	order := m.order()
	rows := strings.Split(meta, "\n")
	if len(rows) != len(order) {
		return ErrContent
	}
	for i := range order {
		m := metaRow.FindStringSubmatch(rows[i])
		if len(m) != 3 || m[1] != order[i] {
			return ErrContent
		}
	}
	return nil
}

type args []interface{}

func setMeta(meta string, m meta) error {
	order, pointers := m.order(), m.pointers()
	rows := strings.Split(strings.TrimSpace(meta), "\n")
	if len(rows) != len(order) {
		return ErrContent
	}
	for i, p := range pointers {
		v := metaRow.FindStringSubmatch(rows[i])[2]
		switch pointer := p.(type) {
		case *string:
			*pointer = v
		case *bool:
			*pointer = v == "true"
		case *int:
			if v == "" {
				break
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return ErrContent
			}
			*pointer = n
		case *float64:
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return ErrContent
			}
			*pointer = n
		case *[]string:
			if strings.TrimSpace(v) == "" {
				break
			}
			*pointer = strings.Split(v, "|")
		default:
			panic(fmt.Sprintf("Unknown type: %T", pointer))
		}
	}
	return nil
}

func getMeta(m meta) string {
	order := m.order()
	b := bytes.NewBuffer(nil)
	for i, v := range m.values() {
		if i > 0 {
			b.WriteRune('\n')
		}
		switch t := v.(type) {
		case []string:
			v = strings.Join(t, "|")
		}
		fmt.Fprintf(b, "[%s]: # (%v)", order[i], v)
	}
	return b.String()
}
