package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
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
	uploadPath    = "./uploads"       // Change to your desired path
)

func main() {
	fmt.Println("hello")
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
		// fmt.Println()

		r.ParseMultipartForm(maxUploadSize)

		// Retrieve the file and prompt from form data
		prompt := r.FormValue("prmpt")
		fmt.Println("prompt: ", prompt)
		flag := r.FormValue("flag")

		fmt.Println("flag::::::::::::::::", flag)

		if flag == "img" || flag == "doc" {

			fmt.Println("inside img if")
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

			// projectID := "devfest2024-438119"
			ctx := context.Background()

			// Replace with your project ID
			// projectID := "devfest2024-438119"
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
				fmt.Println("in ocr")
				// imgForAPI := client.Bucket(bucketName).Object("img" + dtf)

				vClient, _ := vision.NewImageAnnotatorClient(ctx)

				image, _ := vision.NewImageFromReader(file)
				fmt.Println(image)
				fmt.Println(image)
				// image, _ := vision.NewImageFromReader(f)

				annotation, _ := vClient.DetectDocumentText(ctx, image, nil)
				fmt.Println(annotation)

				resp := ""
				if annotation == nil {
					resp = "No text found."
				} else {
					resp = "Document Text:"
					resp = resp + annotation.Text + "\n"

					// fmt.Fprintln(w, "Pages:")
					for _, page := range annotation.Pages {
						for _, block := range page.Blocks {
							for _, paragraph := range block.Paragraphs {
								for _, word := range paragraph.Words {
									symbols := make([]string, len(word.Symbols))
									for i, s := range word.Symbols {
										symbols[i] = s.Text
									}
									wordText := strings.Join(symbols, "")
									// fmt.Fprintf(w, "\t\t\t\tConfidence: %f, Symbols: %s\n", word.Confidence, wordText)
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

			// opts := &storage.SignedURLOptions{
			// 	Scheme:  storage.SigningSchemeV4,
			// 	Method:  "GET",
			// 	Expires: time.Now().Add(15 * time.Minute),
			// }

			// // object = storageClient.Bucket(bucketName).Object(objectName)
			// objectURL, err := storageClient.Bucket(bucketName).SignedURL(dtf, opts)
			// // objectURL, err := object.SignedURL(ctxt, 3600) // Set the expiration time in seconds
			// if err != nil {
			// 	log.Fatalf("Failed to get object URL: %v", err)
			// }

			// // Create a Gemini Vision model client
			// geminiClient, err := genai.NewClient(ctxt, option.WithAPIKey(string(API_KEY)))
			// if err != nil {
			// 	log.Fatalf("Failed to create Gemini Vision model client: %v", err)
			// }
			// defer geminiClient.Close()

			// // Prepare the request
			// request := map[string]interface{}{
			// 	"image": objectURL,
			// }

			// // Send the request
			// response, err := geminiClient.Ca(ctxt, "gemini-1.5-flash", request)
			// if err != nil {
			// 	log.Fatalf("Failed to call Gemini Vision model: %v", err)
			// }

			// // Process the response
			// var result map[string]interface{}
			// err = json.Unmarshal(response, &result)
			// if err != nil {
			// 	log.Fatalf("Failed to parse response: %v", err)
			// }

			// fmt.Println("Result:", result)

			//

			// ImageNewFilePath := filepath.Join(uploadPath, dtf+".jpg")
			// PromptNewFilePath := filepath.Join(uploadPath, dtf+".txt")
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

			// newFile, err := os.Create(ImageNewFilePath)
			// if err != nil {
			// 	http.Error(w, "Unable to create file", http.StatusInternalServerError)
			// 	return
			// }
			// x, e := os.Create(PromptNewFilePath)
			// _, err = x.WriteString(prompt)

			// if e != nil {
			// 	log.Print("err")
			// }

			// defer newFile.Close()

			// // Copy the uploaded file to the new file on disk
			// if _, err := io.Copy(newFile, file); err != nil {
			// 	http.Error(w, "Unable to save file", http.StatusInternalServerError)
			// 	return
			// }

			// Access your API key as an environment variable (see "Set up your API key" above)
			// client, err := genai.NewClient(ctx, option.WithAPIKey(string(API_KEY)))
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// defer client.Close()

			// modelfile, err := client.UploadFileFromPath(ctx, filepath.Join(uploadPath, dtf+".jpg"), nil)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// defer client.DeleteFile(ctx, modelfile.Name)

			// ----------------------------- THIS ONE
			// model := client.GenerativeModel("gemini-1.5-flash")
			// resp, err := model.GenerateContent(ctx,
			// 	genai.FileData{URI: modelfile.URI},
			// 	genai.Text(prompt))
			// if err != nil {
			// 	log.Fatal(err)
			// }

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
			// fmt.Fprintf(w, "Successfully uploaded: %s\n", fileHeader.Filename)
			templ, _ := template.New("t").Parse(response)
			templ.Execute(w, nil)

		} else if flag == "vid" {
			fmt.Print("videoooooooooo")

			file, fileHeader, err := r.FormFile("video")
			defer file.Close()

			// ////////////////////////////////////////////////////////////////

			videoForBucket, _ := io.ReadAll(file)
			promptForBucket := []byte(prompt)
			fmt.Println(videoForBucket)
			fmt.Println(promptForBucket)
			if err != nil {
				http.Error(w, "Invalid file", http.StatusBadRequest)
				fmt.Print(fileHeader)
				return
			}

			// ctxt := context.Background()

			// projectID := "devfest2024-438119"
			ctx := context.Background()
			fmt.Println("context bg")
			client, err := storage.NewClient(ctx)
			fmt.Println("storage newclient")
			if err != nil {
				fmt.Errorf("storage.NewClient: %w", err)
			}
			defer client.Close()

			// Replace with your project ID
			// projectID := "devfest2024-438119"
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

			/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

			data, err := ioutil.ReadAll(rc)
			if err != nil {
				fmt.Errorf("ioutil.ReadAll: %w", err)
			}
			fmt.Println("Blob %v downloaded.\n")
			fmt.Println(data)

			/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
			ctx3 := context.Background()
			fmt.Println("ctx3")

			vClient, err := genai.NewClient(ctx3, option.WithAPIKey(string(API_KEY)))
			if err != nil {
				log.Fatal(err)
			}
			defer vClient.Close()
			fmt.Println("vclient done")

			//open?

			// Optionally set a display name.
			opts := genai.UploadFileOptions{DisplayName: "video" + dtf}
			fmt.Println("opts")

			// Let the API generate a unique `name` for the file by passing an empty string.
			// If you specify a `name`, then it has to be globally unique.
			reader := bytes.NewReader(videoForBucket)
			response3, err := vClient.UploadFile(ctx3, "", reader, &opts)
			fmt.Println("response")
			fmt.Println(response3)

			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(response3)

			// View the response.
			var VideoFile *genai.File = response3
			fmt.Printf("Uploaded file %s as: %q\n", VideoFile.DisplayName, VideoFile.URI)

			// Poll GetFile() on a set interval (10 seconds here) to
			// check file state.
			for response3.State == genai.FileStateProcessing {
				fmt.Print(".")
				// Sleep for 10 seconds
				time.Sleep(10 * time.Second)

				// Fetch the file from the API again.
				response3, err = vClient.GetFile(ctx3, VideoFile.Name)
				if err != nil {
					log.Fatal(err)
				}
			}
			fmt.Println()

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
