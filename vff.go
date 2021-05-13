package vffmod

import (
	"errors"
	"io/fs"
	"strings"
)

var (
	ErrInvalidFormat = errors.New("this file does not appear to be a VFF")
	ErrInvalidMagic  = errors.New("invalid VFF magic detected")
	ErrUnknownFAT    = errors.New("non-FAT16 VFF detected")

	_ fs.DirEntry = (*FATDirEntryInfo)(nil)
	_ fs.FileInfo = (*FATFileInfo)(nil)
)

// VFFFS holds a usable io/fs representation of a VFF.
type VFFFS struct {
	// VolumeSize is used to keep track of the overall size.
	// TODO: make properly mutable
	volumeSize uint32
	// ClusterSize is used to keep track of the expected cluster size.
	// It should effectively always be 512.
	clusterSize uint32
	// clusterCount is used to keep track of how many clusters
	// are within this volume.
	// TODO: make properly mutable
	clusterCount uint32

	// fileData holds the raw VFF in memory.
	// TODO: perhaps use bytes.Buffer or similar?
	fileData []byte

	// tableData holds the current VFF's partition table.
	tableData *FAT16
	// dataOffset holds the offset where data exists within the VFF.
	dataOffset uint32

	fs.ReadDirFS
	fs.FS
}

// getFile loops through an array of entries and returns the matching entry name if possible.
func getFile(entries []FATFile, name string) (*FATFile, error) {
	for _, entry := range entries {
		info, err := entry.Stat()
		if err != nil {
			return nil, err
		}

		if info.Name() == name {
			return &entry, nil
		}
	}

	return nil, fs.ErrNotExist
}

// getDirectory reads a directory and returns its entries.
func (v *VFFFS) getDirectory(entries []FATFile, dirName string) ([]FATFile, error) {
	dirEntry, err := getFile(entries, dirName)
	if err != nil {
		return nil, err
	}

	// Ensure the retrieved file is a directory.
	if !dirEntry.info.IsDir() {
		return nil, err
	}

	// Read entries within this directory.
	newData, err := v.readChain(dirEntry.info.currentFile.ClusterNum)
	if err != nil {
		return nil, err
	}

	// Parse the given entry table for our new directory.
	return v.parseEntries(newData), nil
}

// Open opens a file by the given name.
func (v *VFFFS) Open(name string) (fs.File, error) {
	var err error
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	// Start by reading the very first directory.
	currentDirectory := v.parseEntries(v.readData(0, 4096))
	pathComponents := strings.Split(name, "/")

	// Check if we need to handle other directories per the path.
	for len(pathComponents) > 1 {
		// Get the directory we need to look for.
		dirName := pathComponents[0]

		currentDirectory, err = v.getDirectory(currentDirectory, dirName)
		if err != nil {
			return nil, &fs.PathError{Op: "open", Path: name, Err: err}
		}

		// Strip the currently handled directory from our list.
		pathComponents = pathComponents[1:]
	}

	// Now that we have the directory we need to handle,
	// we can open by only the filename.
	filename := pathComponents[0]
	return getFile(currentDirectory, filename)
}

// ReadDir reads a given directory.
// TODO: have more common code
func (v *VFFFS) ReadDir(name string) ([]fs.DirEntry, error) {
	var err error
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrInvalid}
	}

	// Start by reading the very first directory.
	currentDirectory := v.parseEntries(v.readData(0, 4096))
	pathComponents := strings.Split(name, "/")

	// Check if we need to handle other directories per the path.
	for len(pathComponents) > 1 {
		// Get the directory we need to look for.
		dirName := pathComponents[0]

		currentDirectory, err = v.getDirectory(currentDirectory, dirName)
		if err != nil {
			return nil, &fs.PathError{Op: "read", Path: name, Err: err}
		}

		// Strip the currently handled directory from our list.
		pathComponents = pathComponents[1:]
	}

	infos := make([]fs.DirEntry, len(currentDirectory))

	for i, file := range currentDirectory {
		infos[i] = &FATDirEntryInfo{
			currentInfo: file.info,
		}
	}

	return infos, nil
}
