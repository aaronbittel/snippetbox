# Let's Go - Configuration and error handling

#### Advantages of command line flags vs environment variables
- you get a default value
- you get a `-help` command
- you get type conversion (bool, int, float, ...) (env are always strings)
- you can pass the environment variables into the command line flags
    - ```bash $ export SNIPPETBOX_ADDR=":9999" $ go run ./cmd/web
      -addr=$SNIPPETBOX_ADDR ```

## Structured Logging

### Creating a structured logger
- all structured logger have a **structured logging handler**, that controls
how log entries are formatted and where they are written to
- `logger := slog.New(slog.NewTextHandler(os.Stdout, nil))`
- there is no `log.Fatal()` => `logger.Error(err.Error()) && os.Exit(1)`

### Concurrent logging
- structured loggers are thread-safe and can be savely shared with multiple
goroutines

## Dependency injection
- all handlers are in the same package:
    - create a struct with all the shared dependencies (logger, db pool, ...)
    and register the handlers as methods on the struct
- handlers are in various packages:
    - create another package for the config and use a closure to create the
    handlers:
    - ```Go func ExampleHandler(app *config.Application) http.HandlerFunc {
      return func(w http.ResponseWriter, r *http.Request) { ... ts, err :=
          template.ParseFiles(files...) if err != nil {
          app.Logger.Error(err.Error(), "method", r.Method, "uri",
          r.URL.RequestURI()) http.Error(w, "Internal Server Error",
          http.StatusInternalServerError) return } ... } }
      ```

# Database-driven responses

## Setting up MySQL
- Installing MySQL: `sudo apt install mysql-server`

### Scaffolding the database
- login as root: `sudo mysql`
- create a new database: `CREATE DATABASE snippetbox CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;`
    - sets character set to utf8mb4 (enhanced version of utf-8 with wider
    support for characters and emojis)
    - `COLLATE utf8mb4_unicode_ci`: specifies how string comparions and
    sortings are performed, ci: case insensitiv => a == A `USE snippetbox;`
- Create table and index created column:
```SQL
CREATE TABLE snippets ( id
INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT, title VARCHAR(102) NOT NULL,
content TEXT NOT NULL, created DATETIME NOT NULL, expires DATETIME NOT NULL);

CREATE INDEX idx_snippets_created ON snippets(created);
```

### Create a new user
- create a new user with `SELECT`, `INSERT`, `UPDATE`, and `DELETE` privileges:
```SQL
CREATE USER 'web'@'localhost'; GRANT SELECT, INSERT, UPDATE, DELETE ON
snippetbox.* TO 'web'@'localhost'; ALTER USER 'web'@'localhost' IDENTIFIED BY
'password';
```

## Creating a database connection pool
- create a db connection pool: `sql.Open(driverName, dataSourceName string)
(*DB, error)`
- `driverName`: Name of the driver, e.g. "mysql"
- `dataSourceName`: depending on the db and driver in use (lookup in
documentation)
    - e.g. `web:pass@/snippetbox?parseTime=true`
        + web: username
        + pass: password
        + snippetbox: name of database
        + parseTime=true: convert SQL's `DATE`, `TIME` objects to Go's
        `time.Time`
- returns a database connection pool != database
    - automatically creates and closes database connection as needed
- normally is created once in the main function not in http handlers (too
expensive)
- use is thread-safe
- example: ```Go func openDB(dsn string) (*sql.DB, error) { db, err :=
sql.Open("mysql", dsn) if err != nil { return nil, err }

    err = db.Ping() if err != nil{ db.Close() return nil, err }

    return db, nil } ```
    - `sql.Open()` does not create any connections, creates them lazily when
    needed
    - `db.Ping()` to verify that everything is set up correctly

## Designing a database model
- create a package under `internal/models/snippets.go`
    - internal: non-application specific code
        + only parent directorys can import this code
        + model may be used in another project
- create a struct for the table entry
- create a SnippetModel struct that holds the db connection pool
- register methods like (Insert, Get, Latest) on SnippetModel
    - everything is encapsulated into one object
    - makes it mockable for testing
    - create one instance of SnippetModel and provide it to the handlers
- total control over which database to use at runtime

## Executing SQL statements
- example: ```SQL INSERT INTO snippets (title, content, created, expires)
VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY)) ```
- use placeholder parameters to prevent sql injection

### Executing the query
- three methods:
    - `DB.Query()`: used for `SELECT` queries which return multile rows
    - `DB.QueryRow()`: used for `SELECT` queries which return a single row
    - `DB.Exec()`: used for statements that don't return rows (`INSERT`,
    `DELETE`)

- Go Insert example: ```Go stmt := `INSERT INTO snippets (title, content,
created, expires) VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(),
INTERVAL ? DAY))`

result, err := DB.Exec(stmt, title, content, expries) id, err :=
result.LastInsertedId() ```
- `result` contains basic information about what happened when the statement
was executed
    - `LastInsertedId()`, `RowsAffected()`
    - Not all drivers support these methods, e.g. PostgreSQL does not support
    `LastInsertedId()`

## Single-record SQL queries
- `DB.QueryRow()` returns a pointer to a `sql.Row` which holds the result from
the database
- Use `row.Scan()` to copy the values from each field in `sql.Row` to the
corresponding field in the Snippet struct
    - arguments are pointers to the fields
    - if query returns no rows then `row.Scan()` returns `sql.ErrNoRows`
- create a own wrapper for the database (MySQL) specific errors to make the
handlers independent from the database in use
    - e.g. `var ErrNoRecord = errors.New("models: no matching record found")`

## Muliple-record SQL queries
- `rows, err := DB.Query()` returns a resultset containing the result of the
query
- `defer rows.Close()` to make sure the resultset is closed because as long as
the resultset is not closed a connection to the database is open
- `for rows.Next()` to loop over the resultset
- `rows.Scan()` on each row
- `rows.Err()` check for errors after the loop

## Transactions and other details

### Managing null values
- if a column in the table contains `NULL` value, `rows.Scan()` would error
because it cannot convert a `NULL` value to a string
    - fix: use `sql.NullString`
    - better: avoid `null` => set columns to `not null` or use sensible
    `default` values

### Working with transactions
- calls to `Exec()`, `Query()`, `QueryRow()` can use **any** db connection from
the pool
- no guarantee that two `Exec()` calls right after each other are using the
same connection
- after `LOCK TABLES`, `UNLOCK TABLES` must be called on the same connection to
prevent a deadlock => use transactions
    - `tx, err := DB.Begin()`: to start a transaction
    - `defer tx.Rollback()`
        + if transaction was successful: Changes are already commited; rollback
        is a no-op
        - if transaction was not successful: Reverts all the changes
    - `tx.Exec()`: do all the transactions
    - `tx.Commit()`: if there are no errors, commits the changes
- Either `tx.Rollback()` or `tx.Commit()` must be called before the function
returns, otherwise the connection will not be closed

### Prepared statements
- `Exec()`, `Query()` and `QueryRow()` all use prepared statements behind the
scenes to help prevent SQL injection attacks
    - setup prepared statements
    - run them with the paramters provided
        + parameters are treated as data not executable code
    - close the prepared statements
- `DB.Prepare()`: creates a own prepared statements once
    - good for
        + complex statements like multiple `JOINS`
        + bulk insert opertations
- example: ```Go type ExampleModel struct { DB         *sql.DB InsertStmt
*sql.Stmt }

func NewExampleModel(db *sql.DB) (*ExampleModel, error) { insertStmt, err :=
db.Prepare("INSERT INTO ...") if err != nil { return nil, err } return
&ExampleModel{ DB: db, InsertStmt: insertStmt, }, nil }

func (m *ExampleModel) Insert(args ...) error { _, err :=
m.InsertStmt.Exec(args...) return err }

func main() { ... defer exampleModel.InsertStmt.Close() } ```
- need to defer close the InsertStmt after the program finishes
- prepared statements exist on a database connection
    - the statements remembers which connection it used and tries to use that
    again the next time
    - if the connection is not availabe it is re-prepared on another connection
- on heavy load it could be a lot of effort preparing and re-preparing
statements
    - server-side limits on the number of statements (MySQL: 16382)
