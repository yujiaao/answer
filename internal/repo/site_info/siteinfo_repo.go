package site_info

import (
	"context"
	"encoding/json"

	"github.com/answerdev/answer/internal/base/constant"
	"github.com/answerdev/answer/internal/base/data"
	"github.com/answerdev/answer/internal/base/reason"
	"github.com/answerdev/answer/internal/entity"
	"github.com/answerdev/answer/internal/service/siteinfo_common"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
)

type siteInfoRepo struct {
	data *data.Data
}

func NewSiteInfo(data *data.Data) siteinfo_common.SiteInfoRepo {
	return &siteInfoRepo{
		data: data,
	}
}

// SaveByType save site setting by type
func (sr *siteInfoRepo) SaveByType(ctx context.Context, siteType string, data *entity.SiteInfo) (err error) {
	old := &entity.SiteInfo{}
	exist, err := sr.data.DB.Where(builder.Eq{"type": siteType}).Get(old)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist {
		_, err = sr.data.DB.ID(old.ID).Update(data)
	} else {
		_, err = sr.data.DB.Insert(data)
	}
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	sr.setCache(ctx, siteType, data)
	return
}

// GetByType get site info by type
func (sr *siteInfoRepo) GetByType(ctx context.Context, siteType string) (siteInfo *entity.SiteInfo, exist bool, err error) {
	siteInfo = sr.getCache(ctx, siteType)
	if siteInfo != nil {
		return siteInfo, true, nil
	}
	siteInfo = &entity.SiteInfo{}
	exist, err = sr.data.DB.Where(builder.Eq{"type": siteType}).Get(siteInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist {
		sr.setCache(ctx, siteType, siteInfo)
	}
	return
}

func (sr *siteInfoRepo) getCache(ctx context.Context, siteType string) (siteInfo *entity.SiteInfo) {
	siteInfo = &entity.SiteInfo{}
	siteInfoCache, err := sr.data.Cache.GetString(ctx, constant.SiteInfoCacheKey+siteType)
	if err != nil {
		return nil
	}
	_ = json.Unmarshal([]byte(siteInfoCache), siteInfo)
	return siteInfo
}

func (sr *siteInfoRepo) setCache(ctx context.Context, siteType string, siteInfo *entity.SiteInfo) {
	siteInfoCache, _ := json.Marshal(siteInfo)
	err := sr.data.Cache.SetString(ctx,
		constant.SiteInfoCacheKey+siteType, string(siteInfoCache), constant.SiteInfoCacheTime)
	if err != nil {
		log.Error(err)
	}
}
