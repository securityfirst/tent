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
	optionals() []string
	pointers() args
	values() args
}

var metaRow = regexp.MustCompile(`\[([a-zA-Z]+)\]: # \((.*)\)`)

type args []interface{}

func setMeta(meta string, m meta) error {
	order, pointers, optionals := m.order(), m.pointers(), m.optionals()
	rows := strings.Split(meta, "\n")
	if len(rows) < len(order)-len(optionals) {
		return fmt.Errorf("Expected %v-%v, got %v", len(order)-len(optionals), len(order), len(rows))
	}
	for i, row := range rows {
		m := metaRow.FindStringSubmatch(row)
		if len(m) != 3 {
			return fmt.Errorf("Invalid %q", row)
		}
		for {
			if m[1] == order[i] {
				if len(optionals) > 0 && order[i] == optionals[0] {
					optionals = optionals[1:]
				}
				if err := setMetaValue(pointers[i], m[2]); err != nil {
					return fmt.Errorf("meta: %s", err.Error())
				}
				break
			}
			if len(optionals) == 0 || order[i] != optionals[0] {
				return fmt.Errorf("Expected %v, got %v", order[i], optionals[0])
			}
			optionals = optionals[1:]
			pointers = pointers[:i+copy(pointers[i:], pointers[i+1:])]
			order = order[:i+copy(order[i:], order[i+1:])]
		}
	}
	return nil
}

func setMetaValue(p interface{}, v string) error {
	switch pointer := p.(type) {
	case *string:
		*pointer = v
	case *bool:
		*pointer = v == "true"
	case *int:
		if v == "" {
			return nil
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("Invalid int: %v", v)
		}
		*pointer = n
	case *float64:
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("Invalid float: %v", v)
		}
		*pointer = n
	case *[]string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		*pointer = strings.Split(v, ";")
	default:
		return fmt.Errorf("unknown type %T", pointer)
	}
	return nil
}

func getMeta(m meta) string {
	order, optionals := m.order(), m.optionals()
	b := bytes.NewBuffer(nil)
	for i, v := range m.values() {
		var isZero bool
		switch t := v.(type) {
		case string:
			isZero = t == ""
		case int, float64:
			isZero = t == 0
		case bool:
			isZero = t
		case []string:
			isZero = len(t) == 0 || len(t) == 1 && t[0] == ""
			v = strings.Join(t, ";")
		}
		if len(optionals) != 0 && order[i] == optionals[0] {
			optionals = optionals[1:]
			if isZero {
				continue
			}
		}
		if i > 0 {
			b.WriteRune('\n')
		}
		fmt.Fprintf(b, "[%s]: # (%v)", order[i], v)
	}
	return b.String()
}
