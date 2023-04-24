package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

func Generate(target string, project Project, opt Options, out io.Writer) {

	switch target {
	case "verilog":
		fmt.Fprint(out, Header("//", nil))
		fmt.Fprint(out, "`default_nettype none\n\n")
		VerilogCompiler("main", project, opt, out)
	case "pcf":
		fmt.Fprint(out, Header("#", nil))
		PcfCompiler(project, opt, out)
	}
}

func PcfCompiler(project Project, opt Options, out io.Writer) {
	blocks := project.Design.Graph.Blocks

	var value, code string
	for _, block := range blocks {
		switch block.Type {
		case BASIC_INPUT, BASIC_OUTPUT:
			for _, pin := range block.Data.Pins {
				value = ""
				if !block.Data.Virtual {
					value = pin.Value.(string)
				}
				code += "set_io "
				code += DigestId(block.ID)
				if len(block.Data.Pins) == 1 {
					code += " "
				} else {
					code += "[" + pin.Index + "] "
				}
				code += value
				code += "\n"
			}
		}
	}

	found := func(strs []string, s string) bool {
		for _, u := range strs {
			if u == s {
				return true
			}
		}
		return false
	}

	var used []string
	if opt.BoardRules {
		initPorts := GetInitPorts(project)
		for _, initPort := range initPorts {
			if found(used, initPort.Pin) {
				break
			}
			used = append(used, initPort.Pin)
			found := false
			for _, block := range blocks {
				if block.Type == BASIC_INPUT &&
					len(block.Data.Range) == 0 &&
					!block.Data.Virtual &&
					initPort.Pin == block.Data.Pins[0].Value {
					found = true
					used = append(used, initPort.Pin)
					break
				}
			}
			if !found {
				code += "set_io v"
				code += initPort.Name
				code += " "
				code += initPort.Pin
				code += "\n"
			}
		}

		initPins := GetInitPins(project)
		for i, initPin := range initPins {
			if len(initPins) == 1 {
				code += "set_io vinit "
			} else {
				code += "set_io vinit[" + strconv.Itoa(i) + "] "
			}
			code += initPin.Pin
			code += "\n"
		}
	}
	fmt.Fprint(out, code)
}

func VerilogCompiler(name string, project Project, opt Options, out io.Writer) {

	blocks := project.Design.Graph.Blocks
	dependencies := project.Dependencies

	if len(name) > 0 {

		// Ititialize input ports
		if name == "main" && opt.BoardRules {
			initPorts := opt.InitPorts
			if len(initPorts) == 0 {
				initPorts = GetInitPorts(project)
			}

			for _, initPort := range initPorts {

				source := Resource{
					Block: initPort.Name,
					Port:  "out",
				}

				var found bool
				// Find existing input block with the initPort value
				for _, block := range blocks {
					if block.Type == BASIC_INPUT &&
						len(block.Data.Range) == 0 &&
						!block.Data.Virtual &&
						initPort.Pin == block.Data.Pins[0].Value {
						found = true
						source.Block = block.ID
						break
					}
				}

				if !found {
					blocks = append(project.Design.Graph.Blocks, Block{
						ID:   initPort.Name,
						Type: BASIC_INPUT,
						Data: &Data{
							Name: initPort.Name,
							Pins: []Pin{
								{
									Index: "0",
									Value: initPort.Pin,
								},
							},
							Virtual: false,
						},
					})
					project.Design.Graph.Blocks = blocks
				}

				// Add imaginary wire between the input block and the initPort
				project.Design.Graph.Wires = append(project.Design.Graph.Wires, Wire{
					Source: &Resource{
						Block: source.Block,
						Port:  source.Port,
					},
					Target: &Resource{
						Block: initPort.Block,
						Port:  initPort.Port,
					},
				})
			}
		}

		params := GetParams(project)

		ports := GetPorts(project)

		content := GetContent(name, &project)

		// Initialize output pins
		if name == "main" && opt.BoardRules {
			initPins := opt.InitPins
			if len(initPins) == 0 {
				initPins = GetInitPins(project)
			}
			n := len(initPins)
			if n > 0 {
				ports.Out = append(ports.Out, Port{
					Name:  "vinit",
					Range: "[0:" + strconv.Itoa(n-1) + "]",
				})
				value := strconv.Itoa(n) + "'b"
				for _, pin := range initPins {
					value += pin.Bit
				}
				content += "\nassign vinit = " + value + ";"
			}
		}

		data := Data{
			Name:    name,
			Params:  params,
			Ports:   &ports,
			Content: content,
		}

		fmt.Fprintf(out, "//---- Top entity")
		io.Copy(out, Module(data))
	}
	if project.Package != nil && project.Package.Name != "" {
		fmt.Fprintf(out, "\n/*-------------------------------------------------*/\n")
		fmt.Fprintf(out, "/*-- "+project.Package.Name+"  */\n")
		fmt.Fprintf(out, "/*-- - - - - - - - - - - - - - - - - - - - - - - --*/\n")
		fmt.Fprintf(out, "/*-- "+project.Package.Description+"\n")
		fmt.Fprintf(out, "/*-------------------------------------------------*/\n")
	}

	keys := make([]string, 0, len(dependencies))
	for k := range dependencies {
		keys = append(keys, k)
	}
	// sort.Strings(keys)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})
	for _, dependencieId := range keys {
		dependency := project.Dependencies[dependencieId]
		VerilogCompiler(DigestId(dependencieId), Project{
			Package: &dependency.Package,
			Design:  dependency.Design,
		}, Options{}, out)
	}

	for _, block := range blocks {
		if block.Type == BASIC_CODE {
			data := Data{
				Name:    name + "_" + DigestId(block.ID),
				Params:  block.Data.Params,
				Ports:   block.Data.Ports,
				Content: block.Data.Code,
			}
			io.Copy(out, Module(data))
		}
	}

}

func GetContent(name string, project *Project) string {

	graph := project.Design.Graph
	blocks := graph.Blocks

	connections := Connections{}
	vwiresLutOrder := []string{}
	vwiresLut := map[string]Wires{}
	content := []string{}

	for _, lin := range blocks {
		switch lin.Type {
		case BASIC_INPUT_LABEL:
			for _, vw := range graph.Wires {
				if vw.Target.Block == lin.ID {
					wires, ok := vwiresLut[lin.Data.Name]
					if !ok {
						vwiresLut[lin.Data.Name] = Wires{}
						vwiresLutOrder = append(vwiresLutOrder, lin.Data.Name)
					}
					twire := vw.Source
					twire.Size = vw.Size
					wires.Sources = append(wires.Sources, *twire)
					vwiresLut[lin.Data.Name] = wires
				}
			}
		case BASIC_OUTPUT_LABEL:
			for _, vw := range graph.Wires {
				if vw.Source.Block == lin.ID {
					wires, ok := vwiresLut[lin.Data.Name]
					if !ok {
						vwiresLut[lin.Data.Name] = Wires{}
						vwiresLutOrder = append(vwiresLutOrder, lin.Data.Name)
					}
					twire := vw.Target
					twire.Size = vw.Size
					wires.Targets = append(wires.Targets, *twire)
					vwiresLut[lin.Data.Name] = wires
				}
			}
		}
	}

	// Create virtual wires
	for _, v := range vwiresLutOrder {
		vw := vwiresLut[v]
		if len(vw.Sources) > 0 && len(vw.Targets) > 0 {
			for _, source := range vw.Sources {
				for _, target := range vw.Targets {
					graph.Wires = append(graph.Wires, Wire{
						// ToDelete: true,
						Source: &Resource{source.Block, source.Port, source.Size},
						Target: &Resource{target.Block, target.Port, target.Size},
					})
				}
			}
		}
	}

	// Remove virtual blocks
	// Save temporal wires and delete it
	graph.WiresVirtual = []Wire{}
	wtemp := []Wire{}
	for _, wire := range graph.Wires {
		if wire.Source.Port == "outlabel" ||
			wire.Target.Port == "outlabel" ||
			wire.Source.Port == "inlabel" ||
			wire.Target.Port == "inlabel" {
			graph.WiresVirtual = append(graph.WiresVirtual, wire)
		} else {
			if wire.Source.Size > 0 {
				wire.Size = wire.Source.Size
			}

			wtemp = append(wtemp, wire)
		}
	}

	graph.Wires = wtemp

	for w, wire := range graph.Wires {
		switch wire.Source.Port {
		case "constant-out", "memory-out":
			// Local Parameters
			block := FindBlock(wire.Source.Block, graph)
			paramValue := DigestId(block.ID)
			connections.Localparam = append(connections.Localparam, "localparam p"+strconv.Itoa(w)+" = "+paramValue+";")
		default:
			// Wires
			vrange := " "
			if wire.Size > 0 {
				vrange += "[0:" + strconv.Itoa(wire.Size-1) + "] "
			}
			connections.Wire = append(connections.Wire, "wire"+vrange+"w"+strconv.Itoa(w)+";")
		}
		// Assignations
		for _, block := range graph.Blocks {
			if block.Type == "basic.input" && wire.Source.Block == block.ID {
				connections.Assign = append(connections.Assign, "assign w"+strconv.Itoa(w)+" = "+DigestId(block.ID)+";")
			} else if block.Type == BASIC_OUTPUT && wire.Target.Block == block.ID {
				if wire.Source.Port == "constant-out" || wire.Source.Port == "memory.out" {
					// TODO
				} else {
					connections.Assign = append(connections.Assign, "assign "+DigestId(block.ID)+" = w"+strconv.Itoa(w)+";")
				}
			}
		}
	}
	content = append(content, connections.Localparam...)
	content = append(content, connections.Wire...)
	content = append(content, connections.Assign...)

	// Wires Connections

	numWires := len(graph.Wires)
	for i := 1; i < numWires; i++ {
		gwi := graph.Wires[i]
		for j := 0; j < i; j++ {
			gwj := graph.Wires[j]
			if gwi.Source.Port != "constant-out" &&
				gwi.Source.Port != "memory-out" &&
				gwi.Source.Block == gwj.Source.Block &&
				gwi.Source.Port == gwj.Source.Port {
				content = append(content, "assign w"+strconv.Itoa(i)+" = w"+strconv.Itoa(j)+";")
			}
		}
	}

	content = append(content, GetInstances(name, graph)...)

	// Restore original graph
	// delete temporal wires
	//wtemp = []Wire{}

	return strings.Join(content, "\n")
}

func FindBlock(id string, graph Graph) *Block {
	for _, block := range graph.Blocks {
		if block.ID == id {
			return &block
		}
	}
	return nil
}

func GetInitPorts(project Project) (initPorts []InitPort) {
	blocks := project.Design.Graph.Blocks
	for _, block := range blocks {
		if block.Type == BASIC_CODE || !strings.HasPrefix(block.Type, "basic.") {
			for _, inPort := range block.Data.Ports.In {
				if inPort.Default != nil && inPort.Default.Apply {
					initPorts = append(initPorts, InitPort{
						Block: block.ID,
						Port:  inPort.Name,
						Name:  inPort.Default.Port,
						Pin:   inPort.Default.Pin,
					})
				}
			}
		}
	}
	return
}

func GetInitPins(project Project) (initPins []Output) {
	blocks := project.Design.Graph.Blocks

	var usedPins []string
	for _, block := range blocks {
		if block.Type == BASIC_OUTPUT {
			for _, pin := range block.Data.Pins {
				value := fmt.Sprintf("%v", pin.Value)
				if block.Data.Virtual {
					value = ""
				}
				usedPins = append(usedPins, value)
			}
		}
	}
	sort.Strings(usedPins)
	allInitPins := project.Board.Rules.Outputs
	for _, output := range allInitPins {
		if i := sort.SearchStrings(usedPins, output.Pin); !(i < len(usedPins) && usedPins[i] == output.Pin) {
			initPins = append(initPins, output)
		}

	}
	return
}

func Header(comment string, opt interface{}) (header string) {
	header += comment + " Code generated by GoIce " + "\n"
	return
}

func DigestId(id string) string {
	if strings.Contains(id, "-") {
		id = fmt.Sprintf("%x", sha1.Sum([]byte(id)))
	}

	if len(id) > 5 {
		id = id[:6]
	}

	return "v" + id
}

func GetParams(project Project) (params []Param) {
	blocks := project.Design.Graph.Blocks
	for _, block := range blocks {
		name := DigestId(block.ID)
		switch block.Type {
		case BASIC_CONSTANT:
			params = append(params, Param{
				name,
				block.Data.Value,
			})
		case BASIC_MEMORY:
			params = append(params, Param{
				name,
				`"` + name + `.list"`,
			})
		}
	}
	return
}

func GetPorts(project Project) (ports Ports) {
	blocks := project.Design.Graph.Blocks
	for _, block := range blocks {
		_range := ""
		if block.Data != nil {
			_range = block.Data.Range
		}

		port := Port{
			Name:  DigestId(block.ID),
			Range: _range,
			Size:  0,
		}
		switch block.Type {
		case BASIC_INPUT:
			ports.In = append(ports.In, port)
		case BASIC_OUTPUT:
			ports.Out = append(ports.Out, port)
		}
	}
	return
}

func Module(data Data) io.Reader {
	var buff bytes.Buffer
	if data.Ports == nil {
		return &buff
	}
	// Header
	fmt.Fprintf(&buff, "\nmodule %s", data.Name)
	var params []string
	for _, param := range data.Params {
		value := param.Value
		if value == nil || value.(string) == "" {
			value = 0
		}
		params = append(params, " parameter "+param.Name+" = "+fmt.Sprintf("%v", value))
	}

	if len(params) > 0 {
		fmt.Fprint(&buff, " #(\n")
		fmt.Fprint(&buff, strings.Join(params[:], ",\n"))
		fmt.Fprint(&buff, "\n)")
	}

	//-- Ports

	fn := func(str string, ports []Port) (mports []string) {
		for _, port := range ports {
			value := port.Range
			if len(value) > 0 {
				value += " "
			}
			mports = append(mports, " "+str+" "+value+port.Name)
		}
		return
	}
	ports := append(fn("input", data.Ports.In), fn("output", data.Ports.Out)...)

	if len(ports) > 0 {
		fmt.Fprint(&buff, " (\n")
		fmt.Fprint(&buff, strings.Join(ports, ",\n"))
		fmt.Fprint(&buff, "\n)")
	}

	if len(params) == 0 && len(ports) == 0 {
		fmt.Fprint(&buff, "\n")
	}
	fmt.Fprint(&buff, ";\n")

	// Content
	contens := strings.Split(data.Content, "\n")
	var newContens []string
	for _, element := range contens {
		newContens = append(newContens, " "+element)
	}

	fmt.Fprint(&buff, strings.Join(newContens, "\n"))

	// Footer
	fmt.Fprint(&buff, "\nendmodule\n")
	return &buff
}

func connectPorts(portName string, portsNames *[]string, ports *[]string, block Block, w int) {
	if len(portName) > 0 {
		if block.Type != BASIC_CODE {
			portName = DigestId(portName)
		}
		index := -1
		for i, pn := range *portsNames {
			if pn == portName {
				index = i
				break
			}
		}
		if index == -1 {
			*portsNames = append(*portsNames, portName)
			port := " ." + portName
			port += "(w" + strconv.Itoa(w) + ")"
			*ports = append(*ports, port)
		}
	}
}

func GetInstances(name string, graph Graph) (instances []string) {
	for _, block := range graph.Blocks {
		var instance string
		if block.Type != BASIC_INPUT &&
			block.Type != BASIC_OUTPUT &&
			block.Type != BASIC_CONSTANT &&
			block.Type != BASIC_MEMORY &&
			block.Type != BASIC_INFO &&
			block.Type != BASIC_INPUT_LABEL &&
			block.Type != BASIC_OUTPUT_LABEL {

			// Header
			if block.Type == BASIC_CODE {
				instance = name + "_" + DigestId(block.ID)
			} else {
				instance = DigestId(block.Type)
			}

			// Parameters
			var params []string
			var ports []string
			var portsNames []string
			for w, wire := range graph.Wires {
				if block.ID == wire.Target.Block && (wire.Source.Port == "constant-out" || wire.Source.Port == "memory-out") {
					paramName := wire.Target.Port
					if block.Type != BASIC_CODE {
						paramName = DigestId(paramName)
					}
					param := " ." + paramName
					param += "(p" + strconv.Itoa(w) + ")"
					params = append(params, param)
				}

				if block.ID == wire.Source.Block {
					connectPorts(wire.Source.Port, &portsNames, &ports, block, w)
				}
				if block.ID == wire.Target.Block &&
					wire.Source.Port != "constant-out" &&
					wire.Source.Port != "memory-out" {
					connectPorts(wire.Target.Port, &portsNames, &ports, block, w)
				}
			}

			if len(params) > 0 {
				instance += " #(\n" + strings.Join(params, ",\n") + "\n)"
			}

			// Instance name

			instance += " " + DigestId(block.ID)

			instance += " (\n" + strings.Join(ports, ",\n") + "\n);"

			instances = append(instances, instance)
		}
	}
	return
}

type MemoryList struct {
	Name    string
	Content string
}

func ListCompiler(project Project) (listFiles []MemoryList) {
	for _, block := range project.Design.Graph.Blocks {
		if block.Data != nil && block.Type == BASIC_MEMORY && block.Data.List != "" {
			listFiles = append(listFiles, MemoryList{
				Name:    DigestId(block.ID) + ".list",
				Content: block.Data.List,
			})
		}
	}

	for _, dependency := range project.Dependencies {
		listFiles = append(listFiles, ListCompiler(Project{
			Package: &dependency.Package,
			Design:  dependency.Design,
		})...)
	}

	return
}
