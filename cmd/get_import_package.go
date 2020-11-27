/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/gfes980615/check_tool/utils"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
	"time"
)

func init() {
	rootCmd.AddCommand(getImportPackage)
	getImportPackage.Flags().StringVar(&folder, "folder", "", "enter folder")
	getImportPackage.Flags().StringVar(&ignoreFolder, "ignore_folder", "", "enter ignore folder use ',' split it")
}

var (
	getImportPackage = &cobra.Command{
		Use:     "get_import_package",
		Short:   "get import package",
		Long:    `get import packager`,
		RunE:    runGetImportPackage,
		Example: "  fops get_import_package --folder [folder] --ignore_folder [ignore_folder (use ',' split)]",
	}
	importMap = make(map[string]bool)
)

func runGetImportPackage(cmd *cobra.Command, args []string) error {
	start := time.Now()
	filePaths := utils.GetAllFileInFolder(folder)
	ignoreFolders := strings.Split(ignoreFolder, ",")
	for _, fileName := range filePaths {
		// 忽略不需要檢查的folder
		if checkIgnoreFolder(ignoreFolders, fileName) {
			continue
		}
		lines := utils.TransferFileContentToSlice(fileName)
		setImportContentMap(lines)
	}

	importSlice := []string{}
	for im := range importMap {
		importSlice = append(importSlice, im)
	}
	if err := utils.WriteToFile(strings.Join(importSlice, "\n"), "import.txt"); err != nil {
		return err
	}
	fmt.Printf("success\n%v", time.Since(start))
	return nil
}

func setImportContentMap(lines []string) {
	r := regexp.MustCompile("import \\(")
	startFlag := false
	for _, line := range lines {
		if r.MatchString(line) {
			startFlag = true
			continue
		}
		if startFlag {
			if checkParenthesesEnd(line) {
				break
			} else {
				if len(line) == 0 {
					continue
				}
				if ba := getBehindAnnotation(line); len(ba) > 0 {
					importMap[ba] = true
					continue
				}
				importMap[line] = true
			}
		}
	}
}
