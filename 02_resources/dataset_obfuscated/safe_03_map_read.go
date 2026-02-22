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
	F1    string `json:"f1"`
	F2  string `json:"f2"`
	F3     string `json:"f3"`
	F4 string `json:"f4"`
	F5      string `json:"f5"`
}

func (v *StructA) FuncA(ctx contractapi.TransactionContextInterface, v21 string, v9 string, v22 string) error {
	
	v1 := map[string]string{
		"C001": "Alice Johnson",
		"C002": "Bob Smith",
		"C003": "Carol Williams",
		"C004": "David Brown",
	}

	v2, v3 := v1[v22]
	if !v3 {
		return fmt.Errorf("candidate %s is not registered for this election", v22)
	}
	fmt.Printf("Vote cast for candidate: %s (%s)\n", v22, v2)

	v4 := v21 + "_" + v9
	v5, err := ctx.GetStub().GetState(v4)
	if err != nil {
		return fmt.Errorf("failed to check existing ballot: %v", err)
	}
	if v5 != nil {
		return fmt.Errorf("voter %s has already voted in election %s", v9, v21)
	}

	v6 := StructB{
		F1:    v4,
		F2:  v21,
		F3:     v9,
		F4: v22, 
		F5:      "CAST",
	}

	v7, err := json.Marshal(v6)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v4, v7)
}

func (v *StructA) FuncB(ctx contractapi.TransactionContextInterface, v21 string, v23 []string) error {
	
	v8 := make(map[string]int)

	for _, v9 := range v23 {
		v4 := v21 + "_" + v9
		v7, err := ctx.GetStub().GetState(v4)
		if err != nil {
			return fmt.Errorf("failed to read ballot for voter %s: %v", v9, err)
		}
		if v7 == nil {
			continue
		}

		var v6 StructB
		err = json.Unmarshal(v7, &v6)
		if err != nil {
			return err
		}

		v8[v6.F4]++
	}

	for v10, v11 := range v8 {
		fmt.Printf("[TALLY] %s: %d votes\n", v10, v11)
	}

	v12, err := json.Marshal(v8)
	if err != nil {
		return err
	}

	v13 := "TALLY_" + v21
	return ctx.GetStub().PutState(v13, v12)
}

func (v *StructA) FuncC(ctx contractapi.TransactionContextInterface, v9 string) (bool, error) {
	
	v14 := map[string]bool{
		"DISTRICT_A": true,
		"DISTRICT_B": true,
		"DISTRICT_C": true,
		"DISTRICT_D": false, 
	}

	v15, err := ctx.GetStub().GetState("VOTER_" + v9)
	if err != nil {
		return false, fmt.Errorf("failed to read voter record: %v", err)
	}
	if v15 == nil {
		return false, fmt.Errorf("voter %s is not registered", v9)
	}

	var v20 map[string]string
	err = json.Unmarshal(v15, &v20)
	if err != nil {
		return false, err
	}

	v16 := v20["district"]

	v17, v18 := v14[v16]
	if !v18 {
		return false, nil
	}

	return v17, nil
}

func (v *StructA) FuncD(ctx contractapi.TransactionContextInterface, v21 string) (string, error) {
	v13 := "TALLY_" + v21
	v12, err := ctx.GetStub().GetState(v13)
	if err != nil {
		return "", fmt.Errorf("failed to read tally: %v", err)
	}
	if v12 == nil {
		return "", fmt.Errorf("no tally found for election %s", v21)
	}

	return string(v12), nil
}

func main() {
	v19, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v19.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
