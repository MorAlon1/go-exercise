package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
)

var excludeFile string
var pathToCopy string

var scanCmd = &cobra.Command{
	Use: "scan",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		pathToCopy = filepath.Dir(path) + "/toScan"
		excludeFile, _ = cmd.Flags().GetString("exclude")
		count := findPackageManagers(path, 0)
		scanPackageManagers(count)
		removeCopy()
	},
}

func removeCopy() {
	err := os.RemoveAll(pathToCopy)
	if err != nil {
		fmt.Println(err)
	}
}

func scanPackageManagers(count int) {
	files, _ := ioutil.ReadDir(pathToCopy)
	var wg sync.WaitGroup
	c := make(chan string, count)
	for _, file := range files {
		if file.IsDir() {
			wg.Add(1)
			go func(folderName string, ch chan string, waitGroup *sync.WaitGroup) {
				defer waitGroup.Done()
				res := scanFolder(pathToCopy+"/"+folderName, ch)
				ch <- res
			}(file.Name(), c, &wg)
		}
	}

	wg.Wait()
	close(c)

	for d := range c {
		fmt.Println(d)
	}

}

func scanFolder(path string, channel chan string) string {
	cmd := exec.Command("yarn", "audit")
	cmd.Dir = path
	out, _ := cmd.Output()
	return string(out)
}

func findPackageManagers(path string, count int) int {
	if filepath.Base(path) != excludeFile {
		copyPackageFile(path)
		count++
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	for _, v := range files {
		if v.IsDir() {
			count = findPackageManagers(path+"/"+v.Name(), count)
		}
	}

	return count
}

func copyPackageFile(path string) {
	filepathToCopy := path + "/" + "package.json"
	_, err := os.Stat(filepathToCopy)
	if err == nil {
		bytesRead, _ := ioutil.ReadFile(filepathToCopy)
		err := os.MkdirAll(pathToCopy+"/"+filepath.Base(path), os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
		err = ioutil.WriteFile(pathToCopy+"/"+filepath.Base(path)+"/package.json", bytesRead, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.PersistentFlags().String("path", "", "where")
	rootCmd.PersistentFlags().String("exclude", "", "where not")
}
