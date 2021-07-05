package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/wii-tools/vffmod"
)

func main() {
	if len(os.Args) <= 2 {
		fmt.Printf("Usage: %s [path to a valid VFF] [directory to read in the VFF]", os.Args[0])
		os.Exit(1)
	}

	vffile, err := vffmod.OpenVFF(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to open the VFF \"%s\": \"%s\".\n", os.Args[1], err.Error())
		os.Exit(2)
	}

	file, err := vffile.Open(os.Args[2])
	if err != nil {
		fmt.Printf("Failed to read the file/directory \"%s\" inside \"%s\": \"%s\".\n", os.Args[2], os.Args[1], err.Error())
		os.Exit(3)
	}

	info, err := file.Stat()
	if err != nil {
		fmt.Printf("Failed to fetch information for the file \"%s\": \"%s\".\n", os.Args[2], err.Error())
		os.Exit(4)
	}

	if info.IsDir() {
		fmt.Printf("- Name: \"%s\"\n- Type: directory\n", info.Name())
	} else {
		fmt.Printf("- Name: \"%s\"\n- Type: file\n- File size: %d\n", info.Name(), info.Size())
		buffer := make([]byte, info.Size())
		file.Read(buffer)
		ioutil.WriteFile(info.Name(), buffer, 0777)
	}
}
