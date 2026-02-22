package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1    string  `json:"f1"`
	F2   string  `json:"f2"`
	F3   float64 `json:"f3"`
	F4 string `json:"f4"`
}

type StructC struct {
	F4    string  `json:"f4"`
	F5    int     `json:"f5"`
	F6   float64 `json:"f6"`
	F7 string `json:"f7"`
	F8   float64 `json:"f8"`
}

func (a *StructA) FuncA(ctx contractapi.TransactionContextInterface, v19 string, v20 string, v21 float64, v22 string) error {
	v1 := StructB{
		F1:     v19,
		F2:    v20,
		F3:    v21,
		F4: v22,
	}

	v2, err := json.Marshal(v1)
	if err != nil {
		return err
	}

	v3, err := ctx.GetStub().CreateCompositeKey("Bid", []string{v22, v19})
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v3, v2)
}

func (a *StructA) FuncB(ctx contractapi.TransactionContextInterface, v22 string) error {
	
	v4, err := ctx.GetStub().GetStateByPartialCompositeKey("Bid", []string{v22})
	if err != nil {
		return fmt.Errorf("failed to query bids: %v", err)
	}
	defer v4.Close()

	var v17 float64
	var v18 string
	v5 := 0
	v6 := 0.0

	for v4.HasNext() {
		v7, err := v4.Next()
		if err != nil {
			return err
		}

		var v1 StructB
		json.Unmarshal(v7.Value, &v1)

		v5++
		v6 += v1.F3

		if v1.F3 > v17 {
			v17 = v1.F3
			v18 = v1.F2
		}
	}

	v8 := StructC{
		F4:     v22,
		F5:     v5,
		F6:    v17,
		F7: v18,
		F8:    v6,
	}

	v9, _ := json.Marshal(v8)
	return ctx.GetStub().PutState("AUCTION_RESULT_"+v22, v9)
}

func (a *StructA) FuncC(ctx contractapi.TransactionContextInterface, v23 string) error {
	
	v10, err := ctx.GetStub().GetState(v23)
	if err != nil {
		return fmt.Errorf("failed to read counter: %v", err)
	}

	v11 := 0
	if v10 != nil {
		v11, err = strconv.Atoi(string(v10))
		if err != nil {
			return fmt.Errorf("invalid counter value: %v", err)
		}
	}

	v11++
	return ctx.GetStub().PutState(v23, []byte(strconv.Itoa(v11)))
}

func (a *StructA) FuncD(ctx contractapi.TransactionContextInterface, v24 string, v25 string, v21 float64) error {
	
	v12, err := ctx.GetStub().GetState(v24)
	if err != nil {
		return err
	}
	if v12 == nil {
		return fmt.Errorf("sender %s not found", v24)
	}

	v13, _ := strconv.ParseFloat(string(v12), 64)

	if v13 < v21 {
		return fmt.Errorf("insufficient funds")
	}

	v14, _ := ctx.GetStub().GetState(v25)
	v15 := 0.0
	if v14 != nil {
		v15, _ = strconv.ParseFloat(string(v14), 64)
	}

	ctx.GetStub().PutState(v24, []byte(strconv.FormatFloat(v13-v21, 'f', 2, 64)))
	return ctx.GetStub().PutState(v25, []byte(strconv.FormatFloat(v15+v21, 'f', 2, 64)))
}

func main() {
	v16, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v16.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
