package main

import (
	"github.com/wii-tools/vffmod"
)

func main() {
	vffile, err := vffmod.OpenVFF("/Users/spot/Desktop/wc24recv.mbx")
	if err != nil {
		panic(err)
	}

	_, err = vffile.Open("MB/R0000032.MSG")
	if err != nil {
		panic(err)
	}
}
