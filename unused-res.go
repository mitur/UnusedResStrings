package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const testDir string = "/Users/mitur/gits/TortugaSwift/tortuga_ios"

func main() {
	searcher := NewResSearcher()
	err := searcher.FindExistingStrings(testDir)
	if err != nil {
		fmt.Println("Failed to find existing strings:", err)
		return
	}

	fmt.Println("Existing strings:", len(searcher.ExistingStrings))
	searcher.SearchDir(testDir)
	// searcher.PrintResults()

	unusedStrings := searcher.GetUnusedStrings()

	for _, s := range unusedStrings {
		fmt.Println(s)
	}

	fmt.Println("unused Strings:", len(unusedStrings))
}

type Occurence struct {
	Filename  string
	Line      int
	Character int
}

type ResSearcher struct {
	FoundStrings    map[string]*[]Occurence
	ExistingStrings []string
}

func NewResSearcher() *ResSearcher {
	return &ResSearcher{
		FoundStrings:    map[string]*[]Occurence{},
		ExistingStrings: []string{},
	}
}

func (rs *ResSearcher) TotalFoundLocStrings() int {
	return len(rs.FoundStrings)
}

func (rs *ResSearcher) GetUnusedStrings() []string {
	res := []string{}
	for _, existing := range rs.ExistingStrings {
		if _, exists := rs.FoundStrings[existing]; !exists {
			res = append(res, existing)
		}
	}

	return res

}

func (rs *ResSearcher) PrintResults() {
	for key, occLst := range rs.FoundStrings {
		fmt.Printf("%s, %d Occurences\n", key, len(*occLst))

		for _, occ := range *occLst {
			_, filename := filepath.Split(occ.Filename)
			fmt.Printf("\t%s:%d\n", filename, occ.Line)
		}

		fmt.Printf("\n")
	}

	fmt.Printf("Loc strings used: %d, existing: %d\n", len(rs.FoundStrings), len(rs.ExistingStrings))
}

func (rs *ResSearcher) AddOccurence(locStr, filePath string, lineNumber int) {
	var occurences *[]Occurence
	if lst, exists := rs.FoundStrings[locStr]; exists {
		occurences = lst
	} else {
		occurences = &[]Occurence{}
		rs.FoundStrings[locStr] = occurences
	}
	_, filename := filepath.Split(filePath)

	*occurences = append(*occurences, Occurence{
		Filename:  filename,
		Line:      lineNumber,
		Character: 0,
	})
}

func (rs *ResSearcher) SearchDir(dirPath string) error {
	const filetype string = ".swift"
	files, err := ioutil.ReadDir(dirPath)

	if err != nil {
		return err
	}

	for _, f := range files {
		name := f.Name()
		if len(filetype) < len(name) && name[len(name)-len(filetype):] == filetype {
			if err := rs.SearchFile(filepath.Join(dirPath, name)); err != nil {
				return fmt.Errorf("Error when searching file %s: %s", name, err)
			}
		}
	}

	return nil
}

func (rs *ResSearcher) FindExistingStrings(dirPath string) error {
	locFile := filepath.Join(dirPath, "en.lproj", "Localizable.strings")

	f, err := os.Open(locFile)
	if err != nil {
		return fmt.Errorf("Failed to Localizable.strings in %s: %s", locFile, err)

	}

	r := bufio.NewReader(f)
	for true {
		r.ReadSlice('"')
		// Read to next " and keep it!
		line, err := r.ReadSlice('"')

		if err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("Error while reading existing strings: %s", err)
		}

		rs.ExistingStrings = append(rs.ExistingStrings, string(line[:len(line)-1]))

		// Skip to ;

		r.ReadSlice(';')
	}

	return nil

}

func (rs *ResSearcher) SearchFile(filepath string) error {
	const tok string = ".loc()"
	var searched = []byte(tok)
	//var buf []byte = make([]byte, 1024, 1024)
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}

	r := bufio.NewReader(f)
	lineNumber := 0

	for true {
		//n, err := r.Read(buf)
		lineNumber += 1
		buf, _, err := r.ReadLine()
		n := len(buf)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("Failed to read batch: %s", err)
		}

		l := len(searched)
		for i := 0; i+l < n; i++ {
			if buf[i+l-1] == searched[l-1] {
				ok := false
				for j := 2; j <= l; j++ {
					if buf[i+l-j] != searched[l-j] {
						ok = false
						break
					}
					ok = true
				}

				if ok {
					// We know that i to i+l is ".loc()"
					// i-1 have to be ", going back from there to find next "
					// will yield the LOC string.
					foundString := findLocString(buf, i)

					rs.AddOccurence(foundString, filepath, lineNumber)

				}
			}
		}

	}
	return nil
}

func findLocString(buf []byte, locIndex int) string {
	const delim byte = '"'

	ed := locIndex - 1 // ending delim

	i := 0
	for true {
		if buf[ed-i-1] == delim {
			break
		}
		i += 1
	}
	return string(buf[ed-i : ed])
}

/*










 */
