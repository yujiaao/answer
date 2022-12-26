package migrations

import (
	"xorm.io/xorm"
)

func addTagRecommendedAndReserved(x *xorm.Engine) error {
	type Tag struct {
		ID        string `xorm:"not null pk comment('tag_id') BIGINT(20) id"`
		SlugName  string `xorm:"not null default '' unique VARCHAR(35) slug_name"`
		Recommend bool   `xorm:"not null default false BOOL recommend"`
		Reserved  bool   `xorm:"not null default false BOOL reserved"`
	}
	return x.Sync(new(Tag))
}
