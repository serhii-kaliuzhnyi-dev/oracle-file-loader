# Oracle File Uploader

Oracle File Uploader is a command-line tool for uploading CSV files to an Oracle database. It automates the process of generating a configuration file, determining column lengths, and using SQL*Loader to load data into the database.

## Features

- **Plan**: Generates a configuration file based on the CSV file. This includes determining column names, types, and lengths, and allows you to choose which columns should be included in the final table.
- **Apply**: Creates a table in the Oracle database based on the generated configuration file and uses the SQL*Loader configuration file (`.ctl`) to load data into the database.

## Installation

1. **Prerequisites**:
   - Oracle Instant Client installed.
   - Go installed on your system.

2. **Build the Application**:
   ```bash
   GOOS=windows GOARCH=amd64 go build -o loader.exe
   ```
   
## Usage

### Step 1: Prepare Your Environment

Ensure that your `.env` file contains the necessary environment variables for connecting to your Oracle database. Here’s an example of what the `.env` file might look like:

```env
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_URL=your_db_url
FILE_PATH=/path/to/your/csv/file.csv
TABLE_NAME=your_table_name
CTL_FILE_PATH=/path/to/save/your.ctl
```

### Step 2: Run the `plan` Command

The `plan` command generates a configuration file based on the provided CSV file. This configuration file identifies the columns to be included in the final table and determines the appropriate data types and lengths.

```cmd
./loader.exe plan
```

This will:

- Convert the CSV file to UTF-8 (if necessary) and then back to Windows-1251 for processing.
- Generate a configuration file (your_table_name.config.json) that defines the columns and their properties.
- Generate a .ctl file for SQL*Loader based on the processed data.

### Step 3: Review and Edit the Configuration File

After running the `plan` command, a configuration file (`your_table_name.config.json`) will be generated. You can review this file and make any necessary adjustments to the columns (e.g., changing the `create` flag to `false` for any columns you don't want to include in the final table).

### Step 4: Run the `apply` Command

Once you are satisfied with the configuration, you can run the `apply` command to create the table in the Oracle database and load the data using SQL*Loader.

```cmd
./loader.exe apply --auto-approve
```
This will:
- Create the table in the Oracle database based on the configuration file.
- Load the data into the table using the SQL*Loader configuration generated during the plan step.

### Optional Flags

- `--auto-approve`: Automatically approve the table creation without prompting for confirmation.
- `--skip-table`: Skip the table creation step and only run SQL*Loader.

### Cleanup

After running the `apply` command, the tool will remove any temporary files used during the process, such as the UTF-8 converted file.

## Troubleshooting

- Ensure that all required environment variables are correctly set in the `.env` file.
- Make sure that Oracle Instant Client is installed and accessible by your Go application.

## Contributing

Feel free to fork this repository and make contributions. Pull requests are welcome.

## License

This project is licensed under the MIT License.
