package main

import (

  "encoding/json"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  sc "github.com/hyperledger/fabric/protos/peer"
)

type SimpleAssetChaincode struct{

}

type car struct{
Name string     `json:"name"`
Owner string    `json:"owner"`
Make string     `json:"make"`
Color string    `json:"color"`

}

var logger = shim.NewLogger("simple_asset_chaincode")

func (s *SimpleAssetChaincode) Init(APIstub shim.ChaincodeStubInterface)sc.Response{

logger.Info("**Init***")
return shim.Success(nil)

}

//Invoke
func (s *SimpleAssetChaincode) Invoke(APIstub shim.ChaincodeStubInterface)sc.Response{
  function, args := APIstub.GetFunctionAndParameters()

   if function == "initledger"{
      return s.initledger(APIstub,args)
   }else if function == "queryCar"{
     return s.queryCar(APIstub,args)
   }

   return shim.Error("Invalid fucntion name passed")
}


//function to initledger
func (s *SimpleAssetChaincode)initledger(APIstub shim.ChaincodeStubInterface, args[]string) sc.Response{

lname := args[0]
lowner := args[1]
lmake := args[2]
lcolor := args[3]

car := &car{lname, lowner, lmake, lcolor}

carInBytes, err:= json.Marshal(car)
if err!=nil{

  logger.Error("Error marshalling")
  return shim.Error(err.Error())
}

err = APIstub.PutState(lname,carInBytes)
if err!=nil{

  logger.Error("Error inerting into ledger")
  return shim.Error(err.Error())
}

return shim.Success(nil)

}

//Get Car asset from leger
func (s *SimpleAssetChaincode) queryCar(APIstub shim.ChaincodeStubInterface, args[]string)sc.Response{

  assetInBytes, err:= APIstub.GetState(args[0])

  if err!=nil{

    logger.Error("asset car %d is not present in the KVS",args[0])
    return shim.Error(err.Error())
  }else if assetInBytes==nil {
    logger.Error("asset car %d has data nil",args[0])
    return shim.Error(assetInBytes)
  }

logger.Info("asset car %d has data",string(assetInBytes))
  return shim.Success(assetInBytes)
}

func main(){

  err := shim.Start(new(SimpleAssetChaincode))
  if err != nil {
		logger.Error("Error creating new Smart Contract: %s", err)
	}
}
