package vffmod

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

var (
	ErrInvalidFormat = errors.New("this file does not appear to be a VFF")
	ErrInvalidMagic  = errors.New("invalid VFF magic detected")
	ErrUnknownFAT    = errors.New("non-FAT16 VFF detected")
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

	fs.FS
}

// getEntry loops through an array of entries and returns however possible.
func getEntry(entries []FATFile, name string) (*FATFile, error) {
	for _, entry := range entries {
		info, err := entry.Stat()
		if err != nil {
			return nil, err
		}

		if info.Name() == name {
			return &entry, nil
		}
	}

	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

// Open opens a file by the given name.
func (v *VFFFS) Open(name string) (*FATFile, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	notFound := &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}

	// Start by reading the very first directory.
	currentDirEntries := v.readEntries(0)
	pathIndex := strings.Index(name, "/")
	// Check if we need to handle other directories per the path.
	for pathIndex != -1 {
		// TODO: recurse through other directories
		// and rketrieve them
		fmt.Println("Oops! Unimplemented")
		return nil, notFound
	}

	intendedName := name[pathIndex+1:]
	return getEntry(currentDirEntries, intendedName)
}
