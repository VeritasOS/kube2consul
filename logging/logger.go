/*
Copyright 2018 Veritas Technologies LLC

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

package logging

import (
	"github.com/op/go-logging"
	"os"
)

// Logger defines a wrapper over *logging.Logger type from github.com/op/go-logging package
type Logger struct {
	*logging.Logger
}

var loggerInstance *Logger

// Inititalizes logger and formatter
func init() {
	loggerInstance = &Logger{logging.MustGetLogger("")}

	format := logging.MustStringFormatter(
		`{"application": "kube2consul", "timestamp": "%{time:15:04:05.000}", "source": "%{longfunc}", "loglevel": "%{level}", "message": "%{message}" }`,
	)
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)

}

// GetInstance returns an Logger instance
func GetInstance() *Logger {
	return loggerInstance
}

// InitLogger sets the LogLevel based on the cmd line flag. Must be called before GetInstance() otherwise logger will log everything.
func InitLogger(isDebug *bool) {
	if *isDebug {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
}
