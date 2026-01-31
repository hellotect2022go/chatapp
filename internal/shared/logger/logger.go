package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger 로거 초기화
func InitLogger(mode string) {
	var config zap.Config

	if mode == "prod" {
		// Production 모드 : JSON 형식, Info 레벨이상
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder //시간 값이 사람이 읽기 힘든 숫자(Unix Epoch)에서 읽기 쉬운 문자열로 바뀝니다
	} else {
		// Development 모드
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 로그 파일 출력 추가 (선택사항)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	//AddCallerSkip(1)은 Zap에게 이렇게 명령하는 것입니다. "야, 로그 위치 기록할 때 지금 위치에서 위로 한 단계(Stack Frame) 건너뛰고 기록해!"
	//Step 0: zap 내부 엔진
	//Step 1: logger.Info() (사용자님이 만든 함수) → 여기를 건너뜁니다!
	//Step 2: main.go의 20번째 줄 (실제 호출한 곳) → 여기를 기록합니다.

	var err error
	Log, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

// Sync - 로그 버퍼 플러시 (main 종료 전 호출)
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// ============================================================
// 편의 함수들 (전역 logger 사용)
// ============================================================

// Debug - 디버그 로그
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Info - 정보 로그
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Warn - 경로 로그
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Error - 에러로그
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal - 치명적 에러 로그 (프로그램 종료)
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// With - 필드를 추가한 새로운 logger 를 반환
func With(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

// ============================================================
// 환경변수 기반 모드 결정
// ============================================================

func GetMode() string {
	mode := os.Getenv("APP_MODE")
	if mode == "" {
		mode = "development"
	}
	return mode
}
