// Package validate contains the support for validating models.
package validate

import (
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// validate holds the settings and caches for validating request struct values.
var validate *validator.Validate

// translator is a cache of locale and translation information.
var translator ut.Translator

func init() {
	// Instantiate a validator.
	validate = validator.New()

	// Create a translator for english so the error messages are
	// more human readable than technical.
	translator, _ = ut.New(en.New(), en.New()).GetTranslator("en")

	// Register the english error messages for use.
	if err := en_translations.RegisterDefaultTranslations(validate, translator); err != nil {
		log.Fatalf("register translations: %v", err)
	}

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})
}

func Check(val any) error {
	if err := validate.Struct(val); err != nil {
		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		var fields FieldErrors
		for _, verror := range verrors {
			fields = append(fields, FieldError{
				Field: verror.Field(),
				Err:   verror.Translate(translator),
			})
		}

		return fields
	}

	return nil
}
