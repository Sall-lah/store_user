package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/Sall-lah/store_user/internal/client/order"
	"github.com/Sall-lah/store_user/internal/config"
	"github.com/Sall-lah/store_user/internal/db"
	"github.com/Sall-lah/store_user/internal/handler"
	"github.com/Sall-lah/store_user/internal/kafka"
	"github.com/Sall-lah/store_user/internal/ratelimit"
	"github.com/Sall-lah/store_user/internal/repository"
	"github.com/Sall-lah/store_user/internal/router"
	"github.com/Sall-lah/store_user/internal/service"
)

func TestLive_DatabaseAndRouterWorkflow(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Skipping live test: config load failed (%v)", err)
	}

	// 1. Connect to live Prisma PostgreSQL
	prismaClient := db.NewClient()
	if err := prismaClient.Prisma.Connect(); err != nil {
		t.Fatalf("Live PostgreSQL connection failed: %v", err)
	}
	defer func() {
		_ = prismaClient.Prisma.Disconnect()
	}()

	// 2. Connect to live Redis
	var limiter ratelimit.Limiter
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		t.Logf("Notice: Redis URL parse error (%v), running with mock limiter", err)
		limiter = ratelimit.NewRedisLimiter(nil, false, 50*time.Millisecond)
	} else {
		if cfg.RedisPassword != "" {
			redisOpts.Password = cfg.RedisPassword
		}
		rdb := redis.NewClient(redisOpts)
		limiter = ratelimit.NewRedisLimiter(rdb, cfg.EnableRateLimiter, 50*time.Millisecond)
	}
	defer limiter.Close()

	// 3. Assemble service with live database repository
	repo := repository.NewPrismaUserProfileRepository(prismaClient)
	orderMock := &order.MockClient{}
	kafkaMock := &kafka.MockProducer{}
	svc := service.NewUserService(repo, orderMock, kafkaMock, cfg.KafkaTopicUserEvents)
	profileHandler := handler.NewProfileHandler(svc)
	server := router.NewRouter(cfg, profileHandler, limiter)

	// 4. Generate random test user UUID
	testUserID := uuid.NewString()

	// Clean up any stale records before and after test
	_ = repo.HardDeleteByUserID(context.Background(), testUserID)
	defer func() {
		_ = repo.HardDeleteByUserID(context.Background(), testUserID)
	}()

	// STEP A: GET /api/v1/users/profile (Auto-initializes profile in live PostgreSQL)
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	req1.Header.Set("X-User-Id", testUserID)
	rec1 := httptest.NewRecorder()
	server.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("Live GET profile failed with status %d: %s", rec1.Code, rec1.Body.String())
	}

	var p1 repository.UserProfile
	if err := json.NewDecoder(rec1.Body).Decode(&p1); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}
	if p1.UserID != testUserID || p1.FullName != "User" {
		t.Errorf("Unexpected auto-initialized profile data: %+v", p1)
	}
	t.Logf("✓ Live PostgreSQL: Initial profile auto-created successfully for user %s", testUserID)

	// STEP B: PUT /api/v1/users/profile (Mutates record in live PostgreSQL)
	updatePayload := `{"fullName":"Live Test User","phoneNumber":"+628123456789","bio":"Live PostgreSQL Verification","address":"Jakarta, ID","gender":"MALE"}`
	req2 := httptest.NewRequest(http.MethodPut, "/api/v1/users/profile", bytes.NewBufferString(updatePayload))
	req2.Header.Set("X-User-Id", testUserID)
	rec2 := httptest.NewRecorder()
	server.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Live PUT profile failed with status %d: %s", rec2.Code, rec2.Body.String())
	}

	var p2 repository.UserProfile
	if err := json.NewDecoder(rec2.Body).Decode(&p2); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}
	if p2.FullName != "Live Test User" || *p2.PhoneNumber != "+628123456789" || *p2.Bio != "Live PostgreSQL Verification" {
		t.Errorf("Unexpected updated profile data in database: %+v", p2)
	}
	t.Logf("✓ Live PostgreSQL: Profile updated and persisted successfully")

	// STEP C: DELETE /api/v1/users/account (Hard-deletes from live PostgreSQL)
	req3 := httptest.NewRequest(http.MethodDelete, "/api/v1/users/account", bytes.NewBufferString(`{"reason":"live_test_cleanup"}`))
	req3.Header.Set("X-User-Id", testUserID)
	rec3 := httptest.NewRecorder()
	server.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusOK {
		t.Fatalf("Live DELETE account failed with status %d: %s", rec3.Code, rec3.Body.String())
	}

	// Verify hard-deletion in live database
	_, err = repo.GetByUserID(context.Background(), testUserID)
	if err != repository.ErrNotFound {
		t.Errorf("Expected profile to be purged from PostgreSQL, but found error: %v", err)
	}
	t.Logf("✓ Live PostgreSQL: Profile permanently hard-deleted from database")
}
