async function sendData(data, url) {
  try {
    const response = await fetch(url, {
      method: "POST",
      body: data,
    });
    console.log(await response.statusText);
    location.reload();
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
        formElement.addEventListener(
          "submit",
          (event) => {
            event.preventDefault();
            // Associate the FormData object with the form element
            const formData = new FormData(formElement);
            formData.set("imageFile", file);
            //localStorage.setItem("rewind", 0);
            sendData(formData, sendTo);
          },
          false,
        );

        let sendBtnSelector = "#send_" + firstClass;
        document.querySelector(sendBtnSelector).click();
      }
    });
  } else {
    [...event.dataTransfer.files].forEach((file, i) => {
      console.log("?", file, i);
    });
  }
}

function allowDrop(event) {
  event.preventDefault();
}

function handleRewind(event, idx) {
  event.preventDefault();

  if (idx < 0) {
    return;
  }

  const url = new URL(location);
  url.pathname = "/rewind";
  url.searchParams.set("t", String(idx));
  history.pushState({}, "", url);
  setTimeout(() => history.go(), 300);
}

function handleGoRewind(event, idx) {
  if (typeof event?.preventDefault != "function") {
    console.error("don't call this function directly");
    return
  }
  event.preventDefault();

  if (idx < 0) {
    return;
  }

  if (typeof RollBack == "function") {
    RollBack(idx).then((result) => {
      const limit = Number(result.limit);

      result.monitors?.forEach(i => {
        let styleElement = document.querySelector("#style_" + i.monitor);
        styleElement.textContent = "." + i.monitor + " { background-image: url(" + i.image + "); }";
      });
      const textElement = document.querySelector("#rewindtext");
      const backbtn = document.querySelector("#roll_back")
      const forthbtn = document.querySelector("#roll_forth");
      backbtn?.removeAttribute("disabled");
      forthbtn?.removeAttribute("disabled");

      if (idx >= 0) {
        textElement.textContent = idx; 
      }

      let curr = Number(result.rewind || idx);
      let next = curr - 1;
      let prev = curr + 1;

      if (prev >= limit) {
        prev = limit - 1;
      }

      backbtn?.setAttribute("onclick", `handleGoRewind(event, ${prev})`);

      if (prev - next === 1) {
        backbtn.setAttribute("disabled", true);
      }

      forthbtn?.setAttribute("onclick", `handleGoRewind(event, ${next})`);

      if (next < 0) {
        forthbtn.setAttribute("disabled", true);
      }

    });
  } else {
    console.error("missing RollBack() code");
  }
}

function openFileChooser(monitor) {
  let selector = "#imageFile_" + monitor;
  document.querySelector(selector)?.click();
}

