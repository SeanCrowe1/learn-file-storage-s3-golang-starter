package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(w http.ResponseWriter) (string, error) {
	buffer := make([]byte, 32)
	_, err := rand.Read(buffer)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate URL", nil)
	}

	encoded := make([]byte, base64.RawURLEncoding.EncodedLen(len(buffer)))
	base64.RawURLEncoding.Encode(encoded, buffer)

	filename := string(encoded)
	return filename, nil
}
