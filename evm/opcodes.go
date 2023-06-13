package main

import (
	"fmt"
	"math/big"
)

type EVM struct {
	stack   []*big.Int
	memory  []byte
	storage map[uint64]*big.Int
	pc      int
	gas     int
	opcodes map[uint64]func(*EVM, []byte) bool
}

func NewEVM(initialGas int) *EVM {
	evm := &EVM{
		stack:   []*big.Int{},
		memory:  []byte{},
		storage: make(map[uint64]*big.Int),
		pc:      0,
		gas:     initialGas,
		opcodes: map[uint64]func(*EVM, []byte) bool{
			0x00: (*EVM).opStop,
			0x01: (*EVM).opAdd,
			0x02: (*EVM).opMul,
			0x03: (*EVM).opSub,
			0x04: (*EVM).opDiv,
			0x60: (*EVM).opPush1,
		},
	}
	return evm
}

func (evm *EVM) consumeGas(amount int) {
	if evm.gas < amount {
		panic("Out of gas")
	}
	evm.gas -= amount
}

func (evm *EVM) opStop(bytecode []byte) bool {
	return true
}

func (evm *EVM) opAdd(bytecode []byte) bool {
	n1 := evm.stack[len(evm.stack)-1]
	n2 := evm.stack[len(evm.stack)-2]
	evm.stack = evm.stack[:len(evm.stack)-2]
	result := new(big.Int).Add(n1, n2)
	result.Mod(result, bigPow(256))
	evm.stack = append(evm.stack, result)
	return false
}

func (evm *EVM) opMul(bytecode []byte) bool {
	n1 := evm.stack[len(evm.stack)-1]
	n2 := evm.stack[len(evm.stack)-2]
	evm.stack = evm.stack[:len(evm.stack)-2]
	result := new(big.Int).Mul(n1, n2)
	result.Mod(result, bigPow(256))
	evm.stack = append(evm.stack, result)
	return false
}

func (evm *EVM) opSub(bytecode []byte) bool {
	n1 := evm.stack[len(evm.stack)-1]
	n2 := evm.stack[len(evm.stack)-2]
	evm.stack = evm.stack[:len(evm.stack)-2]
	result := new(big.Int).Sub(n2, n1)
	result.Mod(result, bigPow(256))
	evm.stack = append(evm.stack, result)
	return false
}

func (evm *EVM) opDiv(bytecode []byte) bool {
	n1 := evm.stack[len(evm.stack)-1]
	n2 := evm.stack[len(evm.stack)-2]
	evm.stack = evm.stack[:len(evm.stack)-2]
	result := new(big.Int)
	if n1.Cmp(big.NewInt(0)) != 0 {
		result.Div(n2, n1)
	}
	result.Mod(result, bigPow(256))
	evm.stack = append(evm.stack, result)
	return false
}

func (evm *EVM) opPush1(bytecode []byte) bool {
	if evm.pc >= len(bytecode) {
		panic("Unexpected end of bytecode")
	}
	value := new(big.Int).SetUint64(uint64(bytecode[evm.pc]))
	evm.stack = append(evm.stack, value)
	evm.pc++
	return false
}

func (evm *EVM) execute(bytecode []byte) {
	stopExecution := false
	for evm.pc < len(bytecode) && !stopExecution {
		op := uint64(bytecode[evm.pc])
		evm.pc++

		if opcodeFn, ok := evm.opcodes[op]; ok {
			gasCost := 0
			if _, exists := evm.opcodes[op]; exists {
				gasCost = 3 // Update the gas cost accordingly
			}
			evm.consumeGas(gasCost)
			stopExecution = opcodeFn(evm, bytecode)
		} else {
			if 0x60 <= op && op <= 0x7f {
				numBytes := int(op - 0x5f)
				value := big.NewInt(0)
				for i := 0; i < numBytes; i++ {
					value = value.Lsh(value, 8)
					value = value.Add(value, big.NewInt(int64(bytecode[evm.pc+i])))
				}
				evm.stack = append(evm.stack, value)
				evm.pc += numBytes
			} else {
				panic(fmt.Sprintf("Invalid opcode: %x", op))
			}
		}
	}
}

func bigPow(exp int) *big.Int {
	pow := big.NewInt(1)
	return pow.Lsh(pow, uint(exp))
}

func main() {
	initialGas := 1000
	evm := NewEVM(initialGas)
	bytecode := []byte{0x60, 0x05, 0x60, 0x05, 0x02, 0x00}
	evm.execute(bytecode)
	fmt.Println(evm.stack)
	fmt.Printf("Remaining gas: %d\n", evm.gas)
}
