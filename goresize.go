package main

import (
	"appengine-go/example/moustachio/resize"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	logf, err := os.OpenFile("info.txt", os.O_APPEND|os.O_CREATE, 0640)
	if err != nil {
		log.Fatal(err)
	}
	defer logf.Close()

	log.SetOutput(logf)
	mlog := log.New(io.MultiWriter(logf, os.Stdout), log.Prefix(), log.Flags())

	target := "resized"
	if err := os.Mkdir(target, 0755); os.IsNotExist(err) {
		mlog.Fatalln(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		mlog.Fatalln(err)
	}

	d, err := os.Open(wd)
	if err != nil {
		mlog.Fatalln(err)
	}
	defer d.Close()

	files, err := d.Readdirnames(0)
	if err != nil {
		mlog.Fatalln(err)
	}
	d.Close()

	for _, file := range files {
		lcase := strings.ToLower(file)
		if strings.HasSuffix(lcase, ".jpeg") || strings.HasSuffix(lcase, ".jpg") {
			mlog.Println(file)

			ir, err := os.Open(file)
			if err != nil {
				mlog.Fatalln(err)
			}
			defer ir.Close()

			hasher := md5.New()
			io.Copy(hasher, ir)
			newName := hex.EncodeToString(hasher.Sum(nil))

			mlog.Println(newName)

			if _, err = ir.Seek(0, 0); err != nil {
				mlog.Fatalln(err)
			}

			i, err := jpeg.Decode(ir)
			if err != nil {
				mlog.Fatalln(err)
			}

			ir.Close()

			// Taken from: https://code.google.com/p/appengine-go/source/browse/example/moustachio/moustachio/http.go
			// License: BSD-style
			// Resize if too large, for more efficient moustachioing.
			// We aim for less than 1200 pixels in any dimension; if the
			// picture is larger than that, we squeeze it down to 600.
			const max = 1200
			// If it's gigantic, it's more efficient to downsample first
			// and then resize; resizing will smooth out the roughness.
			if b := i.Bounds(); b.Dx() > 640 || b.Dy() > 480 {
				if b.Dx() > 2*max || b.Dy() > 2*max {
					w, h := max, max
					if b.Dx() > b.Dy() {
						h = b.Dy() * h / b.Dx()
					} else {
						w = b.Dx() * w / b.Dy()
					}
					i = resize.Resample(i, i.Bounds(), w, h)
					b = i.Bounds()
				}
				i = resize.Resize(i, i.Bounds(), 640, 480)
			}

			out, err := os.Create(filepath.FromSlash(fmt.Sprintf("%s/%s.jpg", target, newName)))
			if err != nil {
				mlog.Fatalln(err)
			}
			defer out.Close()

			jpeg.Encode(out, i, nil)
			out.Close()
		}
	}
}
