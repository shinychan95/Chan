package notion

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/shinychan95/Chan/utils"
)

type ImageBlock struct {
	Object         string    `json:"object"`
	ID             string    `json:"id"`
	CreatedTime    time.Time `json:"created_time"`
	LastEditedTime time.Time `json:"last_edited_time"`
	Type           string    `json:"type"`
	Image          struct {
		Type     string `json:"type"`
		File     ImageFile
		External ImageFile
	} `json:"image"`
}

type ImageFile struct {
	URL        string    `json:"url"`
	ExpiryTime time.Time `json:"expiry_time,omitempty"`
}

func SaveImageIfNotExist(rootID, imageId string, wg *sync.WaitGroup, errCh chan error) string {
	imageURL, err := getImageURL(imageId)
	utils.CheckError(err)

	imageFileName := fmt.Sprintf("%s.png", imageId)
	imagePath := filepath.Join(ImgDir, rootID, imageFileName)

	wg.Add(1)

	if !checkImageExist(imagePath) {
		go func(url, path string) {
			defer wg.Done()
			err = downloadImage(url, path)
			if err != nil {
				errCh <- err
				log.Printf("Error downloading image: %s", err)
			}
		}(imageURL, imagePath)
	} else {
		wg.Done()
	}

	return imageFileName
}

func downloadImage(url, imagePath string) error {
	resp, err := http.Get(url)
	utils.CheckError(err)
	defer resp.Body.Close()

	dir := filepath.Dir(imagePath)

	// 경로가 존재하지 않으면 폴더 생성
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		utils.CheckError(err)
	}

	// 이미지 파일 저장
	out, err := os.Create(imagePath)
	utils.CheckError(err)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	utils.CheckError(err)

	return nil
}

func getImageURL(blockID string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.notion.com/v1/blocks/%s", blockID), nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+ApiKey)
	req.Header.Add("Notion-Version", ApiVersion)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var imageBlock ImageBlock
	err = json.Unmarshal(body, &imageBlock)
	if err != nil {
		return "", err
	}

	if imageBlock.Image.Type == "file" {
		return imageBlock.Image.File.URL, nil
	}
	if imageBlock.Image.Type == "external" {
		return imageBlock.Image.External.URL, nil
	}

	return "", fmt.Errorf("unsupported image type or empty URL for block %s", blockID)
}

func checkImageExist(imagePath string) bool {
	// 파일 존재 여부 확인
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return false
	}

	return true
}
