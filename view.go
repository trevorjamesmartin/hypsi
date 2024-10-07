package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
)

func jsonText() string {
	active, err := listActive()
	if err != nil {
		log.Fatal(err)
	}
	bs, err := json.Marshal(active)
	if err != nil {
		log.Fatal(err)
	}

	return string(bs)
}

func configText() string {
	var text string
	sources := make(map[string]bool)
	activeplanes, err := listActive()
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range activeplanes {
		if preloaded, _ := sources[p.Paper]; !preloaded {
			text += fmt.Sprintf("preload = %s\n", p.Paper)
			sources[p.Paper] = true
		}
	}

	for _, p := range activeplanes {
		text += fmt.Sprintf("wallpaper = %s,%s\n", p.Monitor, p.Paper)
	}
	text += "splash = false\n"
	return text
}

func hyperText(w io.Writer, i int) {
	activeplanes, errListing := listActive()

	if errListing != nil {
		log.Fatal(errListing)
	}

	tmpl := `
	<!DOCTYPE html>
	<html>
	  <head>
	    <title>HyprPaperPlanes API {{.Version}}</title>
	      <style>
		* {
			background-color: black;
			color: white;
		}

		.formats {
			display: flex;
			justify-content: space-evenly;
			font-size: 1.5rem;
			letter-spacing: 1px;
			line-height: 4rem;
		}

		div.gallery {
		  border: 5px solid white;
		  border-radius: 12px;
		  margin: 5px;
		  line-height: 4rem;
		  text-align:center; 
		  color: white; 
		  width: 640px; 
		  height: 360px;
  		  background-size: 640px 360px;
		}

		div.gallery:hover {
		  border: 5px solid #777;
		  border-radius: 12px;
		}

		div.desc {
		  padding: 0.25rem;
		  text-align: center;
		  font-size: 2rem; 
		}

		div.monitors {
		  display: flex;
		  flex-direction: row;
		  justify-content: space-around;
		  align-items: stretch;
		  margin-top: 120px;
		}

		div.rewind {
		  height: 150px;
		  width: auto;
		  padding: 60px;
		  font-size: 2rem;
		}
		button {
		  font-size: 2rem;
		  border-radius: 5px;
		}

		.container {
		  display: flex;
		  flex-direction: column;
		  justify-content: space-evenly;
		  align-items: center;
		}
	      </style>
	    </head>
	  <body>
	    {{ if .Ivalue }}
		<div class="formats">
		  <a href="/hyprpaper.conf">hyprpaper.conf</a>
		  <a href="/json">JSON</a>
		</div>
	    {{ end }}
	    <div class="container">
	      <div class="monitors">
		{{ range .Monitors}}
		<div class="{{.Monitor}} gallery" ondrop="handleDrop(event)" ondragover="allowDrop(event)">
			<div class="desc">
			{{.Monitor}}
			</div>

			<form id="form_{{.Monitor}}" enctype="multipart/form-data" action="/upload/{{.Monitor}}" method="post" hidden>
			  <input type="file" name="imageFile" />
			  <input id="send_{{.Monitor}}" type="submit" value="upload" />
			</form>
		</div>
		<style>
		.{{.Monitor}} {
			background-image: url({{.ToBase64|safeURL}});
		};
		</style>
		{{end}}
	      </div>
		{{ if .Ivalue }}
		<div class="rewind">
		  <button onclick="handleRewind(event)">⏮ Previous</button>
		  {{.Rewind}}
		  <button onclick="handleForward(event)">Next</button>
		</div>
		{{ else }}
		<hr>
		{{ end }}
	    </div>
	    <script type="text/javascript">
		var countme = localStorage.getItem("rewind") || 0;
		async function sendData(data, url) {
			  try {
			    const response = await fetch(url, {
			      method: "POST",
			      body: data,
			    });
			    console.log(await response.statusText);
			    location.reload()
			  } catch (e) {
			    console.error(e);
			  }
		}
		function handleDrop(event) {
		  event.preventDefault();
		  if (event.dataTransfer.items) {
		    let filecount = 0;
		    const fileLimit = 1;
		    [...event.dataTransfer.items].forEach((item, i) => {
		      // If dropped items aren't files, reject them
		      if (item.kind === "file" && filecount < fileLimit) {
			filecount++;
			const file = item.getAsFile();
			const firstClass = event.target.className.split(" ").shift();
			const sendTo = "/upload?monitor=" + firstClass;
			const formSelector = "#form_" + firstClass;
			const formElement = document.querySelector(formSelector);
			
			// Take over form submission
			formElement.addEventListener("submit", (event) => {
			  event.preventDefault();
			  // Associate the FormData object with the form element
			  const formData = new FormData(formElement);
			  formData.set("imageFile", file);
			  sendData(formData, sendTo);
			  localStorage.setItem("rewind", 0)
			}, false);

			let sendBtnSelector = "#send_" + firstClass;
			document.querySelector(sendBtnSelector).click()
		      }
		    });
		  } else {
		    [...event.dataTransfer.files].forEach((file, i) => { console.log('?', file, i); });
		  }
		}
		function allowDrop(event) {
		  event.preventDefault();
		}
		function handleRewind(event) {
		  countme++;
		  event.preventDefault();
		  localStorage.setItem("rewind", countme)
		  console.log("⏮ Rewind", countme);
		  const url = new URL(location);
		  url.pathname = "/rewind";
		  url.searchParams.set("t", String(countme));
		  history.pushState( {}, "", url)
		  setTimeout(() => history.go(), 300)
		}
		function handleForward(event) {
		  if (countme <= 0) {
		    countme = 0;
		    return
		  }
		  countme--;
		  event.preventDefault();
		  localStorage.setItem("rewind", countme)
		  console.log("⏮ Rewind", countme);
		  const url = new URL(location);
		  url.pathname = "/rewind";
		  url.searchParams.set("t", String(countme));
		  history.pushState( {}, "", url)
		  setTimeout(() => history.go(), 300)
		}
	    </script>
	  </body>
	</html>`

	funcMap := template.FuncMap{
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
	}

	data := struct {
		Version  string
		Monitors []*Plane
		Ivalue   bool
		Rewind   int
	}{VERSION, activeplanes, i >= 0, i}

	template.Must(template.New("webpage").Funcs(funcMap).Parse(tmpl)).Execute(w, data)
}
