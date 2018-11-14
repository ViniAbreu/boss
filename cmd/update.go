package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/hashload/boss/consts"
	"github.com/spf13/cobra"
	"gopkg.in/cheggaaa/pb.v2"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

var publishCmd = &cobra.Command{
	Use:   "update",
	Short: "update a cli",
	Long:  `update a cli`,
	Run: func(cmd *cobra.Command, args []string) {
		err := rerunDetached()
		if err != nil {
			err.Error()
		}
		err, link, size, version := getLink()
		if err != nil {
			err.Error()
		}

		if !checkVersion(version) {
			return
		}

		exePath, err := filepath.Abs(os.Args[0])
		if err != nil {
			err.Error()
		}
		if err := downloadFile(exePath /*+ "_o"*/, link, size); err != nil {
			os.Remove(exePath + "_o")
			err.Error()
		} else {
			os.Remove(exePath)
			os.Rename(exePath+"_o", exePath)
		}
	},
}

func rerunDetached() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	for _, arg := range os.Args {
		if arg == "--detached" {
			println("LOL")
			return nil
		}
	}

	args := append(os.Args, "--detached")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	err = cmd.Start()
	if err != nil {
		return err
	}
	cmd.Process.Release()
	return nil
}

func checkVersion(newVersion string) bool {
	version, _ := semver.NewVersion(newVersion)
	current, _ := semver.NewVersion(consts.VERSION)
	needUpdate := version.GreaterThan(current)
	if needUpdate {
		println(consts.VERSION, " -> ", newVersion)
	} else {
		println("already up to date!")
	}
	return needUpdate
}

func getLink() (err error, link string, size float64, version string) {
	resp, err := http.Get("https://api.github.com/repos/HashLoad/boss/releases/latest")
	if err != nil {
		return err, "", 0, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status), "", 0, ""
	}
	contents, err := ioutil.ReadAll(resp.Body)
	var obj map[string]interface{}
	json.Unmarshal(contents, &obj)
	bossExe := obj["assets"].([]interface{})[0].(map[string]interface{})
	return nil, bossExe["browser_download_url"].(string), bossExe["size"].(float64), obj["name"].(string)
}

func downloadFile(filepath string, url string, size float64) (err error) {
	os.Remove(filepath)
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	bar := pb.New(int(math.Round(size)))
	bar.Start()
	proxyReader := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(out, proxyReader)
	bar.Finish()
	if err != nil {
		return err
	}

	return nil
}

func init() {
	//RootCmd.AddCommand(publishCmd)
}
