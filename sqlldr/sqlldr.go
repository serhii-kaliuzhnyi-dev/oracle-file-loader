package sqlldr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/serhii-kaliuzhnyi-dev/oracle-file-uploader/db"
)

func GenerateCtlFile(csvFilePath, ctlFilePath, tableName string, tableConfig db.TableConfig, delimiter rune) error {
	var fields []string
	for _, colName := range tableConfig.ColumnsOrder {
		fields = append(fields, colName)
	}
	fieldsStr := strings.Join(fields, ",\n  ")

	var delimiterStr string
	if delimiter == '\t' {
		delimiterStr = "FIELDS TERMINATED BY X'09' OPTIONALLY ENCLOSED BY '\"'"
	} else {
		delimiterStr = fmt.Sprintf("FIELDS TERMINATED BY '%c' OPTIONALLY ENCLOSED BY '\"'", delimiter)
	}

	ctlContent := fmt.Sprintf(`OPTIONS (bad=ksg_bad.log, log=ksg.log, errors=0, skip=1, ROWS=65535, BINDSIZE=65535000, READSIZE=65535000)
LOAD DATA
CHARACTERSET CL8MSWIN1251
INFILE '%s'
INTO TABLE %s
REPLACE
%s

TRAILING NULLCOLS
(
  %s
)
`, csvFilePath, tableName, delimiterStr, fieldsStr)

	err := os.WriteFile(ctlFilePath, []byte(ctlContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing .ctl file: %v", err)
	}

	return nil
}

func RunSQLLoader(user, password, dsn, ctlFilePath string) (string, error) {
	cmd := exec.Command("sqlldr", fmt.Sprintf("userid=%s/%s@%s", user, password, dsn), fmt.Sprintf("control=%s", ctlFilePath), "bad=uz_bad.bad", "log=uz.log", "errors=10")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running sqlldr: %v, output: %s", err, output)
	}

	return string(output), nil
}
