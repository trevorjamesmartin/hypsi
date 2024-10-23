package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func readFromCLI(argsWithoutProg []string) {
	// path to wallpaper ?
	_, err := os.Stat(argsWithoutProg[0])
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
			preloadWallpaper(nextImage)
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

func setWallpaper(image string, monitor string) {
	fmt.Printf("set wallpaper: %s,'%s'\n", monitor, image)

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
