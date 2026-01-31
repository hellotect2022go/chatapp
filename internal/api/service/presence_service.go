package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

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
func (s *PresenceService) onlineKey(userID uint) string {
	return fmt.Sprintf("user:online:%d", userID)
}

// 사용자 온라인 설정
func (s *PresenceService) SetOnline(userID uint) error {
	key := s.onlineKey(userID)

	// 현재 시간 저장 + 5분 TTL
	now := time.Now().Unix()
	return s.rdb.Set(s.ctx, key, now, 5*time.Minute).Err()
}

// 사용자 오프라인 설정
func (s *PresenceService) SetOffline(userID uint) error {
	key := s.onlineKey(userID)
	return s.rdb.Del(s.ctx, key).Err()
}

// 사용자 온라인 상태 확인
func (s *PresenceService) IsOnline(userID uint) (bool, error) {
	key := s.onlineKey(userID)
	exists, err := s.rdb.Exists(s.ctx, key).Result()
	return exists > 0, err
}

// 여러 사용자 온라인 상태 조회
func (s *PresenceService) GetOnlineStatus(userIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool)

	// pipline 으로 한번에 조회 (성능 최적화화)
	pipe := s.rdb.Pipeline()
	cmds := make(map[uint]*redis.IntCmd)
	for _, userID := range userIDs {
		key := s.onlineKey(userID)
		cmds[userID] = pipe.Exists(s.ctx, key)
	}
	_, err := pipe.Exec(s.ctx)
	if err != nil {
		return nil, err
	}
	for userID, cmd := range cmds {
		result[userID] = cmd.Val() > 0
	}
	return result, nil
}

// 마지막 접속 시간 조회
func (s *PresenceService) GetLastSeen(userID uint) (time.Time, error) {
	key := s.onlineKey(userID)
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
func (s *PresenceService) Heartbeat(userID uint) error {
	return s.SetOnline(userID)
}
