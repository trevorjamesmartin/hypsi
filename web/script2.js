const countElement = () => document.querySelector("#svgtext");
const rewindBtn = () => document.querySelector("#roll_back");
const forwardBtn = () => document.querySelector("#roll_forth");
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
      switch (String(idx).length) {
	  case 1:
	    countElement().setAttribute("x", "20");
	    break;
	  case 2:
	    countElement().setAttribute("x", "10");
	    break;
	  default:
	    countElement().setAttribute("x", "0");
	    break;
      }

      const limit = Number(result.limit);

      result.monitors?.forEach(i => {
        let styleElement = document.querySelector("#style_" + i.monitor);
        styleElement.textContent = "." + i.monitor + " { background-image: url(" + i.image + "); }";
      });
      const textElement = document.querySelector("#svgtext");
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

      if (prev - next === 1) {
        backbtn.setAttribute("disabled", true);
      }

      if (next < 0) {
        forthbtn.setAttribute("disabled", true);
      }

    });
  } else {
    console.error("missing RollBack() code");
  }
}


document.addEventListener("DOMContentLoaded", () => {

    rewindBtn().addEventListener("click", (event) => {
      const idx = Number(countElement().innerHTML);
      handleGoRewind(event, idx + 1);
    });

    forwardBtn().addEventListener("click", (event) => {
      const idx = Number(countElement().innerHTML);
      handleGoRewind(event, idx - 1);
    });

});
