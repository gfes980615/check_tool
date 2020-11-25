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
	filePaths := utils.GetAllFileInFolder(folder)
	ignoreFolders := strings.Split(ignoreFolder, ",")
	return utils.WriteToFile(strings.Join(getHasChineseParameter(filePaths, ignoreFolders), "\n"), "parameter.txt")
}

func getHasChineseParameter(filePaths []string, ignoreFolders []string) []string {
	resultParams := []string{}

	for _, fileName := range filePaths {
		if checkIgnoreFolder(ignoreFolders, fileName) {
			continue
		}
		lines := fileContentToLines(fileName)
		packageName := getPackageName(lines)
		parameters := getParameterContent(lines)
		hasChineseParam := filterHasChineseLine(parameters)
		resultParams = append(resultParams, setParameter(trimSpace(hasChineseParam), packageName, fileName)...)
	}
	return resultParams
}

func setParameter(slice []string, packageName, fileName string) []string {
	parameterSlice := []string{}
	for _, s := range slice {
		r := regexp.MustCompile("(^[A-Z]+(.*))=(.*)")
		if r.MatchString(s) {
			parameter := fmt.Sprintf("%s.%s", strings.TrimSpace(packageName), r.FindStringSubmatch(s)[1])
			parameterSlice = append(parameterSlice, parameter)
		}
		//r2 := regexp.MustCompile("(^[a-z]+(.*))=(.*)")
		//if r2.MatchString(s) {
		//	parameter := r2.FindStringSubmatch(s)[1]
		//	parameterSlice = append(parameterSlice, parameter)
		//}
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
			if checkChineseExist(subLine[0]) {
				filterLines = append(filterLines, subLine[0])
			}
			continue
		}
		if checkChineseExist(line) {
			filterLines = append(filterLines, line)
		}
	}
	return filterLines
}

func getParameterContent(lines []string) []string {
	r := regexp.MustCompile("(var|const) \\(")
	startFlag := false
	parameterLines := []string{}
	for index, line := range lines {
		if r.MatchString(line) {
			startFlag = true
			continue
		}
		if startFlag {
			if checkEnd(line) {
				parameterLines = append(parameterLines, getParameterContent(lines[index:])...)
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

func getBehindAnnotation(str string) string {
	r := regexp.MustCompile("(.*)//(.*)")
	if r.MatchString(str) {
		return r.FindStringSubmatch(str)[1]
	}
	return ""
}

func checkEnd(str string) bool {
	for _, s := range str {
		if s == ')' {
			return true
		}
	}
	return false
}

func getPackageName(lines []string) string {
	r := regexp.MustCompile("^package (.*)")
	for _, line := range lines {
		if r.MatchString(line) {
			return r.FindStringSubmatch(line)[1]
		}
	}
	return ""
}
