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

package controller

import (
	"fmt"
	"net/http"

	"github.com/apache/incubator-answer/internal/base/handler"
	"github.com/apache/incubator-answer/internal/base/middleware"
	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/base/translator"
	"github.com/apache/incubator-answer/internal/base/validator"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/action"
	"github.com/apache/incubator-answer/internal/service/content"
	"github.com/apache/incubator-answer/internal/service/permission"
	"github.com/apache/incubator-answer/internal/service/rank"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	"github.com/apache/incubator-answer/pkg/uid"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

// AnswerController answer controller
type AnswerController struct {
	answerService         *content.AnswerService
	rankService           *rank.RankService
	actionService         *action.CaptchaService
	siteInfoCommonService siteinfo_common.SiteInfoCommonService
	rateLimitMiddleware   *middleware.RateLimitMiddleware
}

// NewAnswerController new controller
func NewAnswerController(
	answerService *content.AnswerService,
	rankService *rank.RankService,
	actionService *action.CaptchaService,
	siteInfoCommonService siteinfo_common.SiteInfoCommonService,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
) *AnswerController {
	return &AnswerController{
		answerService:         answerService,
		rankService:           rankService,
		actionService:         actionService,
		siteInfoCommonService: siteInfoCommonService,
		rateLimitMiddleware:   rateLimitMiddleware,
	}
}

// RemoveAnswer delete answer
// @Summary delete answer
// @Description delete answer
// @Tags api-answer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.RemoveAnswerReq true "answer"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/answer [delete]
func (ac *AnswerController) RemoveAnswer(ctx *gin.Context) {
	req := &schema.RemoveAnswerReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ID = uid.DeShortID(req.ID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := ac.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionDelete, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	objectOwner := ac.rankService.CheckOperationObjectOwner(ctx, req.UserID, req.ID)
	canList, err := ac.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanDelete = canList[0] || objectOwner
	if !req.CanDelete {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = ac.answerService.RemoveAnswer(ctx, req)
	if !isAdmin {
		ac.actionService.ActionRecordAdd(ctx, entity.CaptchaActionDelete, req.UserID)
	}
	handler.HandleResponse(ctx, err, nil)
}

// RecoverAnswer recover answer
// @Summary recover answer
// @Description recover deleted answer
// @Tags Answer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.RecoverAnswerReq true "answer"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/answer/recover [post]
func (ac *AnswerController) RecoverAnswer(ctx *gin.Context) {
	req := &schema.RecoverAnswerReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.AnswerID = uid.DeShortID(req.AnswerID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := ac.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerUnDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !canList[0] {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = ac.answerService.RecoverAnswer(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// Get godoc
// @Summary Get Answer
// @Description Get Answer
// @Tags api-answer
// @Accept  json
// @Produce  json
// @Param id query string true "Answer TagID"  default(1)
// @Router  /answer/api/v1/answer/info [get]
// @Success 200 {string} string ""
func (ac *AnswerController) Get(ctx *gin.Context) {
	id := ctx.Query("id")
	id = uid.DeShortID(id)
	userID := middleware.GetLoginUserIDFromContext(ctx)

	info, questionInfo, has, err := ac.answerService.Get(ctx, id, userID)
	if err != nil {
		handler.HandleResponse(ctx, err, gin.H{})
		return
	}
	if !has {
		handler.HandleResponse(ctx, fmt.Errorf(""), gin.H{})
		return
	}
	handler.HandleResponse(ctx, err, gin.H{
		"info":     info,
		"question": questionInfo,
	})
}

// Add godoc
// @Summary Insert Answer
// @Description Insert Answer
// @Tags api-answer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param data body schema.AnswerAddReq  true "AnswerAddReq"
// @Success 200 {string} string ""
// @Router /answer/api/v1/answer [post]
func (ac *AnswerController) Add(ctx *gin.Context) {
	req := &schema.AnswerAddReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	reject, rejectKey := ac.rateLimitMiddleware.DuplicateRequestRejection(ctx, req)
	if reject {
		return
	}
	defer func() {
		// If status is not 200 means that the bad request has been returned, so the record should be cleared
		if ctx.Writer.Status() != http.StatusOK {
			ac.rateLimitMiddleware.DuplicateRequestClear(ctx, rejectKey)
		}
	}()
	req.QuestionID = uid.DeShortID(req.QuestionID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := ac.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerEdit,
		permission.AnswerDelete,
		permission.LinkUrlLimit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	linkUrlLimitUser := canList[2]
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin || !linkUrlLimitUser {
		captchaPass := ac.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionAnswer, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	can, err := ac.rankService.CheckOperationPermission(ctx, req.UserID, permission.AnswerAdd, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	write, err := ac.siteInfoCommonService.GetSiteWrite(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if write.RestrictAnswer {
		// check if there's already an answer by this user
		ids, err := ac.answerService.GetCountByUserIDQuestionID(ctx, req.UserID, req.QuestionID)
		if err != nil {
			handler.HandleResponse(ctx, err, nil)
			return
		}
		if len(ids) >= 1 {
			handler.HandleResponse(ctx, errors.Forbidden(reason.AnswerRestrictAnswer), nil)
			return
		}
	}

	req.UserAgent = ctx.GetHeader("User-Agent")
	req.IP = ctx.ClientIP()

	answerID, err := ac.answerService.Insert(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !isAdmin || !linkUrlLimitUser {
		ac.actionService.ActionRecordAdd(ctx, entity.CaptchaActionAnswer, req.UserID)
	}
	info, questionInfo, has, err := ac.answerService.Get(ctx, answerID, req.UserID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !has {
		handler.HandleResponse(ctx, nil, nil)
		return
	}

	objectOwner := ac.rankService.CheckOperationObjectOwner(ctx, req.UserID, info.ID)
	req.CanEdit = canList[0] || objectOwner
	req.CanDelete = canList[1] || objectOwner
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	info.MemberActions = permission.GetAnswerPermission(ctx, req.UserID, info.UserID,
		0, req.CanEdit, req.CanDelete, false)
	handler.HandleResponse(ctx, nil, gin.H{
		"info":     info,
		"question": questionInfo,
	})
}

// Update godoc
// @Summary Update Answer
// @Description Update Answer
// @Tags api-answer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param data body schema.AnswerUpdateReq  true "AnswerUpdateReq"
// @Success 200 {string} string ""
// @Router /answer/api/v1/answer [put]
func (ac *AnswerController) Update(ctx *gin.Context) {
	req := &schema.AnswerUpdateReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := ac.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerEdit,
		permission.AnswerEditWithoutReview,
		permission.LinkUrlLimit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.QuestionID = uid.DeShortID(req.QuestionID)
	linkUrlLimitUser := canList[2]
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin || !linkUrlLimitUser {
		captchaPass := ac.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEdit, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	objectOwner := ac.rankService.CheckOperationObjectOwner(ctx, req.UserID, req.ID)
	req.CanEdit = canList[0] || objectOwner
	req.NoNeedReview = canList[1] || objectOwner
	if !req.CanEdit {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	_, err = ac.answerService.Update(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !isAdmin || !linkUrlLimitUser {
		ac.actionService.ActionRecordAdd(ctx, entity.CaptchaActionEdit, req.UserID)
	}
	_, _, _, err = ac.answerService.Get(ctx, req.ID, req.UserID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	handler.HandleResponse(ctx, nil, &schema.AnswerUpdateResp{WaitForReview: !req.NoNeedReview})
}

// AnswerList godoc
// @Summary AnswerList
// @Description AnswerList <br> <b>order</b> (default or updated)
// @Tags api-answer
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param question_id query string true "question_id"
// @Param order query string true "order"
// @Param page query string true "page"
// @Param page_size query string true "page_size"
// @Success 200 {string} string ""
// @Router /answer/api/v1/answer/page [get]
func (ac *AnswerController) AnswerList(ctx *gin.Context) {
	req := &schema.AnswerListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.QuestionID = uid.DeShortID(req.QuestionID)

	canList, err := ac.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerEdit,
		permission.AnswerDelete,
		permission.AnswerUnDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanEdit = canList[0]
	req.CanDelete = canList[1]
	req.CanRecover = canList[2]

	list, count, err := ac.answerService.SearchList(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	handler.HandleResponse(ctx, nil, gin.H{
		"list":  list,
		"count": count,
	})
}

// Accepted godoc
// @Summary Accepted
// @Description Accepted
// @Tags api-answer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param data body schema.AcceptAnswerReq  true "AcceptAnswerReq"
// @Success 200 {string} string ""
// @Router /answer/api/v1/answer/acceptance [post]
func (ac *AnswerController) Accepted(ctx *gin.Context) {
	req := &schema.AcceptAnswerReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.AnswerID = uid.DeShortID(req.AnswerID)
	req.QuestionID = uid.DeShortID(req.QuestionID)
	can, err := ac.rankService.CheckOperationPermission(ctx, req.UserID, permission.AnswerAccept, req.QuestionID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = ac.answerService.AcceptAnswer(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// AdminUpdateAnswerStatus update answer status
// @Summary update answer status
// @Description update answer status
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.AdminUpdateAnswerStatusReq true "AdminUpdateAnswerStatusReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/answer/status [put]
func (ac *AnswerController) AdminUpdateAnswerStatus(ctx *gin.Context) {
	req := &schema.AdminUpdateAnswerStatusReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.AnswerID = uid.DeShortID(req.AnswerID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	err := ac.answerService.AdminSetAnswerStatus(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
