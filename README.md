# Chan

Notion에서 글을 작성하고, 원터치로 블로그에 배포하는 macOS 네이티브 앱입니다.

Notion의 다양한 블록 타입(표, 이미지, callout, toggle, 컬럼 레이아웃, 목차 등)을 완벽하게 지원하여 복잡한 레이아웃도 그대로 블로그에 옮길 수 있습니다.

**Example:**
- [Notion 페이지](https://shinychan95.notion.site/Notion-1eafdee6189c46fea0fe5bff83f07309)
- [블로그 페이지](https://shinychan95.github.io/posts/Notion-글,-버튼-하나로-블로그-배포-가능/)

## 🎯 주요 기능

### ✨ 완전한 macOS 네이티브 앱
- **메뉴바 상주**: 🦤 도도새 아이콘으로 항상 접근 가능
- **원클릭 동기화**: 메뉴에서 바로 블로그 동기화 실행
- **실시간 상태 표시**: 동기화 진행상황과 결과를 실시간 확인
- **macOS 알림**: 동기화 완료/실패 시 시스템 알림
- **GUI 설정**: 직관적인 설정 마법사와 폴더 선택 대화상자
- **자동 시작**: macOS 로그인 시 자동 실행 옵션

### 📝 강력한 Notion 블록 지원
- **기본 블록**: 헤더, 텍스트, 리스트, 인용문, 코드 블록
- **고급 블록**: 표, 이미지, callout, toggle, 체크박스
- **레이아웃 블록**: 컬럼 레이아웃 (가로 배치)
- **목차 자동 생성**: Table of Contents 블록 완벽 지원
- **북마크**: 외부 링크를 깔끔한 카드 형태로 변환

### 🔗 Jekyll/GitHub Pages 최적화
- **Jekyll 호환**: Chirpy 테마를 포함한 Jekyll 기반 블로그 완벽 지원
- **자동 앵커 링크**: 목차에서 헤더로 바로 이동 가능
- **GitHub 토큰 인증**: 안전한 자동 배포
- **이미지 최적화**: Notion 이미지를 로컬로 다운로드하여 빠른 로딩

## 🚀 빠른 시작

### 시스템 요구사항

- **macOS**: 10.15 (Catalina) 이상
- **Go**: 1.24.4 이상 (권장)
- **Xcode Command Line Tools**: 최신 버전
- **Notion 앱**: 데스크톱 버전 설치 필요

> ⚠️ **중요**: Go 1.21.10 이하 버전에서는 트레이 아이콘 표시 문제가 발생할 수 있습니다. Go 1.24.4 이상 사용을 강력히 권장합니다.

### 1. 설치

```bash
# 저장소 클론
git clone https://github.com/shinychan95/Chan.git
cd Chan

# 의존성 설치 및 앱 빌드
make install

# Applications 폴더에 설치
make install-app
```

### 2. 앱 실행

```bash
# 방법 1: Applications 폴더에서 실행
# Finder > Applications > "Chan" 더블클릭

# 방법 2: 터미널에서 실행
make run-app
```

### 3. 초기 설정

1. **메뉴바 아이콘 확인**: 🦤 아이콘이 메뉴바에 표시됩니다
2. **설정 마법사 실행**: 아이콘 클릭 → "🛠️ 설정 마법사"
3. **Notion 설정**:
   - Notion API Key 입력
   - Collection View ID 입력 (Notion 데이터베이스 ID)
   - Notion DB 경로 선택
4. **블로그 설정**:
   - GitHub 토큰 입력 (repo 권한 필요)
   - GitHub 저장소 경로 (예: `username/username.github.io`)
   - 로컬 블로그 폴더 선택

### 4. Notion 페이지 준비

1. **데이터베이스 생성**: Notion에서 "블로그 포스팅 캘린더" 템플릿 사용
2. **필수 속성 설정**:
   - `Status`: "Drafting" 상태인 페이지만 동기화됨
   - `Title`: 블로그 포스트 제목
   - `Path`: URL 경로 (선택사항)
   - `Published`: 발행 날짜
   - `Categories`: 카테고리 (쉼표로 구분)
   - `Tags`: 태그 (쉼표로 구분)
3. **Integration 연결**: 
   - [Notion Integrations](https://www.notion.so/my-integrations)에서 새 Integration 생성
   - 데이터베이스 페이지에서 Integration 연결

### 5. 동기화 실행

- 메뉴바 🦤 클릭 → "🦤 동기화" 선택
- 완료 알림까지 기다리기

## 📱 메뉴바 인터페이스

### 상태 표시

| 툴팁 | 상태 | 설명 |
|------|------|------|
| 🦤 동기화 대기 중 | 준비 완료 | 설정 완료, 동기화 가능 |
| ⚠️ 설정이 필요합니다 | 설정 미완료 | 설정 마법사 실행 필요 |
| ⏳ 동기화 중: [진행상황] | 진행 중 | 동기화 실행 중 |
| ✅ 마지막 동기화: N분 전 | 성공 | 마지막 동기화 성공 |
| ❌ 동기화 실패: [오류메시지] | 실패 | 동기화 실패 |

### 메뉴 기능

| 메뉴 항목 | 기능 | 설명 |
|-----------|------|------|
| 🦤 동기화 | 블로그 동기화 실행 | Notion → GitHub Pages |
| 📊 상태 보기 | 현재 상태 확인 | 설정/동기화 상태 표시 |
| 🛠️ 설정 마법사 | 설정 가이드 | 단계별 설정 안내 |
| ⚙️ 설정 편집 | config.json 편집 | 텍스트 에디터로 설정 편집 |
| 🔄 설정 새로고침 | 설정 다시 로드 | 편집한 설정 적용 |
| 📝 로그 보기 | 오류 로그 확인 | 동기화 오류 진단 |
| 🚀 자동 시작 | 로그인 시 자동 실행 | macOS 시작 시 앱 자동 실행 |
| ℹ️ 정보 | 앱 정보 | 버전, 개발자 정보 |
| 🔴 종료 | 앱 종료 | 메뉴바에서 제거 |

## 🛠️ 개발 및 빌드

### 빌드 명령어

```bash
# 전체 설치 (의존성 + 빌드 + 설치)
make install

# macOS 앱 번들 생성
make build-app

# Applications 폴더에 설치
make install-app

# 앱 번들 실행
make run-app

# CLI 버전 빌드
make build-cli

# 개발용 바이너리 실행
make run-binary

# 빌드 파일 정리
make clean

# 도움말
make help
```

### 프로젝트 구조

```
Chan/
├── main.go              # 메뉴바 앱 메인
├── cmd/cli/main.go      # CLI 버전 메인
├── config/              # 설정 관리
├── sync/                # 동기화 로직
├── notion/              # Notion 데이터 파싱
├── markdown/            # 마크다운 생성
├── utils/               # 유틸리티 함수
├── assets/              # 앱 리소스
└── Makefile            # 빌드 스크립트
```

## 📋 설정 파일

앱은 macOS 표준 위치에 설정을 저장합니다:

```
~/Library/Application Support/Chan/config.json
```

**설정 예시:**
```json
{
  "db_path": "/Users/user/Library/Application Support/Notion/notion.db",
  "api_key": "secret_...",
  "post_directory": "/Users/user/github/username.github.io/_posts",
  "image_directory": "/Users/user/github/username.github.io/assets/pages",
  "root_id": "806a2a5d-dce8-4729-916a-387f939bc82b",
  "github_token": "ghp_...",
  "github_repo": "username/username.github.io"
}
```

## 💡 기술적 특징

### 혁신적인 접근 방식
- **Notion 캐시 DB 활용**: Notion 앱의 SQLite 캐시를 직접 파싱하여 API 제한 없이 빠른 동기화
- **리버스 엔지니어링**: Notion의 내부 블록 구조를 분석하여 완벽한 마크다운 변환
- **Jekyll 최적화**: Jekyll/Kramdown의 ID 생성 규칙에 맞춘 앵커 링크 생성

### 고급 기능 구현
- **컬럼 레이아웃**: HTML 테이블을 사용한 가로 배치 구현
- **목차 자동 생성**: 페이지 헤더를 수집하여 클릭 가능한 목차 생성
- **이미지 처리**: Notion API를 통한 이미지 다운로드 및 로컬 저장
- **GitHub 자동 배포**: 토큰 인증을 통한 안전한 자동 커밋/푸시

### 안정성 보장
- **에러 복구**: Panic 상황에서도 앱이 종료되지 않는 안전한 에러 처리
- **대용량 파일 지원**: Git HTTP 버퍼 크기 조정으로 큰 이미지도 안정적으로 처리
- **설정 마이그레이션**: 기존 설정 자동 감지 및 새 위치로 이전

## 🎯 지원되는 Notion 블록

### ✅ 완벽 지원
- **텍스트**: 제목, 본문, 서식 (굵게, 기울임, 밑줄, 취소선, 코드)
- **헤더**: H1, H2, H3 (자동 앵커 링크 생성)
- **리스트**: 순서 있는 목록, 순서 없는 목록
- **인용문**: 블록 인용
- **코드**: 인라인 코드, 코드 블록 (언어별 하이라이팅)
- **구분선**: 수평선
- **체크박스**: To-do 리스트
- **이미지**: 자동 다운로드 및 로컬 저장
- **표**: 완전한 테이블 변환
- **Callout**: 강조 박스
- **Toggle**: 접을 수 있는 콘텐츠
- **컬럼**: 가로 레이아웃 (HTML 테이블 사용)
- **목차**: Table of Contents (자동 헤더 링크)
- **북마크**: 외부 링크 카드

### 🔄 향후 지원 예정
- **데이터베이스 임베드**: 다른 데이터베이스 내용 포함
- **동영상**: 비디오 파일 임베드
- **오디오**: 오디오 파일 임베드
- **PDF**: PDF 파일 임베드

## ⚠️ 제약 사항

1. **macOS 전용**: 현재 메뉴바 앱은 macOS만 지원 (CLI 버전은 크로스 플랫폼)
2. **Jekyll 최적화**: Jekyll 기반 블로그에 최적화됨 (다른 SSG는 추가 설정 필요)
3. **Notion Integration 필요**: 이미지 다운로드를 위해 Notion API 연동 필수
4. **특정 템플릿 의존**: "블로그 포스팅 캘린더" 템플릿의 속성 구조에 의존
5. **Status 필터**: "Drafting" 상태의 페이지만 동기화됨

## 🔧 문제 해결

### Go 버전 관련 문제

#### 트레이 아이콘이 표시되지 않거나 "Killed: 9" 오류 발생
**원인**: Go 1.21.10 이하 버전과 CGO/Objective-C 호환성 문제

**해결 방법**:
```bash
# Homebrew로 Go 업그레이드
brew install go

# PATH 업데이트 (현재 세션)
export PATH="/opt/homebrew/bin:$PATH"

# 영구 적용
echo 'export PATH="/opt/homebrew/bin:$PATH"' >> ~/.zshrc

# 버전 확인
go version  # go1.24.4 이상이어야 함

# 프로젝트 다시 빌드
make clean && make install
```

#### `-lobjc` 링커 경고 (Xcode 15+)
**증상**: `ld: warning: ignoring duplicate libraries: '-lobjc'`
**원인**: Xcode 15의 새로운 링커가 중복 라이브러리에 대해 경고 표시
**상태**: 기능상 문제없음, 무시해도 됨

**경고 숨기기** (선택사항):
```bash
# 환경변수로 설정
export CGO_LDFLAGS="-Wl,-no_warn_duplicate_libraries"
make build

# 또는 Makefile 수정
CGO_LDFLAGS += -Wl,-no_warn_duplicate_libraries
```

### 메뉴바 아이콘이 보이지 않는 경우
1. **메뉴바 정리**: 불필요한 메뉴바 앱 제거
2. **확장 도구 사용**: Bartender, Hidden Bar, Dozer 등 사용
3. **시스템 재시작**: macOS 메뉴바 새로고침

### 동기화 실패 시
1. **로그 확인**: 메뉴 → "📝 로그 보기"
2. **설정 검증**: GitHub 토큰, Notion API 키 확인
3. **네트워크 확인**: 인터넷 연결 상태 점검
4. **권한 확인**: Notion Integration 연결 상태 확인

### 이미지가 표시되지 않는 경우
1. **경로 확인**: 이미지 디렉토리 설정 검증
2. **권한 확인**: 폴더 쓰기 권한 확인
3. **용량 확인**: 디스크 공간 부족 여부 점검

## 🚀 로드맵

### v2.0 (진행 중)
- [x] 메뉴바 네이티브 앱
- [x] GUI 설정 마법사
- [x] 컬럼 레이아웃 지원
- [x] 목차 자동 생성
- [x] GitHub 토큰 인증

### v2.1 (계획)
- [ ] Windows/Linux 지원
- [ ] 웹 인터페이스
- [ ] 자동 백업 기능
- [ ] 배치 동기화

### v3.0 (미래)
- [ ] 다중 블로그 지원
- [ ] 커스텀 템플릿
- [ ] 플러그인 시스템
- [ ] 클라우드 동기화

## 🤝 기여하기

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 라이선스

이 프로젝트는 MIT 라이선스 하에 배포됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 👨‍💻 개발자

**Chanyoung Kim** - [@shinychan95](https://github.com/shinychan95)

프로젝트 링크: [https://github.com/shinychan95/Chan](https://github.com/shinychan95/Chan)

---

**Chan으로 더 쉽고 빠른 블로그 포스팅을 경험해보세요! 🦤**
