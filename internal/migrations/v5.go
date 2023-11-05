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

package migrations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/incubator-answer/internal/entity"
	"xorm.io/xorm"
)

func addThemeAndPrivateMode(ctx context.Context, x *xorm.Engine) error {
	loginConfig := map[string]bool{
		"allow_new_registrations": true,
		"login_required":          false,
	}
	loginConfigDataBytes, _ := json.Marshal(loginConfig)
	siteInfo := &entity.SiteInfo{
		Type:    "login",
		Content: string(loginConfigDataBytes),
		Status:  1,
	}
	exist, err := x.Context(ctx).Get(&entity.SiteInfo{Type: siteInfo.Type})
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if !exist {
		_, err = x.Context(ctx).Insert(siteInfo)
		if err != nil {
			return fmt.Errorf("insert site info failed: %w", err)
		}
	}

	themeConfig := `{"theme":"default","theme_config":{"default":{"navbar_style":"colored","primary_color":"#0033ff"}}}`
	themeSiteInfo := &entity.SiteInfo{
		Type:    "theme",
		Content: themeConfig,
		Status:  1,
	}
	exist, err = x.Context(ctx).Get(&entity.SiteInfo{Type: themeSiteInfo.Type})
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if !exist {
		_, err = x.Context(ctx).Insert(themeSiteInfo)
	}
	return err
}
