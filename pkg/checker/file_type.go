/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package checker

import (
	"fmt"
	"image"
	_ "image/gif" // use init to support decode jpeg,jpg,png,gif
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/segmentfault/pacman/log"
	"golang.org/x/image/webp"
)

const (
	maxImageSize = 8192 * 8192
)

// IsSupportedImageFile currently answers support image type is
// `image/jpeg, image/jpg, image/png, image/gif, image/webp`
func IsSupportedImageFile(localFilePath string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(localFilePath), "."))
	switch ext {
	case "jpg", "jpeg", "png", "gif": // only allow for `image/jpeg,image/jpg,image/png, image/gif`
		if !decodeAndCheckImageFile(localFilePath, standardImageConfigCheck) {
			return false
		}
		if !decodeAndCheckImageFile(localFilePath, standardImageCheck) {
			return false
		}
	case "ico":
		// TODO: There is currently no good Golang library to parse whether the image is in ico format.
		return true
	case "webp":
		if !decodeAndCheckImageFile(localFilePath, webpImageConfigCheck) {
			return false
		}
		if !decodeAndCheckImageFile(localFilePath, webpImageCheck) {
			return false
		}
	default:
		return false
	}
	return true
}

func decodeAndCheckImageFile(localFilePath string, checker func(io.Reader) error) bool {
	file, err := os.Open(localFilePath)
	if err != nil {
		log.Errorf("open file error: %v", err)
		return false
	}
	defer file.Close()

	if err = checker(file); err != nil {
		log.Errorf("check image format error: %v", err)
		return false
	}
	return true
}

func standardImageConfigCheck(file io.Reader) error {
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("decode image config error: %v", err)
	}
	if imageSizeTooLarge(config) {
		return fmt.Errorf("image size too large")
	}
	return nil
}

func standardImageCheck(file io.Reader) error {
	_, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("decode image error: %v", err)
	}
	return nil
}

func webpImageConfigCheck(file io.Reader) error {
	config, err := webp.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("decode webp image config error: %v", err)
	}
	if imageSizeTooLarge(config) {
		return fmt.Errorf("image size too large")
	}
	return nil
}

func webpImageCheck(file io.Reader) error {
	_, err := webp.Decode(file)
	if err != nil {
		return fmt.Errorf("decode webp image error: %v", err)
	}
	return nil
}

func imageSizeTooLarge(config image.Config) bool {
	return config.Width*config.Height > maxImageSize
}
