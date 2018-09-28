package main

import "github.com/kouzdra/go-analyzer/server"
//import "os"
//import "log"
//import "runtime"
//import "runtime/pprof"
//import "github.com/kouzdra/go-analyzer/env"
import "github.com/kouzdra/go-analyzer/names"
import _ "github.com/kouzdra/go-analyzer/paths"
import _ "github.com/kouzdra/go-analyzer/tab/sym"
//import "flag"

var f_main = names.N_main

func main() {
	/*f, err := os.Create("go.prof")
	if err != nil {
	    log.Fatal(err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
	    log.Fatal(err)
	}
        defer pprof.StopCPUProfile ()*/
	server := server.NewServer ()
	server.Run ()
}
