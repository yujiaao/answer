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

package service

import (
	"github.com/apache/incubator-answer/internal/service/action"
	"github.com/apache/incubator-answer/internal/service/activity"
	"github.com/apache/incubator-answer/internal/service/activity_common"
	"github.com/apache/incubator-answer/internal/service/activity_queue"
	answercommon "github.com/apache/incubator-answer/internal/service/answer_common"
	"github.com/apache/incubator-answer/internal/service/auth"
	collectioncommon "github.com/apache/incubator-answer/internal/service/collection_common"
	"github.com/apache/incubator-answer/internal/service/comment"
	"github.com/apache/incubator-answer/internal/service/comment_common"
	"github.com/apache/incubator-answer/internal/service/config"
	"github.com/apache/incubator-answer/internal/service/dashboard"
	"github.com/apache/incubator-answer/internal/service/export"
	"github.com/apache/incubator-answer/internal/service/follow"
	"github.com/apache/incubator-answer/internal/service/meta"
	"github.com/apache/incubator-answer/internal/service/notice_queue"
	"github.com/apache/incubator-answer/internal/service/notification"
	notficationcommon "github.com/apache/incubator-answer/internal/service/notification_common"
	"github.com/apache/incubator-answer/internal/service/object_info"
	"github.com/apache/incubator-answer/internal/service/plugin_common"
	questioncommon "github.com/apache/incubator-answer/internal/service/question_common"
	"github.com/apache/incubator-answer/internal/service/rank"
	"github.com/apache/incubator-answer/internal/service/reason"
	"github.com/apache/incubator-answer/internal/service/report"
	"github.com/apache/incubator-answer/internal/service/report_admin"
	"github.com/apache/incubator-answer/internal/service/report_handle_admin"
	"github.com/apache/incubator-answer/internal/service/revision_common"
	"github.com/apache/incubator-answer/internal/service/role"
	"github.com/apache/incubator-answer/internal/service/search_parser"
	"github.com/apache/incubator-answer/internal/service/siteinfo"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	"github.com/apache/incubator-answer/internal/service/tag"
	tagcommon "github.com/apache/incubator-answer/internal/service/tag_common"
	"github.com/apache/incubator-answer/internal/service/uploader"
	"github.com/apache/incubator-answer/internal/service/user_admin"
	usercommon "github.com/apache/incubator-answer/internal/service/user_common"
	"github.com/apache/incubator-answer/internal/service/user_external_login"
	"github.com/apache/incubator-answer/internal/service/user_notification_config"
	"github.com/google/wire"
)

// ProviderSetService is providers.
var ProviderSetService = wire.NewSet(
	comment.NewCommentService,
	comment_common.NewCommentCommonService,
	report.NewReportService,
	NewVoteService,
	tag.NewTagService,
	follow.NewFollowService,
	NewCollectionGroupService,
	NewCollectionService,
	action.NewCaptchaService,
	auth.NewAuthService,
	NewUserService,
	NewQuestionService,
	NewAnswerService,
	export.NewEmailService,
	tagcommon.NewTagCommonService,
	usercommon.NewUserCommon,
	questioncommon.NewQuestionCommon,
	answercommon.NewAnswerCommon,
	uploader.NewUploaderService,
	collectioncommon.NewCollectionCommon,
	revision_common.NewRevisionService,
	NewRevisionService,
	rank.NewRankService,
	search_parser.NewSearchParser,
	NewSearchService,
	meta.NewMetaService,
	object_info.NewObjService,
	report_handle_admin.NewReportHandle,
	report_admin.NewReportAdminService,
	user_admin.NewUserAdminService,
	reason.NewReasonService,
	siteinfo_common.NewSiteInfoCommonService,
	siteinfo.NewSiteInfoService,
	notficationcommon.NewNotificationCommon,
	notification.NewNotificationService,
	activity.NewAnswerActivityService,
	dashboard.NewDashboardService,
	activity_common.NewActivityCommon,
	activity.NewActivityService,
	role.NewRoleService,
	role.NewUserRoleRelService,
	role.NewRolePowerRelService,
	user_external_login.NewUserExternalLoginService,
	user_external_login.NewUserCenterLoginService,
	plugin_common.NewPluginCommonService,
	config.NewConfigService,
	notice_queue.NewNotificationQueueService,
	activity_queue.NewActivityQueueService,
	user_notification_config.NewUserNotificationConfigService,
	notification.NewExternalNotificationService,
	notice_queue.NewNewQuestionNotificationQueueService,
)
