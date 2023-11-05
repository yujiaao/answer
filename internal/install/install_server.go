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

package install

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/apache/incubator-answer/ui"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

const UIStaticPath = "build/static"

type _resource struct {
	fs embed.FS
}

// Open to implement the interface by http.FS required
func (r *_resource) Open(name string) (fs.File, error) {
	name = fmt.Sprintf(UIStaticPath+"/%s", name)
	log.Debugf("open static path %s", name)
	return r.fs.Open(name)
}

// NewInstallHTTPServer new install http server.
func NewInstallHTTPServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/healthz", func(ctx *gin.Context) { ctx.String(200, "OK") })
	r.StaticFS("/static", http.FS(&_resource{
		fs: ui.Build,
	}))

	installApi := r.Group("")
	installApi.GET("/", CheckConfigFileAndRedirectToInstallPage)
	installApi.GET("/install", WebPage)
	installApi.GET("/50x", WebPage)
	installApi.GET("/installation/language/options", LangOptions)
	installApi.POST("/installation/db/check", CheckDatabase)
	installApi.POST("/installation/config-file/check", CheckConfigFile)
	installApi.POST("/installation/init", InitEnvironment)
	installApi.POST("/installation/base-info", InitBaseInfo)

	r.NoRoute(func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/50x")
	})
	return r
}

func WebPage(c *gin.Context) {
	filePath := ""
	var file []byte
	var err error
	filePath = "build/index.html"
	c.Header("content-type", "text/html;charset=utf-8")
	file, err = ui.Build.ReadFile(filePath)
	if err != nil {
		log.Error(err)
		c.Status(http.StatusNotFound)
		return
	}
	c.String(http.StatusOK, string(file))
}
