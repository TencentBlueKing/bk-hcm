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

package version

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	// validate if the VERSION is valid
	_, err := parseVersion()
	if err != nil {
		msg := fmt.Sprintf("invalid build version, the version(%s) format should be like like v1.0.0 or "+
			"v1.0.0-alpha1, err: %v", VERSION, err)
		fmt.Fprintf(os.Stderr, msg)
		panic(msg)
	}
}

const (
	// LOGO is bk hcm inner logo.
	LOGO = `
===================================================================================
		 ______     ___  ____      ____  ____     ______   ____    ____  
		|_   _ \   |_  ||_  _|    |_   ||   _|  .' ___  | |_   \  /   _| 
		  | |_) |    | |_/ / ______ | |__| |   / .'   \_|   |   \/   |   
		  |  __'.    |  __'.|______||  __  |   | |          | |\  /| |   
		 _| |__) |  _| |  \ \_     _| |  | |_  \ '.___.'\  _| |_\/_| |_  
		|_______/  |____||____|   |____||____|  '.____ .' |_____||_____|
===================================================================================`
)

var (
	// VERSION is version info.
	VERSION = "debug"

	// BUILDTIME  build time.
	BUILDTIME = "unknown"

	// GITHASH git hash for release.
	GITHASH = "unknown"

	// DEBUG if enable debug.
	DEBUG = "false"
)

// Debug show the version if enable debug.
func Debug() bool {
	if DEBUG == "true" {
		return true
	}

	return false
}

// ShowVersion shows the version info.
func ShowVersion() {
	fmt.Println(FormatVersion())
}

// FormatVersion returns service's version.
func FormatVersion() string {
	return fmt.Sprintf("Version: %s\nBuildTime: %s\nGitHash: %s\n", VERSION, BUILDTIME, GITHASH)
}

// GetStartInfo returns start info that includes version and logo.
func GetStartInfo() string {
	startInfo := fmt.Sprintf("%s\n\n%s\n", LOGO, FormatVersion())
	return startInfo
}

// Version ...
func Version() *SysVersion {
	return &SysVersion{
		Version: VERSION,
		Hash:    GITHASH,
		Time:    BUILDTIME,
	}
}

// SemanticVersion return the current process's version with semantic version format.
func SemanticVersion() [3]uint32 {
	ver, _ := parseVersion()
	return ver
}

var versionRegex = regexp.MustCompile(`^v1\.\d?(\.\d?){1,2}(-[a-z]+\d+)?$`)

func parseVersion() ([3]uint32, error) {
	if !versionRegex.MatchString(VERSION) {
		return [3]uint32{}, errors.New("the version should be suffixed with format like v1.0.0")
	}

	ver := strings.Split(VERSION, "-")[0]
	ver = strings.Trim(ver, " ")
	ver = strings.TrimPrefix(ver, "v")
	ele := strings.Split(ver, ".")
	if len(ele) < 3 {
		return [3]uint32{}, errors.New("version should be like v1.0.0")
	}

	major, err := strconv.Atoi(ele[0])
	if err != nil {
		return [3]uint32{}, fmt.Errorf("invalid major version: %s", ele[0])
	}

	minor, err := strconv.Atoi(ele[1])
	if err != nil {
		return [3]uint32{}, fmt.Errorf("invalid minor version: %s", ele[0])
	}

	patch, err := strconv.Atoi(ele[2])
	if err != nil {
		return [3]uint32{}, fmt.Errorf("invalid patch version: %s", ele[0])
	}

	return [3]uint32{uint32(major), uint32(minor), uint32(patch)}, nil
}

// SysVersion describe a binary version
type SysVersion struct {
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Time    string `json:"time"`
}
