package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/answerdev/answer/internal/base/reason"
	"github.com/answerdev/answer/internal/base/translator"
	"github.com/go-playground/locales"
	german "github.com/go-playground/locales/de"
	english "github.com/go-playground/locales/en"
	spanish "github.com/go-playground/locales/es"
	french "github.com/go-playground/locales/fr"
	italian "github.com/go-playground/locales/it"
	japanese "github.com/go-playground/locales/ja"
	korean "github.com/go-playground/locales/ko"
	portuguese "github.com/go-playground/locales/pt"
	russian "github.com/go-playground/locales/ru"
	vietnamese "github.com/go-playground/locales/vi"
	chinese "github.com/go-playground/locales/zh"
	chineseTraditional "github.com/go-playground/locales/zh_Hant_TW"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/translations/en"
	"github.com/go-playground/validator/v10/translations/es"
	"github.com/go-playground/validator/v10/translations/fr"
	"github.com/go-playground/validator/v10/translations/it"
	"github.com/go-playground/validator/v10/translations/ja"
	"github.com/go-playground/validator/v10/translations/pt"
	"github.com/go-playground/validator/v10/translations/ru"
	"github.com/go-playground/validator/v10/translations/vi"
	"github.com/go-playground/validator/v10/translations/zh"
	"github.com/go-playground/validator/v10/translations/zh_tw"
	myErrors "github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/i18n"
	"github.com/segmentfault/pacman/log"
)

type TranslatorLocal struct {
	La           i18n.Language
	Lo           locales.Translator
	RegisterFunc func(v *validator.Validate, trans ut.Translator) (err error)
}

var (
	allLanguageTranslators = []*TranslatorLocal{
		{La: i18n.LanguageChinese, Lo: chinese.New(), RegisterFunc: zh.RegisterDefaultTranslations},
		{La: i18n.LanguageChineseTraditional, Lo: chineseTraditional.New(), RegisterFunc: zh_tw.RegisterDefaultTranslations},
		{La: i18n.LanguageEnglish, Lo: english.New(), RegisterFunc: en.RegisterDefaultTranslations},
		{La: i18n.LanguageGerman, Lo: german.New(), RegisterFunc: nil},
		{La: i18n.LanguageSpanish, Lo: spanish.New(), RegisterFunc: es.RegisterDefaultTranslations},
		{La: i18n.LanguageFrench, Lo: french.New(), RegisterFunc: fr.RegisterDefaultTranslations},
		{La: i18n.LanguageItalian, Lo: italian.New(), RegisterFunc: it.RegisterDefaultTranslations},
		{La: i18n.LanguageJapanese, Lo: japanese.New(), RegisterFunc: ja.RegisterDefaultTranslations},
		{La: i18n.LanguageKorean, Lo: korean.New(), RegisterFunc: nil},
		{La: i18n.LanguagePortuguese, Lo: portuguese.New(), RegisterFunc: pt.RegisterDefaultTranslations},
		{La: i18n.LanguageRussian, Lo: russian.New(), RegisterFunc: ru.RegisterDefaultTranslations},
		{La: i18n.LanguageVietnamese, Lo: vietnamese.New(), RegisterFunc: vi.RegisterDefaultTranslations},
	}
)

// MyValidator my validator
type MyValidator struct {
	Validate *validator.Validate
	Tran     ut.Translator
	Lang     i18n.Language
}

// FormErrorField indicates the current form error content. which field is error and error message.
type FormErrorField struct {
	ErrorField string `json:"error_field"`
	ErrorMsg   string `json:"error_msg"`
}

// GlobalValidatorMapping is a mapping from validator to translator used
var GlobalValidatorMapping = make(map[i18n.Language]*MyValidator, 0)

func init() {
	for _, t := range allLanguageTranslators {
		tran, val := getTran(t.Lo), createDefaultValidator(t.La)
		if t.RegisterFunc != nil {
			if err := t.RegisterFunc(val, tran); err != nil {
				panic(err)
			}
		}
		GlobalValidatorMapping[t.La] = &MyValidator{Validate: val, Tran: tran, Lang: t.La}
	}
}

func getTran(lo locales.Translator) ut.Translator {
	tran, ok := ut.New(lo, lo).GetTranslator(lo.Locale())
	if !ok {
		panic(fmt.Sprintf("not found translator %s", lo.Locale()))
	}
	return tran
}

func createDefaultValidator(la i18n.Language) *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) (res string) {
		defer func() {
			if len(res) > 0 {
				res = translator.GlobalTrans.Tr(la, res)
			}
		}()
		if jsonTag := fld.Tag.Get("json"); len(jsonTag) > 0 {
			if jsonTag == "-" {
				return ""
			}
			return jsonTag
		}
		if formTag := fld.Tag.Get("form"); len(formTag) > 0 {
			return formTag
		}
		return fld.Name
	})
	return validate
}

func GetValidatorByLang(lang i18n.Language) *MyValidator {
	if GlobalValidatorMapping[lang] != nil {
		return GlobalValidatorMapping[lang]
	}
	return GlobalValidatorMapping[i18n.DefaultLanguage]
}

// Check /
func (m *MyValidator) Check(value interface{}) (errFields []*FormErrorField, err error) {
	err = m.Validate.Struct(value)
	if err != nil {
		var valErrors validator.ValidationErrors
		if !errors.As(err, &valErrors) {
			log.Error(err)
			return nil, errors.New("validate check exception")
		}

		for _, fieldError := range valErrors {
			errField := &FormErrorField{
				ErrorField: fieldError.Field(),
				ErrorMsg:   fieldError.Translate(m.Tran),
			}

			// get original tag name from value for set err field key.
			structNamespace := fieldError.StructNamespace()
			_, fieldName, found := strings.Cut(structNamespace, ".")
			if found {
				originalTag := getObjectTagByFieldName(value, fieldName)
				if len(originalTag) > 0 {
					errField.ErrorField = originalTag
				}
			}
			errFields = append(errFields, errField)
		}
		if len(errFields) > 0 {
			errMsg := ""
			if len(errFields) == 1 {
				errMsg = errFields[0].ErrorMsg
			}
			return errFields, myErrors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
		}
	}

	if v, ok := value.(Checker); ok {
		errFields, err = v.Check()
		if err == nil {
			return nil, nil
		}
		for _, errField := range errFields {
			errField.ErrorMsg = translator.GlobalTrans.Tr(m.Lang, errField.ErrorMsg)
		}
		return errFields, err
	}
	return nil, nil
}

// Checker .
type Checker interface {
	Check() (errField []*FormErrorField, err error)
}

func getObjectTagByFieldName(obj interface{}, fieldName string) (tag string) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	objT := reflect.TypeOf(obj)
	objT = objT.Elem()

	structField, exists := objT.FieldByName(fieldName)
	if !exists {
		return ""
	}
	tag = structField.Tag.Get("json")
	if len(tag) == 0 {
		return structField.Tag.Get("form")
	}
	return tag
}
