package querybuilder

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/require"
)

type UserID struct{ uuid.UUID }

func TestDebug_Different_Types(t *testing.T) {
	id := uuid.Must(uuid.NewV4())
	uid := UserID{id}

	tt := []struct {
		name   string
		query  string
		arg    any
		expect string
	}{
		{
			name:   "uuid",
			query:  "select * from access.users where id = ?",
			arg:    id,
			expect: fmt.Sprintf("select * from access.users where id = '%s'", id),
		},
		{
			name:   "custom uuid",
			query:  "select * from access.users where id = ?",
			arg:    uid,
			expect: fmt.Sprintf("select * from access.users where id = '%s'", id),
		},
		{
			name:   "int",
			query:  "select * from access.users where id = ?",
			arg:    10,
			expect: "select * from access.users where id = 10",
		},
		{
			name:   "float",
			query:  "select * from access.users where id = ?",
			arg:    10.1,
			expect: "select * from access.users where id = 10.1",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			qb := New(tc.query, tc.arg)
			require.Equal(t, tc.expect, qb.Debug(), "serialized query should match")
		})
	}
}
