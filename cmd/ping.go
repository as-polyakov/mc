// Copyright (c) 2015-2022 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"errors"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/minio/cli"
	json "github.com/minio/colorjson"
	"github.com/minio/madmin-go"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/pkg/console"
)

var pingFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "count, c",
		Usage: "perform liveliness check for count number of times",
	},
	cli.IntFlag{
		Name:  "error-count, e",
		Usage: "exit after N consecutive ping errors",
	},
	cli.IntFlag{
		Name:  "interval, i",
		Usage: "wait interval between each request in seconds",
		Value: 1,
	},
	cli.BoolFlag{
		Name:  "distributed, a",
		Usage: "ping all the servers in the cluster, use it when you have direct access to nodes/pods",
	},
}

// return latency and liveness probe.
var pingCmd = cli.Command{
	Name:            "ping",
	Usage:           "perform liveness check",
	Action:          mainPing,
	Before:          setGlobalsFromContext,
	OnUsageError:    onUsageError,
	Flags:           append(pingFlags, globalFlags...),
	HideHelpCommand: true,
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} [FLAGS] TARGET [TARGET...]
{{if .VisibleFlags}}
FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}
EXAMPLES:
  1. Return Latency and liveness probe.
     {{.Prompt}} {{.HelpName}} myminio

  2. Return Latency and liveness probe 5 number of times.
     {{.Prompt}} {{.HelpName}} --count 5 myminio

  3. Return Latency and liveness with wait interval set to 30 seconds.
     {{.Prompt}} {{.HelpName}} --interval 30 myminio

  4. Stop pinging when error count > 20.
  	 {{.Prompt}} {{.HelpName}} --error-count 20 myminio	 
`,
}

var stop bool

// Validate command line arguments.
func checkPingSyntax(cliCtx *cli.Context) {
	if !cliCtx.Args().Present() {
		cli.ShowCommandHelpAndExit(cliCtx, "ping", 1) // last argument is exit code
	}
}

// JSON jsonified ping result message.
func (pr PingResult) JSON() string {
	statusJSONBytes, e := json.MarshalIndent(pr, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")

	return string(statusJSONBytes)
}

var colorMap = template.FuncMap{
	"colorWhite": color.New(color.FgWhite).SprintfFunc(),
	"colorRed":   color.New(color.FgRed).SprintfFunc(),
}

// PingDist is the template for ping result in distributed mode
const PingDist = `{{$x := .Counter}}{{range .EndPointsStats}}{{if eq "0" .CountErr}}{{colorWhite $x}}{{colorWhite ": "}}{{colorWhite .Endpoint.Scheme}}{{colorWhite "://"}}{{colorWhite .Endpoint.Host}}{{if ne "" .Endpoint.Port}}{{colorWhite ":"}}{{colorWhite .Endpoint.Port}}{{end}}{{"\t"}}{{ colorWhite "min="}}{{colorWhite .Min}}{{"\t"}}{{colorWhite "max="}}{{colorWhite .Max}}{{"\t"}}{{colorWhite "average="}}{{colorWhite .Average}}{{"\t"}}{{colorWhite "errors="}}{{colorWhite .CountErr}}{{"\t"}}{{colorWhite "roundtrip="}}{{colorWhite .Roundtrip}}{{else}}{{colorRed $x}}{{colorRed ": "}}{{colorRed .Endpoint.Scheme}}{{colorRed "://"}}{{colorRed .Endpoint.Host}}{{if ne "" .Endpoint.Port}}{{colorRed ":"}}{{colorRed .Endpoint.Port}}{{end}}{{"\t"}}{{ colorRed "min="}}{{colorRed .Min}}{{"\t"}}{{colorRed "max="}}{{colorRed .Max}}{{"\t"}}{{colorRed "average="}}{{colorRed .Average}}{{"\t"}}{{colorRed "errors="}}{{colorRed .CountErr}}{{"\t"}}{{colorRed "roundtrip="}}{{colorRed .Roundtrip}}{{end}}
{{end}}`

// Ping is the template for ping result
const Ping = `{{$x := .Counter}}{{range .EndPointsStats}}{{if eq "0" .CountErr}}{{colorWhite $x}}{{colorWhite ": "}}{{colorWhite .Endpoint.Scheme}}{{colorWhite "://"}}{{colorWhite .Endpoint.Host}}{{if ne "" .Endpoint.Port}}{{colorWhite ":"}}{{colorWhite .Endpoint.Port}}{{end}}{{"\t"}}{{ colorWhite "min="}}{{colorWhite .Min}}{{"\t"}}{{colorWhite "max="}}{{colorWhite .Max}}{{"\t"}}{{colorWhite "average="}}{{colorWhite .Average}}{{"\t"}}{{colorWhite "errors="}}{{colorWhite .CountErr}}{{"\t"}}{{colorWhite "roundtrip="}}{{colorWhite .Roundtrip}}{{else}}{{colorRed $x}}{{colorRed ": "}}{{colorRed .Endpoint.Scheme}}{{colorRed "://"}}{{colorRed .Endpoint.Host}}{{if ne "" .Endpoint.Port}}{{colorRed ":"}}{{colorRed .Endpoint.Port}}{{end}}{{"\t"}}{{ colorRed "min="}}{{colorRed .Min}}{{"\t"}}{{colorRed "max="}}{{colorRed .Max}}{{"\t"}}{{colorRed "average="}}{{colorRed .Average}}{{"\t"}}{{colorRed "errors="}}{{colorRed .CountErr}}{{"\t"}}{{colorRed "roundtrip="}}{{colorRed .Roundtrip}}{{end}}{{end}}`

// PingTemplateDist - captures ping template
var PingTemplateDist = template.Must(template.New("ping-list").Funcs(colorMap).Parse(PingDist))

// PingTemplate - captures ping template
var PingTemplate = template.Must(template.New("ping-list").Funcs(colorMap).Parse(Ping))

// String colorized service status message.
func (pr PingResult) String() string {
	var s strings.Builder
	w := tabwriter.NewWriter(&s, 1, 8, 3, ' ', 0)
	var e error
	if len(pr.EndPointsStats) > 1 {
		e = PingTemplateDist.Execute(w, pr)
	} else {
		e = PingTemplate.Execute(w, pr)
	}
	fatalIf(probe.NewError(e), "Unable to initialize template writer")
	w.Flush()
	return s.String()
}

// Endpoint - container to hold server info
type Endpoint struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Port   string `json:"port"`
}

// EndPointStats - container to hold server ping stats
type EndPointStats struct {
	Endpoint  Endpoint `json:"endpoint"`
	Min       string   `json:"min"`
	Max       string   `json:"max"`
	Average   string   `json:"average"`
	CountErr  string   `json:"error-count,omitempty"`
	Error     string   `json:"error,omitempty"`
	Roundtrip string   `json:"roundtrip"`
}

// PingResult contains ping output
type PingResult struct {
	Status         string          `json:"status"`
	Counter        string          `json:"counter"`
	EndPointsStats []EndPointStats `json:"servers"`
}

type serverStats struct {
	min        uint64
	max        uint64
	sum        uint64
	avg        uint64
	errorCount int // used to keep a track of consecutive errors
	err        string
	counter    int // used to find the average, acts as denominator
}

func fetchAdminInfo(admClnt *madmin.AdminClient) (madmin.InfoMessage, error) {
	ctx, cancel := context.WithTimeout(globalContext, 3*time.Second)
	// Fetch the service status of the specified MinIO server
	info, e := admClnt.ServerInfo(ctx)
	cancel()
	if e == nil {
		return info, nil
	}

	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	for {
		select {
		case <-globalContext.Done():
			return madmin.InfoMessage{}, globalContext.Err()
		case <-timer.C:
			ctx, cancel := context.WithTimeout(globalContext, 3*time.Second)
			info, e := admClnt.ServerInfo(ctx)
			cancel()
			if e == nil {
				return info, nil
			}
			timer.Reset(time.Second)
		}
	}
}

func ping(ctx context.Context, cliCtx *cli.Context, anonClient *madmin.AnonymousClient, admInfo madmin.InfoMessage, endPointMap map[string]serverStats, index int) {
	var endPointStats []EndPointStats
	var servers []madmin.ServerProperties
	if cliCtx.Bool("distributed") {
		servers = admInfo.Servers
	}

	for result := range anonClient.Alive(ctx, madmin.AliveOpts{}, servers...) {
		host, port, _ := extractHostPort(result.Endpoint.String())
		endPoint := Endpoint{result.Endpoint.Scheme, host, port}
		stat := getPingInfo(cliCtx, result, endPointMap)
		endPointStat := EndPointStats{
			Endpoint:  endPoint,
			Min:       time.Duration(stat.min).Round(time.Microsecond).String(),
			Max:       time.Duration(stat.max).Round(time.Microsecond).String(),
			Average:   time.Duration(stat.avg).Round(time.Microsecond).String(),
			CountErr:  strconv.Itoa(stat.errorCount),
			Error:     stat.err,
			Roundtrip: result.ResponseTime.Round(time.Microsecond).String(),
		}
		endPointStats = append(endPointStats, endPointStat)
		endPointMap[result.Endpoint.Host] = stat

	}
	printMsg(PingResult{
		Status:         "success",
		Counter:        strconv.Itoa(index),
		EndPointsStats: endPointStats,
	})

	time.Sleep(time.Duration(cliCtx.Int("interval")) * time.Second)
}

func getPingInfo(cliCtx *cli.Context, result madmin.AliveResult, serverMap map[string]serverStats) serverStats {
	var errorString string
	var sum, avg uint64
	min := uint64(math.MaxUint64)
	var max uint64
	var counter, errorCount int

	if result.Error != nil {
		errorString = result.Error.Error()
		if stat, ok := serverMap[result.Endpoint.Host]; ok {
			min = stat.min
			max = stat.max
			sum = stat.sum
			counter = stat.counter
			avg = stat.avg
			errorCount = stat.errorCount + 1

		} else {
			min = 0
			errorCount = 1
		}
		if cliCtx.IsSet("error-count") && errorCount >= cliCtx.Int("error-count") {
			stop = true
		}

	} else {
		// reset consecutive error count
		errorCount = 0
		if stat, ok := serverMap[result.Endpoint.Host]; ok {
			var minVal uint64
			if stat.min == 0 {
				minVal = uint64(result.ResponseTime)
			} else {
				minVal = stat.min
			}
			min = uint64(math.Min(float64(minVal), float64(uint64(result.ResponseTime))))
			max = uint64(math.Max(float64(stat.max), float64(uint64(result.ResponseTime))))
			sum = stat.sum + uint64(result.ResponseTime.Nanoseconds())
			counter = stat.counter + 1

		} else {
			min = uint64(math.Min(float64(min), float64(uint64(result.ResponseTime))))
			max = uint64(math.Max(float64(max), float64(uint64(result.ResponseTime))))
			sum = uint64(result.ResponseTime)
			counter = 1
		}
		avg = sum / uint64(counter)
	}
	return serverStats{min, max, sum, avg, errorCount, errorString, counter}
}

// extractHostPort - extracts host/port from many address formats
// such as, ":9000", "localhost:9000", "http://localhost:9000/"
func extractHostPort(hostAddr string) (string, string, error) {
	var addr, scheme string

	if hostAddr == "" {
		return "", "", errors.New("unable to process empty address")
	}

	// Simplify the work of url.Parse() and always send a url with
	if !strings.HasPrefix(hostAddr, "http://") && !strings.HasPrefix(hostAddr, "https://") {
		hostAddr = "//" + hostAddr
	}

	// Parse address to extract host and scheme field
	u, err := url.Parse(hostAddr)
	if err != nil {
		return "", "", err
	}

	addr = u.Host
	scheme = u.Scheme

	// Use the given parameter again if url.Parse()
	// didn't return any useful result.
	if addr == "" {
		addr = hostAddr
		scheme = "http"
	}

	// At this point, addr can be one of the following form:
	//	":9000"
	//	"localhost:9000"
	//	"localhost" <- in this case, we check for scheme

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if !strings.Contains(err.Error(), "missing port in address") {
			return "", "", err
		}

		host = addr

		switch scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			return "", "", errors.New("unable to guess port from scheme")
		}
	}

	return host, port, nil
}

// mainPing is entry point for ping command.
func mainPing(cliCtx *cli.Context) error {
	// check 'ping' cli arguments.
	checkPingSyntax(cliCtx)

	console.SetColor("Info", color.New(color.FgGreen, color.Bold))
	console.SetColor("InfoFail", color.New(color.FgRed, color.Bold))

	ctx, cancel := context.WithCancel(globalContext)
	defer cancel()

	aliasedURL := cliCtx.Args().Get(0)
	admClient, err := newAdminClient(aliasedURL)
	fatalIf(err.Trace(aliasedURL), "Unable to initialize admin client for `"+aliasedURL+"`.")

	anonClient, err := newAnonymousClient(aliasedURL)
	fatalIf(err.Trace(aliasedURL), "Unable to initialize anonymous client for `"+aliasedURL+"`.")

	var admInfo madmin.InfoMessage
	if cliCtx.Bool("distributed") {
		var e error
		admInfo, e = fetchAdminInfo(admClient)
		fatalIf(probe.NewError(e).Trace(aliasedURL), "Unable to get server info")
	}

	// map to contain server stats for all the servers
	serverMap := make(map[string]serverStats)

	index := 1
	if cliCtx.IsSet("count") {
		count := cliCtx.Int("count")
		if count < 1 {
			fatalIf(errInvalidArgument().Trace(cliCtx.Args()...), "ping count cannot be less than 1")
		}
		for index <= count {
			// return if consecutive error count more then specified value
			if stop {
				return nil
			}
			ping(ctx, cliCtx, anonClient, admInfo, serverMap, index)
			index++
		}
	} else {
		for {
			select {
			case <-globalContext.Done():
				return globalContext.Err()
			default:
				// return if consecutive error count more then specified value
				if stop {
					return nil
				}
				ping(ctx, cliCtx, anonClient, admInfo, serverMap, index)
				index++
			}
		}
	}
	return nil
}
