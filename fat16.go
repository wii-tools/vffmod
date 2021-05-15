package vffmod

import (
	"encoding/binary"
	"errors"
)

const (
	ClusterFlagAvailable   = 0x0000
	ClusterFlagBase        = 0xFFF0
	ClusterFlagReservedMin = ClusterFlagBase
	ClusterFlagReservedMax = ClusterFlagBase + 6
	ClusterFlagBad         = ClusterFlagBase + 7
	ClusterFlagLastBase    = ClusterFlagBase + 8
)

var (
	ErrInvalidSize  = errors.New("partition table size was not aligned to 2")
	ErrInvalidChain = errors.New("encountered non-final cluster where one should be")
)

// Cluster represents the status of a cluster.
type Cluster uint16

// Available determines if the current cluster represents an available one.
func (c *Cluster) Available() bool {
	return *c == 0
}

// Used determines if the current cluster is in use.
func (c *Cluster) Used() bool {
	return *c > ClusterFlagAvailable && *c < ClusterFlagBase
}

// Reserved determines if the current cluster is marked as reserved.
func (c *Cluster) Reserved() bool {
	return *c >= ClusterFlagReservedMin && *c <= ClusterFlagReservedMax
}

// Bad determines if the current cluster is marked as bad.
func (c *Cluster) Bad() bool {
	return *c == ClusterFlagBad
}

// Last determines if the current cluster represents the end of a chain.
func (c *Cluster) Last() bool {
	return *c >= ClusterFlagLastBase
}

// FAT16 represents an array of cluster data.
type FAT16 []Cluster

// NewFAT16 returns the representation of a FAT16 file table
// from the passed data.
func NewFAT16(data []byte) (*FAT16, error) {
	// Ensure we have data that can be properly converted to a slice of uint16.
	if len(data)%2 != 0 {
		return nil, ErrInvalidSize
	}

	var working FAT16

	// For every two bytes, convert them to uint16s.
	for pos := 0; pos < len(data); pos += 2 {
		working = append(working, Cluster(binary.LittleEndian.Uint16(data[pos:pos+2])))
	}

	return &working, nil
}

func (f *FAT16) GetChain(clusterNum uint16) ([]Cluster, error) {
	currentTable := *f
	var chain []Cluster
	var current Cluster
	currentNum := clusterNum

	// Ensure the first cluster's value is present.
	chain = append(chain, Cluster(currentNum))

	for {
		// Add the current cluster.
		current = currentTable[currentNum]

		// If this cluster is not marked as used, we've reached the end of the line.
		if !current.Used() {
			break
		}

		// Keep track of this cluster.
		chain = append(chain, current)

		// Set the next cluster number to retrieve.
		currentNum = uint16(current)
	}

	// Ensure the chain has been properly ended.
	if !current.Last() {
		return nil, ErrInvalidChain
	} else {
		return chain, nil
	}
}

// readCluster reads the data for the given cluster number by the VFF's cluster size.
func (v *VFFFS) readCluster(clusterNum uint16) []byte {
	// All cluster numbers appear to be 2 off.
	clusterNum -= 2
	// Additionally, all cluster data is 4096 bytes off as the first cluster is 4096 bytes.
	return v.readData(v.clusterSize*uint32(clusterNum)+4096, v.clusterSize)
}

func (v *VFFFS) readChain(clusterNum uint16) ([]byte, error) {
	var data []byte

	clusters, err := v.tableData.GetChain(clusterNum)
	if err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		currentData := v.readCluster(uint16(cluster))
		data = append(data, currentData...)
	}

	return data, nil
}
