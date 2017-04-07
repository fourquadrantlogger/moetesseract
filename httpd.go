package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const charWhitelist = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func savejpg(req *http.Request) {
	time := req.FormValue("time")
	bs, _ := ioutil.ReadAll(req.Body)
	ioutil.WriteFile(time+".jpg", bs, os.ModePerm)

}
func fiximg(time string) error {
	fi, err := os.Open(time + ".jpg")
	if err != nil {
		return err
	}
	defer fi.Close()
	img, err := jpeg.Decode(fi)
	if err != nil {
		return err
	}

	//去掉黑边
	newimg := image.RGBA{img.Bounds()}
	for y := 1; y < img.Bounds().Dy(); y++ {
		for x := 1; x < img.Bounds().Dx(); x++ {
			newimg.Set(x, y, img.At(x, y))
		}
	}

	jpeg.Encode(fi, newimg, &jpeg.Options{Quality: 100})
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
	os.Remove(time + ".jpg")
	os.Remove(time + ".temp.txt")

	return strings.TrimSpace(string(buffer)), nil

}

func main() {
	http.HandleFunc("/upload", func(w http.ResponseWriter, req *http.Request) {
		time := req.FormValue("time")
		savejpg(req)
		fiximg(time)
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
