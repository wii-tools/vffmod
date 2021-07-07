package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/wii-tools/vffmod"
)

func main() {
	if len(os.Args) <= 2 {
		usage()
	}

	subcommand := os.Args[1]
	vffpath := os.Args[2]

	vffile, err := vffmod.OpenVFF(vffpath)
	if err != nil {
		fmt.Printf("Failed to open the VFF \"%s\": \"%v\".\n", vffpath, err)
		os.Exit(2)
	}

	switch subcommand {
	case "extract":
		extract(vffile)
	case "tree":
		tree(vffile)
	default:
		usage()
	}
}

func usage() {
	fmt.Printf("Usage: %s [path to a valid VFF] [directory to read in the VFF]\n", os.Args[0])
	os.Exit(1)
}

func stat(info fs.FileInfo) {
	if info.IsDir() {
		fmt.Printf("- Name: \"%s\"\n- Type: directory\n", info.Name())
	} else {
		fmt.Printf("- Name: \"%s\"\n- Type: file\n- File size: %d\n", info.Name(), info.Size())
	}
}
