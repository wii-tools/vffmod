package vffmod

import (
	"bytes"
	"encoding/binary"
	"io/fs"
	"strings"
)

const (
	FATAddrLongName  = 0xf
	FATAddrDirectory = 0x10
)

// FATEntry represents a directory as presented within a FAT filesystem.
type FATEntry struct {
	Name                   [8]byte
	Extension              [3]byte
	Attributes             byte
	Reserved               byte
	CMS                    byte
	CreationTime           uint16
	AccessDate             uint16
	ExtendedAttributeIndex uint16
	ModificationTime       uint16
	Start                  uint16
	FileOffset             uint32
}

// readEntries parses the FAT entry table at the given offset.
func (v *VFFFS) readEntries(offset uint32) []FATFile {
	var fileInfo []FATFile

	for dirOffset := 0; dirOffset < 32*32; dirOffset += 32 {
		var dirEntry FATEntry
		contents := v.readData(uint32(dirOffset), 32)
		err := binary.Read(bytes.NewBuffer(contents), binary.BigEndian, &dirEntry)
		if err != nil {
			panic(err)
		}

		// If we see 0x00, this is not a used entry.
		// If we see 0xe5, the file has been deleted, so we treat it as not used.
		if dirEntry.Name[0] == 0x00 || dirEntry.Name[0] == 0xe5 {
			continue
		}

		// Is this entry meant to be a long name for VFAT?
		if dirEntry.Attributes&FATAddrLongName == FATAddrLongName {
			continue
		}

		// dataOffset must grow by 1024 to be able to read over the current entry table.
		fileInfo = append(fileInfo, FATFile{
			info: FATFileInfo{
				currentFile: dirEntry,
				dataOffset:  offset + 1024,
			},
		})
	}

	return fileInfo
}

// FATFile provides fs.File-like information.
type FATFile struct {
	info FATFileInfo
	fs.File
}

// Stat is supposed to return statistics, and we already have them computed.
func (f *FATFile) Stat() (FATFileInfo, error) {
	return f.info, nil
}

// FATFileInfo provides fs.FileInfo-like information.
type FATFileInfo struct {
	// currentFile is where we are siphoning data from.
	currentFile FATEntry
	// dataOffset is the offset
	dataOffset uint32

	fs.FileInfo
}

// Mode returns a spoofed-together mode for the current entry.
// It only handles directories, as that is all we know how to handle.
func (e *FATFileInfo) Mode() fs.FileMode {
	if e.currentFile.Attributes&FATAddrDirectory == FATAddrDirectory {
		return fs.ModeDir
	} else {
		return 0
	}
}

// IsDir returns whether the entry is a directory or not.
func (e *FATFileInfo) IsDir() bool {
	return e.currentFile.Attributes&FATAddrDirectory == FATAddrDirectory
}

// Name returns the name of the current entry.
func (e *FATFileInfo) Name() string {
	dirEntry := e.currentFile
	name := strings.TrimSpace(string(dirEntry.Name[:]))
	if name[len(name)-1:] == "." {
		name = name[:]
	}

	// Directories have no extension.
	if e.IsDir() {
		return name
	}

	extension := strings.TrimSpace(string(dirEntry.Extension[:]))
	return name + "." + extension
}