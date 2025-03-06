package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
https://invent.kde.org/plasma/plasma-workspace/-/blob/master/wallpapers/image/plasma-apply-wallpaperimage.cpp

	QMap<QString, int> wallpaperFillModeTypes = {
	    {QStringLiteral("stretch"), 0},
	    {QStringLiteral("preserveAspectFit"), 1},
	    {QStringLiteral("preserveAspectCrop"), 2},
	    {QStringLiteral("tile"), 3},
	    {QStringLiteral("tileVertically"), 4},
	    {QStringLiteral("tileHorizontally"), 5},
	    {QStringLiteral("pad"), 6},
	};
*/
var FillModeTypes = map[string]int{
	"stretch":            0,
	"preserveAspectFit":  1,
	"contain":            1, // hyprland
	"preserveAspectCrop": 2,
	"cover":              2, // hyprland
	"tile":               3,
	"tileVertically":     4,
	"tileHorizontally":   5,
	"pad":                6,
}

// Hyprland compatible FillModeTypes
var hyprModeTypes = map[string]string{
	"0": "cover",
	"1": "contain",
	"2": "cover",
	"3": "cover",
	"4": "cover",
	"5": "cover",
	"6": "cover",
}

// $ xrandr --listactivemonitors
func xrandrMonitors() ([]string, error) {
	var result []string
	cmd := exec.Command("xrandr", "--listactivemonitors")

	stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()

	if err != nil {
		return result, err
	}
	scanner := bufio.NewScanner(stdout)

	if scanner.Scan() {
		// first line shows the number of active monitors,
		// this number will equal the length of the return value
		// ...
		_ = scanner.Text() // redundant
	}

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		spc := strings.Split(line, " ")
		monitorName := spc[len(spc)-1]
		result = append(result, monitorName)
	}

	if scanner.Err() != nil {
		return result, scanner.Err()
	}

	return result, nil
}

// $ qdbus org.kde.plasmashell /PlasmaShell org.kde.PlasmaShell.wallpaper N
func plasmaGetWallpapers(names []string) ([]*Plane, error) {
	var result []*Plane

	for idx, monitor := range names {
		cmd := exec.Command("qdbus", "org.kde.plasmashell", "/PlasmaShell", "org.kde.PlasmaShell.wallpaper", fmt.Sprintf("%d", idx))
		stdout, _ := cmd.StdoutPipe()

		err := cmd.Start()

		if err != nil {
			return result, err
		}

		scanner := bufio.NewScanner(stdout)
		pln := &Plane{Monitor: monitor, Mode: "cover"}

		for scanner.Scan() {

			kv := strings.Split(scanner.Text(), " ")

			if len(kv) > 1 && strings.HasPrefix(kv[0], "Image:") {
				pln.Paper = strings.Trim(kv[1], " ")[7:]
			}

			if len(kv) > 1 && strings.HasPrefix(kv[0], "FillMode:") {
				pln.Mode = hyprModeTypes[strings.Trim(kv[1], "")]
			}

		}
		result = append(result, pln)

		if scanner.Err() != nil {
			return result, scanner.Err()
		}
	}
	return result, nil
}

// xrandrMonitors() + plasmaGetWallpapers()
func plasmaListActive() ([]*Plane, error) {
	var result []*Plane
	var err error

	monitors, err := xrandrMonitors() // list of monitor names

	if err != nil {
		log.Fatal(err)
	}

	result, err = plasmaGetWallpapers(monitors) // wallpapers (apply monitor names)

	if err != nil {
		log.Fatal(err)
	}

	return result, err
}

// basically '/usr/bin/plasma-apply-wallpaperimage', per monitor
func plasmaWallpaperCommand(image, monitor, mode string) error {
	if strings.Contains(image, "'") {
		// "If this happens, we might very well assume that there is some kind of funny business going on even if technically it could just be a possessive. But, security first, so..."
		log.Fatalf("\nThere is a stray single quote in the filename of this wallpaper (') - please contact the author of the wallpaper to fix this, or rename the file yourself: %s\n", image)
	}

	var screen int

	result, err := plasmaListActive()

	if err != nil {
		fmt.Println("plasmaWallpaperCommand > plasmaListActive ERROR")
		return err
	}

	for idx, value := range result {
		if value.Monitor == monitor {
			screen = idx
		}
	}

	fmt.Printf("screen %d > apply wallpaper: %s\n", screen, image)

	scriptText := fmt.Sprintf(`const dt = desktops()[%d];`, screen)
	scriptText += fmt.Sprintf(`dt.currentConfigGroup = Array("Wallpaper", "org.kde.image", "General"); dt.writeConfig("Image", "file://%s");`, image)
	scriptText += fmt.Sprintf(`dt.writeConfig("FillMode", %d);`, FillModeTypes[mode])
	scriptText += `dt.reloadConfig();`

	err = exec.Command("qdbus", "org.kde.plasmashell", "/PlasmaShell", "org.kde.PlasmaShell.evaluateScript", scriptText).Run()

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// set the fill-mode, per monitor
func plasmaSetWallpaperMode(monitor, mode string) error {
	var screen int
	var image string

	result, err := plasmaListActive()

	if err != nil {
		log.Fatal(err)
	}

	for idx, value := range result {
		if value.Monitor == monitor {
			screen = idx
			image = value.Paper
		}
	}

	fmt.Printf("screen %d > apply mode: %s\n", screen, mode)

	setModeSetting(monitor, filepath.Base(image), mode)

	if err != nil {
		log.Fatal(err)
	}

	scriptText := fmt.Sprintf(`const dt = desktops()[%d];`, screen)
	scriptText += fmt.Sprintf(`dt.currentConfigGroup = Array("Wallpaper", "org.kde.image", "General"); dt.writeConfig("Image", "file://%s");`, image)
	scriptText += fmt.Sprintf(`dt.writeConfig("FillMode", %d);`, FillModeTypes[mode])
	scriptText += `dt.reloadConfig();`

	err = exec.Command("qdbus", "org.kde.plasmashell", "/PlasmaShell", "org.kde.PlasmaShell.evaluateScript", scriptText).Run()

	if err != nil {
		log.Fatal(err)
	}

	return nil

}

// initialize webview data structures
func plasmaWebInit(page Webpage) Webpage {
	monitorList, err := plasmaListActive()

	if err != nil {
		log.Fatal(err)
	}

	page.data.Monitors = monitorList

	var hardware []*HyprMonitor

	for idx, value := range monitorList {
		hardware = append(hardware, &HyprMonitor{Name: value.Monitor, Id: idx})
	}
	page.data.Hardware = hardware

	return page
}
