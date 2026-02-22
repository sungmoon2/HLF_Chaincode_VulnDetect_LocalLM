package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
)

type User struct {
	ID        string `json:user_id`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	DocType   string `json:"doc_type"`
	Email     string `json:"email"`
	Status    uint8  `json:"status"` //1 = active, 0 = inactive
}

type LoginResponse struct {
	Status  int8        `json:"status"`
	Message string      `json:"message"`
	Id      string      `json:"id"`
	Data    interface{} `json:"data"`

}

type LoginQuery struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Status  int8        `json:"status"`
}

type Response struct {
	Status  int32       `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type UserResponse struct {
	ID        string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type EmailQuery struct {
	Email string `json:"email"`
}

type QueryStruct struct {
	Selector interface{} `json:"selector"`
}

//SmartContract ... The SmartContract
type SmartContract struct {
}

//Init Function
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("Chaincode Successfully initialized"))
}

//Invoke function
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	fun, args := stub.GetFunctionAndParameters()

	if fun == "registerUser" {
		return registerUser(stub,args)
	} else if fun == "login" {
		return login(stub,args)
	}

	return createErrorResponse("Invalid function name = "+fun, 0, nil)
}

//registerUser ... This function will add a new user to database
func registerUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 5 {
		return createErrorResponse("Invalid number of arguments. Expected 5", 422, nil)
	}

	var emailArr = []string{args[4]}

	var emailValidateResponse = CheckEmail(stub, emailArr)

	if emailValidateResponse.Status == 500 {
		return emailValidateResponse
	}


	user := User{
		ID:        args[0],
		FirstName: args[1],
		LastName:  args[2],
		Password:  args[3],
		DocType:   "user",
		Email:     args[4],
		Status:    1,}

	userBytes, err := json.Marshal(user)

	if err != nil {
		return shim.Error(err.Error())
	}

	putStateError := stub.PutState(args[0], userBytes)

	fmt.Println("User Record => " + string(userBytes))
	if putStateError != nil {
		return shim.Error(putStateError.Error())
	}

	userResponse := UserResponse{
		ID:        args[0],
		Email:     args[4],
		FirstName: args[1],
		LastName:  args[2]}

	response := Response{
		Status:  1,
		Message: "User created successfully",
		Data:    userResponse}

	responseBytes, respError := json.Marshal(response)

	if respError != nil {
		return shim.Error(respError.Error())
	}

	return shim.Success(responseBytes)
}

//login ... This function will login user
func login(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return createErrorResponse("Incorrect number of arguments. Expected 2 arguments", 422, nil)
	}

	email := args[0]

	loginQuery := LoginQuery{Email: email, Password: args[1], Status: 1}
	query := QueryStruct{Selector: loginQuery}

	queryByte, err := json.Marshal(query)

	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("Query ==> " + string(queryByte))

	queryIterator, queryError := stub.GetQueryResult(string(queryByte))

	if queryError != nil {
		return shim.Error(queryError.Error())
	}

	if queryIterator.HasNext() {
		queryResult, _ := queryIterator.Next()

		var user User

		err := json.Unmarshal(queryResult.Value, &user)

		if err != nil {
			return shim.Error(err.Error())
		}

		if user.Status == 0 {
			return createErrorResponse("You have been disabled by admin.", 401, nil)
		}

		//send user data with id on successfull login
		loginResponse := LoginResponse{
			Status:  1,
			Message: "User login successful",
			Id:      queryResult.Key,
			Data:    user}

		responseBytes, respError := json.Marshal(loginResponse)

		if respError != nil {
			return createErrorResponse(respError.Error(), 500, nil)
		}

		return shim.Success(responseBytes)
	}
	return createErrorResponse("The email or password you have entered is wrong.", 401, nil)
}

//CheckEmail ... This function will check whether user email exist in database or not
func CheckEmail(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Invalid number of arguments. Expected 1 argument.")
	}

	emailStruct := EmailQuery{Email: args[0]}
	selector := QueryStruct{Selector: emailStruct}

	selectorByte, err := json.Marshal(selector)

	if err != nil {
		return shim.Error(err.Error())
	}

	queryIterator, queryError := stub.GetQueryResult(string(selectorByte))

	if queryError != nil {
		return shim.Error(queryError.Error())
	}

	if queryIterator.HasNext() {
		return shim.Error("The email you have entered already exist.")
	}

	return shim.Success(createSuccessResponse("The email you have entered is a valid email.",
		1,
		nil))

}

// createErrorResponse ... This function will create error response
func createErrorResponse(message string, status int32, data interface{}) peer.Response {
	response := Response{Message: message,
		Status: status,
		Data:   data}

	responseByte, err := json.Marshal(response)

	var peerResponse = peer.Response{Payload: responseByte, Status: status, Message: message}

	if err != nil {
		peerResponse.Payload = []byte("{\"message\":\"Something went wrong while parsing json\"}")
		peerResponse.Status = 500
	}

	return peerResponse
}

//createSuccessResponse... This function will create success response
func createSuccessResponse(message string, status int32, data interface{}) []byte {
	response := Response{Message: message,
		Status: status,
		Data:   data}

	responseByte, err := json.Marshal(response)

	if err != nil {
		return []byte(err.Error())
	}

	return responseByte
}

func main() {
	err := shim.Start(new(SmartContract))

	if err != nil {
		fmt.Print(err)
	}
}
