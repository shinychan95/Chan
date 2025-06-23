package notion

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPageGetMetaStringWithColonInTitle(t *testing.T) {
	// Arrange
	page := Page{
		ID:         "test-id",
		Title:      "Hello: World - A Test Title",
		Status:     "Published",
		Path:       "hello-world",
		Author:     "chanyoung.kim",
		Categories: []string{"Test", "Example"},
		Tags:       []string{"test", "example"},
		Published:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	// Act
	result := page.GetMetaString()

	// Assert
	assert.Contains(t, result, `title: "Hello: World - A Test Title"`)
	assert.Contains(t, result, "author: chanyoung.kim")
	assert.Contains(t, result, "date: 2024-01-15 10:30:00 +0000")
	assert.Contains(t, result, "categories: [Test, Example]")
	assert.Contains(t, result, "tags: [test, example]")

	// YAML 구조 확인
	lines := strings.Split(result, "\n")
	assert.Equal(t, "---", lines[0])
	assert.Equal(t, "---", lines[len(lines)-2]) // 마지막 줄은 빈 문자열
}

func TestPageGetMetaStringWithQuotesInTitle(t *testing.T) {
	// Arrange
	page := Page{
		ID:         "test-id",
		Title:      `He said "Hello World" to me`,
		Status:     "Published",
		Path:       "hello-world",
		Author:     "chanyoung.kim",
		Categories: []string{},
		Tags:       []string{},
		Published:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	// Act
	result := page.GetMetaString()

	// Assert
	assert.Contains(t, result, `title: "He said \"Hello World\" to me"`)
}

func TestPageGetMetaStringWithSimpleTitle(t *testing.T) {
	// Arrange
	page := Page{
		ID:         "test-id",
		Title:      "Simple Title",
		Status:     "Published",
		Path:       "simple-title",
		Author:     "chanyoung.kim",
		Categories: []string{},
		Tags:       []string{},
		Published:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	// Act
	result := page.GetMetaString()

	// Assert
	// 특수문자가 없으면 따옴표 없이 출력
	assert.Contains(t, result, "title: Simple Title")
}

func TestParsePagePropertiesWithDefaults(t *testing.T) {
	// Arrange
	ResetPostCounter() // 카운터 리셋

	page := &Page{}
	rawProperties := `{
		"title": [["Test Title"]],
		"status": [["Published"]],
		"categories": [["Test,Example"]],
		"tags": [["test,example"]]
	}`

	schema := map[string]Schema{
		"title":      {Name: "Title"},
		"status":     {Name: "Status"},
		"categories": {Name: "Categories"},
		"tags":       {Name: "Tags"},
	}

	// Act
	parsePageProperties(page, rawProperties, schema)

	// Assert
	assert.Equal(t, "Test Title", page.Title)
	assert.Equal(t, "Published", page.Status)
	assert.Equal(t, []string{"Test", "Example"}, page.Categories)
	assert.Equal(t, []string{"test", "example"}, page.Tags)
	assert.Equal(t, "chanyoung.kim", page.Author)

	// 기본값 확인
	assert.NotZero(t, page.Published)    // 현재 시간이 설정되어야 함
	assert.Equal(t, "post-1", page.Path) // 첫 번째 글 번호
}

func TestParsePagePropertiesWithCustomPath(t *testing.T) {
	// Arrange
	ResetPostCounter() // 카운터 리셋

	page := &Page{}
	rawProperties := `{
		"title": [["Test Title"]],
		"path": [["custom-path"]],
		"status": [["Published"]]
	}`

	schema := map[string]Schema{
		"title":  {Name: "Title"},
		"path":   {Name: "Path"},
		"status": {Name: "Status"},
	}

	// Act
	parsePageProperties(page, rawProperties, schema)

	// Assert
	assert.Equal(t, "Test Title", page.Title)
	assert.Equal(t, "custom-path", page.Path) // 사용자 정의 경로 사용
	assert.Equal(t, "Published", page.Status)
}

func TestParsePagePropertiesWithEmptyPath(t *testing.T) {
	// Arrange
	ResetPostCounter() // 카운터 리셋

	page := &Page{}
	rawProperties := `{
		"title": [["Test Title"]],
		"path": [[""]],
		"status": [["Published"]]
	}`

	schema := map[string]Schema{
		"title":  {Name: "Title"},
		"path":   {Name: "Path"},
		"status": {Name: "Status"},
	}

	// Act
	parsePageProperties(page, rawProperties, schema)

	// Assert
	assert.Equal(t, "Test Title", page.Title)
	assert.Equal(t, "post-1", page.Path) // 빈 경로일 때 기본값 사용
	assert.Equal(t, "Published", page.Status)
}

func TestPostCounterIncrement(t *testing.T) {
	// Arrange
	ResetPostCounter() // 카운터 리셋

	page1 := &Page{}
	page2 := &Page{}
	page3 := &Page{}

	rawProperties := `{
		"title": [["Test Title"]],
		"status": [["Published"]]
	}`

	schema := map[string]Schema{
		"title":  {Name: "Title"},
		"status": {Name: "Status"},
	}

	// Act
	parsePageProperties(page1, rawProperties, schema)
	parsePageProperties(page2, rawProperties, schema)
	parsePageProperties(page3, rawProperties, schema)

	// Assert
	assert.Equal(t, "post-1", page1.Path)
	assert.Equal(t, "post-2", page2.Path)
	assert.Equal(t, "post-3", page3.Path)
	assert.Equal(t, int64(3), GetPostCounter())
}
