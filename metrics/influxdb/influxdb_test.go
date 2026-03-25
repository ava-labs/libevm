// Copyright 2023 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package influxdb

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ava-labs/libevm/metrics"
	"github.com/ava-labs/libevm/metrics/internal"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// stddev and variance serialization can differ by one ULP across GOARCH / Go versions.
var influxStddevVarRE = regexp.MustCompile(`\b(stddev|variance)=([0-9.]+(?:[eE][+-]?[0-9]+)?)`)

func normalizeInfluxStddevVariance(s string) string {
	return influxStddevVarRE.ReplaceAllStringFunc(s, func(m string) string {
		eq := strings.Index(m, "=")
		key, val := m[:eq], m[eq+1:]
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return m
		}
		return fmt.Sprintf("%s=%.12e", key, f)
	})
}

func TestMain(m *testing.M) {
	metrics.Enabled = true
	os.Exit(m.Run())
}

func TestExampleV1(t *testing.T) {
	r := internal.ExampleMetrics()
	var have, want string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		haveB, _ := io.ReadAll(r.Body)
		have = string(haveB)
		r.Body.Close()
	}))
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	rep := &reporter{
		reg:       r,
		url:       *u,
		namespace: "goth.",
	}
	if err := rep.makeClient(); err != nil {
		t.Fatal(err)
	}
	if err := rep.send(978307200); err != nil {
		t.Fatal(err)
	}
	if wantB, err := os.ReadFile("./testdata/influxdbv1.want"); err != nil {
		t.Fatal(err)
	} else {
		want = string(wantB)
	}
	if normalizeInfluxStddevVariance(have) != normalizeInfluxStddevVariance(want) {
		t.Errorf("\nhave:\n%v\nwant:\n%v\n", have, want)
		t.Logf("have vs want:\n%v", findFirstDiffPos(have, want))
	}
}

func TestExampleV2(t *testing.T) {
	r := internal.ExampleMetrics()
	var have, want string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		haveB, _ := io.ReadAll(r.Body)
		have = string(haveB)
		r.Body.Close()
	}))
	defer ts.Close()

	rep := &v2Reporter{
		reg:       r,
		endpoint:  ts.URL,
		namespace: "goth.",
	}
	rep.client = influxdb2.NewClient(rep.endpoint, rep.token)
	defer rep.client.Close()
	rep.write = rep.client.WriteAPI(rep.organization, rep.bucket)

	rep.send(978307200)

	if wantB, err := os.ReadFile("./testdata/influxdbv2.want"); err != nil {
		t.Fatal(err)
	} else {
		want = string(wantB)
	}
	if normalizeInfluxStddevVariance(have) != normalizeInfluxStddevVariance(want) {
		t.Errorf("\nhave:\n%v\nwant:\n%v\n", have, want)
		t.Logf("have vs want:\n%v", findFirstDiffPos(have, want))
	}
}

func findFirstDiffPos(a, b string) string {
	yy := strings.Split(b, "\n")
	for i, x := range strings.Split(a, "\n") {
		if i >= len(yy) {
			return fmt.Sprintf("have:%d: %s\nwant:%d: <EOF>", i, x, i)
		}
		if y := yy[i]; x != y {
			return fmt.Sprintf("have:%d: %s\nwant:%d: %s", i, x, i, y)
		}
	}
	return ""
}
