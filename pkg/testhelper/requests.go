package testhelper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func NewRequest(t *testing.T, router *gin.Engine, method, path, token, payload string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var (
		req *http.Request
		err error
	)

	if payload == "" {
		req, err = http.NewRequest(method, path, nil)
	} else {
		req, err = http.NewRequest(method, path, strings.NewReader(payload))
		fmt.Printf("\n\nRequest: %s\n\n", payload)
	}

	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(w, req)
	require.NotNil(t, w)

	fmt.Printf("\n\nResponse: %s\n\n", w.Body.String())
	return w
}
