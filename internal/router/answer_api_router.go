package router

import (
	"github.com/answerdev/answer/internal/controller"
	"github.com/answerdev/answer/internal/controller_admin"
	"github.com/gin-gonic/gin"
)

type AnswerAPIRouter struct {
	langController         *controller.LangController
	userController         *controller.UserController
	commentController      *controller.CommentController
	reportController       *controller.ReportController
	voteController         *controller.VoteController
	tagController          *controller.TagController
	followController       *controller.FollowController
	collectionController   *controller.CollectionController
	questionController     *controller.QuestionController
	answerController       *controller.AnswerController
	searchController       *controller.SearchController
	revisionController     *controller.RevisionController
	rankController         *controller.RankController
	adminReportController  *controller_admin.ReportController
	adminUserController    *controller_admin.UserAdminController
	reasonController       *controller.ReasonController
	themeController        *controller_admin.ThemeController
	siteInfoController     *controller_admin.SiteInfoController
	siteinfoController     *controller.SiteinfoController
	notificationController *controller.NotificationController
	dashboardController    *controller.DashboardController
	uploadController       *controller.UploadController
	activityController     *controller.ActivityController
	roleController         *controller_admin.RoleController
}

func NewAnswerAPIRouter(
	langController *controller.LangController,
	userController *controller.UserController,
	commentController *controller.CommentController,
	reportController *controller.ReportController,
	voteController *controller.VoteController,
	tagController *controller.TagController,
	followController *controller.FollowController,
	collectionController *controller.CollectionController,
	questionController *controller.QuestionController,
	answerController *controller.AnswerController,
	searchController *controller.SearchController,
	revisionController *controller.RevisionController,
	rankController *controller.RankController,
	adminReportController *controller_admin.ReportController,
	adminUserController *controller_admin.UserAdminController,
	reasonController *controller.ReasonController,
	themeController *controller_admin.ThemeController,
	siteInfoController *controller_admin.SiteInfoController,
	siteinfoController *controller.SiteinfoController,
	notificationController *controller.NotificationController,
	dashboardController *controller.DashboardController,
	uploadController *controller.UploadController,
	activityController *controller.ActivityController,
	roleController *controller_admin.RoleController,
) *AnswerAPIRouter {
	return &AnswerAPIRouter{
		langController:         langController,
		userController:         userController,
		commentController:      commentController,
		reportController:       reportController,
		voteController:         voteController,
		tagController:          tagController,
		followController:       followController,
		collectionController:   collectionController,
		questionController:     questionController,
		answerController:       answerController,
		searchController:       searchController,
		revisionController:     revisionController,
		rankController:         rankController,
		adminReportController:  adminReportController,
		adminUserController:    adminUserController,
		reasonController:       reasonController,
		themeController:        themeController,
		siteInfoController:     siteInfoController,
		notificationController: notificationController,
		siteinfoController:     siteinfoController,
		dashboardController:    dashboardController,
		uploadController:       uploadController,
		activityController:     activityController,
		roleController:         roleController,
	}
}

func (a *AnswerAPIRouter) RegisterMustUnAuthAnswerAPIRouter(r *gin.RouterGroup) {
	// i18n
	r.GET("/language/config", a.langController.GetLangMapping)
	r.GET("/language/options", a.langController.GetUserLangOptions)

	//siteinfo
	r.GET("/siteinfo", a.siteinfoController.GetSiteInfo)
	r.GET("/siteinfo/legal", a.siteinfoController.GetSiteLegalInfo)

	// user
	r.POST("/user/login/email", a.userController.UserEmailLogin)
	r.POST("/user/register/email", a.userController.UserRegisterByEmail)
	r.GET("/user/register/captcha", a.userController.UserRegisterCaptcha)
	r.POST("/user/email/verification", a.userController.UserVerifyEmail)
	r.PUT("/user/email", a.userController.UserChangeEmailVerify)
	r.GET("/user/action/record", a.userController.ActionRecord)
	r.POST("/user/password/reset", a.userController.RetrievePassWord)
	r.POST("/user/password/replacement", a.userController.UseRePassWord)
	r.GET("/user/info", a.userController.GetUserInfoByUserID)
	r.PUT("/user/email/notification", a.userController.UserUnsubscribeEmailNotification)
}

func (a *AnswerAPIRouter) RegisterUnAuthAnswerAPIRouter(r *gin.RouterGroup) {
	// user
	r.GET("/user/logout", a.userController.UserLogout)
	r.POST("/user/email/change/code", a.userController.UserChangeEmailSendCode)
	r.POST("/user/email/verification/send", a.userController.UserVerifyEmailSend)
	r.GET("/personal/user/info", a.userController.GetOtherUserInfoByUsername)
	r.GET("/user/ranking", a.userController.UserRanking)

	//answer
	r.GET("/answer/info", a.answerController.Get)
	r.GET("/answer/page", a.answerController.AnswerList)
	r.GET("/personal/answer/page", a.questionController.UserAnswerList)

	//question
	r.GET("/question/info", a.questionController.GetQuestion)
	r.GET("/question/page", a.questionController.QuestionPage)
	r.GET("/question/similar/tag", a.questionController.SimilarQuestion)
	r.GET("/personal/qa/top", a.questionController.UserTop)
	r.GET("/personal/question/page", a.questionController.UserList)

	// comment
	r.GET("/comment/page", a.commentController.GetCommentWithPage)
	r.GET("/personal/comment/page", a.commentController.GetCommentPersonalWithPage)
	r.GET("/comment", a.commentController.GetComment)

	//revision
	r.GET("/revisions", a.revisionController.GetRevisionList)

	// tag
	r.GET("/tags/page", a.tagController.GetTagWithPage)
	r.GET("/tags/following", a.tagController.GetFollowingTags)
	r.GET("/tag", a.tagController.GetTagInfo)
	r.GET("/tags", a.tagController.GetTagsBySlugName)
	r.GET("/tag/synonyms", a.tagController.GetTagSynonyms)

	//search
	r.GET("/search", a.searchController.Search)

	//rank
	r.GET("/personal/rank/page", a.rankController.GetRankPersonalWithPage)
}

func (a *AnswerAPIRouter) RegisterAnswerAPIRouter(r *gin.RouterGroup) {
	//revisions
	r.GET("/revisions/unreviewed", a.revisionController.GetUnreviewedRevisionList)
	r.PUT("/revisions/audit", a.revisionController.RevisionAudit)
	r.GET("/revisions/edit/check", a.revisionController.CheckCanUpdateRevision)

	// comment
	r.POST("/comment", a.commentController.AddComment)
	r.DELETE("/comment", a.commentController.RemoveComment)
	r.PUT("/comment", a.commentController.UpdateComment)

	// report
	r.POST("/report", a.reportController.AddReport)

	// vote
	r.POST("/vote/up", a.voteController.VoteUp)
	r.POST("/vote/down", a.voteController.VoteDown)

	// follow
	r.POST("/follow", a.followController.Follow)
	r.PUT("/follow/tags", a.followController.UpdateFollowTags)

	// tag
	r.GET("/question/tags", a.tagController.SearchTagLike)
	r.POST("/tag", a.tagController.AddTag)
	r.PUT("/tag", a.tagController.UpdateTag)
	r.DELETE("/tag", a.tagController.RemoveTag)
	r.PUT("/tag/synonym", a.tagController.UpdateTagSynonym)

	// collection
	r.POST("/collection/switch", a.collectionController.CollectionSwitch)
	r.GET("/personal/collection/page", a.questionController.UserCollectionList)

	// question
	r.POST("/question", a.questionController.AddQuestion)
	r.POST("/question/answer", a.questionController.AddQuestionByAnswer)
	r.PUT("/question", a.questionController.UpdateQuestion)
	r.DELETE("/question", a.questionController.RemoveQuestion)
	r.PUT("/question/status", a.questionController.CloseQuestion)
	r.PUT("/question/operation", a.questionController.OperationQuestion)
	r.PUT("/question/reopen", a.questionController.ReopenQuestion)
	r.GET("/question/similar", a.questionController.SearchByTitleLike)

	// answer
	r.POST("/answer", a.answerController.Add)
	r.PUT("/answer", a.answerController.Update)
	r.POST("/answer/acceptance", a.answerController.Accepted)
	r.DELETE("/answer", a.answerController.RemoveAnswer)

	// user
	r.PUT("/user/password", a.userController.UserModifyPassWord)
	r.PUT("/user/info", a.userController.UserUpdateInfo)
	r.PUT("/user/interface", a.userController.UserUpdateInterface)
	r.POST("/user/notice/set", a.userController.UserNoticeSet)

	// vote
	r.GET("/personal/vote/page", a.voteController.UserVotes)

	// reason
	r.GET("/reasons", a.reasonController.Reasons)

	// notification
	r.GET("/notification/status", a.notificationController.GetRedDot)
	r.PUT("/notification/status", a.notificationController.ClearRedDot)
	r.GET("/notification/page", a.notificationController.GetList)
	r.PUT("/notification/read/state/all", a.notificationController.ClearUnRead)
	r.PUT("/notification/read/state", a.notificationController.ClearIDUnRead)

	// upload file
	r.POST("/file", a.uploadController.UploadFile)
	r.POST("/post/render", a.uploadController.PostRender)

	// activity
	r.GET("/activity/timeline", a.activityController.GetObjectTimeline)
	r.GET("/activity/timeline/detail", a.activityController.GetObjectTimelineDetail)

}

func (a *AnswerAPIRouter) RegisterAnswerAdminAPIRouter(r *gin.RouterGroup) {
	r.GET("/question/page", a.questionController.AdminSearchList)
	r.PUT("/question/status", a.questionController.AdminSetQuestionStatus)
	r.GET("/answer/page", a.questionController.AdminSearchAnswerList)
	r.PUT("/answer/status", a.answerController.AdminSetAnswerStatus)

	// report
	r.GET("/reports/page", a.adminReportController.ListReportPage)
	r.PUT("/report", a.adminReportController.Handle)

	// user
	r.GET("/users/page", a.adminUserController.GetUserPage)
	r.PUT("/user/status", a.adminUserController.UpdateUserStatus)
	r.PUT("/user/role", a.adminUserController.UpdateUserRole)
	r.POST("/user", a.adminUserController.AddUser)
	r.PUT("/user/password", a.adminUserController.UpdateUserPassword)

	// reason
	r.GET("/reasons", a.reasonController.Reasons)

	// language
	r.GET("/language/options", a.langController.GetAdminLangOptions)

	// theme
	r.GET("/theme/options", a.themeController.GetThemeOptions)

	// siteinfo
	r.GET("/siteinfo/general", a.siteInfoController.GetGeneral)
	r.GET("/siteinfo/interface", a.siteInfoController.GetInterface)
	r.GET("/siteinfo/branding", a.siteInfoController.GetSiteBranding)
	r.GET("/siteinfo/write", a.siteInfoController.GetSiteWrite)
	r.GET("/siteinfo/legal", a.siteInfoController.GetSiteLegal)
	r.GET("/siteinfo/seo", a.siteInfoController.GetSeo)
	r.GET("/siteinfo/login", a.siteInfoController.GetSiteLogin)
	r.GET("/siteinfo/custom-css-html", a.siteInfoController.GetSiteCustomCssHTML)
	r.GET("/siteinfo/theme", a.siteInfoController.GetSiteTheme)
	r.PUT("/siteinfo/general", a.siteInfoController.UpdateGeneral)
	r.PUT("/siteinfo/interface", a.siteInfoController.UpdateInterface)
	r.PUT("/siteinfo/branding", a.siteInfoController.UpdateBranding)
	r.PUT("/siteinfo/write", a.siteInfoController.UpdateSiteWrite)
	r.PUT("/siteinfo/legal", a.siteInfoController.UpdateSiteLegal)
	r.PUT("/siteinfo/login", a.siteInfoController.UpdateSiteLogin)
	r.PUT("/siteinfo/custom-css-html", a.siteInfoController.UpdateSiteCustomCssHTML)
	r.PUT("/siteinfo/theme", a.siteInfoController.SaveSiteTheme)
	r.PUT("/siteinfo/seo", a.siteInfoController.UpdateSeo)
	r.GET("/setting/smtp", a.siteInfoController.GetSMTPConfig)
	r.PUT("/setting/smtp", a.siteInfoController.UpdateSMTPConfig)

	// dashboard
	r.GET("/dashboard", a.dashboardController.DashboardInfo)

	// roles
	r.GET("/roles", a.roleController.GetRoleList)
}
