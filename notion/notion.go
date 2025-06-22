package notion

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/shinychan95/Chan/utils"
)

var (
	ApiKey     string
	ApiVersion = "2022-06-28"
	PostDir    string
	ImgDir     string
	db         *sql.DB
)

func Init(apiKey, postDir, imgDir, dbPath string) {
	ApiKey = apiKey
	PostDir = postDir
	ImgDir = imgDir

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	utils.CheckError(err)
}

func Close() {
	err := db.Close()
	utils.CheckError(err)
}

//////////////////////
// Get Data From DB //
//////////////////////

func getCollectionId(rootID string) (colId string) {
	query := "SELECT collection_id FROM block WHERE id = ? AND type = 'collection_view'"
	log.Printf("Executing query: %s, with rootID: %s", query, rootID)
	rows, err := db.Query(query, rootID)
	utils.CheckError(err)
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&colId)
		utils.CheckError(err)

		if rows.Next() {
			utils.ExecError("more than one row returned")
		}
	}

	if colId == "" {
		utils.ExecError("cannot get collection id")
	}

	return
}

func getCollectionSchema(collectionId string) (schemaMap map[string]Schema) {
	query := "SELECT schema FROM collection WHERE id = ?"
	log.Printf("Executing query: %s, with collectionId: %s", query, collectionId)
	rows, err := db.Query(query, collectionId)
	utils.CheckError(err)
	defer rows.Close()

	var rawSchema string
	for rows.Next() {
		err = rows.Scan(&rawSchema)
		utils.CheckError(err)

		if rows.Next() {
			utils.ExecError("more than one row returned")
		}

	}

	err = json.Unmarshal([]byte(rawSchema), &schemaMap)
	utils.CheckError(err)

	return
}

func getPagesWithProperties(parentId string, schema map[string]Schema) (pages []Page) {
	query := "SELECT id, properties FROM block WHERE parent_id = ? AND type = 'page' AND is_template IS NULL AND alive = 1"
	log.Printf("Executing query: %s, with parentId: %s", query, parentId)
	rows, err := db.Query(query, parentId)
	utils.CheckError(err)
	defer rows.Close()

	for rows.Next() {
		var (
			id            string
			rawProperties string
		)
		err = rows.Scan(&id, &rawProperties)
		utils.CheckError(err)

		page := Page{ID: id}
		parsePageProperties(&page, rawProperties, schema)

		pages = append(pages, page)
	}

	return
}

func getRootType(rootID string) (t string) {
	query := "SELECT type FROM block WHERE id = ?"
	log.Printf("Executing query: %s, with rootID: %s", query, rootID)
	rows, err := db.Query(query, rootID)
	utils.CheckError(err)
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&t)
		utils.CheckError(err)

		if rows.Next() {
			utils.ExecError("more than one row returned")
			return
		}
	}

	return
}

func getBlockData(blockID string) (block Block) {
	query := "SELECT id, type, content, properties, format FROM block WHERE id = ?"
	rows, err := db.Query(query, blockID)
	utils.CheckError(err)
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&block.ID, &block.Type, &block.Content, &block.Properties, &block.Format)
		utils.CheckError(err)

		if rows.Next() {
			utils.ExecError("more than one row returned")
			return
		}
	}

	return
}
