package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"testing"
)

type UserRole string

type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	Meta struct {
		Revision int    `validate:"min:10"`
		Owner    string `validate:"in:root,admin"`
	}

	Document struct {
		Title string `validate:"len:5"`
		Meta  Meta   `validate:"nested"`
	}

	privateDocument struct {
		meta Meta `validate:"nested"`
	}
)

func TestValidateSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   interface{}
	}{
		{
			name: "valid struct",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "Ivan",
				Age:    30,
				Email:  "ivan@example.com",
				Role:   "admin",
				Phones: []string{"71234567890"},
			},
		},
		{
			name: "valid pointer to struct",
			in: &App{
				Version: "1.2.3",
			},
		},
		{
			name: "struct without validate tags",
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
		},
		{
			name: "int in rule",
			in: Response{
				Code: 200,
				Body: "ok",
			},
		},
		{
			name: "nested struct",
			in: Document{
				Title: "hello",
				Meta: Meta{
					Revision: 10,
					Owner:    "root",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := Validate(tt.in); err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCollectsValidationErrors(t *testing.T) {
	t.Parallel()

	err := Validate(User{
		ID:     "short",
		Name:   "Ivan",
		Age:    17,
		Email:  "bad-email",
		Role:   "guest",
		Phones: []string{"1", "71234567890"},
	})

	assertValidationErrors(t, err, 5)
	assertErrorIs(t, err, ErrValidationLen)
	assertErrorIs(t, err, ErrValidationMin)
	assertErrorIs(t, err, ErrValidationRegexp)
	assertErrorIs(t, err, ErrValidationIn)
}

func TestValidateCollectsSliceElementErrors(t *testing.T) {
	t.Parallel()

	type Ports struct {
		Values []int `validate:"min:10|max:20|in:10,15,20"`
	}

	err := Validate(Ports{Values: []int{5, 25, 12}})

	assertValidationErrors(t, err, 5)
	assertErrorIs(t, err, ErrValidationMin)
	assertErrorIs(t, err, ErrValidationMax)
	assertErrorIs(t, err, ErrValidationIn)
}

func TestValidateNestedErrors(t *testing.T) {
	t.Parallel()

	err := Validate(Document{
		Title: "hello",
		Meta: Meta{
			Revision: 1,
			Owner:    "guest",
		},
	})

	validationErrors := assertValidationErrors(t, err, 2)
	if validationErrors[0].Field != "Meta.Revision" {
		t.Fatalf("expected first nested field Meta.Revision, got %q", validationErrors[0].Field)
	}
	if validationErrors[1].Field != "Meta.Owner" {
		t.Fatalf("expected second nested field Meta.Owner, got %q", validationErrors[1].Field)
	}
}

func TestValidateUnexportedNestedField(t *testing.T) {
	t.Parallel()

	err := Validate(privateDocument{
		meta: Meta{
			Revision: 1,
			Owner:    "root",
		},
	})

	validationErrors := assertValidationErrors(t, err, 1)
	if validationErrors[0].Field != "meta.Revision" {
		t.Fatalf("expected nested field meta.Revision, got %q", validationErrors[0].Field)
	}
}

func TestValidateProgramErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name:        "not a struct",
			in:          "user",
			expectedErr: ErrInvalidInput,
		},
		{
			name:        "nil pointer",
			in:          (*User)(nil),
			expectedErr: ErrInvalidInput,
		},
		{
			name: "invalid integer rule",
			in: struct {
				Age int `validate:"min:bad"`
			}{Age: 18},
			expectedErr: ErrInvalidRule,
		},
		{
			name: "invalid regexp rule",
			in: struct {
				Value string `validate:"regexp:["`
			}{Value: "test"},
			expectedErr: ErrInvalidRule,
		},
		{
			name: "unsupported type with validate tag",
			in: struct {
				Data []byte `validate:"len:1"`
			}{Data: []byte("a")},
			expectedErr: ErrUnsupportedType,
		},
		{
			name: "nested with scalar field",
			in: struct {
				Value int `validate:"nested"`
			}{Value: 1},
			expectedErr: ErrUnsupportedType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(tt.in)
			assertErrorIs(t, err, tt.expectedErr)

			var validationErrors ValidationErrors
			if errors.As(err, &validationErrors) {
				t.Fatalf("expected program error, got validation errors: %v", validationErrors)
			}
		})
	}
}

func assertValidationErrors(t *testing.T, err error, expectedLen int) ValidationErrors {
	t.Helper()

	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	var validationErrors ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
	}

	if len(validationErrors) != expectedLen {
		t.Fatalf("expected %d validation errors, got %d: %v", expectedLen, len(validationErrors), validationErrors)
	}

	return validationErrors
}

func assertErrorIs(t *testing.T, err, expected error) {
	t.Helper()

	if !errors.Is(err, expected) {
		t.Fatalf("expected error %q in chain, got %v", expected, err)
	}
}
