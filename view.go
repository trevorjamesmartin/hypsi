package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

func dropFileScript() string {
	return fmt.Sprintf(`
	<script type="text/javascript">

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
		const sendTo = "/upload?monitor=" + event.target.className;
		const formSelector = "#form_" + event.target.className;
		const formElement = document.querySelector(formSelector)
		
		// Take over form submission
		formElement.addEventListener("submit", (event) => {
		  event.preventDefault();
		  // Associate the FormData object with the form element
		  const formData = new FormData(formElement);
		  formData.set("imageFile", file);
		  sendData(formData, sendTo);
		});

		let sendBtnSelector = "#send_" + event.target.className;
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


	</script>
	`)
}

func hyperText() string {
	var hypertext string
	activeplanes, errListing := listActive()

	if errListing != nil {
		log.Fatal(errListing)
	}

	hypertext += fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
	<title>HyprPaperPlanes API %s</title>
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
		}

		div.gallery {
		  border: 5px solid white;
		  border-radius: 12px;
		  margin-left: auto;
		  margin-right: auto;
		  margin-top: 80px;
		  line-height: 8rem; 
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
	</style>
	%s
	</head>
	<body>

		<div class="formats">
		<a href="/hyprpaper.conf">hyprpaper.conf</a>
		<a href="/json">JSON</a>
		</div>
	`, VERSION, dropFileScript())

	for _, p := range activeplanes {
		bts, err := os.ReadFile(p.Paper)

		if err != nil {
			log.Fatal(err)
		}

		data := fmt.Sprintf("data:%s;base64,%s",
			http.DetectContentType(bts),
			base64.StdEncoding.EncodeToString(bts))

		form := fmt.Sprintf(`<form id="form_%s" enctype="multipart/form-data" action="/upload/%s" method="post" hidden>
			    <input type="file" name="imageFile" />
		  	    <input id="send_%s" type="submit" value="upload" />
			</form>`, p.Monitor, p.Monitor, p.Monitor)

		hypertext += fmt.Sprintf(`
		<div class="%s gallery" ondrop="handleDrop(event)" ondragover="allowDrop(event)">
			
			<div class="desc">
				%s
			</div>

			%s
		</div>
		<style>
		.%s {
			background-image: url(%s);
		};
		</style>`, p.Monitor, p.Monitor, form, p.Monitor, data)
	}

	hypertext += `</body></html>`
	return hypertext
}
