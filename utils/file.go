package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func GetAllFileInFolder(rootFolder string) []string {
	paths := []string{}
	folders, err := ioutil.ReadDir(rootFolder)
	if err != nil {
		return nil
	}
	for _, f := range folders {
		subFolder := rootFolder + "/" + f.Name()
		if f.IsDir() {
			paths = append(paths, GetAllFileInFolder(subFolder)...)
		} else {
			paths = append(paths, subFolder)
		}
	}
	return paths
}

func ReadFileToString(file string) (string, error) {
	sqlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(sqlFile), nil
}

func WriteToFile(value, fileName string) error {
	sqlFile := []byte(value)

	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(sqlFile)
	if err != nil {
		return err
	}
	return nil
}

// 將檔案的每一行字串拆成 []string
func TransferFileContentToSlice(fileName string) []string {
	content, err := ReadFileToString(fileName)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	return strings.Split(content, "\n")
}
