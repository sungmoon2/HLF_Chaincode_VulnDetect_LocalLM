package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// WabcoChaincode  Chaincode implementation
type WabcoChaincode struct {
}
type Product struct {
	LoadingList           string    `json:"loading_list"`
	DeliveryNbr           string    `json:"delivery_nbr"`
	Vendor                string    `json:"vendor"`
	Recipient             string    `json:"recipient"`
	Street                string    `json:"street"`
	Country               string    `json:"country"`
	Postal                string    `json:"postal"`
	City                  string    `json:"city"`
	NbrPackages           int       `json:"nbr_packages"`
	Pallets               string    `json:"pallets"`
	Weight                float32   `json:"weight"`
	GrossWeight           float32   `json:"gross_weight"`
	Volume                float32   `json:"volume"`
	VolumeOn              float32   `json:"volume_on"`
	Truck                 string    `json:"truck"`
	DeliveryDate          time.Time `json:"delivery_date"`
	NB                    int       `json:"nb"`
	NBR                   int       `json:"nbr"`
	TransportationCharges string    `json:"transportation_charges"`
	GoodsRreceiptDate     time.Time `json:"goods_receipt_date"`
}

func (t *WabcoChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	fmt.Println("hello Wabco")
	return shim.Success([]byte("Chaincode initialize successfully"))
}

// this is the invoke method thid will get execute at the time ogf invocation

func (t *WabcoChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "addProduct" {
		return addProduct(stub, args)
	} else if function == "getProduct" {
		return getProduct(stub, args)
	}
	fmt.Println("hello invoke")
	return shim.Error("Pleas eneter a valid function name!!!!!!!!!!!")
}

func addProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Argument for insert should be equal to 1")
	}

	product := Product{}
	productparseerr := json.Unmarshal([]byte(args[0]), &product)
	if productparseerr != nil {
		return shim.Error(productparseerr.Error())
	}
	key := stub.GetTxID()
	productbytes, productmarserr := json.Marshal(product)
	if productmarserr != nil {
		return shim.Error(productmarserr.Error())
	}
	fmt.Println(product)
	err0 := stub.PutState(key, productbytes)
	if err0 != nil {
		return shim.Error(err0.Error())
	}

	fmt.Println("Printed all the args as given")

	return shim.Success([]byte(key))

}

func getProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Please provide single parameter as key !!!!!!!!!!!!1")
	}
	returndebytes, err := stub.GetState(args[0])

	if err != nil {
		return shim.Error("Unable to fetch the given key something went wrong ")
	}
	return shim.Success(returndebytes)
}

// Initialaizing main

func main() {
	err := shim.Start(new(WabcoChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
