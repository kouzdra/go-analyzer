package main

import "server"
//import "os"
//import "log"
//import "runtime"
//import "runtime/pprof"
import "env"
//import "flag"

var f_main = env.Put("main")

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
