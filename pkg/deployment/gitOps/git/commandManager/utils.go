/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commandManager

import (
	"github.com/go-git/go-billy/v5/osfs"
	"os"
	"path/filepath"
	"strings"
)

func LocateGitRepo(path string) error {

	if _, err := filepath.Abs(path); err != nil {
		return err
	}
	fst := osfs.New(path)
	_, err := fst.Stat(".git")
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func IsAlreadyUpToDateError(response, errMsg string) bool {
	if strings.Contains(response, "already up-to-date") || strings.Contains(errMsg, "already up-to-date") {
		return true
	}
	return false
}
