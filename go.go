package main

import "github.com/kouzdra/go-analyzer/server"
import _ "github.com/kouzdra/go-analyzer/paths"
import _ "github.com/kouzdra/go-analyzer/tab/sym"

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
