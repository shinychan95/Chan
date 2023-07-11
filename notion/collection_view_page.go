package notion

import (
	"github.com/shinychan95/make-notion-blog/utils"
	"sync"
)

type Schema struct {
	Name      string              `json:"name"`
	Type      string              `json:"type"`
	Options   []map[string]string `json:"options,omitempty"`
	OptionIds []string            `json:"optionIds,omitempty"`
}

func HandleCollectionViewPage(rootId string, wg *sync.WaitGroup, errCh chan error) {
	rootType := getRootType(rootId)
	if rootType != "collection_view_page" {
		Close() // db close
		utils.ExecError("block type is not same with exec type")
	}

	// block 테이블 내 collection_id 값을 가져온다., 해당 값을 parent_id 로 하는 페이지들을 구한다.
	collectionId := getCollectionId(rootId)

	// collection 테이블 내 해당 collection 의 스키마를 가져온다.
	collectionSchema := getCollectionSchema(collectionId)

	// block 테이블 내 해당 collection 을 부모로 하는 페이지들을 가져온다. (template is NULL, alive is 1)
	pages := getPagesWithProperties(collectionId, collectionSchema)

	// property 내 Status 가 Drafting 인 글들만 프로세스를 실행한다.
	for _, page := range pages {
		if page.Status == "Published" {
			wg.Add(1)
			go func(page Page) {
				handlePage(page, wg, errCh)
				wg.Done()
			}(page)
		}
	}

	wg.Wait()
}
