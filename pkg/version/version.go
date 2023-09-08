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
)

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

// SysVersion describe a binary version
type SysVersion struct {
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Time    string `json:"time"`
}
