package main

import (
	"bytes"
	"fmt"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

type MyChaincode struct {
}

func (t *MyChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *MyChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *MyChaincode) doGetID(stub shim.ChaincodeStubInterface) pb.Response {
	id, err := cid.GetID(stub)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(id))
}

func (t *MyChaincode) doGetTxID(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf(`GetTxID("%s") \n<Begin>\n`, key)
	txId := stub.GetTxID()

	fmt.Println("<End>")
	return shim.Success([]byte(txId))
}

func (t *MyChaincode) doGetTxTimestamp(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf(`GetTxTimestamp("%s") \n<Begin>\n`, key)
	txTimestamp, err := stub.GetTxTimestamp()

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("<End>")
	return shim.Success([]byte(txTimestamp.String()))
}

func (t *MyChaincode) doGetCreator(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Printf(`GetCreator() \n<Begin>\n`)
	byts, err := stub.GetCreator()

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("<End>")
	return shim.Success(byts)
}

func (t *MyChaincode) doGetChannelID(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf(`GetChannelID("%s") \n<Begin>\n`, key)
	channelID := stub.GetChannelID()

	fmt.Println("<End>")
	return shim.Success([]byte(channelID))
}

func (t *MyChaincode) doGetFunctionAndParameters(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Printf(`GetFunctionAndParameters() \n<Begin>\n`)
	functionName, args := stub.GetFunctionAndParameters()
	msg := fmt.Sprintf(`{"function_name": "%s", "args": ["%s", "%s"]}`, functionName, args[0], args[1])
	fmt.Println("<End>")
	return shim.Success([]byte(msg))
}

func (t *MyChaincode) doGetState(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf(`GetState("%s") \n<Begin>\n`, key)
	byts, err := stub.GetState(key)

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("<End>")
	return shim.Success(byts)
}

func (t *MyChaincode) doPutState(stub shim.ChaincodeStubInterface, key string, val string) pb.Response {
	fmt.Printf(`PutState("%s", "%s") \n<Begin>\n`, key, val)
	err := stub.PutState(key, []byte(val))

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("<End>")
	return shim.Success(nil)
}

func (t *MyChaincode) doDelState(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf(`DelState("%s") \n<Begin>\n`, key)
	err := stub.DelState(key)

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("<End>")
	return shim.Success(nil)
}

func (t *MyChaincode) doGetStateByRange(stub shim.ChaincodeStubInterface, key0, key1 string) pb.Response {
	fmt.Printf(`GetStateByRange("%s","%s") \n<Begin>\n`, key0, key1)
	resultIt, err := stub.GetStateByRange(key0, key1)

	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultIt.Close()

	var buff bytes.Buffer
	buff.WriteString("[")
	for resultIt.HasNext() {
		response, err := resultIt.Next()
		if err != nil {
			fmt.Printf(`[Error] Iterator mistake. Ignored.(%s)`, err.Error())
			continue
		}
		if buff.Len() > 1 {
			buff.WriteString(",")
		}
		buff.WriteString(string(response.Value))
	}
	buff.WriteString("]")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doGetStateByRangeWithPagination(stub shim.ChaincodeStubInterface, key0, key1 string, pageSize int32, bookmark string) pb.Response {
	fmt.Printf(`GetStateByRangeWithPagination("%s","%s","%d","%s") \n<Begin>\n`, key0, key1, pageSize, bookmark)
	resultIt, metadata, err := stub.GetStateByRangeWithPagination(key0, key1, pageSize, bookmark)

	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultIt.Close()

	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf(`{"page_title": {"count": %d, "bookmark": "%s"},\n`, metadata.FetchedRecordsCount, metadata.Bookmark))
	buff.WriteString(`"page_data":[`)
	for resultIt.HasNext() {
		response, err := resultIt.Next()
		if err != nil {
			fmt.Printf(`[Error] Iterator mistake. Ignored.(%s)`, err.Error())
			continue
		}
		if buff.Len() > 1 {
			buff.WriteString(",")
		}
		buff.WriteString(string(response.Value))
	}
	buff.WriteString("]}")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doCreateCompositeKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Printf(`CreateCompositeKey("%s","%s", "%s") \n<Begin>\n`, args[0], args[1], args[2])
	indexName := "sex~name"
	jsonString := fmt.Sprintf(`{"name": "%s", "sex":"%s", "age": %s}`, args[0], args[1], args[2])
	err := stub.PutState(args[0], []byte(jsonString))
	if err != nil {
		return shim.Error(err.Error())
	}

	sexNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{args[1], args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(sexNameIndexKey, []byte{0x00})
	if err != nil {
		return shim.Error("Failed to add state:" + err.Error())
	}
	fmt.Println("<End>")
	return shim.Success(nil)
}

func (t *MyChaincode) doGetStateByPartialCompositeKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Printf(`GetStateByPartialCompositeKey("%s") \n<Begin>\n`, args[0])
	indexName := "sex~name"
	resultIt, err := stub.GetStateByPartialCompositeKey(indexName, []string{args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultIt.Close()

	var buff bytes.Buffer
	buff.WriteString("[")
	for resultIt.HasNext() {
		response, err := resultIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedName := compositeKeyParts[1]
		byts, err := stub.GetState(returnedName)
		if err == nil {
			if buff.Len() > 1 {
				buff.WriteString(",")
			}
			buff.WriteString(string(byts))
		}
	}
	buff.WriteString("]")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doGetStateByPartialCompositeKeyWithPagination(stub shim.ChaincodeStubInterface, args []string, pageSize int32, bookmark string) pb.Response {
	fmt.Printf(`GetStateByPartialCompositeKey("%s") \n<Begin>\n`, args[0])
	indexName := "sex~name"
	resultIt, metadata, err := stub.GetStateByPartialCompositeKeyWithPagination(indexName, []string{args[0]}, pageSize, bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultIt.Close()

	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf(`{"page_title": {"count": %d, "bookmark": "%s"},\n`, metadata.FetchedRecordsCount, metadata.Bookmark))
	buff.WriteString(`"page_data":[`)

	for resultIt.HasNext() {
		response, err := resultIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedName := compositeKeyParts[1]
		byts, err := stub.GetState(returnedName)
		if err == nil {
			if buff.Len() > 1 {
				buff.WriteString(",")
			}
			buff.WriteString(string(byts))
		}
	}
	buff.WriteString("]}")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doGetQueryResult(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Printf(`GetQueryResult("%s") \n<Begin>\n`, args[0])
	age, _ := strconv.Atoi(args[0])
	queryString := fmt.Sprintf(`{"selector":{"age":{ "$gte": %d }}}`, age)
	queryIt, err := stub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer queryIt.Close()

	var buff bytes.Buffer
	buff.WriteString("[")

	for queryIt.HasNext() {
		queryResult, err := queryIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if buff.Len() > 1 {
			buff.WriteString(",")
		}
		buff.WriteString(string(queryResult.GetValue()))
	}

	buff.WriteString("]")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doGetQueryResultWithPagination(stub shim.ChaincodeStubInterface, args []string, pageSize int32, bookmark string) pb.Response {
	fmt.Printf(`GetQueryResult("%s") \n<Begin>\n`, args[0])
	age, _ := strconv.Atoi(args[0])
	queryString := fmt.Sprintf(`{"selector":{"age":{ "$gte": %d }}}`, age)
	queryIt, metadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer queryIt.Close()

	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf(`{"page_title": {"count": %d, "bookmark": "%s"},\n`, metadata.FetchedRecordsCount, metadata.Bookmark))
	buff.WriteString(`"page_data":[`)

	for queryIt.HasNext() {
		queryResult, err := queryIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if buff.Len() > 1 {
			buff.WriteString(",")
		}
		buff.WriteString(string(queryResult.GetValue()))
	}

	buff.WriteString("]}")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doGetHistoryForKey(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf(`GetHistoryForKey("%s") \n<Begin>\n`, key)
	historyIt, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer historyIt.Close()

	var buff bytes.Buffer
	buff.WriteString("[")

	for historyIt.HasNext() {
		historyResult, err := historyIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if buff.Len() > 1 {
			buff.WriteString(",")
		}
		buff.WriteString(fmt.Sprintf(
			`{"tx_id": "%s", 
				"time": "%v", 
				"value": "%s", 
				"del_flg": "%v"
			}`, historyResult.TxId, historyResult.Timestamp, historyResult.Value, historyResult.IsDelete))
	}

	buff.WriteString("]")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doGetPrivateData(stub shim.ChaincodeStubInterface, collection, key string) pb.Response {
	fmt.Printf(`GetGetPrivateData("%s", "%s") \n<Begin>\n`, collection, key)
	byts, err := stub.GetPrivateData(collection, key)

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("<End>")
	return shim.Success(byts)
}

func (t *MyChaincode) doPutPrivateData(stub shim.ChaincodeStubInterface, collection, key, val string) pb.Response {
	fmt.Printf(`PutPrivateData("%s", "%s", "%s") \n<Begin>\n`, collection, key, val)
	err := stub.PutPrivateData(collection, key, []byte(val))

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("<End>")
	return shim.Success(nil)
}

func (t *MyChaincode) doDelPrivateData(stub shim.ChaincodeStubInterface, collection, key string) pb.Response {
	fmt.Printf(`DelPrivateData("%s", "%s") \n<Begin>\n`, collection, key)
	err := stub.DelPrivateData(collection, key)

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("<End>")
	return shim.Success(nil)
}

func (t *MyChaincode) doGetPrivateDataByRange(stub shim.ChaincodeStubInterface, collection, key0, key1 string) pb.Response {
	fmt.Printf(`GetPrivateDataByRange("%s","%s","%s") \n<Begin>\n`, collection, key0, key1)
	resultIt, err := stub.GetPrivateDataByRange(collection, key0, key1)

	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultIt.Close()

	var buff bytes.Buffer
	buff.WriteString("[")
	for resultIt.HasNext() {
		response, err := resultIt.Next()
		if err != nil {
			fmt.Printf(`[Error] Iterator mistake. Ignored.(%s)`, err.Error())
			continue
		}
		if buff.Len() > 1 {
			buff.WriteString(",")
		}
		buff.WriteString(string(response.Value))
	}
	buff.WriteString("]")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doGetPrivateDataByPartialCompositeKey(stub shim.ChaincodeStubInterface, collection string, args []string) pb.Response {
	fmt.Printf(`GetPrivateDataByPartialCompositeKey("%s") \n<Begin>\n`, args[0])
	indexName := "sex~name"
	resultIt, err := stub.GetPrivateDataByPartialCompositeKey(collection, indexName, []string{args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultIt.Close()

	var buff bytes.Buffer
	buff.WriteString("[")
	for resultIt.HasNext() {
		response, err := resultIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedName := compositeKeyParts[1]
		byts, err := stub.GetState(returnedName)
		if err == nil {
			if buff.Len() > 1 {
				buff.WriteString(",")
			}
			buff.WriteString(string(byts))
		}
	}
	buff.WriteString("]")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doGetPrivateDataQueryResult(stub shim.ChaincodeStubInterface, collection string, args []string) pb.Response {
	fmt.Printf(`GetPrivateDataQueryResult("%s", "%s") \n<Begin>\n`, collection, args[0])
	age, _ := strconv.Atoi(args[0])
	queryString := fmt.Sprintf(`{"selector":{"age":{ "$gte": %d }}}`, age)
	queryIt, err := stub.GetPrivateDataQueryResult(collection, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer queryIt.Close()

	var buff bytes.Buffer
	buff.WriteString("[")

	for queryIt.HasNext() {
		queryResult, err := queryIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if buff.Len() > 1 {
			buff.WriteString(",")
		}
		buff.WriteString(string(queryResult.GetValue()))
	}

	buff.WriteString("]")
	fmt.Println("<End>")
	return shim.Success(buff.Bytes())
}

func (t *MyChaincode) doGetStateValidationParameter(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`GetStateValidationParameter() \n<Begin>`)
	key := "my_key"
	peBytes, err := stub.GetStateValidationParameter(key)
	fmt.Println("<End>")
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(peBytes)
}

func (t *MyChaincode) doGetPrivateDataValidationParameter(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`GetPrivateDataValidationParameter() \n<Begin>`)
	collection := "my_collection"
	key := "my_key"
	peBytes, err := stub.GetPrivateDataValidationParameter(collection, key)
	fmt.Println("<End>")
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(peBytes)
}

func (t *MyChaincode) doSetStateValidationParameter(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`SetStateValidationParameter() \n<Begin>`)
	key := "my_key"
	ep := `OR(Org1.member, Org2.member)`
	err := stub.SetStateValidationParameter(key, []byte(ep))
	fmt.Println("<End>")
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *MyChaincode) doSetPrivateDataValidationParameter(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`SetPrivateDataValidationParameter() \n<Begin>`)
	collection := "my_collection"
	key := "my_key"
	ep := `OR(Org1.member, Org2.member)`
	err := stub.SetPrivateDataValidationParameter(collection, key, []byte(ep))
	fmt.Println("<End>")
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *MyChaincode) doInvokeChaincode(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`InvokeChaincode() \n<Begin>`)
	chaincodeName := "exmple03"
	funcName := "query"
	args := []string{funcName, "a", "b"}
	channelName := "channel02"
	fmt.Printf(`InvokeChaincode("%s, %s, %s") >`, chaincodeName, args, channelName)
	byteArgs := util.ArrayToChaincodeArgs(args)
	fmt.Println("<End>")
	return stub.InvokeChaincode(chaincodeName, byteArgs, channelName)
}

func (t *MyChaincode) doSetEvent(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`SetEvent() \n<Begin>`)
	topicName := "car_owner_changed"
	msg := fmt.Sprintf("The owner of car (%s) has been changed successfully.", "xxxxxx")

	err := stub.SetEvent(topicName, []byte(msg))
	fmt.Println("<End>")

	if err != nil {
		return shim.Success(nil)
	} else {
		return shim.Error(err.Error())
	}
}

func (t *MyChaincode) doAssertAttributeValue(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`CID.AssertAttributeValue() \n<Begin>`)
	attrName := "email"
	attrVal := "user1@org1"
	err := cid.AssertAttributeValue(stub, attrName, attrVal)
	if err != nil {
		fmt.Printf(`CID.AssertAttributeValue('%s', '%s') failed. Error: '%s'`, attrName, attrVal, err.Error())
		return shim.Error(err.Error())
	}
	fmt.Printf(`CID.AssertAttributeValue('%s', '%s') successfully.`, attrName, attrVal)
	return shim.Success(nil)
}

func (t *MyChaincode) doGetAttributeValue(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`CID.GetAttributeValue() \n<Begin>`)
	attrName := "admin"
	attrVal, ok, err := cid.GetAttributeValue(stub, attrName)
	if err != nil {
		fmt.Printf(`CID.GetAttributeValue('%s') failed. Error: '%s'`, attrName, err.Error())
		return shim.Error(err.Error())
	} else if !ok {
		fmt.Printf(`CID.GetAttributeValue('%s') failed. Not found.`, attrName)
		return shim.Success(nil)
	}

	fmt.Printf(`CID.GetAttributeValue('%s') successfully. Val: '%s'`, attrName, attrVal)
	return shim.Success(nil)
}

func (t *MyChaincode) doGetMSPID(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`CID.GetMSPID() \n<Begin>`)

	mspID, err := cid.GetMSPID(stub)
	if err != nil {
		fmt.Printf(`CID.GetMSPID() failed. Error: '%s'`, err.Error())
		return shim.Error(err.Error())
	}

	fmt.Printf(`CID.GetMSPID() successfully. MSPID: '%s'`, mspID)
	return shim.Success(nil)
}

func (t *MyChaincode) doGetCIDID(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(`CID.GetID() \n<Begin>`)

	mspID, err := cid.GetID(stub)
	if err != nil {
		fmt.Printf(`CID.GetID() failed. Error: '%s'`, err.Error())
		return shim.Error(err.Error())
	}

	fmt.Printf(`CID.GetID() successfully. ID: '%s'`, mspID)
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(MyChaincode))
	if err != nil {
		fmt.Printf("Error starting MyChaincode: %s", err)
	}
}
