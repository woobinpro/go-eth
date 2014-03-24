package ethchain

import (
	"bytes"
	"fmt"
	"github.com/ethereum/eth-go/ethdb"
	"github.com/ethereum/eth-go/ethutil"
	"github.com/obscuren/mutan"
	"math/big"
	"strings"
	"testing"
)

func TestRun3(t *testing.T) {
	ethutil.ReadConfig("")

	db, _ := ethdb.NewMemDatabase()
	state := NewState(ethutil.NewTrie(db, ""))

	script := Compile([]string{
		"PUSH", "300",
		"PUSH", "0",
		"MSTORE",

		"PUSH", "32",
		"CALLDATA",

		"PUSH", "64",
		"PUSH", "0",
		"RETURN",
	})
	tx := NewTransaction(ContractAddr, ethutil.Big("100000000000000000000000000000000000000000000000000"), script)
	addr := tx.Hash()[12:]
	contract := MakeContract(tx, state)
	state.UpdateContract(contract)

	callerScript := ethutil.Assemble(
		"PUSH", 1337, // Argument
		"PUSH", 65, // argument mem offset
		"MSTORE",
		"PUSH", 64, // ret size
		"PUSH", 0, // ret offset

		"PUSH", 32, // arg size
		"PUSH", 65, // arg offset
		"PUSH", 1000, /// Gas
		"PUSH", 0, /// value
		"PUSH", addr, // Sender
		"CALL",
		"PUSH", 64,
		"PUSH", 0,
		"RETURN",
	)
	callerTx := NewTransaction(ContractAddr, ethutil.Big("100000000000000000000000000000000000000000000000000"), callerScript)

	// Contract addr as test address
	account := NewAccount(ContractAddr, big.NewInt(10000000))
	callerClosure := NewClosure(account, MakeContract(callerTx, state), state, big.NewInt(1000000000), new(big.Int))

	vm := NewVm(state, RuntimeVars{
		origin:      account.Address(),
		blockNumber: 1,
		prevHash:    ethutil.FromHex("5e20a0453cecd065ea59c37ac63e079ee08998b6045136a8ce6635c7912ec0b6"),
		coinbase:    ethutil.FromHex("2adc25665018aa1fe0e6bc666dac8fc2697ff9ba"),
		time:        1,
		diff:        big.NewInt(256),
		// XXX Tx data? Could be just an argument to the closure instead
		txData: nil,
	})
	ret := callerClosure.Call(vm, nil)

	exp := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 44, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 57}
	if bytes.Compare(ret, exp) != 0 {
		t.Errorf("expected return value to be %v, got %v", exp, ret)
	}
}

func TestRun4(t *testing.T) {
	ethutil.ReadConfig("")

	db, _ := ethdb.NewMemDatabase()
	state := NewState(ethutil.NewTrie(db, ""))

	mutan.NewCompiler().Compile(strings.NewReader(`
a = 1337
c = 1
[0] = 50
d = [0]
`))

	asm := mutan.NewCompiler().Compile(strings.NewReader(`
	a = 3 + 3
	[1000] = a
	[1000]
`))
	asm = append(asm, "LOG")
	fmt.Println(asm)

	callerScript := ethutil.Assemble(asm...)
	callerTx := NewTransaction(ContractAddr, ethutil.Big("100000000000000000000000000000000000000000000000000"), callerScript)

	// Contract addr as test address
	account := NewAccount(ContractAddr, big.NewInt(10000000))
	callerClosure := NewClosure(account, MakeContract(callerTx, state), state, big.NewInt(1000000000), new(big.Int))

	vm := NewVm(state, RuntimeVars{
		origin:      account.Address(),
		blockNumber: 1,
		prevHash:    ethutil.FromHex("5e20a0453cecd065ea59c37ac63e079ee08998b6045136a8ce6635c7912ec0b6"),
		coinbase:    ethutil.FromHex("2adc25665018aa1fe0e6bc666dac8fc2697ff9ba"),
		time:        1,
		diff:        big.NewInt(256),
		// XXX Tx data? Could be just an argument to the closure instead
		txData: nil,
	})
	callerClosure.Call(vm, nil)
}
