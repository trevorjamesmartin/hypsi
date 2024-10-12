 
var countme = localStorage.getItem("rewind") || 0;

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
            sendData(formData, sendTo);
            localStorage.setItem("rewind", 0);
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

  localStorage.setItem("rewind", idx);
  const url = new URL(location);
  url.pathname = "/rewind";
  url.searchParams.set("t", String(idx));
  history.pushState({}, "", url);
  setTimeout(() => history.go(), 300);
}

