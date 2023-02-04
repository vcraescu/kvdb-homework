package server_test

import (
	"context"
	"net"
	"testing"

	v1 "emag-homework/gen/proto/go/api/v1"
	"emag-homework/internal/app"
	"emag-homework/internal/app/bootstrap"
	"emag-homework/internal/app/keyword"
	"emag-homework/internal/app/server"
	"emag-homework/pkg/log"
	"emag-homework/pkg/test/require"

	"google.golang.org/grpc"
)

func TestAppServer_Save(t *testing.T) {
	t.Parallel()

	type fields struct {
		repository app.KeywordRepository
		counter    app.KeywordCounter
	}

	type args struct {
		req *v1.SaveRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1.SaveResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				repository: keyword.NewInMemRepository(nil),
				counter:    keyword.NewCounter(),
			},
			args: args{
				req: &v1.SaveRequest{
					Text: "Lorem ipsum text lorem",
				},
			},
			want: &v1.SaveResponse{
				Keywords: map[string]int32{
					"lorem": 2,
					"ipsum": 1,
					"text":  1,
				},
			},
		},
		{
			name: "empty text",
			fields: fields{
				repository: keyword.NewInMemRepository(nil),
				counter:    keyword.NewCounter(),
			},
			args: args{
				req: &v1.SaveRequest{},
			},
			wantErr: true,
		},
	}

	logger := log.NewNopLogger()

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			srv := server.NewAppServer(tt.fields.repository, tt.fields.counter, logger)
			client, tearDown := setupTest(t, srv, logger)
			defer tearDown()

			got, err := client.Save(ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.Equal(t, tt.want.Keywords, got.Keywords)
		})
	}
}

func TestAppServer_Find(t *testing.T) {
	t.Parallel()

	type fields struct {
		repository app.KeywordRepository
	}

	type args struct {
		req *v1.FindRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1.FindResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				repository: keyword.NewInMemRepository(map[string]int{
					"lorem": 5,
					"ipsum": 2,
				}),
			},
			args: args{
				req: &v1.FindRequest{
					Keywords: []string{"lorem", "ipsum", "text", "Lorem"},
				},
			},
			want: &v1.FindResponse{
				Keywords: map[string]int32{
					"lorem": 5,
					"ipsum": 2,
				},
			},
		},
		{
			name: "no keywords",
			args: args{
				req: &v1.FindRequest{},
			},
			wantErr: true,
		},
	}

	logger := log.NewNopLogger()

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			srv := server.NewAppServer(tt.fields.repository, nil, logger)
			client, tearDown := setupTest(t, srv, logger)
			defer tearDown()

			got, err := client.Find(ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.Equal(t, tt.want.Keywords, got.Keywords)
		})
	}
}

func setupTest(t *testing.T, srv v1.AppServiceServer, logger *log.Logger) (client v1.AppServiceClient, tearDown func()) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	addr := lis.Addr().String()

	go func() {
		err := bootstrap.StartGRPCServer(ctx, lis, srv, logger)
		require.NoError(t, err)
	}()

	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)

	client = v1.NewAppServiceClient(cc)
	tearDown = func() {
		cc.Close()
		cancel()
	}

	return client, tearDown
}
