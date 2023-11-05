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

package usercommon

import (
	"context"
	"github.com/apache/incubator-answer/internal/base/constant"
	"github.com/apache/incubator-answer/pkg/converter"
	"strings"

	"github.com/Chain-Zhang/pinyin"
	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/auth"
	"github.com/apache/incubator-answer/internal/service/role"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	"github.com/apache/incubator-answer/pkg/checker"
	"github.com/apache/incubator-answer/pkg/random"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type UserRepo interface {
	AddUser(ctx context.Context, user *entity.User) (err error)
	IncreaseAnswerCount(ctx context.Context, userID string, amount int) (err error)
	IncreaseQuestionCount(ctx context.Context, userID string, amount int) (err error)
	UpdateQuestionCount(ctx context.Context, userID string, count int64) (err error)
	UpdateAnswerCount(ctx context.Context, userID string, count int) (err error)
	UpdateLastLoginDate(ctx context.Context, userID string) (err error)
	UpdateEmailStatus(ctx context.Context, userID string, emailStatus int) error
	UpdateNoticeStatus(ctx context.Context, userID string, noticeStatus int) error
	UpdateEmail(ctx context.Context, userID, email string) error
	UpdateLanguage(ctx context.Context, userID, language string) error
	UpdatePass(ctx context.Context, userID, pass string) error
	UpdateInfo(ctx context.Context, userInfo *entity.User) (err error)
	GetByUserID(ctx context.Context, userID string) (userInfo *entity.User, exist bool, err error)
	BatchGetByID(ctx context.Context, ids []string) ([]*entity.User, error)
	GetByUsername(ctx context.Context, username string) (userInfo *entity.User, exist bool, err error)
	GetByUsernames(ctx context.Context, usernames []string) ([]*entity.User, error)
	GetByEmail(ctx context.Context, email string) (userInfo *entity.User, exist bool, err error)
	GetUserCount(ctx context.Context) (count int64, err error)
	SearchUserListByName(ctx context.Context, name string, limit int) (userList []*entity.User, err error)
}

// UserCommon user service
type UserCommon struct {
	userRepo              UserRepo
	userRoleService       *role.UserRoleRelService
	authService           *auth.AuthService
	siteInfoCommonService siteinfo_common.SiteInfoCommonService
}

func NewUserCommon(
	userRepo UserRepo,
	userRoleService *role.UserRoleRelService,
	authService *auth.AuthService,
	siteInfoCommonService siteinfo_common.SiteInfoCommonService,
) *UserCommon {
	return &UserCommon{
		userRepo:              userRepo,
		userRoleService:       userRoleService,
		authService:           authService,
		siteInfoCommonService: siteInfoCommonService,
	}
}

func (us *UserCommon) GetUserBasicInfoByID(ctx context.Context, ID string) (
	userBasicInfo *schema.UserBasicInfo, exist bool, err error) {
	userInfo, exist, err := us.userRepo.GetByUserID(ctx, ID)
	if err != nil {
		return nil, exist, err
	}
	info := us.FormatUserBasicInfo(ctx, userInfo)
	info.Avatar = us.siteInfoCommonService.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	return info, exist, nil
}

func (us *UserCommon) GetUserBasicInfoByUserName(ctx context.Context, username string) (*schema.UserBasicInfo, bool, error) {
	userInfo, exist, err := us.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, exist, err
	}
	info := us.FormatUserBasicInfo(ctx, userInfo)
	info.Avatar = us.siteInfoCommonService.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	return info, exist, nil
}

func (us *UserCommon) BatchGetUserBasicInfoByUserNames(ctx context.Context, usernames []string) (map[string]*schema.UserBasicInfo, error) {
	infomap := make(map[string]*schema.UserBasicInfo)
	list, err := us.userRepo.GetByUsernames(ctx, usernames)
	if err != nil {
		return infomap, err
	}
	avatarMapping := us.siteInfoCommonService.FormatListAvatar(ctx, list)
	for _, user := range list {
		info := us.FormatUserBasicInfo(ctx, user)
		info.Avatar = avatarMapping[user.ID].GetURL()
		infomap[user.Username] = info
	}
	return infomap, nil
}

func (us *UserCommon) UpdateAnswerCount(ctx context.Context, userID string, num int) error {
	return us.userRepo.UpdateAnswerCount(ctx, userID, num)
}

func (us *UserCommon) UpdateQuestionCount(ctx context.Context, userID string, num int64) error {
	return us.userRepo.UpdateQuestionCount(ctx, userID, num)
}

func (us *UserCommon) BatchUserBasicInfoByID(ctx context.Context, userIDs []string) (map[string]*schema.UserBasicInfo, error) {
	userMap := make(map[string]*schema.UserBasicInfo)
	if len(userIDs) == 0 {
		return userMap, nil
	}
	userList, err := us.userRepo.BatchGetByID(ctx, userIDs)
	if err != nil {
		return userMap, err
	}
	avatarMapping := us.siteInfoCommonService.FormatListAvatar(ctx, userList)
	for _, user := range userList {
		info := us.FormatUserBasicInfo(ctx, user)
		info.Avatar = avatarMapping[user.ID].GetURL()
		userMap[user.ID] = info
	}
	return userMap, nil
}

// FormatUserBasicInfo format user basic info
func (us *UserCommon) FormatUserBasicInfo(ctx context.Context, userInfo *entity.User) *schema.UserBasicInfo {
	userBasicInfo := &schema.UserBasicInfo{}
	userBasicInfo.ID = userInfo.ID
	userBasicInfo.Username = userInfo.Username
	userBasicInfo.Rank = userInfo.Rank
	userBasicInfo.DisplayName = userInfo.DisplayName
	userBasicInfo.Website = userInfo.Website
	userBasicInfo.Location = userInfo.Location
	userBasicInfo.Status = constant.ConvertUserStatus(userInfo.Status, userInfo.MailStatus)
	if userBasicInfo.Status == constant.UserDeleted {
		userBasicInfo.Avatar = ""
		userBasicInfo.DisplayName = "user" + converter.DeleteUserDisplay(userInfo.ID)
	}
	return userBasicInfo
}

// MakeUsername
// Generate a unique Username based on the displayName
func (us *UserCommon) MakeUsername(ctx context.Context, displayName string) (username string, err error) {
	// Chinese processing
	if has := checker.IsChinese(displayName); has {
		str, err := pinyin.New(displayName).Split("").Mode(pinyin.WithoutTone).Convert()
		if err != nil {
			return "", errors.BadRequest(reason.UsernameInvalid)
		} else {
			displayName = str
		}
	}

	username = strings.ReplaceAll(displayName, " ", "-")
	username = strings.ToLower(username)
	suffix := ""

	if checker.IsInvalidUsername(username) {
		return "", errors.BadRequest(reason.UsernameInvalid)
	}

	if checker.IsReservedUsername(username) {
		return "", errors.BadRequest(reason.UsernameInvalid)
	}

	for {
		_, has, err := us.userRepo.GetByUsername(ctx, username+suffix)
		if err != nil {
			return "", err
		}
		if !has {
			break
		}
		suffix = random.UsernameSuffix()
	}
	return username + suffix, nil
}

func (us *UserCommon) CacheLoginUserInfo(ctx context.Context, userID string, userStatus, emailStatus int, externalID string) (
	accessToken string, userCacheInfo *entity.UserCacheInfo, err error) {
	roleID, err := us.userRoleService.GetUserRole(ctx, userID)
	if err != nil {
		log.Error(err)
	}

	userCacheInfo = &entity.UserCacheInfo{
		UserID:      userID,
		EmailStatus: emailStatus,
		UserStatus:  userStatus,
		RoleID:      roleID,
		ExternalID:  externalID,
	}

	accessToken, _, err = us.authService.SetUserCacheInfo(ctx, userCacheInfo)
	if err != nil {
		return "", nil, err
	}
	if userCacheInfo.RoleID == role.RoleAdminID {
		if err = us.authService.SetAdminUserCacheInfo(ctx, accessToken, &entity.UserCacheInfo{UserID: userID}); err != nil {
			return "", nil, err
		}
	}
	return accessToken, userCacheInfo, nil
}
