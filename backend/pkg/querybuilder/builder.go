package querybuilder

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/gofrs/uuid/v5"
)

type Builder struct {
	sb   strings.Builder
	args []any
}

func New(sql string, args ...any) *Builder {
	b := &Builder{}
	if err := b.scan(sql, args...); err != nil {
		panic(err)
	}

	return b
}

func (b *Builder) scan(query string, args ...any) error {
	count := 0
	for _, r := range query {
		if r == '?' {
			count++
		}

		b.sb.WriteRune(r)
	}

	if count != len(args) {
		return fmt.Errorf("invalid number of arguments for %s: %v", query, args)
	}

	b.args = append(b.args, args...)
	return nil
}

type Cond struct {
	Text string
	Args []any
}

func (b *Builder) Where(conds ...Cond) {
	if len(conds) == 0 {
		return
	}

	first := true
	for _, cond := range conds {
		if cond.Text == "" {
			continue
		}

		open := " and ("
		if first {
			open = " where ("
			first = false
		}

		b.sb.WriteString(open)
		if err := b.scan(cond.Text, cond.Args...); err != nil {
			panic(err)
		}
		b.sb.WriteRune(')')
	}
}

func (b *Builder) Append(query string, args ...any) {
	if err := b.scan(query, args...); err != nil {
		panic(err)
	}
}

func Condition(cond string, args []any) Cond {
	return Cond{cond, args}
}

func (b *Builder) SQL() (string, []any) {
	return b.sb.String(), b.args
}

func (b *Builder) Debug() string {
	out := strings.Builder{}
	index := 0
	for _, c := range b.sb.String() {
		if c != '?' {
			out.WriteRune(c)
			continue
		}

		param := b.args[index]
		index++

		if s, ok := canString(param); ok {
			out.WriteString(s)
			continue
		}

		if n, ok := canNumber(param); ok {
			out.WriteString(n)
			continue
		}

		out.WriteString(fmt.Sprintf("'%s'", param))
		// TODO: support can integer
		// TODO: support can array
	}

	return out.String()
}

func canString(param any) (string, bool) {
	t, v := reflect.TypeOf(param), reflect.ValueOf(param)
	if t.Kind() == reflect.Pointer {
		return canString(v.Elem())
	}

	if t.Kind() == reflect.String {
		return fmt.Sprintf("'%s'", v.String()), true
	}

	id := uuid.Nil
	if v.CanConvert(reflect.TypeOf(id)) {
		id = v.Convert(reflect.TypeOf(id)).Interface().(uuid.UUID)
		return fmt.Sprintf("'%s'", id), true
	}

	null := sql.NullString{}
	if v.CanConvert(reflect.TypeOf(null)) {
		null = v.Convert(reflect.TypeOf(null)).Interface().(sql.NullString)
		if null.Valid {
			return fmt.Sprintf("'%s'", null.String), true
		}
		return "null", true
	}

	return "", false
}

func canNumber(param any) (string, bool) {
	t, v := reflect.TypeOf(param), reflect.ValueOf(param)
	if t.Kind() == reflect.Pointer {
		return canNumber(v.Elem())
	}

	switch t.Kind() {
	case
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return fmt.Sprint(v.Interface()), true
	}

	return "", false
}
