/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2019 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package statsite

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v3"

	"go.k6.io/k6/lib/testutils"
	"go.k6.io/k6/lib/types"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
)

func getOutput(
	logger logrus.FieldLogger, addr, namespace null.String, bufferSize null.Int, pushInterval types.NullDuration,
) (*Output, error) {
	return newOutput(
		output.Params{
			Logger: logger,
			JSONConfig: json.RawMessage(fmt.Sprintf(`{
			"addr": "%s",
			"namespace": "%s",
			"bufferSize": %d,
			"pushInterval": "%s"
		}`, addr.String, namespace.String, bufferSize.Int64, pushInterval.Duration.String())),
		})
}

func TestStatsdOutput(t *testing.T) {
	t.Parallel()
	baseTest(t, getOutput,
		func(t *testing.T, _ []metrics.SampleContainer, expectedOutput, output string) {
			assert.Equal(t, expectedOutput, output)
		})
}

func TestStatsdAppendTags(t *testing.T) {
	t.Parallel()
	appended := ".value1.value3"

	appendTest(t, appended, func(
		logger logrus.FieldLogger, addr, namespace null.String, bufferSize null.Int, pushInterval types.NullDuration,
	) (*Output, error) {
		return newOutput(
			output.Params{
				Logger: logger,
				JSONConfig: json.RawMessage(fmt.Sprintf(`{
			"addr": "%s",
			"namespace": "%s",
			"bufferSize": %d,
			"pushInterval": "%s",
			"tagAppend": ["tag1", "tag2", "tag3"],
			"enableTags": false
		}`, addr.String, namespace.String, bufferSize.Int64, pushInterval.Duration.String())),
			})
	}, func(t *testing.T, containers []metrics.SampleContainer, expectedOutput, output string) {
		outputLines := strings.Split(output, "\n")
		expectedOutputLines := strings.Split(expectedOutput, "\n")
		var lines int

		for i, container := range containers {
			for j := range container.GetSamples() {
				lines++
				expectedOutputLine := expectedOutputLines[i*j+i]
				outputLine := outputLines[i*j+i]

				require.Equal(t, expectedOutputLine, outputLine)
			}
		}
		require.Equal(t, lines, len(outputLines))
	})
}

func TestInitWithoutAddressErrors(t *testing.T) {
	t.Parallel()
	c := &Output{
		config: config{},
		logger: testutils.NewLogger(t),
	}
	err := c.Start()
	require.Error(t, err)
}

func TestInitWithBogusAddressErrors(t *testing.T) {
	t.Parallel()
	c := &Output{
		config: config{
			Addr: null.StringFrom("localhost:90000"),
		},
		logger: testutils.NewLogger(t),
	}
	err := c.Start()
	require.Error(t, err)
}

func TestLinkReturnAddress(t *testing.T) {
	t.Parallel()
	bogusValue := "bogus value"
	c := &Output{
		config: config{
			Addr: null.StringFrom(bogusValue),
		},
	}
	require.Equal(t, fmt.Sprintf("statsite (%s)", bogusValue), c.Description())
}
