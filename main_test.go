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
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"testing"

	"github.com/mendersoftware/go-lib-micro/log"
)

var runAcceptanceTests bool

func init() {
	// disable logging thile running unit tests
	// default application settup couses to mich noice
	log.Log.Out = ioutil.Discard

	flag.BoolVar(&runAcceptanceTests, "acceptance-tests", false, "set flag when running acceptance tests")
	flag.Parse()
}

func TestHandleConfigFile(t *testing.T) {
	t.Parallel()

	HandleConfigFile("", false, nil)
}

func TestRunMain(t *testing.T) {
	if !runAcceptanceTests {
		t.Skip()
	}

	go main()

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	<-stopChan
}
