package main

import (
	"fmt"
	"github.com/wii-tools/vffmod"
	"io/fs"
	"log"
)

func tree(vffile *vffmod.VFFFS) {
	fs.WalkDir(vffile, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		fmt.Printf("Traversed to %s\n", d.Name())
		return nil
	})
}

