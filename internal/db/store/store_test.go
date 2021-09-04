package store_test

import (
	"emag-homework/internal/db/store"
	"emag-homework/pkg/test/require"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestStore_Put(t *testing.T) {
	t.Parallel()

	type fields struct {
		filename      string
		flushInterval time.Duration
		entries       []store.Entry
	}

	type args struct {
		entry store.Entry
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    store.Entry
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				filename: func() string {
					rand.Seed(time.Now().UnixNano())

					return fmt.Sprintf("/tmp/test_%d.json", rand.Int())
				}(),
				flushInterval: 0,
			},
			args: args{
				entry: store.Entry{
					Key:     "foobar",
					Value:   []byte(fmt.Sprint(1)),
					Version: time.Now().UnixNano(),
				},
			},
			want: store.Entry{
				Key:     "foobar",
				Value:   []byte(fmt.Sprint(1)),
				Version: time.Now().UnixNano(),
			},
		},
		{
			name: "older key not stored",
			fields: fields{
				filename: func() string {
					rand.Seed(time.Now().UnixNano())

					return fmt.Sprintf("/tmp/test_%d.json", rand.Int())
				}(),
				flushInterval: 0,
				entries: []store.Entry{
					{
						Key:     "foobar",
						Value:   []byte(fmt.Sprint(10)),
						Version: 10,
					},
				},
			},
			args: args{
				entry: store.Entry{
					Key:     "foobar",
					Value:   []byte(fmt.Sprint(5)),
					Version: 5,
				},
			},
			want: store.Entry{
				Key:     "foobar",
				Value:   []byte(fmt.Sprint(10)),
				Version: 10,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer os.Remove(tt.fields.filename)

			s, err := store.New(
				store.WithFilename(tt.fields.filename),
				store.WithFlushInternval(tt.fields.flushInterval),
			)
			require.NoError(t, err, tt.fields.filename)

			for _, entry := range tt.fields.entries {
				err := s.Put(entry)
				require.Error(t, err)
			}

			err = s.Put(tt.args.entry)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			got := s.Get(tt.args.entry.Key)

			require.Equal(t, tt.want.Value, got.Value)
		})
	}
}

func TestStore_Flush(t *testing.T) {
	t.Parallel()

	rand.Seed(time.Now().UnixNano())

	filename := fmt.Sprintf("/tmp/test_%d.json", rand.Int())
	defer os.Remove(filename)

	s, err := store.New(store.WithFilename(filename), store.WithFlushInternval(time.Second))
	require.NoError(t, err, filename)

	entries := []store.Entry{
		{
			Key:     "key-1",
			Value:   []byte(fmt.Sprint(1)),
			Version: 1,
		},
		{
			Key:     "key-2",
			Value:   []byte(fmt.Sprint(2)),
			Version: 1,
		},
		{
			Key:     "key-3",
			Value:   []byte(fmt.Sprint(3)),
			Version: 1,
		},
	}

	for _, entry := range entries {
		err = s.Put(entry)

		require.NoError(t, err, "PUT")
	}

	err = s.Flush()
	require.NoError(t, err, "FLUSH")

	err = s.Close()
	require.NoError(t, err, "CLOSE")

	s, err = store.New(store.WithFilename(filename), store.WithFlushInternval(time.Second))

	require.NoError(t, err, filename)
	require.Equal(t, len(entries), s.Size())

	for _, want := range entries {
		got := s.Get(want.Key)

		require.Equal(t, want.Value, got.Value, "entry value")
		require.Equal(t, want.Version, got.Version, "entry version")
	}

	_ = s.Close()
}
