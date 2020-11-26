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
	rootCmd.AddCommand(getParameter)
	getParameter.Flags().StringVar(&folder, "folder", "", "enter folder")
	getParameter.Flags().StringVar(&ignoreFolder, "ignore_folder", "", "enter ignore folder use ',' split it")
}

var (
	getParameter = &cobra.Command{
		Use:     "get_parameter",
		Short:   "get parameter",
		Long:    `get parameter in all you print folder`,
		RunE:    runGetParameter,
		Example: "  fops get_parameter --folder [folder] --ignore_folder [ignore_folder (use ',' split)]",
	}
)

func runGetParameter(cmd *cobra.Command, args []string) error {
	start := time.Now()
	filePaths := utils.GetAllFileInFolder(folder)
	ignoreFolders := strings.Split(ignoreFolder, ",")
	if err := exportHasChineseParameter(filePaths, ignoreFolders); err != nil {
		return err
	}
	fmt.Printf("success\n%v", time.Since(start))
	return nil
}

func exportHasChineseParameter(filePaths []string, ignoreFolders []string) error {
	resultParams := make(map[string][]string)
	for _, fileName := range filePaths {
		// 忽略不需要檢查的folder
		if checkIgnoreFolder(ignoreFolders, fileName) {
			continue
		}

		lines := utils.TransferFileContentToSlice(fileName)
		params := getParamContent(lines)
		chineseParam := filterHasChineseLine(params)
		if len(chineseParam) > 0 {
			resultParams[fileName] = chineseParam
		}
	}

	return export(resultParams)
}

func export(resultParam map[string][]string) error {
	resultContent := ""
	for fileName, params := range resultParam {
		resultContent += fmt.Sprintf("file: %s\n", fileName)
		for _, param := range params {
			resultContent += fmt.Sprintf("%s\n", param)
		}
		resultContent += "\n-------------------------\n"
	}
	return utils.WriteToFile(resultContent, "parameters.txt")
}

func getChineseGlobParam(filePaths []string, ignoreFolders []string) []string {
	resultParams := []string{}

	for _, fileName := range filePaths {
		// 忽略不需要檢查的folder
		if checkIgnoreFolder(ignoreFolders, fileName) {
			continue
		}
		lines := utils.TransferFileContentToSlice(fileName)

		packageName := getPackageName(lines)
		// 取得所有參數
		params := getParamContent(lines)

		// 過濾出有中文的參數
		hasChineseParam := filterHasChineseLine(params)

		resultParams = append(resultParams, getGlobParam(trimSpace(hasChineseParam), packageName)...)
	}
	return resultParams
}

func getGlobParam(slice []string, packageName string) []string {
	parameterSlice := []string{}
	for _, s := range slice {
		r := regexp.MustCompile("(^[A-Z]+(.*))=(.*)")
		if r.MatchString(s) {
			parameter := fmt.Sprintf("%s.%s", strings.TrimSpace(packageName), r.FindStringSubmatch(s)[1])
			parameterSlice = append(parameterSlice, parameter)
		}
	}
	return parameterSlice
}

func trimSpace(slice []string) []string {
	newSlice := []string{}
	for _, s := range slice {
		newSlice = append(newSlice, strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
	}
	return newSlice
}

func filterHasChineseLine(lines []string) []string {
	filterLines := []string{}
	for _, line := range lines {
		subLine := strings.Split(line, "//")
		if len(subLine) == 2 {
			if utils.CheckChineseExist(subLine[0]) {
				filterLines = append(filterLines, subLine[0])
			}
			continue
		}
		if utils.CheckChineseExist(line) {
			filterLines = append(filterLines, line)
		}
	}
	return filterLines
}

func getParamContent(lines []string) []string {
	r := regexp.MustCompile("(var|const) \\(")
	startFlag := false
	parameterLines := []string{}
	for index, line := range lines {
		if r.MatchString(line) {
			startFlag = true
			continue
		}
		if startFlag {
			if checkParenthesesEnd(line) {
				parameterLines = append(parameterLines, getParamContent(lines[index:])...)
				break
			} else {
				if len(line) == 0 {
					continue
				}
				if ba := getBehindAnnotation(line); len(ba) > 0 {
					parameterLines = append(parameterLines, ba)
					continue
				}
				parameterLines = append(parameterLines, line)
			}
		}
	}
	return parameterLines
}

// 檢查小括弧的終點
func checkParenthesesEnd(str string) bool {
	for _, s := range str {
		if s == ')' {
			return true
		}
	}
	return false
}

// 取得註解前的字串
func getBehindAnnotation(str string) string {
	r := regexp.MustCompile("(.*)//(.*)")
	if r.MatchString(str) {
		return r.FindStringSubmatch(str)[1]
	}
	return ""
}

// 取得檔案的package名稱
func getPackageName(lines []string) string {
	r := regexp.MustCompile("^package (.*)")
	for _, line := range lines {
		if r.MatchString(line) {
			return r.FindStringSubmatch(line)[1]
		}
	}
	return ""
}
