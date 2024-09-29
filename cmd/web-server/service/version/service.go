/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

// Package version ...
package version

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"hcm/cmd/web-server/service/capability"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/version"
)

// InitVersionService ...
func InitVersionService(c *capability.Capability) {
	svr := &service{
		client: c.ApiClient,
	}

	h := rest.NewHandler()
	h.Add("GetVersionList", "GET", "/changelogs", svr.GetVersionList)
	h.Add("GetVersion", "GET", "/changelog/{version}", svr.GetVersion)

	h.Load(c.WebService)
}

type service struct {
	client *client.ClientSet
}

// GetVersionList get user info
func (u *service) GetVersionList(cts *rest.Contexts) (interface{}, error) {

	language := rest.GetLanguageByHTTPRequest(cts.Request)
	changelogFilepath := getVersionDirPath(language)
	files, err := os.ReadDir(changelogFilepath)
	if err != nil {
		logs.Errorf("failed to read directory: %s, err: %v, rid: %s", changelogFilepath, err, cts.Kit.Rid)
		return nil, err
	}

	if len(files) == 0 {
		return nil, nil
	}

	versionInfoList := getVersionInfoList(files)

	return versionInfoList, nil
}

// GetVersion ...
func (u *service) GetVersion(cts *rest.Contexts) (interface{}, error) {
	curVersion := cts.PathParameter("version").String()

	language := rest.GetLanguageByHTTPRequest(cts.Request)
	versionFilePath, err := getVersionFilePath(getVersionDirPath(language), curVersion)
	if err != nil {
		// 找不到指定的版本数据则返回对应错误
		logs.Errorf("the changelog file for %s could not be found, err: %v, rid: %s",
			curVersion, err, cts.Kit.Rid)
		return nil, err
	}

	versionData, err := os.ReadFile(versionFilePath)
	if err != nil {
		logs.Errorf("failed to open file, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return string(versionData), nil
}

func getVersionDirPath(lang constant.Language) string {
	switch lang {
	case constant.English:
		return cc.WebServer().ChangeLogPath.English
	case constant.Chinese:
		return cc.WebServer().ChangeLogPath.Chinese
	default:
		return cc.WebServer().ChangeLogPath.Chinese
	}
}

// getVersionInfoList get all cmdb version info from changelog files
func getVersionInfoList(files []os.DirEntry) []Item {
	versionInfoList := make([]Item, 0)
	curVersion := getCurrentVersion()
	for _, file := range files {
		fileVersion, updateTime := getFileVersion(file.Name())
		if fileVersion == "" {
			continue
		}
		// 找出当前版本并把IsCurrent设为true
		versionInfoList = append(versionInfoList, Item{
			Version:    fileVersion,
			UpdateTime: updateTime,
			IsCurrent:  fileVersion == curVersion,
		})
	}
	return versionInfoList
}

// getFileVersion get version and updateTime in filename
// eg: test.md, _test.md, test_test.txt; return "", ""
//
//	vaa.bb.cc_2006-01-02.md, v3.10.22_2022-02-29.md; return "", ""
//	v3.10.23-rc_2006-01-02.md, v3.10.22-alpha_2006-01-02.md; return "", ""
//	v3.10.22_2022-03-18.md; return v3.10.22, 2022-03-18
func getFileVersion(filename string) (string, string) {
	matched, err := regexp.MatchString("^.+_.+\\.md$", filename)
	if err != nil {
		logs.Errorf("match the changelog file name failed, err: %v", err)
		return "", ""
	}
	if !matched {
		return "", ""
	}

	filename = filename[:len(filename)-3]
	// 判断文件名中版本号的格式是否符合要求
	// 版本日志文件名不会出现版本号带后缀的情况
	versionRegex := "^v\\d+\\.\\d+\\.\\d+$"
	matched, err = regexp.MatchString(versionRegex, strings.Split(filename, "_")[0])
	if err != nil {
		logs.Errorf("matches version in the changelog file name failed, err: %v", err)
		return "", ""
	}
	if !matched {
		return "", ""
	}

	// 判断文件名中的发布时间的格式是否符合要求
	local, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logs.Errorf("matches updateTime in the changelog file name failed, err: %v", err)
		return "", ""
	}
	_, err = time.ParseInLocation("2006-01-02", strings.Split(filename, "_")[1], local)
	if err != nil {
		logs.Errorf("the updateTime in %s.md does not conform to date format, err: %v", filename, err)
		return "", ""
	}

	fileVersion := strings.Split(filename, "_")[0]
	updateTime := strings.Split(filename, "_")[1]
	return fileVersion, updateTime
}

// getCurrentVersion Get the current version
// eg: returns the version according to the existing tag
//
//	release-v3.10.22-alpha1, return "v3.10.22"
//	release-v3.10.18_alpha1, return "v3.10.18"
//	release-v3.10.16, return "v3.10.16"
//	release-v3.10.x_feature-agent-id_alpha, return ""
func getCurrentVersion() string {
	currentVersion := version.Version().Version
	currentVersionRegex := "v\\d+\\.\\d+\\.\\d+(-|_|$)"

	reg := regexp.MustCompile(currentVersionRegex)
	currentVersion = reg.FindString(currentVersion)
	if currentVersion == "" {
		return ""
	}
	//版本日志文件由产品于验收通过版本（rc版本）时一起出
	// 去后缀操作：
	// 用于在产品进行功能验证时保证当前版本号（带后缀）与不带后缀的版本号的版本日志能匹配得上。
	// 例如：当前版本号为v3.10.23-rc，与之对应的版本日志的版本号为v3.10.23
	if strings.Index(currentVersion, "-") != -1 || strings.Index(currentVersion, "_") != -1 {
		return currentVersion[:len(currentVersion)-1]
	}
	return currentVersion
}

// getVersionFilePath gets the version log path specified in the request body
// eg: if the version in the request body is v3.10.aa, v3.10.23-rc, v3.10.22-alpha, return "" and error
func getVersionFilePath(changelogPath string, version string) (string, error) {
	versionRegex := "^v\\d+\\.\\d+\\.\\d+$"

	matched, err := regexp.MatchString(versionRegex, version)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", fmt.Errorf("version: " + version + " does not conform to the version number format")
	}

	files, err := os.ReadDir(changelogPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: " + changelogPath)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files in " + changelogPath)
	}

	for _, file := range files {
		fileVersion, _ := getFileVersion(file.Name())
		if fileVersion == "" {
			continue
		}
		if fileVersion == version {
			return filepath.Join(changelogPath, file.Name()), nil
		}
	}
	return "", fmt.Errorf("no changelog file for " + version + " in " + changelogPath)
}
