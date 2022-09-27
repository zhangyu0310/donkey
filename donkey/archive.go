package donkey

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

var (
	ErrArchiveIncomplete = errors.New("archive file is incomplete")
	ErrReadEndOfFile     = errors.New("read end of archive file")
)

type Entry struct {
	Id        uint64
	Uuid      string
	ExtraUuid []string
}

type Archive struct {
	Id          int
	archive     *os.File
	readOffset  int64
	writeOffset int64
	buffer      []byte
	EntityNum   uint64
}

func NewArchive(routineId int) (*Archive, error) {
	idStr := strconv.Itoa(routineId)
	fileName := "donkey_archive_" + idStr
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Open archive %s failed, err: %s\n", fileName, err)
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		fmt.Printf("Get archive file %s stat failed, err: %s\n", fileName, err)
		return nil, err
	}
	archive := &Archive{
		Id:          routineId,
		archive:     f,
		readOffset:  0,
		writeOffset: stat.Size(),
		buffer:      make([]byte, 0, 10240),
		EntityNum:   0,
	}
	return archive, nil
}

func (archive *Archive) SeekForAppend() error {
	_, err := archive.archive.Seek(archive.writeOffset, 0)
	if err != nil {
		fmt.Println("Append entry archive file seek failed, err:", err)
		return err
	}
	return nil
}

func (archive *Archive) AppendEntry(entry *Entry) error {
	data := make([]byte, 0, 32)
	data = append(data, EncodeVarUint64(entry.Id)...)
	data = append(data, EncodeVarUint64(uint64(len(entry.Uuid)))...)
	data = append(data, []byte(entry.Uuid)...)
	for i := 0; i < len(entry.ExtraUuid); i++ {
		data = append(data, EncodeVarUint64(uint64(len(entry.ExtraUuid[i])))...)
		data = append(data, []byte(entry.ExtraUuid[i])...)
	}

	n, err := archive.archive.Write(data)
	if err != nil {
		fmt.Printf("Append entry to archive failed, err: (%s)\n", err)
		return err
	}
	archive.writeOffset += int64(n)
	archive.EntityNum++
	return nil
}

func (archive *Archive) Flush() {
	_ = archive.archive.Sync()
}

func (archive *Archive) readSomeData() error {
	data := make([]byte, 10240)
	n, err := archive.archive.Read(data)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Get entry read failed, err:", err)
			return err
		}
	}
	archive.readOffset += int64(n)
	archive.buffer = append(archive.buffer, data[:n]...)
	return err
}

func (archive *Archive) getVarIntFromArchive() (VarUint64, error) {
	eof := false
	for {
		varInt, index, err := GetVarUint64(archive.buffer, 0)
		if err != nil {
			if eof {
				fmt.Println("Archive is eof but get var int not enough.")
				return nil, ErrArchiveIncomplete
			} else {
				err = archive.readSomeData()
				if err != nil {
					if err == io.EOF {
						eof = true
					} else {
						fmt.Println("Read archive file failed, err:", err)
						return nil, err
					}
				}
			}
		} else {
			archive.buffer = archive.buffer[index:]
			return varInt, nil
		}
	}
}

func (archive *Archive) getDataFromArchive(dataLen int) ([]byte, error) {
	eof := false
	for {
		if len(archive.buffer) < dataLen {
			if eof {
				return nil, ErrArchiveIncomplete
			}
			err := archive.readSomeData()
			if err != nil {
				if err != io.EOF {
					return nil, err
				} else {
					eof = true
				}
			}
		} else {
			data := archive.buffer[:dataLen]
			archive.buffer = archive.buffer[dataLen:]
			return data, nil
		}
	}
}

func getUuidFromArchive(archive *Archive) ([]byte, error) {
	// Get data length
	dataLenVarInt, err := archive.getVarIntFromArchive()
	if err != nil {
		fmt.Println("Get var int from archive failed, err:", err)
		return nil, err
	}
	dataLen := DecodeVarUint64(dataLenVarInt)
	// Get data
	data, err := archive.getDataFromArchive(int(dataLen))
	if err != nil {
		fmt.Println("Get data from archive failed, err:", err)
		return nil, err
	}
	return data, nil
}

func (archive *Archive) GetOneEntry(extraNum uint) (*Entry, error) {
	if archive.readOffset == archive.writeOffset && len(archive.buffer) == 0 {
		return nil, ErrReadEndOfFile
	}
	_, err := archive.archive.Seek(archive.readOffset, 0)
	if err != nil {
		fmt.Println("Get entry seek failed, err:", err)
		return nil, err
	}
	// Get id
	idVarInt, err := archive.getVarIntFromArchive()
	if err != nil {
		fmt.Println("Get var int from archive failed, err:", err)
		return nil, err
	}
	id := DecodeVarUint64(idVarInt)
	data, err := getUuidFromArchive(archive)
	if err != nil {
		fmt.Println("Get uuid from archive failed, err:", err)
		return nil, err
	}
	// Get extra
	extraUuid := make([]string, 0, extraNum)
	for i := uint(0); i < extraNum; i++ {
		data, err := getUuidFromArchive(archive)
		if err != nil {
			fmt.Println("Get uuid from archive failed, err:", err)
			return nil, err
		}
		extraUuid = append(extraUuid, string(data))
	}

	entry := &Entry{
		Id:        id,
		Uuid:      string(data),
		ExtraUuid: extraUuid,
	}
	return entry, nil
}
