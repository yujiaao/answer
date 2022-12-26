package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/answerdev/answer/internal/base/constant"
	"github.com/answerdev/answer/internal/base/handler"
	templaterender "github.com/answerdev/answer/internal/controller/template_render"
	"github.com/answerdev/answer/internal/schema"
	"github.com/answerdev/answer/internal/service/siteinfo_common"
	"github.com/answerdev/answer/pkg/converter"
	"github.com/answerdev/answer/pkg/htmltext"
	"github.com/answerdev/answer/pkg/obj"
	"github.com/answerdev/answer/ui"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

type TemplateController struct {
	scriptPath               string
	cssPath                  string
	templateRenderController *templaterender.TemplateRenderController
	siteInfoService          *siteinfo_common.SiteInfoCommonService
}

// NewTemplateController new controller
func NewTemplateController(
	templateRenderController *templaterender.TemplateRenderController,
	siteInfoService *siteinfo_common.SiteInfoCommonService,
) *TemplateController {
	script, css := GetStyle()
	return &TemplateController{
		scriptPath:               script,
		cssPath:                  css,
		templateRenderController: templateRenderController,
		siteInfoService:          siteInfoService,
	}
}
func GetStyle() (script, css string) {
	file, err := ui.Build.ReadFile("build/index.html")
	if err != nil {
		return
	}
	scriptRegexp := regexp.MustCompile(`<script defer="defer" src="(.*)"></script>`)
	scriptData := scriptRegexp.FindStringSubmatch(string(file))
	cssRegexp := regexp.MustCompile(`<link href="(.*)" rel="stylesheet">`)
	cssListData := cssRegexp.FindStringSubmatch(string(file))
	if len(scriptData) == 2 {
		script = scriptData[1]
	}
	if len(cssListData) == 2 {
		css = cssListData[1]
	}
	return
}
func (tc *TemplateController) SiteInfo(ctx *gin.Context) *schema.TemplateSiteInfoResp {
	var err error
	resp := &schema.TemplateSiteInfoResp{}
	resp.General, err = tc.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.Interface, err = tc.siteInfoService.GetSiteInterface(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.Branding, err = tc.siteInfoService.GetSiteBranding(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.SiteSeo, err = tc.siteInfoService.GetSiteSeo(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.CustomCssHtml, err = tc.siteInfoService.GetSiteCustomCssHTML(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.Year = fmt.Sprintf("%d", time.Now().Year())
	return resp
}

// Index question list
func (tc *TemplateController) Index(ctx *gin.Context) {
	req := &schema.QuestionSearch{
		Order: "newest",
	}
	if handler.BindAndCheck(ctx, req) {
		tc.Page404(ctx)
		return
	}

	var page = req.Page

	data, count, err := tc.templateRenderController.Index(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}

	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = siteInfo.General.SiteUrl

	UrlUseTitle := false
	if siteInfo.SiteSeo.PermaLink == schema.PermaLinkQuestionIDAndTitle {
		UrlUseTitle = true
	}
	siteInfo.Title = ""
	tc.html(ctx, http.StatusOK, "question.html", siteInfo, gin.H{
		"data":     data,
		"useTitle": UrlUseTitle,
		"page":     templaterender.Paginator(page, req.PageSize, count),
		"path":     "questions",
	})
}

func (tc *TemplateController) QuestionList(ctx *gin.Context) {
	req := &schema.QuestionSearch{
		Order: "newest",
	}
	if handler.BindAndCheck(ctx, req) {
		tc.Page404(ctx)
		return
	}
	var page = req.Page
	data, count, err := tc.templateRenderController.Index(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = fmt.Sprintf("%s/questions", siteInfo.General.SiteUrl)

	UrlUseTitle := false
	if siteInfo.SiteSeo.PermaLink == schema.PermaLinkQuestionIDAndTitle {
		UrlUseTitle = true
	}
	siteInfo.Title = fmt.Sprintf("Questions - %s", siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "question.html", siteInfo, gin.H{
		"data":     data,
		"useTitle": UrlUseTitle,
		"page":     templaterender.Paginator(page, req.PageSize, count),
	})
}

func (tc *TemplateController) QuestionInfo301Jump(ctx *gin.Context, siteInfo *schema.TemplateSiteInfoResp, correctTitle bool) (jump bool, url string) {
	id := ctx.Param("id")
	title := ctx.Param("title")
	titleIsAnswerID := false

	objectType, objectTypeerr := obj.GetObjectTypeStrByObjectID(title)
	if objectTypeerr == nil {
		if objectType == constant.AnswerObjectType {
			titleIsAnswerID = true
		}
	}

	url = fmt.Sprintf("%s/questions/%s", siteInfo.General.SiteUrl, id)
	if siteInfo.SiteSeo.PermaLink == schema.PermaLinkQuestionID {
		//not have title
		if titleIsAnswerID || len(title) == 0 {
			return false, ""
		}
		return true, url
	} else {
		//have title
		if len(title) > 0 && !titleIsAnswerID && correctTitle {
			return false, ""
		}
		detail, err := tc.templateRenderController.QuestionDetail(ctx, id)
		if err != nil {
			tc.Page404(ctx)
			return
		}
		url = fmt.Sprintf("%s/%s", url, htmltext.UrlTitle(detail.Title))
		return true, url
	}
}

// QuestionInfo question and answers info
func (tc *TemplateController) QuestionInfo(ctx *gin.Context) {
	id := ctx.Param("id")
	title := ctx.Param("title")
	answerid := ctx.Param("answerid")

	if id == "ask" {
		file, err := ui.Build.ReadFile("build/index.html")
		if err != nil {
			log.Error(err)
			tc.Page404(ctx)
			return
		}
		ctx.Header("content-type", "text/html;charset=utf-8")
		ctx.String(http.StatusOK, string(file))
		return
	}

	correctTitle := false

	detail, err := tc.templateRenderController.QuestionDetail(ctx, id)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	encodeTitle := htmltext.UrlTitle(detail.Title)
	if encodeTitle == title {
		correctTitle = true
	}

	siteInfo := tc.SiteInfo(ctx)
	jump, jumpurl := tc.QuestionInfo301Jump(ctx, siteInfo, correctTitle)
	if jump {
		ctx.Redirect(http.StatusMovedPermanently, jumpurl)
		return
	}

	// answers
	answerReq := &schema.AnswerListReq{
		QuestionID: id,
		Order:      "",
		Page:       1,
		PageSize:   999,
		UserID:     "",
	}
	answers, answerCount, err := tc.templateRenderController.AnswerList(ctx, answerReq)
	if err != nil {
		tc.Page404(ctx)
		return
	}

	// comments
	objectIDs := []string{id}
	for _, answer := range answers {
		objectIDs = append(objectIDs, answer.ID)
	}
	comments, err := tc.templateRenderController.CommentList(ctx, objectIDs)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	siteInfo.Canonical = fmt.Sprintf("%s/questions/%s/%s", siteInfo.General.SiteUrl, id, encodeTitle)
	if siteInfo.SiteSeo.PermaLink == schema.PermaLinkQuestionID {
		siteInfo.Canonical = fmt.Sprintf("%s/questions/%s", siteInfo.General.SiteUrl, id)
	}
	jsonLD := &schema.QAPageJsonLD{}
	jsonLD.Context = "https://schema.org"
	jsonLD.Type = "QAPage"
	jsonLD.MainEntity.Type = "Question"
	jsonLD.MainEntity.Name = detail.Title
	jsonLD.MainEntity.Text = detail.HTML
	jsonLD.MainEntity.AnswerCount = int(answerCount)
	jsonLD.MainEntity.UpvoteCount = detail.VoteCount
	jsonLD.MainEntity.DateCreated = time.Unix(detail.CreateTime, 0)
	jsonLD.MainEntity.Author.Type = "Person"
	jsonLD.MainEntity.Author.Name = detail.UserInfo.DisplayName
	answerList := make([]*schema.SuggestedAnswerItem, 0)
	for _, answer := range answers {
		if answer.Accepted == schema.AnswerAcceptedEnable {
			acceptedAnswerItem := &schema.AcceptedAnswerItem{}
			acceptedAnswerItem.Type = "Answer"
			acceptedAnswerItem.Text = answer.HTML
			acceptedAnswerItem.DateCreated = time.Unix(answer.CreateTime, 0)
			acceptedAnswerItem.UpvoteCount = answer.VoteCount
			acceptedAnswerItem.URL = fmt.Sprintf("%s/%s", siteInfo.Canonical, answer.ID)
			acceptedAnswerItem.Author.Type = "Person"
			acceptedAnswerItem.Author.Name = answer.UserInfo.DisplayName
			jsonLD.MainEntity.AcceptedAnswer = acceptedAnswerItem
		} else {
			item := &schema.SuggestedAnswerItem{}
			item.Type = "Answer"
			item.Text = answer.HTML
			item.DateCreated = time.Unix(answer.CreateTime, 0)
			item.UpvoteCount = answer.VoteCount
			item.URL = fmt.Sprintf("%s/%s", siteInfo.Canonical, answer.ID)
			item.Author.Type = "Person"
			item.Author.Name = answer.UserInfo.DisplayName
			answerList = append(answerList, item)
		}

	}
	jsonLD.MainEntity.SuggestedAnswer = answerList
	jsonLDStr, err := json.Marshal(jsonLD)
	if err == nil {
		siteInfo.JsonLD = `<script data-react-helmet="true" type="application/ld+json">` + string(jsonLDStr) + ` </script>`
	}

	siteInfo.Description = htmltext.FetchExcerpt(detail.HTML, "...", 240)
	tags := make([]string, 0)
	for _, tag := range detail.Tags {
		tags = append(tags, tag.DisplayName)
	}
	siteInfo.Keywords = strings.Replace(strings.Trim(fmt.Sprint(tags), "[]"), " ", ",", -1)
	siteInfo.Title = fmt.Sprintf("%s - %s", detail.Title, siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "question-detail.html", siteInfo, gin.H{
		"id":       id,
		"answerid": answerid,
		"detail":   detail,
		"answers":  answers,
		"comments": comments,
	})
}

// TagList tags list
func (tc *TemplateController) TagList(ctx *gin.Context) {
	req := &schema.GetTagWithPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	data, err := tc.templateRenderController.TagList(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	page := templaterender.Paginator(req.Page, req.PageSize, data.Count)

	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = fmt.Sprintf("%s/tags", siteInfo.General.SiteUrl)
	siteInfo.Title = fmt.Sprintf("%s - %s", "Tags", siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "tags.html", siteInfo, gin.H{
		"page": page,
		"data": data,
	})
}

// TagInfo taginfo
func (tc *TemplateController) TagInfo(ctx *gin.Context) {
	tag := ctx.Param("tag")
	req := &schema.GetTamplateTagInfoReq{}
	if handler.BindAndCheck(ctx, req) {
		tc.Page404(ctx)
		return
	}
	nowPage := req.Page
	req.Name = tag
	taginifo, questionList, questionCount, err := tc.templateRenderController.TagInfo(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	page := templaterender.Paginator(nowPage, req.PageSize, questionCount)

	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = fmt.Sprintf("%s/tags/%s", siteInfo.General.SiteUrl, tag)
	siteInfo.Description = htmltext.FetchExcerpt(taginifo.ParsedText, "...", 240)
	if len(taginifo.ParsedText) == 0 {
		siteInfo.Description = "The tag has no description."
	}
	siteInfo.Keywords = taginifo.DisplayName

	UrlUseTitle := false
	if siteInfo.SiteSeo.PermaLink == schema.PermaLinkQuestionIDAndTitle {
		UrlUseTitle = true
	}
	siteInfo.Title = fmt.Sprintf("'%s' Questions - %s", taginifo.DisplayName, siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "tag-detail.html", siteInfo, gin.H{
		"tag":           taginifo,
		"questionList":  questionList,
		"questionCount": questionCount,
		"useTitle":      UrlUseTitle,
		"page":          page,
	})
}

// UserInfo user info
func (tc *TemplateController) UserInfo(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		tc.Page404(ctx)
		return
	}
	exist := constant.ExistInPathIgnore(username)
	if exist {
		file, err := ui.Build.ReadFile("build/index.html")
		if err != nil {
			log.Error(err)
			tc.Page404(ctx)
			return
		}
		ctx.Header("content-type", "text/html;charset=utf-8")
		ctx.String(http.StatusOK, string(file))
		return
	}
	req := &schema.GetOtherUserInfoByUsernameReq{}
	req.Username = username
	userinfo, err := tc.templateRenderController.UserInfo(ctx, req)

	if err != nil {
		tc.Page404(ctx)
		return
	}
	if !userinfo.Has {
		tc.Page404(ctx)
		return
	}

	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = fmt.Sprintf("%s/users/%s", siteInfo.General.SiteUrl, username)
	siteInfo.Title = fmt.Sprintf("%s - %s", username, siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "homepage.html", siteInfo, gin.H{
		"userinfo": userinfo,
		"bio":      template.HTML(userinfo.Info.BioHTML),
	})

}

func (tc *TemplateController) Page404(ctx *gin.Context) {
	tc.html(ctx, http.StatusNotFound, "404.html", tc.SiteInfo(ctx), gin.H{})
}

func (tc *TemplateController) html(ctx *gin.Context, code int, tpl string, siteInfo *schema.TemplateSiteInfoResp, data gin.H) {
	data["siteinfo"] = siteInfo
	data["scriptPath"] = tc.scriptPath
	data["cssPath"] = tc.cssPath
	data["keywords"] = siteInfo.Keywords
	if siteInfo.Description == "" {
		siteInfo.Description = siteInfo.General.Description
	}
	data["title"] = siteInfo.Title
	if siteInfo.Title == "" {
		data["title"] = siteInfo.General.Name
	}
	data["description"] = siteInfo.Description
	data["language"] = handler.GetLang(ctx)
	data["timezone"] = siteInfo.Interface.TimeZone
	data["HeadCode"] = siteInfo.CustomCssHtml.CustomHead
	data["HeaderCode"] = siteInfo.CustomCssHtml.CustomHeader
	data["FooterCode"] = siteInfo.CustomCssHtml.CustomFooter
	_, ok := data["path"]
	if !ok {
		data["path"] = ""
	}
	ctx.HTML(code, tpl, data)
}

func (tc *TemplateController) Sitemap(ctx *gin.Context) {
	if tc.checkPrivateMode(ctx) {
		tc.Page404(ctx)
		return
	}
	tc.templateRenderController.Sitemap(ctx)
}

func (tc *TemplateController) SitemapPage(ctx *gin.Context) {
	if tc.checkPrivateMode(ctx) {
		tc.Page404(ctx)
		return
	}
	page := 0
	pageParam := ctx.Param("page")
	pageRegexp := regexp.MustCompile(`question-(.*).xml`)
	pageStr := pageRegexp.FindStringSubmatch(pageParam)
	if len(pageStr) != 2 {
		tc.Page404(ctx)
		return
	}
	page = converter.StringToInt(pageStr[1])
	if page == 0 {
		tc.Page404(ctx)
		return
	}
	err := tc.templateRenderController.SitemapPage(ctx, page)
	if err != nil {
		tc.Page404(ctx)
		return
	}
}

func (tc *TemplateController) checkPrivateMode(ctx *gin.Context) bool {
	resp, err := tc.siteInfoService.GetSiteLogin(ctx)
	if err != nil {
		log.Error(err)
		return false
	}
	if resp.LoginRequired {
		return true
	}
	return false
}
