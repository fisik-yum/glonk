package main

import (
	"io"
	"net/http"
)

func getFile(url string) []byte {
	FileResp, err := http.Get(url)
	if err != nil {
		return nil
	}

	FileBytes, _ := io.ReadAll(FileResp.Body)
	if err != nil {
		return nil
	}

	return FileBytes

}
