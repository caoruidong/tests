//
// Copyright (c) 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testLevel = "info"
	testPid   = 2
	testName  = "kata-agent"
	testMsg   = "hello from the agent"
)

func TestAgentLogEntry(t *testing.T) {
	assert := assert.New(t)

	type testData struct {
		le        LogEntry
		fromAgent bool
	}

	data := []testData{
		{
			LogEntry{},
			false,
		},

		{
			LogEntry{
				Source: "agent",
				Msg:    "",
			},
			false,
		},

		{
			LogEntry{
				Source: "foo",
				Msg:    "time=\"2018-02-24T12:36:35.980548906Z\"",
			},
			false,
		},

		{
			LogEntry{
				Source: "agent",
				Msg:    "wibble",
			},
			false,
		},

		{
			LogEntry{
				Source: "agent",
				Msg:    "time=\"2018-02-24T12:36:35.980548906Z\"",
			},
			true,
		},
	}

	for i, d := range data {
		fromAgent := agentLogEntry(d.le)

		if d.fromAgent {
			assert.True(fromAgent, "test[%d]: %+v", i, d)
		} else {
			assert.False(fromAgent, "test[%d]: %+v", i, d)
		}
	}
}

func TestUnpackAgentLogEntry(t *testing.T) {
	assert := assert.New(t)

	type testData struct {
		le          LogEntry
		expectError bool
	}

	now := time.Now().UTC()
	nano := now.Format(time.RFC3339Nano)

	source := "agent"

	agentMsg := fmt.Sprintf("time=%q source=%s level=%s msg=%q pid=%d name=%s",
		nano, source, testLevel, testMsg, testPid, testName)

	data := []testData{
		{
			LogEntry{},
			true,
		},

		{
			LogEntry{
				Source: "agent",
				Msg:    "",
			},
			true,
		},

		{
			LogEntry{
				Source:   "agent",
				Msg:      "foo",
				Filename: "",
			},
			true,
		},

		{
			LogEntry{
				Source:   "agent",
				Msg:      "foo",
				Filename: "/foo/bar.log",
				Line:     0,
			},
			true,
		},

		{
			LogEntry{
				Count:    123,
				Source:   "agent",
				Filename: "/foo/bar.txt",
				Line:     101,
				Msg:      agentMsg,
			},
			false,
		},
	}

	for i, d := range data {
		agent, err := unpackAgentLogEntry(d.le)

		if d.expectError {
			assert.Errorf(err, "test[%d]: %+v", i, d)
		} else {
			assert.NoErrorf(err, "test[%d]: %+v", i, d)

			// Ensure the newly unpacked LogEntry is valid
			err = agent.Check()
			assert.NoError(err)

			assert.Equal(d.le.Filename, agent.Filename)
			assert.Equal(d.le.Line, agent.Line)
			assert.Equal(d.le.Count, agent.Count)
			assert.Equal(d.le.Source, agent.Source)
			assert.Equal(agent.Pid, testPid)
			assert.Equal(agent.Msg, testMsg)
			assert.Equal(agent.Name, testName)
			assert.Equal(agent.Level, testLevel)
			assert.Equal(agent.Time, now)
			assert.Equal(agent.Source, source)
		}
	}
}
