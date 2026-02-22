package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type SearchShipment struct {
	Shipments []Shipment
}

type Shipment struct {
	ID      string   `json:id`
	Product string   `json:product`
	Price   string   `json:price`
	Fine    string   `json:fine`
	Status  string   `json:status`
	Range   []Ranges `json:ranges`
}

type Ranges struct {
	Param 	 string `json:param`
	MinValue string `json:minValue`
	MaxValue string `json:maxValue`
}

type Logs struct {
	Param 	string `json:param`
	Measure string `json:measure`
}

func Success(rc int32, doc string, payload []byte) peer.Response {
	return peer.Response{
		Status:  rc,
		Message: doc,
		Payload: payload,
	}
}

func Error(rc int32, doc string) peer.Response {
	logger.Errorf("Error %d = %s", rc, doc)
	return peer.Response{
		Status:  rc,
		Message: doc,
	}
}

var logger = shim.NewLogger("chaincode")

type PerishableNet struct {
}

func main() {
	if err := shim.Start(new(PerishableNet)); err != nil {
		fmt.Printf("Main: Error starting chaincode: %s", err)
	}
}

func (cc *PerishableNet) Init(stub shim.ChaincodeStubInterface) peer.Response {
	if _, args := stub.GetFunctionAndParameters(); len(args) > 0 {
		return Error(http.StatusBadRequest, "Init: Incorrect number of arguments; no arguments were expected.")
	}
	return Success(http.StatusOK, "OK", nil)
}

func (cc *PerishableNet) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "get":
		return cc.get(stub, args)
	case "create":
		return cc.create(stub, args)
	case "log":
		return cc.log(stub, args)
	case "history":
		return cc.history(stub, args)
	case "updatestat":
		return cc.updatestat(stub, args)
	case "getall":
		return cc.getall(stub, args)
	default:
		logger.Warningf("Invoke('%s') invalid!", function)
		return Error(http.StatusNotImplemented, "Invalid method! Valid methods are 'create|update|read|history'!")
	}
}

func (cc *PerishableNet) get(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	id := strings.ToLower(args[0])

	if value, err := stub.GetState(id); err != nil || value == nil {
		return Error(http.StatusNotFound, "Not Found")
	} else {
		return Success(http.StatusOK, "OK", value)
	}
}

func (cc *PerishableNet) create(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	id := strings.ToLower(args[0])

	if value, err := stub.GetState(id); !(err == nil && value == nil) {
		return Error(http.StatusConflict, "Shipment Exists")
	}

	jsonShipment := strings.ToLower(args[1])
	var shipment Shipment
	json.Unmarshal([]byte(jsonShipment), &shipment)

	if err := stub.PutState(shipment.ID, []byte(jsonShipment)); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusCreated, "Shipment Created", nil)
}

func (cc *PerishableNet) log(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	id := strings.ToLower(args[0])

	if value, err := stub.GetState(id); err != nil || value == nil {
		return Error(http.StatusNotFound, "Not Found")
	}
	
	jsonLog := strings.ToLower(args[1])
	if err := stub.PutState("log-" + id, []byte(jsonLog)); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusAccepted, "Log Updated", nil)
}

func (cc *PerishableNet) updatestat(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	/**** изменить статус поставки */
	// получить ID поставки (первый аргумент) и преобразовать его в строковый тип (strings.ToLower)
	// по ID проверить существование поставки с данным ID и если поставка существует получить объект в формате json  (GetState)
	// создать переменую типа Shipment
	// скопировать объект поставки в json формате в пермеменную типа Shipment (json.Unmarshal)
	// изменить статус поставки на новый (второй аргумент) и сформировать поставки в формате json с новым статусом (json.Marshal)
	// записать поставку в канал (PutState)
	// вернуть код успеха или ошибки (return Success | return Error)
	// не забыть добавить вызов функции в метод Invoke	
	/****/
}

func (cc *PerishableNet) history(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	id := strings.ToLower(args[0])

	resultsIterator, err := stub.GetHistoryForKey("log-" + id)
	if err != nil {
		return Error(http.StatusNotFound, "Not Found")
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("{ \"values\": [")
	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()

		var log Logs
		json.Unmarshal(it.Value, &log)

		if log.Param == args[1] {
			if buffer.Len() > 15 {
				buffer.WriteString(",")
			}

			buffer.WriteString("{\"timestamp\":\"")
			buffer.WriteString(time.Unix(it.Timestamp.Seconds, int64(it.Timestamp.Nanos)).Format(time.Stamp))

			buffer.WriteString("\", \"measure\":\"")
			buffer.WriteString(log.Measure)
			buffer.WriteString("\"}")
		}
	}
	buffer.WriteString("]}")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}

func (cc *PerishableNet) getall(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// queryString := "{ \"selector\": { \"product\":  { \"$in\": [\"bananas\", \"apples\", \"pears\", \"oranges\"] } } }"
	queryString := `{
		"selector": {
			"$or": [
				{
					"Product": {
						"$in": ["bananas", "apples", "pears", "oranges"]
					}
				},
				{
					"product": {
						"$in": ["bananas", "apples", "pears", "oranges"]
					}
				}
			]
		}
	}`

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	var results SearchShipment
	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()
		jsonShipment, _ := stub.GetState(it.Key)
		var shipment Shipment
		json.Unmarshal(jsonShipment, &shipment)
		results.Shipments = append(results.Shipments, shipment)
	}
	resultJson, _ := json.Marshal(results.Shipments)
	return Success(200, "OK", resultJson)
}