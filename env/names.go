package env

//import "fmt"

const (
        HashSize = 64
)

var Dummy = Put("_")
var Null  = Put("<NONE>")

type Name struct {
        Name string
        Hash int
        Next * Name
}

func (n *Name) Gbl () bool {
    return len(n.Name) != 0 && 'A' <= n.Name[0] && n.Name[0] <= 'Z'
}

var Names [1024*128]*Name

func Hash(name string) int {
        res := 0
        for i := 0; i != len(name); i++ { res = (res << 1) + int(name[i]) }
        return res;
}

func Find (name string) *Name {
        return FindHash(name, Hash(name))
}

func FindHash (name string, hash int) *Name {
        for cell := Names[hash%len(Names)]; cell != nil; cell = cell.Next {
                if (cell.Name == name) {
                        return cell
                }
        }
        return nil
}

func Put(name string) *Name {
        //fmt.Printf("Req: %s\n", name)
        hash := Hash(name)
        cell := FindHash(name, hash)
        if cell == nil {
                cell = &Name {name, hash, Names[hash%len(Names)]}
                //fmt.Printf("New: %s\n", name)
                Names[hash%len(Names)] = cell
        }
        return cell
}

