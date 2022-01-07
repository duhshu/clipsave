package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/viper"
)

var (
	confFile = "config.yaml"
	confExt  = "yaml"
	confDir  = ".clisave"
	pathKey  = "path"
)

func main() {
	if err := Run(); err != nil {
		log.Fatalln(err)
	}
}

func Run() error {
	var savePath string
	var err error

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	confPath := path.Join(home, confDir)
	viper.SetConfigType(confExt)
	viper.AddConfigPath(confPath)

	// 引数のpathを取得
	flag.Parse()
	savePath = flag.Arg(0)

	// 前回実行時の保存先をread
	if savePath == "" {
		savePath, err = readPath(confPath)
		if err != nil {
			return err
		}
	} else {
		// 保存先をconfigにsave
		viper.Set(pathKey, savePath)
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}

	if savePath == "" {
		return errors.New("please input the save path")
	}

	fmt.Println(savePath)

	// clipの中身をsave
	saveFilePath := path.Join(savePath, "out.png")
	f, err := os.Create(saveFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Println("readClip")
	clip, err := readFromClipboard()
	if err != nil {
		return err
	}
	fmt.Println("Copy")
	if _, err := io.Copy(f, clip); err != nil {
		return err
	}

	return nil
}

func readPath(confPath string) (string, error) {

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("config not found")
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {

			if err := os.Mkdir(confPath, 0755); err != nil {
				return "", err
			}
			if _, err := os.Create(path.Join(confPath, confFile)); err != nil {
				return "", err
			}
		}
	}
	return viper.GetString(pathKey), nil
}

func readFromClipboard() (io.Reader, error) {
	// b has new line
	cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return nil, err
	}
	fmt.Println("cmd copy", buf.Len())

	if err := r.Close(); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return buf, nil
}
