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
	"github.com/gfes980615/check_tool/model"
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
		Short:   "check file have Chinese word",
		Long:    `check .go and .sql file have Chinese word, and export its`,
		RunE:    runCheckChinese,
		Example: "  fops check_chinese --folder [folder] --ignore_folder [ignore_folder (use ',' split)]",
	}
	folder       string
	ignoreFolder string
)

func runCheckChinese(cmd *cobra.Command, args []string) error {
	goFileMap := make(map[string][]model.ChineseRow)
	sqlFileMap := make(map[string][]model.ChineseRow)
	specialWordMap := make(map[string][]model.ChineseRow)

	start := time.Now()

	filePaths := utils.GetAllFileInFolder(folder)
	ignoreFolders := strings.Split(ignoreFolder, ",")

	// 取得檔案中有的全域參數
	parameters := getChineseGlobParam(filePaths, ignoreFolders)

	for _, fileName := range filePaths {
		// 忽略不需要檢查的folder
		if checkIgnoreFolder(ignoreFolders, fileName) {
			continue
		}
		// 檢查是否式go檔或sql檔 (需要另外檢查其他種檔案可自行增加case或用指令取代也行)
		if extension, check := checkFileExtension(fileName); check {
			rows := getChineseRows(fileName)
			if len(rows) == 0 {
				continue
			}
			// 將有使用到中文的檔案紀錄下來
			switch extension {
			case "go":
				goFileMap[fileName] = rows
			case "sql":
				sqlFileMap[fileName] = rows
			}
		}
		// 取得有使用到參數是中文的檔案
		result := globParamInFile(fileName, parameters)
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
	err = createFileByParam(specialWordMap, "glob_parameter.txt")
	if err != nil {
		return err
	}
	fmt.Printf("success\n%v", time.Since(start))
	return nil
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

func getChineseRows(fileName string) []model.ChineseRow {
	lines := utils.TransferFileContentToSlice(fileName)
	chineseRows := []model.ChineseRow{}
	for index, line := range lines {
		// 中文是註解就忽略
		subLine := strings.Split(line, "//")
		if len(subLine) == 2 {
			tmp := getChinese(subLine[0], index+1)
			if len(tmp.Chinese) > 0 {
				chineseRows = append(chineseRows, tmp)
			}
			continue
		}
		tmp := getChinese(line, index+1)
		if len(tmp.Chinese) > 0 {
			chineseRows = append(chineseRows, tmp)
		}
	}
	return chineseRows
}

// 取得中文內容，若有多個字段用逗號隔開紀錄
func getChinese(str string, row int) model.ChineseRow {
	chinese := model.ChineseRow{
		Row:     row,
		Chinese: "",
	}

	if utils.CheckChineseExist(str) {
		chinese.Chinese = ignoreNotChineseCharacter(str)
	}
	return chinese
}

// 忽略不是中文的字元
func ignoreNotChineseCharacter(str string) string {
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
	return utils.ReplaceSpaceToComma(str)
}

func createFileByParam(files map[string][]model.ChineseRow, writeToFileName string) error {
	resetContent := []string{}
	index := 1
	for file, rows := range files {
		str := fmt.Sprintf("file %d: %s\n", index, file)
		index++
		for _, row := range rows {
			for key := range row.Parameter {
				str += fmt.Sprintf("%d\t%s\n", row.Row, key)
			}
		}
		resetContent = append(resetContent, str)
	}
	return utils.WriteToFile(strings.Join(resetContent, "\n--------------------------------------------------\n"), writeToFileName)

}

func globParamInFile(fileName string, parameters []string) []model.ChineseRow {
	content, _ := utils.ReadFileToString(fileName)
	exist, words := checkFileHasWhichParam(content, parameters)
	if !exist {
		return []model.ChineseRow{}
	}

	lines := utils.TransferFileContentToSlice(fileName)
	rows := []model.ChineseRow{}
	for index, line := range lines {
		paramMap := getUsedParamFromStr(line, words)
		wordInfo := setParamInfo(paramMap, index+1)
		rows = append(rows, wordInfo)
	}
	return rows
}

func setParamInfo(paramMap map[string]bool, row int) model.ChineseRow {
	return model.ChineseRow{
		Row:       row,
		Parameter: paramMap,
	}
}

// 從字串中取得有使用到的參數
func getUsedParamFromStr(line string, params []string) map[string]bool {
	content := strings.Split(line, " ")
	paramMap := make(map[string]bool)
	for _, param := range params {
		r := regexp.MustCompile(fmt.Sprintf("(.*)%s(.*)", param))
		for _, str := range content {
			if r.MatchString(str) {
				paramMap[param] = true
			}
		}
	}
	return paramMap
}

// 檢查檔案有多少參數
func checkFileHasWhichParam(content string, parameters []string) (bool, []string) {
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

// 檢查是否是需要忽略的folder
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

func createFileByExtension(files map[string][]model.ChineseRow, writeToFileName string) error {
	resetContent := []string{}
	index := 1
	for file, rows := range files {
		str := fmt.Sprintf("file %d: %s\n", index, file)
		index++
		for _, row := range rows {
			str += fmt.Sprintf("%d\t%s\n", row.Row, row.Chinese)
		}
		resetContent = append(resetContent, str)
	}
	return utils.WriteToFile(strings.Join(resetContent, "\n--------------------------------------------------\n"), writeToFileName)
}
