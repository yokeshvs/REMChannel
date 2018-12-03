// DISCLAIMER:
// THIS SAMPLE CODE MAY BE USED SOLELY AS PART OF THE TEST AND EVALUATION OF THE SAP CLOUD PLATFORM BLOCKCHAIN SERVICE (THE "SERVICE")
// AND IN ACCORDANCE WITH THE TERMS OF THE TEST AND EVALUATION AGREEMENT FOR THE SERVICE. THIS SAMPLE CODE PROVIDED "AS IS", WITHOUT
// ANY WARRANTY, ESCROW, TRAINING, MAINTENANCE, OR SERVICE OBLIGATIONS WHATSOEVER ON THE PART OF SAP.

// go test --tags nopkcs11

package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type TestEnvironment struct {
	uuid      int
	framework *testing.T
	stub      *shim.MockStub
}

func (test *TestEnvironment) Validate(comment string, expectedReturnCode int32, expectedPayload string, function string, args ...string) {

	fmt.Println("==============================================================================")
	fmt.Println("=== Test:    ", comment)
	fmt.Println("=== Call:    ", function, "(", strings.TrimSpace(strings.Join(args, ",")), ")")
	fmt.Println()

	cc_args := make([][]byte, 1+len(args))
	cc_args[0] = []byte(function)
	for i, arg := range args {
		cc_args[i+1] = []byte(arg)
	}

	test.uuid++
	res := test.stub.MockInvoke(strconv.Itoa(test.uuid), cc_args /*[][]byte{[]byte("function"), []byte("arg1")}*/)

	fmt.Println()
	fmt.Println("=== RetCode: ", res.Status)
	fmt.Println("=== RetMsg:  ", res.Message)

	if res.Status != expectedReturnCode {
		fmt.Println("ERR Unexpected RetCode =", res.Status, "( Expected:", expectedReturnCode, ")")
		test.framework.FailNow()
	}

	fmt.Println("=== Payload: ", string(res.Payload))

	if len(expectedPayload) > 0 && string(res.Payload) != expectedPayload {
		fmt.Println("ERR Unexpected Payload =", string(res.Payload), "( Expected:", expectedPayload, ")")
		test.framework.FailNow()
	}

	fmt.Println()
}

func TestHelloWorld(t *testing.T) {

	// Test Environment
	testEnv := &TestEnvironment{framework: t, stub: shim.NewMockStub("helloWorldStub", new(HelloWorld))}

	// Initialize the chaincode
	if res := testEnv.stub.MockInit("000", nil); res.Status != shim.OK {
		fmt.Println(res.Status, "Init failed ==> ", string(res.Message))
		testEnv.framework.FailNow()
	}

	// Invalid Function
	testEnv.Validate("Invalid function call", 501, "", "invalidFunction")

	// Exist
	testEnv.Validate("Not enough parameters", 400, "", "exist")
	testEnv.Validate("Too many parameters", 400, "", "exist", "ID001", "Hello World")
	testEnv.Validate("Parameter length too short", 400, "", "exist", "")
	testEnv.Validate("Non-existing ID", 404, "", "exist", "ID001")

	// Create
	testEnv.Validate("Create call", 201, "", "create", "ID001", "Hello World")
	testEnv.Validate("Validate ID001 exist", 204, "", "exist", "ID001")
	testEnv.Validate("Create call on existing ID", 409, "", "create", "ID001", "Hello World")

	// Update
	testEnv.Validate("Update existing", 204, "", "update", "ID001", "Hello World!")
	testEnv.Validate("Read updated ID001", 200, "{\"Text\":\"Hello World!\"}", "read", "ID001")
	testEnv.Validate("Update non-existing ID", 404, "", "update", "ID002", "Hallo World!")

	// Delete
	testEnv.Validate("Create new ID003", 201, "", "create", "ID003", "ID to be deleted!")
	testEnv.Validate("Validate ID003 exist", 204, "", "exist", "ID003")
	testEnv.Validate("Delete ID003", 204, "", "delete", "ID003")
	testEnv.Validate("Validate ID003 does NOT exist", 404, "", "exist", "ID003")

	// History/Search
	// ... the History and Search API calls are not supported in the mock interface
}
