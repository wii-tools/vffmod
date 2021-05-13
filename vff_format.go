package vffmod

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

var (
	// VFFMagic is recognized to be 4 bytes: the literal "VFF", and a space.
	VFFMagic = [4]byte{0x56, 0x46, 0x46, 0x20}
)

// VFFHeader allows us to keep track of the
type VFFHeader struct {
	Magic       [4]byte
	Unknown     uint32
	VolumeSize  uint32
	ClusterSize uint16
}

// OpenVFF reads a VFF at the given path, returning
// a usable filesystem representation.
func OpenVFF(path string) (*VFFFS, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ReadVFF(contents)
}

// ReadVFF interprets a VFF from passed bytes, returning
// a usable filesystem representation.
func ReadVFF(contents []byte) (*VFFFS, error) {
	// We can immediately reject smaller files.
	// The maximum size comes from https://superuser.com/a/74399
	// The author of this code has questions on where that number came from.
	if 0x419999 > len(contents) {
		return nil, ErrInvalidFormat
	}

	var header VFFHeader
	err := binary.Read(bytes.NewBuffer(contents), binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	if VFFMagic != header.Magic {
		return nil, ErrInvalidMagic
	}

	fs := VFFFS{
		fileData: contents,
	}

	// Multiply by 16. Apparently, this is necessary.
	fs.volumeSize = header.VolumeSize
	fs.clusterSize = uint32(header.ClusterSize) * 16
	fs.clusterCount = header.VolumeSize / fs.clusterSize
	// The first FAT file table is 32 bytes into the file.
	fs.dataOffset = 32

	fmt.Printf("Volume size: %d\n", fs.volumeSize)
	fmt.Printf("Cluster size: %d\n", fs.clusterSize)
	fmt.Printf("Cluster count: %d\n", fs.clusterCount)

	// Ensure the cluster count is one we can work with.
	// If greater than 0xFFF5, it's most likely FAT32.
	// If less than 0xFF5, it's most likely FAT12.
	if fs.clusterCount > 0xFFF5 || fs.clusterCount < 0xFF5 {
		return nil, ErrUnknownFAT
	}

	// There are two FAT file tables at the start of the file.
	// For the purpose of this library, we will only utilize the first.
	tableSize := fs.clusterCount * 2
	tableSize = (tableSize + fs.clusterSize - 1) & ^(fs.clusterSize - 1)

	// We already have an offset past the header.
	tableData, err := NewFAT16(fs.readData(0, tableSize))
	if err != nil {
		return nil, err
	}
	fs.tableData = tableData

	log.Println(tableData.GetChain(26))

	// Advance the tracked data offset past both two tables.
	fs.dataOffset = tableSize * 2

	return &fs, nil
}

// readData returns data from the VFF adjusting to the set data offset.
func (v *VFFFS) readData(offset uint32, length uint32) []byte {
	offset += v.dataOffset
	return v.fileData[offset : offset+length]
}
