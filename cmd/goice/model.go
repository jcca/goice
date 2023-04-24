package main

type Project struct {
	Version      string                 `json:"version,omitempty"`
	Design       Design                 `json:"design,omitempty"`
	Package      *Package               `json:"package,omitempty"`
	Dependencie  Dependencie            `json:"dependencie,omitempty"`
	Dependencies map[string]Dependencie `json:"dependencies,omitempty"`
	Board        *Board                 `json:"-"`
}

type Dependencie struct {
	Package Package `json:"package,omitempty"`
	Design  Design  `json:"design,omitempty"`
}

type Data struct {
	Name    string      `json:"name,omitempty"`
	Ports   *Ports      `json:"ports,omitempty"`
	Clock   bool        `json:"clock,omitempty"`
	Range   string      `json:"range,omitempty"`
	Virtual bool        `json:"virtual,omitempty"`
	Pins    []Pin       `json:"pins,omitempty"`
	Params  []Param     `json:"params,omitempty"`
	Content string      `json:"content,omitempty"`
	Code    string      `json:"code,omitempty"`
	Value   interface{} `json:"value,omitempty"`
	List    string      `json:"list,omitempty"`
	//Local   bool        `json:"local"`
}

type Pin struct {
	Index string      `json:"index,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type Resource struct {
	Block string `json:"block,omitempty"`
	Port  string `json:"port,omitempty"`
	Size  int    `json:"size,omitempty"`
}

type Ports struct {
	In  []Port `json:"in,omitempty"`
	Out []Port `json:"out,omitempty"`
}

type Port struct {
	Name    string `json:"name,omitempty"`
	Default *Input `json:"default,omitempty"`
	Range   string `json:"range,omitempty"`
	Size    int    `json:"size,omitempty"`
}

type (
	Section struct {
		Type   string
		Boards []string
	}

	Pinout struct {
		Type  string
		Name  string
		Value string
	}

	Info struct {
		Label         string
		Datasheet     string
		Interface     string
		FPGAResources map[string]int
	}

	Input struct {
		Port  string
		Pin   string
		Apply bool
	}

	Output struct {
		Pin string
		Bit string
	}

	Rules struct {
		Inputs  []Input  `json:"input"`
		Outputs []Output `json:"output"`
	}

	Board struct {
		Name    string
		Info    Info
		Pinouts []Pinout
		Rules   Rules
		Type    string
	}
)

type Options struct {
	DateTime   bool       `json:"datetime,omitempty"`
	BoardRules bool       `json:"boardrules,omitempty"`
	InitPorts  []InitPort `json:"initPorts,omitempty"`
	InitPins   []Output   `json:"initPins,omitempty"`
}

type Connections struct {
	Localparam []string `json:"localparam,omitempty"`
	Wire       []string `json:"wire,omitempty"`
	Assign     []string
}
type Wires struct {
	Targets []Resource `json:"targets,omitempty"`
	Sources []Resource `json:"sources,omitempty"`
}

type Package struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`
	Image       string `json:"image,omitempty"`
}

type Design struct {
	Board string `json:"board,omitempty"`
	Graph Graph  `json:"graph,omitempty"`
}

type Graph struct {
	Blocks       []Block `json:"blocks,omitempty"`
	Wires        []Wire  `json:"wires,omitempty"`
	WiresVirtual []Wire  `json:"wiresvirtual,omitempty"`
}

type Block struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Data     *Data  `json:"data,omitempty"`
	Position Point  `json:"position,omitempty"`
	Size     Size   `json:"size,omitempty"`
}

type Size struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

type Point struct {
	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`
}

type Wire struct {
	Source   *Resource `json:"source,omitempty"`
	Target   *Resource `json:"target,omitempty"`
	Size     int       `json:"size,omitempty"`
	ToDelete bool      `json:"todelete,omitempty"`
	//Vertices X, Y
}

type InitPort struct {
	Block string `json:"block,omitempty"`
	Port  string `json:"port,omitempty"`
	Name  string `json:"name,omitempty"`
	Pin   string `json:"pin,omitempty"`
}

type Param struct {
	Name  string      `json:"name,omitempty"`
	Value interface{} `json:"value,omitempty"`
}
