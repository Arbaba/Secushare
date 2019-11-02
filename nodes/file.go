package nodes

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

type FileMetadata struct {
	FileName string
	FileSize int64
	MetaFile [][sha256.Size]byte
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func (gossiper *Gossiper) ScanFile(filename string) {

	buffer := make([]byte, 1<<13)
	f, err := os.Open(fmt.Sprintf("_SharedFiles/%s", filename))
	check(err)
	reader := bufio.NewReader(f)
	var metaFile [][sha256.Size]byte
	for {
		nbytes, err := reader.Read(buffer)

		if err == io.EOF {
			break
		}
		check(err)

		sum := sha256.Sum256(buffer[:nbytes])
		metaFile = append(metaFile, sum)
	}
	var tmp []byte
	for _, chunk := range metaFile {
		tmp = append(tmp, chunk[:]...)
	}
	metaHash := sha256.Sum256(tmp[:])
	stat, err := f.Stat()
	check(err)
	fileSize := stat.Size()
	file := FileMetadata{filename, fileSize, metaFile}
	gossiper.FilesMux.Lock()
	defer gossiper.FilesMux.Unlock()
	gossiper.Files[fmt.Sprintf("%x", metaHash)] = file
	fmt.Printf("filename: %s\nmetaFile: %x\nmetaHash: %x\nfilesize: %d\n", filename, metaFile, metaHash, fileSize)

}
