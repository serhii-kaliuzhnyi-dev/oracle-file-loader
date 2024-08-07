package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/serhii-kaliuzhnyi-dev/oracle-file-uploader/config"
	"github.com/serhii-kaliuzhnyi-dev/oracle-file-uploader/convertor"
	"github.com/serhii-kaliuzhnyi-dev/oracle-file-uploader/db"
	"github.com/serhii-kaliuzhnyi-dev/oracle-file-uploader/sqlldr"
	"github.com/serhii-kaliuzhnyi-dev/oracle-file-uploader/util"
)

func main() {
	config, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	// Define paths for the original and converted CSV files
	convertedFilePath := util.GenerateConvertedFilePath(config.FilePath)

	// Convert the CSV file to ANSI encoding
	err = convertor.ConvertFileToANSI(config.FilePath, convertedFilePath)
	
	if err != nil {
		log.Fatalf("error converting CSV file to ANSI: %v", err)
	}

	delimiter, err := util.DetectDelimiter(convertedFilePath)
	if err != nil {
		log.Fatalf("error detecting delimiter: %v", err)
	}

	tableConfig, err := db.GenerateTableConfig(convertedFilePath, config.TableName, delimiter)
	if err != nil {
		log.Fatalf("error generating table config: %v", err)
	}

	tableConfigJSON, err := json.MarshalIndent(tableConfig, "", "  ")
	if err != nil {
		log.Fatalf("error marshalling table config to JSON: %v", err)
	}

	fmt.Printf("Table Config:\n%s\n", string(tableConfigJSON))

	// Generate dynamic filename for table config
	tableConfigFilePath := filepath.Join(filepath.Dir(config.FilePath), fmt.Sprintf("%s.config", config.TableName))
	tableConfigFile, err := os.Create(tableConfigFilePath)
	if err != nil {
		log.Fatalf("error creating table config file: %v", err)
	}
	defer tableConfigFile.Close()

	err = sqlldr.GenerateCtlFile(convertedFilePath, config.CtlFilePath, config.TableName, tableConfig, delimiter)
	if err != nil {
		log.Fatalf("error generating .ctl file: %v", err)
	}

	// err = db.CreateTableFromConfig(config.DBUser, config.DBPassword, config.DBUrl, tableConfig)
	// if err != nil {
	// 	log.Fatalf("error creating table: %v", err)
	// }

	// output, err := sqlldr.RunSQLLoader(config.DBUser, config.DBPassword, config.DBUrl, config.CtlFilePath)
	// if err != nil {
	// 	log.Fatalf("error running sqlldr: %v", err)
	// }

	// fmt.Println("Output from SQLLoader:\n", output)

	// fmt.Println("Data uploaded successfully")

	// // Remove the converted file after upload
	// err = os.Remove(convertedFilePath)
	// if err != nil {
	// 	log.Fatalf("error removing converted file: %v", err)
	// }

	// fmt.Println("Converted file removed successfully")

	// 	// Wait for user input before closing
	// fmt.Println("Press Enter to exit...")
	// bufio.NewReader(os.Stdin).ReadBytes('\n')
}
