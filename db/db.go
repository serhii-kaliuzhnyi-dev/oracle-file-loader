package db

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/godror/godror"
	"github.com/serhii-kaliuzhnyi-dev/oracle-file-uploader/util"
)

type ColumnInfo struct {
	OriginalName string `json:"original_name"`
	Type   string `json:"type"`
	Length int    `json:"length"`
	Create bool   `json:"create"`
}

type TableConfig struct {
	Columns      map[string]ColumnInfo `json:"columns"`
	Metadata     Metadata              `json:"metadata"`
	ColumnsOrder []string              `json:"columns_order"`
}

type Metadata struct {
	RowCount  int    `json:"rowCount"`
	TableName string `json:"tableName"`
}

func GenerateTableConfig(filePath string, tableName string, delimiter rune) (TableConfig, error) {
	file, err := os.Open(filePath)

	if err != nil {
		return TableConfig{}, err
	}

	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = delimiter
	records, err := reader.ReadAll()

	if err != nil {
		return TableConfig{}, err
	}

	originalHeaders := records[0]
	headers := util.TransliterateHeaders(originalHeaders)
	for i, header := range headers {
		headers[i] = util.ToLowerSnakeCase(header)
	}

	columnInfo := make(map[string]*ColumnInfo, len(headers))

	for i, header := range headers {
		columnInfo[header] = &ColumnInfo{
			Type: "VARCHAR2", 
			Length: 1, // Set minimum length for oracle columns is 1
			Create: originalHeaders[i] != "",
			OriginalName: originalHeaders[i],
		}
	}

	for _, row := range records[1:] {
		for i, value := range row {
			header := headers[i]
			col := columnInfo[header]
			if len(value) > col.Length {
				col.Length = len(value)
			}
			if col.Type != "VARCHAR2" { // If already set to VARCHAR2, no need to check further
				if util.IsNumeric(value) {
					if len(value) > 11 {
						col.Type = "VARCHAR2" // Handle long numbers as VARCHAR2
					} else {
						col.Type = "NUMBER"
					}
				} else {
					col.Type = "VARCHAR2"
				}
			}
		}
	}

	result := make(map[string]ColumnInfo, len(headers))
	for header, col := range columnInfo {
		result[header] = *col
	}

	tableConfig := TableConfig{
		Columns:      result,
		Metadata:     Metadata{RowCount: len(records) - 1, TableName: tableName},
		ColumnsOrder: headers,
	}

	return tableConfig, nil
}

func CreateTableFromConfig(user, password, dsn string, tableConfig *TableConfig) error {
	connString := fmt.Sprintf("%s/%s@%s", user, password, dsn)
	db, err := sql.Open("godror", connString)
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}
	defer db.Close()

	tableExists, err := checkIfTableExists(db, tableConfig.Metadata.TableName)
	if err != nil {
		return fmt.Errorf("error checking if table exists: %v", err)
	}

	if tableExists {
		log.Printf("Table %s already exists", tableConfig.Metadata.TableName)
		return nil
	}

	createTableSQL := GenerateCreateTableSQL(tableConfig)
	log.Printf("SQL script for creating a table: %s", createTableSQL)

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error executing script for table: %v", err)
	}

	log.Printf("Table %s created successfully", tableConfig.Metadata.TableName)
	return nil
}

func checkIfTableExists(db *sql.DB, tableName string) (bool, error) {
	query := `
	SELECT COUNT(*) 
	FROM all_tables 
	WHERE table_name = :1
	`
	var count int
	err := db.QueryRow(query, strings.ToUpper(tableName)).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}


// GenerateCreateTableSQL generates a SQL CREATE TABLE statement from the given TableConfig.
func GenerateCreateTableSQL(tableConfig *TableConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableConfig.Metadata.TableName))
	first := true
	for _, colName := range tableConfig.ColumnsOrder {
		colInfo := tableConfig.Columns[colName]
		if !colInfo.Create {
			continue
		}
		if !first {
			sb.WriteString(",\n")
		}
		first = false
		if colInfo.Type == "NUMBER" || colInfo.Type == "NUMERIC" || colInfo.Type == "DATE" || colInfo.Type == "TIMESTAMP WITH TIME ZONE" {
			sb.WriteString(fmt.Sprintf("  %s %s", colName, colInfo.Type))
		} else {
			sb.WriteString(fmt.Sprintf("  %s VARCHAR2(%d)", colName, colInfo.Length))
		}
	}
	sb.WriteString("\n);")
	return sb.String()
}