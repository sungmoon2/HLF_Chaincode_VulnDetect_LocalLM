package main
import(
	"fmt"
	/*"strconv"*/
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type simplechaincode struct {

}

func (t *simplechaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	fmt.Println("<< ===== success init it is view in docker ==========>>")
	_, args := stub.GetFunctionAndParameters()
	var logger = shim.NewLogger("simplechaincode")
	/*
	var A, B string
	var Aval, Bval int

	var err error*/
	if len(args)!=4{
		return shim.Error(args[2])
	}else{
		logger.Info("Number of arguments is 4")
		logger.Info(args[0])
		logger.Info(args[1])
		logger.Info(args[2])
		logger.Info(args[3])
	}
	/*
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil{
		return shim.Error("Expecting integer value for asset holding")
	}
*/


	return shim.Success([] byte("success init"))
}

func (t *simplechaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response{
	fmt.Println("======= Invoke function is call successfully ========")

	/*_,args := stub.GetFunctionAndParameters();
	var a_parm = args[0]
	var b_parm = args[1]
	var c_parm = args[2]*/
	var logger = shim.NewLogger("simplechaincode")
	logger.Warning("----------------------------------test Warning--------------------------------------------")
	/*logger.info("args0:%s",a_parm)
	logger.info("args1:%s",b_parm)
	logger.info("args2:%s",c_parm)*/

	return shim.Success([] byte("Success invoke"))
}

func main(){
	err := shim.Start(new (simplechaincode))
	if err != nil{
		fmt.Println("Error starting Simple chaincode:%s", err)
	}
}
