package main

import (
	"fmt"
	"github.com/wii-tools/vffmod"
	"io/fs"
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

	fs.WalkDir(vffile, "MB", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		fmt.Printf("File Name: %s\n", d.Name())
		return nil
	})
}

func dumpFile(file fs.File) {
	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}
	log.Printf("found %s with a size of %d", stat.Name(), stat.Size())
}
