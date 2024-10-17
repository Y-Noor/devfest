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
	API_KEY, _ := os.ReadFile("keys.txt")

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

		if flag == "img" {
			fmt.Println("inside img if")
			file, fileHeader, err := r.FormFile("image")
			fmt.Println(file)

			// ==============================================================================================================================================
			// UPLOAD TO CLOUD BUCKET AND READ AND FEED INTO MODEL

			// file, fileHeader, err := r.FormFile("image")
			// imageForBucket, _ := io.ReadAll(file)
			// promptForBucket := []byte(prompt)

			// if err != nil {
			// 	http.Error(w, "Invalid file", http.StatusBadRequest)
			// 	fmt.Print(fileHeader)
			// 	return
			// }

			// defer file.Close()

			// ctxt := context.Background()

			// // projectID := "devfest2024-438119"
			// ctx := context.Background()

			// // Replace with your project ID
			// // projectID := "devfest2024-438119"
			// bucketName := "devfest2024bucket"

			// // Create a Cloud Storage client
			// client, err := storage.NewClient(ctx)
			// if err != nil {
			// 	fmt.Errorf("failed to create storage client: %w", err)
			// }
			// defer client.Close()

			// // Create a bucket object
			// bucket := client.Bucket(bucketName)

			// // Create an object for writing
			// imageObject := bucket.Object(dtf)
			// wc := imageObject.NewWriter(ctx)
			// if err != nil {
			// 	fmt.Errorf("failed to create object writer: %w", err)
			// }
			// defer wc.Close()

			// // Write the data to the object
			// _, err = wc.Write(imageForBucket)
			// if err != nil {
			// 	fmt.Errorf("failed to write data to object: %w", err)
			// }
			// _, err = wc.Write(promptForBucket)
			// if err != nil {
			// 	fmt.Errorf("failed to write data to object: %w", err)
			// }

			// client2, err := genai.NewClient(ctx, option.WithAPIKey(string(API_KEY)))
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// defer client.Close()
			// model := client2.GenerativeModel("gemini-1.5-flash")

			// genaiImgData1 := genai.ImageData("jpeg", imageForBucket)

			// prompt := []genai.Part{
			// 	genaiImgData1,
			// 	genai.Text(
			// 		prompt),
			// }
			// resp, err := model.GenerateContent(ctxt, prompt...)

			// ==================================================================================================================================================

			if err != nil {
				http.Error(w, "Invalid file", http.StatusBadRequest)
				fmt.Print(fileHeader)
				return
			}

			defer file.Close()

			// Create a new file in the uploads directory                                        // LLLLLLLOOOOOOOOCCCCCCCAAAAAALLLLLL
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

			// ----------------------------- THIS ONE
			model := client.GenerativeModel("gemini-1.5-flash")
			resp, err := model.GenerateContent(ctx,
				genai.FileData{URI: modelfile.URI},
				genai.Text(prompt))
			if err != nil {
				log.Fatal(err)
			}

			response := ""
			if resp.Candidates != nil {
				for _, v := range resp.Candidates {
					for _, k := range v.Content.Parts {
						response = response + string(k.(genai.Text))
					}
				}
			}

			// fmt.Fprintf(w, response)
			// fmt.Fprintf(w, "Successfully uploaded: %s\n", fileHeader.Filename)
			templ, _ := template.New("t").Parse(response)
			templ.Execute(w, nil)
		} else if flag == "vid" {
			fmt.Print("videoooooooooo")

			file, fileHeader, err := r.FormFile("video")

			if err != nil {
				http.Error(w, "Invalid file", http.StatusBadRequest)
				fmt.Print(fileHeader)
				return
			}

			defer file.Close()

			ImageNewFilePath := filepath.Join(uploadPath, dtf+".mp4")
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

			ctx := context.Background()
			// Access your API key as an environment variable
			client, err := genai.NewClient(ctx, option.WithAPIKey(string(API_KEY)))
			if err != nil {
				log.Fatal(err)
			}
			defer client.Close()

			// Use client.UploadFile to upload a file to the service.
			// Pass it an io.Reader.
			f, err := os.Open(ImageNewFilePath)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			// Optionally set a display name.
			opts := genai.UploadFileOptions{DisplayName: "dtf"}
			// Let the API generate a unique `name` for the file by passing an empty string.
			// If you specify a `name`, then it has to be globally unique.
			response, err := client.UploadFile(ctx, "", f, &opts)
			if err != nil {
				log.Fatal(err)
			}

			// View the response.
			var VideoFile *genai.File = response
			fmt.Printf("Uploaded file %s as: %q\n", VideoFile.DisplayName, VideoFile.URI)

			// Poll GetFile() on a set interval (10 seconds here) to
			// check file state.
			for response.State == genai.FileStateProcessing {
				fmt.Print(".")
				// Sleep for 10 seconds
				time.Sleep(10 * time.Second)

				// Fetch the file from the API again.
				response, err = client.GetFile(ctx, VideoFile.Name)
				if err != nil {
					log.Fatal(err)
				}
			}
			fmt.Println()

			// View the response.
			fmt.Printf("File %s is ready for inference as: %q\n",
				response.DisplayName, response.URI)

			vPrompt := []genai.Part{
				genai.FileData{URI: response.URI},
				genai.Text(prompt),
			}
			model := client.GenerativeModel("gemini-1.5-flash")
			// Generate content using the prompt.

			vResp, err := model.GenerateContent(ctx, vPrompt...)
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

			// //////////////////////////////////////////////////////////////////////////////////// working part for cloud run

			// fmt.Print("videoooooooooo")

			// file, fileHeader, err := r.FormFile("video")
			// defer file.Close()

			// // ////////////////////////////////////////////////////////////////

			// videoForBucket, _ := io.ReadAll(file)
			// promptForBucket := []byte(prompt)
			// fmt.Println(videoForBucket)
			// fmt.Println(promptForBucket)
			// if err != nil {
			// 	http.Error(w, "Invalid file", http.StatusBadRequest)
			// 	fmt.Print(fileHeader)
			// 	return
			// }

			// // ctxt := context.Background()

			// // projectID := "devfest2024-438119"
			// ctx := context.Background()
			// fmt.Println("context bg")
			// client, err := storage.NewClient(ctx)
			// fmt.Println("storage newclient")
			// if err != nil {
			// 	fmt.Errorf("storage.NewClient: %w", err)
			// }
			// defer client.Close()

			// // Replace with your project ID
			// // projectID := "devfest2024-438119"
			// bucketName := "devfest2024bucket"

			// ctx, cancel := context.WithTimeout(ctx, time.Second*50)
			// defer cancel()

			// // Upload an object with storage.Writer.
			// wc := client.Bucket(bucketName).Object("video" + dtf).NewWriter(ctx)
			// fmt.Println("wc")
			// wc.ChunkSize = 0 // note retries are not supported for chunk size 0.

			// fmt.Println("storage writer")

			// // Write the data to the object
			// _, err = wc.Write(videoForBucket)
			// if err != nil {
			// 	fmt.Errorf("failed to write data to object: %w", err)
			// }

			// fmt.Println("wc write")

			// if err := wc.Close(); err != nil {
			// 	fmt.Errorf("Writer.Close: %w", err)
			// }
			// fmt.Println("%v uploaded to %v.\n", "video"+dtf, bucketName)

			// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

			// ctx2 := context.Background()
			// client2, err := storage.NewClient(ctx2)
			// if err != nil {
			// 	fmt.Errorf("storage.NewClient: %w", err)
			// }
			// defer client2.Close()

			// ctx2, cancel2 := context.WithTimeout(ctx2, time.Second*50)
			// defer cancel2()

			// rc, err := client.Bucket(bucketName).Object("video" + dtf).NewReader(ctx2)
			// if err != nil {
			// 	fmt.Errorf("Object(%q).NewReader: %w", "video"+dtf, err)
			// }
			// defer rc.Close()

			// data, err := ioutil.ReadAll(rc)
			// if err != nil {
			// 	fmt.Errorf("ioutil.ReadAll: %w", err)
			// }
			// fmt.Println("Blob %v downloaded.\n")
			// fmt.Println(data)

			// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
			// ctx3 := context.Background()
			// fmt.Println("ctx3")

			// // Access your API key as an environment variable
			// vClient, err := genai.NewClient(ctx3, option.WithAPIKey(string(API_KEY)))
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// defer vClient.Close()
			// fmt.Println("vclient done")

			// //open?

			// // Optionally set a display name.
			// opts := genai.UploadFileOptions{DisplayName: "video" + dtf}
			// fmt.Println("opts")

			// // Let the API generate a unique `name` for the file by passing an empty string.
			// // If you specify a `name`, then it has to be globally unique.
			// // fromBucket :=
			// reader := bytes.NewReader(videoForBucket)
			// response3, err := vClient.UploadFile(ctx3, "", reader, &opts)
			// fmt.Println("response")
			// fmt.Println(response3)

			// if err != nil {
			// 	log.Fatal(err)
			// }
			// fmt.Println(response3)

			// // View the response.
			// var VideoFile *genai.File = response3
			// fmt.Printf("Uploaded file %s as: %q\n", VideoFile.DisplayName, VideoFile.URI)

			// // Poll GetFile() on a set interval (10 seconds here) to
			// // check file state.
			// for response3.State == genai.FileStateProcessing {
			// 	fmt.Print(".")
			// 	// Sleep for 10 seconds
			// 	time.Sleep(10 * time.Second)

			// 	// Fetch the file from the API again.
			// 	response3, err = vClient.GetFile(ctx3, VideoFile.Name)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}
			// }
			// fmt.Println()

			// // View the response.
			// fmt.Printf("File %s is ready for inference as: %q\n",
			// 	response3.DisplayName, response3.URI)

			// vPrompt := []genai.Part{
			// 	genai.FileData{URI: response3.URI},
			// 	genai.Text(prompt),
			// }
			// model := vClient.GenerativeModel("gemini-1.5-flash")
			// // Generate content using the prompt.

			// vResp, err := model.GenerateContent(ctx3, vPrompt...)
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// // Handle the response of generated text.
			// toDisplay := ""
			// for _, c := range vResp.Candidates {
			// 	if c.Content != nil {

			// 		for _, k := range c.Content.Parts {
			// 			toDisplay = toDisplay + string(k.(genai.Text))
			// 		}
			// 	}
			// }

			// templ, _ := template.New("t").Parse(toDisplay)
			// templ.Execute(w, nil)

		}
	}

	http.HandleFunc("/", h1)
	http.HandleFunc("/upload", h2)
	log.Fatal(http.ListenAndServe(":8080", nil))
	http.Handle("/", http.FileServer(http.Dir("/")))

}
