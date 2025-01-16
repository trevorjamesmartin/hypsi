package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MaestroError/go-libheif"
	"github.com/trevorjamesmartin/resize"
)

func readFromCLI(argsWithoutProg []string) {
	// mode ?
	var fname string
	x := strings.Split(argsWithoutProg[0], `:`)

	if len(x) == 2 {
		fname = x[1]
	} else {
		fname = argsWithoutProg[0]
	}

	// path to wallpaper ?
	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		fmt.Println("not a valid background image")
	} else {
		activeplanes, err := listActive()

		if err != nil {
			fmt.Println(err)
			return
		}

		// file exists
		var nextImage string

		switch filepath.Ext(fname) {
		case ".heic", ".heif":
			newfile := fmt.Sprintf("%s.jpg", fname)

			err = libheif.HeifToJpeg(fname, newfile, 100)

			if err != nil {
				log.Fatal(err)
			}
			nextImage = newfile

		case ".bmp":
			log.Fatal("Unsupported file type")
		default:
			nextImage = fname
		}

		monitor := activeMonitor()

		var prevImage string

		for _, p := range activeplanes {
			if p.Monitor == monitor {
				prevImage = p.Paper
				break
			}
		}

		if prevImage != nextImage {
			unloadWallpaper(prevImage)
			preloadWallpaper(nextImage)
			setWallpaper(nextImage, monitor)
			writeConfig(true)
		}
	}

}

// readFromWeb Function (webview) - read from webview
func readFromWeb(monitor, filename string) {
	if activeplanes, err := listActive(); err != nil {
		log.Fatal(err)
	} else {
		var prevImage string
		nextImage := filename
		for _, p := range activeplanes {
			if p.Monitor == monitor {
				prevImage = p.Paper
				break
			}
		}

		if prevImage != nextImage {
			unloadWallpaper(prevImage)
			preloadWallpaper(nextImage)
			setWallpaper(nextImage, monitor)
			writeConfig(true)
		}
	}
}

func listMonitors() ([]*HyprMonitor, error) {

	cmd := exec.Command("hyprctl", "monitors", "-j")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	var mons []*HyprMonitor

	var rawjson string

	for scanner.Scan() {
		rawjson += scanner.Text()
	}

	if scanner.Err() != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return nil, scanner.Err()
	}

	err = json.Unmarshal([]byte(rawjson), &mons)

	if err != nil {
		fmt.Println("ERROR", err)
		return mons, err
	}

	return mons, nil

}

func listActive() ([]*Plane, error) {
	cmd := exec.Command("hyprctl", "hyprpaper", "listactive")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	var planes []*Plane

	for scanner.Scan() {
		kv := strings.Split(scanner.Text(), "=")
		if len(kv) > 1 {
			planes = append(planes, &Plane{Monitor: strings.TrimSpace(kv[0]), Paper: strings.TrimSpace(kv[1])})
		}
	}

	if scanner.Err() != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return nil, scanner.Err()
	}

	return planes, nil
}

func listUnassigned() ([]*HyprMonitor, error) {
	var m, result []*HyprMonitor

	active, err := listActive()
	available := make(map[string]*HyprMonitor)

	if err != nil {
		return m, err
	}

	m, err = listMonitors()

	if err != nil {
		return m, err
	}

	for _, m := range m {
		available[m.Name] = m
	}

	for _, a := range active {
		_, ok := available[a.Monitor]
		if ok {
			available[a.Monitor] = nil
			delete(available, a.Monitor)
		}
	}

	for k := range available {
		result = append(result, available[k])
	}
	return result, nil
}

// unloadWallpaper Function
// $ hyprctl hyprpaper unload {image}
func unloadWallpaper(image string) {
	cmd := exec.Command("hyprctl", "hyprpaper", "unload", image)
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatal(err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()

	if err != nil {
		log.Fatal(err)
	}

	var textLog string

	for scanner.Scan() {
		textLog += scanner.Text()
	}

	if scanner.Err() != nil {
		cmd.Process.Kill()
		cmd.Wait()
		log.Fatal(scanner.Err())
		return
	}

	if textLog != "ok" {
		fmt.Println(textLog)
	}

}

// preloadWallpaper Function
// $ hyprctl hypaper preload {image}
func preloadWallpaper(image string) {
	fmt.Printf("preload: %s\n", image)

	cmd := exec.Command("hyprctl", "hyprpaper", "preload", image)
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatal(err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()

	if err != nil {
		log.Fatal(err)
	}

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if scanner.Err() != nil {
		cmd.Process.Kill()
		cmd.Wait()
		log.Fatal(scanner.Err())
		return
	}

}

// setWallpaper Function
// $ hyprctrl hyprpaper wallpaper {monitor},{image}
func setWallpaper(image, monitor string) {
	fmt.Printf("set wallpaper: %s,%s\n", monitor, image)

	cmd := exec.Command("hyprctl", "hyprpaper", "wallpaper", fmt.Sprintf("%s,%s", monitor, image))
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatal(err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()

	if err != nil {
		log.Fatal(err)
	}

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if scanner.Err() != nil {
		cmd.Process.Kill()
		cmd.Wait()
		log.Fatal(scanner.Err())
		return
	}

}

func setWallpaperMode(monitor, mode string) {
	monitors, err := listActive()

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, p := range monitors {
		if p.Monitor == monitor && p.Mode != mode {
			unloadWallpaper(p.Paper)

			preloadWallpaper(p.Paper)
			if mode == "cover" {
				// default mode
				setWallpaper(p.Paper, p.Monitor)
			} else {
				setWallpaper(fmt.Sprintf("%s:%s", mode, p.Paper), p.Monitor)
			}

			break
		}
	}
}

func writeConfig(historical bool) {
	base := os.Getenv("HOME")

	configfile := fmt.Sprintf("%s/.config/hypr/hyprpaper.conf", base)

	// remove old file if it exists
	if errRemoving := os.Remove(configfile); errRemoving != nil {
		fmt.Println(errRemoving)
	}

	// create new
	f, err := os.Create(configfile)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	fmt.Fprint(f, configText())

	if historical {
		writeHistory()
	}

}

func rewind(n int) (bool, int) {
	past, grief := readHistory()

	if n > len(past) {
		return false, 0
	}

	if grief != nil {
		log.Fatal(grief)
		return false, 0
	}

	current := len(past) - 1

	var target History

	if current >= 0 {
		if len(past)-n > 0 {
			target = past[current-n]

		} else {
			target = past[current]
		}
	}

	for _, v := range target.unfold() {
		preloadWallpaper(v.Paper)
		setWallpaper(v.Paper, v.Monitor)
		// note, the config file isn't being written here
	}
	return true, len(past)
}

func activeMonitor() string {
	buf, err := exec.Command("hyprctl", "activeworkspace", "-j").CombinedOutput()

	if err != nil {
		fmt.Println(err)
		return ""
	}

	var ws HyprCtlActiveWorkspace

	if err = json.Unmarshal(buf, &ws); err != nil {
		fmt.Println("ERROR:")
		fmt.Println(buf)
		log.Fatal(err)
	}

	return ws.Monitor
}

func makeThumbNail(inputPath, thumb string) {
	var err error

	file, err := os.Open(inputPath)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	var img image.Image

	switch filepath.Ext(inputPath) {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
	case ".png":
		img, err = png.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
	case ".gif":
		img, err = gif.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
	case ".bmp":
		// img, err := bmp.Decode(file)
		// if err != nil {
		// 	log.Fatal(err)
		// }
	default:
		log.Fatal("Unsupported file type")
	}

	m := resize.Thumbnail(640, 360, img, resize.Lanczos3)

	out, err := os.Create(thumb)

	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = jpeg.Encode(out, m, nil)

	if err != nil {
		log.Fatal(err)
	}

}

type HyprCtlVersion struct {
	Branch          string   `json:"branch"`
	Commit          string   `json:"commit"`
	CommitMessage   string   `json:"commit_message"`
	CommitDate      string   `json:"commit_date"`
	Tag             string   `json:"tag"`
	Commits         string   `json:"commits"`
	BuildAquamarine string   `json:"buildAquamarine"`
	Flags           []string `json:"flags,omitempty"`
	Dirty           bool     `json:"dirty"`
}

func hyprCtlVersion() (HyprCtlVersion, error) {
	var hyprCtlVersiion HyprCtlVersion
	buf, err := exec.Command("hyprctl", "version", "-j").CombinedOutput()

	if err != nil {
		fmt.Println(err)
		return hyprCtlVersiion, err
	}

	err = json.Unmarshal(buf, &hyprCtlVersiion)
	if err != nil {
		log.Fatal(err)
	}
	return hyprCtlVersiion, nil
}

func fileExtensionFromURL(validURL string) string {
	u, _ := url.Parse(validURL)
	_, fname := filepath.Split(u.Path)
	arr := strings.Split(fname, ".")
	return arr[len(arr)-1]
}

func downloadImage(validURL string) {
	resp, err := http.Get(validURL)

	if err != nil {
		log.Fatal(err)
	}

	ext := fileExtensionFromURL(validURL)

	defer resp.Body.Close()
	fileBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	h := sha256.New()
	h.Write(fileBytes)
	bs := h.Sum(nil)

	fmt.Printf("sha256: %x\n", bs)
	fmt.Println("Status Code: ", resp.StatusCode)
	fmt.Println("Content Length: ", resp.ContentLength)

	tempFolder := fmt.Sprintf("%s/wallpaper", os.Getenv("HOME"))

	fname := fmt.Sprintf("%x", bs)

	if len(ext) > 0 {
		fname += fmt.Sprintf(".%s", ext)
	}

	if tempFile, err := os.Create(filepath.Join(tempFolder, fname)); err != nil {
		log.Fatal(err)
	} else {
		filename := tempFile.Name()
		fmt.Println("Saving to : ", filename)

		defer tempFile.Close()
		tempFile.Write(fileBytes)
		fmt.Println("Successfully Downloaded File")
		monitor := activeMonitor()
		defer readFromWeb(monitor, filename)
	}
}
