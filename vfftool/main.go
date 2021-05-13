package main

import (
	"github.com/wii-tools/vffmod"
	"log"
)

func main() {
	vffile, err := vffmod.OpenVFF("/Users/spot/Desktop/wc24recv.mbx")
	if err != nil {
		panic(err)
	}

	// Attempt opening the main directory "MB"
	testing, err := vffile.Open("MB")
	if err != nil {
		panic(err)
	}
	dumpFile(testing)

	// Attempt opening a file within "MB"
	second, err := vffile.Open("MB/R0000032.MSG")
	if err != nil {
		panic(err)
	}
	dumpFile(second)
}

func dumpFile(file *vffmod.FATFile) {
	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}
	log.Printf("I found %s, and to say it's a directory is %t", stat.Name(), stat.IsDir())
}
