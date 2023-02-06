/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Product describes basic details of what makes up a simple asset
type Product struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	Origin      string `json:"origin"`
	Price       int`json:"price"`
}


// AddProduct issues a new product to the world state with given details.
func (s *SmartContract) AddProduct(ctx contractapi.TransactionContextInterface, name string, description string, quantity int, origin string, price int) error {
	exists, err := s.ProductExists(ctx, name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the product %s already exists", name)
	}

	product := Product{
		Name:            name,
		Description:     description,
		Quantity:        quantity,
		Origin:          origin,
		Price:           price,
	}
	assetJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(name, assetJSON)
}

// ViewProduct returns the product stored in the world state with given name.
func (s *SmartContract) ViewProduct(ctx contractapi.TransactionContextInterface, name string) (*Product, error) {
	productJSON, err := ctx.GetStub().GetState(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if productJSON == nil {
		return nil, fmt.Errorf("the product %s does not exist", name)
	}

	var product Product
	err = json.Unmarshal(productJSON, &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// ViewAllProducts returns all products found in world state
func (s *SmartContract) ViewAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var products []*Product
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product Product
		err = json.Unmarshal(queryResponse.Value, &product)
		if err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}


// UpdateProduct updates an existing product in the world state with provided parameters.
func (s *SmartContract) UpdateProduct(ctx contractapi.TransactionContextInterface, name string, description string, quantity int, origin string, price int) error {
	exists, err := s.ProductExists(ctx, name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the product %s does not exist", name)
	}

	// overwriting original asset with new asset
	product := Product{
		Name:            name,
		Description:     description,
		Quantity:        quantity,
		Origin:          origin,
		Price:           price,
	}
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(name, productJSON)
}

// DeleteProduct deletes an given product from the world state.
func (s *SmartContract) DeleteProduct(ctx contractapi.TransactionContextInterface, name string) error {
	exists, err := s.ProductExists(ctx, name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the product %s does not exist", name)
	}

	return ctx.GetStub().DelState(name)
}

// DeleteAllProducts deletes all products from the world state.
func (s *SmartContract) DeleteAllProducts(ctx contractapi.TransactionContextInterface) error {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return  err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return  err
		}

		var existingProduct Product
		err = json.Unmarshal(queryResponse.Value, &existingProduct)
		if err != nil {
			return  err
		}

		errStub := ctx.GetStub().DelState(queryResponse.Key)
		if errStub != nil {
			return  errStub
		}
	}
	return err
}


// ProductExists returns true when product with given name exists in world state
func (s *SmartContract) ProductExists(ctx contractapi.TransactionContextInterface, name string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(name)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferProduct updates the origin field of product with given name in world state.
func (s *SmartContract) TransferProduct(ctx contractapi.TransactionContextInterface, name string, newOrigin string) error {
	product, err := s.ViewProduct(ctx, name)
	if err != nil {
		return err
	}

	product.Origin = newOrigin
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(name, productJSON)
}


func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
