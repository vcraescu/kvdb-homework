package healthz_test

import (
	"context"
	v1 "emag-homework/internal/db/api/v1"
	"emag-homework/internal/db/controller/healthz"
	"emag-homework/pkg/test/require"
	"testing"
	"time"
)

func TestChecker_Add(t *testing.T) {
	t.Parallel()

	type fields struct {
		checkInterval time.Duration
		checkTimeout  time.Duration
	}

	type args struct {
		id string
		fn healthz.CheckFunc
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1.HealthzResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				checkInterval: time.Millisecond * 100,
				checkTimeout:  time.Second * 5,
			},
			args: args{
				id: "foobar",
				fn: func(ctx context.Context, req *v1.HealthzRequest) (*v1.HealthzResponse, error) {
					return &v1.HealthzResponse{
						Code: v1.HealthzResponse_HEALTHZ_OK,
						Id:   "foobar",
					}, nil
				},
			},
			want: &v1.HealthzResponse{
				Code: v1.HealthzResponse_HEALTHZ_OK,
				Id:   "foobar",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := healthz.NewChecker(
				healthz.WithCheckInterval(tt.fields.checkInterval),
				healthz.WithCheckTimeout(tt.fields.checkTimeout),
			)

			go c.Start()
			defer c.Stop()

			c.Add(tt.args.id, tt.args.fn)

			safeStopper := time.NewTimer(time.Second)

			times := int(time.Second/tt.fields.checkInterval) / 2
			var called int

			for i := 0; i < times; i++ {
				called++

				select {
				case <-safeStopper.C:
					require.True(t, false, "safe stop")
				case got := <-c.Events():
					if tt.wantErr {
						require.Error(t, got.Err, "want err")

						return
					}

					require.NoError(t, got.Err)
					require.Equal(t, tt.want.Id, got.Res.Id, "ID")
					require.Equal(t, tt.want.Code, got.Res.Code, "Code")
				}
			}

			require.True(t, called > 0, "must be called at least once")
			require.Equal(t, times, called, "calls")
		})
	}
}
