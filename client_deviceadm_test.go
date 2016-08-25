// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main

import (
	"github.com/mendersoftware/deviceauth/log"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// return mock http server returning status code 'status'
func newMockServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))

}

func TestGetDevAdmClient(t *testing.T) {
	c := GetDevAdmClient(DevAdmClientConfig{AddDeviceUrl: "/foo"},
		log.New(log.Ctx{}))
	assert.NotNil(t, c)
}

func TestDevAdmClientReqSuccess(t *testing.T) {
	s := newMockServer(200)
	defer s.Close()

	addDevUrl := s.URL + "/devices"
	c := NewDevAdmClient(DevAdmClientConfig{
		AddDeviceUrl: addDevUrl,
	})

	err := c.AddDevice(&Device{}, &http.Client{})
	assert.NoError(t, err, "expected no errors")
}

func TestDevAdmClientReqFail(t *testing.T) {
	s := newMockServer(400)
	defer s.Close()

	addDevUrl := s.URL + "/devices"
	c := NewDevAdmClient(DevAdmClientConfig{
		AddDeviceUrl: addDevUrl,
	})

	err := c.AddDevice(&Device{}, &http.Client{})
	assert.NoError(t, err, "expected an error")
}

func TestDevAdmClientReqNoHost(t *testing.T) {
	c := NewDevAdmClient(DevAdmClientConfig{
		AddDeviceUrl: "http://somehost:1234/devices",
	})

	err := c.AddDevice(&Device{}, &http.Client{})

	assert.Error(t, err, "expected an error")
}

func TestDevAdmClientTImeout(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// channel for nofitying the responder that the test is
	// complete
	testdone := make(chan bool)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// wait for the test to notify us about timeout
		select {
		case <-testdone:
			// test finished, can leave now
		case <-time.After(defaultDevAdmReqTimeout * 2):
			// don't block longer than default timeout * 2
		}
		w.WriteHeader(400)
	}))

	addDevUrl := s.URL + "/devices"
	c := NewDevAdmClient(DevAdmClientConfig{
		AddDeviceUrl: addDevUrl,
	})

	t1 := time.Now()
	err := c.AddDevice(&Device{}, &http.Client{Timeout: defaultDevAdmReqTimeout})
	t2 := time.Now()

	// let the responder know we're done
	testdone <- true

	s.Close()

	assert.Error(t, err, "expected timeout error")
	// allow some slack in timeout, add 20% of the default timeout
	maxdur := defaultDevAdmReqTimeout +
		time.Duration(0.2*float64(defaultDevAdmReqTimeout))

	assert.WithinDuration(t, t2, t1, maxdur, "timeout took too long")
}
