package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// import
// import

const (
	maxUploadSize = 100 * 1024 * 1024 // 2 MB
	uploadPath    = "./uploads"       // Change to your desired path
)

func main() {
	fmt.Println("hello")
	API_KEY, err := os.ReadFile("keys.txt")
	if err == nil {
		fmt.Print(string(API_KEY))
	}

	h1 := func(w http.ResponseWriter, r *http.Request) {

		templ := template.Must(template.ParseFiles("index.html"))
		templ.Execute(w, nil)
	}

	h2 := func(w http.ResponseWriter, r *http.Request) {
		log.Print("Request received")
		log.Print(r.Header.Get("HX-Request"))

		dt := time.Now()
		dtf := dt.Format("01_02_2006_15_04_05")
		// fmt.Println()

		r.ParseMultipartForm(maxUploadSize)

		// Retrieve the file and prompt from form data
		prompt := r.FormValue("prmpt")
		fmt.Println("prompt: ", prompt)
		file, fileHeader, err := r.FormFile("image")

		if err != nil {
			http.Error(w, "Invalid file", http.StatusBadRequest)
			fmt.Print(fileHeader)
			return
		}

		defer file.Close()

		// Create a new file in the uploads directory
		ImageNewFilePath := filepath.Join(uploadPath, dtf+".jpg")
		PromptNewFilePath := filepath.Join(uploadPath, dtf+".txt")

		newFile, err := os.Create(ImageNewFilePath)
		if err != nil {
			http.Error(w, "Unable to create file", http.StatusInternalServerError)
			return
		}
		x, e := os.Create(PromptNewFilePath)
		_, err = x.WriteString(prompt)

		if e != nil {
			log.Print("err")
		}

		defer newFile.Close()

		// Copy the uploaded file to the new file on disk
		if _, err := io.Copy(newFile, file); err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}

		//
		//

		ctx := context.Background()
		// Access your API key as an environment variable (see "Set up your API key" above)
		client, err := genai.NewClient(ctx, option.WithAPIKey(string(API_KEY)))
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()

		modelfile, err := client.UploadFileFromPath(ctx, filepath.Join(uploadPath, dtf+".jpg"), nil)
		if err != nil {
			log.Fatal(err)
		}
		defer client.DeleteFile(ctx, modelfile.Name)

		model := client.GenerativeModel("gemini-1.5-flash")
		resp, err := model.GenerateContent(ctx,
			genai.FileData{URI: modelfile.URI},
			genai.Text(prompt))
		if err != nil {
			log.Fatal(err)
		}

		// printResponse(resp)
		// model := client.GenerativeModel("gemini-1.5-flash")
		// resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		// if err != nil {
		// 	log.Fatal(err)
		// }

		if resp.Candidates != nil {
			for _, v := range resp.Candidates {
				for _, k := range v.Content.Parts {
					fmt.Println(k.(genai.Text))
				}
			}
		}

		// fmt.Fprintf(w, "Successfully uploaded: %s\n", fileHeader.Filename)
	}

	http.HandleFunc("/", h1)
	http.HandleFunc("/upload", h2)
	log.Fatal(http.ListenAndServe(":8000", nil))

}
