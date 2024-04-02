package main

import (
	//"archive/zip"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

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

func zipit(dir string) {
	file, err := os.Create("tmp.zip")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		fmt.Println("Crawling:", path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := w.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	err = filepath.Walk(dir, walker)
	if err != nil {
		panic(err)
	}
}

func main() {
	loadConfig()

	if len(os.Args) < 2 {
		fmt.Println("Missing filename")
		os.Exit(1)
	}

	fileName := os.Args[1]
	url := conf.Url

	a, _ := os.Stat(fileName)
	var file *os.File

	if a.IsDir() {
		zipit(fileName)
		fileName = "tmp.zip"
	}

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
		fmt.Println("Couldn't connect to server", url)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println(err.Error())
	}

	fmt.Printf("%s\n", string(body))
}
