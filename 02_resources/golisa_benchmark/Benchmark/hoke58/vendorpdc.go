// peer chaincode install -n vendor -v 1.0 -p github.com/chaincode/hoke/
// peer chaincode instantiate -o orderer.example.com:7050 --tls --cafile $ORDERER_CA -C mychannel -n vendor -v 1.0 -c '{"Args":["init"]}' -P "OR('Org1MSP.member','Org2MSP.member')" --collections-config $GOPATH/src/github.com/chaincode/hoke/collections_config.json
// export VENDOR=$(echo -n "{\"Name\":\"test0\",\"Project\":\"supplychain\",\"Status\":\"yes\",\"Expiry\":\"2020-05-01\",\"Price\":6666}" | base64 | tr -d \\n)
// export VENDOR=$(echo -n "{\"Name\":\"test1\",\"Project\":\"supplychain\",\"Status\":\"yes\",\"Expiry\":\"2020-05-01\",\"Price\":7777}" | base64 | tr -d \\n)
// export VENDOR=$(echo -n "{\"Name\":\"test2\",\"Project\":\"supplychain\",\"Status\":\"yes\",\"Expiry\":\"2020-05-01\",\"Price\":8888}" | base64 | tr -d \\n)
// export VENDOR=$(echo -n "{\"Name\":\"test3\",\"Project\":\"supplychain\",\"Status\":\"no\",\"Expiry\":\"2020-05-01\",\"Price\":9999}" | base64 | tr -d \\n)
// export VENDOR=$(echo -n "{\"Name\":\"test4\",\"Project\":\"supplychain\",\"Status\":\"no\",\"Expiry\":\"2020-05-01\",\"Price\":55555}" | base64 | tr -d \\n)
// export VENDOR=$(echo -n "{\"Name\":\"test5\",\"Project\":\"supplychain\",\"Status\":\"ok\",\"Expiry\":\"2020-05-01\",\"Price\":6666}" | base64 | tr -d \\n)
// peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n vendor -c '{"Args":["putVendor"]}'  --transient "{\"vendor\":\"$VENDOR\"}"
// peer chaincode query -C mychannel -n vendor -c '{"Args":["getVendor","test0"]}'
// peer chaincode query -C mychannel -n vendor -c '{"Args":["getVendorPrice","test0"]}'
// peer chaincode query -C mychannel -n vendor -c '{"Args":["getVendorByRange","test0","test2"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type Vendor struct {
	Name    string `json:"Name"`    // 供应商名字
	Project string `json:"Project"` // 项目
	Status  string `json:"Status"`  // 状态
	Expiry  string `json:"Expiry"`  // 有效期
}

type vendorPrice struct {
	Vendor
	Price float64 `json:"Price"` // 价格
}

type vendorChaincode struct {
}

func main() {
	err := shim.Start(new(vendorChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s\n", err)
	}
}

func (t *vendorChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init Chaincode...")
	return shim.Success(nil)
}

func (t *vendorChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()
	fmt.Printf("stub.GetFunctionAndParameters function= %v, args=%v\n", function, args)

	fmt.Printf("Invoke %v start\n", function)
	defer fmt.Printf("Invoke %v stop\n", function)
	switch function {
	case "putVendor":
		return t.putVendor(stub, args)
	case "getVendor":
		return t.getVendor(stub, args)
	case "getVendorPrice":
		return t.getVendorPrice(stub, args)
	case "getVendorByRange":
		return t.getVendorByRange(stub, args)
	default:
		fmt.Printf("Invoke %v not find\n", function)
		return shim.Error("Received unknown function invocation")
	}
}

func (t *vendorChaincode) putVendor(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("args=", args)
	type vendorTransientInput struct {
		Name    string  `json:"Name"`
		Project string  `json:"Project"`
		Status  string  `json:"Status"`
		Expiry  string  `json:"Expiry"`
		Price   float64 `json:"Price"`
	}

	fmt.Println("- start init vendor")

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Private data must be passed in transient map.")
	}

	transMap, err := stub.GetTransient()
	fmt.Printf("transMap=%v, err=%v\n", transMap, err)

	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["vendor"]; !ok {
		return shim.Error("vendor must be a key in the transient map")
	}

	if len(transMap["vendor"]) == 0 {
		return shim.Error("vendor value in the transient map must be a non-empty JSON string")
	}

	var vendorInput vendorTransientInput
	err = json.Unmarshal(transMap["vendor"], &vendorInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(transMap["vendor"]))
	}

	if len(vendorInput.Name) == 0 {
		return shim.Error("name field must be a non-empty string")
	}
	if len(vendorInput.Project) == 0 {
		return shim.Error("project field must be a non-empty string")
	}
	if len(vendorInput.Status) == 0 {
		return shim.Error("status field must be a positive integer")
	}
	if len(vendorInput.Expiry) == 0 {
		return shim.Error("expiry field must be a non-empty string")
	}
	if vendorInput.Price <= 0 {
		return shim.Error("price field must be a positive integer")
	}
	// vendor collection
	vendor := &Vendor{
		Name:    vendorInput.Name,
		Project: vendorInput.Project,
		Status:  vendorInput.Status,
		Expiry:  vendorInput.Expiry,
	}
	vendorJSONasBytes, err := json.Marshal(vendor)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutPrivateData("vendor", vendorInput.Name, vendorJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// vendorPrice collection
	vendorPrice := &vendorPrice{
		Vendor: Vendor{
			Name:    vendorInput.Name,
			Project: vendorInput.Project,
			Status:  vendorInput.Status,
			Expiry:  vendorInput.Expiry,
		},
		Price: vendorInput.Price,
	}
	vendorPriceBytes, err := json.Marshal(vendorPrice)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutPrivateData("vendorPrice", vendorInput.Name, vendorPriceBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the vendor to enable status-based range queries, e.g. return all vendors that status is ok ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~status~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~status~*
	indexName := "status~name"
	statusNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{vendor.Status, vendor.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the vendor.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutPrivateData("vendor", statusNameIndexKey, value)

	fmt.Println("- end init marble")
	return shim.Success(nil)
}

func (t *vendorChaincode) getVendor(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetPrivateData("vendor", name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Vendor does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

func (t *vendorChaincode) getVendorPrice(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetPrivateData("vendorPrice", name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get vendor price for " + name + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Vendor price does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

func (t *vendorChaincode) getVendorByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetPrivateDataByRange("vendor", startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getMarblesByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
