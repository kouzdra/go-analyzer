package options

import "os"
import "fmt"

var Codeblocks = false
var  InLogName = "/tmp/go-server-in"
var OutLogName = "/tmp/go-server-out"

func init () {
        i := 1
        for i != len (os.Args) {
                switch os.Args[i] {
                case "--codeblocks": Codeblocks = true; i++
                case "--inLogName" : i++; if i != len(os.Args) {  InLogName = os.Args[i]; i ++}
                case "--outLogName": i++; if i != len(os.Args) { OutLogName = os.Args[i]; i ++}
                default: fmt.Fprintf(os.Stderr, "unknown option: `%s'\n", os.Args[i]); os.Exit(1)
                }
        }
}
