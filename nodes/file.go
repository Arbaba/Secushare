package nodes

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/Arbaba/Peerster/packets"
	"io"
	"os"
	"time"
	//errors"
)
//Struct used to store file infos 
type FileMetaData struct {
	FileName string
	FileSize uint32
	MetaFile [][sha256.Size]byte
	ChunksOwners map[int][]string
	MatchedPeers []string //peers who have the whole file 
}



func check(e error) {
	if e != nil {
		panic(e)
	}
}


func (gossiper *Gossiper) ScanFile(filename string) {

	f, err := os.Open(fmt.Sprintf("_SharedFiles/%s", filename))
	check(err)
	reader := bufio.NewReader(f)
	var metaFile [][sha256.Size]byte
	for {
		buffer := make([]byte, 1<<13)

		nbytes, err := reader.Read(buffer)

		if err == io.EOF {
			break
		}
		check(err)
		sum := sha256.Sum256(buffer[:nbytes])

		gossiper.FilesMux.Lock()
		//fmt.Println("hex to str ", hex.EncodeToString(sum[:]))
		gossiper.Files[hex.EncodeToString(sum[:])] = buffer[:nbytes]
		gossiper.FilesMux.Unlock()
		metaFile = append(metaFile, sum)
		
	}
		var tmp []byte
		for _, chunk := range metaFile {
			tmp = append(tmp, chunk[:]...)
		}
		metaHash := sha256.Sum256(tmp[:])
		stat, err := f.Stat()
		check(err)
		fileSize := uint32(stat.Size())
		file := FileMetaData{FileName:filename, FileSize:fileSize,MetaFile: metaFile,}
		gossiper.FilesInfoMux.Lock()
		gossiper.FilesInfo[hex.EncodeToString(metaHash[:])] = &file
		gossiper.FilesInfoMux.Unlock()
		fmt.Printf("filename: %s\nmetaHash: %x\nfilesize: %d\n", filename, metaHash, fileSize)
	}

	func HexToString(hexrepr []byte) string {
		return hex.EncodeToString(hexrepr)
	}

	func (gossiper *Gossiper) CreateDataRequest(destination string, hashvalue []byte) packets.DataRequest {
		return packets.DataRequest{Origin: gossiper.Name,
			Destination: destination,
			HopLimit:    gossiper.HOPLIMIT,
			HashValue:   hashvalue,
		}
	}
	func (gossiper *Gossiper) DownloadMetaFile(metahash, destination, filename string) (packets.DataReply, *FileMetaData) {
		decoded, _ := hex.DecodeString(metahash)
		request := gossiper.CreateDataRequest(destination, decoded)
		go gossiper.SendDataRequest(request)
		fmt.Printf("DOWNLOADING metafile of %s from %s\n", filename, destination)
		metafileReceiver := make(chan packets.DataReply)
		gossiper.DataBufferMux.Lock()
		gossiper.DataBuffer[metahash] = &metafileReceiver
		gossiper.DataBufferMux.Unlock()
		ticker := time.NewTicker(time.Second * time.Duration(5))
		defer ticker.Stop()

		for {
			select {
			case dataReply := <-metafileReceiver:
				metaFile := make([][sha256.Size]byte, len(dataReply.Data)/sha256.Size)
				fmt.Println("-- METAFILE PREPARED --")
				gossiper.FilesInfoMux.Lock()
				for chunkNb := 0; chunkNb < len(dataReply.Data)/sha256.Size; chunkNb++ {
					nextChunkidx := (chunkNb + 1) * 32
					//fmt.Println("Chunknb ", chunkNb, "nextchunkidx", nextChunkidx, len(metaFile), len(dataReply.Data))
					copy(metaFile[chunkNb][:], dataReply.Data[chunkNb*32:nextChunkidx])
					//fmt.Println("Verify ",HexToString(metaFile[chunkNb][:]), HexToString(dataReply.Data[chunkNb:nextChunkidx]))
				}
				fileMetaData := FileMetaData{FileName: filename, MetaFile: metaFile}

				gossiper.FilesInfo[metahash] = &fileMetaData
				gossiper.FilesInfoMux.Unlock()
				return dataReply, &fileMetaData
			case <-ticker.C:
				gossiper.SendDataRequest(request)
			}

		}
	}

	func (gossiper *Gossiper) DownloadFile(metafileReply packets.DataReply, fileMetaData *FileMetaData) {
		var fileData []byte
		for chunkNb := 0; chunkNb < len(metafileReply.Data)/32; chunkNb++ {
			fmt.Printf("DOWNLOADING %s chunk %d from %s\n", fileMetaData.FileName, chunkNb, metafileReply.Origin)

		nextChunkidx := (chunkNb + 1) * 32
		request := packets.DataRequest{Origin: gossiper.Name,
			Destination: metafileReply.Origin,
			HopLimit:    gossiper.HOPLIMIT,
			HashValue:   metafileReply.Data[chunkNb*32 : nextChunkidx],
		}

		go gossiper.SendDataRequest(request)
		ticker := time.NewTicker(time.Second * time.Duration(5))
		defer ticker.Stop()

		chunkReceiver := make(chan packets.DataReply)
		gossiper.DataBufferMux.Lock()
		chunkHash := HexToString(fileMetaData.MetaFile[chunkNb][:])
		gossiper.DataBuffer[chunkHash] = &chunkReceiver
		gossiper.DataBufferMux.Unlock()

		func () {
			for {

				select {
				case chunkReply := <-chunkReceiver:
					if CheckChunk(chunkReply.Data, chunkNb, fileMetaData) {
						gossiper.FilesMux.Lock()
						gossiper.Files[chunkHash] =  chunkReply.Data[:]
						fileData = append(fileData, chunkReply.Data[:]...)
						gossiper.FilesMux.Unlock()
						return
					} else if len(chunkReply.Data) == 0{
						fmt.Println("Empty chunk download")
						return
						//gossiper.SendDataRequest(request)
					}
				case <-ticker.C:
					gossiper.SendDataRequest(request)
				}
			}
		}()

	}
	//fmt.Printf(" file data %x", fileData)
	fileMetaData.FileSize = uint32(len(fileData))
	gossiper.StoreFile(fileData, fileMetaData)
}

func (gossiper *Gossiper) FindFileInfo(filename string) (*FileMetaData, string, bool){
	gossiper.FilesInfoMux.Lock()
	defer gossiper.FilesInfoMux.Unlock()
	for metahash, fileInfo := range gossiper.FilesInfo{
		if fileInfo.FileName == filename{
			return fileInfo, metahash, true
		}
	}
	return nil, "", false

}

func (gossiper * Gossiper) GetChunk(hash string) ([]byte, bool){
	gossiper.FilesMux.Lock()
	defer gossiper.FilesMux.Unlock()
	data, found := gossiper.Files[hash] 
	return data, found	
}

// TODO: Merge or modularize download functions when you have enough time,
func (gossiper *Gossiper) DownloadFoundFile(filename string) {
	
	var fileData []byte
	fileMetaData, metahash,  found := gossiper.FindFileInfo(filename)
	
	if !found{
		fmt.Println("Metafile for ",filename ,"not found")
		return
	}
	//metafileData, found := gossiper.GetChunk(metahash)
	fileLocations := gossiper.Matches.FindLocations(filename)
	if len(fileLocations) == 0 {
		fmt.Println("No location found for ", filename)
		return
	} 
	fmt.Println("metafile data ", len(fileMetaData.MetaFile), metahash)
	for chunkNb := 0; chunkNb < len(fileMetaData.MetaFile); chunkNb++ {
		for _, location := range fileLocations{
			fmt.Printf("DOWNLOADING %s chunk %d from %s\n", fileMetaData.FileName, chunkNb, location)

			request := packets.DataRequest{Origin: gossiper.Name,
				Destination: location,
				HopLimit:    gossiper.HOPLIMIT,
				HashValue:   fileMetaData.MetaFile[chunkNb][:],
			}

			go gossiper.SendDataRequest(request)
			ticker := time.NewTicker(time.Second * time.Duration(5))
			defer ticker.Stop()

			chunkReceiver := make(chan packets.DataReply)
			gossiper.DataBufferMux.Lock()
			chunkHash := HexToString(fileMetaData.MetaFile[chunkNb][:])
			gossiper.DataBuffer[chunkHash] = &chunkReceiver
			gossiper.DataBufferMux.Unlock()

			success := func () bool {
				for {

					select {
					case chunkReply := <-chunkReceiver:
						if CheckChunk(chunkReply.Data, chunkNb, fileMetaData) {
							gossiper.FilesMux.Lock()
							gossiper.Files[chunkHash] =  chunkReply.Data[:]
							fileData = append(fileData, chunkReply.Data[:]...)
							gossiper.FilesMux.Unlock()
							return true
						} else if len(chunkReply.Data) == 0{
							fmt.Println("Empty chunk download")
							return false
							//gossiper.SendDataRequest(request)
						}
					case <-ticker.C:
						gossiper.SendDataRequest(request)
					}
				}
			}()
			if success {break}
		}

}
//fmt.Printf(" file data %x", fileData)
fileMetaData.FileSize = uint32(len(fileData))
gossiper.StoreFile(fileData, fileMetaData)

}
func (gossiper *Gossiper) StoreFile(data []byte, filemetadata *FileMetaData) {
	f, err := os.Create(fmt.Sprintf("_Downloads/%s", filemetadata.FileName))
	check(err)
	defer f.Close()
	writer := bufio.NewWriter(f)
	writer.Write(data)
	writer.Flush()
	fmt.Printf("Reconstructed file %s\n", filemetadata.FileName)
}
func CheckChunk(data []byte, chunknb int, meta *FileMetaData) bool {
	sum :=sha256.Sum256(data)
	//fmt.Println(HexToString(sum[:]), HexToString(meta.MetaFile[chunknb][:]))
	return HexToString(sum[:]) == HexToString(meta.MetaFile[chunknb][:])
}

func (gossiper *Gossiper) SendDataRequest(request packets.DataRequest) {
	pkt := packets.GossipPacket{DataRequest: &request}
	gossiper.SendDirect(pkt, request.Destination)
}

