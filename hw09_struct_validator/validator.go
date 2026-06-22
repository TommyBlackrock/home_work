package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidInput    = errors.New("input must be a struct or pointer to struct")
	ErrInvalidRule     = errors.New("invalid validation rule")
	ErrUnsupportedType = errors.New("unsupported field type")

	ErrValidationLen    = errors.New("length validation failed")
	ErrValidationRegexp = errors.New("regexp validation failed")
	ErrValidationIn     = errors.New("in validation failed")
	ErrValidationMin    = errors.New("min validation failed")
	ErrValidationMax    = errors.New("max validation failed")
)

type ValidationError struct {
	Field string
	Err   error
}

func (v ValidationError) Error() string {
	if v.Err == nil {
		return v.Field
	}
	if v.Field == "" {
		return v.Err.Error()
	}

	return fmt.Sprintf("%s: %v", v.Field, v.Err)
}

func (v ValidationError) Unwrap() error {
	return v.Err
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}

	parts := make([]string, 0, len(v))
	for _, validationErr := range v {
		parts = append(parts, validationErr.Error())
	}

	return strings.Join(parts, "; ")
}

func (v ValidationErrors) Unwrap() []error {
	errs := make([]error, 0, len(v))
	for _, validationErr := range v {
		errs = append(errs, validationErr)
	}

	return errs
}

type rule struct {
	name  string
	value string
}

type scalarRuleValidator func(reflect.Value, rule, string) (error, error)

const (
	validateTag = "validate"

	ruleNested = "nested"
	ruleLen    = "len"
	ruleRegexp = "regexp"
	ruleIn     = "in"
	ruleMin    = "min"
	ruleMax    = "max"
)

var scalarRuleHandlers = map[reflect.Kind]map[string]scalarRuleValidator{
	reflect.String: {
		ruleLen:    validateStringLen,
		ruleRegexp: validateStringRegexp,
		ruleIn:     validateStringIn,
	},
	reflect.Int: {
		ruleMin: validateIntMin,
		ruleMax: validateIntMax,
		ruleIn:  validateIntIn,
	},
	reflect.Int8: {
		ruleMin: validateIntMin,
		ruleMax: validateIntMax,
		ruleIn:  validateIntIn,
	},
	reflect.Int16: {
		ruleMin: validateIntMin,
		ruleMax: validateIntMax,
		ruleIn:  validateIntIn,
	},
	reflect.Int32: {
		ruleMin: validateIntMin,
		ruleMax: validateIntMax,
		ruleIn:  validateIntIn,
	},
	reflect.Int64: {
		ruleMin: validateIntMin,
		ruleMax: validateIntMax,
		ruleIn:  validateIntIn,
	},
}

func Validate(v interface{}) error {
	value, err := structValue(v)
	if err != nil {
		return err
	}

	return validateStruct(value, "")
}

func structValue(v interface{}) (reflect.Value, error) {
	if v == nil {
		return reflect.Value{}, ErrInvalidInput
	}

	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return reflect.Value{}, fmt.Errorf("%w: nil pointer", ErrInvalidInput)
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("%w: got %s", ErrInvalidInput, value.Kind())
	}

	return value, nil
}

func validateStruct(value reflect.Value, prefix string) error {
	var validationErrors ValidationErrors
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		fieldType := valueType.Field(i)
		tag, ok := fieldType.Tag.Lookup(validateTag)
		if !ok || tag == "" {
			continue
		}

		if !isExported(fieldType) && tag != ruleNested {
			continue
		}

		fieldPath := joinField(prefix, fieldType.Name)
		fieldErrors, err := validateField(value.Field(i), fieldPath, tag)
		if err != nil {
			return err
		}

		validationErrors = append(validationErrors, fieldErrors...)
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

func isExported(field reflect.StructField) bool {
	return field.PkgPath == ""
}

func validateField(value reflect.Value, fieldPath, tag string) (ValidationErrors, error) {
	rules, err := parseRules(tag)
	if err != nil {
		return nil, fmt.Errorf("%w: field %s: %w", ErrInvalidRule, fieldPath, err)
	}

	return validateValue(value, fieldPath, rules)
}

func parseRules(tag string) ([]rule, error) {
	rawRules := strings.Split(tag, "|")
	rules := make([]rule, 0, len(rawRules))
	for _, rawRule := range rawRules {
		if rawRule == "" {
			return nil, errors.New("empty rule")
		}

		if rawRule == ruleNested {
			rules = append(rules, rule{name: rawRule})
			continue
		}

		parts := strings.SplitN(rawRule, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("rule %q must have name and value", rawRule)
		}

		rules = append(rules, rule{name: parts[0], value: parts[1]})
	}

	return rules, nil
}

func isNestedRule(rules []rule) bool {
	return len(rules) == 1 && rules[0].name == ruleNested
}

func hasNestedRule(rules []rule) bool {
	for _, validationRule := range rules {
		if validationRule.name == ruleNested {
			return true
		}
	}

	return false
}

func validateValue(value reflect.Value, fieldPath string, rules []rule) (ValidationErrors, error) {
	nested := isNestedRule(rules)
	if !nested && hasNestedRule(rules) {
		return nil, fmt.Errorf("%w: field %s: nested cannot be combined with other rules",
			ErrInvalidRule, fieldPath)
	}

	if nested {
		indirectValue, ok := dereferenceNestedValue(value)
		if !ok {
			return nil, nil
		}
		value = indirectValue
	}

	//nolint:exhaustive // The validator dispatches supported kinds here; default returns a typed error.
	switch value.Kind() {
	case reflect.Struct:
		if nested {
			return collectNestedErrors(validateStruct(value, fieldPath))
		}

		return nil, unsupportedTypeError(fieldPath, value.Kind())
	case reflect.Slice:
		return validateSlice(value, fieldPath, rules)
	default:
		if nested {
			return nil, unsupportedTypeError(fieldPath, value.Kind())
		}

		return validateScalar(value, fieldPath, rules)
	}
}

func dereferenceNestedValue(value reflect.Value) (reflect.Value, bool) {
	if value.Kind() != reflect.Ptr {
		return value, true
	}
	if value.IsNil() {
		return reflect.Value{}, false
	}

	return value.Elem(), true
}

func collectNestedErrors(err error) (ValidationErrors, error) {
	if err == nil {
		return nil, nil
	}

	var validationErrors ValidationErrors
	if errors.As(err, &validationErrors) {
		return validationErrors, nil
	}

	return nil, err
}

func validateSlice(value reflect.Value, fieldPath string, rules []rule) (ValidationErrors, error) {
	if !isNestedRule(rules) {
		if _, ok := scalarRuleHandlers[value.Type().Elem().Kind()]; !ok {
			return nil, unsupportedTypeError(fieldPath, value.Type())
		}
	}

	var validationErrors ValidationErrors
	for i := 0; i < value.Len(); i++ {
		itemErrors, err := validateValue(value.Index(i), fieldPath, rules)
		if err != nil {
			return nil, err
		}

		validationErrors = append(validationErrors, itemErrors...)
	}

	return validationErrors, nil
}

func validateScalar(value reflect.Value, fieldPath string, rules []rule) (ValidationErrors, error) {
	handlers, ok := scalarRuleHandlers[value.Kind()]
	if !ok {
		return nil, unsupportedTypeError(fieldPath, value.Kind())
	}

	var validationErrors ValidationErrors
	for _, validationRule := range rules {
		handler, ok := handlers[validationRule.name]
		if !ok {
			return nil, unsupportedRuleError(fieldPath, validationRule, value.Kind())
		}

		validationErr, err := handler(value, validationRule, fieldPath)
		if err != nil {
			return nil, err
		}
		if validationErr != nil {
			validationErrors = append(validationErrors, ValidationError{
				Field: fieldPath,
				Err:   validationErr,
			})
		}
	}

	return validationErrors, nil
}

func validateStringLen(value reflect.Value, validationRule rule, fieldPath string) (error, error) {
	expectedLen, err := parseInt(validationRule, fieldPath)
	if err != nil {
		return nil, err
	}

	actualLen := int64(utf8.RuneCountInString(value.String()))
	if actualLen != expectedLen {
		return fmt.Errorf("%w: expected %d, got %d", ErrValidationLen, expectedLen, actualLen), nil
	}

	return nil, nil
}

func validateStringRegexp(value reflect.Value, validationRule rule, fieldPath string) (error, error) {
	re, err := regexp.Compile(validationRule.value)
	if err != nil {
		return nil, fmt.Errorf("%w: field %s regexp %q: %w",
			ErrInvalidRule, fieldPath, validationRule.value, err)
	}

	stringValue := value.String()
	match := re.FindStringIndex(stringValue)
	if match == nil || match[0] != 0 || match[1] != len(stringValue) {
		return fmt.Errorf("%w: value %q does not match %q",
			ErrValidationRegexp, stringValue, validationRule.value), nil
	}

	return nil, nil
}

func validateStringIn(value reflect.Value, validationRule rule, _ string) (error, error) {
	stringValue := value.String()
	if !containsString(strings.Split(validationRule.value, ","), stringValue) {
		return fmt.Errorf("%w: value %q is not in [%s]",
			ErrValidationIn, stringValue, validationRule.value), nil
	}

	return nil, nil
}

func validateIntMin(value reflect.Value, validationRule rule, fieldPath string) (error, error) {
	minValue, err := parseInt(validationRule, fieldPath)
	if err != nil {
		return nil, err
	}

	intValue := value.Int()
	if intValue < minValue {
		return fmt.Errorf("%w: expected at least %d, got %d", ErrValidationMin, minValue, intValue), nil
	}

	return nil, nil
}

func validateIntMax(value reflect.Value, validationRule rule, fieldPath string) (error, error) {
	maxValue, err := parseInt(validationRule, fieldPath)
	if err != nil {
		return nil, err
	}

	intValue := value.Int()
	if intValue > maxValue {
		return fmt.Errorf("%w: expected at most %d, got %d", ErrValidationMax, maxValue, intValue), nil
	}

	return nil, nil
}

func validateIntIn(value reflect.Value, validationRule rule, fieldPath string) (error, error) {
	allowedValues, err := parseIntList(validationRule, fieldPath)
	if err != nil {
		return nil, err
	}

	intValue := value.Int()
	if !containsInt(allowedValues, intValue) {
		return fmt.Errorf("%w: value %d is not in [%s]",
			ErrValidationIn, intValue, validationRule.value), nil
	}

	return nil, nil
}

func unsupportedTypeError(fieldPath string, fieldType fmt.Stringer) error {
	return fmt.Errorf("%w: field %s has type %s", ErrUnsupportedType, fieldPath, fieldType)
}

func unsupportedRuleError(fieldPath string, validationRule rule, kind reflect.Kind) error {
	return fmt.Errorf(
		"%w: field %s rule %q is not applicable to %s",
		ErrInvalidRule,
		fieldPath,
		validationRule.name,
		kind,
	)
}

func parseInt(validationRule rule, fieldPath string) (int64, error) {
	value, err := strconv.ParseInt(validationRule.value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: field %s rule %q expects integer value: %w",
			ErrInvalidRule, fieldPath, validationRule.name, err,
		)
	}

	return value, nil
}

func parseIntList(validationRule rule, fieldPath string) ([]int64, error) {
	rawValues := strings.Split(validationRule.value, ",")
	values := make([]int64, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: field %s rule %q expects integer list: %w",
				ErrInvalidRule, fieldPath, validationRule.name, err,
			)
		}

		values = append(values, value)
	}

	return values, nil
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}

	return false
}

func containsInt(values []int64, value int64) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}

	return false
}

func joinField(prefix, name string) string {
	if prefix == "" {
		return name
	}

	return prefix + "." + name
}
