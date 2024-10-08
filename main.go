package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	maxUploadSize = 100 * 1024 * 1024 // 2 MB
	uploadPath    = "./uploads"       // Change to your desired path
)

func main() {
	fmt.Println("hello")

	h1 := func(w http.ResponseWriter, r *http.Request) {
		templ := template.Must(template.ParseFiles("index.html"))
		templ.Execute(w, nil)
	}

	h2 := func(w http.ResponseWriter, r *http.Request) {
		log.Print("Request received")
		log.Print(r.Header.Get("HX-Request"))

		r.ParseMultipartForm(maxUploadSize)

		// Retrieve the file and prompt from form data
		prompt := r.FormValue("prmpt")
		fmt.Println(prompt)
		file, fileHeader, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "Invalid file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Create a new file in the uploads directory
		newFilePath := filepath.Join(uploadPath, fileHeader.Filename)
		newFile, err := os.Create(newFilePath)
		if err != nil {
			http.Error(w, "Unable to create file", http.StatusInternalServerError)
			return
		}
		defer newFile.Close()

		// Copy the uploaded file to the new file on disk
		if _, err := io.Copy(newFile, file); err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Successfully uploaded: %s\n", fileHeader.Filename)

	}

	http.HandleFunc("/", h1)
	http.HandleFunc("/upload", h2)
	log.Fatal(http.ListenAndServe(":8000", nil))

}
