
package main


import (
	"fmt"
	"strconv"
    "encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("moneyToticket")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//=====================================================================================
//account asset struct
//=====================================================================================
type Account struct {
    ID                 string `json:"ID"`               // the owner of assets
    Role               string `json:"role"`                  
    RMB                int    `json:"rmb"`                //捆包
    Coin               int    `json:"ouyeelcoin"`         

}

//=====================================================================================
//PayMoney is for customet paying money to ouyeel
// A:Ouyell   B:Customer
// R:RMB      T:Ticket  
// RA:Ouyell money     
// RB: Customer money
// TA:Ouyell ticket
// TB:Customer ticket 
//=====================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### moneyToticket Init ###########")
        _, args := stub.GetFunctionAndParameters()
	
	var err error
	var account Account
	
	// Initialize the asset of account
	account.ID = args[0]
	account.Role = args[1]
	account.RMB, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	account.Coin, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	logger.Info("account.RMB = %d, account.Coin = %d\n", account.RMB, account.Coin)

	/* Initialize the asset of customer
	B = args[3]
	RB, err = strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	TB, err = strconv.Atoi(args[5])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	logger.Info("RB = %d, TB = %d\n", RB, TB)
	*/
	
	Bytes, _ := json.Marshal(account) 
	// Write the state to the ledger
	err = stub.PutState(account.ID, Bytes)
	if err != nil {
		return shim.Error(err.Error())
	}
    
	return shim.Success(nil)


}

//===========================================================================
//customer account
//===========================================================================
func (t *SimpleChaincode) CreateCustomerAccount (stub shim.ChaincodeStubInterface, args []string) pb.Response {
        var err error
        var account Account   

  // Initialize the asset of account
	account.ID = args[0]
	account.Role = args[1]
	account.RMB, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	account.Coin, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	logger.Info("account.RMB = %d, account.OuyeelCoin = %d\n", account.RMB, account.Coin)
	
	Bytes, _ := json.Marshal(account) 
	// Write the state to the ledger
	err = stub.PutState(account.ID, Bytes)
	if err != nil {
		return shim.Error(err.Error())
	}
    
	return shim.Success(nil)
	
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### example_cc0 Invoke ###########")

	function, args := stub.GetFunctionAndParameters()
	
	if function == "PayMoney" {
		return t.PayMoney(stub, args)
	}
        if function == "Exchange" {
		return t.Exchange(stub, args)
	}
	if function == "QueryTicket" {
		return t.QueryTicket(stub, args)
	}
	if function == "CreateCustomerAccount" {
		return t.CreateCustomerAccount(stub, args)
	}
	
	logger.Errorf("Unknown action, check the first argument, must be one of 'CreateCustomerAccount', 'PayMoney', 'Exchange', or 'QueryTicket'. But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'CreateCustomerAccount', 'PayMoney', 'Exchange', or 'QueryTicket'. But got: %v", args[0]))
}



//===========================================================================
// PayMoney is for customet paying money to ouyeel
// A:Ouyell   B:Customer
// R:RMB      T:Ticket  
// RA:Ouyell money     
// RB: Customer money
// TA:Ouyell ticket
// TB:Customer ticket    
//===========================================================================
func (t *SimpleChaincode) PayMoney (stub shim.ChaincodeStubInterface, args []string) pb.Response {
    fmt.Println("PayMoney")
	
	var err error
    var customer Account
	var ouyeel Account
    var RMB int  
	
	if len(args) != 3 {
	    return shim.Error("Incorrect number of arguments. Expecting 3")
	}	
    
	customer.ID = args[0]
	ouyeel.ID = args[1]
	ouyeel
	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	customerbytes, err := stub.GetState(customer.ID)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if customerbytes == nil {
		return shim.Error("Entity not found")
	}
	customer.RMB, _ = strconv.Atoi(string(customerbytes))

	ouyeelbytes, err := stub.GetState(ouyeel.ID)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if ouyeelbytes == nil {
		return shim.Error("Entity not found")
	}
        ouyeel.RMB, _ = strconv.Atoi(string(ouyeelbytes))

    // Perform the money_pay execution
	RMB, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	ouyeel.RMB = ouyeel.RMB + RMB
	customer.RMB = customer.RMB - RMB
	fmt.Printf(" ouyeel.RMB= %d, customer.RMB = %d\n", ouyeel.RMB, customer.RMB)

	ouyeelBytes, _ := json.Marshal(ouyeel)               
	// Write the state back to the ledger
	err = stub.PutState(ouyeel.ID, ouyeelBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	customerBytes, _ := json.Marshal(customer)               
	err = stub.PutState(customer.ID, customerBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

    return shim.Success(nil)

}

//==========================================================================
// Exchange is for ouyeel exchanging ticket to customer
// A:Ouyell   B:Customer
// R:RMB      T:Ticket  
// RA:Ouyell money     
// RB: Customer money
// TA:Ouyell ticket
// TB:Customer ticket  
//==========================================================================
func (t *SimpleChaincode) Exchange (stub shim.ChaincodeStubInterface, args []string) pb.Response {
    fmt.Println("Exchange")
	
	var err error
        var customer Account
	var ouyeel Account
        var coin int  
	
	if len(args) != 3 {
	    return shim.Error("Incorrect number of arguments. Expecting 3")
	}	
    
	customer.ID = args[0]
	ouyeel.ID = args[1]
	
	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	customerbytes, err := stub.GetState(customer.ID)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if customerbytes == nil {
		return shim.Error("Entity not found")
	}
	customer.Coin, _ = strconv.Atoi(string(customerbytes))

	ouyeelbytes, err := stub.GetState(ouyeel.ID)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if ouyeelbytes == nil {
		return shim.Error("Entity not found")
	}
        ouyeel.Coin, _ = strconv.Atoi(string(ouyeelbytes))

        // Perform the money_pay execution
	coin, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	ouyeel.Coin = ouyeel.Coin - coin
	customer.Coin = customer.Coin - coin
	fmt.Printf(" ouyeel.Coin= %d, customer.Coin = %d\n", ouyeel.Coin, customer.Coin)

	ouyeelBytes, _ := json.Marshal(ouyeel)               
	// Write the state back to the ledger
	err = stub.PutState(ouyeel.ID, ouyeelBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	customerBytes, _ := json.Marshal(customer)               
	err = stub.PutState(customer.ID, customerBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

    return shim.Success(nil)

}
//==========================================================================
// QueryTicket is for customer quering ticket(是否需要登录账户)
//==========================================================================
func (t *SimpleChaincode) QueryTicket (stub shim.ChaincodeStubInterface, args []string) pb.Response {
        fmt.Println("QueryTicket")
	
	if len(args) != 1 {
	    return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	
	var account Account
	account.ID = args[0]
        Bytes, err := stub.GetState(account.ID)
	if err != nil {
		return shim.Error("Failed to get digital asset: " + err.Error())
	} 
	if Bytes == nil {
		return shim.Error("This ticket detail does not exists: " + account.ID)
        }
	

	var accountTicket []string
	json.Unmarshal(Bytes, &accountTicket)
	
	return shim.Success(Bytes)
}



//=======================================================================================
//function main
//=======================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
