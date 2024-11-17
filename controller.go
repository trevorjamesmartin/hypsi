package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		nextImage := argsWithoutProg[0]

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
			preloadWallpaper(fname)
			setWallpaper(nextImage, monitor)
			writeConfig(true)
		}
	}

}

func readFromWeb(monitor string, filename string) {
	activeplanes, err := listActive()

	if err != nil {
		fmt.Println(err)
		return
	}

	// file exists
	nextImage := filename

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
func setWallpaper(image string, monitor string) {
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

	Map := make(map[string]string)
	err = json.Unmarshal(buf, &Map)
	monitor := Map["monitor"]
	return string(monitor)
}

func makeThumbNail(image string, thumb string) {
	cmd := exec.Command("magick", "-define", "jpeg:size=640x360",
		image, "-thumbnail", "230400@", "-gravity", "center",
		"-background", "black", "-extent", "640x360", thumb)

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

	tempFile, err := os.Create(filepath.Join(tempFolder, fname))

	if err != nil {
		fmt.Println(err)
	}

	filename := tempFile.Name()
	fmt.Println("Saving to : ", filename)

	defer tempFile.Close()

	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
	} else {
		tempFile.Write(fileBytes)
		fmt.Println("Successfully Downloaded File")
		monitor := activeMonitor()
		defer readFromWeb(monitor, filename)
	}
}
