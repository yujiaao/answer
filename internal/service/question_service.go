package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/answerdev/answer/internal/base/constant"
	"github.com/answerdev/answer/internal/base/data"
	"github.com/answerdev/answer/internal/base/handler"
	"github.com/answerdev/answer/internal/base/reason"
	"github.com/answerdev/answer/internal/base/translator"
	"github.com/answerdev/answer/internal/base/validator"
	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/schema"
	"github.com/answerdev/answer/internal/service/activity"
	"github.com/answerdev/answer/internal/service/activity_queue"
	collectioncommon "github.com/answerdev/answer/internal/service/collection_common"
	"github.com/answerdev/answer/internal/service/meta"
	"github.com/answerdev/answer/internal/service/notice_queue"
	"github.com/answerdev/answer/internal/service/permission"
	questioncommon "github.com/answerdev/answer/internal/service/question_common"
	"github.com/answerdev/answer/internal/service/revision_common"
	tagcommon "github.com/answerdev/answer/internal/service/tag_common"
	usercommon "github.com/answerdev/answer/internal/service/user_common"
	"github.com/answerdev/answer/pkg/htmltext"
	"github.com/answerdev/answer/pkg/uid"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/i18n"
	"github.com/segmentfault/pacman/log"
	"golang.org/x/net/context"
)

// QuestionRepo question repository

// QuestionService user service
type QuestionService struct {
	questionRepo          questioncommon.QuestionRepo
	tagCommon             *tagcommon.TagCommonService
	questioncommon        *questioncommon.QuestionCommon
	userCommon            *usercommon.UserCommon
	revisionService       *revision_common.RevisionService
	metaService           *meta.MetaService
	collectionCommon      *collectioncommon.CollectionCommon
	answerActivityService *activity.AnswerActivityService
	data                  *data.Data
}

func NewQuestionService(
	questionRepo questioncommon.QuestionRepo,
	tagCommon *tagcommon.TagCommonService,
	questioncommon *questioncommon.QuestionCommon,
	userCommon *usercommon.UserCommon,
	revisionService *revision_common.RevisionService,
	metaService *meta.MetaService,
	collectionCommon *collectioncommon.CollectionCommon,
	answerActivityService *activity.AnswerActivityService,
	data *data.Data,

) *QuestionService {
	return &QuestionService{
		questionRepo:          questionRepo,
		tagCommon:             tagCommon,
		questioncommon:        questioncommon,
		userCommon:            userCommon,
		revisionService:       revisionService,
		metaService:           metaService,
		collectionCommon:      collectionCommon,
		answerActivityService: answerActivityService,
		data:                  data,
	}
}

func (qs *QuestionService) CloseQuestion(ctx context.Context, req *schema.CloseQuestionReq) error {
	questionInfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	questionInfo.Status = entity.QuestionStatusClosed
	err = qs.questionRepo.UpdateQuestionStatus(ctx, questionInfo)
	if err != nil {
		return err
	}

	closeMeta, _ := json.Marshal(schema.CloseQuestionMeta{
		CloseType: req.CloseType,
		CloseMsg:  req.CloseMsg,
	})
	err = qs.metaService.AddMeta(ctx, req.ID, entity.QuestionCloseReasonKey, string(closeMeta))
	if err != nil {
		return err
	}

	activity_queue.AddActivity(&schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionClosed,
	})
	return nil
}

// ReopenQuestion reopen question
func (qs *QuestionService) ReopenQuestion(ctx context.Context, req *schema.ReopenQuestionReq) error {
	questionInfo, has, err := qs.questionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	questionInfo.Status = entity.QuestionStatusAvailable
	err = qs.questionRepo.UpdateQuestionStatus(ctx, questionInfo)
	if err != nil {
		return err
	}
	activity_queue.AddActivity(&schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionReopened,
	})
	return nil
}

// CloseMsgList list close question condition
func (qs *QuestionService) CloseMsgList(ctx context.Context, lang i18n.Language) (
	resp []*schema.GetCloseTypeResp, err error,
) {
	resp = make([]*schema.GetCloseTypeResp, 0)
	err = json.Unmarshal([]byte(constant.QuestionCloseJSON), &resp)
	if err != nil {
		return nil, errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}
	for _, t := range resp {
		t.Name = translator.Tr(lang, t.Name)
		t.Description = translator.Tr(lang, t.Description)
	}
	return resp, err
}

func (qs *QuestionService) AddQuestionCheckTags(ctx context.Context, Tags []*entity.Tag) ([]string, error) {
	list := make([]string, 0)
	for _, tag := range Tags {
		if tag.Reserved {
			list = append(list, tag.DisplayName)
		}
	}
	if len(list) > 0 {
		return list, errors.BadRequest(reason.RequestFormatError)
	}
	return []string{}, nil
}
func (qs *QuestionService) CheckAddQuestion(ctx context.Context, req *schema.QuestionAdd) (errorlist any, err error) {
	if len(req.Tags) == 0 {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.TagNotFound),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}
	recommendExist, err := qs.tagCommon.ExistRecommend(ctx, req.Tags)
	if err != nil {
		return
	}
	if !recommendExist {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.RecommendTagEnter),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}

	tagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tagNameList = append(tagNameList, tag.SlugName)
	}
	Tags, tagerr := qs.tagCommon.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		return errorlist, tagerr
	}
	if !req.QuestionPermission.CanUseReservedTag {
		taglist, err := qs.AddQuestionCheckTags(ctx, Tags)
		errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
			strings.Join(taglist, ","))
		if err != nil {
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RecommendTagEnter)
			return errorlist, err
		}
	}
	return nil, nil
}

// AddQuestion add question
func (qs *QuestionService) AddQuestion(ctx context.Context, req *schema.QuestionAdd) (questionInfo any, err error) {
	if len(req.Tags) == 0 {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.TagNotFound),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}
	recommendExist, err := qs.tagCommon.ExistRecommend(ctx, req.Tags)
	if err != nil {
		return
	}
	if !recommendExist {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.RecommendTagEnter),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}

	tagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tag.SlugName = strings.ReplaceAll(tag.SlugName, " ", "-")
		tagNameList = append(tagNameList, tag.SlugName)
	}
	Tags, tagerr := qs.tagCommon.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		return questionInfo, tagerr
	}
	if !req.QuestionPermission.CanUseReservedTag {
		taglist, err := qs.AddQuestionCheckTags(ctx, Tags)
		errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
			strings.Join(taglist, ","))
		if err != nil {
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RecommendTagEnter)
			return errorlist, err
		}
	}

	question := &entity.Question{}
	now := time.Now()
	question.UserID = req.UserID
	question.Title = req.Title
	question.OriginalText = req.Content
	question.ParsedText = req.HTML
	question.AcceptedAnswerID = "0"
	question.LastAnswerID = "0"
	question.LastEditUserID = "0"
	//question.PostUpdateTime = nil
	question.Status = entity.QuestionStatusAvailable
	question.RevisionID = "0"
	question.CreatedAt = now
	question.Pin = entity.QuestionUnPin
	question.Show = entity.QuestionShow
	//question.UpdatedAt = nil
	err = qs.questionRepo.AddQuestion(ctx, question)
	if err != nil {
		return
	}
	objectTagData := schema.TagChange{}
	objectTagData.ObjectID = question.ID
	objectTagData.Tags = req.Tags
	objectTagData.UserID = req.UserID
	err = qs.ChangeTag(ctx, &objectTagData)
	if err != nil {
		return
	}

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   question.UserID,
		ObjectID: question.ID,
		Title:    question.Title,
	}

	questionWithTagsRevision, err := qs.changeQuestionToRevision(ctx, question, Tags)
	if err != nil {
		return nil, err
	}
	infoJSON, _ := json.Marshal(questionWithTagsRevision)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := qs.revisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return
	}

	// user add question count
	err = qs.userCommon.UpdateQuestionCount(ctx, question.UserID, 1)
	if err != nil {
		log.Error("user IncreaseQuestionCount error", err.Error())
	}

	activity_queue.AddActivity(&schema.ActivityMsg{
		UserID:           question.UserID,
		ObjectID:         question.ID,
		OriginalObjectID: question.ID,
		ActivityTypeKey:  constant.ActQuestionAsked,
		RevisionID:       revisionID,
	})

	questionInfo, err = qs.GetQuestion(ctx, question.ID, question.UserID, req.QuestionPermission)
	return
}

// OperationQuestion
func (qs *QuestionService) OperationQuestion(ctx context.Context, req *schema.OperationQuestionReq) (err error) {
	questionInfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}
	// Hidden question cannot be placed at the top
	if questionInfo.Show == entity.QuestionHide && req.Operation == schema.QuestionOperationPin {
		return nil
	}
	// Question cannot be hidden when they are at the top
	if questionInfo.Pin == entity.QuestionPin && req.Operation == schema.QuestionOperationHide {
		return nil
	}

	switch req.Operation {
	case schema.QuestionOperationHide:
		questionInfo.Show = entity.QuestionHide
	case schema.QuestionOperationShow:
		questionInfo.Show = entity.QuestionShow
	case schema.QuestionOperationPin:
		questionInfo.Pin = entity.QuestionPin
	case schema.QuestionOperationUnPin:
		questionInfo.Pin = entity.QuestionUnPin
	}

	err = qs.questionRepo.UpdateQuestionOperation(ctx, questionInfo)
	if err != nil {
		return err
	}

	actMap := make(map[string]constant.ActivityTypeKey)
	actMap[schema.QuestionOperationPin] = constant.ActQuestionPin
	actMap[schema.QuestionOperationUnPin] = constant.ActQuestionUnPin
	actMap[schema.QuestionOperationHide] = constant.ActQuestionHide
	actMap[schema.QuestionOperationShow] = constant.ActQuestionShow
	_, ok := actMap[req.Operation]
	if ok {
		activity_queue.AddActivity(&schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  actMap[req.Operation],
		})
	}

	return nil
}

// RemoveQuestion delete question
func (qs *QuestionService) RemoveQuestion(ctx context.Context, req *schema.RemoveQuestionReq) (err error) {
	questionInfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	//if the status is deleted, return directly
	if questionInfo.Status == entity.QuestionStatusDeleted {
		return nil
	}
	if !has {
		return nil
	}
	if !req.IsAdmin {
		if questionInfo.UserID != req.UserID {
			return errors.BadRequest(reason.QuestionCannotDeleted)
		}

		if questionInfo.AcceptedAnswerID != "0" {
			return errors.BadRequest(reason.QuestionCannotDeleted)
		}
		if questionInfo.AnswerCount > 1 {
			return errors.BadRequest(reason.QuestionCannotDeleted)
		}

		if questionInfo.AnswerCount == 1 {
			answersearch := &entity.AnswerSearch{}
			answersearch.QuestionID = req.ID
			answerList, _, err := qs.questioncommon.AnswerCommon.Search(ctx, answersearch)
			if err != nil {
				return err
			}
			for _, answer := range answerList {
				if answer.VoteCount > 0 {
					return errors.BadRequest(reason.QuestionCannotDeleted)
				}
			}
		}
	}

	questionInfo.Status = entity.QuestionStatusDeleted
	err = qs.questionRepo.UpdateQuestionStatusWithOutUpdateTime(ctx, questionInfo)
	if err != nil {
		return err
	}

	// user add question count
	err = qs.userCommon.UpdateQuestionCount(ctx, questionInfo.UserID, -1)
	if err != nil {
		log.Error("user IncreaseQuestionCount error", err.Error())
	}

	err = qs.answerActivityService.DeleteQuestion(ctx, questionInfo.ID, questionInfo.CreatedAt, questionInfo.VoteCount)
	if err != nil {
		log.Errorf("user DeleteQuestion rank rollback error %s", err.Error())
	}
	activity_queue.AddActivity(&schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionDeleted,
	})
	return nil
}

func (qs *QuestionService) UpdateQuestionCheckTags(ctx context.Context, req *schema.QuestionUpdate) (errorlist []*validator.FormErrorField, err error) {
	dbinfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return
	}
	if !has {
		return
	}

	oldTags, tagerr := qs.tagCommon.GetObjectEntityTag(ctx, req.ID)
	if tagerr != nil {
		log.Error("GetObjectEntityTag error", tagerr)
		return nil, nil
	}

	tagNameList := make([]string, 0)
	oldtagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tagNameList = append(tagNameList, tag.SlugName)
	}
	for _, tag := range oldTags {
		oldtagNameList = append(oldtagNameList, tag.SlugName)
	}

	isChange := qs.tagCommon.CheckTagsIsChange(ctx, tagNameList, oldtagNameList)

	//If the content is the same, ignore it
	if dbinfo.Title == req.Title && dbinfo.OriginalText == req.Content && !isChange {
		return
	}

	Tags, tagerr := qs.tagCommon.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		log.Error("GetTagListByNames error", tagerr)
		return nil, nil
	}

	// if user can not use reserved tag, old reserved tag can not be removed and new reserved tag can not be added.
	if !req.CanUseReservedTag {
		CheckOldTag, CheckNewTag, CheckOldTaglist, CheckNewTaglist := qs.CheckChangeReservedTag(ctx, oldTags, Tags)
		if !CheckOldTag {
			errMsg := fmt.Sprintf(`The reserved tag "%s" must be present.`,
				strings.Join(CheckOldTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
		if !CheckNewTag {
			errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
				strings.Join(CheckNewTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
	}
	return nil, nil
}

// UpdateQuestion update question
func (qs *QuestionService) UpdateQuestion(ctx context.Context, req *schema.QuestionUpdate) (questionInfo any, err error) {
	var canUpdate bool
	questionInfo = &schema.QuestionInfo{}

	_, existUnreviewed, err := qs.revisionService.ExistUnreviewedByObjectID(ctx, req.ID)
	if err != nil {
		return

	}
	if existUnreviewed {
		err = errors.BadRequest(reason.QuestionCannotUpdate)
		return
	}

	dbinfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return
	}
	if !has {
		return
	}
	if dbinfo.Status == entity.QuestionStatusDeleted {
		err = errors.BadRequest(reason.QuestionCannotUpdate)
		return nil, err
	}

	now := time.Now()
	question := &entity.Question{}
	question.Title = req.Title
	question.OriginalText = req.Content
	question.ParsedText = req.HTML
	question.ID = uid.DeShortID(req.ID)
	question.UpdatedAt = now
	question.PostUpdateTime = now
	question.UserID = dbinfo.UserID
	question.LastEditUserID = req.UserID

	oldTags, tagerr := qs.tagCommon.GetObjectEntityTag(ctx, question.ID)
	if tagerr != nil {
		return questionInfo, tagerr
	}

	tagNameList := make([]string, 0)
	oldtagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tag.SlugName = strings.ReplaceAll(tag.SlugName, " ", "-")
		tagNameList = append(tagNameList, tag.SlugName)
	}
	for _, tag := range oldTags {
		oldtagNameList = append(oldtagNameList, tag.SlugName)
	}

	isChange := qs.tagCommon.CheckTagsIsChange(ctx, tagNameList, oldtagNameList)

	//If the content is the same, ignore it
	if dbinfo.Title == req.Title && dbinfo.OriginalText == req.Content && !isChange {
		return
	}

	Tags, tagerr := qs.tagCommon.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		return questionInfo, tagerr
	}

	// if user can not use reserved tag, old reserved tag can not be removed and new reserved tag can not be added.
	if !req.CanUseReservedTag {
		CheckOldTag, CheckNewTag, CheckOldTaglist, CheckNewTaglist := qs.CheckChangeReservedTag(ctx, oldTags, Tags)
		if !CheckOldTag {
			errMsg := fmt.Sprintf(`The reserved tag "%s" must be present.`,
				strings.Join(CheckOldTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
		if !CheckNewTag {
			errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
				strings.Join(CheckNewTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
	}
	// Check whether mandatory labels are selected
	recommendExist, err := qs.tagCommon.ExistRecommend(ctx, req.Tags)
	if err != nil {
		return
	}
	if !recommendExist {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.RecommendTagEnter),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}

	//Administrators and themselves do not need to be audited

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   question.UserID,
		ObjectID: question.ID,
		Title:    question.Title,
		Log:      req.EditSummary,
	}

	if req.NoNeedReview {
		canUpdate = true
	}

	// It's not you or the administrator that needs to be reviewed
	if !canUpdate {
		revisionDTO.Status = entity.RevisionUnreviewedStatus
		revisionDTO.UserID = req.UserID //use revision userid
	} else {
		//Direct modification
		revisionDTO.Status = entity.RevisionReviewPassStatus
		//update question to db
		saveerr := qs.questionRepo.UpdateQuestion(ctx, question, []string{"title", "original_text", "parsed_text", "updated_at", "post_update_time", "last_edit_user_id"})
		if saveerr != nil {
			return questionInfo, saveerr
		}
		objectTagData := schema.TagChange{}
		objectTagData.ObjectID = question.ID
		objectTagData.Tags = req.Tags
		objectTagData.UserID = req.UserID
		tagerr := qs.ChangeTag(ctx, &objectTagData)
		if err != nil {
			return questionInfo, tagerr
		}
	}

	questionWithTagsRevision, err := qs.changeQuestionToRevision(ctx, question, Tags)
	if err != nil {
		return nil, err
	}
	infoJSON, _ := json.Marshal(questionWithTagsRevision)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := qs.revisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return
	}
	if canUpdate {
		activity_queue.AddActivity(&schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         question.ID,
			ActivityTypeKey:  constant.ActQuestionEdited,
			RevisionID:       revisionID,
			OriginalObjectID: question.ID,
		})
	}

	questionInfo, err = qs.GetQuestion(ctx, question.ID, question.UserID, req.QuestionPermission)
	return
}

// GetQuestion get question one
func (qs *QuestionService) GetQuestion(ctx context.Context, questionID, userID string,
	per schema.QuestionPermission) (resp *schema.QuestionInfo, err error) {
	question, err := qs.questioncommon.Info(ctx, questionID, userID)
	if err != nil {
		return
	}
	// If the question is deleted, only the administrator and the author can view it
	if question.Status == entity.QuestionStatusDeleted && !per.CanReopen && question.UserID != userID {
		return nil, errors.NotFound(reason.QuestionNotFound)
	}
	if question.Status != entity.QuestionStatusClosed {
		per.CanReopen = false
	}
	if question.Status == entity.QuestionStatusClosed {
		per.CanClose = false
	}
	if question.Pin == entity.QuestionPin {
		per.CanPin = false
		per.CanHide = false
	}
	if question.Pin == entity.QuestionUnPin {
		per.CanUnPin = false
	}
	if question.Show == entity.QuestionShow {
		per.CanShow = false
	}
	if question.Show == entity.QuestionHide {
		per.CanHide = false
		per.CanPin = false
	}

	if question.Status == entity.QuestionStatusDeleted {
		operation := &schema.Operation{}
		operation.Msg = translator.Tr(handler.GetLangByCtx(ctx), reason.QuestionAlreadyDeleted)
		operation.Level = schema.OperationLevelDanger
		question.Operation = operation
	}

	question.Description = htmltext.FetchExcerpt(question.HTML, "...", 240)
	question.MemberActions = permission.GetQuestionPermission(ctx, userID, question.UserID,
		per.CanEdit, per.CanDelete, per.CanClose, per.CanReopen, per.CanPin, per.CanHide, per.CanUnPin, per.CanShow)
	return question, nil
}

// GetQuestionAndAddPV get question one
func (qs *QuestionService) GetQuestionAndAddPV(ctx context.Context, questionID, loginUserID string,
	per schema.QuestionPermission) (
	resp *schema.QuestionInfo, err error) {
	err = qs.questioncommon.UpdatePv(ctx, questionID)
	if err != nil {
		log.Error(err)
	}
	return qs.GetQuestion(ctx, questionID, loginUserID, per)
}

func (qs *QuestionService) ChangeTag(ctx context.Context, objectTagData *schema.TagChange) error {
	return qs.tagCommon.ObjectChangeTag(ctx, objectTagData)
}

func (qs *QuestionService) CheckChangeReservedTag(ctx context.Context, oldobjectTagData, objectTagData []*entity.Tag) (bool, bool, []string, []string) {
	return qs.tagCommon.CheckChangeReservedTag(ctx, oldobjectTagData, objectTagData)
}

func (qs *QuestionService) SearchUserList(ctx context.Context, userName, order string, page, pageSize int, loginUserID string) ([]*schema.UserQuestionInfo, int64, error) {
	userlist := make([]*schema.UserQuestionInfo, 0)

	userinfo, Exist, err := qs.userCommon.GetUserBasicInfoByUserName(ctx, userName)
	if err != nil {
		return userlist, 0, err
	}
	if !Exist {
		return userlist, 0, nil
	}
	search := &schema.QuestionPageReq{}
	search.OrderCond = order
	search.Page = page
	search.PageSize = pageSize
	search.UserIDBeSearched = userinfo.ID
	search.LoginUserID = loginUserID
	questionlist, count, err := qs.GetQuestionPage(ctx, search)
	if err != nil {
		return userlist, 0, err
	}
	for _, item := range questionlist {
		info := &schema.UserQuestionInfo{}
		_ = copier.Copy(info, item)
		status, ok := entity.AdminQuestionSearchStatusIntToString[item.Status]
		if ok {
			info.Status = status
		}
		userlist = append(userlist, info)
	}
	return userlist, count, nil
}

func (qs *QuestionService) SearchUserAnswerList(ctx context.Context, userName, order string, page, pageSize int, loginUserID string) ([]*schema.UserAnswerInfo, int64, error) {
	answerlist := make([]*schema.AnswerInfo, 0)
	userAnswerlist := make([]*schema.UserAnswerInfo, 0)
	userinfo, Exist, err := qs.userCommon.GetUserBasicInfoByUserName(ctx, userName)
	if err != nil {
		return userAnswerlist, 0, err
	}
	if !Exist {
		return userAnswerlist, 0, nil
	}
	answersearch := &entity.AnswerSearch{}
	answersearch.UserID = userinfo.ID
	answersearch.PageSize = pageSize
	answersearch.Page = page
	if order == "newest" {
		answersearch.Order = entity.AnswerSearchOrderByTime
	} else {
		answersearch.Order = entity.AnswerSearchOrderByDefault
	}
	questionIDs := make([]string, 0)
	answerList, count, err := qs.questioncommon.AnswerCommon.Search(ctx, answersearch)
	if err != nil {
		return userAnswerlist, count, err
	}
	for _, item := range answerList {
		answerinfo := qs.questioncommon.AnswerCommon.ShowFormat(ctx, item)
		answerlist = append(answerlist, answerinfo)
		questionIDs = append(questionIDs, uid.DeShortID(item.QuestionID))
	}
	questionMaps, err := qs.questioncommon.FindInfoByID(ctx, questionIDs, loginUserID)
	if err != nil {
		return userAnswerlist, count, err
	}

	for _, item := range answerlist {
		_, ok := questionMaps[item.QuestionID]
		if ok {
			item.QuestionInfo = questionMaps[item.QuestionID]
		}
		info := &schema.UserAnswerInfo{}
		_ = copier.Copy(info, item)
		info.AnswerID = item.ID
		info.QuestionID = item.QuestionID
		if item.QuestionInfo.Status != entity.QuestionStatusDeleted {
			userAnswerlist = append(userAnswerlist, info)
		}
	}

	return userAnswerlist, count, nil
}

func (qs *QuestionService) SearchUserCollectionList(ctx context.Context, page, pageSize int, loginUserID string) ([]*schema.QuestionInfo, int64, error) {
	list := make([]*schema.QuestionInfo, 0)
	userinfo, Exist, err := qs.userCommon.GetUserBasicInfoByID(ctx, loginUserID)
	if err != nil {
		return list, 0, err
	}
	if !Exist {
		return list, 0, nil
	}
	collectionSearch := &entity.CollectionSearch{}
	collectionSearch.UserID = userinfo.ID
	collectionSearch.Page = page
	collectionSearch.PageSize = pageSize
	collectionlist, count, err := qs.collectionCommon.SearchList(ctx, collectionSearch)
	if err != nil {
		return list, 0, err
	}
	questionIDs := make([]string, 0)
	for _, item := range collectionlist {
		questionIDs = append(questionIDs, item.ObjectID)
	}

	questionMaps, err := qs.questioncommon.FindInfoByID(ctx, questionIDs, loginUserID)
	if err != nil {
		return list, count, err
	}
	for _, id := range questionIDs {
		_, ok := questionMaps[uid.EnShortID(id)]
		if ok {
			questionMaps[uid.EnShortID(id)].LastAnsweredUserInfo = nil
			questionMaps[uid.EnShortID(id)].UpdateUserInfo = nil
			questionMaps[uid.EnShortID(id)].Content = ""
			questionMaps[uid.EnShortID(id)].HTML = ""
			list = append(list, questionMaps[uid.EnShortID(id)])
		}
	}

	return list, count, nil
}

func (qs *QuestionService) SearchUserTopList(ctx context.Context, userName string, loginUserID string) ([]*schema.UserQuestionInfo, []*schema.UserAnswerInfo, error) {
	answerlist := make([]*schema.AnswerInfo, 0)

	userAnswerlist := make([]*schema.UserAnswerInfo, 0)
	userQuestionlist := make([]*schema.UserQuestionInfo, 0)

	userinfo, Exist, err := qs.userCommon.GetUserBasicInfoByUserName(ctx, userName)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	if !Exist {
		return userQuestionlist, userAnswerlist, nil
	}
	search := &schema.QuestionPageReq{}
	search.OrderCond = "score"
	search.Page = 0
	search.PageSize = 5
	search.UserIDBeSearched = userinfo.ID
	search.LoginUserID = loginUserID
	questionlist, _, err := qs.GetQuestionPage(ctx, search)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	answersearch := &entity.AnswerSearch{}
	answersearch.UserID = userinfo.ID
	answersearch.PageSize = 5
	answersearch.Order = entity.AnswerSearchOrderByVote
	questionIDs := make([]string, 0)
	answerList, _, err := qs.questioncommon.AnswerCommon.Search(ctx, answersearch)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	for _, item := range answerList {
		answerinfo := qs.questioncommon.AnswerCommon.ShowFormat(ctx, item)
		answerlist = append(answerlist, answerinfo)
		questionIDs = append(questionIDs, item.QuestionID)
	}
	questionMaps, err := qs.questioncommon.FindInfoByID(ctx, questionIDs, loginUserID)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	for _, item := range answerlist {
		_, ok := questionMaps[item.QuestionID]
		if ok {
			item.QuestionInfo = questionMaps[item.QuestionID]
		}
	}

	for _, item := range questionlist {
		info := &schema.UserQuestionInfo{}
		_ = copier.Copy(info, item)
		info.UrlTitle = htmltext.UrlTitle(info.Title)
		userQuestionlist = append(userQuestionlist, info)
	}

	for _, item := range answerlist {
		info := &schema.UserAnswerInfo{}
		_ = copier.Copy(info, item)
		info.AnswerID = item.ID
		info.QuestionID = item.QuestionID
		info.QuestionInfo.UrlTitle = htmltext.UrlTitle(info.QuestionInfo.Title)
		userAnswerlist = append(userAnswerlist, info)
	}

	return userQuestionlist, userAnswerlist, nil
}

// SearchByTitleLike
func (qs *QuestionService) SearchByTitleLike(ctx context.Context, title string, loginUserID string) ([]*schema.QuestionBaseInfo, error) {
	list := make([]*schema.QuestionBaseInfo, 0)
	dblist, err := qs.questionRepo.SearchByTitleLike(ctx, title)
	if err != nil {
		return list, err
	}
	for _, question := range dblist {
		item := &schema.QuestionBaseInfo{}
		item.ID = question.ID
		item.Title = question.Title
		item.ViewCount = question.ViewCount
		item.AnswerCount = question.AnswerCount
		item.CollectionCount = question.CollectionCount
		item.FollowCount = question.FollowCount
		status, ok := entity.AdminQuestionSearchStatusIntToString[question.Status]
		if ok {
			item.Status = status
		}
		if question.AcceptedAnswerID != "0" {
			item.AcceptedAnswer = true
		}
		list = append(list, item)
	}

	return list, nil
}

// SimilarQuestion
func (qs *QuestionService) SimilarQuestion(ctx context.Context, questionID string, loginUserID string) ([]*schema.QuestionPageResp, int64, error) {
	question, err := qs.questioncommon.Info(ctx, questionID, loginUserID)
	if err != nil {
		return nil, 0, nil
	}
	tagNames := make([]string, 0, len(question.Tags))
	for _, tag := range question.Tags {
		tagNames = append(tagNames, tag.SlugName)
	}
	search := &schema.QuestionPageReq{}
	search.OrderCond = "frequent"
	search.Page = 0
	search.PageSize = 6
	if len(tagNames) > 0 {
		search.Tag = tagNames[0]
	}
	search.LoginUserID = loginUserID
	return qs.GetQuestionPage(ctx, search)
}

// GetQuestionPage query questions page
func (qs *QuestionService) GetQuestionPage(ctx context.Context, req *schema.QuestionPageReq) (
	questions []*schema.QuestionPageResp, total int64, err error) {
	questions = make([]*schema.QuestionPageResp, 0)

	// query by tag condition
	if len(req.Tag) > 0 {
		tagInfo, exist, err := qs.tagCommon.GetTagBySlugName(ctx, strings.ToLower(req.Tag))
		if err != nil {
			return nil, 0, err
		}
		if exist {
			req.TagID = tagInfo.ID
		}
	}

	// query by user condition
	if req.Username != "" {
		userinfo, exist, err := qs.userCommon.GetUserBasicInfoByUserName(ctx, req.Username)
		if err != nil {
			return nil, 0, err
		}
		if !exist {
			return questions, 0, nil
		}
		req.UserIDBeSearched = userinfo.ID
	}

	questionList, total, err := qs.questionRepo.GetQuestionPage(ctx, req.Page, req.PageSize,
		req.UserIDBeSearched, req.TagID, req.OrderCond)
	if err != nil {
		return nil, 0, err
	}
	questions, err = qs.questioncommon.FormatQuestionsPage(ctx, questionList, req.LoginUserID, req.OrderCond)
	if err != nil {
		return nil, 0, err
	}
	return questions, total, nil
}

func (qs *QuestionService) AdminSetQuestionStatus(ctx context.Context, questionID string, setStatusStr string) error {
	setStatus, ok := entity.AdminQuestionSearchStatus[setStatusStr]
	if !ok {
		return fmt.Errorf("question status does not exist")
	}
	questionInfo, exist, err := qs.questionRepo.GetQuestion(ctx, questionID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.QuestionNotFound)
	}
	err = qs.questionRepo.UpdateQuestionStatus(ctx, &entity.Question{ID: questionInfo.ID, Status: setStatus})
	if err != nil {
		return err
	}

	if setStatus == entity.QuestionStatusDeleted {
		err = qs.answerActivityService.DeleteQuestion(ctx, questionInfo.ID, questionInfo.CreatedAt, questionInfo.VoteCount)
		if err != nil {
			log.Errorf("admin delete question then rank rollback error %s", err.Error())
		}
		activity_queue.AddActivity(&schema.ActivityMsg{
			UserID:           questionInfo.UserID,
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionDeleted,
		})
	}
	if setStatus == entity.QuestionStatusAvailable && questionInfo.Status == entity.QuestionStatusClosed {
		activity_queue.AddActivity(&schema.ActivityMsg{
			UserID:           questionInfo.UserID,
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionReopened,
		})
	}
	if setStatus == entity.QuestionStatusClosed && questionInfo.Status != entity.QuestionStatusClosed {
		activity_queue.AddActivity(&schema.ActivityMsg{
			UserID:           questionInfo.UserID,
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionClosed,
		})
	}
	msg := &schema.NotificationMsg{}
	msg.ObjectID = questionInfo.ID
	msg.Type = schema.NotificationTypeInbox
	msg.ReceiverUserID = questionInfo.UserID
	msg.TriggerUserID = questionInfo.UserID
	msg.ObjectType = constant.QuestionObjectType
	msg.NotificationAction = constant.YourQuestionWasDeleted
	notice_queue.AddNotification(msg)
	return nil
}

func (qs *QuestionService) AdminSearchList(ctx context.Context, search *schema.AdminQuestionSearch, loginUserID string) ([]*schema.AdminQuestionInfo, int64, error) {
	list := make([]*schema.AdminQuestionInfo, 0)

	status, ok := entity.AdminQuestionSearchStatus[search.StatusStr]
	if ok {
		search.Status = status
	}

	if search.Status == 0 {
		search.Status = 1
	}
	dblist, count, err := qs.questionRepo.AdminSearchList(ctx, search)
	if err != nil {
		return list, count, err
	}
	userIds := make([]string, 0)
	for _, dbitem := range dblist {
		item := &schema.AdminQuestionInfo{}
		_ = copier.Copy(item, dbitem)
		item.CreateTime = dbitem.CreatedAt.Unix()
		item.UpdateTime = dbitem.PostUpdateTime.Unix()
		item.EditTime = dbitem.UpdatedAt.Unix()
		list = append(list, item)
		userIds = append(userIds, dbitem.UserID)
	}
	userInfoMap, err := qs.userCommon.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return list, count, err
	}
	for _, item := range list {
		_, ok = userInfoMap[item.UserID]
		if ok {
			item.UserInfo = userInfoMap[item.UserID]
		}
	}

	return list, count, nil
}

// AdminSearchList
func (qs *QuestionService) AdminSearchAnswerList(ctx context.Context, search *entity.AdminAnswerSearch, loginUserID string) ([]*schema.AdminAnswerInfo, int64, error) {
	answerlist := make([]*schema.AdminAnswerInfo, 0)

	status, ok := entity.AdminAnswerSearchStatus[search.StatusStr]
	if ok {
		search.Status = status
	}

	if search.Status == 0 {
		search.Status = 1
	}
	dblist, count, err := qs.questioncommon.AnswerCommon.AdminSearchList(ctx, search)
	if err != nil {
		return answerlist, count, err
	}
	questionIDs := make([]string, 0)
	userIds := make([]string, 0)
	for _, item := range dblist {
		answerinfo := qs.questioncommon.AnswerCommon.AdminShowFormat(ctx, item)
		answerlist = append(answerlist, answerinfo)
		questionIDs = append(questionIDs, item.QuestionID)
		userIds = append(userIds, item.UserID)
	}
	userInfoMap, err := qs.userCommon.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return answerlist, count, err
	}

	questionMaps, err := qs.questioncommon.FindInfoByID(ctx, questionIDs, loginUserID)
	if err != nil {
		return answerlist, count, err
	}
	for _, item := range answerlist {
		_, ok := questionMaps[item.QuestionID]
		if ok {
			item.QuestionInfo.Title = questionMaps[item.QuestionID].Title
		}
		_, ok = userInfoMap[item.UserID]
		if ok {
			item.UserInfo = userInfoMap[item.UserID]
		}
	}
	return answerlist, count, nil
}

func (qs *QuestionService) changeQuestionToRevision(ctx context.Context, questionInfo *entity.Question, tags []*entity.Tag) (
	questionRevision *entity.QuestionWithTagsRevision, err error) {
	questionRevision = &entity.QuestionWithTagsRevision{}
	questionRevision.Question = *questionInfo

	for _, tag := range tags {
		item := &entity.TagSimpleInfoForRevision{}
		_ = copier.Copy(item, tag)
		questionRevision.Tags = append(questionRevision.Tags, item)
	}
	return questionRevision, nil
}

func (qs *QuestionService) SitemapCron(ctx context.Context) {
	data := &schema.SiteMapList{}
	questionNum, err := qs.questionRepo.GetQuestionCount(ctx)
	if err != nil {
		log.Error("GetQuestionCount error", err)
		return
	}
	if questionNum <= schema.SitemapMaxSize {
		questionIDList, err := qs.questionRepo.GetQuestionIDsPage(ctx, 0, int(questionNum))
		if err != nil {
			log.Error("GetQuestionIDsPage error", err)
			return
		}
		data.QuestionIDs = questionIDList

	} else {
		nums := make([]int, 0)
		totalpages := int(math.Ceil(float64(questionNum) / float64(schema.SitemapMaxSize)))
		for i := 1; i <= totalpages; i++ {
			siteMapPagedata := &schema.SiteMapPageList{}
			nums = append(nums, i)
			questionIDList, err := qs.questionRepo.GetQuestionIDsPage(ctx, i, int(schema.SitemapMaxSize))
			if err != nil {
				log.Error("GetQuestionIDsPage error", err)
				return
			}
			siteMapPagedata.PageData = questionIDList
			if setCacheErr := qs.SetCache(ctx, fmt.Sprintf(schema.SitemapPageCachekey, i), siteMapPagedata); setCacheErr != nil {
				log.Errorf("set sitemap cron SetCache failed: %s", setCacheErr)
			}
		}
		data.MaxPageNum = nums
	}
	if setCacheErr := qs.SetCache(ctx, schema.SitemapCachekey, data); setCacheErr != nil {
		log.Errorf("set sitemap cron SetCache failed: %s", setCacheErr)
	}
}

func (qs *QuestionService) SetCache(ctx context.Context, cachekey string, info interface{}) error {
	infoStr, err := json.Marshal(info)
	if err != nil {
		return errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}

	err = qs.data.Cache.SetString(ctx, cachekey, string(infoStr), schema.DashBoardCacheTime)
	if err != nil {
		return errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}
	return nil
}
