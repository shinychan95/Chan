package markdown

import (
	"fmt"
	"strings"
)

func Header(indent string, text string, anchor string) string {
	return fmt.Sprintf("\n<h1 id=\"%s\">%s</h1>\n", anchor, text)
}

func SubHeader(indent string, text string, anchor string) string {
	return fmt.Sprintf("\n<h2 id=\"%s\">%s</h2>\n", anchor, text)
}

func SubSubHeader(indent string, text string, anchor string) string {
	return fmt.Sprintf("\n<h3 id=\"%s\">%s</h3>\n", anchor, text)
}

func Text(indent string, text string) string {
	if text == "" {
		text = "<br/>\n\n"
	}

	return fmt.Sprintf("%s%s\n\n", indent, text)
}

func Code(indent, lang, text string) string {
	return fmt.Sprintf("%s```%s\n%s%s\n%s```\n\n", indent, lang, indent, text, indent)
}

func Divider(indent string) string {
	return fmt.Sprintf("%s\n---\n\n", indent)
}

func BulletedList(indent, text string) string {
	return fmt.Sprintf("%s- %s\n\n", indent, text)
}

func NumberedList(indent string, number uint8, text string) string {
	return fmt.Sprintf("%s%d. %s\n\n", indent, number, text)
}

func Toggle(indent, text, content string) string {
	summary := fmt.Sprintf("%s<summary>%s</summary>\n", indent, text)
	return fmt.Sprintf("%s<details>\n%s%s%s</details>\n\n", indent, summary, content, indent)
}

func Quote(indent, text string) string {
	text = strings.Replace(text, "\n", "<br/>", -1)
	return fmt.Sprintf("%s> %s\n\n", indent, text)
}

func Callout(indent, text string) string {
	return fmt.Sprintf("%s> 🦖 %s\n\n", indent, text)
}

func Image(indent, imagePath string) string {
	return fmt.Sprintf("%s![](%s)\n", indent, imagePath)
}

func ToDo(indent string, text string, checked bool) string {
	if checked {
		return fmt.Sprintf("%s- [x] %s\n", indent, text)
	}
	return fmt.Sprintf("%s- [ ] %s\n", indent, text)
}

func Bookmark(indent, url, title string) string {
	if title == "" {
		title = url
	}
	return fmt.Sprintf("%s> 🔗 [%s](%s)\n", indent, title, url)
}
