// Obdi - a REST interface and GUI for deploying software
// Copyright (C) 2014  Mark Clarkson
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"
	"os"
)

type PostedData struct {
	Hostname string
}

// ***************************************************************************
// GO RPC PLUGIN
// ***************************************************************************

func (t *Plugin) GetRequest(args *Args, response *[]byte) error {

	// List all keys

	if len(args.QueryString["env_id"]) == 0 {
		ReturnError("'env_id' must be set", response)
		return nil
	}

	sa := ScriptArgs{
		ScriptName: "saltkey-showkeys.sh",
		CmdArgs:    "",
		EnvVars:    "",
		EnvCapDesc: "SALT_WORKER",
		Type:       2,
	}

	var jobid int64
	var err error
	if jobid, err = t.RunScript(args, sa, response); err != nil {
		// RunScript wrote the error
		return nil
	}

	reply := Reply{jobid, "", SUCCESS, ""}
	jsondata, err := json.Marshal(reply)
	if err != nil {
		ReturnError("Marshal error: "+err.Error(), response)
		return nil
	}
	*response = jsondata

	return nil
}

func (t *Plugin) PostRequest(args *Args, response *[]byte) error {

	// Check for required query string entries

	if len(args.QueryString["hostname"]) == 0 {
		ReturnError("'type' must be set", response)
		return nil
	}

	if len(args.QueryString["type"]) == 0 {
		ReturnError("'type' must be set", response)
		return nil
	}

	if len(args.QueryString["env_id"]) == 0 {
		ReturnError("'env_id' must be set", response)
		return nil
	}

	// Get the ScriptId from the scripts table for:
	scriptName := ""
	if args.QueryString["type"][0] == "accept" {
		scriptName = "saltkey-acceptkeys.sh"
	} else {
		scriptName = "saltkey-rejectkeys.sh"
	}

	var jobid int64

	sa := ScriptArgs{
		ScriptName: scriptName,
		CmdArgs:    args.QueryString["hostname"][0],
		EnvVars:    "",
		EnvCapDesc: "SALT_WORKER",
		Type:       2,
	}

	var err error
	if jobid, err = t.RunScript(args, sa, response); err != nil {
		// RunScript wrote the error
		return nil
	}

	reply := Reply{jobid, "", SUCCESS, ""}
	jsondata, err := json.Marshal(reply)
	if err != nil {
		ReturnError("Marshal error: "+err.Error(), response)
		return nil
	}
	*response = jsondata

	return nil
}

func (t *Plugin) DeleteRequest(args *Args, response *[]byte) error {

	// Check for required query string entries

	if len(args.QueryString["env_id"]) == 0 {
		ReturnError("'env_id' must be set", response)
		return nil
	}

	var jobid int64

	sa := ScriptArgs{
		ScriptName: "saltkey-deletekeys.sh",
		CmdArgs:    args.PathParams["id"],
		EnvVars:    "",
		EnvCapDesc: "SALT_WORKER",
		Type:       2,
	}

	var err error
	if jobid, err = t.RunScript(args, sa, response); err != nil {
		// RunScript wrote the error
		return nil
	}

	reply := Reply{jobid, "", SUCCESS, ""}
	jsondata, err := json.Marshal(reply)
	if err != nil {
		ReturnError("Marshal error: "+err.Error(), response)
		return nil
	}
	*response = jsondata

	return nil
}

func (t *Plugin) HandleRequest(args *Args, response *[]byte) error {

	// All plugins must have this.

	if len(args.QueryType) > 0 {
		switch args.QueryType {
		case "GET":
			t.GetRequest(args, response)
			return nil
		case "POST":
			t.PostRequest(args, response)
			return nil
		case "DELETE":
			t.DeleteRequest(args, response)
			return nil
		}
		ReturnError("Internal error: Invalid HTTP request type for this plugin "+
			args.QueryType, response)
		return nil
	} else {
		ReturnError("Internal error: HTTP request type was not set", response)
		return nil
	}
}

func main() {

	//logit("Plugin starting")

	plugin := new(Plugin)
	rpc.Register(plugin)

	listener, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		txt := fmt.Sprintf("Listen error. ", err)
		logit(txt)
	}

	//logit("Plugin listening on port " + os.Args[1])

	if conn, err := listener.Accept(); err != nil {
		txt := fmt.Sprintf("Accept error. ", err)
		logit(txt)
	} else {
		//logit("New connection established")
		rpc.ServeConn(conn)
	}
}
