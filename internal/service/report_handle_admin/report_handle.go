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

package report_handle_admin

import (
	"context"

	"github.com/apache/incubator-answer/internal/service/config"
	"github.com/apache/incubator-answer/internal/service/notice_queue"

	"github.com/apache/incubator-answer/internal/base/constant"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/comment"
	questioncommon "github.com/apache/incubator-answer/internal/service/question_common"
	"github.com/apache/incubator-answer/pkg/obj"
)

type ReportHandle struct {
	questionCommon           *questioncommon.QuestionCommon
	commentRepo              comment.CommentRepo
	configService            *config.ConfigService
	notificationQueueService notice_queue.NotificationQueueService
}

func NewReportHandle(
	questionCommon *questioncommon.QuestionCommon,
	commentRepo comment.CommentRepo,
	configService *config.ConfigService,
	notificationQueueService notice_queue.NotificationQueueService,
) *ReportHandle {
	return &ReportHandle{
		questionCommon:           questionCommon,
		commentRepo:              commentRepo,
		configService:            configService,
		notificationQueueService: notificationQueueService,
	}
}

// HandleObject this handle object status
func (rh *ReportHandle) HandleObject(ctx context.Context, reported *entity.Report, req schema.ReportHandleReq) (err error) {
	reasonDeleteCfg, err := rh.configService.GetConfigByKey(ctx, "reason.needs_delete")
	if err != nil {
		return err
	}
	reasonCloseCfg, err := rh.configService.GetConfigByKey(ctx, "reason.needs_close")
	if err != nil {
		return err
	}
	var (
		objectID       = reported.ObjectID
		reportedUserID = reported.ReportedUserID
		objectKey      string
	)

	objectKey, err = obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return err
	}
	switch objectKey {
	case "question":
		switch req.FlaggedType {
		case reasonDeleteCfg.ID:
			err = rh.questionCommon.RemoveQuestion(ctx, &schema.RemoveQuestionReq{ID: objectID})
		case reasonCloseCfg.ID:
			err = rh.questionCommon.CloseQuestion(ctx, &schema.CloseQuestionReq{
				ID:        objectID,
				CloseType: req.FlaggedType,
				CloseMsg:  req.FlaggedContent,
			})
		}
	case "answer":
		switch req.FlaggedType {
		case reasonDeleteCfg.ID:
			err = rh.questionCommon.RemoveAnswer(ctx, objectID)
		}
	case "comment":
		switch req.FlaggedType {
		case reasonCloseCfg.ID:
			err = rh.commentRepo.RemoveComment(ctx, objectID)
			rh.sendNotification(ctx, reportedUserID, objectID, constant.NotificationYourCommentWasDeleted)
		}
	}
	return
}

// sendNotification send rank triggered notification
func (rh *ReportHandle) sendNotification(ctx context.Context, reportedUserID, objectID, notificationAction string) {
	msg := &schema.NotificationMsg{
		TriggerUserID:      reportedUserID,
		ReceiverUserID:     reportedUserID,
		Type:               schema.NotificationTypeInbox,
		ObjectID:           objectID,
		ObjectType:         constant.ReportObjectType,
		NotificationAction: notificationAction,
	}
	rh.notificationQueueService.Send(ctx, msg)
}
