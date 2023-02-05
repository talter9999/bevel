package supplychain

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/chaincode/common"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type marble struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Color      string `json:"color"`
	Size       string  `json:"size"`
	Owner      string `json:"owner"`
}


// createProduct creates a new Product on the blockchain using the  with the supplied ID
func (s *SmartContract) createProduct(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	identity, err := GetInvokerIdentity(stub)
	if err != nil {
		shim.Error(fmt.Sprintf("Error getting invoker identity: %s\n", err.Error()))
	}
	s.logger.Infof("%+v\n", identity.Cert.Subject.String())

	if !identity.CanInvoke("createProduct") {
		return peer.Response{
			Status:  403,
			Message: fmt.Sprintf("You are not authorized to perform this transaction, cannot invoke createProduct"),
		}
	}

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// Create ProductRequest struct from input JSON.
	argBytes := []byte(args[0])
	var request ProductRequest
	if err := json.Unmarshal(argBytes, &request); err != nil {
		return shim.Error(err.Error())
	}
	//Check if product  state using id as key exsists
	testProductAsBytes, err := stub.GetState(request.ID)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Return 403 if item exisits
	if len(testProductAsBytes) != 0 {
		return peer.Response{
			Status:  403,
			Message: fmt.Sprintf("Existing Product %s Found", args[0]),
		}
	}

	product := Product{
		ID:           request.ID,
		Type:         "product",
		Name:         request.ProductName, sampleproduct
		Health:       "",  good
		Metadata:     request.Metadata, misc
		Location:     request.Location, india
		Sold:         false, false
		Recalled:     false, false
		ContainerID:  "", abac
		Custodian:    identity.Cert.Subject.String(), org1
		Timestamp:    int64(s.clock.Now().UTC().Unix()), 1532009163
		Participants: request.Participants, abc
	}

	product.Participants = append(product.Participants, identity.Cert.Subject.String())

	// Put new Product onto blockchain
	productAsBytes, _ := json.Marshal(product)
	if err := stub.PutState(product.ID, productAsBytes); err != nil {
		return shim.Error(err.Error())
	}

	response := map[string]interface{}{
		"generatedID": product.ID,
	}
	bytes, _ := json.Marshal(response)

	s.logger.Infof("Wrote Product: %s\n", product.ID)
	return shim.Success(bytes)
}

//getAllProducts retrieves all products on the ledger
func (s *SmartContract) getAllProducts(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//Get user identity
	identity, err := GetInvokerIdentity(stub)
	if err != nil {
		shim.Error(fmt.Sprintf("Error getting invoker identity: %s\n", err.Error()))
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}

	// Get iterator for all entries
	iterator, err := stub.GetStateByRange("", "")
	if err != nil {
		shim.Error(fmt.Sprintf("Error getting state iterator: %s", err))
	}
	defer iterator.Close()

	// Create array
	var buffer bytes.Buffer
	buffer.WriteString("[")
	for iterator.HasNext() {
		state, iterErr := iterator.Next()
		if iterErr != nil {
			return shim.Error(fmt.Sprintf("Error accessing state: %s", err))
		}

		// Don't return products issuer isn't a party to
		var product Product
		err = json.Unmarshal(state.Value, &product)
		if err != nil && err.Error() != "Not a Product" {
			return shim.Error(err.Error())
		}
		if product.AccessibleBy(identity) {
			if buffer.Len() != 1 {
				buffer.WriteString(",")
			}
			buffer.WriteString(string(state.Value))
		}
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

//getSingleProducts retrieves all products on the ledger
func (s *SmartContract) getSingleProduct(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//get user identity
	identity, err := GetInvokerIdentity(stub)
	if err != nil {
		shim.Error(fmt.Sprintf("Error getting invoker identity: %s\n", err.Error()))
	}

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	//get single state using id as key
	productAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	// Return 404 if result's empty
	if len(productAsBytes) == 0 {
		return peer.Response{
			Status:  404,
			Message: fmt.Sprintf("Product %s Not Found", args[0]),
		}
	}

	//check to see if result is a product or not and unmarsal if so
	var product Product
	err = json.Unmarshal(productAsBytes, &product)
	if err != nil {
		return peer.Response{
			Status:  400,
			Message: fmt.Sprintf("Error: %s ", err),
		}
	}
	//check if user is allowed to see this product
	if !product.AccessibleBy(identity) {
		return peer.Response{
			Status:  404,
			Message: fmt.Sprintf("Product %s Not Found", args[0]),
		}
	}
	return shim.Success(productAsBytes)
}

//getContainerlessProducts retrieves all products on the ledger where containerID is empty
func (s *SmartContract) getContainerlessProducts(stub shim.ChaincodeStubInterface) peer.Response {
	//Get user identity
	identity, err := GetInvokerIdentity(stub)
	if err != nil {
		shim.Error(fmt.Sprintf("Error getting invoker identity: %s\n", err.Error()))
	}

	queryString := "{\"selector\":{\"docType\":\"product\",\"containerID\":\"\"}}"
	// Get iterator for all entries
	iterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		shim.Error(fmt.Sprintf("Error getting query iterator: %s", err.Error()))
	}
	defer iterator.Close()

	// Create array
	var buffer bytes.Buffer
	buffer.WriteString("[")
	for iterator.HasNext() {
		state, iterErr := iterator.Next()
		if iterErr != nil {
			return shim.Error(fmt.Sprintf("Error accessing state: %s", err))
		}

		// Don't return products issuer isn't a party to
		var product Product
		err = json.Unmarshal(state.Value, &product)
		if err != nil {
			s.logger.Errorf("Error unmarshalling product: %s", err)
			return shim.Error(fmt.Sprintf("Error unmarshalling product: %s", err))
		}
		s.logger.Infof("%+v\n", product)
		s.logger.Infof("%+v\n", identity.Cert.Subject.String())
		if product.AccessibleBy(identity) {
			if buffer.Len() != 1 {
				buffer.WriteString(",")
			}
			buffer.WriteString(string(state.Value))
		}
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

//updateCustodian claims current user as the custodian
func (s *SmartContract) updateProductCustodian(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//get user identity
	identity, err := GetInvokerIdentity(stub)
	if err != nil {
		shim.Error(fmt.Sprintf("Error getting invoker identity: %s\n", err.Error()))
	}

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	trackingID := args[0]
	newLocation := args[1]
	newCustodian := identity.Cert.Subject.String()
	//get state by id as key
	existingsBytes, _ := stub.GetState(trackingID)

	// return 404 is not found
	if len(existingsBytes) == 0 {
		return peer.Response{
			Status:  404,
			Message: fmt.Sprintf("Item with trackingID %s not found", trackingID),
		}
	}

	//try to unmarshal as container
	var product Product
	err = json.Unmarshal(existingsBytes, &product)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Ensure user is a participant
	if !(product.AccessibleBy(identity)) {
		return peer.Response{
			Status:  403,
			Message: fmt.Sprintf("You are not authorized to perform this transaction, product not accesible by identity"),
		}
	}
	//Ensure new custodian isnt the same as old one
	if newCustodian == product.Custodian {
		return peer.Response{
			Status:  403,
			Message: fmt.Sprintf("You are already custodian"),
		}
	}

	//make sure user cant claim a product separately from the container
	if product.ContainerID != "" {
		containerBytes, _ := stub.GetState(product.ContainerID)
		var container Container
		err = json.Unmarshal(containerBytes, &container)
		if err != nil {
			return shim.Error(err.Error())
		}
		if container.Custodian != newCustodian {
			return peer.Response{
				Status:  403,
				Message: fmt.Sprintf("Product needs to be unpackaged before claiming a new owner"),
			}
		}
	}

	//change custodian
	product.Custodian = newCustodian
	product.Location = newLocation
	product.Timestamp = int64(s.clock.Now().UTC().Unix())

	newBytes, _ := json.Marshal(product)

	if err := stub.PutState(trackingID, newBytes); err != nil {
		return shim.Error(err.Error())
	}

	s.logger.Infof("Updated state: %s\n", trackingID)
	return shim.Success([]byte(trackingID))

}

func (s *SmartContract) createMarble(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	marbleName := args[0]
	color := args[1]
	owner := args[3]
	size := args[2]

	// ==== Check if marble already exists ====
	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if marbleAsBytes != nil {
		fmt.Println("This marble already exists: " + marbleName)
		return shim.Error("This marble already exists: " + marbleName)
	}

	// ==== Create marble object and marshal to JSON ====
	objectType := "marble"
	marble := &marble{objectType, marbleName, color, size, owner}
	marbleJSONasBytes, err := json.Marshal(marble)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the marble json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + marbleName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//marbleJSONasBytes := []byte(str)

	// === Save marble to state ===
	err = stub.PutState(marbleName, marbleJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	// indexName := "color~name"
	// colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marble.Color, marble.Name})
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	// //  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	// //  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	// value := []byte{0x00}
	// stub.PutState(colorNameIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("Marble added")
	return shim.Success(nil)
}

// ===============================================
// readMarble - read a marble from chaincode state
// ===============================================
func (s *SmartContract) readMarble(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a marble key/value pair from state
// ==================================================
func (s *SmartContract) deleteMarble(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var jsonResp string
	var marbleJSON marble
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	marbleName := args[0]

	// to maintain the color~name index, we need to read the marble first and get its color
	valAsbytes, err := stub.GetState(marbleName) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + marbleName + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + marbleName + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &marbleJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + marbleName + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(marbleName) //remove the marble from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// // maintain the index
	// indexName := "color~name"
	// colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marbleJSON.Color, marbleJSON.Name})
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	// //  Delete index entry to state.
	// err = stub.DelState(colorNameIndexKey)
	// if err != nil {
	// 	return shim.Error("Failed to delete state:" + err.Error())
	// }
	return shim.Success(nil)
}

// ===========================================================
// transfer a marble by setting a new owner name on the marble
// ===========================================================
func (s *SmartContract) transferMarble(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	marbleName := args[0]
	newOwner := args[1]
	fmt.Println("- start transferMarble ", marbleName, newOwner)

	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return shim.Error("Failed to get marble:" + err.Error())
	} else if marbleAsBytes == nil {
		return shim.Error("Marble does not exist")
	}

	marbleToTransfer := marble{}
	err = json.Unmarshal(marbleAsBytes, &marbleToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	marbleToTransfer.Owner = newOwner //change the owner

	marbleJSONasBytes, _ := json.Marshal(marbleToTransfer)
	err = stub.PutState(marbleName, marbleJSONasBytes) //rewrite the marble
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferMarble (success)")
	return shim.Success(nil)
}
