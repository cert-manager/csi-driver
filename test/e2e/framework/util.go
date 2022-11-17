/*
Copyright 2021 The cert-manager Authors.

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

package framework

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/cert-manager/csi-driver/test/e2e/framework/log"
)

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func Failf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	Logf(msg)
	Fail(nowStamp()+": "+msg, 1)
}

func Skipf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	Logf("INFO", msg)
	Skip(nowStamp() + ": " + msg)
}

func boolPtr(b bool) *bool {
	return &b
}
