package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func MergePDFSubmitHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Invalid upload", http.StatusBadRequest)
		return
	}

	order := strings.Split(
		r.FormValue("pdf_order"),
		",",
	)

	files := r.MultipartForm.File["pdfs"]
	ids := r.MultipartForm.Value["pdf_ids"]

	if len(files) < 2 {
		http.Error(
			w,
			"Please upload at least 2 PDFs",
			http.StatusBadRequest,
		)
		return
	}

	if len(files) != len(ids) {
		http.Error(
			w,
			"Invalid upload data",
			http.StatusBadRequest,
		)
		return
	}

	fileMap :=
		make(map[string]*multipart.FileHeader)

	for i, id := range ids {
		fileMap[id] = files[i]
	}

	var readers []io.ReadSeeker

	for _, id := range order {

		header := fileMap[id]

		if header == nil {
			continue
		}

		file, err := header.Open()
		if err != nil {
			http.Error(
				w,
				"Failed to read PDF",
				http.StatusInternalServerError,
			)
			return
		}

		data, err := io.ReadAll(file)

		file.Close()

		if err != nil {
			http.Error(
				w,
				"Failed to read PDF",
				http.StatusInternalServerError,
			)
			return
		}

		if len(data) < 5 ||
			string(data[:5]) != "%PDF-" {

			http.Error(
				w,
				"Invalid PDF uploaded",
				http.StatusBadRequest,
			)
			return
		}

		readers = append(
			readers,
			bytes.NewReader(data),
		)
	}

	var merged bytes.Buffer

	conf := model.NewDefaultConfiguration()

	err = api.MergeRaw(
		readers,
		&merged,
		false,
		conf,
	)

	if err != nil {
		log.Println("pdf merge error:", err)

		http.Error(
			w,
			"Failed to merge PDFs",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/pdf",
	)

	w.Header().Set(
		"Content-Disposition",
		`attachment; filename="merged.pdf"`,
	)

	_, _ = w.Write(
		merged.Bytes(),
	)
}


func MergePDFHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	tpl.ExecuteTemplate(
		w,
		"mergepdf.html",
		nil,
	)
}
