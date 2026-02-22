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
	F1          string `json:"f1"`
	F2      string `json:"f2"`
	F3 string `json:"f3"`
	F4      string `json:"f4"`
	F5   string `json:"f5"`
	F6   string `json:"f6"`
	F7 string `json:"f7"`
}

func (s *StructA) FuncA(ctx contractapi.TransactionContextInterface, v8 string, v9 string, v10 string) error {
	v1, err := s.FuncD(ctx, v8)
	if err != nil {
		return err
	}
	if v1 {
		return fmt.Errorf("shipment %s already exists", v8)
	}

	v2 := time.Now()

	v3 := StructB{
		F1:          v8,
		F2:      v9,
		F3: v10,
		F4:      "CREATED",
		F5:   v2.Format(time.RFC3339),     
		F6:   v2.Format(time.RFC3339Nano), 
	}

	v4, err := json.Marshal(v3)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v8, v4)
}

func (s *StructA) FuncB(ctx contractapi.TransactionContextInterface, v8 string, v11 string) error {
	v4, err := ctx.GetStub().GetState(v8)
	if err != nil {
		return fmt.Errorf("failed to read shipment %s: %v", v8, err)
	}
	if v4 == nil {
		return fmt.Errorf("shipment %s does not exist", v8)
	}

	var v3 StructB
	err = json.Unmarshal(v4, &v3)
	if err != nil {
		return err
	}

	v3.F4 = v11
	
	v3.F6 = time.Now().Format(time.RFC3339)

	v5, err := json.Marshal(v3)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v8, v5)
}

func (s *StructA) FuncC(ctx contractapi.TransactionContextInterface, v8 string, v12 string) error {
	v4, err := ctx.GetStub().GetState(v8)
	if err != nil {
		return fmt.Errorf("failed to read shipment %s: %v", v8, err)
	}
	if v4 == nil {
		return fmt.Errorf("shipment %s does not exist", v8)
	}

	var v3 StructB
	err = json.Unmarshal(v4, &v3)
	if err != nil {
		return err
	}

	v3.F7 = fmt.Sprintf("%s°C at %s", v12, time.Now().Format("15:04:05.000"))
	v3.F6 = time.Now().Format(time.RFC3339) 

	v5, err := json.Marshal(v3)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v8, v5)
}

func (s *StructA) FuncD(ctx contractapi.TransactionContextInterface, v8 string) (bool, error) {
	v6, err := ctx.GetStub().GetState(v8)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return v6 != nil, nil
}

func main() {
	v7, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v7.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
