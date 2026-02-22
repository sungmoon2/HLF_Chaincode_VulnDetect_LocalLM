package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

var G1 int
var G2 string

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1   string  `json:"f1"`
	F2 float64 `json:"f2"`
	F3 int     `json:"f3"`
	F4  string  `json:"f4"`
}

func (t *StructA) FuncA(ctx contractapi.TransactionContextInterface, v12 string, v13 float64) error {
	
	G1++

	v1 := StructB{
		F1:   v12,
		F2: v13,
		F3: G1, 
		F4:  "ISSUE",
	}

	v2, err := json.Marshal(v1)
	if err != nil {
		return err
	}

	G2 = v12

	return ctx.GetStub().PutState(v12, v2)
}

func (t *StructA) FuncB(ctx contractapi.TransactionContextInterface, v14 string, v15 string, v13 float64) error {
	v3, err := ctx.GetStub().GetState(v14)
	if err != nil {
		return fmt.Errorf("failed to read sender: %v", err)
	}
	if v3 == nil {
		return fmt.Errorf("sender %s does not exist", v14)
	}

	v4, err := ctx.GetStub().GetState(v15)
	if err != nil {
		return fmt.Errorf("failed to read recipient: %v", err)
	}
	if v4 == nil {
		return fmt.Errorf("recipient %s does not exist", v15)
	}

	var fromToken, toToken StructB
	json.Unmarshal(v3, &fromToken)
	json.Unmarshal(v4, &toToken)

	if fromToken.F2 < v13 {
		return fmt.Errorf("insufficient balance")
	}

	fromToken.F2 -= v13
	toToken.F2 += v13

	G1++
	fromToken.F3 = G1 
	fromToken.F4 = "SEND"
	toToken.F4 = "RECEIVE"

	v5, _ := json.Marshal(fromToken)
	v6, _ := json.Marshal(toToken)

	ctx.GetStub().PutState(v14, v5)
	return ctx.GetStub().PutState(v15, v6)
}

func (t *StructA) FuncC(ctx contractapi.TransactionContextInterface) (string, error) {
	
	return fmt.Sprintf("counter=%d, lastID=%s", G1, G2), nil
}

func (t *StructA) FuncD(ctx contractapi.TransactionContextInterface, v12 string, v13 float64) error {
	v2, err := ctx.GetStub().GetState(v12)
	if err != nil {
		return fmt.Errorf("failed to read owner: %v", err)
	}
	if v2 == nil {
		return fmt.Errorf("owner %s does not exist", v12)
	}

	var v1 StructB
	json.Unmarshal(v2, &v1)

	if v1.F2 < v13 {
		return fmt.Errorf("insufficient balance to burn")
	}

	v1.F2 -= v13

	G1++
	v7 := "BURN_RECEIPT_" + strconv.Itoa(G1)

	v8 := map[string]interface{}{
		"owner":  v12,
		"burned": v13,
		"seq":    G1, 
	}
	v9, _ := json.Marshal(v8)
	ctx.GetStub().PutState(v7, v9) 

	v1.F3 = G1
	v1.F4 = "BURN"
	v10, _ := json.Marshal(v1)
	return ctx.GetStub().PutState(v12, v10)
}

func main() {
	v11, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v11.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
