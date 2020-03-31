package server

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/stretchr/testify/assert"
)

func TestForwardError_Error(t *testing.T) {
	e := &forwardError{err: errors.New("test_error")}
	assert.NotNil(t, e)
	assert.Contains(t, e.Error(), "test_error")
}

func TestWaitBeforeReturns(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	i1 := &mckmaps.Interval{Min: 100, Max: 100}
	i1Start := time.Now()
	waitBeforeReturns(i1)
	i1Elapsed := time.Since(i1Start)
	assert.True(i1Elapsed.Milliseconds() >= 100)

	i2 := &mckmaps.Interval{Min: 10, Max: 30}
	i2Start := time.Now()
	for i := 0; i < 10; i++ {
		waitBeforeReturns(i2)
	}
	i2Elapsed := time.Since(i2Start)
	assert.True(i2Elapsed.Milliseconds() >= 100 && i2Elapsed.Milliseconds() <= 350)
}

func TestPolicyExecutor_executor(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	rr1 := httptest.NewRecorder()
	var rw1 http.ResponseWriter = rr1
	exe1 := &policyExecutor{
		r:      httptest.NewRequest("GET", "/TestPolicyExecutor_executor", nil),
		w:      &rw1,
		policy: newStatusJSONPolicy(myhttp.StatusOk, "OK"),
	}
	e1 := exe1.execute()
	if assert.Nil(e1) {
		assert.Equal(http.StatusOK, rr1.Code)
		assert.Equal(fmt.Sprintf(`{"statusCode": %d, "message": "%s"}`, http.StatusOK, "OK"), rr1.Body.String())
	}

	rr2 := httptest.NewRecorder()
	var rw2 http.ResponseWriter = rr2
	exe2 := &policyExecutor{
		r:      httptest.NewRequest("GET", "/TestPolicyExecutor_executor", nil),
		w:      &rw2,
		policy: pNoPolicyMatched,
	}
	e2 := exe2.execute()
	if assert.Nil(e2) {
		assert.Equal(http.StatusBadRequest, rr2.Code)
		assert.Equal(string(pNoPolicyMatched.Returns.Body), rr2.Body.String())
	}

	rr3Start := time.Now()
	rr3 := httptest.NewRecorder()
	var rw3 http.ResponseWriter = rr3
	exe3 := &policyExecutor{
		r: httptest.NewRequest("GET", "/TestPolicyExecutor_executor", nil),
		w: &rw3,
		policy: &mckmaps.Policy{
			CmdType: mckmaps.CmdTypeReturns,
			Returns: &mckmaps.Returns{
				StatusCode: myhttp.StatusOk,
				Latency: &mckmaps.Interval{
					Min: 100,
					Max: 100,
				},
				Headers: []*mckmaps.NameValuesPair{
					{Name: myhttp.HeaderContentType, Values: []string{myhttp.ContentTypeJSON}},
				},
				Body: []byte(fmt.Sprintf(`{"statusCode": %d, "message": "%s"}`, myhttp.StatusOk, "test")),
			},
		},
	}
	e3 := exe3.execute()
	assert.Nil(e3)
	rr3Elapsed := time.Since(rr3Start)
	assert.True(rr3Elapsed.Milliseconds() >= 100)

	rr4 := httptest.NewRecorder()
	var rw4 http.ResponseWriter = rr4
	exe4 := &policyExecutor{
		r: httptest.NewRequest("GET", "/TestPolicyExecutor_executor", nil),
		w: &rw4,
		policy: &mckmaps.Policy{
			CmdType: mckmaps.CmdTypeForwards,
			Forwards: &mckmaps.Forwards{
				Latency: &mckmaps.Interval{
					Min: 0,
					Max: 0,
				},
				Path: "https://www.example.com",
			},
		},
	}
	e4 := exe4.execute()
	assert.Nil(e4)
	assert.Equal(http.StatusOK, rr4.Code)

	handler := newMockHandler(mappings)
	rr5 := httptest.NewRecorder()
	var rw5 http.ResponseWriter = rr5
	exe5 := &policyExecutor{
		h: handler,
		r: httptest.NewRequest("POST", "/TestPolicyExecutor_executor", bytes.NewReader([]byte("120"))),
		w: &rw5,
		policy: &mckmaps.Policy{
			CmdType: mckmaps.CmdTypeForwards,
			Forwards: &mckmaps.Forwards{
				Path: "/m",
			},
		},
	}
	e5 := exe5.execute()
	assert.Nil(e5)
	assert.Equal(http.StatusOK, rr5.Code)
	assert.Equal("TEST/v1", rr5.Header().Get("Server"))

	rr6 := httptest.NewRecorder()
	var rw6 http.ResponseWriter = rr6
	exe6 := &policyExecutor{
		h: handler,
		r: httptest.NewRequest("GET", "/c?r1=120", nil),
		w: &rw6,
		policy: &mckmaps.Policy{
			CmdType: mckmaps.CmdTypeForwards,
			Forwards: &mckmaps.Forwards{
				Path: "m",
			},
		},
	}
	e6 := exe6.execute()
	assert.Nil(e6)
	assert.Equal(http.StatusOK, rr6.Code)

	rr7 := httptest.NewRecorder()
	var rw7 http.ResponseWriter = rr7
	exe7 := &policyExecutor{
		r: httptest.NewRequest("GET", "/TestPolicyExecutor_executor", nil),
		w: &rw7,
		policy: &mckmaps.Policy{
			CmdType: "wow",
		},
	}
	e7 := exe7.execute()
	assert.NotNil(e7)
	assert.Contains(e7.Error(), "wow")

	rr8 := httptest.NewRecorder()
	var rw8 http.ResponseWriter = rr8
	exe8 := &policyExecutor{
		r: httptest.NewRequest("GET", "/TestPolicyExecutor_executor", nil),
		w: &rw8,
		policy: &mckmaps.Policy{
			CmdType: mckmaps.CmdTypeForwards,
			Forwards: &mckmaps.Forwards{
				Path: "http://localhost:8080",
			},
		},
	}
	e8 := exe8.execute()
	assert.NotNil(e8)
}
