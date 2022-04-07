# earth-rest-api

## Prerequisite:
- Install PostgreSQL and Go Environment.
- Create your local database name, database username & database password. You should create two database for normal usage and testing purpose.
- Create an `env` file & store your database configuration into the `env` file at the root directory of the project. the content is following:
```
export APP_USERNAME=
export APP_PASSWORD=
export APP_DATABASE=

export TEST_USERNAME=
export TEST_PASSWORD=
export TEST_DATABASE=
```
- Unfortunaly, application is built without docker-compose for now.

## Running application:

```
# server
cd server
go build
./server

# test
# 1. booting server in test mode:
cd server
go build
./server --test-build

# 2. running test:
cd test
npm install
npm run test

# or in vscode
run vscode tasks by `Ctrl + Shift + P` (On window) or `Cmd + Shift + P` (On mac) to "run test" and "run server"
```
## Server project structure:

```
├── config.go
├── database.go
├── database_name.go
├── http_handlers_name.go
├── http_server.go
├── main.go
├── main_test.go
├── pkg
│   ├── name.go
│   ├── mhttp/mhttp.go
│   ├── msql/msql.go
│   ├── mstring/mstring.go
│   └── muuid/muuid.go
├── server
└── sql
    ├── 01-create-table.sql
    ├── 02-trigger-function.sql
    └── 03-create-trigger.sql
```

### 1. Project base:

- `main.go`: Handling main application operation, for example: server boot, initialized database from `database.go` and server http from `http_server.go`

- `database.go`: responsible for main database operation, such as running initial SQL migration in `sever/sql/0d-operation.sql` in numericial order. File also contains basic database function.

- `http_server.go`: manage route and server API at configured port.

- `config.go`: handle configuration for application framework and database settings.

- `/server/sql/01-create-table.sql`: contains basic table schema.

- `/server/pkg/mutil/mutil.go`: contains go utils package related to SQL, string modification, http and uuid.


### 2. Directory & data structure for feature:

When structure file & schema for a new feature, for example continent, the data would follow the belowing generic rules:



- base SQL structure:
```sql
CREATE TABLE IF NOT EXISTS country (
    -- id
    index bigserial PRIMARY KEY,
    uuid uuid NOT NULL UNIQUE,
    -- base
    created timestamp DEFAULT NOW(),
    updated timestamp,
    creator jsonb,
    deleted_state smallint default 0
);
```

the basic schema is stored in `01-create-table.sql` for schema to be created on server boot.


- base go structure:
```go
type Data struct {
    // id
	Index msql.DatabaseIndex `json:"-"`
	Uuid  muuid.UUID         `json:"uuid"`

    // data
	Name      string        `json:"name"`
	Type      DataType `json:"type"`

    // base
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Creator *UserMinimal `json:"creator"`
	DeletedState msql.DeletedState `json:"-"`
}
```

These data structure is created as own `data.go` file name inside `server/pkg`. each file contains the go struct of the data & related function for the data struct.

- http & database function: each feature will have their own `http_handler_<data>.go` to handle API functionality & `database_<data>.go` to handle related database functionality.

## Test project structure:

```
├── jest-earth-rest-api-preset.js
├── jest-puppeteer.config.js
├── jest-test-sequencer.js
├── jest.config.js
├── package-lock.json
├── package.json
├── run
│   ├── 01-init-data.spec.ts
│   ├── src
│   │   ├── api.ts
│   │   └── testing.ts
│   └── types
│       └── earth.ts
├── test-report.html
└── tsconfig.json
```