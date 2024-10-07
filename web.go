package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func api() {
	UPLOADS := fmt.Sprintf("%s/wallpaper", os.Getenv("HOME"))
	// check for existing wallpaper folder
	ok, _ := exists(UPLOADS)
	if !ok {
		errorMessage := fmt.Sprintf(`NOTE: this option requires writing uploaded images to "%s" please ensure the folder exists`, UPLOADS)
		log.Fatal(errorMessage)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		t := r.URL.Query().Get("t")

		i, err := strconv.Atoi(t)

		if err != nil {
			hyperText(w, 0)
			return
		}
		hyperText(w, i)
	})

	handleConfig := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(configText()))
	}

	handleJSON := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(jsonText()))
	}

	uploadFile := func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)
		file, handler, err := r.FormFile("imageFile")

		if err != nil {
			fmt.Println("ERROR getting file")
			fmt.Println(err)
			return
		}
		defer file.Close()

		fmt.Printf("Uploaded File: %+v\n", handler.Filename)
		fmt.Printf("File Size: %+v\n", handler.Size)
		fmt.Printf("MIME Header: %+v\n", handler.Header)

		monitor := r.URL.Query().Get("monitor")

		fmt.Printf("Monitor: %s\n", monitor)

		tempFolder := fmt.Sprintf("%s/wallpaper", os.Getenv("HOME"))

		tempFile, err := os.CreateTemp(tempFolder, "hyprPaperPlane_*-"+handler.Filename)

		if err != nil {
			fmt.Println(err)
		}

		filename := tempFile.Name()
		fmt.Println(filename)

		defer tempFile.Close()
		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := io.ReadAll(file)

		if err != nil {
			fmt.Println("ERROR")
			fmt.Println(err)
		} else {
			// write this byte array to our temporary file
			tempFile.Write(fileBytes)
			// return that we have successfully uploaded our file!
			fmt.Fprintf(w, "Successfully Uploaded File\n")
			defer readFromWeb(monitor, filename)
		}
	}
	mux.HandleFunc("POST /upload", uploadFile)

	mux.HandleFunc("GET /conf", handleConfig)
	mux.HandleFunc("GET /config", handleConfig)
	mux.HandleFunc("GET /configuration", handleConfig)
	mux.HandleFunc("GET /hyprpaper.conf", handleConfig)

	mux.HandleFunc("GET /json", handleJSON)

	mux.HandleFunc("GET /rewind", func(w http.ResponseWriter, r *http.Request) {
		pages := r.URL.Query().Get("t")

		i, err := strconv.Atoi(pages)

		if err != nil {
			http.Redirect(w, r, "/rewind?t=0", http.StatusSeeOther)
			return
		}

		fmt.Println(pages)
		rewind(i)
		hyperText(w, i)
	})

	server := http.Server{Addr: ":3000", Handler: mux}
	fmt.Println("Listening @ http://0.0.0.0:3000")
	server.ListenAndServe()
}
