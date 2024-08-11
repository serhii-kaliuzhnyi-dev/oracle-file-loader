package main

import (
	"encoding/json"
	"flag"
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
	planCmd := flag.NewFlagSet("plan", flag.ExitOnError)
	applyCmd := flag.NewFlagSet("apply", flag.ExitOnError)
	autoApprove := applyCmd.Bool("auto-approve", false, "Automatically approve the plan without prompt")
	skipTable := applyCmd.Bool("skip-table", false, "Skip table creation")

	flag.Parse()

	if len(os.Args) < 2 {
		log.Println("expected 'plan' or 'apply' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "plan":
		planCmd.Parse(os.Args[2:])
		handlePlan()
	case "apply":
		applyCmd.Parse(os.Args[2:])
		handleApply(*autoApprove, *skipTable)
	default:
		log.Println("expected 'plan' or 'apply' subcommands")
		os.Exit(1)
	}
}

func handlePlan() {
	cfg := loadConfig()
	delimiter := detectDelimiter(cfg.FilePath)
	utf8FilePath := util.GenerateUtf8FilePath(cfg.FilePath)
	convertedFilePath := util.GenerateConvertedFilePath(cfg.FilePath)

	// Step 1: Convert the original file from Windows-1251 to UTF-8
	err := convertor.ConvertFileToUtf8(cfg.FilePath, utf8FilePath)
	if err != nil {
		log.Fatalf("error converting file to UTF-8: %v", err)
	}

	tableConfigFilePath := getTableConfigFilePath(cfg)
	var tableConfig *db.TableConfig

	// Step 2: Interact with the UTF-8 file and load the configuration
	if fileExists(tableConfigFilePath) {
		log.Printf("Table config file found: %s", tableConfigFilePath)
		tableConfig = loadTableConfigFromFile(tableConfigFilePath)

		// Remove columns marked as create: false from the UTF-8 file
		err = convertor.FilterConvertedFile(utf8FilePath, tableConfig, delimiter)
		if err != nil {
			log.Fatalf("error filtering UTF-8 file: %v", err)
		}
	} else {
		// Step 3: Generate the table configuration using the UTF-8 file
		tableConfig = generateTableConfig(utf8FilePath, cfg, delimiter)
		saveTableConfigToFile(cfg, tableConfig)
	}

	// Step 4: Convert the filtered UTF-8 file back to Windows-1251
	err = convertor.ConvertFileToANSI(utf8FilePath, convertedFilePath)
	if err != nil {
		log.Fatalf("error converting file back to Windows-1251: %v", err)
	}

	// Step 5: Regenerate the .ctl file using the Windows-1251 file name
	generateCtlFile(cfg, tableConfig, delimiter)

	// Step 6: Generate and display the SQL statement for the table
	sqlStatement := db.GenerateCreateTableSQL(tableConfig)
	fmt.Printf("Planned table:\n%s\n", sqlStatement)

	// Step 7: Clean up by removing the temporary UTF-8 file
	err = os.Remove(utf8FilePath)
	if err != nil {
		log.Fatalf("error removing UTF-8 file: %v", err)
	}
	log.Printf("Temporary UTF-8 file removed: %s\n", utf8FilePath)
}


func handleApply(autoApprove, skipTable bool) {
	cfg := loadConfig()
	tableConfigFilePath := getTableConfigFilePath(cfg)

	if !fileExists(tableConfigFilePath) {
		log.Fatalf("Table config file not found. Please run the 'plan' command first.")
	}

	tableConfig := loadTableConfigFromFile(tableConfigFilePath)

	if !skipTable {
		sqlStatement := db.GenerateCreateTableSQL(tableConfig)
		fmt.Printf("The following table will be created:\n%s\n", sqlStatement)

		if !autoApprove && !confirmCreation() {
			fmt.Println("Operation cancelled.")
			return
		}

		createTable(cfg, tableConfig)
	} else {
		log.Println("Table creation skipped due to --skip-table flag.")
	}

	runSQLLoader(cfg)
}

func loadConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	return cfg
}

func generateTableConfig(filePath string, cfg *config.Config, delimiter rune) *db.TableConfig {
	tableConfig, err := db.GenerateTableConfig(filePath, cfg.TableName, delimiter)
	if err != nil {
		log.Fatalf("error generating table config: %v", err)
	}
	return &tableConfig
}

func saveTableConfigToFile(cfg *config.Config, tableConfig *db.TableConfig) {
	tableConfigJSON := marshalTableConfig(tableConfig)
	tableConfigFilePath := getTableConfigFilePath(cfg)
	saveToFile(tableConfigFilePath, tableConfigJSON)
	log.Printf("Table configuration saved to %s\n", tableConfigFilePath)
}

func generateCtlFile(cfg *config.Config, tableConfig *db.TableConfig, delimiter rune) {
	convertedFilePath := util.GenerateConvertedFilePath(cfg.FilePath)
	ctlFilePath := getCtlFilePath(cfg)

	// Use the converted file name in the INFILE directive
	infile := fmt.Sprintf("INFILE '%s'", filepath.Base(convertedFilePath))
	err := sqlldr.GenerateCtlFile(convertedFilePath, ctlFilePath, cfg.TableName, tableConfig, delimiter, infile)
	if err != nil {
		log.Fatalf("error generating .ctl file: %v", err)
	}
	log.Printf("SQL*Loader control file generated: %s\n", ctlFilePath)
}

func confirmCreation() bool {
	fmt.Print("Are you sure you want to create it? (yes/no): ")
	var response string
	fmt.Scanln(&response)
	return response == "yes"
}

func createTable(cfg *config.Config, tableConfig *db.TableConfig) {
	err := db.CreateTableFromConfig(cfg.DBUser, cfg.DBPassword, cfg.DBUrl, tableConfig)
	if err != nil {
		log.Fatalf("error creating table: %v", err)
	}
}

func runSQLLoader(cfg *config.Config) {
	output, err := sqlldr.RunSQLLoader(cfg.DBUser, cfg.DBPassword, cfg.DBUrl, getCtlFilePath(cfg))
	if err != nil {
		log.Fatalf("error running sqlldr: %v", err)
	}
	fmt.Println("Output from SQLLoader:\n", output)
	fmt.Println("Data uploaded successfully")
}

func detectDelimiter(filePath string) rune {
	delimiter, err := util.DetectDelimiter(filePath)
	if err != nil {
		log.Fatalf("error detecting delimiter: %v", err)
	}
	return delimiter
}

func marshalTableConfig(tableConfig *db.TableConfig) []byte {
	tableConfigJSON, err := json.MarshalIndent(tableConfig, "", "  ")
	if err != nil {
		log.Fatalf("error marshalling table config to JSON: %v", err)
	}
	return tableConfigJSON
}

func getTableConfigFilePath(cfg *config.Config) string {
	return filepath.Join(filepath.Dir(cfg.FilePath), fmt.Sprintf("%s.config.json", cfg.TableName))
}

func getCtlFilePath(cfg *config.Config) string {
	return filepath.Join(filepath.Dir(cfg.FilePath), fmt.Sprintf("%s.ctl", cfg.TableName))
}

func saveToFile(filePath string, data []byte) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		log.Fatalf("error writing to file: %v", err)
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func loadTableConfigFromFile(filePath string) *db.TableConfig {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error reading table config file: %v", err)
	}
	var tableConfig db.TableConfig
	err = json.Unmarshal(data, &tableConfig)
	if err != nil {
		log.Fatalf("error unmarshalling table config JSON: %v", err)
	}
	return &tableConfig
}
