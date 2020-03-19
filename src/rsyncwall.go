package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/disintegration/imaging"
)

func main() {
	var anchor imaging.Anchor
	// TODO: check if exist this directory
	home, err := os.UserHomeDir()
	os.MkdirAll(home+"/.config/syncwall/", os.ModePerm)
	// Open a test image.
	// src, err := imaging.Open("testdata/flowers.png")

	_desktop := os.Args[1]
	_anchor := os.Args[2]
	_wallpaper := os.Args[3]

	if _desktop == "" {
		_desktop = "x11"
	}

	if _wallpaper == "--listen" {
		log.Println("Waiting...")
		s, err := net.ResolveUDPAddr("udp4", "255.255.255.255:3411")
		c, err := net.ListenUDP("udp4", s)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer c.Close()

		buffer := make([]byte, 128)
		n, _, err := c.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}
		_wallpaper = string(buffer[0:n])

	} else {
		s, err := net.ResolveUDPAddr("udp4", "255.255.255.255:3411")
		c, err := net.ListenUDP("udp4", s)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer c.Close()

		_, err = c.WriteTo([]byte(_wallpaper), s)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

	}

	res, err := http.Get(_wallpaper)
	if err != nil {
		log.Fatalf("failed to download image: %v", err)
	}
	// img, err := ioutil.ReadAll(res.Body)

	src, err := imaging.Decode(res.Body)
	if err != nil {
		log.Fatalf("failed to read image: %v", err)
	}
	res.Body.Close()
	// Crop the original image to 300x300px size using the center anchor.
	switch _anchor {
	case "left":
		anchor = imaging.Left
	case "center":
		anchor = imaging.Center
	case "right":
		anchor = imaging.Right
	}
	src = imaging.CropAnchor(src, 1920, 1080, anchor)

	// Create a new image and paste the four produced images into it.
	dst := imaging.New(1920, 1080, color.NRGBA{0, 0, 0, 0})
	dst = imaging.Paste(dst, src, image.Pt(0, 0))

	// Save the resulting image as JPEG.
	err = imaging.Save(dst, home+"/.config/syncwall/current.jpg")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}

	// Change desktop wallpaper
	switch _desktop {
	case "x11":
		cmd := exec.Command("pcmanfm", "--set-wallpaper", home+"/.config/syncwall/current.jpg")
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Command finished with error: %v", err)
		}
	case "weston":
		cmd := exec.Command("killall", "-s", "KILL", "sighup", "weston")
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Command finished with error: %v", err)
		}
	}
}
