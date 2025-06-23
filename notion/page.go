package notion

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shinychan95/Chan/utils"
)

// 글로벌 카운터 (글 번호 생성용)
var postCounter int64 = 0

// ResetPostCounter 카운터를 리셋합니다 (테스트용)
func ResetPostCounter() {
	atomic.StoreInt64(&postCounter, 0)
}

// GetPostCounter 현재 카운터 값을 반환합니다 (테스트용)
func GetPostCounter() int64 {
	return atomic.LoadInt64(&postCounter)
}

type Page struct {
	ID         string
	Title      string
	Status     string
	Path       string
	Author     string
	Categories []string
	Tags       []string
	Published  time.Time
}

// escapeYAMLString YAML에서 특수문자를 적절히 이스케이프합니다
func escapeYAMLString(s string) string {
	// 이미 따옴표로 감싸져 있거나 특수문자가 없는 경우 그대로 반환
	if !strings.ContainsAny(s, ":\"'\\\n\r\t") && !strings.HasPrefix(s, " ") && !strings.HasSuffix(s, " ") {
		return s
	}

	// 따옴표를 이스케이프하고 전체를 큰따옴표로 감싸기
	escaped := strings.ReplaceAll(s, `"`, `\"`)
	return fmt.Sprintf(`"%s"`, escaped)
}

func (pg *Page) GetMetaString() string {
	var sb strings.Builder

	sb.WriteString("---\n")

	// title을 YAML에 안전하게 삽입
	sb.WriteString(fmt.Sprintf("title: %s\n", escapeYAMLString(pg.Title)))
	sb.WriteString(fmt.Sprintf("author: %s\n", pg.Author))
	sb.WriteString(fmt.Sprintf("date: %s\n", pg.Published.Format("2006-01-02 15:04:05 -0700")))

	sb.WriteString("categories: [" + utils.SliceToString(pg.Categories, nil) + "]\n")
	sb.WriteString("tags: [" + utils.SliceToString(pg.Tags, strings.ToLower) + "]\n")

	sb.WriteString("---\n")

	return sb.String()
}

func handlePage(page Page, wg *sync.WaitGroup, errCh chan error) {
	fmt.Println("Page title:", page.Path)

	if page.ID == "1519a0a9-70f1-444e-95b4-f6e6fac46131" {
		fmt.Println("")
	}

	// page block 하위 모든 block parsing
	pageBlock := getBlockData(page.ID)
	parseChildBlocks(&pageBlock)
	setNumberedListValue(&pageBlock.Children)

	//////////////////////
	// markdown 결과 출력 //
	//////////////////////

	var markdownOutput string

	// 내부 헤더
	markdownOutput += page.GetMetaString() + "\n"

	// Table of Contents 생성을 위해 헤더 정보 수집
	headers := CollectHeaders(pageBlock.Children)

	// 내부 컨텐츠
	for _, block := range pageBlock.Children {
		markdownOutput += ParseBlock(page.ID, block, 0, headers, wg, errCh)
	}

	if _, err := os.Stat(PostDir); os.IsNotExist(err) {
		os.MkdirAll(PostDir, os.ModePerm)
	}

	datePrefix := page.Published.Format("2006-01-02")
	markdownFileName := fmt.Sprintf("%s-%s.md", datePrefix, utils.SanitizeFileName(page.Path))
	markdownFilePath := filepath.Join(PostDir, "", markdownFileName)

	err := ioutil.WriteFile(markdownFilePath, []byte(markdownOutput), 0644)
	utils.CheckError(err)

	log.Printf("📄 Page saved: %s (%s)", page.Title, markdownFilePath)
}

func parsePageProperties(page *Page, rawProperties string, schema map[string]Schema) {
	var propertiesMap map[string][][]interface{}
	err := json.Unmarshal([]byte(rawProperties), &propertiesMap)
	utils.CheckError(err)

	// INFO - Author 의 경우, static 하게 입력한다.
	page.Author = "chanyoung.kim"

	// 기본값 설정
	page.Published = time.Now()                                          // 현재 시간을 기본값으로 설정
	page.Path = fmt.Sprintf("post-%d", atomic.AddInt64(&postCounter, 1)) // 글 번호를 기본값으로 설정

	for key, value := range propertiesMap {
		schemaInfo := schema[key]
		propertyValue := value[0]

		switch schemaInfo.Name {
		case "Categories":
			// `block` 테이블의 `properties`에는 옵션의 '값'이 쉼표로 구분된 문자열로 저장되어 있습니다.
			// 예: [["Value1,Value2"]]
			page.Categories = strings.Split(propertyValue[0].(string), ",")
		case "Tags":
			page.Tags = strings.Split(propertyValue[0].(string), ",")
		case "Status":
			// `block` 테이블의 `properties`에는 상태의 '값'이 직접 저장되어 있습니다.
			// 예: [["Archived"]]
			page.Status = propertyValue[0].(string)
		case "Title":
			page.Title = propertyValue[0].(string)
		case "Path":
			// Path가 비어있지 않은 경우에만 설정
			if pathValue := propertyValue[0].(string); pathValue != "" {
				page.Path = pathValue
			}
		case "Published":
			dateProperty := propertyValue[1].([]interface{})[0].([]interface{})[1].(map[string]interface{})
			dateString := dateProperty["start_date"].(string)
			timeString, ok := dateProperty["start_time"]
			if !ok {
				timeString = "00:00"
			}
			dateTime := dateString + "T" + timeString.(string) + ":00"
			location, _ := time.LoadLocation("Asia/Seoul")
			page.Published, err = time.ParseInLocation("2006-01-02T15:04:05", dateTime, location)
			utils.CheckError(err)
		}
	}

	return
}
