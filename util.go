package buildpack

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func filterURI(rawURL string) (string, error) {
	unsafeURL, err := url.Parse(rawURL)

	if err != nil {
		return "", err
	}

	var safeURL string

	if unsafeURL.User == nil {
		safeURL = rawURL
		return safeURL, nil
	}

	redactedUserInfo := url.UserPassword("-redacted-", "-redacted-")

	unsafeURL.User = redactedUserInfo
	safeURL = unsafeURL.String()

	return safeURL, nil
}

func checkMD5(filePath, expectedMD5 string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	hashInBytes := hash.Sum(nil)[:16]
	actualMD5 := hex.EncodeToString(hashInBytes)

	if actualMD5 != expectedMD5 {
		return newBuildpackError("FIXME", "md5 mismatch: expected: %s got: %s", expectedMD5, actualMD5)
	}
	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return writeToFile(resp.Body, dest)
}

func copyFile(source, dest string) error {
	fh, err := os.Open(source)
	if err != nil {
		return err
	}
	defer fh.Close()

	return writeToFile(fh, dest)
}

func writeToFile(source io.Reader, dest string) error {
	err := os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	if err != nil {
		return err
	}

	fh, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = io.Copy(fh, source)
	if err != nil {
		return err
	}

	return nil
}
