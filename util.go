package buildpack

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
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
		Log.Error("DEPENDENCY_MD5_MISMATCH: expected md5: %s, actual md5: %s", expectedMD5, actualMD5)
		return fmt.Errorf("expected md5: %s actual md5: %s", expectedMD5, actualMD5)
	}
	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		Log.Error("Could not download: %d", resp.StatusCode)
		return errors.New("file download failed")
	}

	return writeToFile(resp.Body, dest)
}

func copyFile(source, dest string) error {
	fh, err := os.Open(source)
	if err != nil {
		Log.Error("Could not be found")
		return err
	}
	defer fh.Close()

	return writeToFile(fh, dest)
}

func writeToFile(source io.Reader, dest string) error {
	err := os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	if err != nil {
		Log.Error("Could not create %s", filepath.Dir(dest))
		return err
	}

	fh, err := os.Create(dest)
	if err != nil {
		Log.Error("Could not write to %s", dest)
		return err
	}
	defer fh.Close()

	_, err = io.Copy(fh, source)
	if err != nil {
		Log.Error("Could not write to %s", dest)
		return err
	}

	return nil
}
