package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1   string `json:"f1"`
	F2  string `json:"f2"`
	F3    int    `json:"f3"`
	F4    string `json:"f4"`
	F5  string `json:"f5"`
}

func (ic *StructA) FuncA(ctx contractapi.TransactionContextInterface, v11 string, v12 string, v13 string) error {
	if v13 == "" {
		return fmt.Errorf("approver cannot be empty")
	}

	v1, err := ctx.GetStub().GetStateByRange(v11, v12)
	if err != nil {
		return fmt.Errorf("failed to get state by range: %v", err)
	}

	v2 := 0
	for v1.HasNext() {
		v2++
		v1.Next()
	}
	if v2 == 0 {
		
		return fmt.Errorf("no claims found in range %s to %s", v11, v12)
	}

	defer v1.Close()

	v3, err := ctx.GetStub().GetStateByRange(v11, v12)
	if err != nil {
		return err
	}
	defer v3.Close()

	for v3.HasNext() {
		v4, err := v3.Next()
		if err != nil {
			return err
		}

		var v9 StructB
		err = json.Unmarshal(v4.Value, &v9)
		if err != nil {
			return err
		}

		if v9.F4 == "PENDING" {
			v9.F4 = "APPROVED"
			v9.F5 = v13
			v5, _ := json.Marshal(v9)
			ctx.GetStub().PutState(v9.F1, v5)
		}
	}

	return nil
}

func (ic *StructA) FuncB(ctx contractapi.TransactionContextInterface, v11 string, v12 string, v14 int) ([]*StructB, error) {
	v1, err := ctx.GetStub().GetStateByRange(v11, v12)
	if err != nil {
		return nil, fmt.Errorf("failed to query range: %v", err)
	}

	var v10 []*StructB

	if v14 > 0 {
		
		for v1.HasNext() {
			v4, err := v1.Next()
			if err != nil {
				return nil, err 
			}

			var v9 StructB
			json.Unmarshal(v4.Value, &v9)
			if v9.F3 >= v14 {
				v10 = append(v10, &v9)
			}
		}
		
	} else {
		
		v1.Close()
		return nil, fmt.Errorf("threshold must be positive, got %d", v14)
	}

	return v10, nil 
}

func (ic *StructA) FuncC(ctx contractapi.TransactionContextInterface, v15 string) (int, error) {
	v1, err := ctx.GetStub().GetStateByRange(v15+"_0000", v15+"_9999")
	if err != nil {
		return 0, err
	}
	
	v6 := 0
	for v1.HasNext() {
		v4, err := v1.Next()
		if err != nil {
			return v6, err 
		}

		var v9 StructB
		json.Unmarshal(v4.Value, &v9)

		if v9.F4 == "CLOSED" {
			v7 := "ARCHIVE_" + v9.F1
			ctx.GetStub().PutState(v7, v4.Value)
			ctx.GetStub().DelState(v9.F1)
			v6++

			if v6 >= 100 {
				
				return v6, nil
			}
		}
	}

	v1.Close() 
	return v6, nil
}

func main() {
	v8, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v8.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
