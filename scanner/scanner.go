package scanner

import "os"
import "fmt"
import "bufio"
import "strings"
import "commands"
import "options"

type Scanner struct {
	*bufio.Scanner
	Log *bufio.Writer
}

func New (f *os.File) *Scanner {
	log, err := os.Create(options.InLogName)
	if err != nil {
		fmt.Println(err)
	}
	return &Scanner{bufio.NewScanner(f), bufio.NewWriter(log)}}

func (s *Scanner) ReadCmd () commands.Cmd {
	if s.Scan() {
		line := s.Text()
		s.Log.WriteString(line)
		s.Log.WriteString("\n")
		s.Log.Flush()
		line = strings.TrimSpace (line)
		if len(line) == 0 || line[0] == '#' { return s.ReadCmd() }
		return commands.ParseCmd(s.Text())
	}
	return commands.ParseCmd("quit")
}

