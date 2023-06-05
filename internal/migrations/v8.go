package migrations

import (
	"fmt"
	"time"

	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/service/permission"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

func addRolePinAndHideFeatures(x *xorm.Engine) error {

	powers := []*entity.Power{
		{ID: 34, Name: "question pin", PowerType: permission.QuestionPin, Description: "top the question"},
		{ID: 35, Name: "question hide", PowerType: permission.QuestionHide, Description: "hide  the question"},
		{ID: 36, Name: "question unpin", PowerType: permission.QuestionUnPin, Description: "untop the question"},
		{ID: 37, Name: "question show", PowerType: permission.QuestionShow, Description: "show the question"},
	}
	// insert default powers
	for _, power := range powers {
		exist, err := x.Get(&entity.Power{ID: power.ID})
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

		{RoleID: 2, PowerType: permission.QuestionPin},
		{RoleID: 2, PowerType: permission.QuestionHide},
		{RoleID: 2, PowerType: permission.QuestionUnPin},
		{RoleID: 2, PowerType: permission.QuestionShow},

		{RoleID: 3, PowerType: permission.QuestionPin},
		{RoleID: 3, PowerType: permission.QuestionHide},
		{RoleID: 3, PowerType: permission.QuestionUnPin},
		{RoleID: 3, PowerType: permission.QuestionShow},
	}

	// insert default powers
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
		{ID: 119, Key: "question.pin", Value: `-1`},
		{ID: 120, Key: "question.unpin", Value: `-1`},
		{ID: 121, Key: "question.show", Value: `-1`},
		{ID: 122, Key: "question.hide", Value: `-1`},
	}
	for _, c := range defaultConfigTable {
		exist, err := x.Get(&entity.Config{ID: c.ID, Key: c.Key})
		if err != nil {
			return fmt.Errorf("get config failed: %w", err)
		}
		if exist {
			if _, err = x.Update(c, &entity.Config{ID: c.ID, Key: c.Key}); err != nil {
				log.Errorf("update %+v config failed: %s", c, err)
				return fmt.Errorf("update config failed: %w", err)
			}
			continue
		}
		if _, err = x.Insert(&entity.Config{ID: c.ID, Key: c.Key, Value: c.Value}); err != nil {
			log.Errorf("insert %+v config failed: %s", c, err)
			return fmt.Errorf("add config failed: %w", err)
		}
	}

	type Question struct {
		ID               string    `xorm:"not null pk BIGINT(20) id"`
		CreatedAt        time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
		UpdatedAt        time.Time `xorm:"updated_at TIMESTAMP"`
		UserID           string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
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
	err := x.Sync(new(Question))
	if err != nil {
		return err
	}

	return nil
}
