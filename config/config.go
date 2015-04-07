// Package config reads a configuration file and generates an instance of
// Configuration.
//
// A valid configuration file is in JSON format.  The file contains a top-level
// object with the following members:
//
//     - "listen type": {string} (required) is the type of socket to open to
//     listen for hits.  On Windows, the only accepted value is "tcp".  Other
//     systems also accept "unix".
//
//     - "listen address": {string} (required) is the address that the system
//     should listen on for hits.  For TCP connections, use the format ":xxxx"
//     where xxxx is the port number.  For UNIX sockets, provide the path to
//     where the socket will be open.
//
//     - "dashboard address": {string} (optional) is the port on which to
//     launch the dashboard (e.g. ":3000").  If omitted or blank, the dashboard
//     will not be used.
//
//     - "log": {string} (optional) is the name of the file to write logs to.
//
//     - "directions": {array} (required) is the collection of directions you
//     want to track.  Each element of the array is an object with the
//     following members:
//
//         - "name": {string} (required) is the name that identifies the
//         direction.
//
//         - "type": {string} (required) is the type of data to track.  The
//         only accepted values are "string" and "int32".
//
//         - "window size": {number} (required) is the duration in seconds that
//         subsequent hits are tracked.
//
//         - "max hits": {number} (required) is the total number of hits that
//         are permitted within the window.  Any more hits within the window
//         will be flagged as an attack.
//
//         - "clean up time": {number} (optional) is the interval in seconds
//         that the clean up routine will run.  This routine clears irrelevant
//         values from memory.
//
//         - "max tracked": {number} (optional) is the maximum number of values
//         to track.  Once the limit has been reached, all new hits will be
//         flagged as attacks.  This is mainly to limit the amount of memory
//         that the system uses.  WARNING: don't use this unless you are
//         running out of memory.  The system can easily track millions of
//         values without using all of your memory.
package config

import (
	"fmt"
	"github.com/BruteForceFencer/core/hitcounter"
	"github.com/BruteForceFencer/core/logger"
	"github.com/BruteForceFencer/core/store"
	"os"
)

// Configuration is a struct that represents the contents of a configuration
// file.
type Configuration struct {
	Directions       []hitcounter.Direction
	ListenAddress    string
	ListenType       string
	DashboardAddress string
	AcceptedSources  []string
	Logger           *logger.Logger
}

// ReadConfig parses a configuration file and returns an instance of
// Configuration.
func ReadConfig(filename string) (*Configuration, []error) {
	parsed, err := parseJsonFile(filename)
	if err != nil {
		return nil, []error{err}
	}
	if errs := parsed.Validate(); len(errs) != 0 {
		return nil, errs
	}

	result := new(Configuration)
	result.ListenAddress = parsed.ListenAddress
	result.ListenType = parsed.ListenType
	result.DashboardAddress = parsed.DashboardAddress
	result.Directions = make([]hitcounter.Direction, 0, len(parsed.Directions))
	result.AcceptedSources = parsed.AcceptedSources

	// Logger
	if parsed.Log != "" {
		file, err := os.OpenFile(parsed.Log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, []error{fmt.Errorf("unable to open file %s", parsed.Log)}
		}
		result.Logger = logger.New(file)
	}

	// Directions
	for _, jsonDir := range parsed.Directions {
		// Create the direction according to its type
		dir := hitcounter.Direction{
			Store:       store.NewShardMap(int64(jsonDir.MaxTracked)),
			Name:        jsonDir.Name,
			CleanUpTime: jsonDir.CleanUpTime,
			MaxHits:     jsonDir.MaxHits,
			WindowSize:  jsonDir.WindowSize,
		}
		dir.Store.Type = jsonDir.Typ

		// Add it to the list
		result.Directions = append(result.Directions, dir)
	}

	return result, nil
}
