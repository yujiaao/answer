package report_admin

import (
	"context"

	"github.com/answerdev/answer/internal/service/config"
	"github.com/answerdev/answer/pkg/htmltext"
	"github.com/segmentfault/pacman/log"

	"github.com/answerdev/answer/internal/base/pager"
	"github.com/answerdev/answer/internal/base/reason"
	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/repo/common"
	"github.com/answerdev/answer/internal/schema"
	answercommon "github.com/answerdev/answer/internal/service/answer_common"
	"github.com/answerdev/answer/internal/service/comment_common"
	questioncommon "github.com/answerdev/answer/internal/service/question_common"
	"github.com/answerdev/answer/internal/service/report_common"
	"github.com/answerdev/answer/internal/service/report_handle_admin"
	usercommon "github.com/answerdev/answer/internal/service/user_common"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
)

// ReportAdminService user service
type ReportAdminService struct {
	reportRepo        report_common.ReportRepo
	commonUser        *usercommon.UserCommon
	commonRepo        *common.CommonRepo
	answerRepo        answercommon.AnswerRepo
	questionRepo      questioncommon.QuestionRepo
	commentCommonRepo comment_common.CommentCommonRepo
	reportHandle      *report_handle_admin.ReportHandle
	configRepo        config.ConfigRepo
}

// NewReportAdminService new report service
func NewReportAdminService(
	reportRepo report_common.ReportRepo,
	commonUser *usercommon.UserCommon,
	commonRepo *common.CommonRepo,
	answerRepo answercommon.AnswerRepo,
	questionRepo questioncommon.QuestionRepo,
	commentCommonRepo comment_common.CommentCommonRepo,
	reportHandle *report_handle_admin.ReportHandle,
	configRepo config.ConfigRepo) *ReportAdminService {
	return &ReportAdminService{
		reportRepo:        reportRepo,
		commonUser:        commonUser,
		commonRepo:        commonRepo,
		answerRepo:        answerRepo,
		questionRepo:      questionRepo,
		commentCommonRepo: commentCommonRepo,
		reportHandle:      reportHandle,
		configRepo:        configRepo,
	}
}

// ListReportPage list report pages
func (rs *ReportAdminService) ListReportPage(ctx context.Context, dto schema.GetReportListPageDTO) (pageModel *pager.PageModel, err error) {
	var (
		resp  []*schema.GetReportListPageResp
		flags []entity.Report
		total int64

		flaggedUserIds,
		userIds []string

		flaggedUsers,
		users map[string]*schema.UserBasicInfo
	)

	pageModel = &pager.PageModel{}

	flags, total, err = rs.reportRepo.GetReportListPage(ctx, dto)
	if err != nil {
		return
	}

	_ = copier.Copy(&resp, flags)
	for _, r := range resp {
		flaggedUserIds = append(flaggedUserIds, r.ReportedUserID)
		userIds = append(userIds, r.UserID)
		r.Format()
	}

	// flagged users
	flaggedUsers, err = rs.commonUser.BatchUserBasicInfoByID(ctx, flaggedUserIds)
	if err != nil {
		return nil, err
	}

	// flag users
	users, err = rs.commonUser.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return nil, err
	}
	for _, r := range resp {
		r.ReportedUser = flaggedUsers[r.ReportedUserID]
		r.ReportUser = users[r.UserID]
	}

	rs.parseObject(ctx, &resp)
	return pager.NewPageModel(total, resp), nil
}

// HandleReported handle the reported object
func (rs *ReportAdminService) HandleReported(ctx context.Context, req schema.ReportHandleReq) (err error) {
	var (
		reported   *entity.Report
		handleData = entity.Report{
			FlaggedContent: req.FlaggedContent,
			FlaggedType:    req.FlaggedType,
			Status:         entity.ReportStatusCompleted,
		}
		exist bool
	)

	reported, exist, err = rs.reportRepo.GetByID(ctx, req.ID)
	if err != nil {
		err = errors.BadRequest(reason.ReportHandleFailed).WithError(err).WithStack()
		return
	}
	if !exist {
		err = errors.NotFound(reason.ReportNotFound)
		return
	}

	// check if handle or not
	if reported.Status != entity.ReportStatusPending {
		return
	}

	if err = rs.reportHandle.HandleObject(ctx, reported, req); err != nil {
		return
	}

	err = rs.reportRepo.UpdateByID(ctx, reported.ID, handleData)
	return
}

func (rs *ReportAdminService) parseObject(ctx context.Context, resp *[]*schema.GetReportListPageResp) {
	var (
		res = *resp
	)

	for i, r := range res {
		var (
			objIds map[string]string
			exists,
			ok bool
			err error
			questionId,
			answerId,
			commentId string
			question *entity.Question
			answer   *entity.Answer
			cmt      *entity.Comment
		)

		objIds, err = rs.commonRepo.GetObjectIDMap(r.ObjectID)
		if err != nil {
			log.Error(err)
			continue
		}

		questionId, ok = objIds["question"]
		if !ok {
			continue
		}

		question, exists, err = rs.questionRepo.GetQuestion(ctx, questionId)
		if err != nil || !exists {
			continue
		}

		answerId, ok = objIds["answer"]
		if ok {
			answer, _, err = rs.answerRepo.GetAnswer(ctx, answerId)
			if err != nil {
				log.Error(err)
				continue
			}
		}

		commentId, ok = objIds["comment"]
		if ok {
			cmt, _, err = rs.commentCommonRepo.GetComment(ctx, commentId)
			if err != nil {
				log.Error(err)
				continue
			}
		}

		switch r.OType {
		case "question":
			r.QuestionID = questionId
			r.Title = question.Title
			r.Excerpt = htmltext.FetchExcerpt(question.ParsedText, "...", 240)

		case "answer":
			r.QuestionID = questionId
			r.AnswerID = answerId
			r.Title = question.Title
			r.Excerpt = htmltext.FetchExcerpt(answer.ParsedText, "...", 240)

		case "comment":
			r.QuestionID = questionId
			r.AnswerID = answerId
			r.CommentID = commentId
			r.Title = question.Title
			r.Excerpt = htmltext.FetchExcerpt(cmt.ParsedText, "...", 240)
		}

		// parse reason
		if r.ReportType > 0 {
			r.Reason = &schema.ReasonItem{
				ReasonType: r.ReportType,
			}
			err = rs.configRepo.GetJsonConfigByIDAndSetToObject(r.ReportType, r.Reason)
			if err != nil {
				log.Error(err)
			}
		}
		if r.FlaggedType > 0 {
			r.FlaggedReason = &schema.ReasonItem{
				ReasonType: r.FlaggedType,
			}
			err = rs.configRepo.GetJsonConfigByIDAndSetToObject(r.FlaggedType, r.FlaggedReason)
			if err != nil {
				log.Error(err)
			}
		}

		res[i] = r
	}
}
