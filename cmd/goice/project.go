package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const VERSION = "1.2"

func NewDefaultProject() Project {
	return Project{
		Version: VERSION,
		Dependencie: Dependencie{
			Package: Package{
				Name:        "",
				Version:     "",
				Description: "",
				Author:      "",
				Image:       "",
			},
			Design: Design{
				Board: "",
				Graph: Graph{
					Blocks: []Block{},
					Wires:  []Wire{},
				},
			},
		},
		Dependencies: map[string]Dependencie{},
	}
}

func Open(filePath string) error {
	f, err := os.Open(filePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}
	defer f.Close()
	name := filepath.Base(f.Name())
	// base name without extension
	fmt.Println(strings.TrimSuffix(filepath.Base(f.Name()), filepath.Ext(f.Name())))

	// load
	return Load(name, f)
}

func Load(name string, r io.Reader) error {
	safeLoad(r)
	return nil
}

func safeLoad(r io.Reader) *Project {
	var project Project
	json.NewDecoder(r).Decode(&project)
	fmt.Println(project)
	return nil
}
