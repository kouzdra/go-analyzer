package names

//import "fmt"

const (
        hashSize = 64*1024
)

type NameNo uint
type Name struct {
        Name string
	No   NameNo
        Hash uint
        Next *Name
}

var cnt uint = 0

func (n *Name) Gbl () bool {
    return len(n.Name) != 0 && 'A' <= n.Name[0] && n.Name[0] <= 'Z'
}

var namesTab [hashSize]*Name
var names = make ([]Name, 0, 1024)

func (n *Name) Repr () string {
	return n.Name
}

func hash(name string) uint {
        res := uint (0)
        for i := 0; i != len(name); i++ { res = uint (res << 1) + uint(name[i]) }
        return res;
}

func Find (name string) *Name {
        return FindHash(name, hash(name))
}

func FindHash (name string, hash uint) *Name {
        for cell := namesTab[hash%uint (hashSize)]; cell != nil; cell = cell.Next {
                if (cell.Name == name) {
                        return cell
                }
        }
        return nil
}

func Put(name string) *Name {
        //fmt.Printf("Req: %s\n", name)
        hash := hash(name)
        cell := FindHash(name, hash)
        if cell == nil {
		hh := hash%uint (hashSize)
		cnt ++
                newCell := Name {name, NameNo (cnt), hash, namesTab[hh]}
		names = append (names, newCell)
                //fmt.Printf("New: %s\n", name)
		cell = &newCell
                namesTab[hh] = cell
        }
        return cell
}

