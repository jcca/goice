package main

const (
	BASIC_INPUT         = "basic.input"       //-- Input ports
	BASIC_OUTPUT        = "basic.output"      //-- Output ports
	BASIC_INPUT_LABEL   = "basic.inputLabel"  //-- Input labels
	BASIC_OUTPUT_LABEL  = "basic.outputLabel" //-- Output labels
	BASIC_PAIRED_LABELS = "basic.pairedLabel" //-- Paired labels
	BASIC_CODE          = "basic.code"        //-- Verilog code
	BASIC_MEMORY        = "basic.memory"      //-- Memory parameter
	BASIC_CONSTANT      = "basic.constant"    //-- Constant parameter
	BASIC_INFO          = "basic.info"        //-- Info block
)

type PortBlock struct {
	Block
	Pins []string
}

type InputPortBlock struct {
	PortBlock
	Clock string
}

func LoadGeneric(project Project, instance *Block, block Dependencie, wires []Wire, disable bool) {

	instance.Data = &Data{Ports: &Ports{}}
	for _, item := range block.Design.Graph.Blocks {
		if item.Type == BASIC_INPUT && len(item.Data.Range) == 0 {
			name := item.Data.Name
			if item.Data.Clock {
				name = "clk"
			}

			_default := HasInputRule(project, name, true)
			if _default == nil {
				continue
			}

			for _, wire := range wires {
				if wire.Target != nil && wire.Target.Port == item.ID {
					_default.Apply = false
					break
				}
			}

			instance.Data.Ports.In = append(instance.Data.Ports.In, Port{
				Name:    item.ID,
				Default: _default,
			})
		}
	}
}

func HasInputRule(project Project, port string, apply bool) *Input {
	var result *Input
	allInitPorts := project.Board.Rules.Inputs
	for _, p := range allInitPorts {
		if port == p.Port {
			result = &p
			result.Apply = apply
			break
		}
	}
	return result
}
