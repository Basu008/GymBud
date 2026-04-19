package validator

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Validator wraps go-playground validator
// Validator container validator library and transaltor
type Validator struct {
	V *validator.Validate
	T *ut.Translator
}

// NewValidation create new Validator struct instance
func NewValidator() *Validator {
	v := &Validator{
		V: validator.New(),
	}
	trans := initializeTranslation(v.V)
	v.T = trans
	registerFunc(v.V)
	return v
}

// Initialize initializes and returns the UniversalTranslator instance for the application
func initializeTranslation(validate *validator.Validate) *ut.Translator {

	// initialize translator
	en := en.New()
	uni := ut.New(en, en)

	trans, _ := uni.GetTranslator("en")
	// initialize translations
	en_translations.RegisterDefaultTranslations(validate, trans)
	return &trans
}

func registerFunc(validate *validator.Validate) {
	// register function to get tag name from json tags.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Validate validates the struct
// Note: do not pass slice of struct
func (v *Validator) Validate(form interface{}) []error {
	var errResp []error
	if err := v.V.Struct(form); err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			key := strings.SplitN(e.Namespace(), ".", 2)
			field := e.Field()
			if len(key) > 1 {
				field = key[1]
			}
			errResp = append(errResp, errors.New(field+": "+e.Translate(*v.T)))
		}
	}
	return errResp
}
