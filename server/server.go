package server

import "os"
import "fmt"
import "strings"
import "github.com/kouzdra/go-analyzer/results"
import "github.com/kouzdra/go-analyzer/writer"
import "github.com/kouzdra/go-analyzer/scanner"
import "strconv"
import "github.com/kouzdra/go-analyzer/commands"

type Server struct {
	*scanner.Scanner
	*writer .Writer
	*Project
}

func NewServer () * Server {
	server := &Server{scanner.New(os.Stdin),writer.New(os.Stdout), nil}
	server.Project = NewProject(server)
	return server
}

func (s *Server) Msg (msg string) {
	results.Message{msg}.Write(s.Writer)
}

func (s *Server) MsgF (f string, args... interface{}) {
	results.Message{fmt.Sprintf (f, args...)}.Write(s.Writer)
}

func (s *Server) Run () {
	for {
		s.Process (s.ReadCmd ())
	}
}

func (s *Server) Process (cmd commands.Cmd) {
	chk := func (min int) bool {
		if len(cmd.Args) < min {
			s.MsgF("min %d args required in cmd \"%s\"", min, cmd.Repr())
			return false
		}
		return true
	}
	switch cmd.Name {
	case "go-path": if chk(1) { s.Project.Project.SetPath (cmd.Args [0]) }
	case "go-root": if chk(1) { s.Project.Project.SetRoot (cmd.Args [0]) }
	case "load-project": {
		s.Project.Project.Load ()
		s.Msg(fmt.Sprintf ("%d dirs found", len (s.Project.Project.Dirs)))
		//s.Msg(fmt.Sprintf ("%d pkgs found", len (s.Project.Pkgs)))
	}
	case "reload": if chk (1) {
		if src, err := s.Project.Project.GetSrc(cmd.Args[0]); src != nil {
			src.Reload()
		} else {
			s.Msg(err.Error())
		}
	}
	case "changed": if chk(4) {
		if src, err := s.Project.Project.GetSrc(cmd.Args[0]); src != nil {
			beg, _ := strconv.Atoi(cmd.Args[2])
			end, _ := strconv.Atoi(cmd.Args[3])
			src.Changed(beg, end, cmd.Args[1])
		} else {
			s.Msg(err.Error())
		}
	}
	case "analyze": if chk(1) {
		s.Project.Analyze(cmd.No, cmd.Args[0])
	}
	case "complete": if chk(2) {
		pos, err := strconv.Atoi(cmd.Args[1])
		if err == nil {
			s.Project.Complete(cmd.No, pos)
		}
	}
	case "tooltip-info": if chk(1) {
		_, err := strconv.Atoi(cmd.Args[1])
		if err == nil {
			// TODO: tooltips
		}
	}
	case "find-files": if chk(3) {
		max, _ := strconv.Atoi(cmd.Args[2])
		s.Project.FindFiles(cmd.No, cmd.Args[0], cmd.Args[1] == "system", max)
	}
	case "file-in-project": if chk(1) {
		if strings.HasSuffix(cmd.Args[0], ".go") {
			s.Writer.Beg("SET-FILE-MODE").Write(cmd.Args[0]).Eol()
			s.Writer.End("SET-FILE-MODE")
		} else {
			s.Writer.Beg("DROP-FILE-MODE").Write(cmd.Args[0]).Eol()
			s.Writer.End("DROP-FILE-MODE")
		}
	}
	case "quit": os.Exit (0)
	default: s.Msg(fmt.Sprintf ("<UnknownCommand: %s>", cmd.Repr ()))
	}
	//results.PlainText{fmt.Sprintf ("<OK: %s>\n", cmd.Repr ())}.Write(s.Writer)
}
