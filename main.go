package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Token string
	Url   string
}

var conf Config

func loadConfig() {
	userConfDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println(err.Error())
	}

	path := userConfDir + "/dingo/dingo.toml"
	_, err = toml.DecodeFile(path, &conf)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func main() {
	loadConfig()

	fileName := os.Args[1]
	url := conf.Url

	file, err := os.Open(fileName)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	client := &http.Client{}
	buf := new(bytes.Buffer)
	bw := multipart.NewWriter(buf)
	fbw, err := bw.CreateFormFile("fil", fileName)
	if err != nil {
		println(err.Error())
	}

	io.Copy(fbw, file)
	bw.Close()

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		println(err.Error())
	}

	req.Header.Set("Authorization", conf.Token)
	req.Header.Add("Content-Type", bw.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		println(err.Error())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println(err.Error())
	}

	fmt.Printf("%s\n", string(body))
}
