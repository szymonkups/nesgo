package instructions_test

import (
	"bufio"
	"fmt"
	"github.com/szymonkups/nesgo/core/addressing"
	"github.com/szymonkups/nesgo/core/instructions"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestInstructions(t *testing.T) {
	const fileName = "./testdata/instructions.txt"
	file, err := os.Open(fileName)

	if err != nil {
		t.Errorf("Cannot open testdata file, be sure that \"%s\" file exists", fileName)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var instr *instructions.Instruction
	for scanner.Scan() {
		line := scanner.Text()

		if checkIfIgnored(line, t) {
			continue
		}

		newName, err := checkIfName(line)

		if err != nil {
			t.Error(err)
			return
		}

		if newName != "" {
			var ok bool
			instr, ok = instructions.GetInstructionByName(newName)

			if !ok {
				t.Errorf("Found Instruction name \"%s\" in test file but cannot find in Instruction list", newName)
				return
			}

			continue
		}

		info, err := getInstructionInfo(line)

		if err != nil {
			t.Error(err)
			return
		}

		if info != nil && instr != nil {
			addr, c, found := findByOpCode(instr, info.opCode)

			if !found {
				t.Errorf("Could not find op code %x in instrction %s", info.opCode, instr.Name)
				return
			}

			addrMode, ok := addressing.GetAddressingById(addr)

			if !ok {
				t.Errorf("Could not find addressing mode %d for Instruction %s", addr, instr.Name)
				return
			}

			if info.noCycles != c {
				t.Errorf("Mismatch in Cycles for instricion %s. Test file: %d, implementation: %d", instr.Name, info.noCycles, c)
				return
			}

			if info.length != addrMode.Size {
				t.Errorf("Mismatch Instruction size for Instruction %s. Test file: %d, implementation: %d", instr.Name, info.length, addrMode.Size)
			}

			continue
		}
	}

	if err := scanner.Err(); err != nil {
		t.Error(err)
	}
}

func checkIfIgnored(line string, t *testing.T) bool {
	// Ignored pattern
	matched, err := regexp.MatchString("^(\\s*|##.*|[+-]*)$", line)

	if err != nil {
		t.Error(err)
	}

	return matched
}

func checkIfName(line string) (string, error) {
	// Ignored pattern
	r, err := regexp.Compile("^# ([A-Z]*)$")

	if err != nil {
		return "", err
	}

	res := r.FindStringSubmatch(line)

	if len(res) == 2 {
		return res[1], nil
	}

	return "", nil
}

type instructionInfo struct {
	addrMode  int
	assembler string
	opCode    uint8
	length    uint8
	noCycles  uint8
}

var nameToAddressing map[string]int = map[string]int{
	"Accumulator":  addressing.AccumulatorAddressing,
	"Implied":      addressing.ImpliedAddressing,
	"Immediate":    addressing.ImmediateAddressing,
	"Zero Page":    addressing.ZeroPageAddressing,
	"Zero Page,X":  addressing.ZeroPageXAddressing,
	"Zero Page,Y":  addressing.ZeroPageYAddressing,
	"Relative":     addressing.RelativeAddressing,
	"Absolute":     addressing.AbsoluteAddressing,
	"Absolute,X":   addressing.AbsoluteXAddressing,
	"Absolute,Y":   addressing.AbsoluteYAddressing,
	"Indirect":     addressing.IndirectAddressing,
	"(Indirect,X)": addressing.IndirectXAddressing,
	"(Indirect),Y": addressing.IndirectYAddressing,
}

func getInstructionInfo(line string) (*instructionInfo, error) {
	r, err := regexp.Compile(`^\|\s*([^|]*)\s*\|([^|]*)\|\s*(\S*)\s*\|\s*(\d+)\s*\|\s*(\d+)\*?\s*\|$`)

	if err != nil {
		return nil, err
	}

	res := r.FindStringSubmatch(line)

	if len(res) == 6 {
		addrName := strings.Trim(res[1], " ")
		a, ok := nameToAddressing[addrName]

		if !ok {
			return nil, fmt.Errorf("found unknown addressing name in test file: \"%s\"", addrName)
		}

		opCode, err := strconv.ParseInt(res[3], 16, 64)

		if err != nil {
			return nil, err
		}

		length, err := strconv.ParseInt(res[4], 16, 64)

		if err != nil {
			return nil, err
		}

		cycles, err := strconv.ParseInt(res[5], 16, 64)

		if err != nil {
			return nil, err
		}

		info := instructionInfo{
			addrMode:  a,
			assembler: strings.Trim(res[2], " "),
			opCode:    uint8(opCode),
			noCycles:  uint8(cycles),
			length:    uint8(length),
		}

		return &info, nil
	}

	return nil, nil
}

func findByOpCode(inst *instructions.Instruction, opCode uint8) (addrMode int, cycles uint8, found bool) {
	for c, val := range inst.AddrByOpCode {
		if c == opCode {
			return val.AddrMode, val.Cycles, true
		}
	}

	return 0, 0, false
}
