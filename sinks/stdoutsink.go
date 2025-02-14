/*
Copyright 2017 Heptio Inc.

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

package sinks

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

type StdoutSink struct {
	// TODO: create a channel and buffer for scaling
}

// NewStdoutSink will create a new StdoutSink with default options, returned as
// an EventSinkInterface
func NewStdoutSink() EventSinkInterface {
	return &StdoutSink{}
}

// UpdateEvents implements the EventSinkInterface
func (gs *StdoutSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	eData := NewEventData(eNew, eOld)

	if eJSONBytes, err := json.Marshal(eData); err == nil {
		fmt.Println(string(eJSONBytes))
	} else {
		klog.Warningf("Failed to json serialize event: %v", err)
	}
}
