package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testCases := []*struct {
		Name          string
		Binary        string
		Args          []string
		WantResult    []byte
		WantErr       error
		WantErrString string
	}{
		{
			Name:          "binary not found",
			Binary:        "ls-not-found",
			Args:          nil,
			WantResult:    nil,
			WantErr:       ErrBinNotFound,
			WantErrString: "cannot start command \"ls-not-found\" with args : binary not found",
		},
		{
			Name:          "binary not found with args",
			Binary:        "ls-not-found",
			Args:          []string{"arg1", "arg2"},
			WantResult:    nil,
			WantErr:       ErrBinNotFound,
			WantErrString: "cannot start command \"ls-not-found\" with args arg1, arg2: binary not found",
		},
		{
			Name:          "failed",
			Binary:        "ls",
			Args:          []string{"123asdf"},
			WantErrString: "cmd.Wait: command \"ls\", args \"123asdf\": exit status 2",
		},
		{
			Name:       "success",
			Binary:     "ls",
			WantResult: []byte("errors.go\nexecute.go\nexecute_test.go\n"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			buf, err := Execute(ctx, "", testCase.Binary, testCase.Args...)
			if err != nil {
				if testCase.WantErr != nil {
					assert.ErrorIs(t, err, testCase.WantErr)
				}

				if testCase.WantErrString != "" {
					assert.EqualError(t, err, testCase.WantErrString)
				} else {
					assert.NoError(t, err)
				}
			}

			if testCase.WantResult != nil {
				if assert.NotNil(t, buf) {
					assert.Equal(t, testCase.WantResult, buf.Bytes())
				}
			} else {
				t.Log(buf.String())
			}
		})
	}
}
