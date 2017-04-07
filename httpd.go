package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const charWhitelist = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func fiximg(req *http.Request, time string) error {

	img, err := jpeg.Decode(req.Body)
	if err != nil {
		return err
	}

	//去掉黑边
	buf := make([]uint8, 4*img.Bounds().Dx()*img.Bounds().Dy())
	var newimg *image.RGBA = &image.RGBA{buf, 4 * img.Bounds().Dx(), img.Bounds()}
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			avarge := int(float64(r)*0.3 + float64(g)*0.51 + float64(b)*0.19)

			if avarge <= 255*50 {
				newimg.Set(x, y, img.At(x, y))
			} else {
				newimg.Set(x, y, color.White)
			}

		}
	}
	for x := 0; x < img.Bounds().Dx(); x++ {
		newimg.Set(x, 0, color.White)
	}
	for y := 0; y < img.Bounds().Dy(); y++ {
		newimg.Set(0, y, color.White)
	}

	dst, err := os.Create(time + ".jpg")
	if err != nil {
		return err
	}
	defer dst.Close()
	jpeg.Encode(dst, newimg, &jpeg.Options{Quality: 100})
	return nil
}
func exec_tesseract(time string) (string, error) {
	cmd := exec.Command("tesseract", time+".jpg", time+".temp", "-l", "eng", "-psm", "6", "-c", "tessedit_char_whitelist="+charWhitelist)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if e := cmd.Run(); e != nil {
		return "", e
	}

	buffer, _ := ioutil.ReadFile(time + ".temp.txt")
	os.Remove(time + ".temp.txt")
	os.Remove(time + ".jpg")
	return strings.TrimSpace(string(buffer)), nil

}

func main() {
	http.HandleFunc("/upload", func(w http.ResponseWriter, req *http.Request) {
		time := req.FormValue("time")
		fiximg(req, time)
		result, err := exec_tesseract(time)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte(result))
		}

	})
	http.ListenAndServe(":8092", nil)
}
