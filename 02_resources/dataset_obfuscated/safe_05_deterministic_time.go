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
	F1     string `json:"f1"`
	F2 string `json:"f2"`
	F3   string `json:"f3"`
	F4   string `json:"f4"`
	F5  string `json:"f5"`
	F6     string `json:"f6"`
}

func (c *StructA) FuncA(ctx contractapi.TransactionContextInterface, v12 string, v13 string, v14 string) error {
	v1, err := c.FuncD(ctx, v12)
	if err != nil {
		return err
	}
	if v1 {
		return fmt.Errorf("certificate %s already exists", v12)
	}

	v2 := time.Now()

	v3, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	v4 := time.Unix(v3.Seconds, int64(v3.Nanos))

	v5 := StructB{
		F1:     v12,
		F2: v13,
		F3:   v14,
		F4:   v4.Format(time.RFC3339),                          
		F5:  v4.AddDate(2, 0, 0).Format(time.RFC3339),         
		F6:     "ACTIVE",
	}

	v6, err := json.Marshal(v5)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(v12, v6)
	if err != nil {
		return err
	}

	v7 := time.Since(v2)
	fmt.Printf("[PERF] IssueCertificate(%s) took %v\n", v12, v7)

	return nil
}

func (c *StructA) FuncB(ctx contractapi.TransactionContextInterface, v12 string) error {
	v6, err := ctx.GetStub().GetState(v12)
	if err != nil {
		return fmt.Errorf("failed to read certificate %s: %v", v12, err)
	}
	if v6 == nil {
		return fmt.Errorf("certificate %s does not exist", v12)
	}

	var v5 StructB
	err = json.Unmarshal(v6, &v5)
	if err != nil {
		return err
	}

	v8 := time.Now()
	fmt.Printf("[LOG] RenewCertificate called at local time: %s\n", v8.Format(time.RFC3339))

	v3, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	
	v8 = time.Unix(v3.Seconds, int64(v3.Nanos))

	v5.F5 = v8.AddDate(2, 0, 0).Format(time.RFC3339) 
	v5.F6 = "RENEWED"

	v9, err := json.Marshal(v5)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v12, v9)
}

func (c *StructA) FuncC(ctx contractapi.TransactionContextInterface, v12 string, v15 string) error {
	v6, err := ctx.GetStub().GetState(v12)
	if err != nil {
		return fmt.Errorf("failed to read certificate %s: %v", v12, err)
	}
	if v6 == nil {
		return fmt.Errorf("certificate %s does not exist", v12)
	}

	var v5 StructB
	err = json.Unmarshal(v6, &v5)
	if err != nil {
		return err
	}

	fmt.Printf("[REVOKE] Certificate %s revoked at %s. Reason: %s\n",
		v12, time.Now().Format(time.RFC3339), v15)

	v5.F6 = "REVOKED" 

	v9, err := json.Marshal(v5)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v12, v9)
}

func (c *StructA) FuncD(ctx contractapi.TransactionContextInterface, v16 string) (bool, error) {
	v10, err := ctx.GetStub().GetState(v16)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return v10 != nil, nil
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
