package gongular

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// HandlerStat keeps duration, error (if exists) and whether the chain was stopped for
//  a single handler
type HandlerStat struct {
	FuncName  string
	Duration  time.Duration
	Error     error
	StopChain bool
}

// RouteStat holds information for the whole route, which path it matched, the written
// response size and the final status code for the request, and finally the logs generated
// by all handlers and it includes the individual HandlerStat this route consists of.
type RouteStat struct {
	Request       *http.Request
	Handlers      []HandlerStat
	MatchedPath   string
	TotalDuration time.Duration
	ResponseSize  int
	ResponseCode  int
	Logs          *bytes.Buffer
}

// RouteCallback is the interface to what to do with a given route
type RouteCallback func(stat RouteStat)

// DefaultRouteCallback prints many information about the the request including
// the individual  handler(s) information as well
var DefaultRouteCallback RouteCallback = func(stat RouteStat) {
	s := fmt.Sprintln(stat.Request.Method, stat.Request.RemoteAddr, stat.MatchedPath,
		stat.Request.RequestURI, stat.TotalDuration, stat.ResponseSize, stat.ResponseCode)

	for idx, h := range stat.Handlers {
		s += fmt.Sprintln("\t", idx, " ", h.FuncName, " ", h.Duration)
	}

	// All the information is concatenated to avoid race conditions that might occur
	fmt.Println(s)
}

// NoOpRouteCallback is doing nothing for a RouteCallback which should increases the performance
// which can be desirable when too many requests arrive
var NoOpRouteCallback = func(stat RouteStat) {}
