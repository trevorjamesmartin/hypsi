package main

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/MaestroError/go-libheif"
)

//go:embed web/*
var WEBFOLDER embed.FS

func newFile(fname string) (*os.File, error) {
	// todo: configure this for customization
	folderName := "wallpaper"
	return os.Create(filepath.Join(fmt.Sprintf("%s/%s", os.Getenv("HOME"), folderName), fname))
}

func api() {
	var port string
	port = os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	serverAddress := fmt.Sprintf("0.0.0.0:%s", port)

	mux := http.NewServeMux()
	page := webInit()

	user_template := os.Getenv("HYPSI_WEBVIEW")

	user_html_template := os.Getenv("HYPSI_WEBPAGE")

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		if len(user_html_template) > 0 {
			f, err := os.ReadFile(user_html_template)
			if err != nil {
				log.Fatal(err)
				return
			}
			page.template = string(f)
		} else {
			page.template = page._Template()
		}

		t := r.URL.Query().Get("t")
		if n, err := strconv.Atoi(t); err != nil {
			page.Print(w, HYPSI_STATE.GetRewind())
		} else {
			page.Print(w, n)
		}
	})

	mux.HandleFunc("GET /webview", func(w http.ResponseWriter, r *http.Request) {
		if len(user_template) > 0 {
			f, err := os.ReadFile(user_template)
			if err != nil {
				log.Fatal(err)
				return
			}
			page.template = string(f)
		} else {
			default_template, _ := WEBFOLDER.ReadFile("web/webview.html.tmpl")
			page.template = string(default_template)
		}
		page.data.Rewind = HYPSI_STATE.GetRewind()
		page.Print(w, HYPSI_STATE.GetRewind())
	})

	mux.HandleFunc("GET /static/{filename}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, WEBFOLDER, fmt.Sprintf("web/%s", r.PathValue("filename")))
	})

	handleConfig := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(configText()))
	}

	handleJSON := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(jsonText()))
	}

	uploadFile := func(w http.ResponseWriter, r *http.Request) {
		var filename string
		r.ParseMultipartForm(10 << 20)

		monitor := r.FormValue("monitor")

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
		fmt.Printf("Monitor: %s\n", monitor)

		fileBytes, err := io.ReadAll(file)

		h := sha256.New()
		h.Write(fileBytes)
		bs := h.Sum(nil)
		fmt.Printf("sha256: %x\n", bs)

		if err != nil {
			fmt.Println(err)
		}

		ext := filepath.Ext(handler.Filename)
		fname := fmt.Sprintf("%x%s", bs, ext)

		// convert unsupported file types
		switch filepath.Ext(fname) {
		case ".heic", ".heif":
			// write 2 files
			originalFile, err := newFile(fname)
			filename = fmt.Sprintf("%s.jpg", originalFile.Name())

			if err != nil {
				log.Fatal(err)
			}

			originalFile.Write(fileBytes)
			originalFile.Close()

			// convert to jpeg for hyprpaper
			err = libheif.HeifToJpeg(originalFile.Name(), filename, 100)
			if err != nil {
				fmt.Println("ERROR")
				fmt.Println(err)
			} else {
				fmt.Println(filename)
				fmt.Fprintf(w, "Successfully Uploaded File\n")
				defer readFromWeb(monitor, filename)
			}

		default:
			tempFile, err := newFile(fname)
			if err != nil {
				log.Fatal(err)
			}
			filename = tempFile.Name()
			fmt.Println(filename)

			defer tempFile.Close()

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

	}
	mux.HandleFunc("POST /upload", uploadFile)
	mux.HandleFunc("GET /conf", handleConfig)
	mux.HandleFunc("GET /config", handleConfig)
	mux.HandleFunc("GET /configuration", handleConfig)
	mux.HandleFunc("GET /hyprpaper.conf", handleConfig)
	mux.HandleFunc("GET /json", handleJSON)
	mux.HandleFunc("GET /rewind", func(w http.ResponseWriter, r *http.Request) {
		t := r.URL.Query().Get("t")
		n, err := strconv.Atoi(t)
		if err != nil {
			http.Redirect(w, r, "/rewind?t=0", http.StatusSeeOther)
			return
		}
		rewind(n)
		page.Print(w, n)
	})

	mux.HandleFunc("GET /interrupt", func(w http.ResponseWriter, r *http.Request) {
		os.Exit(0)
	})

	server := http.Server{Addr: serverAddress, Handler: mux}
	fmt.Printf("[ listening @ http://%s ] ", serverAddress)
	fmt.Println("...")
	server.ListenAndServe()
}
