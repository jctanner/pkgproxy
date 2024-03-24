package hashing

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func Sha1Hash(text string) string {
	hasher := sha1.New()
	hasher.Write([]byte(text))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func Sha256Hash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func Sha256HashFile(fileName string) (string, error) {
	// Open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new SHA256 hash object
	hash := sha256.New()

	// Copy the file content to the hash object
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Compute the SHA256 checksum
	checksum := hash.Sum(nil)

	// Encode the checksum to a hexadecimal string
	return hex.EncodeToString(checksum), nil
}
