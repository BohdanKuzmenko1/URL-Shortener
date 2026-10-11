package service_test

import (
	"context"
	pb "github.com/BohdanKuzmenko1/URLShortener/proto"
	"github.com/BohdanKuzmenko1/URLShortener/services/stats-service/internal"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockStatsRepository struct {
	mock.Mock
}

func (m *MockStatsRepository) RecordClick(ctx context.Context, click internal.Click) error {
	args := m.Called(ctx, click)
	return args.Error(0)
}

func (m *MockStatsRepository) RecordClickBatch(ctx context.Context, clicks []internal.Click) error {
	args := m.Called(ctx, clicks)
	return args.Error(0)
}

func (m *MockStatsRepository) GetURLStats(ctx context.Context, URLId int32, date string) ([]internal.URLStats, error) {
	args := m.Called(ctx, URLId, date)
	return args.Get(0).([]internal.URLStats), args.Error(1)
}

type MockURLServiceClient struct {
	mock.Mock
}

func (m *MockURLServiceClient) CreateShortURL(ctx context.Context, in *pb.CreateShortUrlRequest, opts ...grpc.CallOption) (*pb.CreateShortUrlResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.CreateShortUrlResponse), args.Error(1)
}

func (m *MockURLServiceClient) ResolveSlug(ctx context.Context, in *pb.ResolveSlugRequest, opts ...grpc.CallOption) (*pb.ResolveSlugResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ResolveSlugResponse), args.Error(1)
}

func (m *MockURLServiceClient) GetURL(ctx context.Context, in *pb.GetURLRequest, opts ...grpc.CallOption) (*pb.GetURLResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.GetURLResponse), args.Error(1)
}
