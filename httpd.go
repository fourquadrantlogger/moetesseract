package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const charWhitelist = ""

func savejpg(req *http.Request) {
	time := req.FormValue("time")
	bs, _ := ioutil.ReadAll(req.Body)
	ioutil.WriteFile(time+".jpg", bs, os.ModePerm)
}
func exec_tesseract(time string) (string, error) {
	cmd := exec.Command("tesseract", time+".jpg", time+".txt", "-l", "eng", "-psm", "6", "-c", "tessedit_char_whitelist="+charWhitelist)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if e := cmd.Run(); e != nil {
		return "", e
	}

	buffer, _ := ioutil.ReadFile(time + ".txt")
	os.Remove(time + ".txt")

	return strings.TrimSpace(string(buffer)), nil

}

func main() {
	http.HandleFunc("/upload", func(w http.ResponseWriter, req *http.Request) {
		time := req.FormValue("time")
		savejpg(req)
		result, err := exec_tesseract(time)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(result))
	})
	http.ListenAndServe(":8092", nil)
}
