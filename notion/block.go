package notion

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/shinychan95/Chan/markdown"
	"github.com/shinychan95/Chan/utils"
)

type Block struct {
	ID         string
	Type       string
	Number     uint8
	ParsedProp ParsedProp
	Content    sql.NullString
	Children   []Block
	Properties sql.NullString
	Format     sql.NullString
	Table      *Table
}

type ParsedProp struct {
	Title    string
	Language string // for code type
}

func parseChildBlocks(block *Block) {
	childIDs, err := extractChildIDs(block.Content)
	utils.CheckError(err)

	for _, childID := range childIDs {
		childBlock := getBlockData(childID)
		parseChildBlocks(&childBlock)

		block.Children = append(block.Children, childBlock)
	}

	return
}

func setNumberedListValue(blocks *[]Block) {
	var currentNumber uint8 = 1

	for i := range *blocks {
		if (*blocks)[i].Type == "numbered_list" {
			(*blocks)[i].Number = currentNumber
			currentNumber++

			setNumberedListValue(&((*blocks)[i].Children))
		} else {
			currentNumber = 1
		}
	}
}

// extractChildIDs 함수 추가
func extractChildIDs(content sql.NullString) (childIDs []string, err error) {
	if !content.Valid {
		return
	}

	err = json.Unmarshal([]byte(content.String), &childIDs)
	utils.CheckError(err)

	return
}

/////////////////////////////////////////
// Block 자체 그리고 Type 에 따른 Parse 로직 //
/////////////////////////////////////////

type HeaderInfo struct {
	Title  string
	Anchor string
	Level  int
}

func CollectHeaders(blocks []Block) []HeaderInfo {
	var headers []HeaderInfo
	for _, block := range blocks {
		level := 0
		switch block.Type {
		case "header":
			level = 1
		case "sub_header":
			level = 2
		case "sub_sub_header":
			level = 3
		}

		if level > 0 {
			title := ParsePropTitle(block.Properties.String)
			anchor := utils.SanitizeFileName(title)

			headers = append(headers, HeaderInfo{
				Title:  title,
				Anchor: anchor,
				Level:  level,
			})
		}

		// 재귀적으로 자식 블록 탐색
		if len(block.Children) > 0 {
			headers = append(headers, CollectHeaders(block.Children)...)
		}
	}
	return headers
}

func ParseBlock(pageID string, block Block, indentLv int, headers []HeaderInfo, wg *sync.WaitGroup, errCh chan error) string {
	var output string

	if block.Properties.String != "" {
		block.ParsedProp.Title = ParsePropTitle(block.Properties.String)
		block.ParsedProp.Language = ParsePropLanguage(block.Properties.String)
	}

	indent := strings.Repeat("   ", indentLv)
	text := strings.ReplaceAll(block.ParsedProp.Title, "\n", "\n"+indent)

	anchor := ""
	if block.Type == "header" || block.Type == "sub_header" || block.Type == "sub_sub_header" {
		anchor = utils.SanitizeFileName(text)
	}

	switch block.Type {
	case "header":
		output = markdown.Header(indent, text, anchor)
	case "sub_header":
		output = markdown.SubHeader(indent, text, anchor)
	case "sub_sub_header":
		output = markdown.SubSubHeader(indent, text, anchor)
	case "text":
		output = markdown.Text(indent, text)
	case "code":
		output = markdown.Code(indent, block.ParsedProp.Language, text)
	case "divider":
		output = markdown.Divider(indent)
	case "bulleted_list":
		output = markdown.BulletedList(indent, text)
	case "numbered_list":
		output = markdown.NumberedList(indent, block.Number, text)
	case "toggle":
		var content string
		for _, child := range block.Children {
			content += ParseBlock(pageID, child, indentLv+1, headers, wg, errCh)
		}
		output = markdown.Toggle(indent, text, content)
		block.Children = nil
	case "quote":
		output = markdown.Quote(indent, text)
	case "callout":
		output = markdown.Callout(indent, text)
	case "image":
		imageFileName := SaveImageIfNotExist(pageID, block.ID, wg, errCh)
		output = markdown.Image(indent, filepath.Join("/assets/pages", pageID, imageFileName))
	case "to_do":
		output = markdown.ToDo(indent, text, ParseChecked(block.Properties.String))
	case "table":
		output = createTableMarkdown(&block, block.Children)
		block.Children = nil
	case "column_list":
		var content string
		for _, child := range block.Children {
			// 컬럼 리스트의 자식(컬럼)은 들여쓰기를 추가하지 않음
			content += ParseBlock(pageID, child, indentLv, headers, wg, errCh)
		}
		output = content
		block.Children = nil // 자식 블록은 이미 처리되었으므로 nil로 설정
	case "column":
		var content string
		for _, child := range block.Children {
			// 컬럼 내의 블록은 들여쓰기를 추가하지 않음
			content += ParseBlock(pageID, child, indentLv, headers, wg, errCh)
		}
		output = content
		block.Children = nil
	case "table_of_contents":
		var tocBuilder strings.Builder
		for _, h := range headers {
			tocIndent := strings.Repeat("  ", h.Level-1)
			escapedTitle := strings.ReplaceAll(h.Title, "|", "\\|")
			tocBuilder.WriteString(fmt.Sprintf("%s- [%s](#%s)\n", tocIndent, escapedTitle, h.Anchor))
		}
		output = tocBuilder.String()
	case "bookmark":
		url, title, _ := ParseBookmark(block.Properties.String)
		output = markdown.Bookmark(indent, url, title)
	default:
		if block.Type != "" {
			log.Printf("Unsupported block type: %s", block.Type)
		}
		output = ""
	}

	for _, child := range block.Children {
		output += ParseBlock(pageID, child, indentLv+1, headers, wg, errCh)
	}

	return output
}

func ParsePropLanguage(properties string) (language string) {
	var props map[string]interface{}
	if err := json.Unmarshal([]byte(properties), &props); err != nil {
		panic(err)
	}

	if langValue, ok := props["language"]; ok {
		langArray := langValue.([]interface{})
		language = langArray[0].([]interface{})[0].(string)
	} else {
		language = ""
	}

	return
}

func ParsePropTitle(properties string) (text string) {
	var props map[string]interface{}
	if err := json.Unmarshal([]byte(properties), &props); err != nil {
		panic(err)
	}

	text = ParseText(props["title"])

	return
}

func ParseText(text interface{}) (parsedText string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("패닉 복구:", err)
		}
	}()

	// INFO - [ ["type",[["b"]]], [" "], ["자체가",[["i"]]], [" "], ["하나의",[["_"]]], [" "], ["변환으로",[["s"]]], ...]
	for _, value := range text.([]interface{}) {
		values := value.([]interface{})
		v := values[0].(string)

		// 길이가 1보다 큰 경우, text 에 대한 추가 형식 변환이 존재한다.
		if len(values) > 1 {
			for _, format := range values[1].([]interface{}) {
				f := format.([]interface{})
				switch f[0].(string) {
				case "b":
					v = markdown.Bold(v)
				case "i":
					v = markdown.Italic(v)
				case "s":
					v = markdown.Strikethrough(v)
				case "c":
					v = markdown.InlineCode(v)
				case "_":
					v = markdown.Underline(v)
				case "e":
					v = markdown.Equation(f[1].(string)) // [ "⁍", [["e","x+1"]] ]
				case "a":
					v = markdown.Link(v, f[1].(string))
				case "h":
					// 배경색이므로 무시
				case "‣":
					// 페이지 혹은 기타 노션 내부 링크이므로 무시
				default:
					//fmt.Printf("Error: Failed to parse properties. (%v) (%s) type\n", properties, f[0].(string))
				}
			}
			parsedText += v
		} else {
			parsedText += v
		}
	}

	return
}

func ParseChecked(properties string) bool {
	var propData map[string]interface{}
	err := json.Unmarshal([]byte(properties), &propData)
	utils.CheckError(err)

	checkedData := propData["checked"].([]interface{})
	checkedValue := checkedData[0].([]interface{})[0].(string)

	return checkedValue == "Yes"
}

func ParseBookmark(properties string) (url string, title string, description string) {
	var propData map[string]interface{}
	if err := json.Unmarshal([]byte(properties), &propData); err != nil {
		return "", "", ""
	}

	if link, ok := propData["link"]; ok {
		url = link.([]interface{})[0].([]interface{})[0].(string)
	}

	if titleArray, ok := propData["title"]; ok {
		title = titleArray.([]interface{})[0].([]interface{})[0].(string)
	}

	if descriptionArray, ok := propData["description"]; ok {
		description = descriptionArray.([]interface{})[0].([]interface{})[0].(string)
	}
	return
}
