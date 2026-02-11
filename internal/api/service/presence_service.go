package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type PresenceService struct {
	rdb *redis.Client
	ctx context.Context
}

func NewPresenceService(rdb *redis.Client) *PresenceService {
	return &PresenceService{rdb: rdb, ctx: context.Background()}
}

// redis 온라인 키 생성
func (s *PresenceService) onlineKey(userUID uuid.UUID) string {
	return fmt.Sprintf("user:online:%s", userUID.String())
}

// 사용자 온라인 설정
func (s *PresenceService) SetOnline(userUID uuid.UUID) error {
	key := s.onlineKey(userUID)

	// 현재 시간 저장 + 5분 TTL
	now := time.Now().Unix()
	return s.rdb.Set(s.ctx, key, now, 5*time.Minute).Err()
}

// 사용자 오프라인 설정
func (s *PresenceService) SetOffline(userUID uuid.UUID) error {
	key := s.onlineKey(userUID)
	return s.rdb.Del(s.ctx, key).Err()
}

// 사용자 온라인 상태 확인
func (s *PresenceService) IsOnline(userUID uuid.UUID) (bool, error) {
	key := s.onlineKey(userUID)
	exists, err := s.rdb.Exists(s.ctx, key).Result()
	return exists > 0, err
}

// 여러 사용자 온라인 상태 조회
func (s *PresenceService) GetOnlineStatus(userUIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	result := make(map[uuid.UUID]bool)

	// pipline 으로 한번에 조회 (성능 최적화화)
	pipe := s.rdb.Pipeline()
	cmds := make(map[uuid.UUID]*redis.IntCmd)
	for _, userUID := range userUIDs {
		key := s.onlineKey(userUID)
		cmds[userUID] = pipe.Exists(s.ctx, key)
	}
	_, err := pipe.Exec(s.ctx)
	if err != nil {
		return nil, err
	}
	for userUID, cmd := range cmds {
		result[userUID] = cmd.Val() > 0
	}
	return result, nil
}

// 마지막 접속 시간 조회
func (s *PresenceService) GetLastSeen(userUID uuid.UUID) (time.Time, error) {
	key := s.onlineKey(userUID)
	val, err := s.rdb.Get(s.ctx, key).Result()
	if err == redis.Nil {
		return time.Time{}, errors.New("user not found")
	}
	if err != nil {
		return time.Time{}, err
	}

	timestamp, _ := strconv.ParseInt(val, 10, 64)
	return time.Unix(timestamp, 0), nil
}

// Heartbeat (주기적으로 호출)
func (s *PresenceService) Heartbeat(userUID uuid.UUID) error {
	return s.SetOnline(userUID)
}
