package migrations

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/answerdev/answer/internal/base/constant"
	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/schema"
	"github.com/answerdev/answer/internal/service/permission"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

func updateCount(x *xorm.Engine) error {
	fns := []func(*xorm.Engine) error{
		inviteAnswer,
		addPrivilegeForInviteSomeoneToAnswer,
		addGravatarBaseURL,
		updateQuestionCount,
		updateTagCount,
		updateUserQuestionCount,
		updateUserAnswerCount,
		inBoxData,
	}
	for _, fn := range fns {
		if err := fn(x); err != nil {
			return err
		}
	}
	return nil
}

func addGravatarBaseURL(x *xorm.Engine) error {
	usersSiteInfo := &entity.SiteInfo{
		Type: constant.SiteTypeUsers,
	}
	exist, err := x.Get(usersSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if exist {
		content := &schema.SiteUsersReq{}
		_ = json.Unmarshal([]byte(usersSiteInfo.Content), content)
		content.GravatarBaseURL = "https://www.gravatar.com/avatar/"
		data, _ := json.Marshal(content)
		usersSiteInfo.Content = string(data)

		_, err = x.ID(usersSiteInfo.ID).Cols("content").Update(usersSiteInfo)
		if err != nil {
			return fmt.Errorf("update site info failed: %w", err)
		}
	}
	return nil
}

func addPrivilegeForInviteSomeoneToAnswer(x *xorm.Engine) error {
	// add rank for invite to answer
	powers := []*entity.Power{
		{ID: 38, Name: "invite someone to answer", PowerType: permission.AnswerInviteSomeoneToAnswer, Description: "invite someone to answer"},
	}
	for _, power := range powers {
		exist, err := x.Get(&entity.Power{PowerType: power.PowerType})
		if err != nil {
			return err
		}
		if exist {
			_, err = x.ID(power.ID).Update(power)
		} else {
			_, err = x.Insert(power)
		}
		if err != nil {
			return err
		}
	}
	rolePowerRels := []*entity.RolePowerRel{
		{RoleID: 2, PowerType: permission.AnswerInviteSomeoneToAnswer},
		{RoleID: 3, PowerType: permission.AnswerInviteSomeoneToAnswer},
	}
	for _, rel := range rolePowerRels {
		exist, err := x.Get(&entity.RolePowerRel{RoleID: rel.RoleID, PowerType: rel.PowerType})
		if err != nil {
			return err
		}
		if exist {
			continue
		}
		_, err = x.Insert(rel)
		if err != nil {
			return err
		}
	}

	defaultConfigTable := []*entity.Config{
		{ID: 127, Key: "rank.answer.invite_someone_to_answer", Value: `1000`},
	}
	for _, c := range defaultConfigTable {
		exist, err := x.Get(&entity.Config{ID: c.ID})
		if err != nil {
			return fmt.Errorf("get config failed: %w", err)
		}
		if exist {
			if _, err = x.Update(c, &entity.Config{ID: c.ID}); err != nil {
				return fmt.Errorf("update config failed: %w", err)
			}
			continue
		}
		if _, err = x.Insert(&entity.Config{ID: c.ID, Key: c.Key, Value: c.Value}); err != nil {
			return fmt.Errorf("add config failed: %w", err)
		}
	}
	return nil
}

func updateQuestionCount(x *xorm.Engine) error {
	//question answer count
	answers := make([]AnswerV13, 0)
	err := x.Find(&answers, &AnswerV13{Status: entity.AnswerStatusAvailable})
	if err != nil {
		return fmt.Errorf("get answers failed: %w", err)
	}
	questionAnswerCount := make(map[string]int)
	for _, answer := range answers {
		_, ok := questionAnswerCount[answer.QuestionID]
		if !ok {
			questionAnswerCount[answer.QuestionID] = 1
		} else {
			questionAnswerCount[answer.QuestionID]++
		}
	}
	questionList := make([]QuestionV13, 0)
	err = x.Find(&questionList, &QuestionV13{})
	if err != nil {
		return fmt.Errorf("get questions failed: %w", err)
	}
	for _, item := range questionList {
		_, ok := questionAnswerCount[item.ID]
		if ok {
			item.AnswerCount = questionAnswerCount[item.ID]
			if _, err = x.Cols("answer_count").Update(item, &QuestionV13{ID: item.ID}); err != nil {
				log.Errorf("update %+v config failed: %s", item, err)
				return fmt.Errorf("update question failed: %w", err)
			}
		}
	}

	return nil
}

// updateTagCount update tag count
func updateTagCount(x *xorm.Engine) error {
	tagRelList := make([]entity.TagRel, 0)
	err := x.Find(&tagRelList, &entity.TagRel{})
	if err != nil {
		return fmt.Errorf("get tag rel failed: %w", err)
	}
	questionIDs := make([]string, 0)
	questionsAvailableMap := make(map[string]bool)
	questionsHideMap := make(map[string]bool)
	for _, item := range tagRelList {
		questionIDs = append(questionIDs, item.ObjectID)
		questionsAvailableMap[item.ObjectID] = false
		questionsHideMap[item.ObjectID] = false
	}
	questionList := make([]QuestionV13, 0)
	err = x.In("id", questionIDs).In("question.status", []int{entity.QuestionStatusAvailable, entity.QuestionStatusClosed}).Find(&questionList, &QuestionV13{})
	if err != nil {
		return fmt.Errorf("get questions failed: %w", err)
	}
	for _, question := range questionList {
		_, ok := questionsAvailableMap[question.ID]
		if ok {
			questionsAvailableMap[question.ID] = true
			if question.Show == entity.QuestionHide {
				questionsHideMap[question.ID] = true
			}
		}
	}

	for id, ok := range questionsHideMap {
		if ok {
			if _, err = x.Cols("status").Update(&entity.TagRel{Status: entity.TagRelStatusHide}, &entity.TagRel{ObjectID: id}); err != nil {
				log.Errorf("update %+v config failed: %s", id, err)
			}
		}
	}

	for id, ok := range questionsAvailableMap {
		if !ok {
			if _, err = x.Cols("status").Update(&entity.TagRel{Status: entity.TagRelStatusDeleted}, &entity.TagRel{ObjectID: id}); err != nil {
				log.Errorf("update %+v config failed: %s", id, err)
			}
		}
	}

	//select tag count
	newTagRelList := make([]entity.TagRel, 0)
	err = x.Find(&newTagRelList, &entity.TagRel{Status: entity.TagRelStatusAvailable})
	if err != nil {
		return fmt.Errorf("get tag rel failed: %w", err)
	}
	tagCountMap := make(map[string]int)
	for _, v := range newTagRelList {
		_, ok := tagCountMap[v.TagID]
		if !ok {
			tagCountMap[v.TagID] = 1
		} else {
			tagCountMap[v.TagID]++
		}
	}
	TagList := make([]entity.Tag, 0)
	err = x.Find(&TagList, &entity.Tag{})
	if err != nil {
		return fmt.Errorf("get tag  failed: %w", err)
	}
	for _, tag := range TagList {
		_, ok := tagCountMap[tag.ID]
		if ok {
			tag.QuestionCount = tagCountMap[tag.ID]
			if _, err = x.Update(tag, &entity.Tag{ID: tag.ID}); err != nil {
				log.Errorf("update %+v tag failed: %s", tag.ID, err)
				return fmt.Errorf("update tag failed: %w", err)
			}
		} else {
			tag.QuestionCount = 0
			if _, err = x.Cols("question_count").Update(tag, &entity.Tag{ID: tag.ID}); err != nil {
				log.Errorf("update %+v tag failed: %s", tag.ID, err)
				return fmt.Errorf("update tag failed: %w", err)
			}
		}
	}
	return nil
}

// updateUserQuestionCount update user question count
func updateUserQuestionCount(x *xorm.Engine) error {
	questionList := make([]QuestionV13, 0)
	err := x.In("status", []int{entity.QuestionStatusAvailable, entity.QuestionStatusClosed}).Find(&questionList, &QuestionV13{})
	if err != nil {
		return fmt.Errorf("get question  failed: %w", err)
	}
	userQuestionCountMap := make(map[string]int)
	for _, question := range questionList {
		_, ok := userQuestionCountMap[question.UserID]
		if !ok {
			userQuestionCountMap[question.UserID] = 1
		} else {
			userQuestionCountMap[question.UserID]++
		}
	}
	userList := make([]entity.User, 0)
	err = x.Find(&userList, &entity.User{})
	if err != nil {
		return fmt.Errorf("get user  failed: %w", err)
	}
	for _, user := range userList {
		_, ok := userQuestionCountMap[user.ID]
		if ok {
			user.QuestionCount = userQuestionCountMap[user.ID]
			if _, err = x.Cols("question_count").Update(user, &entity.User{ID: user.ID}); err != nil {
				log.Errorf("update %+v user failed: %s", user.ID, err)
				return fmt.Errorf("update user failed: %w", err)
			}
		} else {
			user.QuestionCount = 0
			if _, err = x.Cols("question_count").Update(user, &entity.User{ID: user.ID}); err != nil {
				log.Errorf("update %+v user failed: %s", user.ID, err)
				return fmt.Errorf("update user failed: %w", err)
			}
		}
	}
	return nil
}

type AnswerV13 struct {
	ID         string `xorm:"not null pk autoincr BIGINT(20) id"`
	QuestionID string `xorm:"not null default 0 BIGINT(20) question_id"`
	UserID     string `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	Status     int    `xorm:"not null default 1 INT(11) status"`
	Accepted   int    `xorm:"not null default 1 INT(11) adopted"`
}

func (AnswerV13) TableName() string {
	return "answer"
}

// updateUserAnswerCount update user answer count
func updateUserAnswerCount(x *xorm.Engine) error {
	answers := make([]AnswerV13, 0)
	err := x.Find(&answers, &AnswerV13{Status: entity.AnswerStatusAvailable})
	if err != nil {
		return fmt.Errorf("get answers failed: %w", err)
	}
	userAnswerCount := make(map[string]int)
	for _, answer := range answers {
		_, ok := userAnswerCount[answer.UserID]
		if !ok {
			userAnswerCount[answer.UserID] = 1
		} else {
			userAnswerCount[answer.UserID]++
		}
	}
	userList := make([]entity.User, 0)
	err = x.Find(&userList, &entity.User{})
	if err != nil {
		return fmt.Errorf("get user failed: %w", err)
	}
	for _, user := range userList {
		_, ok := userAnswerCount[user.ID]
		if ok {
			user.AnswerCount = userAnswerCount[user.ID]
			if _, err = x.Cols("answer_count").Update(user, &entity.User{ID: user.ID}); err != nil {
				log.Errorf("update %+v user failed: %s", user.ID, err)
				return fmt.Errorf("update user failed: %w", err)
			}
		} else {
			user.AnswerCount = 0
			if _, err = x.Cols("answer_count").Update(user, &entity.User{ID: user.ID}); err != nil {
				log.Errorf("update %+v user failed: %s", user.ID, err)
				return fmt.Errorf("update user failed: %w", err)
			}
		}
	}
	return nil
}

type QuestionV13 struct {
	ID               string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt        time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt        time.Time `xorm:"updated_at TIMESTAMP"`
	UserID           string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	InviteUserID     string    `xorm:"TEXT invite_user_id"`
	LastEditUserID   string    `xorm:"not null default 0 BIGINT(20) last_edit_user_id"`
	Title            string    `xorm:"not null default '' VARCHAR(150) title"`
	OriginalText     string    `xorm:"not null MEDIUMTEXT original_text"`
	ParsedText       string    `xorm:"not null MEDIUMTEXT parsed_text"`
	Status           int       `xorm:"not null default 1 INT(11) status"`
	Pin              int       `xorm:"not null default 1 INT(11) pin"`
	Show             int       `xorm:"not null default 1 INT(11) show"`
	ViewCount        int       `xorm:"not null default 0 INT(11) view_count"`
	UniqueViewCount  int       `xorm:"not null default 0 INT(11) unique_view_count"`
	VoteCount        int       `xorm:"not null default 0 INT(11) vote_count"`
	AnswerCount      int       `xorm:"not null default 0 INT(11) answer_count"`
	CollectionCount  int       `xorm:"not null default 0 INT(11) collection_count"`
	FollowCount      int       `xorm:"not null default 0 INT(11) follow_count"`
	AcceptedAnswerID string    `xorm:"not null default 0 BIGINT(20) accepted_answer_id"`
	LastAnswerID     string    `xorm:"not null default 0 BIGINT(20) last_answer_id"`
	PostUpdateTime   time.Time `xorm:"post_update_time TIMESTAMP"`
	RevisionID       string    `xorm:"not null default 0 BIGINT(20) revision_id"`
}

func (QuestionV13) TableName() string {
	return "question"
}

func inviteAnswer(x *xorm.Engine) error {
	err := x.Sync(new(QuestionV13))
	if err != nil {
		return err
	}
	return nil
}

// inBoxData Classify messages
func inBoxData(x *xorm.Engine) error {
	type Notification struct {
		ID        string    `xorm:"not null pk autoincr BIGINT(20) id"`
		CreatedAt time.Time `xorm:"created TIMESTAMP created_at"`
		UpdatedAt time.Time `xorm:"TIMESTAMP updated_at"`
		UserID    string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
		ObjectID  string    `xorm:"not null default 0 INDEX BIGINT(20) object_id"`
		Content   string    `xorm:"not null TEXT content"`
		Type      int       `xorm:"not null default 0 INT(11) type"`
		MsgType   int       `xorm:"not null default 0 INT(11) msg_type"`
		IsRead    int       `xorm:"not null default 1 INT(11) is_read"`
		Status    int       `xorm:"not null default 1 INT(11) status"`
	}
	err := x.Sync(new(Notification))
	if err != nil {
		return err
	}
	msglist := make([]entity.Notification, 0)
	err = x.Find(&msglist, &entity.Notification{})
	if err != nil {
		return fmt.Errorf("get Notification failed: %w", err)
	}
	for _, v := range msglist {
		Content := &schema.NotificationContent{}
		err := json.Unmarshal([]byte(v.Content), Content)
		if err != nil {
			continue
		}
		_, ok := constant.NotificationMsgTypeMapping[Content.NotificationAction]
		if ok {
			v.MsgType = constant.NotificationMsgTypeMapping[Content.NotificationAction]
			if _, err = x.Update(v, &entity.Notification{ID: v.ID}); err != nil {
				log.Errorf("update %+v Notification failed: %s", v.ID, err)
			}
		}
	}

	return nil
}
