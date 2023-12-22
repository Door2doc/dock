package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/door2doc/d2d-uploader/pkg/uploader/db"
	"github.com/door2doc/d2d-uploader/pkg/uploader/dlog"
	"github.com/door2doc/d2d-uploader/pkg/uploader/rest"
	"github.com/shibukawa/configdir"
)

var (
	Server     = "https://integration.door2doc.net/"
	configDirs = configdir.New("door2doc", "Upload Service")
)

const (
	PathPing             = "/services/v3/upload/ping"
	PathVisitorUpload    = "/services/v3/upload/bezoeken"
	PathRadiologieUpload = "/services/v3/upload/orders/radiologie"
	PathLabUpload        = "/services/v3/upload/orders/lab"
	PathConsultUpload    = "/services/v3/upload/orders/consult"
	DBValidationTimeout  = 5 * time.Second

	config = "door2doc.json"
)

type QueryResult interface {
	AsTable() template.HTML
}

// ValidationResult contains the results of validating the current configuration.
type ValidationResult struct {
	DatabaseConnection error
	QueryTimeout       error

	VisitorQuery         error
	VisitorQueryDuration time.Duration
	VisitorQueryResults  QueryResult

	RadiologieQuery         error
	RadiologieQueryDuration time.Duration
	RadiologieQueryResults  QueryResult

	LabQuery         error
	LabQueryDuration time.Duration
	LabQueryResults  QueryResult

	ConsultQuery         error
	ConsultQueryDuration time.Duration
	ConsultQueryResults  QueryResult

	D2DConnection  error
	D2DCredentials error

	Access error
}

// IsValid returns true if all fatal validation errors are nil.
func (v *ValidationResult) IsValid() bool {
	return v.DatabaseConnection == nil &&
		v.QueryTimeout == nil &&
		v.VisitorQuery == nil &&
		v.D2DConnection == nil &&
		v.D2DCredentials == nil
}

// Configuration contains the configuration options for the service.
type Configuration struct {
	mu sync.RWMutex

	// Set to true if the service should be active
	active bool

	data ConfigDataV1

	// results of the last call to UpdateValidation
	validationResult *ValidationResult
}

func NewConfiguration() *Configuration {
	return &Configuration{
		active: true,
		data: ConfigDataV1{
			Timeout: 5 * time.Second,
		},
	}
}

// Credentials returns the door2doc credentials stored in the configuration.
func (c *Configuration) Credentials() (string, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.data.Username, c.data.Password
}

func (c *Configuration) SetCredentials(username, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.Username = username
	c.data.Password = password
	dlog.SetUsername(username)
}

func (c *Configuration) Proxy() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data.Proxy
}

func (c *Configuration) SetProxy(proxy string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.Proxy = proxy
}

// Connection returns the connection data stored in the configuration.
func (c *Configuration) Connection() db.ConnectionData {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.data.Connection
}

func (c *Configuration) SetConnection(cd db.ConnectionData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.Connection = cd
}

// Timeout returns the timeout used for all queries
func (c *Configuration) Timeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Timeout
}

func (c *Configuration) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.Timeout = timeout
}

// VisitorQuery returns the visitor query stored in the configuration.
func (c *Configuration) VisitorQuery() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.VisitorQuery
}

func (c *Configuration) SetVisitorQuery(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.VisitorQuery = query
}

// RadiologieQuery returns the radiologie query stored in the configuration.
func (c *Configuration) RadiologieQuery() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.RadiologieQuery
}

func (c *Configuration) SetRadiologieQuery(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.RadiologieQuery = query
}

// LabQuery returns the lab query stored in the configuration.
func (c *Configuration) LabQuery() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.LabQuery
}

func (c *Configuration) SetLabQuery(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.LabQuery = query
}

// ConsultQuery returns the consult query stored in the configuration.
func (c *Configuration) ConsultQuery() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.ConsultQuery
}

func (c *Configuration) SetConsultQuery(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.ConsultQuery = query
}

func (c *Configuration) AccessCredentials() (username, password string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.AccessUsername, c.data.AccessPassword
}

func (c *Configuration) SetAccessCredentials(username, password string) {
	c.mu.Lock()
	c.mu.Unlock()

	if username == "" || password == "" {
		username = ""
		password = ""
	}
	c.data.AccessUsername = username
	c.data.AccessPassword = password
}

func (c *Configuration) Active() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.active
}

func (c *Configuration) Interval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return time.Minute
}

// UpdateBaseValidation validates the base configuration and returns the results of those checks.
func (c *Configuration) UpdateBaseValidation(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	res := &ValidationResult{}

	// check d2d connection
	connCtx, timeout := context.WithTimeout(ctx, DBValidationTimeout)
	defer timeout()

	res.D2DConnection, res.D2DCredentials = c.checkConnection(connCtx)

	if c.data.AccessUsername == "" && c.data.AccessPassword == "" {
		res.Access = ErrAccessNotConfigured
	}

	// check timeout
	if c.data.Timeout <= 0 {
		res.QueryTimeout = ErrInvalidTimeout
	}

	// check db connection
	res.VisitorQueryDuration, res.VisitorQueryResults, res.DatabaseConnection, res.VisitorQuery = c.checkDatabase(ctx, c.data.VisitorQuery, func(ctx context.Context, tx *sql.Tx, query string) (QueryResult, error) {
		return db.ExecuteVisitorQuery(ctx, tx, query, c.data.Timeout)
	})

	c.validationResult = res
	c.active = c.validationResult.IsValid()
}

// UpdateRadiologieValidation validates the order configuration and returns the results of those checks.
func (c *Configuration) UpdateRadiologieValidation(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.validationResult == nil {
		c.validationResult = new(ValidationResult)
	}
	res := c.validationResult

	// check db connection
	res.RadiologieQueryDuration, res.RadiologieQueryResults, res.DatabaseConnection, res.RadiologieQuery = c.checkDatabase(ctx, c.data.RadiologieQuery, func(ctx context.Context, tx *sql.Tx, query string) (QueryResult, error) {
		return db.ExecuteRadiologieQuery(ctx, tx, query, c.data.Timeout)
	})
}

// UpdateLabValidation validates the order configuration and returns the results of those checks.
func (c *Configuration) UpdateLabValidation(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.validationResult == nil {
		c.validationResult = new(ValidationResult)
	}
	res := c.validationResult

	// check db connection
	res.LabQueryDuration, res.LabQueryResults, res.DatabaseConnection, res.LabQuery = c.checkDatabase(ctx, c.data.LabQuery, func(ctx context.Context, tx *sql.Tx, query string) (QueryResult, error) {
		return db.ExecuteLabQuery(ctx, tx, query, c.data.Timeout)
	})
}

// UpdateConsultValidation validates the order configuration and returns the results of those checks.
func (c *Configuration) UpdateConsultValidation(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.validationResult == nil {
		c.validationResult = new(ValidationResult)
	}
	res := c.validationResult

	// check db connection
	res.ConsultQueryDuration, res.ConsultQueryResults, res.DatabaseConnection, res.ConsultQuery = c.checkDatabase(ctx, c.data.ConsultQuery, func(ctx context.Context, tx *sql.Tx, query string) (QueryResult, error) {
		return db.ExecuteConsultQuery(ctx, tx, query, c.data.Timeout)
	})
}

// Validate returns the result of the last validation.
func (c *Configuration) Validate() *ValidationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.validationResult
}

// Reload loads the configuration form a well-known location and updates the values accordingly.
func (c *Configuration) Reload() error {
	folders := configDirs.QueryFolders(configdir.System)
	if len(folders) == 0 {
		return nil
	}
	bs, err := folders[0].ReadFile(config)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := json.Unmarshal(bs, &c); err != nil {
		return err
	}
	return nil
}

// Save stores the latest configuration values to a well-known location.
func (c *Configuration) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bs, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	folders := configDirs.QueryFolders(configdir.System)
	if len(folders) == 0 {
		return errors.New("failed to find configuration folder")
	}
	if err := folders[0].WriteFile(config, bs); err != nil {
		return fmt.Errorf("while writing configuration file: %w", err)
	}
	dlog.Info("Updated %s/%s", folders[0].Path, config)
	return nil
}

func (c *Configuration) MarshalJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return json.Marshal(c.data)
}

func (c *Configuration) UnmarshalJSON(v []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := json.Unmarshal(v, &c.data); err != nil {
		return err
	}
	if c.data.Timeout == 0 {
		c.data.Timeout = 5 * time.Second
	}

	dlog.SetUsername(c.data.Username)

	return nil
}

func (c *Configuration) checkConnection(ctx context.Context) (connErr error, credErr error) {
	if c.data.Username == "" || c.data.Password == "" {
		credErr = ErrD2DCredentialsNotConfigured
	}

	req, err := http.NewRequest(http.MethodGet, Server, nil)
	if err != nil {
		dlog.Error("Failed to initialize connection to %s: %v", Server, err)
		return err, credErr
	}
	req.URL.Path = PathPing
	req.SetBasicAuth(c.data.Username, c.data.Password)

	res, err := rest.Do(ctx, c.data.Proxy, req)
	if err != nil {
		dlog.Error("Failed to connect to %s: %v", Server, err)
		return ErrD2DConnectionFailed, credErr
	}
	_, err = io.Copy(io.Discard, res.Body)
	if err != nil {
		dlog.Error("Failed to drain response: %v", err)
		return err, credErr
	}
	dlog.Close(res.Body)

	if credErr != nil {
		return nil, credErr
	}

	switch res.StatusCode {
	case http.StatusOK:
		dlog.Info("Ping successful")
		return nil, nil
	case http.StatusUnauthorized:
		dlog.Info("Ping not authorized")
		return nil, ErrD2DCredentialsInvalid
	default:
		dlog.Info("Ping failed")
		return nil, D2DCredentialsStatusError{StatusCode: res.StatusCode}
	}
}

type checker func(context.Context, *sql.Tx, string) (QueryResult, error)

func (c *Configuration) checkDatabase(ctx context.Context, query string, f checker) (queryDuration time.Duration, queryResult QueryResult, connErr, queryErr error) {
	if query == "" {
		queryErr = ErrQueryNotConfigured
	}

	if !c.data.Connection.IsValid() {
		connErr = ErrDatabaseNotConfigured
		return
	}

	conn, err := sql.Open(c.data.Connection.Driver, c.data.Connection.DSN())
	if err != nil {
		dlog.Error("Failed to connect to database: %v", err)
		connErr = &DatabaseInvalidError{Cause: err.Error()}
		return
	}

	err = conn.PingContext(ctx)
	if err != nil {
		dlog.Error("Failed to ping database %s: %v", c.data.Connection.Host, err)
		connErr = &DatabaseInvalidError{Cause: err.Error()}
		return
	}
	defer func() {
		dlog.Close(conn)
	}()

	if queryErr != nil {
		return
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		dlog.Error("Failed to start transaction: %v", err)
		connErr = &DatabaseInvalidError{Cause: err.Error()}
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			dlog.Error("Failed to roll back transaction: %v", err)
		}
	}()

	queryStart := time.Now()
	queryResult, err = f(ctx, tx, query)
	var selectionError *db.SelectionError
	errIsSelection := errors.As(err, &selectionError)

	switch {
	case err == nil:
	case errIsSelection:
		queryErr = err
		return
	case errors.Is(err, context.DeadlineExceeded):
		queryErr = err
		return
	default:
		queryErr = &QueryError{Cause: err.Error()}
		return
	}

	queryDuration = time.Since(queryStart)

	return
}

func (c *Configuration) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	req.SetBasicAuth(c.data.Username, c.data.Proxy)
	return rest.Do(ctx, c.data.Proxy, req)
}
