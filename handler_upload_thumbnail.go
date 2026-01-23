package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory int64 = 10 * 1024 * 1024
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Content-Type for thumbnail", nil)
		return
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type header", err)
		return
	}

	var extension string
	switch mediaType {
	case "image/jpeg":
		extension = "jpeg"
	case "image/png":
		extension = "png"
	default:
		respondWithError(w, http.StatusBadRequest, "Invalid file format", nil)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}

	buffer := make([]byte, 32)
	_, err = rand.Read(buffer)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate URL", nil)
	}

	encoded := make([]byte, base64.RawURLEncoding.EncodedLen(len(buffer)))
	base64.RawURLEncoding.Encode(encoded, buffer)

	filename := string(encoded) + "." + extension
	path := filepath.Join(cfg.assetsRoot, filename)
	videoFile, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create file", err)
		return
	}
	defer videoFile.Close()

	_, err = io.Copy(videoFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving file", err)
		return
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%v/assets/%v", cfg.port, filename)
	video.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
