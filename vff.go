package vffmod

import (
	"errors"
	"io/fs"
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
	// ClusterSize is used to keep track of
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
}

// Open
func (v *VFFFS) Open(name string) (fs.File, error) {
	return nil, nil
}
