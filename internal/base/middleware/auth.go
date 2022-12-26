package middleware

import (
	"strings"

	"github.com/answerdev/answer/internal/schema"
	"github.com/answerdev/answer/internal/service/siteinfo_common"

	"github.com/answerdev/answer/internal/base/handler"
	"github.com/answerdev/answer/internal/base/reason"
	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/service/auth"
	"github.com/answerdev/answer/pkg/converter"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

var ctxUUIDKey = "ctxUuidKey"

// AuthUserMiddleware auth user middleware
type AuthUserMiddleware struct {
	authService           *auth.AuthService
	siteInfoCommonService *siteinfo_common.SiteInfoCommonService
}

// NewAuthUserMiddleware new auth user middleware
func NewAuthUserMiddleware(
	authService *auth.AuthService,
	siteInfoCommonService *siteinfo_common.SiteInfoCommonService) *AuthUserMiddleware {
	return &AuthUserMiddleware{
		authService:           authService,
		siteInfoCommonService: siteInfoCommonService,
	}
}

// Auth get token and auth user, set user info to context if user is already login
func (am *AuthUserMiddleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if len(token) == 0 {
			ctx.Next()
			return
		}
		userInfo, err := am.authService.GetUserCacheInfo(ctx, token)
		if err != nil {
			ctx.Next()
			return
		}
		if userInfo != nil {
			ctx.Set(ctxUUIDKey, userInfo)
		}
		ctx.Next()
	}
}

// EjectUserBySiteInfo if admin config the site can access by nologin user, eject user.
func (am *AuthUserMiddleware) EjectUserBySiteInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mustLogin := false
		siteInfo, _ := am.siteInfoCommonService.GetSiteLogin(ctx)
		if siteInfo != nil {
			mustLogin = siteInfo.LoginRequired
		}
		if !mustLogin {
			ctx.Next()
			return
		}

		_, isLogin := ctx.Get(ctxUUIDKey)
		if !isLogin {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// MustAuth auth user info. If the user does not log in, an unauthenticated error is displayed
func (am *AuthUserMiddleware) MustAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if len(token) == 0 {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		userInfo, err := am.authService.GetUserCacheInfo(ctx, token)
		if err != nil || userInfo == nil {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		if userInfo.EmailStatus != entity.EmailStatusAvailable {
			handler.HandleResponse(ctx, errors.Forbidden(reason.EmailNeedToBeVerified),
				&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeInactive})
			ctx.Abort()
			return
		}
		if userInfo.UserStatus == entity.UserStatusSuspended {
			handler.HandleResponse(ctx, errors.Forbidden(reason.UserSuspended),
				&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeUserSuspended})
			ctx.Abort()
			return
		}
		if userInfo.UserStatus == entity.UserStatusDeleted {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		ctx.Set(ctxUUIDKey, userInfo)
		ctx.Next()
	}
}

func (am *AuthUserMiddleware) AdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if len(token) == 0 {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		userInfo, err := am.authService.GetAdminUserCacheInfo(ctx, token)
		if err != nil {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		if userInfo != nil {
			if userInfo.UserStatus == entity.UserStatusDeleted {
				handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
				ctx.Abort()
				return
			}
			ctx.Set(ctxUUIDKey, userInfo)
		}
		ctx.Next()
	}
}

// GetLoginUserIDFromContext get user id from context
func GetLoginUserIDFromContext(ctx *gin.Context) (userID string) {
	userInfo := GetUserInfoFromContext(ctx)
	if userInfo == nil {
		return ""
	}
	return userInfo.UserID
}

// GetIsAdminFromContext get user is admin from context
func GetIsAdminFromContext(ctx *gin.Context) (isAdmin bool) {
	userInfo := GetUserInfoFromContext(ctx)
	if userInfo == nil {
		return false
	}
	return userInfo.IsAdmin
}

// GetUserInfoFromContext get user info from context
func GetUserInfoFromContext(ctx *gin.Context) (u *entity.UserCacheInfo) {
	userInfo, exist := ctx.Get(ctxUUIDKey)
	if !exist {
		return nil
	}
	u, ok := userInfo.(*entity.UserCacheInfo)
	if !ok {
		return nil
	}
	return u
}

func GetLoginUserIDInt64FromContext(ctx *gin.Context) (userID int64) {
	userIDStr := GetLoginUserIDFromContext(ctx)
	return converter.StringToInt64(userIDStr)
}

// ExtractToken extract token from context
func ExtractToken(ctx *gin.Context) (token string) {
	token = ctx.GetHeader("Authorization")
	if len(token) == 0 {
		token = ctx.Query("Authorization")
	}
	return strings.TrimPrefix(token, "Bearer ")
}
