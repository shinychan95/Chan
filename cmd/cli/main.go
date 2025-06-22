package main

import (
	"flag"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shinychan95/Chan/sync"
)

func main() {
	// flag
	configPath := flag.String("config", "config.json", "Path to config.json file")
	flag.Parse()

	log.Println("Notion Blog CLI 시작")

	// 동기화 객체 생성
	syncer, err := sync.NewBlogSyncer(*configPath)
	if err != nil {
		log.Fatalf("Config 로드 실패: %v", err)
	}

	// 동기화 실행
	log.Println("블로그 동기화 시작...")
	result := syncer.SyncToBlog()

	if result.Success {
		log.Printf("✅ 동기화 완료! (소요시간: %s)", result.Duration)
	} else {
		log.Fatalf("❌ 동기화 실패: %s (오류: %v)", result.Message, result.Error)
	}
}
