package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockHandler_ServeHTTP(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	handler := newMockHandler(mappings)

	req1 := httptest.NewRequest("POST", "/hello", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(http.StatusOK, rr1.Code)

	req2 := httptest.NewRequest("", "/hello", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(http.StatusMethodNotAllowed, rr2.Code)

	req3 := httptest.NewRequest("", "/notfound", nil)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	assert.Equal(http.StatusNotFound, rr3.Code)

	req4 := httptest.NewRequest("", "/m", nil)
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req4)
	assert.Equal(http.StatusBadRequest, rr4.Code)

	req5 := httptest.NewRequest("POST", "/m", strings.NewReader("120"))
	rr5 := httptest.NewRecorder()
	handler.ServeHTTP(rr5, req5)
	assert.Equal(http.StatusOK, rr5.Code)
	assert.Equal("TEST/v1", rr5.Header().Get("Server"))

	req6 := httptest.NewRequest("GET", "/m?p1=v1&p2=v1&p2=v2", nil)
	rr6 := httptest.NewRecorder()
	handler.ServeHTTP(rr6, req6)
	assert.Equal(http.StatusBadGateway, rr6.Code)
}

func TestMockHandler_handleExecuteError(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	handler := newMockHandler(mappings)

	req1 := httptest.NewRequest("GET", "/", nil)
	rr1 := httptest.NewRecorder()
	handler.handleExecuteError(rr1, req1, &forwardError{err: errors.New("test")})
	assert.Equal(http.StatusBadGateway, rr1.Code)

	req2 := httptest.NewRequest("GET", "/", nil)
	rr2 := httptest.NewRecorder()
	handler.handleExecuteError(rr2, req2, errors.New("test"))
	assert.Equal(http.StatusInternalServerError, rr2.Code)
}

func TestMockHandler_listAllMappings(t *testing.T) {
	handler := newMockHandler(mappings)
	handler.listAllMappings()
}
