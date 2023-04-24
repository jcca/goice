package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
)

var output string
var boardsPath string

var boards map[string]*Board

func main() {
	flag.StringVar(&output, "o", "", "Name of output file")
	flag.StringVar(&boardsPath, "b", "", "Boards path")
	flag.Parse()

	if flag.NArg() != 1 || boardsPath == "" {
		fmt.Println("Usage: ", "goice -o ./main -b ../icestudio/app/resources/boards demo.ice")
		os.Exit(0)
	}

	iceFile := flag.Arg(0)

	os.MkdirAll(output, os.ModePerm)

	var err error
	if boards, err = LoadBoards(boardsPath); err != nil {
		fmt.Println("Can't read boards")
		os.Exit(0)
	}

	var project Project
	if err := readJson(iceFile, &project); err != nil {
		panic(err)
	}

	project.Board = boards[project.Design.Board]

	for i, blockInstance := range project.Design.Graph.Blocks {
		if dependencies, ok := project.Dependencies[blockInstance.Type]; ok {
			LoadGeneric(project, &project.Design.Graph.Blocks[i], dependencies, project.Design.Graph.Wires, false)
		}
	}

	for _, data := range ListCompiler(project) {
		f, _ := os.Create(path.Join(output, data.Name))
		f.WriteString(data.Content)
		f.Close()
	}

	var out io.WriteCloser
	out, err = os.Create(path.Join(output, "main.v"))
	if err != nil {
		panic(err)
	}
	opt := Options{
		DateTime:   false,
		BoardRules: true,
		InitPorts:  []InitPort{},
		InitPins:   []Output{},
	}
	Generate("verilog", project, opt, out)
	out.Close()
	out, err = os.Create(path.Join(output, "main.pcf"))
	if err != nil {
		panic(err)
	}
	Generate("pcf", project, Options{BoardRules: true}, out)
	out.Close()
}
