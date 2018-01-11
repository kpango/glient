package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/kpango/glg"
	"github.com/kpango/glient"
)

func main() {
	glient.Init(nil)

	start := time.Now()
	res, err := http.Get("https://github.com/kpango")
	glg.Infof("http Default Request\t:\t%v", time.Since(start))
	if err != nil {
		glg.Error(err)
		return
	}
	res.Body.Close()

	start = time.Now()
	res, err = glient.Get("https://github.com/kpango")
	glg.Infof("First Request\t:\t%v", time.Since(start))
	if err != nil {
		glg.Error(err)
		return
	}
	res.Body.Close()

	start = time.Now()
	res, err = glient.Get("https://github.com/kpango")
	glg.Infof("Second Request Uses Cache DNS Resoleved IP\t:\t%v", time.Since(start))
	if err != nil {
		glg.Error(err)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glg.Error(err)
		return
	}

	ioutil.WriteFile("kpango.html", body, os.ModePerm)
}
