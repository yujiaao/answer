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

package action

import (
	"context"
	"image/color"
	"strings"

	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/mojocn/base64Captcha"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// CaptchaRepo captcha repository
type CaptchaRepo interface {
	SetCaptcha(ctx context.Context, key, captcha string) (err error)
	GetCaptcha(ctx context.Context, key string) (captcha string, err error)
	DelCaptcha(ctx context.Context, key string) (err error)
	SetActionType(ctx context.Context, unit, actionType, config string, amount int) (err error)
	GetActionType(ctx context.Context, unit, actionType string) (actioninfo *entity.ActionRecordInfo, err error)
	DelActionType(ctx context.Context, unit, actionType string) (err error)
}

// CaptchaService kit service
type CaptchaService struct {
	captchaRepo CaptchaRepo
}

// NewCaptchaService captcha service
func NewCaptchaService(captchaRepo CaptchaRepo) *CaptchaService {
	return &CaptchaService{
		captchaRepo: captchaRepo,
	}
}

// ActionRecord action record
func (cs *CaptchaService) ActionRecord(ctx context.Context, req *schema.ActionRecordReq) (resp *schema.ActionRecordResp, err error) {
	resp = &schema.ActionRecordResp{}
	unit := req.IP
	switch req.Action {
	case entity.CaptchaActionEditUserinfo:
		unit = req.UserID
	case entity.CaptchaActionQuestion:
		unit = req.UserID
	case entity.CaptchaActionAnswer:
		unit = req.UserID
	case entity.CaptchaActionComment:
		unit = req.UserID
	case entity.CaptchaActionEdit:
		unit = req.UserID
	case entity.CaptchaActionInvitationAnswer:
		unit = req.UserID
	case entity.CaptchaActionSearch:
		if req.UserID != "" {
			unit = req.UserID
		}
	case entity.CaptchaActionReport:
		unit = req.UserID
	case entity.CaptchaActionDelete:
		unit = req.UserID
	case entity.CaptchaActionVote:
		unit = req.UserID
	}
	verificationResult := cs.ValidationStrategy(ctx, unit, req.Action)
	if !verificationResult {
		resp.CaptchaID, resp.CaptchaImg, err = cs.GenerateCaptcha(ctx)
		resp.Verify = true
	}
	return
}

func (cs *CaptchaService) UserRegisterCaptcha(ctx context.Context) (resp *schema.ActionRecordResp, err error) {
	resp = &schema.ActionRecordResp{}
	resp.CaptchaID, resp.CaptchaImg, err = cs.GenerateCaptcha(ctx)
	resp.Verify = true
	return
}

func (cs *CaptchaService) UserRegisterVerifyCaptcha(
	ctx context.Context, id string, VerifyValue string,
) bool {
	if id == "" || VerifyValue == "" {
		return false
	}
	pass, err := cs.VerifyCaptcha(ctx, id, VerifyValue)
	if err != nil {
		return false
	}
	return pass
}

// ActionRecordVerifyCaptcha
// Verify that you need to enter a CAPTCHA, and that the CAPTCHA is correct
func (cs *CaptchaService) ActionRecordVerifyCaptcha(
	ctx context.Context, actionType string, unit string, id string, VerifyValue string,
) bool {
	verificationResult := cs.ValidationStrategy(ctx, unit, actionType)
	if !verificationResult {
		if id == "" || VerifyValue == "" {
			return false
		}
		pass, err := cs.VerifyCaptcha(ctx, id, VerifyValue)
		if err != nil {
			return false
		}
		return pass
	}
	return true
}

func (cs *CaptchaService) ActionRecordAdd(ctx context.Context, actionType string, unit string) (int, error) {
	info, err := cs.captchaRepo.GetActionType(ctx, unit, actionType)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	amount := 1
	if info != nil {
		amount = info.Num + 1
	}
	err = cs.captchaRepo.SetActionType(ctx, unit, actionType, "", amount)
	if err != nil {
		return 0, err
	}
	return amount, nil
}

func (cs *CaptchaService) ActionRecordDel(ctx context.Context, actionType string, unit string) {
	err := cs.captchaRepo.DelActionType(ctx, unit, actionType)
	if err != nil {
		log.Error(err)
	}
}

// GenerateCaptcha generate captcha
func (cs *CaptchaService) GenerateCaptcha(ctx context.Context) (key, captchaBase64 string, err error) {
	driverString := base64Captcha.DriverString{
		Height:          60,
		Width:           200,
		NoiseCount:      0,
		ShowLineOptions: 2 | 4,
		Length:          4,
		Source:          "1234567890qwertyuioplkjhgfdsazxcvbnm",
		BgColor:         &color.RGBA{R: 211, G: 211, B: 211, A: 0},
		Fonts:           []string{"wqy-microhei.ttc"},
	}
	driver := driverString.ConvertFonts()

	id, content, answer := driver.GenerateIdQuestionAnswer()
	item, err := driver.DrawCaptcha(content)
	if err != nil {
		return "", "", errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}
	err = cs.captchaRepo.SetCaptcha(ctx, id, answer)
	if err != nil {
		return "", "", err
	}

	captchaBase64 = item.EncodeB64string()
	return id, captchaBase64, nil
}

// VerifyCaptcha generate captcha
func (cs *CaptchaService) VerifyCaptcha(ctx context.Context, key, captcha string) (isCorrect bool, err error) {
	realCaptcha, err := cs.captchaRepo.GetCaptcha(ctx, key)
	if err != nil {
		log.Error("VerifyCaptcha GetCaptcha Error", err.Error())
		return false, nil
	}
	err = cs.captchaRepo.DelCaptcha(ctx, key)
	if err != nil {
		log.Error("VerifyCaptcha DelCaptcha Error", err.Error())
		return false, nil
	}
	return strings.TrimSpace(captcha) == realCaptcha, nil
}
