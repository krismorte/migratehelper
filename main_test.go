package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestValidateAndLoadFile(t *testing.T) {

	testEnvFile := ".envTestFile"
	testEnvFileInvalid := ".envTestFileInvalid"
	testEnvContent := "CONN_TYPE=FILE\nCONN_INFO=./test/\nSQL_PATH=./test/sql\nFLYWAY_CONF_PATH=./test/\nDB_USER=root\nDB_PASS=localhost\nDB_URL=123456"
	testEnvContentInvalid := "CONN_TYPE=FILE\nCONN_INFO=./test/\nSQL_PATH=./sql\nFLYWAY_CONF_PATH=./test/\nDB_USER=root\nDB_PASS=localhost\nDB_URL=123456"
	createFileHelper(testEnvFile, testEnvContent)
	createFileHelper(testEnvFileInvalid, testEnvContentInvalid)

	err := validateAndLoadFile(testEnvFile)
	if err != nil {
		t.Error("Valid .env file type FILE failed")
	}
	os.Clearenv()
	err = validateAndLoadFile(testEnvFileInvalid)
	if err == nil {
		t.Error("Invalid .env file type FILE read as valid")
	}

	removeFileHelper(testEnvFile)
	removeFileHelper(testEnvFileInvalid)
}

func TestValidateAndLoadSecret(t *testing.T) {
	testEnvFile := ".envTestSecret"
	testEnvFileInvalid := ".envTestSecretInvalid"
	testEnvContent := "CONN_TYPE=AWSSECRET\nCONN_INFO=test-secret\nSQL_PATH=./test/sql\nFLYWAY_CONF_PATH=./test/\nAWSREGION=us-east-1\nLOCALURL=http://localhost:4566"
	testEnvContentInvalid := "CONN_TYPE=AWSSECRET\nCONN_INFO=test-secret\nSQL_PATH=./sql\nFLYWAY_CONF_PATH=./test/\nAWSREGION=us-east-1\nLOCALURL=http://localhost:4566"
	createFileHelper(testEnvFile, testEnvContent)
	createFileHelper(testEnvFileInvalid, testEnvContentInvalid)
	os.Clearenv()
	err := validateAndLoadFile(testEnvFile)
	if err != nil {
		fmt.Println(err)
		t.Error("Valid .env file type SECRET failed")
	}
	os.Clearenv()
	err = validateAndLoadFile(testEnvFileInvalid)
	if err == nil {
		t.Error("Invalid .env file type SECRET read as valid")
	}

	removeFileHelper(testEnvFile)
	removeFileHelper(testEnvFileInvalid)
}

func TestValidateCommands(t *testing.T) {
	validCommands := []string{"lint", "flyway-conf", "validate", "send-log"}
	invalidCommands := []string{"linte", "conf", "cmd"}

	for _, v := range validCommands {
		err := validateCommands(v)
		if err != nil {
			t.Errorf("Doesn't recognized %s command", v)
		}
	}

	for _, v := range invalidCommands {
		err := validateCommands(v)
		if err == nil {
			t.Errorf("Recognized a invalid command %s as valid", v)
		}
	}

}

func TestCheckIfIsTest(t *testing.T) {
	testEnvFile := ".envTestMode"
	testEnvFilePrd := ".envTestModePROD"
	testEnvFileInvalid := ".envTestModeInvalid"
	testEnvContent := "MODE=DEV"
	testEnvContentPrd := "MODE=PROD"
	testEnvContentInvalid := "MODE=ANOTHER"
	createFileHelper(testEnvFile, testEnvContent)
	createFileHelper(testEnvFilePrd, testEnvContentPrd)
	createFileHelper(testEnvFileInvalid, testEnvContentInvalid)

	godotenv.Load(testEnvFile)
	if !checkIfIsTest() {
		t.Error("DEV Mode failed")
	}

	os.Clearenv()
	godotenv.Load(testEnvFilePrd)
	if checkIfIsTest() {
		t.Error("PROD Mode was recognized as DEV!")
	}

	os.Clearenv()
	godotenv.Load(testEnvFileInvalid)
	if checkIfIsTest() {
		t.Error("Invalid Mode was recognized as DEV!")
	}

	removeFileHelper(testEnvFile)
	removeFileHelper(testEnvFilePrd)
	removeFileHelper(testEnvFileInvalid)
}

func TestListFiles(t *testing.T) {

	files, _ := listFiles(".", ".mod")
	if len(files) != 1 {
		t.Error("listFiles faild in finding files!")
	}

	files, _ = listFiles("./test/sql", ".sql")
	if len(files) < 1 {
		t.Error("listFiles faild in finding files!")
	}

	files, _ = listFiles(".", ".sql")
	if len(files) > 0 {
		t.Error("listFiles faild in finding files!")
	}

	_, err := listFiles("NonExistinDir", ".sql")
	if err == nil {
		t.Error("listFiles faild in finding files!")
	}
}

func createFileHelper(fileName, content string) {
	ioutil.WriteFile(fileName, []byte(content), 0644)
}

func removeFileHelper(fileName string) {
	os.Remove(fileName)
}
