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
			if p.monitor == monitor {
				prevImage = p.paper
				break
			}
		}

		if prevImage != nextImage {
			unloadWallpaper(prevImage)
			preloadWallpaper(nextImage)
			setWallpaper(nextImage, monitor)
			writeConfig()
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
		if p.monitor == monitor {
			prevImage = p.paper
			break
		}
	}

	if prevImage != nextImage {
		unloadWallpaper(prevImage)
		preloadWallpaper(nextImage)
		setWallpaper(nextImage, monitor)
		writeConfig()
	}

}

func listActive() ([]*plane, error) {
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

	var planes []*plane

	for scanner.Scan() {
		kv := strings.Split(scanner.Text(), "=")
		if len(kv) > 1 {
			planes = append(planes, &plane{monitor: strings.TrimSpace(kv[0]), paper: strings.TrimSpace(kv[1])})
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
	fmt.Printf("unload: %s\n", image)
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

func writeConfig() {
	base := os.Getenv("HOME")

	configfile := fmt.Sprintf("%s/.config/hypr/hyprpaper.conf", base)

	fmt.Printf("writing: %s\n\n", configfile)
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

	defer fmt.Println("ok")

	hist := configText()

	// log changes
	historyfile := fmt.Sprintf("%s/wallpaper/hyprpaperplanes.log", base)
	file, err := os.OpenFile(historyfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.Println(jsonText())
	// -

	fmt.Fprint(f, hist)
}

func rewind(n int) {
	base := os.Getenv("HOME")
	historyfile := fmt.Sprintf("%s/wallpaper/hyprpaperplanes.log", base)
	file, err := os.Open(historyfile)

	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	var past []history
	for scanner.Scan() {

		line := scanner.Text()
		if len(line) > 0 {
			idx := strings.IndexRune(line, '[')
			if idx == -1 {
				// catch for single monitor
				idx = strings.IndexRune(line, '{')
			}

			if idx >= 0 {
				past = append(past, history{dt: line[:idx], data: line[idx:]})
			}
		}

	}
	current := len(past) - 1

	var target int

	if current >= 0 {
		if len(past)-n > 0 {
			target = current - n

		} else {
			target = current
		}
	}
	hist := past[target]

	for i, v := range hist.unfold() {
		fmt.Println(i, v)
		preloadWallpaper(v.paper)
		setWallpaper(v.paper, v.monitor)
		// note, the config file isn't being written here
	}
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
