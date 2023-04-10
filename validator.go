package homework

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var errs []string
	for _, err := range v {
		errs = append(errs, err.Err.Error())

	}
	return strings.Join(errs, "\n")
}

func Validate(v any) error {
	valueOf := reflect.ValueOf(v)
	var validationErrors ValidationErrors
	if valueOf.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	for i := 0; i < valueOf.NumField(); i++ {
		field := valueOf.Type().Field(i)
		fieldValue := valueOf.Field(i)

		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		if field.PkgPath != "" {
			validationErrors = append(validationErrors, ValidationError{ErrValidateForUnexportedFields})
			continue
		}

		fieldError := validateField(fieldValue, validateTag)
		if fieldError != nil {
			validationErrors = append(validationErrors, ValidationError{fieldError})
		}
	}

	if len(validationErrors) == 0 {
		return nil
	}
	return validationErrors
}

func validateField(value reflect.Value, tag string) error {
	switch {
	case strings.HasPrefix(tag, "len:"):
		length, err := strconv.Atoi(tag[4:])
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		return validateLength(value, length)
	case strings.HasPrefix(tag, "in:"):
		options := strings.Split(tag[3:], ",")
		if tag[3:] == "" {
			return fmt.Errorf("empty in")
		}
		return validateIn(value, options)
	case strings.HasPrefix(tag, "min:"):
		minimum, err := strconv.Atoi(tag[4:])
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		return validateMinimum(value, minimum)
	case strings.HasPrefix(tag, "max:"):
		maximum, err := strconv.Atoi(tag[4:])
		if err != nil {
			return ErrInvalidValidatorSyntax
		}
		return validateMaximum(value, maximum)
	default:
		return ErrInvalidValidatorSyntax
	}
}

func validateLength(value reflect.Value, length int) error {
	switch value.Kind() {
	case reflect.String:
		if value.Len() != length {
			return fmt.Errorf("string length must be %d", length)
		}
	default:
		return fmt.Errorf("length validation not supported for %s", value.Kind())
	}
	return nil
}

func validateIn(value reflect.Value, options []string) error {
	switch value.Kind() {
	case reflect.String:
		for _, opt := range options {
			if value.String() == opt {
				return nil
			}
		}
		return fmt.Errorf("value must be one of %s", strings.Join(options, ", "))
	case reflect.Int:
		for _, opt := range options {
			val, err := strconv.Atoi(opt)
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			if value.Int() == int64(val) {
				return nil
			}
		}
		return fmt.Errorf("value must be one of %s", strings.Join(options, ", "))
	default:
		return fmt.Errorf("in validation not supported for %s", value.Kind())
	}
}

func validateMinimum(value reflect.Value, minimum int) error {
	switch value.Kind() {
	case reflect.String:
		if value.Len() < minimum {
			return fmt.Errorf("string length must be at least %d", minimum)
		}
	case reflect.Int:
		if value.Int() < int64(minimum) {
			return fmt.Errorf("value must be greater than or equal to %d", minimum)
		}
	default:
		return fmt.Errorf("minimum validation not supported for %s", value.Kind())
	}
	return nil
}

func validateMaximum(value reflect.Value, maximum int) error {
	switch value.Kind() {
	case reflect.String:
		if value.Len() > maximum {
			return fmt.Errorf("string length must be at most %d", maximum)
		}
	case reflect.Int:
		if value.Int() > int64(maximum) {
			return fmt.Errorf("value must be less than or equal to %d", maximum)
		}
	default:
		return fmt.Errorf("maximum validation not supported for %s", value.Kind())
	}
	return nil
}