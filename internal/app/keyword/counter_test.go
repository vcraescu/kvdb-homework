package keyword_test

import (
	"context"
	"emag-homework/internal/app/keyword"
	"emag-homework/pkg/test/require"
	"testing"
)

func TestCounter_Count(t *testing.T) {
	t.Parallel()

	type args struct {
		text string
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]int
		wantErr bool
	}{
		{
			name: "text with punctuation, numbers and spaces",
			args: args{
				text: "Lorem Ipsum  , is   simply   dummy 1500 text of the printing and typesetting industry! Dummy text simply lorem    ipsum ",
			},
			want: map[string]int{
				"lorem":       2,
				"ipsum":       2,
				"is":          1,
				"simply":      2,
				"dummy":       2,
				"text":        2,
				"of":          1,
				"the":         1,
				"printing":    1,
				"and":         1,
				"typesetting": 1,
				"industry":    1,
			},
		},
		{
			name: "empty text",
			args: args{
				text: "",
			},
		},
		{
			name: "empty text with punctuation and numbers",
			args: args{
				text: "! ?;125",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			counter := keyword.NewCounter()
			got, err := counter.Count(context.Background(), tt.args.text)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
