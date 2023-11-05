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

package repo_test

import (
	"context"
	"testing"

	"github.com/apache/incubator-answer/internal/repo/config"
	serviceconfig "github.com/apache/incubator-answer/internal/service/config"

	"github.com/apache/incubator-answer/internal/repo/reason"
	"github.com/stretchr/testify/assert"
)

func Test_reasonRepo_ListReasons(t *testing.T) {
	configRepo := config.NewConfigRepo(testDataSource)
	reasonRepo := reason.NewReasonRepo(serviceconfig.NewConfigService(configRepo))
	reasonItems, err := reasonRepo.ListReasons(context.TODO(), "question", "close")
	assert.NoError(t, err)
	assert.Equal(t, 4, len(reasonItems))
}
