// Copyright 2015-2018 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"

	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/go-nats"

	gnatsd "github.com/nats-io/gnatsd/test"
)

////////////////////////////////////////////////////////////////////////////////
// Creating client connections
////////////////////////////////////////////////////////////////////////////////

// NewDefaultConnection opens a default connection
func NewDefaultConnection() (*nats.Conn, error) {
	return NewConnection(nats.DefaultPort)
}

// NewConnection forms connection on a given port.
func NewConnection(port int) (*nats.Conn, error) {
	url := fmt.Sprintf("nats://localhost:%d", port)
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return nc, err
}

// NewEConn opens an encoded connection
func NewEConn() (*nats.EncodedConn, error) {
	c, err := NewDefaultConnection()
	if err != nil {
		return nil, err
	}
	ec, err := nats.NewEncodedConn(c, nats.DEFAULT_ENCODER)
	if err != nil {
		return nil, err
	}
	return ec, nil
}

////////////////////////////////////////////////////////////////////////////////
// Running gnatsd server in separate Go routines
////////////////////////////////////////////////////////////////////////////////

// RunDefaultServer will run a server on the default port.
func RunDefaultServer() *server.Server {
	return RunServerOnPort(nats.DefaultPort)
}

// RunServerOnPort will run a server on the given port.
func RunServerOnPort(port int) *server.Server {
	opts := gnatsd.DefaultTestOptions
	opts.Port = port
	return RunServerWithOptions(opts)
}

// RunServerWithOptions will run a server with the given options.
func RunServerWithOptions(opts server.Options) *server.Server {
	return gnatsd.RunServer(&opts)
}

// RunServerWithConfig will run a server with the given configuration file.
func RunServerWithConfig(configFile string) (*server.Server, *server.Options) {
	return gnatsd.RunServerWithConfig(configFile)
}
