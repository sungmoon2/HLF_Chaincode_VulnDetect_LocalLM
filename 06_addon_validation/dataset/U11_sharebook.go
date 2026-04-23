package main

import (
  "encoding/json"
  "fmt"
  "encoding/hex"
  "strconv"
  "strings"
  "log"
  "encoding/pem"
  "regexp"
  "crypto/x509"
  "github.com/hyperledger/fabric-contract-api-go/contractapi"
)



// SmartContract provides functions for managing an Book
type SmartContract struct {
  contractapi.Contract
}

type StudentConfig struct {
  Count          int `json:"Count"`
}
type BookConfig struct {
  Count          int `json:"Count"`
}
type Student struct {
  Org             string `json:"Org"`
  StudentID       string `json:"StudentID"`
  Name            string `json:"Name"`
  Phone           string `json:"Phone"`
  Email           string `json:"Email"`
}
type BookRequester struct {
  Org             string `json:"Org"`
  StudentID       string `json:"StudentID"`
}
// Book describes basic details of what makes up a simple book
type Book struct {
  ID             string `json:"ID"`
  Title          string `json:"Title"`
  Author         string `json:"Author"`
  ISBN           string `json:"Isbn"`
  Owner          string `json:"Owner"`
  Holder         BookRequester `json:"Holder"`
  IsBorrowed     bool   `json:"IsBorrowed"`
  RequestQueue   []BookRequester `json:"RequestQueue"`
  EntitleList    []string    `json:"EntitleList"`
  ReaderList    []string    `json:"ReaderList"`
}

func (s *SmartContract) GetClientName(ctx contractapi.TransactionContextInterface) (string,error) {
  caller,err := ctx.GetStub().GetCreator()
  if err != nil {
    return "",err
  }
  re := regexp.MustCompile("-----BEGIN CERTIFICATE-----[^ ]+-----END CERTIFICATE-----\n") 
  match := re.FindStringSubmatch(string(caller))
  pemBlock, _ := pem.Decode([]byte(match[0]))
  cert, err := x509.ParseCertificate(pemBlock.Bytes)
  if err != nil {
    return "",err
  }
  owner := cert.Subject.CommonName
  return owner,nil
}
// Register a new student to the world state with given details.
func (s *SmartContract) RegStudent(ctx contractapi.TransactionContextInterface) error {
  // Get new asset from transient map
  transientMap, err := ctx.GetStub().GetTransient()
  if err != nil {
        return fmt.Errorf("error getting transient: %v", err)
  }

  // Asset properties are private, therefore they get passed in transient field, instead of func args
  transientStudentJSON, ok := transientMap["student_properties"]
  if !ok {
	return fmt.Errorf("asset not found in the transient map input")
  }

  type StudentTransientInput struct {
		Name           string `json:"Name"`
		Phone          string `json:"Phone"`
		Email          string `json:"Email"`
  }

  var studentInput StudentTransientInput
  err = json.Unmarshal(transientStudentJSON, &studentInput)
  fmt.Println(studentInput)
  if err != nil {
    return err
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  fmt.Println("registering student by:")
  fmt.Println(caller)
  tokens := strings.Split(caller, "@")
  domain := ""
  if len(tokens) >= 2 {
      domain = tokens[1]
  }
  re := regexp.MustCompile("Admin@org([0-9]).example.com")
  match := re.FindStringSubmatch(string(caller))
  fmt.Println(match)
  PrivateCollection := ""
  if len(match) >= 2 {
     PrivateCollection = "_implicit_org_Org"+match[1]+"MSP"
  } else {
     return fmt.Errorf("failed to parse org string")
  }
  studentcfg := StudentConfig{ Count: 0 }
  configJSON, err := ctx.GetStub().GetPrivateData(PrivateCollection,"StudentConfig")
  if configJSON != nil {
     err = json.Unmarshal(configJSON, &studentcfg)
  }
  studentcfg.Count += 1
  configJSON2, err := json.Marshal(studentcfg)
  ctx.GetStub().PutPrivateData(PrivateCollection,"StudentConfig", configJSON2)
  if err != nil {
    return err
  }
  id := "student_" + strconv.Itoa(studentcfg.Count) + "@" + domain
  fmt.Println(id)
  student := Student{
    Org:            caller,
    StudentID:      id,
    Name:           studentInput.Name,
    Phone:          studentInput.Phone,
    Email:          studentInput.Email,
  }
  studentJSON, err := json.Marshal(student)
  if err != nil {
    return err
  }
  
  return ctx.GetStub().PutPrivateData(PrivateCollection,id, studentJSON)
}
  
// CreateBook issues a new book to the world state with given details.
func (s *SmartContract) CreateBook(ctx contractapi.TransactionContextInterface, title string, author string, isbn string ) error {
  owner,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  log.Println("creating book by:")
  log.Println(owner)

  bookcfg := BookConfig{ Count: 0 }
  configJSON, err := ctx.GetStub().GetState("BookConfig")
  if configJSON != nil {
     err = json.Unmarshal(configJSON, &bookcfg)
  }
  bookcfg.Count += 1
  configJSON2, err := json.Marshal(bookcfg)
  ctx.GetStub().PutState("BookConfig", configJSON2)
  id := "book_" + strconv.Itoa(bookcfg.Count)
  log.Println(id)
  book := Book{
    ID:             id,
    Title:          title,
    Author:         author,
    ISBN:           isbn,
    Owner:          owner,
    Holder:         BookRequester{"",""},
    IsBorrowed:     false,
    RequestQueue:   make([]BookRequester,0),
    EntitleList:    make([]string,0),
    ReaderList:     make([]string,0),
  }
  book.EntitleList = append(book.EntitleList,owner)
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, bookJSON)
}

// GetBook returns the book stored in the world state with given id.
func (s *SmartContract) GetBook(ctx contractapi.TransactionContextInterface, id string) (*Book, error) {
  bookJSON, err := ctx.GetStub().GetState(id)
  if err != nil {
    return nil, fmt.Errorf("failed to read from world state: %v", err)
  }
  if bookJSON == nil {
    return nil, fmt.Errorf("the book %s does not exist", id)
  }

  var book Book
  err = json.Unmarshal(bookJSON, &book)
  if err != nil {
    return nil, err
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return nil,err
  }
  for _, v := range book.EntitleList {
	if v == caller {
  		return &book, nil
	}
  }
  return nil,fmt.Errorf("the book %s not entitled", id)

}

func (s *SmartContract) GetStudentHash(ctx contractapi.TransactionContextInterface, client string , student string) (string,error) {
  PrivateCollection,err := s.GetPrivateCollection(ctx,client)
  if err != nil {
    return "",err
  }
  studentHash,err := ctx.GetStub().GetPrivateDataHash(PrivateCollection, student)
  if err != nil {
    return "",err
  }
  if studentHash == nil {
    return "",fmt.Errorf("cannot find the student")
  }
  return hex.EncodeToString(studentHash),nil
}
// AddRequest  let the client to add request an existing book in the world state with provided parameters.
func (s *SmartContract) AddRequest(ctx contractapi.TransactionContextInterface, id string , student string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }
  if !book.IsBorrowed {
    return fmt.Errorf("the book %s is not borrowed", id)
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  studentHash,err := s.GetStudentHash(ctx,caller,student)
  if err != nil {
    return err
  }
  fmt.Printf("student hash: %s",studentHash) 
  for _, v := range book.EntitleList {
	if v == caller {
           for _, s := range book.ReaderList {
	      if s == studentHash {
                requester := BookRequester{caller,student}
	        book.RequestQueue = append(book.RequestQueue,requester)
		bookJSON, err := json.Marshal(book)
		if err != nil {
		    return err
		}
		return ctx.GetStub().PutState(id, bookJSON)
             }
           }
	}
  }
  return fmt.Errorf("the book %s is not entitled by you %s", id,caller)
}
// ReturnBook  let the client to return an existing book in the world state with provided parameters.
func (s *SmartContract) ReturnBook(ctx contractapi.TransactionContextInterface, id string , student string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }
  if !book.IsBorrowed {
    return fmt.Errorf("the book %s is not borrowed", id)
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  if caller !=book.Holder.Org {
    return fmt.Errorf("the book %s is not borrowed by you %s", id,caller)
  }
  if student !=book.Holder.StudentID {
    return fmt.Errorf("the book %s is not borrowed by you %s", id,caller)
  }
  _,err = s.GetStudentHash(ctx,caller,student)
  if err != nil {
    return err
  }
  n := len(book.RequestQueue) 
  if n > 0 {
    new := make([]BookRequester,n-1)
    for i, v := range book.RequestQueue {
         if i == 0 {
            book.Holder = v
         } else {
	    new = append(new,v)
	 }
    }
    book.RequestQueue = new
  } else {
       book.IsBorrowed = false
  }
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }
  return ctx.GetStub().PutState(id, bookJSON)
}
// BorrowBook  let the client to borrow an existing book in the world state with provided parameters.
func (s *SmartContract) BorrowBook(ctx contractapi.TransactionContextInterface, id string, student string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }
  if book.IsBorrowed {
    return fmt.Errorf("the book %s has been borrowed", id)
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  studentHash,err := s.GetStudentHash(ctx,caller,student)
  if err != nil {
    return err
  }
  fmt.Printf("student hash: %s",studentHash) 
  for _, v := range book.EntitleList {
	if v == caller {
           for _, s := range book.ReaderList {
	      if s == studentHash {
		book.Holder = BookRequester{caller,student}
                book.IsBorrowed = true
                bookJSON, err := json.Marshal(book)
		if err != nil {
		    return err
		}
                return ctx.GetStub().PutState(id, bookJSON)
              }
           }
	}
  }
  return fmt.Errorf("the book %s not entitled", id)
}
// GrantBook grant client to read an existing book in the world state with provided parameters.
func (s *SmartContract) GrantBook(ctx contractapi.TransactionContextInterface, id string, client string, student string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  if book.Owner != caller {
    return fmt.Errorf("the book %s is not owned by you", id)
  }
  book.EntitleList = append(book.EntitleList,client)

  studentHash,err := s.GetStudentHash(ctx,client,student)
  if err != nil {
    return err
  }
  fmt.Printf("student hash: %s",studentHash) 
  book.ReaderList = append(book.ReaderList,studentHash)
  
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, bookJSON)
}
// UpdateBook updates an existing book in the world state with provided parameters.
func (s *SmartContract) UpdateBook(ctx contractapi.TransactionContextInterface, id string, title string, author string, isbn string, owner string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  // overwriting original book with new book
  book := Book{
    ID:             id,
    Title:          title,
    Author:         author,
    ISBN:           isbn,
    Owner:          owner,
  }
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, bookJSON)
}

// DeleteBook deletes an given book from the world state.
func (s *SmartContract) DeleteBook(ctx contractapi.TransactionContextInterface, id string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  return ctx.GetStub().DelState(id)
}

// BookExists returns true when book with given ID exists in world state
func (s *SmartContract) BookExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
  bookJSON, err := ctx.GetStub().GetState(id)
  if err != nil {
    return false, fmt.Errorf("failed to read from world state: %v", err)
  }

  return bookJSON != nil, nil
}

// TransferBook updates the owner field of book with given id in world state.
func (s *SmartContract) TransferBook(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }

  book.Owner = newOwner
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, bookJSON)
}

// GetAllBooks returns all book found in world state
func (s *SmartContract) GetAllBooks(ctx contractapi.TransactionContextInterface) ([]*Book, error) {
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return nil,err
  }
  // range query with empty string for startKey and endKey does an
  // open-ended query of all book in the chaincode namespace.
  resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  var books []*Book
  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }
    if strings.Contains(queryResponse.Key, "book_") {
       var book Book
       err = json.Unmarshal(queryResponse.Value, &book)
       if err != nil {
         return nil, err
       }
       for _, v := range book.EntitleList {
	   if v == caller {
 	   	books = append(books, &book)
                break;
           }

        }
      }
   }

  return books, nil
}

func (s *SmartContract) GetPrivateCollection(ctx contractapi.TransactionContextInterface, caller string) (string, error) {
  re := regexp.MustCompile("Admin@org([0-9]).example.com")
  match := re.FindStringSubmatch(string(caller))
  fmt.Println(match)
  PrivateCollection := ""
  if len(match) >= 2 {
     PrivateCollection = "_implicit_org_Org"+match[1]+"MSP"
  } else {
     return "",fmt.Errorf("failed to parse org string")
  }
  return PrivateCollection,nil
}
// GetAllStudents returns all book found in world state
func (s *SmartContract) GetAllStudents(ctx contractapi.TransactionContextInterface) ([]*Student, error) {
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return nil,err
  }
  PrivateCollection,err := s.GetPrivateCollection(ctx,caller)
  if err != nil {
    return nil,err
  }

  // range query with empty string for startKey and endKey does an
  // open-ended query of all book in the chaincode namespace.
  resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(PrivateCollection,"", "")
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  var students []*Student
  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }
    if strings.Contains(queryResponse.Key, "student_") {
       var student Student
       err = json.Unmarshal(queryResponse.Value, &student)
       if err != nil {
         return nil, err
       }
       if student.Org == caller {
 	   	students = append(students, &student)
       }
    }
  }

  return students, nil
}
func main() {
  bookChaincode, err := contractapi.NewChaincode(&SmartContract{})
  if err != nil {
    log.Panicf("Error creating book-transfer-basic chaincode: %v", err)
  }

  if err := bookChaincode.Start(); err != nil {
    log.Panicf("Error starting book-transfer-basic chaincode: %v", err)
  }
}
