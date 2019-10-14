package core

import (
	"bufio"
	"fmt"
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
	var instr *instruction

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
			instr = getInstructionByName(newName)

			if instr == nil {
				t.Errorf("Found instruction name \"%s\" in test file but cannot find in instruction list", newName)
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
				t.Errorf("Could not find op code %x in instrction %s", info.opCode, instr.name)
				return
			}

			if info.noCycles != c {
				t.Errorf("Mismatch in cycles for instricion %s. Test file: %d, implementation: %d", instr.name, info.noCycles, c)
				return
			}

			if info.length != addressingSize[addr] {
				t.Errorf("Mismatch instruction size for instruction %s. Test file: %d, implementation: %d", instr.name, info.length, addressingSize[addr])
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

func getInstructionByName(name string) *instruction {
	for _, in := range instructions {
		if in.name == name {
			return in
		}
	}

	return nil
}

type instructionInfo struct {
	addrMode  int
	assembler string
	opCode    uint8
	length    uint8
	noCycles  uint8
}

var nameToAddressing map[string]int = map[string]int{
	"Accumulator":  accumulatorAddressing,
	"Implied":      impliedAddressing,
	"Immediate":    immediateAddressing,
	"Zero Page":    zeroPageAddressing,
	"Zero Page,X":  zeroPageXAddressing,
	"Zero Page,Y":  zeroPageYAddressing,
	"Relative":     relativeAddressing,
	"Absolute":     absoluteAddressing,
	"Absolute,X":   absoluteXAddressing,
	"Absolute,Y":   absoluteYAddressing,
	"Indirect":     indirectAddressing,
	"(Indirect,X)": indirectXAddressing,
	"(Indirect),Y": indirectYAddressing,
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

func findByOpCode(inst *instruction, opCode uint8) (addrMode int, cycles uint8, found bool) {
	for c, val := range inst.opCodes {
		if c == opCode {
			return val.addrMode, val.cycles, true
		}
	}

	return 0, 0, false
}
