package usercommon

import (
	"context"
	"encoding/hex"
	"math/rand"
	"regexp"
	"strings"

	"github.com/Chain-Zhang/pinyin"
	"github.com/answerdev/answer/internal/base/reason"
	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/schema"
	"github.com/answerdev/answer/pkg/checker"
	"github.com/segmentfault/pacman/errors"
)

type UserRepo interface {
	AddUser(ctx context.Context, user *entity.User) (err error)
	IncreaseAnswerCount(ctx context.Context, userID string, amount int) (err error)
	IncreaseQuestionCount(ctx context.Context, userID string, amount int) (err error)
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
	GetByEmail(ctx context.Context, email string) (userInfo *entity.User, exist bool, err error)
	GetUserCount(ctx context.Context) (count int64, err error)
}

// UserCommon user service
type UserCommon struct {
	userRepo UserRepo
}

func NewUserCommon(userRepo UserRepo) *UserCommon {
	return &UserCommon{
		userRepo: userRepo,
	}
}

func (us *UserCommon) GetUserBasicInfoByID(ctx context.Context, ID string) (
	userBasicInfo *schema.UserBasicInfo, exist bool, err error) {
	userInfo, exist, err := us.userRepo.GetByUserID(ctx, ID)
	if err != nil {
		return nil, exist, err
	}
	info := us.FormatUserBasicInfo(ctx, userInfo)
	return info, exist, nil
}

func (us *UserCommon) GetUserBasicInfoByUserName(ctx context.Context, username string) (*schema.UserBasicInfo, bool, error) {
	userInfo, exist, err := us.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, exist, err
	}
	info := us.FormatUserBasicInfo(ctx, userInfo)
	return info, exist, nil
}

func (us *UserCommon) UpdateAnswerCount(ctx context.Context, userID string, num int) error {
	return us.userRepo.IncreaseAnswerCount(ctx, userID, num)
}

func (us *UserCommon) UpdateQuestionCount(ctx context.Context, userID string, num int) error {
	return us.userRepo.IncreaseQuestionCount(ctx, userID, num)
}

func (us *UserCommon) BatchUserBasicInfoByID(ctx context.Context, IDs []string) (map[string]*schema.UserBasicInfo, error) {
	userMap := make(map[string]*schema.UserBasicInfo)
	dbInfo, err := us.userRepo.BatchGetByID(ctx, IDs)
	if err != nil {
		return userMap, err
	}
	for _, item := range dbInfo {
		info := us.FormatUserBasicInfo(ctx, item)
		userMap[item.ID] = info
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
	userBasicInfo.Avatar = schema.FormatAvatarInfo(userInfo.Avatar)
	userBasicInfo.Website = userInfo.Website
	userBasicInfo.Location = userInfo.Location
	userBasicInfo.IPInfo = userInfo.IPInfo
	userBasicInfo.Status = schema.UserStatusShow[userInfo.Status]
	if userBasicInfo.Status == schema.UserDeleted {
		userBasicInfo.Avatar = ""
		userBasicInfo.DisplayName = "Anonymous"
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

	username = strings.ReplaceAll(displayName, " ", "_")
	username = strings.ToLower(username)
	suffix := ""

	re := regexp.MustCompile(`^[a-z0-9._-]{4,30}$`)
	match := re.MatchString(username)
	if !match {
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
		bytes := make([]byte, 2)
		_, _ = rand.Read(bytes)
		suffix = hex.EncodeToString(bytes)
	}
	return username + suffix, nil
}
