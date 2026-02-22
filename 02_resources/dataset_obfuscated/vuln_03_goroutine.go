package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1          string   `json:"f1"`
	F2       string   `json:"f2"`
	F3   int      `json:"f3"`
	F4      []string `json:"f4"`
	F5      string   `json:"f5"`
	F6  string   `json:"f6"`
}

func (v *StructA) FuncA(ctx contractapi.TransactionContextInterface, v11 string, v12 string) error {
	v1 := StructB{
		F1:        v11,
		F2:     v12,
		F3: 0,
		F4:    []string{},
		F5:    "OPEN",
	}

	v2, err := json.Marshal(v1)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v11, v2)
}

func (v *StructA) FuncB(ctx contractapi.TransactionContextInterface, v13 string, v14 string) error {
	v2, err := ctx.GetStub().GetState(v13)
	if err != nil {
		return fmt.Errorf("failed to read proposal: %v", err)
	}
	if v2 == nil {
		return fmt.Errorf("proposal %s does not exist", v13)
	}

	var v1 StructB
	json.Unmarshal(v2, &v1)

	if v1.F5 != "OPEN" {
		return fmt.Errorf("proposal is not open for voting")
	}

	v1.F3++
	v1.F4 = append(v1.F4, v14)

	v3, _ := json.Marshal(v1)
	return ctx.GetStub().PutState(v13, v3)
}

func (v *StructA) FuncC(ctx contractapi.TransactionContextInterface, v15 []string) error {
	var mu sync.Mutex
	var wg sync.WaitGroup
	v4 := make(map[string]int)

	for _, v5 := range v15 {
		wg.Add(1)
		
		go func(v11 string) {
			defer wg.Done()
			v2, err := ctx.GetStub().GetState(v11)
			if err != nil {
				return
			}
			if v2 == nil {
				return
			}

			var v1 StructB
			json.Unmarshal(v2, &v1)

			mu.Lock()
			v4[v11] = v1.F3
			mu.Unlock()
		}(v5)
	}

	wg.Wait()

	v6, _ := json.Marshal(v4)
	return ctx.GetStub().PutState("TALLY_RESULT", v6)
}

func (v *StructA) FuncD(ctx contractapi.TransactionContextInterface, v15 []string, v16 string) error {
	var wg sync.WaitGroup
	v7 := make(chan error, len(v15))

	for _, v5 := range v15 {
		wg.Add(1)
		
		go func(v11 string) {
			defer wg.Done()
			v2, err := ctx.GetStub().GetState(v11)
			if err != nil {
				v7 <- err
				return
			}
			if v2 == nil {
				return
			}

			var v1 StructB
			json.Unmarshal(v2, &v1)
			v1.F5 = v16

			v3, _ := json.Marshal(v1)
			
			err = ctx.GetStub().PutState(v11, v3)
			if err != nil {
				v7 <- err
			}
		}(v5)
	}

	wg.Wait()
	close(v7)

	for err := range v7 {
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *StructA) FuncE(ctx contractapi.TransactionContextInterface, v13 string, v17 string) error {
	
	go func() {
		v8 := map[string]string{
			"proposal": v13,
			"message":  v17,
		}
		v9, _ := json.Marshal(v8)
		ctx.GetStub().SetEvent("NOTIFY", v9) 
	}()

	return nil
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
