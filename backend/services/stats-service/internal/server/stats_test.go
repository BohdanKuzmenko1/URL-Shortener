package server_test

import (
	"context"
	"fmt"
	pb "github.com/BohdanKuzmenko1/URLShortener/proto"
	"github.com/BohdanKuzmenko1/URLShortener/services/stats-service/internal"
	"github.com/BohdanKuzmenko1/URLShortener/services/stats-service/internal/broker"
	"github.com/BohdanKuzmenko1/URLShortener/services/stats-service/internal/server"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"os"
	"testing"
	"time"
)

type testTokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

func generateToken(t *testing.T, userId int, isExpired bool) (string, error) {
	t.Helper()

	if isExpired {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &testTokenClaims{
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(-24 * time.Hour).Unix(),
				IssuedAt:  time.Now().Add(-48 * time.Hour).Unix(),
			},
			userId,
		})

		signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SIGNING_KEY")))
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Bearer %s", signedToken), nil
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &testTokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userId,
	})

	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SIGNING_KEY")))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Bearer %s", signedToken), nil
}

type MockStatsService struct {
	mock.Mock
}

func (m *MockStatsService) RecordClick(ctx context.Context, event broker.RedirectEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockStatsService) RecordClickBatch(ctx context.Context, events []broker.RedirectEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockStatsService) GetURLStats(ctx context.Context, token string, urlID int32, date string) ([]internal.URLStats, error) {
	args := m.Called(ctx, token, urlID, date)
	return args.Get(0).([]internal.URLStats), args.Error(1)
}

func (m *MockStatsService) CheckURLOwnership(ctx context.Context, token string, urlID int32) error {
	args := m.Called(ctx, token, urlID)
	return args.Error(0)
}

func TestStatsServer_GetURLStats(t *testing.T) {
	tests := []struct {
		name           string
		userId         int
		urlID          int32
		date           string
		mockResult     []internal.URLStats
		mockError      error
		expectedError  bool
		tokenExpired   bool
		tokenExists    bool
		mockCall       bool
		metadataExists bool
	}{
		{
			name:   "Success",
			userId: 1,
			urlID:  1,
			date:   "01.01.1970",
			mockResult: []internal.URLStats{
				{
					URLId:   1,
					Device:  "desktop",
					Country: "XX",
					Date:    "01.01.1970",
					Clicks:  23,
				},
			},
			mockCall:       true,
			tokenExpired:   false,
			tokenExists:    true,
			metadataExists: true,
			expectedError:  false,
		},
		{
			name:           "NoMetadata",
			userId:         1,
			urlID:          1,
			date:           "01.01.1970",
			mockCall:       false,
			tokenExists:    false,
			metadataExists: false,
			expectedError:  true,
		},
		{
			name:           "ExpiredToken",
			userId:         1,
			urlID:          1,
			date:           "01.01.1970",
			mockCall:       false,
			tokenExists:    true,
			tokenExpired:   true,
			metadataExists: true,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(MockStatsService)
			s := server.NewStatsServer(mockService)

			var token string
			if tt.tokenExists {
				var err error
				token, err = generateToken(t, tt.userId, tt.tokenExpired)
				require.NoError(t, err, "failed to generate token")
			}

			if tt.mockCall {
				mockService.On("GetURLStats", mock.Anything, token, tt.urlID, tt.date).
					Return(tt.mockResult, tt.mockError)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if tt.metadataExists {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", token))
			}

			req := &pb.GetURLStatsRequest{
				UrlId: tt.urlID,
				Date:  tt.date,
			}

			// Action
			resp, err := s.GetURLStats(ctx, req)

			// Assert
			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Len(t, resp.UrlStats, len(tt.mockResult))

				for i, stat := range resp.UrlStats {
					require.Equal(t, tt.mockResult[i].URLId, stat.UrlId)
					require.Equal(t, tt.mockResult[i].Device, stat.Device)
					require.Equal(t, tt.mockResult[i].Country, stat.Country)
					require.Equal(t, tt.mockResult[i].Date, stat.Date)
					require.Equal(t, tt.mockResult[i].Clicks, stat.Clicks)
				}

				mockService.AssertExpectations(t)
			}
		})
	}
}
