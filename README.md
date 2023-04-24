# GoIce

Herramienta para usar .ice desde la consola

## Limitaciones

* Solo puede **leer** archivos *.ice y generar archivos main.v y main.pfc


## Instalar

```bash
[jcca@arch]$ go install github.com/jcca/goice/cmd/goice@latest
[jcca@arch]$ ~/go/bin/goice 
Usage:  goice -o ./main -b ../icestudio/app/resources/boards demo.ice
```

## Ejemplos

```bash
[jcca@arch]$ ~/go/bin/goice -o ./main -b ../icestudio/app/resources/boards z80-soc-16KB-boot.ice
[jcca@arch]$ apio build --board alhambra-ii -p ./main
[jcca@arch]$ iceprog -d i:0x0403:0x6010:0 ./main/hardware.bin
init..
cdone: high
reset..
cdone: low
flash ID: 0xEF 0x40 0x16 0x00
file size: 135100
erase 64kB sector at 0x000000..
erase 64kB sector at 0x010000..
erase 64kB sector at 0x020000..
programming..
reading..
VERIFY OK
cdone: high
Bye.
```

Utilizando `yosys`, `nextpnr-ice40` y `icepack` que es lo que internamente ejecuta `apio`:

```bash
[jcca@arch]$ ~/go/bin/goice -o ./main -b ../icestudio/app/resources/boards z80-soc-16KB-boot.ice
[jcca@arch]$ cd main
[jcca@arch]$ yosys -q -f verilog -p 'synth_ice40 -json hardware.json' main.v
[jcca@arch]$ nextpnr-ice40 --hx8k --package tq144:4k --json hardware.json --asc hardware.asc --pcf main.pcf -q
[jcca@arch]$ icepack hardware.asc hardware.bin
[jcca@arch]$ iceprog -d i:0x0403:0x6010:0 hardware.bin
init..
cdone: high
reset..
cdone: low
flash ID: 0xEF 0x40 0x16 0x00
file size: 135100
erase 64kB sector at 0x000000..
erase 64kB sector at 0x010000..
erase 64kB sector at 0x020000..
programming..
reading..
VERIFY OK
cdone: high
Bye.
```
