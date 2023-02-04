package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v1 "emag-homework/gen/proto/go/api/v1"
	"emag-homework/internal/app"
	"emag-homework/internal/app/keyword"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ v1.AppServiceServer = (*AppServer)(nil)

type AppServer struct {
	v1.UnimplementedAppServiceServer

	repository app.KeywordRepository
	counter    app.KeywordCounter
	logger     app.Logger
}

func NewAppServer(repository app.KeywordRepository, counter app.KeywordCounter, logger app.Logger) *AppServer {
	return &AppServer{
		repository: repository,
		counter:    counter,
		logger:     logger,
	}
}

func (s *AppServer) Save(ctx context.Context, req *v1.SaveRequest) (*v1.SaveResponse, error) {
	text := strings.TrimSpace(req.Text)

	if text == "" {
		return nil, status.Error(codes.InvalidArgument, "empty text")
	}

	keywordCounters, err := s.counter.Count(ctx, text)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &v1.SaveResponse{
		Keywords: make(map[string]int32),
	}

	for k, count := range keywordCounters {
		if err := s.repository.Increment(ctx, k, count); err != nil {
			s.logger.Error("failed to increment %q occurences", k)

			return nil, status.Error(codes.Internal, err.Error())
		}

		res.Keywords[k] = int32(count)

		s.logger.Info("keyword %q occurs %d time(s)", k, count)
	}

	return res, nil
}

func (s *AppServer) Find(ctx context.Context, req *v1.FindRequest) (*v1.FindResponse, error) {
	if len(req.Keywords) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no keywords")
	}

	find := s.newFindFunc(ctx)
	res := &v1.FindResponse{
		Keywords: make(map[string]int32),
	}

	for _, k := range req.Keywords {
		k = strings.TrimSpace(strings.ToLower(k))

		value, err := find(k)
		if err != nil {
			s.logger.Info("failed finding keyword %q: %s", k, err)

			if errors.Is(err, keyword.ErrNotFound) {
				continue
			}

			return nil, status.Error(codes.Internal, fmt.Sprintf("failed finding keyword: %s", err))
		}

		res.Keywords[k] = int32(value)
	}

	return res, nil
}

func (s *AppServer) newFindFunc(ctx context.Context) func(keyword string) (int, error) {
	cache := make(map[string]int)

	return func(keyword string) (int, error) {
		if v, ok := cache[keyword]; ok {
			return v, nil
		}

		v, err := s.repository.Find(ctx, keyword)
		if err != nil {
			return 0, err
		}

		cache[keyword] = v

		return v, nil
	}
}
