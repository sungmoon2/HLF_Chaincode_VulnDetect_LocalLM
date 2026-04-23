package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts

	"bytes"

	"math"
	"strconv"
	"time"
*/
import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

// Define the Smart Contract structure
type SmartContract struct {
}

/*
const (
	MAX_MOVIE  = 5
	MAX_SHOWS = 4
	MAX_TICKETS = 100
	MAX_WINDOWS = 4
)
*/
const (
	NEXT_SHOW_ID   = "NEXT_SHOW_ID"
	NEXT_TICKET_ID = "NEXT_TICKET_ID"
)

//doctypes
const (
	SHOW   = "SHOW"
	TICKET = "TICKET"
	WINDOW = "WINDOW"
	SODA   = "SODA"
)

type Theatre struct {
	TheatreNo      int    `json:"theatreNo"`
	TheatreName    string `json:"theatreName"`
	Windows        int    `json:"windows, omitempty"`
	TicketsPerShow int    `json:"ticketsPerShow, omitempty"`
	ShowsDaily     int    `json:"showsDaily, omitempty"`
	SodaStock      int    `json:"sodaStock, omitempty"`
	Halls          int    `json:"halls, omitempty"`
	DocType        string `json:"docType"`
}

type Window struct {
	WindowNo    int    `json:"windowNo"`
	TicketsSold int    `json:"ticketsSold"`
	DocType     string `json:"docType"`
}

type Ticket struct {
	TicketNo        int     `json:"ticketNo"`
	Show            Show    `json:"show"`
	Window          Window  `json:"window"`
	Quantity        int     `json:"quantity,number"`
	Amount          float64 `json:"amount,string"`
	CouponNumber    string  `json:"couponNumber"`
	CouponAvailed   bool    `json:"couponAvailed"`
	ExchangeAvailed bool    `json:"exchangeAvailed"`
	DocType         string  `json:"docType"`
}

type Show struct {
	ShowID    int    `json:"showID"`
	Movie     string `json:"movie"`
	ShowSlot  string `json:"showSlot"`
	Quantity  int    `json:"quantity,number"`
	HallNo    int    `json:"hallNo"`
	TheatreNo int    `json:"theatreNo"`
	DocType   string `json:"docType"`
}

type Soda struct {
	Stock        int    `json:"stock"`
	TicketNo     int    `json:"ticketNo"`
	CouponNumber string `json:"couponNumber"`
	DocType      string `json:"docType"`
}

type Property struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CreateShows struct {
	TheatreNo int    `json:"theatreNo"`
	Shows     []Show `json:"shows"`
}

// =========================================================================================
// The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
// Best practice is to have any Ledger initialization in separate function -- see initLedger()
// =========================================================================================
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {

	_, err := set(APIstub, NEXT_SHOW_ID, "0")
	_, err = set(APIstub, NEXT_TICKET_ID, "0")

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// =========================================================================================
// The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
// The calling application program has also specified the particular smart contract function to be called, with arguments
// =========================================================================================
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "registerTheatre" {
		return s.registerTheatre(APIstub, args)
	} else if function == "createShow" {
		return s.createShow(APIstub, args)
	} else if function == "purchaseTicket" {
		return s.purchaseTicket(APIstub, args)
	} else if function == "issueCoupon" {
		return s.issueCoupon(APIstub, args)
	} else if function == "availExchange" {
		return s.availExchange(APIstub, args)
	} else if function == "queryByString" {
		return s.queryByString(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.:" + function)
}

func (s *SmartContract) registerTheatre(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::registerTheatre:Start")
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	var theatre Theatre
	if err := json.Unmarshal([]byte(args[0]), &theatre); err != nil {
		fmt.Println("Cannot unmarshal theatre Object", err)
		return shim.Error(err.Error())
	}
	// Create unique theatre number & save theatre
	txnId := APIstub.GetTxID()
	var number int
	for _, c := range txnId {
		number = number + int(c)
	}
	theatre.TheatreNo = number
	theatre.DocType = "THEATRE"
	theatreAsBytes, _ := json.Marshal(theatre)
	err := APIstub.PutState("THEATRE"+strconv.Itoa(theatre.TheatreNo), theatreAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// create windows for the theatre
	for i := 1; i <= theatre.Windows; i++ {
		var window Window
		window.WindowNo = i
		window.TicketsSold = 0
		window.DocType = WINDOW
		windowAsBytes, _ := json.Marshal(window)
		err := APIstub.PutState("WINDOW"+strconv.Itoa(i), windowAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	fmt.Println("API::registerTheatre:End")
	return shim.Success([]byte("MovieTheatre Number:" + strconv.Itoa(theatre.TheatreNo)))
}

func (s *SmartContract) createShow(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::createShow:Start")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var createShows CreateShows

	if err := json.Unmarshal([]byte(args[0]), &createShows); err != nil {
		fmt.Println("Cannot unmarshal createShows Object", err)
		return shim.Error(err.Error())
	}

	showSeq, err := get(APIstub, NEXT_SHOW_ID)
	fmt.Println("Generating show for showSeq", showSeq)

	if err != nil {
		return shim.Error(err.Error())
	}

	shows := createShows.Shows
	var theatre Theatre
	theatreBytes, err := APIstub.GetState("THEATRE" + strconv.Itoa(createShows.TheatreNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(theatreBytes, &theatre)

	if len(shows) > theatre.Halls {
		return shim.Error("Number of Movies cannot exceed" + strconv.Itoa(theatre.Halls))
	}
	for _, show := range shows {

		for i := 1; i <= theatre.ShowsDaily; i++ {
			showSeq = showSeq + 1
			show.ShowID = +showSeq
			show.ShowSlot = strconv.Itoa(i)
			show.Quantity = theatre.TicketsPerShow
			show.TheatreNo = theatre.TheatreNo
			show.DocType = SHOW
			showAsBytes, _ := json.Marshal(show)
			err = APIstub.PutState("SHOW"+strconv.Itoa(show.ShowID), showAsBytes)
			if err != nil {
				return shim.Error(err.Error())
			}
		}
	}
	fmt.Println("saving current showSeq", showSeq)
	_, err = set(APIstub, NEXT_SHOW_ID, strconv.Itoa(showSeq))
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("API::createShow:End")
	return shim.Success([]byte(APIstub.GetTxID()))
}

func (s *SmartContract) purchaseTicket(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::purchaseTicket:Start")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var ticket Ticket

	if err := json.Unmarshal([]byte(args[0]), &ticket); err != nil {
		fmt.Println("Cannot unmarshal ticket Object", err)
		return shim.Error(err.Error())
	}

	ticketSeq, err := get(APIstub, NEXT_TICKET_ID)
	fmt.Println("Generating Ticket for ticketSeq", ticketSeq)

	if err != nil {
		return shim.Error(err.Error())
	}

	showBytes, err := APIstub.GetState("SHOW" + strconv.Itoa(ticket.Show.ShowID))
	if err != nil {
		return shim.Error(err.Error())
	}
	var show Show
	json.Unmarshal(showBytes, &show)

	windowBytes, err := APIstub.GetState("WINDOW" + strconv.Itoa(ticket.Window.WindowNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	var window Window
	json.Unmarshal(windowBytes, &window)
	// check the show for number of seats remaining
	if show.Quantity < 0 || show.Quantity-ticket.Quantity < 0 {
		return shim.Error("Seats Full for the requested show or Not enough seats as requested. Available:" + strconv.Itoa(show.Quantity))
	}

	show.Quantity = show.Quantity - ticket.Quantity
	window.TicketsSold = window.TicketsSold + ticket.Quantity
	fmt.Println(window.TicketsSold)
	fmt.Println(ticket.Quantity)
	ticketSeq = ticketSeq + 1
	ticket.TicketNo = ticketSeq
	ticket.Show = show
	ticket.Window = window
	ticket.DocType = TICKET

	showAsBytes, _ := json.Marshal(show)
	err = APIstub.PutState("SHOW"+strconv.Itoa(show.ShowID), showAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	windowAsBytes, _ := json.Marshal(window)
	err = APIstub.PutState("WINDOW"+strconv.Itoa(window.WindowNo), windowAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("saving current ticketSeq", ticketSeq)
	_, err = set(APIstub, NEXT_TICKET_ID, strconv.Itoa(ticketSeq))
	if err != nil {
		return shim.Error(err.Error())
	}

	ticketAsBytes, _ := json.Marshal(ticket)
	err = APIstub.PutState("TICKET"+strconv.Itoa(ticketSeq), ticketAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("API::purchaseTicket:End")
	return shim.Success([]byte(APIstub.GetTxID()))
}

// Issue coupon for the waterbottle and popcorn also for the soda exchange
func (s *SmartContract) issueCoupon(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::issueCoupon:Start")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var ticket Ticket

	if err := json.Unmarshal([]byte(args[0]), &ticket); err != nil {
		fmt.Println("Cannot unmarshal ticket Object", err)
		return shim.Error(err.Error())
	}
	ticketBytes, err := APIstub.GetState("TICKET" + strconv.Itoa(ticket.TicketNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(ticketBytes, &ticket)

	if ticket.CouponAvailed {
		fmt.Println("Coupon Availed Already")
		return shim.Error("Coupon Availed Already")
	}

	txnId := APIstub.GetTxID()
	var number int
	for _, c := range txnId {
		number = number + int(c)
	}
	ticket.CouponNumber = strconv.Itoa(number)
	ticket.CouponAvailed = true
	ticket.ExchangeAvailed = false
	ticketAsBytes, _ := json.Marshal(ticket)
	err = APIstub.PutState("TICKET"+strconv.Itoa(ticket.TicketNo), ticketAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("API::issueCoupon:End")
	return shim.Success([]byte("Coupon Number:" + ticket.CouponNumber))
}

func (s *SmartContract) availExchange(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::availExchange:Start")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var ticket Ticket

	if err := json.Unmarshal([]byte(args[0]), &ticket); err != nil {
		fmt.Println("Cannot unmarshal ticket Object", err)
		return shim.Error(err.Error())
	}
	ticketBytes, err := APIstub.GetState("TICKET" + strconv.Itoa(ticket.TicketNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(ticketBytes, &ticket)

	if ticket.ExchangeAvailed {
		fmt.Println("Exchange Availed Already")
		return shim.Error("Exchange Availed Already")
	}
	// check if even number for the eligible soda exchange
	couponNo, err := strconv.Atoi(ticket.CouponNumber)
	if err != nil {
		fmt.Println("Ticket Not eligible for exchange")
		return shim.Error("Ticket Not eligible for exchange")
	}
	if couponNo%2 != 0 {
		fmt.Println("Ticket Not eligible for exchange")
		return shim.Error("Ticket Not eligible for exchange")
	}
	ticket.ExchangeAvailed = true
	ticketAsBytes, _ := json.Marshal(ticket)
	err = APIstub.PutState("TICKET"+strconv.Itoa(ticket.TicketNo), ticketAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	var theatre Theatre
	theatreBytes, err := APIstub.GetState("THEATRE" + strconv.Itoa(ticket.Show.TheatreNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(theatreBytes, &theatre)

	theatre.SodaStock = theatre.SodaStock - ticket.Quantity

	theatreAsBytes, _ := json.Marshal(theatre)
	err = APIstub.PutState("THEATRE"+strconv.Itoa(theatre.TheatreNo), theatreAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("API::availExchange:End")
	return shim.Success([]byte(APIstub.GetTxID()))
}

func (s *SmartContract) queryByString(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	queryString := args[0]
	fmt.Println("queryString" + queryString)
	queryResults, err := getQueryResultForQueryString(APIstub, queryString)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func get(APIstub shim.ChaincodeStubInterface, key string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("Incorrect arguments. Expecting a key")
	}
	value, err := APIstub.GetState(key)
	if err != nil {
		return 0, fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
	}
	if value == nil {
		return 0, fmt.Errorf("Asset not found: %s", key)
	}
	var property Property
	json.Unmarshal(value, &property)
	fmt.Println(property)
	fmt.Println(property.Value)
	i, err := strconv.Atoi(property.Value)
	if err != nil {
		return 0, fmt.Errorf("Failed to get next sequence number", err)
	}
	fmt.Println("Got the the value for %s : value : %s", key, i)
	return i, nil
}

func set(APIstub shim.ChaincodeStubInterface, key string, value string) (string, error) {
	fmt.Println("setting value", key, value)

	var property Property
	property.Key = key
	property.Value = value

	propertyAsBytes, _ := json.Marshal(property)
	err := APIstub.PutState(key, propertyAsBytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return value, nil
}

// =========================================================================================
// The main function is only relevant in unit test mode. Only included here for completeness.
// =========================================================================================
func main() {
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
