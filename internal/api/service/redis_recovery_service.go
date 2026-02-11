package service

import (
	"context"
	"fmt"
	"log"

	"github.com/hellotect2022go/chatapp/internal/api/repository"
	"github.com/redis/go-redis/v9"
)

type RedisRecoveryService interface {
	RecoverUserRoomFromDB() error
	CheckRedisHealth() (bool, error)
}

type redisRecoveryService struct {
	rdb      *redis.Client
	roomRepo repository.RoomRepository
}

func NewRedisRecoveryService(rdb *redis.Client, roomRepo repository.RoomRepository) RedisRecoveryService {
	return &redisRecoveryService{
		rdb:      rdb,
		roomRepo: roomRepo,
	}
}

func (s *redisRecoveryService) RecoverUserRoomFromDB() error {
	ctx := context.Background()

	log.Println("🔄 Starting Redis recovery from DB...")
	s.rdb.FlushAll(ctx).Err()

	// 1. 모든 방 조회
	rooms, err := s.roomRepo.FindAll()
	if err != nil {
		return fmt.Errorf("failed to fetch rooms from DB: %w", err)
	}

	log.Printf("📊 Found %d rooms in DB", len(rooms))

	// 2. Redis Pipeline 생성 (성능 최적화)
	pipe := s.rdb.Pipeline()

	totalMembers := 0

	// 3. 각 방의 멤버 정보를 Redis에 복원
	for _, room := range rooms {
		// 방의 멤버 조회
		members, err := s.roomRepo.GetMembers(room.RoomID)
		if err != nil {
			log.Printf("⚠️ Failed to get members for room %d: %v", room.RoomID, err)
			continue
		}

		if len(members) == 0 {
			log.Printf("⚠️ Room %d has no members, skipping", room.RoomID)
			continue
		}

		totalMembers += len(members)

		// room:X:members Set 생성
		// 방별 : 참여자 정보 저장
		roomMembersKey := fmt.Sprintf("room:%d:members", room.RoomID)
		for _, member := range members {
			pipe.SAdd(ctx, roomMembersKey, member.UID)
		}

		// user:X:rooms Set 생성
		// 사용자별 참여한 방 정보 저장
		for _, member := range members {
			userRoomsKey := fmt.Sprintf("user:%d:rooms", member.UID)
			pipe.SAdd(ctx, userRoomsKey, room.RoomID)
		}

		log.Printf("✅ Queued room %d (%s) with %d members", room.RoomID, room.Name, len(members))
	}

	// 4. Pipeline 실행
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("✅ Redis recovery completed!")
	log.Printf("📊 Recovered %d rooms with %d total memberships", len(rooms), totalMembers)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

// CheckRedisHealth - Redis 상태 확인
func (s *redisRecoveryService) CheckRedisHealth() (bool, error) {
	ctx := context.Background()

	// Redis 연결 확인
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return false, fmt.Errorf("redis connection failed: %w", err)
	}

	// 샘플 키 확인 (user:1:rooms 같은 키가 있는지)
	keys, err := s.rdb.Keys(ctx, "user:*:rooms").Result()
	if err != nil {
		return false, fmt.Errorf("failed to check keys: %w", err)
	}

	// 키가 하나도 없으면 복원 필요
	if len(keys) == 0 {
		log.Println("⚠️ No user:*:rooms keys found in Redis - recovery may be needed")
		return false, nil
	}

	log.Printf("✅ Redis health check passed - found %d user keys", len(keys))
	return true, nil
}
