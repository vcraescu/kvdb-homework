package server_test

import (
	"context"
	v1 "emag-homework/internal/db/api/v1"
	"emag-homework/internal/db/bootstrap"
	"emag-homework/internal/db/node"
	"emag-homework/internal/db/node/server"
	"emag-homework/internal/db/store"
	"emag-homework/pkg/log"
	"emag-homework/pkg/test/require"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"testing"
)

func TestNodeServer_Put(t *testing.T) {
	t.Parallel()

	type fields struct {
		store node.Store
	}

	type args struct {
		req *v1.PutRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1.GetResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				store: func() node.Store {
					t.Helper()

					s, err := store.New()
					require.NoError(t, err)

					return s
				}(),
			},
			args: args{
				req: &v1.PutRequest{
					Key:     "key-1",
					Value:   []byte(fmt.Sprint(1)),
					Version: 1,
				},
			},
			want: &v1.GetResponse{
				Value:   []byte(fmt.Sprint(1)),
				Version: 1,
			},
		},
	}

	logger := log.NewNopLogger()

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			srv := server.NewNodeServer(tt.fields.store, "100")
			client, tearDown := setupTest(t, srv, logger)
			defer tearDown()

			_, err := client.Put(ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			got, err := client.Get(ctx, &v1.GetRequest{Key: tt.args.req.Key})

			require.NoError(t, err)
			require.Equal(t, tt.want.Value, got.Value, "value")
			require.Equal(t, tt.want.Version, got.Version, "version")
		})
	}
}

func setupTest(t *testing.T, srv v1.NodeServer, logger node.Logger) (client v1.NodeClient, tearDown func()) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	addr := lis.Addr().String()

	go func() {
		err := bootstrap.StartNodeGRPCServer(ctx, lis, srv, logger)
		require.NoError(t, err)
	}()

	cc, err := grpc.Dial(addr, grpc.WithInsecure())

	client = v1.NewNodeClient(cc)
	tearDown = func() {
		cc.Close()
		cancel()
	}

	return client, tearDown
}
