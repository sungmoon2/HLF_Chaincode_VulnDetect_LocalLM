package main

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/api/gmail/v1"
)

type OptimumChaincode struct {
}

func (t *OptimumChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("### Optimum Chaincode: Initialization")
	_, args := stub.GetFunctionAndParameters()
	var credKey string
	var credValue string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	credKey = "credentials"
	credValue = args[0]
	
	fmt.Printf("OAuth2 Credentials = %s", credValue)

	err = stub.PutState(credKey, []byte(credValue))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *OptimumChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("### Optimum Chaincode: Invocation")
	function, args := stub.GetFunctionAndParameters()
	if function == "getAuthUrl" {
		return t.getAuthUrl(stub, args)
	} else if function == "setVar" {
		return t.setVar(stub, args)
	} else if function == "getLabels" {
		return t.getLabels(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"getAuthUrl\" \"setVar\" \"getLabels\"")
}

func (t *OptimumChaincode) getAuthUrl(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments.")
	}

	credentials, err := stub.GetState("credentials")
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for credentials\"}"
		return shim.Error(jsonResp)
	}

	config, err := google.ConfigFromJSON(credentials, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret credentials to config: %v", err)
	}
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser to get auth token: \n%v\n", authURL)

	return shim.Success(nil)
}

func (t *OptimumChaincode) setVar(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var A string
	var Aval string
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	A = args[0]
	Aval = args[1]
	
	err = stub.PutState(A, []byte(Aval))
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Printf("authToken = %s", Aval)

	return shim.Success(nil)
}


func (t *OptimumChaincode) getLabels(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments.")
	}

	Avalbytes, err := stub.GetState("authToken")
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for authToken\"}"
		return shim.Error(jsonResp)
	}
	authCode := string(Avalbytes)

	credentials, err := stub.GetState("credentials")
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for credentials\"}"
		return shim.Error(jsonResp)
	}

	config, err := google.ConfigFromJSON(credentials, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret credentials to config: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	client := config.Client(context.Background(), tok)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}
	if len(r.Labels) == 0 {
		fmt.Println("No labels found.")
	}
	fmt.Print("Gmail Labels: ")
	for _, l := range r.Labels {
		fmt.Printf("[%s] ", l.Name)
	}
	fmt.Println()

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(OptimumChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
