package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	PREFIX_TOPIC_SENDER   = "TOPIC_SEND_"
	PREFIX_TOPIC_RECEIVER = "TOPIC_RECV_"
	COLLECTION_PRIV_KEYS  = "private_keys"
)

// RelayAdapter - relay adapter for communication cross blockchain network
type RelayAdapter struct {
	bccspInst bccsp.BCCSP
}

func main() {
	factory.InitFactories(nil)
	// bccspName := "PKCS11"
	// bccsp, err := factory.GetBCCSP(bccspName)
	// if err != nil {
	// 	fmt.Printf("Error getting BCCSP(%s): %s", bccspName, err)
	// 	fmt.Println()
	// }
	err := shim.Start(&RelayAdapter{factory.GetDefault()})
	if err != nil {
		fmt.Printf("Error starting RelayAdapter: %s", err)
		fmt.Println()
	}
}

// Init - Initializing smart contract
func (t *RelayAdapter) Init(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, _ := stub.GetFunctionAndParameters()
	if funcName == "init" {
		// init deployment of smart contract
		return t.init(stub)
	} else if funcName == "upgrade" {
		// upgrade deployment of smart contract
		return t.upgrade(stub)
	}

	return shim.Error(fmt.Sprintf(`Invalid function name for deployment. 
		(expecting 'init' or 'upgrade', actual: '%s')`, funcName))
}

func (t *RelayAdapter) init(stub shim.ChaincodeStubInterface) pb.Response {
	// TODO:
	return shim.Success(nil)
}

func (t *RelayAdapter) upgrade(stub shim.ChaincodeStubInterface) pb.Response {
	// TODO:
	return shim.Success(nil)
}

// Invoke - accessing smart contract interface(query or invoke)
func (t *RelayAdapter) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, params := stub.GetFunctionAndParameters()
	if funcName == "topic" {
		// query installed topics (IN or OUT)
		return t.findTopic(stub, params)
	} else if funcName == "topics" {
		// list installed topics (IN or OUT)
		return t.listTopic(stub, params)
	} else if funcName == "sessions" {
		// list message (IN or OUT)
		return t.listSession(stub, params)
	} else if funcName == "install" {
		// install relay topic (IN or OUT)
		return t.install(stub, params)
	} else if funcName == "uninstall" {
		// uninstall relay topic (IN or OUT)
		return t.uninstall(stub, params)
	} else if funcName == "register" {
		// register relay topic (IN or OUT)
		return t.register(stub, params)
	} else if funcName == "send" {
		// send message to topic (IN or OUT)
		return t.send(stub, params)
	} else if funcName == "read" {
		return t.readMessage(stub, params)
	}
	return shim.Success(nil)
}

func (t *RelayAdapter) findTopic(stub shim.ChaincodeStubInterface, params []string) pb.Response {
	if len(params) != 2 {
		return shim.Error(fmt.Sprintf(`Incorrect number of arguments. (expecting 2, actual: "%d")`, len(params)))
	}

	topicType := params[0]
	topicName := params[1]

	var topic *Topic
	var err error
	if topicType == string(OUT) {
		topic, err = findTopic(stub, DOC_TOPIC_OUT, topicName)
	} else if topicType == string(IN) {
		topic, err = findTopic(stub, DOC_TOPIC_IN, topicName)
	} else {
		return shim.Error("topic type is wrong")
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	bytes, err := json.Marshal(topic)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(bytes)
}

func (t *RelayAdapter) listTopic(stub shim.ChaincodeStubInterface, params []string) pb.Response {
	if len(params) != 2 {
		return shim.Error(fmt.Sprintf(`Incorrect number of arguments. (expecting 2, actual: "%d")`, len(params)))
	}

	topicType := params[0]
	orgID := params[1]

	var topics []*Topic
	var err error
	if topicType == string(OUT) {
		topics, err = queryTopicsByOrg(stub, DOC_TOPIC_OUT, orgID)
	} else if topicType == string(IN) {
		topics, err = queryTopicsByOrg(stub, DOC_TOPIC_IN, orgID)
	} else {
		return shim.Error("topic type is wrong")
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	bytes, err := json.Marshal(topics)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(bytes)
}

func (t *RelayAdapter) readMessage(stub shim.ChaincodeStubInterface, params []string) pb.Response {
	if len(params) != 3 {
		return shim.Error(fmt.Sprintf(`Incorrect number of arguments. (expecting 3, actual: "%d")`, len(params)))
	}

	sessionType := params[0]
	sessionID := params[1]
	orgID := params[2]

	var session *Session
	var err error
	var topicDocType, sessionDocType DocumentType
	if sessionType == string(OUT) {
		sessionDocType = DOC_SESSION_OUT
		topicDocType = DOC_TOPIC_OUT
	} else if sessionType == string(IN) {
		sessionDocType = DOC_SESSION_IN
		topicDocType = DOC_TOPIC_IN
	} else {
		return shim.Error("read message failed. cause: message type is wrong")
	}

	session, err = findSession(stub, sessionDocType, sessionID)

	if err != nil {
		return shim.Error(err.Error())
	}

	if session == nil {
		return shim.Error(fmt.Sprintf(`read message failed. cause: session not found.(id: %s)`, sessionID))
	}

	topic, err := findTopic(stub, topicDocType, session.TopicName)

	if err != nil {
		return shim.Error(err.Error())
	}

	reader, err := topic.GetReader(orgID)
	if err != nil {
		shim.Error(fmt.Sprintf(`read message failed. cause: reader not existing.(error: %s)`, err.Error()))
	}

	privateKey, err := getReaderPrivateKey(stub, reader.PrivateHash)
	if err != nil {
		shim.Error(fmt.Sprintf(`read message failed. cause: get private key failed.(error: %s)`, err.Error()))
	}

	decoded, err := reader.DecryptMessage(orgID, session.Message, privateKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	session.Message = decoded

	bytes, err := json.Marshal(session)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(bytes)
}

func (t *RelayAdapter) listSession(stub shim.ChaincodeStubInterface, params []string) pb.Response {
	if len(params) != 2 {
		return shim.Error(fmt.Sprintf(`Incorrect number of arguments. (expecting 2, actual: "%d")`, len(params)))
	}

	topicType := params[0]
	topicName := params[1]

	var sessions []*Session
	var err error
	if topicType == string(OUT) {
		sessions, err = querySessionByTopic(stub, DOC_SESSION_OUT, topicName)
	} else if topicType == string(IN) {
		sessions, err = querySessionByTopic(stub, DOC_SESSION_IN, topicName)
	} else {
		return shim.Error("session type is wrong")
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	bytes, err := json.Marshal(sessions)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(bytes)
}

// install topic for sender|receiver
func (t *RelayAdapter) install(stub shim.ChaincodeStubInterface, params []string) pb.Response {
	if len(params) != 2 {
		return shim.Error(fmt.Sprintf(`Incorrect number of arguments. (expecting 2, actual: "%d")`, len(params)))
	}

	topicType := params[0]
	topicName := params[1]

	var topic *Topic
	var err error
	if topicType == string(IN) {
		topic, err = findTopic(stub, DOC_TOPIC_IN, topicName)
	} else if topicType == string(OUT) {
		topic, err = findTopic(stub, DOC_TOPIC_OUT, topicName)
	} else {
		return shim.Error("topic type is wrong")
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	if topic == nil {
		if topicType == string(IN) {
			topic = NewTopic(DOC_TOPIC_IN, topicName)
		} else if topicType == string(OUT) {
			topic = NewTopic(DOC_TOPIC_OUT, topicName)
		} else {
			return shim.Error("topic type is wrong")
		}
		topic.Id = stub.GetTxID()
		data, err := ToJSON(topic)
		if err != nil {
			return shim.Error(err.Error())
		}
		fmt.Printf(`Putting state '%s'`, data)
		fmt.Println()
		err = stub.PutState(topic.Id, []byte(data))
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success(nil)
	}

	return shim.Error(fmt.Sprintf(`topic already existing. (topic: %s)`, topic.Name))
}

func (t *RelayAdapter) register(stub shim.ChaincodeStubInterface, params []string) pb.Response {
	if len(params) != 6 {
		return shim.Error(fmt.Sprintf(`Incorrect number of arguments. (expecting 6, actual: "%d")`, len(params)))
	}

	topicType := params[0]
	topicName := params[1]
	registerType := params[2]
	orgID := params[3]
	publicKey := params[4]
	privateKey := params[5]

	var topic *Topic
	var err error
	if topicType == string(IN) {
		topic, err = findTopic(stub, DOC_TOPIC_IN, topicName)
	} else if topicType == string(OUT) {
		topic, err = findTopic(stub, DOC_TOPIC_OUT, topicName)
	} else {
		return shim.Error("topic type is wrong")
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	if topic == nil {
		return shim.Error(fmt.Sprintf(`topic not existing. (topic: %s)`, topicName))
	}

	if topicType == string(IN) && topic.SenderExist(orgID) {
		return shim.Error(fmt.Sprintf(`topic already registered. (topic: %s ,org: %s)`, topic.Name, orgID))
	} else if topicType == string(OUT) && topic.SenderExist(orgID) {
		return shim.Error(fmt.Sprintf(`topic already registered. (topic: %s ,org: %s)`, topic.Name, orgID))
	}

	if registerType == string(REGISTER_SENDER) {
		if topicType == string(IN) && topic.SenderExist(orgID) {
			return shim.Error(fmt.Sprintf(`topic already registered. (topic: %s ,org: %s)`, topic.Name, orgID))
		} else if topicType == string(OUT) && topic.SenderExist(orgID) {
			return shim.Error(fmt.Sprintf(`topic already registered. (topic: %s ,org: %s)`, topic.Name, orgID))
		}
		topic.AddSender(orgID, publicKey)
	} else if registerType == string(REGISTER_READER) {
		if topicType == string(IN) && topic.ReaderExist(orgID) {
			return shim.Error(fmt.Sprintf(`topic already registered. (topic: %s ,org: %s)`, topic.Name, orgID))
		} else if topicType == string(OUT) && topic.ReaderExist(orgID) {
			return shim.Error(fmt.Sprintf(`topic already registered. (topic: %s ,org: %s)`, topic.Name, orgID))
		}
		privateKeyHash := stub.GetTxID()
		topic.AddReader(orgID, publicKey, privateKeyHash)

		err := saveReaderPrivateKey(stub, privateKeyHash, []byte(privateKey))
		if err != nil {
			return shim.Error(fmt.Sprintf(`private key save failed. cause: %s`, err.Error()))
		}
	} else {
		return shim.Error("register type is wrong")
	}

	data, err := ToJSON(topic)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf(`Putting state '%s'`, data)
	fmt.Println()
	err = stub.PutState(topic.Id, []byte(data))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *RelayAdapter) uninstall(stub shim.ChaincodeStubInterface, params []string) pb.Response {
	// TODO:
	return shim.Success(nil)
}

func (t *RelayAdapter) send(stub shim.ChaincodeStubInterface, params []string) pb.Response {
	if len(params) != 3 {
		return shim.Error(fmt.Sprintf(`Incorrect number of arguments. (expecting 3, actual: "%d")`, len(params)))
	}

	topicType := params[0]
	messageJSON := params[1]
	routeJSON := params[2]

	session := new(Session)
	err := json.Unmarshal([]byte(messageJSON), &session)
	if err != nil {
		return shim.Error(err.Error())
	}

	route := new(Route)
	err = json.Unmarshal([]byte(routeJSON), route)
	if err != nil {
		return shim.Error(err.Error())
	}

	var topic *Topic
	if topicType == string(OUT) {
		topic, err = findTopic(stub, DOC_TOPIC_OUT, session.TopicName)
		if err != nil {
			return shim.Error(err.Error())
		} else if topic == nil {
			return shim.Error("send message(OUT) failed, cause: topic not found")
		} else if topic.SenderExist(route.OrgID) == false {
			return shim.Error(fmt.Sprintf(`send message(OUT) failed, cause: sender not registered.(org:%s)`, route.OrgID))
		}
		session.Message, err = topic.EncryptMessage(route.OrgID, session.Message)
		if err != nil {
			return shim.Error(fmt.Sprintf(`send message failed, cause: %s`, err.Error()))
		}
		session.DocType = DOC_SESSION_OUT
	} else if topicType == string(IN) {
		topic, err = findTopic(stub, DOC_TOPIC_IN, session.TopicName)

		if err != nil {
			return shim.Error(err.Error())
		} else if topic == nil {
			return shim.Error("send message(IN) failed, cause: topic not found")
		} else if topic.SenderExist(route.OrgID) == false {
			return shim.Error(fmt.Sprintf(`send message(IN) failed, cause: sender not registered.(topic:%s, org:%s)`, topic.Name, route.OrgID))
		}
		session.DocType = DOC_SESSION_IN
	} else {
		return shim.Error("topic type is wrong")
	}
	session.Id = stub.GetTxID()
	route.TrxID = stub.GetTxID()
	session.Histories = append(session.Histories, route)

	//encryption failed.
	if err != nil {
		return shim.Error(err.Error())
	}

	data, err := ToJSON(session)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf(`Putting state '%s'`, data)
	fmt.Println()
	err = stub.PutState(session.Id, []byte(data))
	if err != nil {
		return shim.Error(err.Error())
	}

	if topicType == string(OUT) {
		//send message to event hub
		fmt.Printf(`send message: |%s|`, data)
		fmt.Println()
		err = stub.SetEvent(session.TopicName, []byte(data))
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	return shim.Success(nil)
}

func saveReaderPrivateKey(stub shim.ChaincodeStubInterface, privateHash string, privateKey []byte) error {
	err := stub.PutPrivateData(COLLECTION_PRIV_KEYS, privateHash, privateKey)
	return err
}

func getReaderPrivateKey(stub shim.ChaincodeStubInterface, privateHash string) ([]byte, error) {
	return stub.GetPrivateData(COLLECTION_PRIV_KEYS, privateHash)
}
