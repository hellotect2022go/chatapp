# 실시간 채팅 애플리케이션

Go 언어 기반의 확장 가능한 실시간 채팅 시스템입니다. RESTful API와 WebSocket을 활용한 마이크로서비스 아키텍처로 설계되었습니다.

## 📋 목차
- [시스템 아키텍처](#시스템-아키텍처)
- [기술 스택](#기술-스택)
- [프로젝트 구조](#프로젝트-구조)
- [핵심 기능](#핵심-기능)
- [아키텍처 패턴](#아키텍처-패턴)
- [시작하기](#시작하기)

## 🏗 시스템 아키텍처

본 프로젝트는 **마이크로서비스 아키텍처**를 채택하여 두 개의 독립적인 서버로 구성되어 있습니다:

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Layer                             │
│                    (Web Browser / Mobile)                        │
└───────────────────┬─────────────────────────┬───────────────────┘
                    │                         │
                    │ HTTP/REST               │ WebSocket
                    │                         │
        ┌───────────▼─────────┐   ┌──────────▼───────────┐
        │   API Server        │   │  WebSocket Server    │
        │   (Port: 8888)      │   │   (Port: 9999)       │
        │                     │   │                      │
        │ - User Auth         │   │ - Real-time Chat     │
        │ - Room Management   │   │ - Room Hub           │
        │ - Message CRUD      │   │ - Connection Mgmt    │
        │ - File Upload       │   │                      │
        └──────────┬──────────┘   └──────────┬───────────┘
                   │                         │
                   │         Redis Pub/Sub   │
                   └────────────┬────────────┘
                                │
                    ┌───────────▼───────────┐
                    │   Infrastructure      │
                    │                       │
                    │ - PostgreSQL (GORM)   │
                    │ - Redis (Pub/Sub)     │
                    │ - File Storage        │
                    └───────────────────────┘
```

### 서버 분리 이유

1. **관심사의 분리 (Separation of Concerns)**
   - API Server: 비즈니스 로직, 데이터 영속성
   - WebSocket Server: 실시간 통신, 연결 관리

2. **독립적인 확장성 (Independent Scalability)**
   - 각 서버를 독립적으로 스케일 아웃 가능
   - 트래픽 패턴에 따라 선택적 확장

3. **장애 격리 (Fault Isolation)**
   - 한 서버의 장애가 다른 서버에 영향 최소화
   - 서비스 안정성 향상

## 🛠 기술 스택

### Backend
- **언어**: Go 1.25.5
- **웹 프레임워크**: Gin (v1.11.0)
- **WebSocket**: Gorilla WebSocket (v1.5.3)
- **ORM**: GORM (v1.31.1)
- **데이터베이스**: PostgreSQL (pgx v5.6.0)
- **캐시/메시지 큐**: Redis (go-redis v9.17.2)
- **인증**: JWT (golang-jwt v5.3.1)
- **로깅**: Uber Zap (v1.27.1)
- **설정 관리**: godotenv (v1.5.1)

### Infrastructure
- **Database**: PostgreSQL
- **Cache & Pub/Sub**: Redis
- **File Storage**: Local File System (확장 가능)

## 📁 프로젝트 구조

```
chatapp/
├── cmd/                          # 애플리케이션 엔트리포인트
│   ├── api/                      # REST API 서버
│   │   └── main.go              # API 서버 시작점 (Port: 8888)
│   └── websocket/               # WebSocket 서버
│       └── main.go              # WebSocket 서버 시작점 (Port: 9999)
│
├── internal/                     # 내부 패키지 (외부 노출 불가)
│   ├── api/                     # API 서버 레이어
│   │   ├── dto/                 # Data Transfer Objects
│   │   │   ├── message.go       # 메시지 DTO
│   │   │   ├── room.go          # 채팅방 DTO
│   │   │   └── user.go          # 사용자 DTO
│   │   │
│   │   ├── handler/             # HTTP 핸들러 (Controller)
│   │   │   ├── 01.user_handler.go           # 사용자 관리
│   │   │   ├── 02.auth_handler.go           # 인증/인가
│   │   │   ├── 03.presence_handler.go       # 접속 상태
│   │   │   ├── 04.message_handler.go        # 메시지 CRUD
│   │   │   ├── 05.room_handler.go           # 채팅방 관리
│   │   │   ├── 06.file_handler.go           # 파일 업로드/다운로드
│   │   │   └── 99.redis_recovery_handler.go # Redis 복구
│   │   │
│   │   ├── service/             # 비즈니스 로직
│   │   │   ├── auth_service.go
│   │   │   ├── file_service.go
│   │   │   ├── message_service.go
│   │   │   ├── presence_service.go
│   │   │   ├── redis_recovery_service.go
│   │   │   ├── room_service.go
│   │   │   └── user_service.go
│   │   │
│   │   ├── repository/          # 데이터 접근 계층
│   │   │   ├── file_repository.go
│   │   │   ├── message_repository.go
│   │   │   ├── redis_session_repository.go
│   │   │   ├── room_repository.go
│   │   │   └── user_repository.go
│   │   │
│   │   ├── middleware/          # HTTP 미들웨어
│   │   │   ├── auth.go          # JWT 인증
│   │   │   ├── cors.go          # CORS 설정
│   │   │   ├── error_handler.go # 전역 에러 처리
│   │   │   ├── logger.go        # 요청 로깅
│   │   │   └── rate_limit.go    # Rate Limiting
│   │   │
│   │   └── infra/               # 인프라 설정
│   │       └── container.go     # DI 컨테이너 (의존성 주입)
│   │
│   ├── websocket/               # WebSocket 서버 레이어
│   │   ├── handler/
│   │   │   └── connection_handler.go  # WebSocket 연결 관리
│   │   │
│   │   ├── hub/                 # 실시간 통신 허브
│   │   │   ├── client.go        # 클라이언트 연결 관리
│   │   │   └── room_hub.go      # 채팅방별 브로드캐스팅
│   │   │
│   │   ├── service/             # WebSocket 비즈니스 로직
│   │   │
│   │   └── infra/               # WebSocket 인프라
│   │       └── container.go     # WebSocket DI 컨테이너
│   │
│   └── shared/                  # 공유 컴포넌트
│       ├── model/               # 도메인 모델 (엔티티)
│       │   ├── file.go          # 파일 모델
│       │   ├── message.go       # 메시지 모델
│       │   ├── room.go          # 채팅방 모델
│       │   └── user.go          # 사용자 모델
│       │
│       ├── database/            # 데이터베이스 연결
│       │   ├── db.go            # PostgreSQL 연결
│       │   └── redis.go         # Redis 연결
│       │
│       ├── pubsub/              # Redis Pub/Sub
│       │   ├── publisher.go     # 메시지 발행
│       │   └── subscriber.go    # 메시지 구독
│       │
│       ├── logger/              # 로깅
│       │   └── logger.go        # Zap 로거
│       │
│       ├── errors/              # 에러 정의
│       │   └── errors.go
│       │
│       └── util/                # 유틸리티
│           ├── file.go
│           └── token.go         # JWT 토큰 관리
│
├── client/                      # 테스트 클라이언트
│   ├── test_client.html
│   └── test_client.js
│
├── go.mod                       # Go 모듈 정의
├── go.sum                       # 의존성 체크섬
└── README.md                    # 프로젝트 문서
```

## 🎯 핵심 기능

### 1. 사용자 관리
- 회원가입 / 로그인
- JWT 기반 인증 (Access Token + Refresh Token)
- 사용자 프로필 관리
- 최근 접속 사용자 조회

### 2. 채팅방 관리
- 1:1 채팅 (Direct Message)
- 그룹 채팅 (Group Chat)
- 채팅방 생성 / 참여 / 퇴장
- 내 채팅방 목록 조회

### 3. 실시간 메시지
- WebSocket 기반 실시간 통신
- 텍스트 메시지
- 이미지 업로드 및 전송
- 입장/퇴장 알림
- 읽음 표시

### 4. 파일 관리
- 파일 업로드
- 파일 다운로드
- 이미지 미리보기

### 5. 인프라 기능
- Redis를 통한 Pub/Sub 메시지 브로드캐스팅
- Redis 데이터 복구 (DB → Redis)
- Rate Limiting
- CORS 설정
- 전역 에러 처리
- 구조화된 로깅 (Zap)

## 🏛 아키텍처 패턴

### 1. Clean Architecture (계층형 아키텍처)

본 프로젝트는 Clean Architecture 원칙을 따라 다음과 같이 계층을 분리했습니다:

```
┌─────────────────────────────────────────────┐
│           Presentation Layer                │
│         (Handler / Controller)              │
│  - HTTP 요청/응답 처리                        │
│  - WebSocket 연결 관리                        │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│          Business Logic Layer               │
│              (Service)                      │
│  - 비즈니스 규칙                              │
│  - 도메인 로직                                │
│  - 트랜잭션 관리                              │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│         Data Access Layer                   │
│           (Repository)                      │
│  - DB 쿼리                                   │
│  - 데이터 매핑                                │
│  - 캐시 관리                                  │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│          Infrastructure Layer               │
│      (Database, Redis, External API)        │
│  - PostgreSQL                               │
│  - Redis                                    │
│  - File Storage                             │
└─────────────────────────────────────────────┘
```

**계층 간 의존성 규칙:**
- 각 계층은 하위 계층에만 의존
- 상위 계층은 하위 계층을 알지 못함
- DTO를 통한 계층 간 데이터 전달

### 2. Dependency Injection (의존성 주입)

**컨테이너 패턴**을 사용하여 의존성을 관리합니다:

```go
// API Server Container
type Container struct {
    DB              *gorm.DB
    Redis           *redis.Client
    AuthHandler     *handler.AuthHandler
    RoomHandler     *handler.RoomHandler
    MessageHandler  *handler.MessageHandler
    // ...
}
```

**장점:**
- 테스트 용이성 (Mock 객체 주입 가능)
- 낮은 결합도
- 의존성 관리 중앙화

### 3. Repository Pattern

데이터 접근 로직을 추상화하여 비즈니스 로직과 분리:

```go
type UserRepository interface {
    Create(user *model.User) error
    FindByEmail(email string) (*model.User, error)
    // ...
}
```

**장점:**
- DB 교체 용이
- 테스트 용이성
- 비즈니스 로직과 데이터 접근 로직 분리

### 4. Pub/Sub Pattern (메시지 브로커)

Redis Pub/Sub을 활용한 서버 간 통신:

```
API Server                 Redis                WebSocket Server
    │                        │                         │
    │─── Publish Message ───▶│                         │
    │                        │─── Subscribe ──────────▶│
    │                        │                         │
    │                        │◀─── Broadcast ──────────│
```

**흐름:**
1. API Server가 메시지를 DB에 저장
2. Redis Pub/Sub으로 메시지 발행
3. WebSocket Server가 메시지 수신
4. 연결된 클라이언트들에게 브로드캐스트

**장점:**
- 서버 간 느슨한 결합
- 확장 가능한 아키텍처
- 실시간 이벤트 처리

### 5. Hub Pattern (WebSocket)

채팅방별 클라이언트 연결 관리:

```go
type RoomHub struct {
    rooms map[uint]map[*Client]bool  // room_id → clients
    users map[uint]*Client            // user_id → client
    Join      chan *JoinRequest
    Leave     chan *LeaveRequest
    Broadcast chan *BroadcastMessage
}
```

**특징:**
- 채널 기반 동시성 제어
- 고루틴을 통한 비동기 처리
- 방별 브로드캐스팅

### 6. Middleware Chain Pattern

요청 처리 파이프라인:

```
Request → Recovery → CORS → Logger → RateLimit → Auth → Handler
```

**적용된 미들웨어:**
- `Recovery`: Panic 복구
- `CORS`: Cross-Origin 설정
- `Logger`: 요청/응답 로깅
- `RateLimit`: API 호출 제한
- `Auth`: JWT 검증

## 🔐 보안

### 인증 방식
- **JWT (JSON Web Token)**
  - Access Token (15분 유효)
  - Refresh Token (7일 유효, DB 저장)
  - HTTP Only Cookie (선택적)

### 데이터 보호
- 비밀번호 해싱 (bcrypt)
- SQL Injection 방지 (GORM ORM)
- XSS 방지 (입력 검증)

### API 보호
- Rate Limiting (Redis 기반)
- CORS 설정
- 인증 미들웨어

## 🚀 시작하기

### 사전 요구사항
- Go 1.25.5+
- PostgreSQL
- Redis

### 환경 설정

`.env` 파일을 생성하고 다음 항목을 설정하세요:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=chatapp

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your_secret_key

# Server
API_PORT=8888
WS_PORT=9999
```

### 설치 및 실행

1. **의존성 설치**
```bash
go mod download
```

2. **API 서버 실행**
```bash
go run cmd/api/main.go
```

3. **WebSocket 서버 실행** (별도 터미널)
```bash
go run cmd/websocket/main.go
```

4. **테스트 클라이언트 실행**
```bash
# client/test_client.html을 브라우저로 열기
```

### API 엔드포인트

#### 인증
- `POST /api/v1/auth/signup` - 회원가입
- `POST /api/v1/auth/login` - 로그인
- `POST /api/v1/auth/refresh` - 토큰 갱신

#### 채팅방
- `POST /api/v1/rooms` - 채팅방 생성
- `GET /api/v1/rooms` - 내 채팅방 목록
- `GET /api/v1/rooms/:id` - 채팅방 상세
- `POST /api/v1/rooms/:id/join` - 채팅방 참여
- `DELETE /api/v1/rooms/:id/leave` - 채팅방 퇴장

#### 메시지
- `POST /api/v1/rooms/:id/messages` - 메시지 전송
- `GET /api/v1/rooms/:id/messages` - 메시지 조회
- `POST /api/v1/rooms/:id/messages/image` - 이미지 메시지 전송

#### WebSocket
- `GET /ws` - WebSocket 연결

## 📊 데이터베이스 스키마

### Users (사용자)
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    nickname VARCHAR(100) NOT NULL,
    user_role VARCHAR(100) NOT NULL,
    age INT NOT NULL,
    last_login_at TIMESTAMP,
    refresh_token TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Rooms (채팅방)
```sql
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    type VARCHAR(20) NOT NULL, -- 'direct' or 'group'
    created_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### RoomMembers (채팅방 멤버)
```sql
CREATE TABLE room_members (
    id SERIAL PRIMARY KEY,
    room_id INT NOT NULL,
    user_id INT NOT NULL,
    joined_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(room_id, user_id)
);
```

### Messages (메시지)
```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    room_id INT NOT NULL,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'text', 'image', 'file'
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Files (파일)
```sql
CREATE TABLE files (
    id SERIAL PRIMARY KEY,
    message_id INT,
    filename VARCHAR(255) NOT NULL,
    filepath VARCHAR(500) NOT NULL,
    filesize BIGINT NOT NULL,
    mimetype VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## 🔧 향후 개선 계획

### 기능 개선
- [ ] 메시지 검색 기능
- [ ] 읽지 않은 메시지 카운트
- [ ] 사용자 차단 기능
- [ ] 푸시 알림
- [ ] 음성/영상 통화

### 인프라 개선
- [ ] Docker 컨테이너화
- [ ] Kubernetes 배포
- [ ] CI/CD 파이프라인
- [ ] 모니터링 (Prometheus + Grafana)
- [ ] 로그 수집 (ELK Stack)

### 성능 개선
- [ ] 메시지 페이지네이션 최적화
- [ ] Redis 캐싱 전략 개선
- [ ] DB 인덱스 최적화
- [ ] CDN 연동 (파일 스토리지)

### 테스트
- [ ] 단위 테스트
- [ ] 통합 테스트
- [ ] 부하 테스트
- [ ] E2E 테스트

## 📝 라이센스

이 프로젝트는 MIT 라이센스를 따릅니다.

## 👥 기여

이슈와 풀 리퀘스트를 환영합니다!

---

**Made with ❤️ using Go**

