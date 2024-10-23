package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	vision "cloud.google.com/go/vision/apiv1"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// import
// import

const (
	maxUploadSize = 100 * 1024 * 1024 // 2 MB
)

func main() {
	fmt.Println("Server's up")
	API_KEY := "YOUR_API_KEY_GOES_HERE"

	h1 := func(w http.ResponseWriter, r *http.Request) {

		templ := template.Must(template.ParseFiles("index.html"))
		templ.Execute(w, nil)
	}

	h2 := func(w http.ResponseWriter, r *http.Request) {

		log.Print("Request received")
		log.Print(r.Header.Get("HX-Request"))

		dt := time.Now()
		dtf := dt.Format("01_02_2006_15_04_05")

		r.ParseMultipartForm(maxUploadSize)

		// Retrieve the file and prompt from form data
		prompt := r.FormValue("prmpt")
		fmt.Println("prompt: ", prompt)
		flag := r.FormValue("flag")

		if flag == "img" || flag == "doc" {

			file, fileHeader, err := r.FormFile("image")
			imageForBucket, _ := io.ReadAll(file)
			promptForBucket := []byte(prompt)

			if err != nil {
				http.Error(w, "Invalid file", http.StatusBadRequest)
				fmt.Print(fileHeader)
				return
			}

			defer file.Close()

			ctxt := context.Background()

			ctx := context.Background()

			bucketName := "devfest2024bucket"

			// Create a Cloud Storage client
			client, err := storage.NewClient(ctx)
			if err != nil {
				fmt.Errorf("failed to create storage client: %w", err)
			}
			defer client.Close()

			// Create a bucket object
			bucket := client.Bucket(bucketName)

			// Create an object for writing
			imageObject := bucket.Object("img" + dtf)
			promptObject := bucket.Object("prompt" + dtf)
			imagewc := imageObject.NewWriter(ctx)
			if err != nil {
				fmt.Errorf("failed to create object writer: %w", err)
			}
			defer imagewc.Close()

			promptwc := promptObject.NewWriter(ctx)
			if err != nil {
				fmt.Errorf("failed to create object writer: %w", err)
			}
			defer promptwc.Close()

			// Write the data to the object
			_, err = imagewc.Write(imageForBucket)
			if err != nil {
				fmt.Errorf("failed to write data to object: %w", err)
			}
			_, err = promptwc.Write(promptForBucket)
			if err != nil {
				fmt.Errorf("failed to write data to object: %w", err)
			}

			if flag == "doc" {

				vClient, _ := vision.NewImageAnnotatorClient(ctx)

				image, _ := vision.NewImageFromReader(file)

				annotation, _ := vClient.DetectDocumentText(ctx, image, nil)

				resp := ""
				if annotation == nil {
					resp = "No text found."
				} else {
					resp = "Document Text:"
					resp = resp + annotation.Text + "\n"

					for _, page := range annotation.Pages {
						for _, block := range page.Blocks {
							for _, paragraph := range block.Paragraphs {
								for _, word := range paragraph.Words {
									symbols := make([]string, len(word.Symbols))
									for i, s := range word.Symbols {
										symbols[i] = s.Text
									}
									wordText := strings.Join(symbols, "")
									resp = resp + wordText
								}
							}
						}
					}
					fmt.Print(resp)
				}

				templ, _ := template.New("t").Parse(resp)
				templ.Execute(w, nil)
				return
			}

			client2, err := genai.NewClient(ctx, option.WithAPIKey(string(API_KEY)))
			if err != nil {
				log.Fatal(err)
			}
			defer client.Close()
			model := client2.GenerativeModel("gemini-1.5-flash")

			genaiImgData1 := genai.ImageData("jpeg", imageForBucket)

			prompt := []genai.Part{
				genaiImgData1,
				genai.Text(
					prompt),
			}
			resp, err := model.GenerateContent(ctxt, prompt...)

			response := ""
			if resp.Candidates != nil {
				for _, v := range resp.Candidates {
					for _, k := range v.Content.Parts {
						response = response + string(k.(genai.Text))
					}
				}
			}

			responseObject := bucket.Object("response" + dtf)
			responsewc := responseObject.NewWriter(ctx)
			if err != nil {
				fmt.Errorf("failed to create object writer: %w", err)
			}
			defer imagewc.Close()
			_, err = responsewc.Write([]byte(response))
			if err != nil {
				fmt.Errorf("failed to write data to object: %w", err)
			}

			fmt.Println(response)
			templ, _ := template.New("t").Parse(response)
			templ.Execute(w, nil)

		} else if flag == "vid" {

			file, fileHeader, err := r.FormFile("video")
			defer file.Close()

			videoForBucket, _ := io.ReadAll(file)

			if err != nil {
				http.Error(w, "Invalid file", http.StatusBadRequest)
				fmt.Print(fileHeader)
				return
			}

			ctx := context.Background()
			fmt.Println("context bg")
			client, err := storage.NewClient(ctx)
			fmt.Println("storage newclient")
			if err != nil {
				fmt.Errorf("storage.NewClient: %w", err)
			}
			defer client.Close()

			bucketName := "devfest2024bucket"

			ctx, cancel := context.WithTimeout(ctx, time.Second*50)
			defer cancel()

			// Upload an object with storage.Writer.
			wc := client.Bucket(bucketName).Object("video" + dtf).NewWriter(ctx)
			fmt.Println("wc")
			wc.ChunkSize = 0 // note retries are not supported for chunk size 0.

			fmt.Println("storage writer")

			// Write the data to the object
			_, err = wc.Write(videoForBucket)
			if err != nil {
				fmt.Errorf("failed to write data to object: %w", err)
			}

			fmt.Println("wc write")

			if err := wc.Close(); err != nil {
				fmt.Errorf("Writer.Close: %w", err)
			}
			fmt.Println("%v uploaded to %v.\n", "video"+dtf, bucketName)

			ctx2 := context.Background()
			client2, err := storage.NewClient(ctx2)
			if err != nil {
				fmt.Errorf("storage.NewClient: %w", err)
			}
			defer client2.Close()

			ctx2, cancel2 := context.WithTimeout(ctx2, time.Second*50)
			defer cancel2()

			rc, err := client.Bucket(bucketName).Object("video" + dtf).NewReader(ctx2)
			if err != nil {
				fmt.Errorf("Object(%q).NewReader: %w", "video"+dtf, err)
			}
			defer rc.Close()

			ctx3 := context.Background()

			vClient, err := genai.NewClient(ctx3, option.WithAPIKey(string(API_KEY)))
			if err != nil {
				log.Fatal(err)
			}
			defer vClient.Close()

			// Optionally set a display name.
			opts := genai.UploadFileOptions{DisplayName: "video" + dtf}

			// Let the API generate a unique `name` for the file by passing an empty string.
			// If you specify a `name`, then it has to be globally unique.
			reader := bytes.NewReader(videoForBucket)
			response3, err := vClient.UploadFile(ctx3, "", reader, &opts)

			if err != nil {
				log.Fatal(err)
			}

			// View the response.
			var VideoFile *genai.File = response3
			fmt.Printf("Uploaded file %s as: %q\n", VideoFile.DisplayName, VideoFile.URI)

			// Poll GetFile() on a set interval (10 seconds here) to
			// check file state.
			for response3.State == genai.FileStateProcessing {
				// Sleep for 10 seconds
				time.Sleep(10 * time.Second)

				// Fetch the file from the API again.
				response3, err = vClient.GetFile(ctx3, VideoFile.Name)
				if err != nil {
					log.Fatal(err)
				}
			}

			// View the response.
			fmt.Printf("File %s is ready for inference as: %q\n",
				response3.DisplayName, response3.URI)

			vPrompt := []genai.Part{
				genai.FileData{URI: response3.URI},
				genai.Text(prompt),
			}
			model := vClient.GenerativeModel("gemini-1.5-flash")
			// Generate content using the prompt.

			vResp, err := model.GenerateContent(ctx3, vPrompt...)
			if err != nil {
				log.Fatal(err)
			}

			// Handle the response of generated text.
			toDisplay := ""
			for _, c := range vResp.Candidates {
				if c.Content != nil {

					for _, k := range c.Content.Parts {
						toDisplay = toDisplay + string(k.(genai.Text))
					}
				}
			}

			templ, _ := template.New("t").Parse(toDisplay)
			templ.Execute(w, nil)
		}
	}

	http.HandleFunc("/", h1)
	http.HandleFunc("/upload", h2)
	log.Fatal(http.ListenAndServe(":8080", nil))
	http.Handle("/", http.FileServer(http.Dir("/")))

}
