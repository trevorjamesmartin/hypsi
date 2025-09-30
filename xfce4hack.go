package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// $ xconf-query -c xfce-4desktop -p /backdrop/screen0/monitor[NAME]/workspace0/last-image
func xfce4GetWallpapers(names []string) ([]*Plane, error) {
	var result []*Plane

	for _, monitor := range names {
		cmd := exec.Command("xfconf-query", "-c", "xfce4-desktop", "-p", fmt.Sprintf("/backdrop/screen0/monitor%s/workspace0/last-image", monitor))
		stdout, _ := cmd.StdoutPipe()

		err := cmd.Start()

		if err != nil {
			return result, err
		}

		scanner := bufio.NewScanner(stdout)
		pln := &Plane{Monitor: monitor, Mode: "cover"}

		for scanner.Scan() {
			// only expecting 1 line of output
			pln.Paper = scanner.Text()
		}

		result = append(result, pln)

		if scanner.Err() != nil {
			return result, scanner.Err()
		}
	}
	return result, nil
}

// xrandrMonitors() + xfce4GetWallpapers()
func xfce4ListActive() ([]*Plane, error) {
	var result []*Plane
	var err error

	monitors, err := xrandrMonitors() // list of monitor names

	if err != nil {
		log.Fatal(err)
	}

	result, err = xfce4GetWallpapers(monitors) // wallpapers (apply monitor names)

	if err != nil {
		log.Fatal(err)
	}

	return result, err
}

// initialize webview data structures
func xfce4WebInit(page Webpage) Webpage {
	monitorList, err := xfce4ListActive()

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

// xfconf-query -c xfce4-desktop -p /backdrop/screen0/monitor0/workspace0/last-image -s /path/to/your/image.jpg
func xfce4WallpaperCommand(image, monitor, mode string) error {
	if strings.Contains(image, "'") {
		// "If this happens, we might very well assume that there is some kind of funny business going on even if technically it could just be a possessive. But, security first, so..."
		log.Fatalf("\nThere is a stray single quote in the filename of this wallpaper (') - please contact the author of the wallpaper to fix this, or rename the file yourself: %s\n", image)
	}

	fmt.Printf("monitor: %s > apply wallpaper: %s (%s)\n", monitor, image, mode)

	err := exec.Command("xfconf-query", "-c", "xfce4-desktop", "-p", fmt.Sprintf("/backdrop/screen0/monitor%s/workspace0/last-image", monitor), "-s", image).Run()

	if err != nil {
		log.Fatal(err)
	}

	return nil
}
