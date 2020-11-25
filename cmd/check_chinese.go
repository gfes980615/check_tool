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
	"unicode"
)

func init() {
	rootCmd.AddCommand(checkChinese)
	checkChinese.Flags().StringVar(&folder, "folder", "", "enter folder")
	checkChinese.Flags().StringVar(&ignoreFolder, "ignore_folder", "", "enter ignore folder use ',' split it")
}

var (
	checkChinese = &cobra.Command{
		Use:     "check_chinese",
		Short:   "check file have chinese word",
		Long:    `check .go and .sql file have chinese word, and export its`,
		RunE:    runCheckChinese,
		Example: "  fops check_chinese --folder [folder] --ignore_folder [ignore_folder (use ',' split)]",
	}
	folder       string
	ignoreFolder string
)

func runCheckChinese(cmd *cobra.Command, args []string) error {
	start := time.Now()
	filePaths := utils.GetAllFileInFolder(folder)
	goFileMap := make(map[string][]ChineseRow)
	sqlFileMap := make(map[string][]ChineseRow)
	specialWordMap := make(map[string][]ChineseRow)
	ignoreFolders := strings.Split(ignoreFolder, ",")
	parameters := getHasChineseParameter(filePaths, ignoreFolders)
	for _, fileName := range filePaths {
		if checkIgnoreFolder(ignoreFolders, fileName) {
			continue
		}
		// 檢查是否式go檔或sql檔
		if extension, check := checkFileExtension(fileName); check {
			if file, rows := getChineseFileName(fileName); len(rows) > 0 {
				switch extension {
				case "go":
					goFileMap[file] = rows
				case "sql":
					sqlFileMap[file] = rows
				}
			}
		}
		result := globParameterInFile(fileName, parameters)
		if len(result) > 0 {
			specialWordMap[fileName] = result
		}
	}

	var err error
	err = createFileByExtension(goFileMap, "go.txt")
	if err != nil {
		return err
	}
	err = createFileByExtension(sqlFileMap, "sql.txt")
	if err != nil {
		return err
	}
	err = createFileByParameter(specialWordMap, "glob_parameter.txt")
	if err != nil {
		return err
	}
	fmt.Println(time.Since(start))
	return nil
}

func createFileByParameter(files map[string][]ChineseRow, writeToFileName string) error {
	resetContent := []string{}
	index := 1
	for file, rows := range files {
		str := fmt.Sprintf("file %d: %s\n", index, file)
		index++
		for _, row := range rows {
			for key := range row.special {
				str += fmt.Sprintf("%d\t%s\n", row.row, key)
			}
		}
		resetContent = append(resetContent, str)
	}
	return utils.WriteToFile(strings.Join(resetContent, "\n--------------------------------------------------\n"), writeToFileName)

}

func globParameterInFile(fileName string, parameters []string) []ChineseRow {
	content, _ := utils.ReadFileToString(fileName)
	exist, words := checkFileHasSpecialWord(content, parameters)
	if !exist {
		return []ChineseRow{}
	}

	lines := fileContentToLines(fileName)
	rows := []ChineseRow{}
	for index, line := range lines {
		rows = append(rows, setSpecialWordInfo(hasSpecialWord(line, words), index+1))
	}
	return rows
}

func setSpecialWordInfo(wordMap map[string]bool, row int) ChineseRow {
	return ChineseRow{
		row:     row,
		special: wordMap,
	}
}

func hasSpecialWord(line string, words []string) map[string]bool {
	content := strings.Split(line, " ")
	wordMap := make(map[string]bool)
	for _, word := range words {
		r := regexp.MustCompile(fmt.Sprintf("(.*)%s(.*)", word))
		for _, str := range content {
			if r.MatchString(str) {
				wordMap[word] = true
			}
		}
	}
	return wordMap
}

func checkFileHasSpecialWord(content string, parameters []string) (bool, []string) {
	flag := false
	words := []string{}
	for _, parameter := range parameters {
		if len(strings.Split(content, parameter)) > 1 {
			flag = true
			words = append(words, parameter)
		}
	}
	return flag, words
}

func checkIgnoreFolder(ignoreFolders []string, fileName string) bool {
	for _, ignore := range ignoreFolders {
		if len(ignore) == 0 {
			continue
		}
		if len(strings.Split(fileName, ignore+"/")) > 1 {
			return true
		}
	}
	return false
}

func createFileByExtension(files map[string][]ChineseRow, writeToFileName string) error {
	resetContent := []string{}
	index := 1
	for file, rows := range files {
		str := fmt.Sprintf("file %d: %s\n", index, file)
		index++
		for _, row := range rows {
			str += fmt.Sprintf("%d\t%s\n", row.row, row.chinese)
		}
		resetContent = append(resetContent, str)
	}
	return utils.WriteToFile(strings.Join(resetContent, "\n--------------------------------------------------\n"), writeToFileName)
}

func checkFileExtension(fileName string) (string, bool) {
	r, _ := regexp.MatchString("(.*).(go|sql)", fileName)
	re := regexp.MustCompile("(.*).(go|sql)")
	e := re.FindStringSubmatch(fileName)
	if len(e) != 3 {
		return "", r
	}
	return e[2], r
}

type ChineseRow struct {
	row     int
	chinese string
	special map[string]bool
}

func fileContentToLines(fileName string) []string {
	content, err := utils.ReadFileToString(fileName)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	return strings.Split(content, "\n")
}

func getChineseFileName(fileName string) (string, []ChineseRow) {
	lines := fileContentToLines(fileName)
	chineseRows := []ChineseRow{}
	for index, line := range lines {
		// 排除註解
		subLine := strings.Split(line, "//")
		if len(subLine) == 2 {
			tmp := getChinese(subLine[0], index+1)
			if len(tmp.chinese) > 0 {
				chineseRows = append(chineseRows, tmp)
			}
			continue
		}
		tmp := getChinese(line, index+1)
		if len(tmp.chinese) > 0 {
			chineseRows = append(chineseRows, tmp)
		}
	}
	return fileName, chineseRows
}

func getChinese(str string, row int) ChineseRow {
	chinese := ChineseRow{
		row:     row,
		chinese: "",
	}

	if checkChineseExist(str) {
		conn := false
		lineContent := ""
		for _, r := range str {
			if unicode.Is(unicode.Han, r) {
				conn = true
			} else {
				conn = false
			}
			if conn {
				lineContent += string(r)
			} else {
				lineContent += " "
			}
		}
		chinese.chinese = replaceSpaceToComma(lineContent)
	}
	return chinese
}

func replaceSpaceToComma(str string) string {
	words := []string{}
	for _, word := range strings.Split(str, " ") {
		if len(word) > 0 {
			words = append(words, word)
		}
	}
	return strings.Join(words, ",")
}

func checkChineseExist(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
