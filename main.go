package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
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

func zipit(dir string) string {
	cache_dir, err := os.UserCacheDir()
	if err != nil {
		panic(err.Error())
	}

	zip_file := filepath.Join(cache_dir, "dingo.zip")
	file, err := os.Create(zip_file)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	walker := func(path string, info os.FileInfo, err error) error {
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

	return zip_file
}

func file_upload(fileNames []string) {
	file, err := os.Open(fileNames[0])
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	buf := new(bytes.Buffer)

	io.Copy(buf, file)
	uploadBuf(*buf, path.Ext(fileNames[0]))
}

func noPipe() {
	if len(os.Args) < 2 {
		fmt.Println("Missing filename")
		os.Exit(1)
	}

	file_names := os.Args[1:]

	f, _ := os.Stat(file_names[0])

	if f.IsDir() {
		zip_name := zipit(file_names[0])
		file_upload([]string{zip_name}) // temp fix
		err := os.Remove(zip_name)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// It's not a dir
		file_upload(file_names)
	}
}

func uploadBuf(buf_upload bytes.Buffer, file_extension string) {
	client := &http.Client{}
	buf := new(bytes.Buffer)
	bw := multipart.NewWriter(buf)

	fbw, err := bw.CreateFormFile("fil", file_extension)
	if err != nil {
		println(err.Error())
	}

	io.Copy(fbw, bytes.NewReader(buf_upload.Bytes()))
	bw.Close()

	req, err := http.NewRequest("POST", conf.Url, buf)
	if err != nil {
		println(err.Error())
	}

	req.Header.Set("Authorization", conf.Token)
	req.Header.Add("Content-Type", bw.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Couldn't connect to server", conf.Url)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println(err.Error())
	}

	fmt.Printf("%s\n", string(body))
}

func main() {
	loadConfig()

	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err.Error())
	}

	if (fi.Mode() & os.ModeNamedPipe) != 0 {
		buf := new(bytes.Buffer)
		io.Copy(buf, os.Stdin)
		uploadBuf(*buf, "") // tmp fix
	} else {
		noPipe()
	}

}
