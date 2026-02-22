package chaincode

import (
        "errors"
	"encoding/json"
	"fmt"
	"strconv"
	"tokenmarket"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type NekoTokenCC struct{
}

func checkLen(logger *shim.ChaincodeLogger, expected int, args []string) error {

        if len(args) < expected {
	         mes := fmt.Sprintf("not enough number of arguments: %d given, %d expected", len(args), expected,)
		 logger.Warning(mes)
		 return errors.New(mes)
	}
	return nil
}

func (cc *NekoTokenCC) Init(stub shim.ChaincodeStubInterface) pb.Response{

        fmt.Println("Init is running")

        var humans [2]tokenmarket.Status
	var humanbytes [2][]byte

        humans[0] = tokenmarket.Status{Id: "abc", Name: "Nekomata", TokenQuan: 100, MoneyQuan: 3000}
	humans[1] = tokenmarket.Status{Id: "xyz", Name: "Inumata", TokenQuan: 20, MoneyQuan: 20000}

        humanbytes[0], _ = json.Marshal(humans[0])
	humanbytes[1], _ = json.Marshal(humans[1])
	stub.PutState("abc", humanbytes[0])
	stub.PutState("xyz", humanbytes[1])

        return shim.Success(nil)
}

func (cc *NekoTokenCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
        logger := shim.NewLogger("tokenmarket")

        fmt.Println("Invoke is running")

        var (
	        function string
		args     []string
        )
        function, args = stub.GetFunctionAndParameters()
        logger.Infof("function name = %s", function)

        switch function {

	       case "buy":
       	       	    if err := checkLen(logger, 3, args); err != nil {
              	            return shim.Error(err.Error())
       		    }

       		    var id1, id2 string
       		    var token string

       		    err := json.Unmarshal([]byte(args[0]), &id1)
       		    if err != nil {
              	       	   mes := fmt.Sprintf("failed to unmarshal the 1st argument: %s", err.Error(),)
	      		   logger.Warning(mes)
	      		   return shim.Error(mes)
       	            }

       		    err = json.Unmarshal([]byte(args[1]), &id2)
       		    if err != nil {
              	       	      mes := fmt.Sprintf("failed to unmarshal the 2nd argument: %s", err.Error(),)
	      		      logger.Warning(mes)
	      		      return shim.Error(mes)
       		    }

       		    err = json.Unmarshal([]byte(args[2]), &token)
       		    if err != nil{
              	       	      mes := fmt.Sprintf("failed to unmarshal the 3rd arument: %s", err.Error(),)
	      		      logger.Warning(mes)
	      		      return shim.Error(mes)
       		    }

       		    err = cc.buy(stub,id1,id2,token)
       		    if err != nil{
              	       	      return shim.Error(err.Error())
       		    }

       		    return shim.Success([]byte{})

	       case "sell":
			if err := checkLen(logger, 3, args); err != nil {
              		       	  return shim.Error(err.Error())
       	        	}

       	       		var id1, id2 string
       			var token string

       			err := json.Unmarshal([]byte(args[0]), &id1)
       			if err != nil {
              		       	  mes := fmt.Sprintf("failed to unmarshal the 1st argument: %s", err.Error(),)
	     			  logger.Warning(mes)
	      			  return shim.Error(mes)
       		        }

       			err = json.Unmarshal([]byte(args[1]), &id2)
       			if err != nil {
              		       	  mes := fmt.Sprintf("failed to unmarshal the 2nd argument: %s", err.Error(),)
	      			  logger.Warning(mes)
	      			  return shim.Error(mes)
       		        }

       			err = json.Unmarshal([]byte(args[2]), &token)
       			if err != nil {
               		       	  mes := fmt.Sprintf("failed to unmarshal the 3rd argument: %s", err.Error(),)
	       			  logger.Warning(mes)
	       			  return shim.Error(mes)
       			}

       			err = cc.buy(stub,id1,id2,token)
       			if err != nil {
             		       	  return shim.Error(err.Error())
       		        }
       
			return shim.Success([]byte{})

	       case "check":

			if err := checkLen(logger, 1, args); err != nil{
              		       	  return shim.Error(err.Error())
        		}

        		var id string

        		err := json.Unmarshal([]byte(args[0]), &id)
			if err != nil {
	        	       	  mes := fmt.Sprintf("failed to unmarshal the 1st argument: %s", err.Error(),)
				  logger.Warning(mes)
				  return shim.Error(mes)
			}

        		err = cc.check(stub, id)
			if err != nil {
	       		       	  return shim.Error(err.Error())
			}

	       case "charge":
			if err := checkLen(logger, 2, args); err != nil {
	        	       	  return shim.Error(err.Error())
			}

        		var id string
			var money string

        		err := json.Unmarshal([]byte(args[0]), &id)
			if err != nil{
	        	       	  mes := fmt.Sprintf("failed to unmarshal the 1st argument: %s", err.Error(),)
				  logger.Warning(mes)
				  return shim.Error(mes)
			}

        		err = json.Unmarshal([]byte(args[1]), &money)
			if err != nil{
	        	       	  mes := fmt.Sprintf("failed to unmarshal the 2nd argument: %s", err.Error(),)
				  logger.Warning(mes)
				  return shim.Error(mes)
			}

        		err = cc.charge(stub, id, money)
			if err != nil {
	        	       	  return shim.Error(err.Error())
			}

		return shim.Success([]byte{})
		}
		
	return shim.Success(nil)
}

func (cc *NekoTokenCC) check(stub shim.ChaincodeStubInterface, id string) error {

        logger := shim.NewLogger("tokenmarket")
	logger.Infof("check %s's status", id)

        status1 := tokenmarket.Status{}
	jsonBytes, _ := stub.GetState(id)
	json.Unmarshal(jsonBytes, &status1)
	fmt.Println("Status:")
	fmt.Print("----  ")
	fmt.Printf("Id: %s, Name: %s, Token: %d, Money: %d", status1.Id, status1.Name, status1.TokenQuan, status1.MoneyQuan)
	fmt.Println("  ----\n")

        jsonBytes, _ = json.Marshal(status1)
	stub.PutState(id, jsonBytes)


        return nil
}

func (cc *NekoTokenCC) buy(stub shim.ChaincodeStubInterface, id1 string, id2 string, token string) error {

        logger := shim.NewLogger("tokenmarket")

        status1 := tokenmarket.Status{}
	status2 := tokenmarket.Status{}
	statusjson1, _ := stub.GetState(id1)
	statusjson2, _ := stub.GetState(id2)
	json.Unmarshal(statusjson1, &status1)
	json.Unmarshal(statusjson2, &status2)

        logger.Infof("%s buys the %s token from %s", status1.Name, token, status2.Name)

        tokenint, _ := strconv.Atoi(token)
	moneymove := tokenint * 100

        if status1.MoneyQuan - moneymove >= 0 && status2.TokenQuan - tokenint >= 0 {
	         status1.TokenQuan += tokenint
		 status1.MoneyQuan -= moneymove
		 status2.TokenQuan -= tokenint
		 status2.MoneyQuan += moneymove
        } else if status1.MoneyQuan - moneymove < 0 {
	         fmt.Println(status1.Name + "の残金が不足しています")
		 fmt.Println("取引が成立しませんでした")
	} else if status2.TokenQuan - tokenint < 0 {
	         fmt.Println(status2.Name + "のトークンが不足しています")
		 fmt.Println("取引が成立しませんでした")
	}

        statusjson1, _ = json.Marshal(status1)
	statusjson2, _ = json.Marshal(status2)
	stub.PutState(id1, statusjson1)
	stub.PutState(id2, statusjson2)

        return nil

}

func (cc *NekoTokenCC) sell(stub shim.ChaincodeStubInterface, id1 string, id2 string, token string) error {

        logger := shim.NewLogger("tokenmarket")

        status1 := tokenmarket.Status{}
	status2 := tokenmarket.Status{}
	statusjson1, _ := stub.GetState(id1)
	statusjson2, _ := stub.GetState(id2)
	json.Unmarshal(statusjson1, &status1)
	json.Unmarshal(statusjson2, &status2)

        logger.Infof("%s sells the %s token to %s", status1.Name, token, status2.Name)

        tokenint, _ := strconv.Atoi(token)
	moneymove := tokenint * 100
	if status2.MoneyQuan - moneymove >= 0 && status1.TokenQuan - tokenint >= 0 {
	         status1.TokenQuan -= tokenint
		 status1.MoneyQuan += moneymove
		 status2.TokenQuan += tokenint
		 status2.MoneyQuan -= moneymove
	} else if status2.MoneyQuan - moneymove < 0 {
	         fmt.Println(status2.Name + "の残金が不足しています")
		 fmt.Println("取引が成立しませんでした")
	} else if status1.TokenQuan - tokenint < 0 {
	         fmt.Println(status1.Name + "のトークンが不足しています")
		 fmt.Println("取引が成立しませんでした")
	}

        statusjson1, _ = json.Marshal(status1)
	statusjson2, _ = json.Marshal(status2)
	stub.PutState(id1, statusjson1)
	stub.PutState(id2, statusjson2)

        return nil

}

func (cc *NekoTokenCC) charge(stub shim.ChaincodeStubInterface, id string, money string) error{

         logger := shim.NewLogger("tokenmarket")

         status := tokenmarket.Status{}
	 statusjson, _ := stub.GetState(id)
	 json.Unmarshal(statusjson, &status)

         logger.Infof("%s円を%sへ入金します", money, status.Name)

         moneyint, _ := strconv.Atoi(money)

         status.MoneyQuan += moneyint

         fmt.Println("入金が完了しました")
	 statusjson, _ = json.Marshal(status)
	 stub.PutState(id, statusjson)

         return nil

}

