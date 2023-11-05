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

package templaterender

import (
	"context"
	"github.com/apache/incubator-answer/internal/base/pager"
	"github.com/apache/incubator-answer/internal/schema"
)

func (t *TemplateRenderController) CommentList(
	ctx context.Context,
	objectIDs []string,
) (
	comments map[string][]*schema.GetCommentResp,
	err error,
) {

	comments = make(map[string][]*schema.GetCommentResp, len(objectIDs))

	for _, objectID := range objectIDs {
		var (
			req = &schema.GetCommentWithPageReq{
				Page:      1,
				PageSize:  3,
				ObjectID:  objectID,
				QueryCond: "vote",
				UserID:    "",
			}
			pageModel *pager.PageModel
		)
		pageModel, err = t.commentService.GetCommentWithPage(ctx, req)
		if err != nil {
			return
		}
		li := pageModel.List
		comments[objectID] = li.([]*schema.GetCommentResp)
	}
	return
}
