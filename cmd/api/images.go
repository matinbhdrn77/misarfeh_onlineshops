package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const BULK_FILE_SIZE = 32 << 20

func (app *application) uploadImagesHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20+512)

	if err := r.ParseMultipartForm(BULK_FILE_SIZE); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Get a reference to the fileHeaders.
	// They are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File["file"]

	var errNew string

	diskFiles := make([]*os.File, 0)

	for _, fileHeader := range files {
		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			errNew = err.Error()
			continue
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			errNew = err.Error()
			continue
		}

		// checking the content type
		// so we don't allow files other than images
		filetype := http.DetectContentType(buff)
		if filetype != "image/jpeg" && filetype != "image/png" && filetype != "image/jpg" {
			errNew = fmt.Sprintf("The %s file format is not allowed. Please upload a JPEG,JPG or PNG image", fileHeader.Filename)
			continue
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			errNew = err.Error()
			continue
		}

		err = os.MkdirAll("./uploads", os.ModePerm)
		if err != nil {
			errNew = err.Error()
			continue
		}

		f, err := os.Create(fmt.Sprintf("./uploads/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
		if err != nil {
			errNew = err.Error()
			continue
		}

		defer f.Close()

		diskFiles = append(diskFiles, f)

		_, err = io.Copy(f, file)
		if err != nil {
			errNew = err.Error()
			continue
		}
	}

	if errNew != "" {
		for _, file := range diskFiles {
			os.Remove(file.Name())
		}
		app.badRequestResponse(w, r, fmt.Errorf(errNew))
		return
	}

	serverUrl := fmt.Sprintf("localhost:%d/v1/images/", app.config.port)

	resp := map[string]string{}

	for i, file := range diskFiles {
		resp[fmt.Sprint(i+1)] = serverUrl + strings.Split(file.Name(), "/")[2]
	}

	app.writeJSON(w, http.StatusOK, envelope{"img_urls": resp}, nil)
}
