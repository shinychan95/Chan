# Chan Fyne 앱
APP_NAME = Chan
APP_ID = com.shinychan95.Chan
VERSION = 1.0.0
BUILD = 1

# 빌드 디렉토리
BUILD_DIR = ./build

.PHONY: all build build-app install-app run-app run-binary clean deps help

# 기본 타겟
all: build-app

# 바이너리 빌드
build:
	@echo "🦤 바이너리 빌드 중..."
	go build -o $(APP_NAME) .
	@echo "✅ 바이너리 빌드 완료: $(APP_NAME)"

# macOS 앱 번들 생성
build-app: build
	@echo "🦤 Fyne macOS 앱 번들 생성 중..."
	@mkdir -p $(BUILD_DIR)
	"$(shell go env GOPATH)/bin/fyne" package -os darwin -icon assets/dodo.png .
	@echo "🔧 생성된 .app 파일을 $(BUILD_DIR)로 이동 중..."
	@mv "Chan.app" "$(BUILD_DIR)/$(APP_NAME).app"
	@echo "✅ macOS 앱 번들 생성 완료: $(BUILD_DIR)/$(APP_NAME).app"

# 앱 번들 설치 (Applications 폴더로 복사)
install-app: build-app
	@echo "🚀 Applications 폴더에 설치 중..."
	@cp -R "$(BUILD_DIR)/$(APP_NAME).app" "/Applications/"
	@echo "✅ 설치 완료! Applications 폴더에서 '$(APP_NAME)' 앱을 실행하세요."

# 앱 번들 실행
run-app: build-app
	@echo "🦤 macOS 앱 번들 실행 중..."
	@open "$(BUILD_DIR)/$(APP_NAME).app"

# 바이너리 직접 실행
run-binary: build
	@echo "🦤 바이너리 직접 실행 중..."
	@./$(APP_NAME)

# 의존성 설치
deps:
	@echo "Go 의존성 설치 중..."
	go mod tidy
	go mod download
	@echo "Fyne CLI 도구 설치 중..."
	go install fyne.io/tools/cmd/fyne@latest

# 테스트 실행
test:
	go test ./...

# 빌드 파일 정리
clean:
	@echo "빌드 파일 정리 중..."
	rm -rf $(BUILD_DIR)
	rm -f $(APP_NAME)
	rm -f $(APP_NAME).app
	@echo "✅ 정리 완료"

# 도움말
help:
	@echo "🦤 Chan 빌드 도구"
	@echo ""
	@echo "사용 가능한 명령어:"
	@echo "  make build        - 바이너리 빌드"
	@echo "  make build-app    - macOS 앱 번들 생성 (LSUIElement 포함)"
	@echo "  make install-app  - 앱 번들을 Applications에 설치"
	@echo "  make run-app      - 앱 번들 실행"
	@echo "  make run-binary   - 바이너리 직접 실행"
	@echo "  make deps         - 의존성 및 도구 설치"
	@echo "  make test         - 테스트 실행"
	@echo "  make clean        - 빌드 파일 정리"
	@echo "  make help         - 이 도움말 표시"
	@echo ""
	@echo "ℹ️  빠른 시작: make deps && make run-app" 