package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1    string         `json:"f1"`
	F2 int            `json:"f2"`
	F3     map[string]int `json:"f3"`
}

func (w *StructA) FuncA(ctx contractapi.TransactionContextInterface, v27 string) error {
	
	var v22 map[string]map[string]int
	err := json.Unmarshal([]byte(v27), &v22)
	if err != nil {
		return fmt.Errorf("invalid update JSON: %v", err)
	}

	for v1, v2 := range v22 {
		v3 := "ZONE_" + v1
		v4, err := ctx.GetStub().GetState(v3)
		if err != nil {
			return fmt.Errorf("failed to read zone %s: %v", v1, err)
		}

		var v23 StructB
		if v4 != nil {
			json.Unmarshal(v4, &v23)
		} else {
			v23 = StructB{F1: v1, F3: make(map[string]int)}
		}

		for v5, v6 := range v2 {
			v23.F3[v5] += v6
			v23.F2 += v6

			v7 := v1 + "_ITEM_" + v5
			v8 := fmt.Sprintf(`{"zone":"%s","item":"%s","qty":%d}`, v1, v5, v23.F3[v5])
			ctx.GetStub().PutState(v7, []byte(v8))
		}

		v9, _ := json.Marshal(v23)
		ctx.GetStub().PutState(v3, v9) 
	}

	return nil
}

func (w *StructA) FuncB(ctx contractapi.TransactionContextInterface, v28 string) error {
	var v24 map[string]map[string]int
	err := json.Unmarshal([]byte(v28), &v24)
	if err != nil {
		return fmt.Errorf("invalid zone data: %v", err)
	}

	var v25 []string

	for v1, v10 := range v24 {
		var v26 []string

		for v5, v11 := range v10 {
			v26 = append(v26, fmt.Sprintf("%s=%d", v5, v11))
		}

		v12 := fmt.Sprintf("[%s: %s]", v1, strings.Join(v26, ","))
		v25 = append(v25, v12)
	}

	v13 := strings.Join(v25, " | ")
	return ctx.GetStub().PutState("WAREHOUSE_REPORT", []byte(v13))
}

func (w *StructA) FuncC(ctx contractapi.TransactionContextInterface, v29 string, v30 []string) error {
	v14 := make(map[string]int)

	for _, v15 := range v30 {
		v3 := "ZONE_" + v15
		v4, err := ctx.GetStub().GetState(v3)
		if err != nil || v4 == nil {
			continue
		}

		var v23 StructB
		json.Unmarshal(v4, &v23)

		for v5, v11 := range v23.F3 {
			v14[v5] += v11
		}

		ctx.GetStub().DelState(v3) 
	}

	for v5, v16 := range v14 {
		v17 := v29 + "_ITEM_" + v5
		v18 := fmt.Sprintf(`{"zone":"%s","item":"%s","qty":%d}`, v29, v5, v16)
		ctx.GetStub().PutState(v17, []byte(v18)) 
	}

	v19 := StructB{F1: v29, F3: v14, F2: len(v14)}
	v20, _ := json.Marshal(v19)
	return ctx.GetStub().PutState("ZONE_"+v29, v20)
}

func main() {
	v21, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v21.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
