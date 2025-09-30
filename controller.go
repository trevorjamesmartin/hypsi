package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"

	//"image/png"
	//"image/gif"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MaestroError/go-libheif"
	"github.com/adrg/xdg"
	"github.com/trevorjamesmartin/resize"
	//"golang.org/x/image/webp"
)

// url should return JSON format
func FetchJSON(url string) json.RawMessage {
	valid, _ := regexp.MatchString("(((https?)://)([-%()_.!~*';/?:@&=+$,A-Za-z0-9])+)", url)

	if !valid {
		return nil
	}

	fmt.Println(url)

	response, err := http.Get(url)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	return responseData
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func readEnvFile(path string) (string, error) {
	file, err := os.Open(path)
	var result string
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)

	result += fmt.Sprintf("[environment: %s]", path)

	for scanner.Scan() {
		kv := strings.Split(scanner.Text(), "=")
		if len(kv) == 2 {
			result += fmt.Sprintf("\n%s : %s", kv[0], kv[1])
			os.Setenv(kv[0], kv[1])
		}
	}

	if scanner.Err() != nil {
		return result, scanner.Err()
	}

	return result, nil
}

func ReadInput(args []string) {
	onWeb, _ := regexp.MatchString("(((https?)://)([-%()_.!~*';/?:@&=+$,A-Za-z0-9])+)", args[0])
	if onWeb {
		monitors := args[1:]
		DownloadImage(args[0], monitors)
	} else {
		readFromCLI(args)
	}
}

func getContentType(fname string) (string, error) {
	var contentType string

	file, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buf := make([]byte, 1024)

	_, err = file.Read(buf)

	if err != nil {
		return contentType, err
	}
	contentType = http.DetectContentType(buf)

	ext := filepath.Ext(fname)

	if contentType == "application/octet-stream" {
		// base content on file extension ?
		switch ext {
		case ".avif":
			contentType = "image/avif"
		default:
			// do nothing
		}
	}

	return contentType, nil
}

func readFromCLI(argsWithoutProg []string) {
	var mode string
	currentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")

	// mode ?
	var fname string
	x := strings.Split(argsWithoutProg[0], `:`)

	if len(x) == 2 {
		mode = x[0]
		fname = x[1]
	} else {
		fname = argsWithoutProg[0]
	}

	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		fmt.Printf("not a valid background image: [ %s ]", fname)
		return
	}

	fname, _ = filepath.Abs(fname)

	contentType, err := getContentType(fname)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(contentType)

	var activeplanes []*Plane

	switch currentDesktop {
	case "KDE":
		activeplanes, err = plasmaListActive()
	default:
		activeplanes, err = listActive()
	}

	if err != nil {
		log.Fatal(err)
		return
	}

	// file exists
	var nextImage string

	// can Hyprland read this image file?
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
		if strings.HasPrefix(contentType, "image") {
			nextImage = fname
		} else {
			log.Fatalf("Unsupported mime type %v\n", contentType)
		}
	}

	monitor := activeMonitor()

	var prevImage string

	for _, p := range activeplanes {
		if p.Monitor == monitor {
			prevImage = p.Paper
			break
		}
	}

	switch currentDesktop {
	case "KDE":
		fname := filepath.Base(nextImage)

		if len(mode) == 0 {
			mode = getModeSetting(monitor, fname)
		}

		err = plasmaWallpaperCommand(nextImage, monitor, mode)
		if err != nil {
			log.Fatal(err)
			return
		}
	default:
		if prevImage != nextImage {
			unloadWallpaper(prevImage)
			preloadWallpaper(nextImage)
			fname := filepath.Base(nextImage)
			mode = getModeSetting(monitor, fname)

			if err = setWallpaper(nextImage, monitor, mode); err != nil {
				log.Fatal(err)
				return
			}

		}
	}

	HYPSI_STATE.SetRewind(0)
	writeConfig(true)

}

// readFromWeb Function (webview) - read from webview
func readFromWeb(monitor, filename string) {
	currentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
	var prevImage, mode string
	nextImage := filename

	switch currentDesktop {
	case "KDE":
		mode = getModeSetting(monitor, filename)
		err := plasmaWallpaperCommand(filename, monitor, mode)
		if err != nil {
			log.Fatal(err)
			return
		}
	case "XFCE":
		mode = getModeSetting(monitor, filename)
		err := xfce4WallpaperCommand(filename, monitor, mode)
		if err != nil {
			log.Fatal(err)
			return
		}
	default:
		// Hyprland
		activeplanes, err := listActive()
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range activeplanes {
			if p.Monitor == monitor {
				prevImage = p.Paper
				break
			}
		}
		if prevImage != nextImage {
			unloadWallpaper(prevImage)
			preloadWallpaper(nextImage)
			fname := filepath.Base(nextImage)
			mode = getModeSetting(monitor, fname)

			if err = setWallpaper(nextImage, monitor, mode); err != nil {
				log.Fatal(err)
				HYPSI_STATE.SetMessage(err.Error())
				return
			}

		}
	}
	HYPSI_STATE.SetRewind(0)
	writeConfig(true)

}

// $ hyprctl monitors -j
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

// $ hyprctl hyprpaper listactive
func listActive() ([]*Plane, error) {
	switch os.Getenv("XDG_CURRENT_DESKTOP") {
	case "KDE":
		return plasmaListActive()
	case "XFCE":
		return xfce4ListActive()
	default:
		// continue
	}

	cmd := exec.Command("hyprctl", "hyprpaper", "listactive")
	// note: "mode" setting is not displayed by this command

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

// unloadWallpaper Function
// $ hyprctl hyprpaper unload {image}
func unloadWallpaper(image string) {
	if os.Getenv("XDG_CURRENT_DESKTOP") == "XFCE" {
		return
	}

	if os.Getenv("XDG_CURRENT_DESKTOP") == "KDE" {
		return
	}
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
	if os.Getenv("XDG_CURRENT_DESKTOP") == "KDE" {
		return
	}
	if os.Getenv("XDG_CURRENT_DESKTOP") == "XFCE" {
		return
	}
	extension := filepath.Ext(image)

	switch extension {
	case ".avif", ".heif":
		fmt.Printf("Hyprland can't handle %s\n", extension)
		return
	default:
		fmt.Printf("preload: %s\n", image)
	}

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
func setWallpaper(image, monitor, mode string) error {
	var cmd *exec.Cmd
	var err error

	extension := filepath.Ext(image)

	switch extension {
	case ".avif", ".heif":
		fmt.Printf("Hyprland can't handle %s\n", extension)
		return nil
	default:
		switch mode {
		case "cover", "":
			// "monitor,image"
			fmt.Printf("set wallpaper: %s,%s\n", monitor, image)
			cmd = exec.Command("hyprctl", "hyprpaper", "wallpaper", fmt.Sprintf("%s,%s", monitor, image))
		default:
			// monitor,mode:image
			fmt.Printf("set wallpaper: %s,%s:%s\n", monitor, mode, image)
			cmd = exec.Command("hyprctl", "hyprpaper", "wallpaper", fmt.Sprintf("%s,%s:%s\n", monitor, mode, image))
		}
	}

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatal(err)

	}

	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()

	if err != nil {
		log.Fatal(err)
	}

	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		fmt.Println(text)

		if strings.HasPrefix(text, "wallpaper failed") {
			err = errors.New(text)
		}

		if strings.HasPrefix(text, "couldn't connect to") {
			err = errors.New(text)
		}
	}
	cmd.Wait()

	if scanner.Err() != nil {
		cmd.Process.Kill()
		return scanner.Err()
	}

	return err
}

// => $ hyprctl hyprpaper listactive
func monitorFilename(monitor string) string {
	var fname string
	monitors, err := listActive()

	if err != nil {
		return err.Error()
	}

	for _, p := range monitors {
		if p.Monitor == monitor {
			fname = filepath.Base(p.Paper)
			break
		}
	}

	return fname
}

func setWallpaperMode(monitor, mode string) {
	currentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")

	if currentDesktop == "KDE" {
		plasmaSetWallpaperMode(monitor, mode)
		return
	}

	if currentDesktop == "XFCE" {
		// todo
		return
	}

	monitors, err := listActive()

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, p := range monitors {
		if p.Monitor == monitor && p.Mode != mode {
			// save the mode adjustment
			fname := filepath.Base(p.Paper)
			if getModeSetting(monitor, fname) != mode {
				unloadWallpaper(p.Paper)
				preloadWallpaper(p.Paper)
				setModeSetting(p.Monitor, fname, mode)
				// update wallpaper
				if err = setWallpaper(p.Paper, p.Monitor, mode); err != nil {
					HYPSI_STATE.SetMessage(err.Error())
					return
				}
			}
			break
		}
	}
}

// hyprpaper.conf
func writeConfig(historical bool) {
	if os.Getenv("XDG_CURRENT_DESKTOP") == "XFCE" {
		if historical {
			writeHistory()
		}
		return
	}
	configfile := fmt.Sprintf("%s/hypr/hyprpaper.conf", xdg.ConfigHome)

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

	if len(past) == 0 {
		return false, 0
	}

	if grief != nil {
		log.Fatal(grief)
		return false, 0
	}

	current := len(past) - 1

	var target History

	if len(past)-n > 0 {
		target = past[current-n]
	} else {
		target = past[current]
	}

	for _, v := range target.unfold() {

		currentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
		fname := filepath.Base(v.Paper)
		mode := getModeSetting(v.Monitor, fname)
		switch currentDesktop {
		case "KDE":
			plasmaWallpaperCommand(v.Paper, v.Monitor, mode)
		case "XFCE":
			xfce4WallpaperCommand(v.Paper, v.Monitor, mode)
		default:
			preloadWallpaper(v.Paper)
			// update wallpaper
			if err := setWallpaper(v.Paper, v.Monitor, mode); err != nil {
				log.Fatal(err)
			}
		}
		// note, the config file isn't being written here
	}
	return true, len(past)
}

func activeMonitor() string {
	currentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
	var monitorName string

	switch currentDesktop {
	case "KDE":
		monitors, err := xrandrMonitors()
		if err != nil {
			log.Fatal(err)
		}
		// TODO: provide this information ?
		// for now,
		// just return the first monitor listed
		monitorName = monitors[0]
	case "XFCE":
		monitors, err := xrandrMonitors()
		if err != nil {
			log.Fatal(err)
		}
		monitorName = monitors[0]

	default:
		// Hyprland's hyprctl
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

		monitorName = ws.Monitor
	}
	return monitorName
}

func makeThumbNail(inputPath, thumb string) {
	var err error

	file, err := os.Open(inputPath)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	var img image.Image
	var fmat string
	img, fmat, err = image.Decode(file)

	if err != nil {
		fmt.Printf(`makeThumbNail ERROR [imagePath: "%s", decode: "%s"]\n`, inputPath, fmat)
		log.Fatal(err)
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

// download image, then set wallpaper
func DownloadImage(validURL string, monitors []string) {
	resp, err := http.Get(validURL)

	if err != nil {
		log.Fatal(err)
	}

	ext := filepath.Ext(validURL)

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

	fname := fmt.Sprintf("%x", bs)

	if len(ext) > 0 {
		fname += ext
	}

	if tempFile, err := os.Create(filepath.Join(HYPSI_STATE.GetStorePath(), fname)); err != nil {
		log.Fatal(err)
	} else {
		filename := tempFile.Name()
		fmt.Println("Saving to : ", filename)

		defer tempFile.Close()
		tempFile.Write(fileBytes)
		fmt.Println("Successfully Downloaded File")

		if len(monitors) == 0 {
			monitors = []string{activeMonitor()}
		}

		defer func() {
			for _, monitor := range monitors {
				readFromWeb(monitor, filename)
			}
		}()
	}
}
