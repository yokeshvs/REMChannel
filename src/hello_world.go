// DISCLAIMER:
// THIS SAMPLE CODE MAY BE USED SOLELY AS PART OF THE TEST AND EVALUATION OF THE SAP CLOUD PLATFORM
// BLOCKCHAIN SERVICE (THE "SERVICE") AND IN ACCORDANCE WITH THE TERMS OF THE AGREEMENT FOR THE SERVICE.
// THIS SAMPLE CODE PROVIDED "AS IS", WITHOUT ANY WARRANTY, ESCROW, TRAINING, MAINTENANCE, OR SERVICE
// OBLIGATIONS WHATSOEVER ON THE PART OF SAP.

package main

//=================================================================================================
//========================================================================================== IMPORT
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	//"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

//=================================================================================================
//============================================================================= BLOCKCHAIN DOCUMENT
// Doc writes string to the blockchain (as JSON object) for a specific key
type Doc struct {
	DStatus string `json:DStatus`
	Timestamp string `json:Timestamp`
	EdgeID string `json:EdgeID`
}

func (doc *Doc) FromJson(input []byte) *Doc {
	json.Unmarshal(input, doc)
	return doc
}

func (doc *Doc) ToJson() []byte {
	jsonDoc, _ := json.Marshal(doc)
	return jsonDoc
}

//=================================================================================================
//================================================================================= RETURN HANDLING
// Return handling: for return, we either return "shim.Success (payload []byte) with HttpRetCode=200"
// or "shim.Error(doc string) with HttpRetCode=500". However, we want to set our own status codes to
// map into HTTP return codes. A few utility functions:

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

//=================================================================================================
//====================================================================================== VALIDATION
// Validation: all arguments for a function call is passed as a string array args[]. Validate that
// the number, type and length of the arguments are correct. Only string validations are supported
// here, as all parameters are strings of different lengths.

func Validate(funcName string, args []string, desc ...interface{}) peer.Response {

	logger.Debugf("Function: %s(%s)", funcName, strings.TrimSpace(strings.Join(args, ",")))

	nrArgs := len(desc) / 3

	if len(args) != nrArgs {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	for i := 0; i < nrArgs; i++ {
		switch desc[i*3] {
		case "%s":
			var minLength int = desc[i*3+1].(int)
			var maxLength int = desc[i*3+2].(int)
			if len(args[i]) < minLength || len(args[i]) > maxLength {
				return Error(http.StatusBadRequest, "Parameter Length Error")
			}
		}
	}

	return Success(0, "OK", nil)
}

//=================================================================================================
//============================================================================================ MAIN
// Main function starts up the chaincode in the container during instantiate
//
var logger = shim.NewLogger("chaincode")

type HelloWorld struct {
	// use this structure for information that is held (in-memory) within chaincode
	// instance and available over all chaincode calls
}

func main() {
	if err := shim.Start(new(HelloWorld)); err != nil {
		fmt.Printf("Main: Error starting chaincode: %s", err)
	}
}

//=================================================================================================
//============================================================================================ INIT
// Init is called during Instantiate transaction after the chaincode container
// has been established for the first time, allowing the chaincode to
// initialize its internal data. Note that chaincode upgrade also calls this
// function to reset or to migrate data, so be careful to avoid a scenario
// where you inadvertently clobber your ledger's data!
//
func (cc *HelloWorld) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Validate supplied init parameters, in this case zero arguments!
	if _, args := stub.GetFunctionAndParameters(); len(args) > 0 {
		return Error(http.StatusBadRequest, "Init: Incorrect number of arguments; no arguments were expected.")
	}
	return Success(http.StatusOK, "OK", nil)
}

//=================================================================================================
//========================================================================================== INVOKE
// Invoke is called to update or query the ledger in a proposal transaction.
// Updated state variables are not committed to the ledger until the
// transaction is committed.
//
func (cc *HelloWorld) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	// Which function is been called?
	function, args := stub.GetFunctionAndParameters()

	// Route call to the correct function
	switch function {
	case "exist":
		return cc.exist(stub, args)
	case "read":
		return cc.read(stub, args)
	case "list":
		return cc.list(stub)	
	case "create":
		return cc.create(stub, args)
	case "update":
		return cc.update(stub, args)
	case "delete":
		return cc.delete(stub, args)
	case "history":
		return cc.history(stub, args)
	case "search":
		return cc.search(stub, args)
	default:
		logger.Warningf("Invoke('%s') invalid!", function)
		return Error(http.StatusNotImplemented, "Invalid method! Valid methods are 'create|update|delete|exist|read|list|history|search'!")
	}
}

//=================================================================================================
//=========================================================================================== EXIST
// Validate message's existance by ID
//
func (cc *HelloWorld) exist(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	if rc := Validate("exist", args /*args[0]=id*/, "%s", 1, 64); rc.Status > 0 {
		return rc
	}
	id := strings.ToLower(args[0])

	// If we can read the ID, then it exists
	if value, err := stub.GetState(id); err != nil || value == nil {
		return Error(http.StatusNotFound, "Not Found")
	}

	return Success(http.StatusNoContent, " Exists", nil)
}

//=================================================================================================
//============================================================================================ READ
// Read Message by DeviceID
//
func (cc *HelloWorld) read(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	if rc := Validate("read", args /*args[0]=id*/, "%s", 1, 255); rc.Status > 0 {
		return rc
	}
	id := strings.ToLower(args[0])

	// Read the value for the ID
	if value, err := stub.GetState(id); err != nil || value == nil {
		return Error(http.StatusNotFound, "Not Found")
	} else {
		return Success(http.StatusOK, "OK", value)
	}
}



//=================================================================================================
//============================================================================================ LIST
// List of DeviceID
//
func (cc *HelloWorld) list(stub shim.ChaincodeStubInterface) peer.Response {

	startKey := ""
	endKey := ""

	// Get list of ID
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return Error(http.StatusNotFound, "Not Found")
	}
	defer resultsIterator.Close()

	// Write return buffer
	var buffer bytes.Buffer
	buffer.WriteString("{ \"values\": [")
	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()
		if buffer.Len() > 15 {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"id\":\"")
		buffer.WriteString(it.Key)
		buffer.WriteString("\"}")


	}
	buffer.WriteString("]}")

	return Success(200, "OK", buffer.Bytes())
}

//=================================================================================================
//========================================================================================== CREATE
// Creates a Message by ID
//
func (cc *HelloWorld) create(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	if rc := Validate("create", args /*args[0]=id*/, "%s", 1, 64 /*args[1]=DStatus*/, "%s", 1, 255 /*args[2]=Timestamp*/, "%s", 1, 255 /*args[3]=EdgeID*/, "%s", 1, 255); rc.Status > 0 {
		return rc
	}
	id := strings.ToLower(args[0])
	doc := &Doc{DStatus: args[1], Timestamp: args[2], EdgeID:args[3]}



	// Write the message
	if err := stub.PutState(id, doc.ToJson()); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusCreated, "Entry Created", nil)
}

//=================================================================================================
//========================================================================================== UPDATE
// Updates a Message by ID
//
func (cc *HelloWorld) update(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	if rc := Validate("update", args /*args[0]=id*/, "%s", 1, 64 /*args[1]=DStatus*/, "%s", 1, 255 /*args[2]=Timestamp*/, "%s", 1, 255 /*args[3]=EdgeID*/, "%s", 1, 255); rc.Status > 0 {
		return rc
	}
	id := strings.ToLower(args[0])
	doc := &Doc{DStatus: args[1], Timestamp:args[2], EdgeID:args[3]}

	// Validate that this ID exist
	if value, err := stub.GetState(id); err != nil || value == nil {
		return Error(http.StatusNotFound, "Not Found")
	}

	// Write the message
	if err := stub.PutState(id, doc.ToJson()); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusNoContent, "Message Updated", nil)
}

//=================================================================================================
//========================================================================================== DELETE
// Delete Message by ID
//
func (cc *HelloWorld) delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	if rc := Validate("delete", args /*args[0]=id*/, "%s", 1, 64); rc.Status > 0 {
		return rc
	}
	id := strings.ToLower(args[0])

	// Validate that this ID exist
	if value, err := stub.GetState(id); err != nil || value == nil {
		return Error(http.StatusNotFound, "Not Found")
	}

	// Delete the message
	if err := stub.DelState(id); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusNoContent, "Message Deleted", nil)
}

//=================================================================================================
//========================================================================================= HISTORY
// Return history by ID
//
func (cc *HelloWorld) history(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	if rc := Validate("history", args /*args[0]=id*/, "%s", 1, 64); rc.Status > 0 {
		return rc
	}
	id := strings.ToLower(args[0])

	// Read history
	resultsIterator, err := stub.GetHistoryForKey(id)
	if err != nil {
		return Error(http.StatusNotFound, "Not Found")
	}
	defer resultsIterator.Close()

	// Write return buffer
	var buffer bytes.Buffer
	buffer.WriteString("{ \"values\": [")
	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()
		if buffer.Len() > 15 {
			buffer.WriteString(",")
		}
		var doc Doc
		//buffer.WriteString("{\"timestamp\":\"")
		//buffer.WriteString(time.Unix(it.Timestamp.Seconds, int64(it.Timestamp.Nanos)).Format(time.Stamp))
		buffer.WriteString("{\"DStatus\":\"")
		buffer.WriteString(doc.FromJson(it.Value).DStatus)
		buffer.WriteString("\", \"Timestamp\":\"")
		buffer.WriteString(doc.FromJson(it.Value).Timestamp)
		buffer.WriteString("\", \"EdgeID\":\"")
		buffer.WriteString(doc.FromJson(it.Value).EdgeID)
		buffer.WriteString("\"}")
	}
	buffer.WriteString("]}")

	return Success(200, "OK", buffer.Bytes())
}

//=================================================================================================
//========================================================================================== SEARCH
// Search for all matching IDs, given a (regex) value expression and return both the IDs and Status.
// For example: '^H.llo' will match any string starting with 'Hello' or 'Hallo'.
//
func (cc *HelloWorld) search(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	if rc := Validate("search", args /*args[0]=searchString*/, "%s", 1, 64); rc.Status > 0 {
		return rc
	}
	searchString := strings.Replace(args[0], "\"", ".", -1) // protect against SQL injection

	// stub.GetQueryResult takes a verbatim CouchDB (assuming this is used DB). See CouchDB documentation:
	//     http://docs.couchdb.org/en/2.0.0/api/database/find.html
	// For example:
	//	{
	//		"selector": {
	//			"value": {"$regex": %s"}
	//		},
	//		"fields": ["ID","value"],
	//		"limit":  99
	//	}
	queryString := fmt.Sprintf("{\"selector\": {\"DStatus\": {\"$regex\": \"%s\"}}, \"fields\": [\"DStatus\"], \"limit\":99}", strings.Replace(searchString, "\"", ".", -1))
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	// Write return buffer
	var buffer bytes.Buffer
	buffer.WriteString("{ \"values\": [")
	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()
		if buffer.Len() > 15 {
			buffer.WriteString(",")
		}
		var doc Doc
		buffer.WriteString("{\"id\":\"")
		buffer.WriteString(it.Key)
		buffer.WriteString("\", \"DStatus\":\"")
		buffer.WriteString(doc.FromJson(it.Value).DStatus)
		buffer.WriteString("\", \"Timestamp\":\"")
		buffer.WriteString(doc.FromJson(it.Value).Timestamp)
		buffer.WriteString("\", \"EdgeID\":\"")
		buffer.WriteString(doc.FromJson(it.Value).EdgeID)
		buffer.WriteString("\"}")
	}
	buffer.WriteString("]}")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}
