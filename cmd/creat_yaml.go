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
)

func init() {
	rootCmd.AddCommand(createYaml)
	createYaml.Flags().StringVar(&folder, "folder", "", "enter folder")
	createYaml.Flags().StringVar(&file, "file", "", "enter file")
	createYaml.Flags().StringVar(&ignoreFolder, "ignore_folder", "", "enter ignore folder use ',' split it")
}

const (
	path     = "  - path: %s\n"
	redirect = "    redirect: %s\n"
	method   = "    method: %s\n"
)

var (
	createYaml = &cobra.Command{
		Use:     "create_yaml",
		Short:   "create yaml",
		Long:    `create yaml`,
		RunE:    runCreateYaml,
		Example: "  fops create_yaml --folder [folder]",
	}
	file string
)

func runCreateYaml(cmd *cobra.Command, args []string) error {
	start := time.Now()
	var err error
	if len(file) != 0 {
		if err = createYamlFile(file); err != nil {
			fmt.Println(err)
		}
	}
	if len(folder) != 0 {
		filePaths := utils.GetAllFileInFolder(folder)
		for _, file := range filePaths {
			if err = createYamlFile(file); err != nil {
				fmt.Println(err)
			}
		}
	}
	fmt.Printf("success\n%v", time.Since(start))
	return err
}

func createYamlFile(file string) error {
	contentLines := utils.TransferFileContentToSlice(file)
	methodContent := getMethodContent(contentLines, "SetupRouter")
	groupMap := getRouterGroupToMap(methodContent)
	routers := setRouter(methodContent, groupMap)

	exportContent := "service:\napi_level:\napis:\n"
	for _, route := range routers {
		exportContent += fmt.Sprintf(path, route.Router) + fmt.Sprintf(redirect, route.Router) + fmt.Sprintf(method, route.Method)
	}
	if err := utils.WriteToFile(exportContent, setExportFileName(file)); err != nil {
		return err
	}
	return nil
}

func setExportFileName(file string) string {
	name := strings.Split(file, "/")
	str := name[len(name)-1]
	r := regexp.MustCompile("(.*).go")
	if !r.MatchString(str) {
		return "unknow.yaml"
	}
	return r.FindStringSubmatch(str)[1] + ".yaml"
}

func setRouter(methodContent []string, groupMap map[string]string) []model.RouterItem {
	re := make(map[string]*regexp.Regexp)
	for group, url := range groupMap {
		re[url] = regexp.MustCompile(fmt.Sprintf(`%s.(GET|PUT|POST|DELETE)\("(.*)",(.*)\)`, group))
	}
	resRouter := []model.RouterItem{}
	for _, content := range methodContent {
		flag := false
		for groupURL, r := range re {
			if r.MatchString(content) {
				matchContent := r.FindStringSubmatch(content)
				tmpRouter := model.RouterItem{
					Router: groupURL + matchContent[2],
					Method: matchContent[1],
				}
				resRouter = append(resRouter, tmpRouter)
				flag = true
			}
		}
		if !flag {
			r := regexp.MustCompile(`(.*).(GET|PUT|POST|DELETE)\("(.*)",(.*)\)`)
			if r.MatchString(content) {
				matchContent := r.FindStringSubmatch(content)
				tmpRouter := model.RouterItem{
					Router: matchContent[3],
					Method: matchContent[2],
				}
				resRouter = append(resRouter, tmpRouter)
			}
		}
	}
	return resRouter
}

func getRouterGroupToMap(methodContent []string) map[string]string {
	r := regexp.MustCompile(`(.*):=(.*).Group\("(.*)"\)`)
	slice := []model.GroupItem{}
	resultMap := make(map[string]string)
	for _, content := range methodContent {
		if r.MatchString(content) {
			subContent := r.FindStringSubmatch(content)
			tmp := model.GroupItem{
				Group:  strings.TrimSpace(subContent[1]),
				Parent: strings.TrimSpace(subContent[2]),
				URL:    strings.TrimSpace(subContent[3]),
			}
			slice = append(slice, tmp)
		}
	}

	for i := 0; i < len(slice); i++ {
		parent := slice[i].Parent
		for j := 0; j < len(slice); j++ {
			if slice[j].Group == parent {
				slice[i].URL = slice[j].URL + slice[i].URL
			}
		}
	}

	for _, s := range slice {
		resultMap[s.Group] = s.URL
	}

	return resultMap
}

func getMethodContent(contentLines []string, method string) []string {
	reStr := fmt.Sprintf("func(.*) %s(.*)", method)
	r := regexp.MustCompile(reStr)
	startFlag := false
	count := 1
	result := []string{}
	for _, line := range contentLines {
		if r.MatchString(line) {
			startFlag = true
			continue
		}
		if startFlag {
			r := regexp.MustCompile("(.*)//(.*)")
			if r.MatchString(line) {
				line = r.FindStringSubmatch(line)[1]
			}
			count := checkCurlyBracketEnd(line, count)
			if count == 0 {
				break
			}
			result = append(result, line)
		}
	}
	return result
}

func checkCurlyBracketEnd(str string, count int) int {
	for _, s := range str {
		if s == '{' {
			count++
		}
		if s == '}' {
			count--
		}
		if count == 0 {
			return count
		}
	}
	return count
}
