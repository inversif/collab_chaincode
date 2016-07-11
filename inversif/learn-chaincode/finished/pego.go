/*
Copyright IBM Corp 2016 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var projectIndexStr = "_projectindex" //name for the key that will store list of project index/project name
var employeeIndexStr = "_employeeindex" //name for the key that will store list of employee index/employeeID

//same as employee
type Member struct{
	MemberID string `json:"memberid"`
	MemberName string `json:"membername"`
    JobTitle string `json:"jobtitle"`
    Level int `json:"level"`
    JobGroup string `json:"jobgroup"`
}

type Project struct{
	Name string `json:"name"`		  // Project name
    Members []string `json:"members"` // ID
}

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	err := stub.PutState("hello_world", []byte(args[0]))
	if err != nil {
		return nil, err
	}

return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, 
		function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "write" {											    //writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "add_employee"{									//add new employee
        return t.add_employee(stub, args)
    } else if function == "update_employee" {								//update attributes of existing employee
		return t.update_employee(stub, args)
	} else if function == "create_project"{									//create new project
		return t.create_project(stub, args)
	} else if function == "add_project_member"{								//add new member to the project
		return t.add_project_member(stub, args)
	} else if function == "delete_project_member"{							//delete member from a project
		return t.delete_project_member(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation")
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query")
}

// read - query function to read key/value pair
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// Get Employee's information, and returns a Member struct and necessary error
// if error exists.
func InquireEmployee (stub *shim.ChaincodeStub, args []string) (Member, error) {
	//   0       1       2           3          4
	// "id", "name", "job title", "level", "job group"
	if len(args) != 5 {
        fmt.Println("Incorrect number of arguments. Expecting 5")
		return Member{}, errors.New("Incorrect number of arguments. Expecting 5")
	}

	//input sanitation
	fmt.Println("- start adding new employee(s)")
	if len(args[0]) <= 0 || len(args[1]) <= 0 || len(args[2]) <= 0 || len(args[3]) <= 0  || len(args[4]) <= 0  {
		return Member{}, errors.New("All arguments must be non-empty string")
	}

	//check if employee is exists
	employeeAsBytes, err := stub.GetState(args[0])	//get employee detail from chaincode state
	if err != nil {
		return Member{}, errors.New("Failed to get employee")
	}

	employee := Member{}
	json.Unmarshal(employeeAsBytes, &employee)

	return employee, nil
}

func AssignToEmployee(id string, name string, title string, level string, 
		group string, nemployee *Member){
	nemployee.MemberID = id
    nemployee.MemberName = name
    nemployee.JobTitle = title
    nemployee.JobGroup = group
    nemployee.Level, _ = strconv.Atoi(level)
}

// TODO: Might have to change key's type
// Function to invoke Marshal & PutState consecutively.
func PutBack(stub *shim.ChaincodeStub, employee Member, key int) ([]byte, error) {
	employeeAsBytes, _ := json.Marshal(employee)

	strkey := strconv.Itoa(key)
	err := stub.PutState(strkey, employeeAsBytes) //write the new employee to the chaincode state
	if err != nil {
		return nil, errors.New("Error on PutState")
	}
	return employeeAsBytes, nil
}

func (t *SimpleChaincode) update_employee(stub *shim.ChaincodeStub, args []string) ([]byte, error){
	employeeObj, err := InquireEmployee(stub, args)
	if err != nil{
		fmt.Println(err)
		return nil, err
	}

	person_id, _ := strconv.Atoi(args[0])
	AssignToEmployee(args[0], args[1], args[2], args[3], args[4], &employeeObj)

	_, errvar := PutBack(stub, employeeObj, person_id)
    if errvar != nil {
		fmt.Println(errvar)
    	return nil, errvar
    }

    return nil, nil
}

func (t *SimpleChaincode) add_employee(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	//   0       1       2           3          4
	// "id", "name", "job title", "level", "job group"
	new_employee, err := InquireEmployee(stub, args)
	if err != nil { 
		fmt.Println(err)
		return nil, err	// TODO: Think of a good alternative to the current return value.
	}

	// if emp_exist := FindEmployee(new_employee.MemberID, args[0]); emp_exist {
	if new_employee.MemberID == args[0] {
		fmt.Println("This employee arleady exists: ", args[0])
		return nil, errors.New("This employee arleady exists")
	}

	conv_id, _ := strconv.Atoi(args[0])
	AssignToEmployee(args[0], args[1], args[2], args[3], args[4], &new_employee)

    employeeIndexAsBytes, errvar := PutBack(stub, new_employee, conv_id)
    if errvar != nil {
    	fmt.Println(errvar)
    	return nil, errors.New("")
    }
	
	anyerror := GetIndex(employeeIndexAsBytes, stub, args[0], false)
	if anyerror != nil {
		anyerror = errors.New("Failed to retrieve employee's index")
		return nil, anyerror
	}

	fmt.Println("- end add employee")
	return nil, nil
}

// Get Project's information, and returns a Project struct and necessary error
// if error exists.
func InquireProject (stub *shim.ChaincodeStub, argument string) (Project, error) {
	//get project from chaincode state
	projectAsBytes, err := stub.GetState(argument);

	if err != nil{
		fmt.Println("Failed to get project")
		return Project{}, errors.New("Failed to get project")
	}

	project := Project{}
	json.Unmarshal(projectAsBytes, &project)	//equals to JSON.parse()

	return project, nil
}

func GetIndex(somethingAsBytes []byte, stub *shim.ChaincodeStub, 
		name string, whichone bool) error{
	var indexList []string
	var err error
	json.Unmarshal(somethingAsBytes, &indexList)
	
	//append
	indexList = append(indexList, name)
	fmt.Println("! index: ", indexList)
	jsonAsBytes, _ := json.Marshal(indexList)
	// projectIndexStr is a global variable
	if whichone == true{
		err = stub.PutState(projectIndexStr, jsonAsBytes)	// TODO: Might want to reconsider return value(s)
	} else {
		err = stub.PutState(employeeIndexStr, jsonAsBytes)
	}

	return err
}

// Initiate a project
// TODO: Check if key & name is arg[0].
func (t *SimpleChaincode) create_project(stub *shim.ChaincodeStub, args []string) ([]byte, error){

	//input sanitation
	fmt.Println("- start creating project")
	if (len(args) != 1) && (len(args[0]) <= 0) { 
		fmt.Println("Invalid argument! Consider these cases \n1. Incorrect number of argument 1\n2. argument must be a non-empty string\n")
		return nil, errors.New("Incorrect invocation of function! See log!")
	}

	//get the project index
	projectAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get project index")
	}

	res := Project{}
	json.Unmarshal(projectAsBytes, &res)	//equals to JSON.parse()

	//check if project already exists
	if res.Name == args[0]{
		fmt.Println("This project arleady exists: ", args[0])
		return nil, errors.New("This project already exists")
	}

	res.Name = args[0]
	
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)	//String replace is used for get rid of white space, because key can contain white space
	if err != nil {
		return nil, err
	}

	//get the project index
	projectAsBytes, err = stub.GetState(projectIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get project index")
	}
	
	anyerror := GetIndex(projectAsBytes, stub, args[0], true)
	if anyerror != nil{
		fmt.Println(anyerror)
		return nil, anyerror
	}

	fmt.Println("- end create project")
	return nil, nil
}

func (t *SimpleChaincode) add_project_member(stub *shim.ChaincodeStub, 
		args []string) ([]byte, error){
	new_project, err := InquireProject(stub, args[0])

	for i:=1; i < len(args); i++ {	// TODO: Might be off by 1, prolly start at 0
		isExists := 0 //0 means member still not in this project

		if len(new_project.Members) == 0 {
			new_project.Members = append(new_project.Members, args[i])	//append memberID/employeeID to project members array 
			fmt.Println("! Success add new member: ", args[i])
		}

		for j := range new_project.Members{
			if args[i] == new_project.Members[j] {
				isExists = 1 //1 means member already exists in this project
				break
			}
		}

		if isExists == 0 {
			new_project.Members = append(new_project.Members, args[i])	//append memberID/employeeID to project members array 
			fmt.Println("! Success add new member: ", args[i])
		}
	}

	jsonAsBytes, _ := json.Marshal(new_project)	//equals to JSON.stringify
	err = stub.PutState(args[0], jsonAsBytes)	//rewrite project to the chaincode state

	if err != nil {
		return nil, err
	}
	
	fmt.Println("- end add new member")
	return nil, nil
}

func (t *SimpleChaincode) delete_project_member(stub *shim.ChaincodeStub, 
		args []string) ([]byte, error){
	//   0                  1
	// "project name", "member id"
	if len(args) != 2 {
		fmt.Println("Incorrect number of arguments. Expecting 2")
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	new_project, err := InquireProject(stub, args[0])

	//remove member from project
	for i := range new_project.Members{
		//looking for member ID
		if new_project.Members[i] == args[1]{
			fmt.Println("member found")
			new_project.Members = append(new_project.Members[:i], new_project.Members[i+1:]...)			//remove it
			break
		}
	}

	projectAsBytes, _ := json.Marshal(new_project)	//stringify
	err = stub.PutState(args[0], projectAsBytes)	//rewrite project to the chaincode state

	if err != nil {
		return nil, errors.New("Failed to delete member from project chaincode state")
	}

	return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}
