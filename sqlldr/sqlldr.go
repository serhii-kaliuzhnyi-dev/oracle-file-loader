package sqlldr

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/serhii-kaliuzhnyi-dev/oracle-file-uploader/db"
)

func GenerateCtlFile(csvFilePath, ctlFilePath, tableName string, tableConfig *db.TableConfig, delimiter rune, infile string) error {
	var fields []string
	for _, colName := range tableConfig.ColumnsOrder {
		if tableConfig.Columns[colName].Create {
			fields = append(fields, colName)
		}
	}
	fieldsStr := strings.Join(fields, ",\n  ")

	var delimiterStr string
	if delimiter == '\t' {
		delimiterStr = "FIELDS TERMINATED BY X'09' OPTIONALLY ENCLOSED BY '\"'"
	} else {
		delimiterStr = fmt.Sprintf("FIELDS TERMINATED BY '%c' OPTIONALLY ENCLOSED BY '\"'", delimiter)
	}

	// Use the table name for the bad and log files
	badFileName := fmt.Sprintf("%s_bad.log", tableName)
	logFileName := fmt.Sprintf("%s.log", tableName)

	ctlContent := fmt.Sprintf(`OPTIONS (bad=%s, log=%s, errors=0, skip=1, ROWS=65535, BINDSIZE=65535000, READSIZE=65535000)
LOAD DATA
CHARACTERSET CL8MSWIN1251
%s
INTO TABLE %s
REPLACE
%s

TRAILING NULLCOLS
(
  %s
)
`, badFileName, logFileName, infile, tableName, delimiterStr, fieldsStr)

	err := os.WriteFile(ctlFilePath, []byte(ctlContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing .ctl file: %v", err)
	}

	return nil
}

func RunSQLLoader(user, password, dsn, ctlFilePath string) (string, error) {
	// Extract the table name from the control file path to use for the log and bad files
	tableName := strings.TrimSuffix(filepath.Base(ctlFilePath), ".ctl")
	badFileName := fmt.Sprintf("%s_bad.bad", tableName)
	logFileName := fmt.Sprintf("%s.log", tableName)

	cmd := exec.Command("sqlldr", fmt.Sprintf("userid=%s/%s@%s", user, password, dsn), fmt.Sprintf("control=%s", ctlFilePath), fmt.Sprintf("bad=%s", badFileName), fmt.Sprintf("log=%s", logFileName), "errors=10")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running sqlldr: %v, output: %s", err, output)
	}

	return string(output), nil
}
