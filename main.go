package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/krismorte/secrets/aws"
	"github.com/stretchr/stew/slice"
)

// Secret
type AWSSECRET struct {
	Username            string
	Password            string
	Engine              string
	Host                string
	Port                int
	Dbname              string
	DbClusterIdentifier string
	Url                 string
	DatadogKey          string `json:"DATADOG_KEY"`
}

func main() {

	if len(os.Args) < 3 {
		fmt.Errorf("Two arguments are waited!!!\nflyway-validate <conf file> <command>")
	}

	err := validateCommands(os.Args[2])
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = validateAndLoadFile(os.Args[1])
	if err != nil {
		log.Fatalln(err.Error())
	}

	if os.Args[2] == "lint" {
		lint()
	}

}

func lint() {

	files, err := listFiles(os.Getenv("SQL_PATH"), ".sql")
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = validateFilesNames(files)
	if err != nil {
		ioutil.WriteFile(os.Getenv("OUTPUT"), []byte(err.Error()), 0644)
		log.Fatalln(err.Error())
	}
}

func generateConf() {
	fileContent := "flyway.sqlMigrationSuffixes=.sql\nflyway.table=flyway_migrations\nflyway.baselineOnMigrate=true\n"

	if os.Getenv("CONN_TYPE") == "FILE" {
		fileContent = fileContent + "flyway.user=" + os.Getenv("DB_USER") + "\n"
		fileContent = fileContent + "flyway.password=" + os.Getenv("DB_PASS") + "\n"
		fileContent = fileContent + "flyway.url=" + os.Getenv("DB_URL") + "\n"
	}

	if os.Getenv("CONN_TYPE") == "AWSSECRET" {
		var secret AWSSECRET
		value := aws.GetSecret("tst-aurorao-migrate")
		b, _ := json.Marshal(value)
		json.Unmarshal(b, &secret)
		fileContent = fileContent + "flyway.user=" + secret.Username + "\n"
		fileContent = fileContent + "flyway.password=" + secret.Password + "\n"
		fileContent = fileContent + "flyway.url=" + secret.Url + "\n"
	}

	ioutil.WriteFile(os.Getenv("FLYWAY_CONF_PATH")+"\\flyway.conf", []byte(fileContent), 0644)
}

func validateFilesNames(files []os.FileInfo) error {
	var versionNumbers []string
	var errorMessages []string
	for _, file := range files {

		if strings.HasPrefix(file.Name(), "v") || strings.HasPrefix(file.Name(), "u") || strings.HasPrefix(file.Name(), "r") {
			errorMessages = append(errorMessages, fmt.Sprintf("Wrong File Name: %s. These are the valids prefix [V,R,U]", file.Name()))
		}
		if !strings.Contains(file.Name(), "__") {
			errorMessages = append(errorMessages, fmt.Sprintf("Wrong File Name: %s. Name is missing 2 underlines", file.Name()))
		}
		underlineIndex := strings.Index(file.Name(), "_")
		actualVersion := file.Name()[1:underlineIndex]

		if slice.Contains(versionNumbers, actualVersion) {
			errorMessages = append(errorMessages, fmt.Sprintf("Duplicate Version at file: %s. Increase version %s", file.Name(), actualVersion))
		}
		versionNumbers = append(versionNumbers, actualVersion)
	}
	if len(errorMessages) > 1 {
		errorMessage := "\n"
		for _, v := range errorMessages {
			errorMessage += v + "\n\n"
		}
		return fmt.Errorf(errorMessage)
	}
	return nil
}

func getSecret(secretID string) AWSSECRET {
	var secret AWSSECRET
	var rawSecret interface{}

	if checkIfIsTest() {
		rawSecret = aws.GetSecretLocally(secretID, os.Getenv("AWSREGION"), os.Getenv("LOCALURL"))
	} else {
		rawSecret = aws.GetSecret(secretID)
	}

	b, _ := json.Marshal(rawSecret)

	json.Unmarshal(b, &secret)

	return secret
}

func validateAndLoadFile(fileName string) error {
	_, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	err = godotenv.Load(fileName)
	if err != nil {
		return err
	}

	if os.Getenv("CONN_TYPE") == "FILE" {
		_, err = os.Stat(os.Getenv("CONN_INFO"))
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(os.Getenv("SQL_PATH"))
	if err != nil {
		return err
	}

	_, err = os.Stat(os.Getenv("FLYWAY_CONF_PATH"))
	if err != nil {
		return err
	}

	return nil
}

func validateCommands(command string) error {
	commands := []string{"lint", "flyway-conf", "validate", "send-log"}
	if !slice.Contains(commands, command) {
		return errors.New("Invalid Command: " + command)
	}
	return nil
}

func listFiles(dir string, extension string) ([]os.FileInfo, error) {
	var filteredFiles []os.FileInfo
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.Contains(file.Name(), extension) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles, nil
}

func checkIfIsTest() bool {

	if os.Getenv("MODE") == "DEV" {
		return true
	}
	return false
}
