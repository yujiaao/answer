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

package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/apache/incubator-answer/internal/service/review"
	"github.com/apache/incubator-answer/internal/service/revision"
	"github.com/apache/incubator-answer/pkg/converter"
	"xorm.io/xorm/schemas"

	"github.com/apache/incubator-answer/internal/base/constant"
	"github.com/apache/incubator-answer/internal/base/data"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/activity_common"
	answercommon "github.com/apache/incubator-answer/internal/service/answer_common"
	"github.com/apache/incubator-answer/internal/service/comment_common"
	"github.com/apache/incubator-answer/internal/service/config"
	"github.com/apache/incubator-answer/internal/service/export"
	questioncommon "github.com/apache/incubator-answer/internal/service/question_common"
	"github.com/apache/incubator-answer/internal/service/report_common"
	"github.com/apache/incubator-answer/internal/service/service_config"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	usercommon "github.com/apache/incubator-answer/internal/service/user_common"
	"github.com/apache/incubator-answer/pkg/dir"
	"github.com/segmentfault/pacman/log"
)

type dashboardService struct {
	questionRepo    questioncommon.QuestionRepo
	answerRepo      answercommon.AnswerRepo
	commentRepo     comment_common.CommentCommonRepo
	voteRepo        activity_common.VoteRepo
	userRepo        usercommon.UserRepo
	reportRepo      report_common.ReportRepo
	configService   *config.ConfigService
	siteInfoService siteinfo_common.SiteInfoCommonService
	serviceConfig   *service_config.ServiceConfig
	reviewService   *review.ReviewService
	revisionRepo    revision.RevisionRepo
	data            *data.Data
}

func NewDashboardService(
	questionRepo questioncommon.QuestionRepo,
	answerRepo answercommon.AnswerRepo,
	commentRepo comment_common.CommentCommonRepo,
	voteRepo activity_common.VoteRepo,
	userRepo usercommon.UserRepo,
	reportRepo report_common.ReportRepo,
	configService *config.ConfigService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	serviceConfig *service_config.ServiceConfig,
	reviewService *review.ReviewService,
	revisionRepo revision.RevisionRepo,
	data *data.Data,
) DashboardService {
	return &dashboardService{
		questionRepo:    questionRepo,
		answerRepo:      answerRepo,
		commentRepo:     commentRepo,
		voteRepo:        voteRepo,
		userRepo:        userRepo,
		reportRepo:      reportRepo,
		configService:   configService,
		siteInfoService: siteInfoService,
		serviceConfig:   serviceConfig,
		reviewService:   reviewService,
		revisionRepo:    revisionRepo,
		data:            data,
	}
}

type DashboardService interface {
	Statistical(ctx context.Context) (resp *schema.DashboardInfo, err error)
}

func (ds *dashboardService) Statistical(ctx context.Context) (*schema.DashboardInfo, error) {
	dashboardInfo := ds.getFromCache(ctx)
	if dashboardInfo == nil {
		dashboardInfo = &schema.DashboardInfo{}
		dashboardInfo.QuestionCount = ds.questionCount(ctx)
		dashboardInfo.AnswerCount = ds.answerCount(ctx)
		dashboardInfo.CommentCount = ds.commentCount(ctx)
		dashboardInfo.UserCount = ds.userCount(ctx)
		dashboardInfo.VoteCount = ds.voteCount(ctx)
		dashboardInfo.OccupyingStorageSpace = ds.calculateStorage()
		general, err := ds.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get general site info failed: %s", err)
			return dashboardInfo, nil
		}
		if general.CheckUpdate {
			dashboardInfo.VersionInfo.RemoteVersion = ds.remoteVersion(ctx)
		}
		dashboardInfo.DatabaseVersion = ds.getDatabaseInfo()
		dashboardInfo.DatabaseSize = ds.GetDatabaseSize()
	}

	dashboardInfo.ReportCount = ds.reportCount(ctx)
	dashboardInfo.SMTP = ds.smtpStatus(ctx)
	dashboardInfo.HTTPS = ds.httpsStatus(ctx)
	dashboardInfo.TimeZone = ds.getTimezone(ctx)
	dashboardInfo.UploadingFiles = true
	dashboardInfo.AppStartTime = fmt.Sprintf("%d", time.Now().Unix()-schema.AppStartTime.Unix())
	dashboardInfo.VersionInfo.Version = constant.Version
	dashboardInfo.VersionInfo.Revision = constant.Revision
	dashboardInfo.GoVersion = constant.GoVersion
	if siteLogin, err := ds.siteInfoService.GetSiteLogin(ctx); err == nil {
		dashboardInfo.LoginRequired = siteLogin.LoginRequired
	}

	ds.setCache(ctx, dashboardInfo)
	return dashboardInfo, nil
}

func (ds *dashboardService) getFromCache(ctx context.Context) (dashboardInfo *schema.DashboardInfo) {
	infoStr, exist, err := ds.data.Cache.GetString(ctx, schema.DashboardCacheKey)
	if err != nil {
		log.Errorf("get dashboard statistical from cache failed: %s", err)
		return nil
	}
	if !exist {
		return nil
	}
	dashboardInfo = &schema.DashboardInfo{}
	if err = json.Unmarshal([]byte(infoStr), dashboardInfo); err != nil {
		return nil
	}
	return dashboardInfo
}

func (ds *dashboardService) setCache(ctx context.Context, info *schema.DashboardInfo) {
	infoStr, _ := json.Marshal(info)
	err := ds.data.Cache.SetString(ctx, schema.DashboardCacheKey, string(infoStr), schema.DashboardCacheTime)
	if err != nil {
		log.Errorf("set dashboard statistical failed: %s", err)
	}
}

func (ds *dashboardService) questionCount(ctx context.Context) int64 {
	questionCount, err := ds.questionRepo.GetQuestionCount(ctx)
	if err != nil {
		log.Errorf("get question count failed: %s", err)
	}
	return questionCount
}

func (ds *dashboardService) answerCount(ctx context.Context) int64 {
	answerCount, err := ds.answerRepo.GetAnswerCount(ctx)
	if err != nil {
		log.Errorf("get answer count failed: %s", err)
	}
	return answerCount
}

func (ds *dashboardService) commentCount(ctx context.Context) int64 {
	commentCount, err := ds.commentRepo.GetCommentCount(ctx)
	if err != nil {
		log.Errorf("get comment count failed: %s", err)
	}
	return commentCount
}

func (ds *dashboardService) userCount(ctx context.Context) int64 {
	userCount, err := ds.userRepo.GetUserCount(ctx)
	if err != nil {
		log.Errorf("get user count failed: %s", err)
	}
	return userCount
}

func (ds *dashboardService) reportCount(ctx context.Context) int64 {
	reviewCount, err := ds.reviewService.GetReviewPendingCount(ctx)
	if err != nil {
		log.Errorf("get review count failed: %s", err)
	}
	reportCount, err := ds.reportRepo.GetReportCount(ctx)
	if err != nil {
		log.Errorf("get report count failed: %s", err)
	}
	countUnreviewedRevision, err := ds.revisionRepo.CountUnreviewedRevision(ctx, []int{
		constant.ObjectTypeStrMapping[constant.AnswerObjectType],
		constant.ObjectTypeStrMapping[constant.QuestionObjectType],
		constant.ObjectTypeStrMapping[constant.TagObjectType],
	})
	if err != nil {
		log.Errorf("get revision count failed: %s", err)
	}
	return reviewCount + reportCount + countUnreviewedRevision
}

// count vote
func (ds *dashboardService) voteCount(ctx context.Context) int64 {
	typeKeys := []string{
		"question.vote_up",
		"question.vote_down",
		"answer.vote_up",
		"answer.vote_down",
	}
	var activityTypes []int
	for _, typeKey := range typeKeys {
		cfg, err := ds.configService.GetConfigByKey(ctx, typeKey)
		if err != nil {
			continue
		}
		activityTypes = append(activityTypes, cfg.ID)
	}
	voteCount, err := ds.voteRepo.GetVoteCount(ctx, activityTypes)
	if err != nil {
		log.Errorf("get vote count failed: %s", err)
	}
	return voteCount
}

func (ds *dashboardService) remoteVersion(ctx context.Context) string {
	req, _ := http.NewRequest("GET", "https://getlatest.answer.dev/", nil)
	req.Header.Set("User-Agent", "Answer/"+constant.Version)
	httpClient := &http.Client{}
	httpClient.Timeout = 15 * time.Second
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("request remote version failed: %s", err)
		return ""
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("read response body failed: %s", err)
		return ""
	}
	remoteVersion := &schema.RemoteVersion{}
	if err := json.Unmarshal(respByte, remoteVersion); err != nil {
		log.Errorf("parsing response body failed: %s", err)
		return ""
	}
	return remoteVersion.Release.Version
}

func (ds *dashboardService) smtpStatus(ctx context.Context) (smtpStatus string) {
	smtpStatus = "not_configured"
	emailConf, err := ds.configService.GetStringValue(ctx, "email.config")
	if err != nil {
		log.Errorf("get email config failed: %s", err)
		return "disabled"
	}
	ec := &export.EmailConfig{}
	err = json.Unmarshal([]byte(emailConf), ec)
	if err != nil {
		log.Errorf("parsing email config failed: %s", err)
		return "disabled"
	}
	if ec.SMTPHost != "" {
		smtpStatus = "enabled"
	}
	return smtpStatus
}

func (ds *dashboardService) httpsStatus(ctx context.Context) (enabled bool) {
	siteGeneral, err := ds.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Errorf("get site general failed: %s", err)
		return false
	}
	siteUrl, err := url.Parse(siteGeneral.SiteUrl)
	if err != nil {
		log.Errorf("parse site url failed: %s", err)
		return false
	}
	return siteUrl.Scheme == "https"
}

func (ds *dashboardService) getTimezone(ctx context.Context) string {
	siteInfoInterface, err := ds.siteInfoService.GetSiteInterface(ctx)
	if err != nil {
		return ""
	}
	return siteInfoInterface.TimeZone
}

func (ds *dashboardService) calculateStorage() string {
	dirSize, err := dir.DirSize(ds.serviceConfig.UploadPath)
	if err != nil {
		log.Errorf("get upload dir size failed: %s", err)
		return ""
	}
	return dir.FormatFileSize(dirSize)
}

func (ds *dashboardService) getDatabaseInfo() (versionDesc string) {
	dbVersion, err := ds.data.DB.DBVersion()
	if err != nil {
		log.Errorf("get db version failed: %s", err)
	} else {
		versionDesc = fmt.Sprintf("%s %s", ds.data.DB.Dialect().URI().DBType, dbVersion.Number)
	}
	return versionDesc
}

func (ds *dashboardService) GetDatabaseSize() (dbSize string) {
	switch ds.data.DB.Dialect().URI().DBType {
	case schemas.MYSQL:
		sql := fmt.Sprintf("SELECT SUM(DATA_LENGTH) as db_size FROM information_schema.TABLES WHERE table_schema = '%s'",
			ds.data.DB.Dialect().URI().DBName)
		res, err := ds.data.DB.QueryInterface(sql)
		if err != nil {
			log.Warnf("get db size failed: %s", err)
		} else {
			if res != nil && len(res) > 0 && res[0]["db_size"] != nil {
				dbSizeStr, _ := res[0]["db_size"].(string)
				dbSize = dir.FormatFileSize(converter.StringToInt64(dbSizeStr))
			}
		}
	case schemas.POSTGRES:
		sql := fmt.Sprintf("SELECT pg_database_size('%s') AS db_size",
			ds.data.DB.Dialect().URI().DBName)
		res, err := ds.data.DB.QueryInterface(sql)
		if err != nil {
			log.Warnf("get db size failed: %s", err)
		} else {
			if res != nil && len(res) > 0 && res[0]["db_size"] != nil {
				dbSizeStr, _ := res[0]["db_size"].(int32)
				dbSize = dir.FormatFileSize(int64(dbSizeStr))
			}
		}
	case schemas.SQLITE:
		dirSize, err := dir.DirSize(ds.data.DB.DataSourceName())
		if err != nil {
			log.Errorf("get upload dir size failed: %s", err)
			return ""
		}
		dbSize = dir.FormatFileSize(dirSize)
	}
	return dbSize
}
