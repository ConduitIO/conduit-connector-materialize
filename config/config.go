// Copyright © 2022 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"errors"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.uber.org/multierr"
)

const (
	// ConfigKeyURL is the config name for a connection URL.
	ConfigKeyURL = "url"
	// ConfigKeyTable is the config name for a table.
	ConfigKeyTable = "table"
	// ConfigKeyKey is the config name for a key.
	ConfigKeyKey = "key"
)

// Config represents configuration needed for Materialize.
type Config struct {
	URL string `validate:"required,url"`
	// The maximum identifier length is 63.
	// See https://www.postgresql.org/docs/current/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS.
	Table string `validate:"max=63"`
	Key   string `validate:"max=63"`
}

// Validate validates the Config.
func (c Config) Validate() error {
	// init a translator and a universal translator
	translator := en.New()
	uni := ut.New(translator, translator)

	uniTranslator, found := uni.GetTranslator("en")
	if !found {
		return errors.New("translator not found")
	}

	// init a new instance of a validator
	validate := validator.New()

	// register custom translations
	if err := registerTranslations(validate, uniTranslator); err != nil {
		return err
	}

	// collect all validation errors into one
	if err := validate.Struct(c); err != nil {
		var resultErr error
		validationErrors := err.(validator.ValidationErrors)
		for _, validationError := range validationErrors {
			resultErr = multierr.Append(resultErr, errors.New(
				validationError.Translate(uniTranslator),
			))
		}

		return resultErr
	}

	return nil
}

// registerTranslations registers custom tranlations (error messages) for validation errors.
func registerTranslations(validate *validator.Validate, uniTranslator ut.Translator) error {
	// register a custom translation for the required tag
	err := validate.RegisterTranslation("required", uniTranslator, func(ut ut.Translator) error {
		return ut.Add("required", "\"{0}\" config value must be set", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return strings.ToLower(t)
	})
	if err != nil {
		return err
	}

	// register a custom translation for the url tag
	err = validate.RegisterTranslation("url", uniTranslator, func(ut ut.Translator) error {
		return ut.Add("url", "\"{0}\" config value must be a valid url", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("url", fe.Field())
		return strings.ToLower(t)
	})
	if err != nil {
		return err
	}

	// register a custom translation for the max tag
	err = validate.RegisterTranslation("max", uniTranslator, func(ut ut.Translator) error {
		return ut.Add("max", "\"{0}\" config value is too long", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("max", fe.Field())
		return strings.ToLower(t)
	})
	if err != nil {
		return err
	}

	return nil
}

// Parse attempts to parse plugins.Config into a Config struct
func Parse(cfg map[string]string) (Config, error) {
	config := Config{
		URL:   cfg[ConfigKeyURL],
		Table: cfg[ConfigKeyTable],
		Key:   cfg[ConfigKeyKey],
	}

	if err := config.Validate(); err != nil {
		return Config{}, err
	}

	return config, nil
}
