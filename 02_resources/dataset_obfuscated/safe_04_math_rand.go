package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1  string `json:"f1"`
	F2      string `json:"f2"`
	F3    string `json:"f3"`
	F4      int    `json:"f4"`
	F5      string `json:"f5"`
	F6 string `json:"f6"`
}

func (p *StructA) FuncA(ctx contractapi.TransactionContextInterface, v11 string, v12 string, v13 string, v14 int) error {
	v1, err := p.FuncD(ctx, v11)
	if err != nil {
		return err
	}
	if v1 {
		return fmt.Errorf("package %s already exists", v11)
	}

	v2 := rand.Intn(999999)
	fmt.Printf("[TRACE-%06d] RegisterPackage called: id=%s, from=%s, to=%s\n",
		v2, v11, v12, v13)

	v3 := StructB{
		F1:  v11,     
		F2:      v12,         
		F3:    v13,       
		F4:      v14,         
		F5:      "REGISTERED",   
		F6: "WH_INTAKE",    
	}

	v4, err := json.Marshal(v3)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v11, v4)
}

func (p *StructA) FuncB(ctx contractapi.TransactionContextInterface, v11 string, v15 string) error {
	v4, err := ctx.GetStub().GetState(v11)
	if err != nil {
		return fmt.Errorf("failed to read package %s: %v", v11, err)
	}
	if v4 == nil {
		return fmt.Errorf("package %s does not exist", v11)
	}

	var v3 StructB
	err = json.Unmarshal(v4, &v3)
	if err != nil {
		return err
	}

	v5 := fmt.Sprintf("CORR-%08d", rand.Intn(99999999))
	fmt.Printf("[%s] TransferPackage: %s from %s to %s\n",
		v5, v11, v3.F6, v15)

	v3.F6 = v15 
	v3.F5 = "IN_TRANSIT"       

	v6, err := json.Marshal(v3)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v11, v6)
}

func (p *StructA) FuncC(ctx contractapi.TransactionContextInterface, v11 string) error {
	v4, err := ctx.GetStub().GetState(v11)
	if err != nil {
		return fmt.Errorf("failed to read package %s: %v", v11, err)
	}
	if v4 == nil {
		return fmt.Errorf("package %s does not exist", v11)
	}

	var v3 StructB
	err = json.Unmarshal(v4, &v3)
	if err != nil {
		return err
	}

	v7 := []string{
		"Package delivered successfully!",
		"Delivery confirmed.",
		"Shipment complete.",
	}
	v8 := v7[rand.Intn(len(v7))]
	fmt.Printf("[DELIVERY] %s: %s\n", v11, v8)

	v3.F5 = "DELIVERED" 

	v6, err := json.Marshal(v3)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v11, v6)
}

func (p *StructA) FuncD(ctx contractapi.TransactionContextInterface, v16 string) (bool, error) {
	v9, err := ctx.GetStub().GetState(v16)
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
