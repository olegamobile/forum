package handlers

import (
	"forum/internal/db"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofrs/uuid"
)

func imageType(file string) bool {
	extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webm"}
	imageExtensions := make(map[string]bool)
	for _, ext := range extensions {
		imageExtensions[ext] = true
	}
	ext := strings.ToLower(filepath.Ext(file))
	return imageExtensions[ext]
}

func uniqueFileName(file string) (string, error) {
	fileExt := filepath.Ext(file)
	UUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	fileID := UUID.String() + fileExt
	return fileID, nil
}

func saveImageData(fileID string, postID int64, userID, originalName string, fileSize int, w http.ResponseWriter, uploadedFile multipart.File) (string, error) {

	imageUploadDir := "images"
	err := os.MkdirAll(imageUploadDir, 0777)
	if err != nil {
		log.Println("Error creating directory:", err)
		errMsg := "Internal error"
		return errMsg, err
	}
	filePath := filepath.Join(imageUploadDir, fileID)
	os.Chmod(filePath, 0644)
	savedFile, err := os.Create(filePath)
	if err != nil {
		errMsg := "Error while crreating a file."
		return errMsg, err
	}
	defer savedFile.Close()
	_, err = io.Copy(savedFile, uploadedFile)
	if err != nil {
		errMsg := "Error while saving file content."
		return errMsg, err
	}

	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Println("File does not exist immediately after save")
	}

	if err != nil {
		log.Println("Error writing to file:", err)
		errMsg := "Error while saving file content."
		return errMsg, err
	}

	_, err = db.DB.Exec(`INSERT INTO images (id, post_id, user_id, original_name, file_size) VALUES (?, ?, ?, ?, ?)`,
		fileID, postID, userID, originalName, fileSize)
	if err != nil {
		log.Println("Error inserting into DB:", err)
		errMsg := "Internal error"
		return errMsg, err
	}
	return "", nil
}
func getThreadImageURL(threadID int) ([]string, error) {
	rows, err := db.DB.Query(`SELECT id FROM images WHERE post_id = ?`, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var images []string
	for rows.Next() {
		var imageID string
		err := rows.Scan(&imageID)
		if err != nil {
			log.Println("Error scanning image ID:", err)
			return nil, err
		}
		imageURL := "/images/" + imageID
		images = append(images, imageURL)
	}
	return images, nil
}
