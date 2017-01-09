package commands

import "strings"
import "strconv"

type Cmd struct {
	No   int
	Name string
	Args [] string
}

func isSpace (c byte) bool {
	switch c {
	case  ' ': fallthrough
	case '\t': fallthrough
	case '\r': fallthrough
	case '\n': return true
	default: return false
	}
}

func readWords (line string) []string {
	single := func (c byte) string { return string([]byte{c}) }
	at := func (i int) byte { 
		if i == len(line) { return '\000' } else { return line[i] }
	}
	words := make ([]string, 0, 10)
	pos := 0
	for {
		for isSpace (at(pos)) { pos++ }
		if pos == len(line) { return words }
		quoted := false
		word := ""
		for pos != len(line) && (!isSpace(line[pos]) || quoted) {
			switch line[pos] {
			case '"': quoted = !quoted
			case '\\': 
				if pos+1 != len(line) { pos++ }
				switch(at(pos)) {
				case 'n': word += "\n"
			case 't': word += "\t"
			case 'r': word += "\r"
				case 'b': word += "\b"
				default: word += single(line[pos])
				}
			default: word += single(line[pos])
			}
			pos++
		}
		words = append (words, word)
	}
}

func ParseCmd (line string) Cmd {
	args := readWords(line)
	res := Cmd{0, args[0], args[1:]}
	if len (res.Args) > 2 && res.Args[0] == "-n" {
		res.No, _ = strconv.Atoi(res.Args[1])
		res.Args = res.Args[2:]
	}
	return res
}

func (cmd Cmd) Repr () string {
	if cmd.No != 0 {
		return cmd.Name + " -n " + strconv.Itoa (cmd.No) + " " + strings.Join(cmd.Args, " ")
	}
	return cmd.Name + " " + strings.Join(cmd.Args, " ")
}
