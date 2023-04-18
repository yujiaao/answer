package migrations

import (
	"encoding/json"
	"fmt"

	"github.com/answerdev/answer/internal/base/data"
	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/service/permission"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/xorm"
)

const (
	defaultSEORobotTxt = `User-agent: *
Disallow: /admin
Disallow: /search
Disallow: /install
Disallow: /review
Disallow: /users/login
Disallow: /users/register
Disallow: /users/account-recovery
Disallow: /users/oauth/*
Disallow: /users/*/*
Disallow: /answer/api
Disallow: /*?code*

Sitemap: `
)

var tables = []interface{}{
	&entity.Activity{},
	&entity.Answer{},
	&entity.Collection{},
	&entity.CollectionGroup{},
	&entity.Comment{},
	&entity.Config{},
	&entity.Meta{},
	&entity.Notification{},
	&entity.Question{},
	&entity.Report{},
	&entity.Revision{},
	&entity.SiteInfo{},
	&entity.Tag{},
	&entity.TagRel{},
	&entity.Uniqid{},
	&entity.User{},
	&entity.Version{},
	&entity.Role{},
	&entity.RolePowerRel{},
	&entity.Power{},
	&entity.UserRoleRel{},
}

// InitDB init db
func InitDB(dataConf *data.Database) (err error) {
	engine, err := data.NewDB(false, dataConf)
	if err != nil {
		fmt.Println("new database failed: ", err.Error())
		return err
	}

	exist, err := engine.IsTableExist(&entity.Version{})
	if err != nil {
		return fmt.Errorf("check table exists failed: %s", err)
	}
	if exist {
		fmt.Println("[database] already exists")
		return nil
	}

	err = engine.Sync(tables...)
	if err != nil {
		return fmt.Errorf("sync table failed: %s", err)
	}
	_, err = engine.InsertOne(&entity.Version{ID: 1, VersionNumber: ExpectedVersion()})
	if err != nil {
		return fmt.Errorf("init version table failed: %s", err)
	}

	err = initAdminUser(engine)
	if err != nil {
		return fmt.Errorf("init admin user failed: %s", err)
	}

	err = initConfigTable(engine)
	if err != nil {
		return fmt.Errorf("init config table: %s", err)
	}

	err = initRolePower(engine)
	if err != nil {
		return fmt.Errorf("init role and power failed: %s", err)
	}
	return nil
}

func initAdminUser(engine *xorm.Engine) error {
	_, err := engine.InsertOne(&entity.User{
		ID:           "1",
		Username:     "admin",
		Pass:         "$2a$10$.gnUnpW.8ssRNaEvx.XwvOR2NuPsGzFLWWX2rqSIVAdIvLNZZYs5y", // admin
		EMail:        "admin@admin.com",
		MailStatus:   1,
		NoticeStatus: 1,
		Status:       1,
		Rank:         1,
		DisplayName:  "admin",
	})
	return err
}

func initSiteInfo(engine *xorm.Engine, language, siteName, siteURL, contactEmail string) error {
	interfaceData := map[string]string{
		"logo":     "",
		"theme":    "black",
		"language": language,
	}
	interfaceDataBytes, _ := json.Marshal(interfaceData)
	_, err := engine.InsertOne(&entity.SiteInfo{
		Type:    "interface",
		Content: string(interfaceDataBytes),
		Status:  1,
	})
	if err != nil {
		return err
	}

	generalData := map[string]string{
		"name":          siteName,
		"site_url":      siteURL,
		"contact_email": contactEmail,
	}
	generalDataBytes, _ := json.Marshal(generalData)
	_, err = engine.InsertOne(&entity.SiteInfo{
		Type:    "general",
		Content: string(generalDataBytes),
		Status:  1,
	})
	if err != nil {
		return err
	}

	loginConfig := map[string]bool{
		"allow_new_registrations": true,
		"login_required":          false,
	}
	loginConfigDataBytes, _ := json.Marshal(loginConfig)
	_, err = engine.InsertOne(&entity.SiteInfo{
		Type:    "login",
		Content: string(loginConfigDataBytes),
		Status:  1,
	})
	if err != nil {
		return err
	}

	themeConfig := `{"theme":"default","theme_config":{"default":{"navbar_style":"colored","primary_color":"#0033ff"}}}`
	_, err = engine.InsertOne(&entity.SiteInfo{
		Type:    "theme",
		Content: themeConfig,
		Status:  1,
	})
	if err != nil {
		return err
	}

	seoData := map[string]string{
		"robots": defaultSEORobotTxt + siteURL + "/sitemap.xml",
	}
	seoDataBytes, _ := json.Marshal(seoData)
	_, err = engine.InsertOne(&entity.SiteInfo{
		Type:    "seo",
		Content: string(seoDataBytes),
		Status:  1,
	})
	if err != nil {
		return err
	}
	return err
}

func updateAdminInfo(engine *xorm.Engine, adminName, adminPassword, adminEmail string) error {
	generateFromPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	adminPassword = string(generateFromPassword)

	// update admin info
	_, err = engine.ID("1").Update(&entity.User{
		Username:    adminName,
		Pass:        adminPassword,
		EMail:       adminEmail,
		DisplayName: adminName,
	})
	if err != nil {
		return fmt.Errorf("update admin user info failed: %s", err)
	}
	return nil
}

// UpdateInstallInfo update some init data about the admin interface and admin password
func UpdateInstallInfo(dataConf *data.Database, language string,
	siteName string,
	siteURL string,
	contactEmail string,
	adminName string,
	adminPassword string,
	adminEmail string) error {

	engine, err := data.NewDB(false, dataConf)
	if err != nil {
		return fmt.Errorf("database connection error: %s", err)
	}

	err = updateAdminInfo(engine, adminName, adminPassword, adminEmail)
	if err != nil {
		return fmt.Errorf("update admin info failed: %s", err)
	}

	err = initSiteInfo(engine, language, siteName, siteURL, contactEmail)
	if err != nil {
		return fmt.Errorf("init site info failed: %s", err)
	}
	return err
}

func initConfigTable(engine *xorm.Engine) error {
	defaultConfigTable := []*entity.Config{
		{ID: 1, Key: "answer.accepted", Value: `15`},
		{ID: 2, Key: "answer.voted_up", Value: `10`},
		{ID: 3, Key: "question.voted_up", Value: `10`},
		{ID: 4, Key: "tag.edit_accepted", Value: `2`},
		{ID: 5, Key: "answer.accept", Value: `2`},
		{ID: 6, Key: "answer.voted_down_cancel", Value: `2`},
		{ID: 7, Key: "question.voted_down_cancel", Value: `2`},
		{ID: 8, Key: "answer.vote_down_cancel", Value: `1`},
		{ID: 9, Key: "question.vote_down_cancel", Value: `1`},
		{ID: 10, Key: "user.activated", Value: `1`},
		{ID: 11, Key: "edit.accepted", Value: `2`},
		{ID: 12, Key: "answer.vote_down", Value: `-1`},
		{ID: 13, Key: "question.voted_down", Value: `-2`},
		{ID: 14, Key: "answer.voted_down", Value: `-2`},
		{ID: 15, Key: "answer.accept_cancel", Value: `-2`},
		{ID: 16, Key: "answer.deleted", Value: `-5`},
		{ID: 17, Key: "question.voted_up_cancel", Value: `-10`},
		{ID: 18, Key: "answer.voted_up_cancel", Value: `-10`},
		{ID: 19, Key: "answer.accepted_cancel", Value: `-15`},
		{ID: 20, Key: "object.reported", Value: `-100`},
		{ID: 21, Key: "edit.rejected", Value: `-2`},
		{ID: 22, Key: "daily_rank_limit", Value: `200`},
		{ID: 23, Key: "daily_rank_limit.exclude", Value: `["answer.accepted"]`},
		{ID: 24, Key: "user.follow", Value: `0`},
		{ID: 25, Key: "comment.vote_up", Value: `0`},
		{ID: 26, Key: "comment.vote_up_cancel", Value: `0`},
		{ID: 27, Key: "question.vote_down", Value: `0`},
		{ID: 28, Key: "question.vote_up", Value: `0`},
		{ID: 29, Key: "question.vote_up_cancel", Value: `0`},
		{ID: 30, Key: "answer.vote_up", Value: `0`},
		{ID: 31, Key: "answer.vote_up_cancel", Value: `0`},
		{ID: 32, Key: "question.follow", Value: `0`},
		{ID: 33, Key: "email.config", Value: `{"from_name":"","from_email":"","smtp_host":"","smtp_port":465,"smtp_password":"","smtp_username":"","smtp_authentication":true,"encryption":"","register_title":"[{{.SiteName}}] Confirm your new account","register_body":"Welcome to {{.SiteName}}<br><br>\n\nClick the following link to confirm and activate your new account:<br>\n<a href='{{.RegisterUrl}}' target='_blank'>{{.RegisterUrl}}</a><br><br>\n\nIf the above link is not clickable, try copying and pasting it into the address bar of your web browser.\n","pass_reset_title":"[{{.SiteName }}] Password reset","pass_reset_body":"Somebody asked to reset your password on [{{.SiteName}}].<br><br>\n\nIf it was not you, you can safely ignore this email.<br><br>\n\nClick the following link to choose a new password:<br>\n<a href='{{.PassResetUrl}}' target='_blank'>{{.PassResetUrl}}</a>\n","change_title":"[{{.SiteName}}] Confirm your new email address","change_body":"Confirm your new email address for {{.SiteName}}  by clicking on the following link:<br><br>\n\n<a href='{{.ChangeEmailUrl}}' target='_blank'>{{.ChangeEmailUrl}}</a><br><br>\n\nIf you did not request this change, please ignore this email.\n","test_title":"[{{.SiteName}}] Test Email","test_body":"This is a test email.","new_answer_title":"[{{.SiteName}}] {{.DisplayName}} answered your question","new_answer_body":"<strong><a href='{{.AnswerUrl}}'>{{.QuestionTitle}}</a></strong><br><br>\n\n<small>{{.DisplayName}}:</small><br>\n<blockquote>{{.AnswerSummary}}</blockquote><br>\n<a href='{{.AnswerUrl}}'>View it on {{.SiteName}}</a><br><br>\n\n<small>You are receiving this because you authored the thread. <a href='{{.UnsubscribeUrl}}'>Unsubscribe</a></small>","new_comment_title":"[{{.SiteName}}] {{.DisplayName}} commented on your post","new_comment_body":"<strong><a href='{{.CommentUrl}}'>{{.QuestionTitle}}</a></strong><br><br>\n\n<small>{{.DisplayName}}:</small><br>\n<blockquote>{{.CommentSummary}}</blockquote><br>\n<a href='{{.CommentUrl}}'>View it on {{.SiteName}}</a><br><br>\n\n<small>You are receiving this because you authored the thread. <a href='{{.UnsubscribeUrl}}'>Unsubscribe</a></small>"}`},
		{ID: 35, Key: "tag.follow", Value: `0`},
		{ID: 36, Key: "rank.question.add", Value: `1`},
		{ID: 37, Key: "rank.question.edit", Value: `200`},
		{ID: 38, Key: "rank.question.delete", Value: `-1`},
		{ID: 39, Key: "rank.question.vote_up", Value: `15`},
		{ID: 40, Key: "rank.question.vote_down", Value: `125`},
		{ID: 41, Key: "rank.answer.add", Value: `1`},
		{ID: 42, Key: "rank.answer.edit", Value: `200`},
		{ID: 43, Key: "rank.answer.delete", Value: `-1`},
		{ID: 44, Key: "rank.answer.accept", Value: `1`},
		{ID: 45, Key: "rank.answer.vote_up", Value: `15`},
		{ID: 46, Key: "rank.answer.vote_down", Value: `125`},
		{ID: 47, Key: "rank.comment.add", Value: `1`},
		{ID: 48, Key: "rank.comment.edit", Value: `-1`},
		{ID: 49, Key: "rank.comment.delete", Value: `-1`},
		{ID: 50, Key: "rank.report.add", Value: `1`},
		{ID: 51, Key: "rank.tag.add", Value: `1`},
		{ID: 52, Key: "rank.tag.edit", Value: `100`},
		{ID: 53, Key: "rank.tag.delete", Value: `-1`},
		{ID: 54, Key: "rank.tag.synonym", Value: `20000`},
		{ID: 55, Key: "rank.link.url_limit", Value: `10`},
		{ID: 56, Key: "rank.vote.detail", Value: `0`},
		{ID: 57, Key: "reason.spam", Value: `{"name":"spam","description":"This post is an advertisement, or vandalism. It is not useful or relevant to the current topic."}`},
		{ID: 58, Key: "reason.rude_or_abusive", Value: `{"name":"rude or abusive","description":"A reasonable person would find this content inappropriate for respectful discourse."}`},
		{ID: 59, Key: "reason.something", Value: `{"name":"something else","description":"This post requires staff attention for another reason not listed above.","content_type":"textarea"}`},
		{ID: 60, Key: "reason.a_duplicate", Value: `{"name":"a duplicate","description":"This question has been asked before and already has an answer.","content_type":"text"}`},
		{ID: 61, Key: "reason.not_a_answer", Value: `{"name":"not a answer","description":"This was posted as an answer, but it does not attempt to answer the question. It should possibly be an edit, a comment, another question, or deleted altogether.","content_type":""}`},
		{ID: 62, Key: "reason.no_longer_needed", Value: `{"name":"no longer needed","description":"This comment is outdated, conversational or not relevant to this post."}`},
		{ID: 63, Key: "reason.community_specific", Value: `{"name":"a community-specific reason","description":"This question doesn’t meet a community guideline."}`},
		{ID: 64, Key: "reason.not_clarity", Value: `{"name":"needs details or clarity","description":"This question currently includes multiple questions in one. It should focus on one problem only.","content_type":"text"}`},
		{ID: 65, Key: "reason.normal", Value: `{"name":"normal","description":"A normal post available to everyone."}`},
		{ID: 66, Key: "reason.normal.user", Value: `{"name":"normal","description":"A normal user can ask and answer questions."}`},
		{ID: 67, Key: "reason.closed", Value: `{"name":"closed","description":"A closed question can’t answer, but still can edit, vote and comment."}`},
		{ID: 68, Key: "reason.deleted", Value: `{"name":"deleted","description":"All reputation gained and lost will be restored."}`},
		{ID: 69, Key: "reason.deleted.user", Value: `{"name":"deleted","description":"Delete profile, authentication associations."}`},
		{ID: 70, Key: "reason.suspended", Value: `{"name":"suspended","description":"A suspended user can’t log in."}`},
		{ID: 71, Key: "reason.inactive", Value: `{"name":"inactive","description":"An inactive user must re-validate their email."}`},
		{ID: 72, Key: "reason.looks_ok", Value: `{"name":"looks ok","description":"This post is good as-is and not low quality."}`},
		{ID: 73, Key: "reason.needs_edit", Value: `{"name":"needs edit, and I did it","description":"Improve and correct problems with this post yourself."}`},
		{ID: 74, Key: "reason.needs_close", Value: `{"name":"needs close","description":"A closed question can’t answer, but still can edit, vote and comment."}`},
		{ID: 75, Key: "reason.needs_delete", Value: `{"name":"needs delete","description":"All reputation gained and lost will be restored."}`},
		{ID: 76, Key: "question.flag.reasons", Value: `["reason.spam","reason.rude_or_abusive","reason.something","reason.a_duplicate"]`},
		{ID: 77, Key: "answer.flag.reasons", Value: `["reason.spam","reason.rude_or_abusive","reason.something","reason.not_a_answer"]`},
		{ID: 78, Key: "comment.flag.reasons", Value: `["reason.spam","reason.rude_or_abusive","reason.something","reason.no_longer_needed"]`},
		{ID: 79, Key: "question.close.reasons", Value: `["reason.a_duplicate","reason.community_specific","reason.not_clarity","reason.something"]`},
		{ID: 80, Key: "question.status.reasons", Value: `["reason.normal","reason.closed","reason.deleted"]`},
		{ID: 81, Key: "answer.status.reasons", Value: `["reason.normal","reason.deleted"]`},
		{ID: 82, Key: "comment.status.reasons", Value: `["reason.normal","reason.deleted"]`},
		{ID: 83, Key: "user.status.reasons", Value: `["reason.normal.user","reason.suspended","reason.deleted.user","reason.inactive"]`},
		{ID: 84, Key: "question.review.reasons", Value: `["reason.looks_ok","reason.needs_edit","reason.needs_close","reason.needs_delete"]`},
		{ID: 85, Key: "answer.review.reasons", Value: `["reason.looks_ok","reason.needs_edit","reason.needs_delete"]`},
		{ID: 86, Key: "comment.review.reasons", Value: `["reason.looks_ok","reason.needs_edit","reason.needs_delete"]`},
		{ID: 87, Key: "question.asked", Value: `0`},
		{ID: 88, Key: "question.closed", Value: `0`},
		{ID: 89, Key: "question.reopened", Value: `0`},
		{ID: 90, Key: "question.answered", Value: `0`},
		{ID: 91, Key: "question.commented", Value: `0`},
		{ID: 92, Key: "question.accept", Value: `0`},
		{ID: 93, Key: "question.edited", Value: `0`},
		{ID: 94, Key: "question.rollback", Value: `0`},
		{ID: 95, Key: "question.deleted", Value: `0`},
		{ID: 96, Key: "question.undeleted", Value: `0`},
		{ID: 97, Key: "answer.answered", Value: `0`},
		{ID: 98, Key: "answer.commented", Value: `0`},
		{ID: 99, Key: "answer.edited", Value: `0`},
		{ID: 100, Key: "answer.rollback", Value: `0`},
		{ID: 101, Key: "answer.undeleted", Value: `0`},
		{ID: 102, Key: "tag.created", Value: `0`},
		{ID: 103, Key: "tag.edited", Value: `0`},
		{ID: 104, Key: "tag.rollback", Value: `0`},
		{ID: 105, Key: "tag.deleted", Value: `0`},
		{ID: 106, Key: "tag.undeleted", Value: `0`},
		{ID: 107, Key: "rank.comment.vote_up", Value: `1`},
		{ID: 108, Key: "rank.comment.vote_down", Value: `1`},
		{ID: 109, Key: "rank.question.edit_without_review", Value: `2000`},
		{ID: 110, Key: "rank.answer.edit_without_review", Value: `2000`},
		{ID: 111, Key: "rank.tag.edit_without_review", Value: `20000`},
		{ID: 112, Key: "rank.answer.audit", Value: `2000`},
		{ID: 113, Key: "rank.question.audit", Value: `2000`},
		{ID: 114, Key: "rank.tag.audit", Value: `20000`},
		{ID: 115, Key: "rank.question.close", Value: `-1`},
		{ID: 116, Key: "rank.question.reopen", Value: `-1`},
		{ID: 117, Key: "rank.tag.use_reserved_tag", Value: `-1`},
	}
	_, err := engine.Insert(defaultConfigTable)
	return err
}

func initRolePower(engine *xorm.Engine) (err error) {
	roles := []*entity.Role{
		{ID: 1, Name: "User", Description: "Default with no special access."},
		{ID: 2, Name: "Admin", Description: "Have the full power to access the site."},
		{ID: 3, Name: "Moderator", Description: "Has access to all posts except admin settings."},
	}
	_, err = engine.Insert(roles)
	if err != nil {
		return err
	}

	powers := []*entity.Power{
		{ID: 1, Name: "admin access", PowerType: permission.AdminAccess, Description: "admin access"},
		{ID: 2, Name: "question add", PowerType: permission.QuestionAdd, Description: "question add"},
		{ID: 3, Name: "question edit", PowerType: permission.QuestionEdit, Description: "question edit"},
		{ID: 4, Name: "question edit without review", PowerType: permission.QuestionEditWithoutReview, Description: "question edit without review"},
		{ID: 5, Name: "question delete", PowerType: permission.QuestionDelete, Description: "question delete"},
		{ID: 6, Name: "question close", PowerType: permission.QuestionClose, Description: "question close"},
		{ID: 7, Name: "question reopen", PowerType: permission.QuestionReopen, Description: "question reopen"},
		{ID: 8, Name: "question vote up", PowerType: permission.QuestionVoteUp, Description: "question vote up"},
		{ID: 9, Name: "question vote down", PowerType: permission.QuestionVoteDown, Description: "question vote down"},
		{ID: 10, Name: "answer add", PowerType: permission.AnswerAdd, Description: "answer add"},
		{ID: 11, Name: "answer edit", PowerType: permission.AnswerEdit, Description: "answer edit"},
		{ID: 12, Name: "answer edit without review", PowerType: permission.AnswerEditWithoutReview, Description: "answer edit without review"},
		{ID: 13, Name: "answer delete", PowerType: permission.AnswerDelete, Description: "answer delete"},
		{ID: 14, Name: "answer accept", PowerType: permission.AnswerAccept, Description: "answer accept"},
		{ID: 15, Name: "answer vote up", PowerType: permission.AnswerVoteUp, Description: "answer vote up"},
		{ID: 16, Name: "answer vote down", PowerType: permission.AnswerVoteDown, Description: "answer vote down"},
		{ID: 17, Name: "comment add", PowerType: permission.CommentAdd, Description: "comment add"},
		{ID: 18, Name: "comment edit", PowerType: permission.CommentEdit, Description: "comment edit"},
		{ID: 19, Name: "comment delete", PowerType: permission.CommentDelete, Description: "comment delete"},
		{ID: 20, Name: "comment vote up", PowerType: permission.CommentVoteUp, Description: "comment vote up"},
		{ID: 21, Name: "comment vote down", PowerType: permission.CommentVoteDown, Description: "comment vote down"},
		{ID: 22, Name: "report add", PowerType: permission.ReportAdd, Description: "report add"},
		{ID: 23, Name: "tag add", PowerType: permission.TagAdd, Description: "tag add"},
		{ID: 24, Name: "tag edit", PowerType: permission.TagEdit, Description: "tag edit"},
		{ID: 25, Name: "tag edit without review", PowerType: permission.TagEditWithoutReview, Description: "tag edit without review"},
		{ID: 26, Name: "tag edit slug name", PowerType: permission.TagEditSlugName, Description: "tag edit slug name"},
		{ID: 27, Name: "tag delete", PowerType: permission.TagDelete, Description: "tag delete"},
		{ID: 28, Name: "tag synonym", PowerType: permission.TagSynonym, Description: "tag synonym"},
		{ID: 29, Name: "link url limit", PowerType: permission.LinkUrlLimit, Description: "link url limit"},
		{ID: 30, Name: "vote detail", PowerType: permission.VoteDetail, Description: "vote detail"},
		{ID: 31, Name: "answer audit", PowerType: permission.AnswerAudit, Description: "answer audit"},
		{ID: 32, Name: "question audit", PowerType: permission.QuestionAudit, Description: "question audit"},
		{ID: 33, Name: "tag audit", PowerType: permission.TagAudit, Description: "tag audit"},
	}
	_, err = engine.Insert(powers)
	if err != nil {
		return err
	}

	rolePowerRels := []*entity.RolePowerRel{
		{RoleID: 2, PowerType: permission.AdminAccess},
		{RoleID: 2, PowerType: permission.QuestionAdd},
		{RoleID: 2, PowerType: permission.QuestionEdit},
		{RoleID: 2, PowerType: permission.QuestionEditWithoutReview},
		{RoleID: 2, PowerType: permission.QuestionDelete},
		{RoleID: 2, PowerType: permission.QuestionClose},
		{RoleID: 2, PowerType: permission.QuestionReopen},
		{RoleID: 2, PowerType: permission.QuestionVoteUp},
		{RoleID: 2, PowerType: permission.QuestionVoteDown},
		{RoleID: 2, PowerType: permission.AnswerAdd},
		{RoleID: 2, PowerType: permission.AnswerEdit},
		{RoleID: 2, PowerType: permission.AnswerEditWithoutReview},
		{RoleID: 2, PowerType: permission.AnswerDelete},
		{RoleID: 2, PowerType: permission.AnswerAccept},
		{RoleID: 2, PowerType: permission.AnswerVoteUp},
		{RoleID: 2, PowerType: permission.AnswerVoteDown},
		{RoleID: 2, PowerType: permission.CommentAdd},
		{RoleID: 2, PowerType: permission.CommentEdit},
		{RoleID: 2, PowerType: permission.CommentDelete},
		{RoleID: 2, PowerType: permission.CommentVoteUp},
		{RoleID: 2, PowerType: permission.CommentVoteDown},
		{RoleID: 2, PowerType: permission.ReportAdd},
		{RoleID: 2, PowerType: permission.TagAdd},
		{RoleID: 2, PowerType: permission.TagEdit},
		{RoleID: 2, PowerType: permission.TagEditSlugName},
		{RoleID: 2, PowerType: permission.TagEditWithoutReview},
		{RoleID: 2, PowerType: permission.TagDelete},
		{RoleID: 2, PowerType: permission.TagSynonym},
		{RoleID: 2, PowerType: permission.LinkUrlLimit},
		{RoleID: 2, PowerType: permission.VoteDetail},
		{RoleID: 2, PowerType: permission.AnswerAudit},
		{RoleID: 2, PowerType: permission.QuestionAudit},
		{RoleID: 2, PowerType: permission.TagAudit},
		{RoleID: 2, PowerType: permission.TagUseReservedTag},

		{RoleID: 3, PowerType: permission.QuestionAdd},
		{RoleID: 3, PowerType: permission.QuestionEdit},
		{RoleID: 3, PowerType: permission.QuestionEditWithoutReview},
		{RoleID: 3, PowerType: permission.QuestionDelete},
		{RoleID: 3, PowerType: permission.QuestionClose},
		{RoleID: 3, PowerType: permission.QuestionReopen},
		{RoleID: 3, PowerType: permission.QuestionVoteUp},
		{RoleID: 3, PowerType: permission.QuestionVoteDown},
		{RoleID: 3, PowerType: permission.AnswerAdd},
		{RoleID: 3, PowerType: permission.AnswerEdit},
		{RoleID: 3, PowerType: permission.AnswerEditWithoutReview},
		{RoleID: 3, PowerType: permission.AnswerDelete},
		{RoleID: 3, PowerType: permission.AnswerAccept},
		{RoleID: 3, PowerType: permission.AnswerVoteUp},
		{RoleID: 3, PowerType: permission.AnswerVoteDown},
		{RoleID: 3, PowerType: permission.CommentAdd},
		{RoleID: 3, PowerType: permission.CommentEdit},
		{RoleID: 3, PowerType: permission.CommentDelete},
		{RoleID: 3, PowerType: permission.CommentVoteUp},
		{RoleID: 3, PowerType: permission.CommentVoteDown},
		{RoleID: 3, PowerType: permission.ReportAdd},
		{RoleID: 3, PowerType: permission.TagAdd},
		{RoleID: 3, PowerType: permission.TagEdit},
		{RoleID: 3, PowerType: permission.TagEditSlugName},
		{RoleID: 3, PowerType: permission.TagEditWithoutReview},
		{RoleID: 3, PowerType: permission.TagDelete},
		{RoleID: 3, PowerType: permission.TagSynonym},
		{RoleID: 3, PowerType: permission.LinkUrlLimit},
		{RoleID: 3, PowerType: permission.VoteDetail},
		{RoleID: 3, PowerType: permission.AnswerAudit},
		{RoleID: 3, PowerType: permission.QuestionAudit},
		{RoleID: 3, PowerType: permission.TagAudit},
		{RoleID: 3, PowerType: permission.TagUseReservedTag},
	}
	_, err = engine.Insert(rolePowerRels)
	if err != nil {
		return err
	}

	adminUserRoleRel := &entity.UserRoleRel{
		UserID: "1",
		RoleID: 2,
	}
	_, err = engine.Insert(adminUserRoleRel)
	if err != nil {
		return err
	}
	return nil
}
