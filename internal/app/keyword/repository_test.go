package keyword_test

import (
	"context"
	"emag-homework/internal/app/keyword"
	"emag-homework/pkg/test/require"
	"testing"
)

func TestInMemRepository_Increment(t *testing.T) {
	t.Parallel()

	type fields struct {
		data map[string]int
	}

	type args struct {
		keyword   string
		increment int
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "keyword already exists",
			fields: fields{
				data: map[string]int{
					"foo": 3,
					"bar": 2,
				},
			},
			args: args{
				keyword:   "foo",
				increment: 2,
			},
			want: 5,
		},
		{
			name: "keyword does not exists",
			fields: fields{
				data: map[string]int{
					"foo": 3,
					"bar": 2,
				},
			},
			args: args{
				keyword:   "baz",
				increment: 6,
			},
			want: 6,
		},
		{
			name: "empty controller",
			args: args{
				keyword:   "foo",
				increment: 3,
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			repo := keyword.NewInMemRepository(tt.fields.data)
			err := repo.Increment(ctx, tt.args.keyword, tt.args.increment)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			got, err := repo.Find(ctx, tt.args.keyword)

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
