package main

import (
	"fmt"
	"github.com/wii-tools/vffmod"
	"io/ioutil"
	"os"
)

func extract(vffile *vffmod.VFFFS) {
	if len(os.Args) <= 3 {
		fmt.Printf("Usage: %s extract [path to a valid VFF] [directory to read in the VFF]\n", os.Args[0])
		os.Exit(1)
	}

	fileToRead := os.Args[3]

	file, err := vffile.Open(fileToRead)
	if err != nil {
		fmt.Printf("Failed to read the file/directory \"%s\" inside the given VFF: \"%v\".\n", fileToRead, err)
		os.Exit(3)
	}

	info, err := file.Stat()
	if err != nil {
		fmt.Printf("Failed to fetch information for the file \"%s\": \"%v\".\n", fileToRead, err)
		os.Exit(4)
	}

	stat(info)
	if !info.IsDir() {
		// Extract the file.
		buffer := make([]byte, info.Size())
		file.Read(buffer)
		ioutil.WriteFile(info.Name(), buffer, 0777)
	}
}
