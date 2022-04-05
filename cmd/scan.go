package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var excludeFile string
var destinationDirectory string
var count int = 0

type ScanResult struct {
	Data struct {
		Advisory struct {
			Module_name         string
			Vulnerable_versions string
			Severity            string
		}
	}
}

var scanCmd = &cobra.Command{
	Use: "scan",
	Run: func(cmd *cobra.Command, args []string) {
		destinationDirectory = filepath.Dir(path) + destinationRelativePath
		excludeFile = exclude
		findPackageManagers(path)
		scanPackageManagers()
		removeCopy()
	},
}

func removeCopy() {
	err := os.RemoveAll(destinationDirectory)
	if err != nil {
		fmt.Println(err)
	}
}

func scanPackageManagers() {
	files, err := ioutil.ReadDir(destinationDirectory)
	if err != nil {
		fmt.Println(err)
	} else {
		var wg sync.WaitGroup
		c := make(chan string, count)
		for _, file := range files {
			if file.IsDir() {
				wg.Add(1)
				go func(folderName string, ch chan string, waitGroup *sync.WaitGroup) {
					defer waitGroup.Done()
					scanningResults := scanFolder(destinationDirectory + "/" + folderName)
					resultToDisplay := ""

					for _, r := range scanningResults {
						if r.Data.Advisory.Module_name != "" {
							resultToDisplay = resultToDisplay + strings.NewReplacer("{name}", r.Data.Advisory.Module_name,
								"{version}", r.Data.Advisory.Vulnerable_versions,
								"{severity}", r.Data.Advisory.Severity).Replace("valrebility found - package name:{name} verion:{version} severity:{severity} \n")
						}
					}
					ch <- resultToDisplay

				}(file.Name(), c, &wg)
			}
		}

		wg.Wait()
		close(c)

		for d := range c {
			fmt.Println(d)
		}
	}
}

func scanFolder(path string) []ScanResult {
	cmd := exec.Command("yarn", "audit", "--json")
	cmd.Dir = path
	out, _ := cmd.Output()

	allResults := []ScanResult{}
	arrayToScan := strings.Split(string(out), "\n")

	for _, res := range arrayToScan {
		var result ScanResult
		json.Unmarshal([]byte(res), &result)
		allResults = append(allResults, result)
	}
	return allResults
}

func findPackageManagers(path string) {
	if filepath.Base(path) != excludeFile {
		wasCopied := copyPackageFile(path)
		if wasCopied {
			count++
		}
	}

	subDirectories := getSubDierctories(path)
	for _, d := range subDirectories {
		findPackageManagers(path + "/" + d.Name())
	}
}

func getSubDierctories(path string) []fs.FileInfo {
	res := []fs.FileInfo{}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if file.IsDir() {
			res = append(res, file)
		}
	}

	return res
}

func copyPackageFile(path string) bool {
	inputFile := path + "/" + "package.json"
	folderName := filepath.Base(path)
	succeed := false
	_, err := os.Stat(inputFile)
	if err == nil {
		bytesRead, err := ioutil.ReadFile(inputFile)
		if err != nil {
			fmt.Println(err)
		} else {
			err := os.MkdirAll(destinationDirectory+"/"+folderName, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			} else {
				err = ioutil.WriteFile(destinationDirectory+"/"+folderName+"/package.json", bytesRead, 0755)
				if err != nil {
					fmt.Println(err)
				} else {
					succeed = true
				}
			}
		}
	}

	return succeed
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.PersistentFlags().StringVarP(&path, pathFlagName, "", "", "directory to scan")
	rootCmd.PersistentFlags().StringVarP(&exclude, excludeFlagName, "", "", "folder name to avoid in the scan procces")
	rootCmd.MarkPersistentFlagRequired(pathFlagName)
}
