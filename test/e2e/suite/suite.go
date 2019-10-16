/*
Copyright 2019 The Jetstack cert-manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package suite

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"

	"github.com/jetstack/cert-manager-csi/test/e2e/environment"
)

var (
	Writer = GinkgoWriter
	env    *environment.Environment
)

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	env, err = environment.Create(1, 3)
	if err != nil {
		failf(err.Error())
	}

	return nil
}, func([]byte) {
})

var globalLogs map[string]string

var _ = SynchronizedAfterSuite(func() {},
	func() {
		if err := env.Destory(); err != nil {
			failf(err.Error())
		}
	},
)

func failf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logf(msg)
	Fail(nowStamp()+": "+msg, 1)
}

func log(level string, format string, args ...interface{}) {
	fmt.Fprintf(Writer, nowStamp()+": "+level+": "+format+"\n", args...)
}

func logf(format string, args ...interface{}) {
	log("INFO", format, args...)
}
func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}
