package vffmod

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/fs"
	"strings"
)

const (
	FATAddrLongName  = 0xf
	FATAddrDirectory = 0x10
)

var (
	// ForbiddenSingleDot is an ASCII "." followed by 7 spaces.
	ForbiddenSingleDot = []byte{0x2e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	// ForbiddenDoubleDot is ".." in ASCII followed by 6 spaces.
	ForbiddenDoubleDot = []byte{0x2e, 0x2e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	// ForbiddenExtension is 3 spaces.
	ForbiddenExtension = []byte{0x20, 0x20, 0x20}
)

// FATEntry represents a directory as presented within a FAT filesystem.
type FATEntry struct {
	Name                   [8]byte
	Extension              [3]byte
	Attributes             byte
	Reserved               byte
	CMS                    byte
	CreationTime           uint16
	CreationDate           uint16
	AccessDate             uint16
	ExtendedAttributeIndex uint16
	ModificationTime       uint16
	ModificationDate       uint16
	ClusterNum             uint16
	Size                   uint32
}

// parseEntries parses the FAT entry table at the given offset.
func (v *VFFFS) parseEntries(data []byte) []FATFile {
	var fileInfo []FATFile

	for dirOffset := 0; dirOffset < len(data); dirOffset += 32 {
		var dirEntry FATEntry
		contents := data[dirOffset : dirOffset+32]
		err := binary.Read(bytes.NewBuffer(contents), binary.LittleEndian, &dirEntry)
		if err != nil {
			panic(err)
		}

		// If we see 0x00, this is not a used entry.
		// If we see 0xe5, the file has been deleted, so we treat it as not used.
		if dirEntry.Name[0] == 0x00 || dirEntry.Name[0] == 0xe5 {
			continue
		}

		// Is this entry meant to be a long name for VFAT?
		// TODO: consider implementing this
		// TODO: maybe don't consid
		if dirEntry.Attributes&FATAddrLongName == FATAddrLongName {
			continue
		}

		// We are not dealing with . or .. as names. All end with 3 spaces as an extension.
		if (bytes.Equal(dirEntry.Name[:], ForbiddenSingleDot) || bytes.Equal(dirEntry.Name[:], ForbiddenDoubleDot)) && bytes.Equal(dirEntry.Extension[:], ForbiddenExtension) {
			continue
		}

		fmt.Printf("%s.%s\n", strings.Trim(string(dirEntry.Name[:]), " "), strings.Trim(string(dirEntry.Extension[:]), " "))

		// dataOffset must grow by 1024 to be able to read over the current entry table.
		fileInfo = append(fileInfo, FATFile{
			info: FATFileInfo{
				currentFile: dirEntry,
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
func (f FATFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

// Close is supposed to close a file. We had nothing open to begin with.
func (f *FATFile) Close() error {
	return nil
}

// FATFileInfo provides fs.FileInfo-like information.
type FATFileInfo struct {
	// currentFile is where we are siphoning data from.
	currentFile FATEntry

	fs.FileInfo
}

// Size returns the size of the current entry.
func (e FATFileInfo) Size() int64 {
	return int64(e.currentFile.Size)
}

// Mode returns a spoofed-together mode for the current entry.
// It only handles directories, as that is all we know how to handle.
func (e FATFileInfo) Mode() fs.FileMode {
	if e.currentFile.Attributes&FATAddrDirectory == FATAddrDirectory {
		return fs.ModeDir
	} else {
		return 0
	}
}

// IsDir returns whether the entry is a directory or not.
func (e FATFileInfo) IsDir() bool {
	return e.currentFile.Attributes&FATAddrDirectory == FATAddrDirectory
}

// Name returns the name of the current entry.
func (e FATFileInfo) Name() string {
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

// FATDirEntryInfo provides fs.DirEntry-like information.
type FATDirEntryInfo struct {
	// currentInfo is where we are siphoning info from.
	currentInfo FATFileInfo

	fs.DirEntry
}

// Name returns the name of the current entry.
func (d FATDirEntryInfo) Name() string {
	return d.currentInfo.Name()
}

// IsDir returns whether the entry is a directory or not.
func (d FATDirEntryInfo) IsDir() bool {
	return d.currentInfo.IsDir()
}

// Type returns a directory-specific fs.FileMode.
func (d FATDirEntryInfo) Type() fs.FileMode {
	return d.currentInfo.Mode().Type()
}

// Info returns our mocked-up statistics.
func (d FATDirEntryInfo) Info() (fs.FileInfo, error) {
	return d.currentInfo, nil
}
