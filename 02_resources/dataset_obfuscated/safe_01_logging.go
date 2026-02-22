package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1         string `json:"f1"`
	F2      string `json:"f2"`
	F3     string `json:"f3"`
	F4     string `json:"f4"`
	F5 string `json:"f5"`
	F6  string `json:"f6"`
	F7  string `json:"f7"`
}

func (a *StructA) FuncA(ctx contractapi.TransactionContextInterface, v11 string, v12 string, v13 string) error {
	v1, err := a.FuncD(ctx, v11)
	if err != nil {
		return err
	}
	if v1 {
		return fmt.Errorf("document %s already exists", v11)
	}

	fmt.Printf("[%s] CreateDocument called: id=%s, title=%s\n",
		time.Now().Format(time.RFC3339), v11, v12)

	v2, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	v3 := time.Unix(v2.Seconds, int64(v2.Nanos)).Format(time.RFC3339)

	v4 := StructB{
		F1:        v11,
		F2:     v12,
		F3:    v13,
		F4:    "PENDING",
		F6: v3, 
		F7: v3,
	}

	v5, err := json.Marshal(v4)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v11, v5)
}

func (a *StructA) FuncB(ctx contractapi.TransactionContextInterface, v11 string, v14 string) error {
	v5, err := ctx.GetStub().GetState(v11)
	if err != nil {
		return fmt.Errorf("failed to read document %s: %v", v11, err)
	}
	if v5 == nil {
		return fmt.Errorf("document %s does not exist", v11)
	}

	var v4 StructB
	err = json.Unmarshal(v5, &v4)
	if err != nil {
		return err
	}

	if v4.F4 != "PENDING" {
		return fmt.Errorf("document %s is not pending approval (current: %s)", v11, v4.F4)
	}

	fmt.Printf("[%s] ApproveDocument: id=%s approved by %s\n",
		time.Now().Format("15:04:05.000"), v11, v14)

	v2, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	v3 := time.Unix(v2.Seconds, int64(v2.Nanos)).Format(time.RFC3339)

	v4.F4 = "APPROVED"
	v4.F5 = v14
	v4.F7 = v3 

	v6, err := json.Marshal(v4)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v11, v6)
}

func (a *StructA) FuncC(ctx contractapi.TransactionContextInterface, v11 string) (*StructB, error) {
	v7 := time.Now() 

	v5, err := ctx.GetStub().GetState(v11)
	if err != nil {
		return nil, fmt.Errorf("failed to read document %s: %v", v11, err)
	}
	if v5 == nil {
		return nil, fmt.Errorf("document %s does not exist", v11)
	}

	var v4 StructB
	err = json.Unmarshal(v5, &v4)
	if err != nil {
		return nil, err
	}

	v8 := time.Since(v7)
	fmt.Printf("[PERF] QueryDocument(%s) took %v\n", v11, v8)

	return &v4, nil
}

func (a *StructA) FuncD(ctx contractapi.TransactionContextInterface, v11 string) (bool, error) {
	v9, err := ctx.GetStub().GetState(v11)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return v9 != nil, nil
}

func main() {
	v10, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v10.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
