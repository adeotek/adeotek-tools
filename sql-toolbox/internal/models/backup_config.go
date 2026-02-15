package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
)

// BackupConfig represents the top-level YAML configuration for backup-only command
type BackupConfig struct {
	Defaults  BackupDefaults   `yaml:"defaults"`
	Databases []DatabaseTarget `yaml:"databases"`
	S3        S3Config         `yaml:"s3"`
}

// BackupDefaults holds the default settings applied to all databases
type BackupDefaults struct {
	OutputDir    string           `yaml:"output_dir"`
	Compress     *bool            `yaml:"compress"`
	UploadToS3   *bool            `yaml:"upload_to_s3"`
	NoOwner      *bool            `yaml:"no_owner"`
	Clean        *bool            `yaml:"clean"`
	Schemas      []string         `yaml:"schemas"`
	BackupMethod *string          `yaml:"backup_method"`
	SSHTunnel    *SSHTunnelConfig `yaml:"ssh_tunnel"`
}

// DatabaseTarget represents a single database to back up
type DatabaseTarget struct {
	Name                    string           `yaml:"name"`
	Host                    string           `yaml:"host"`
	Port                    int              `yaml:"port"`
	Database                string           `yaml:"database"`
	User                    string           `yaml:"user"`
	Password                string           `yaml:"password"`
	SSLMode                 string           `yaml:"ssl_mode"`
	ConnectionString        string           `yaml:"connection_string"`
	OutputDir               string           `yaml:"output_dir"`
	Compress                *bool            `yaml:"compress"`
	UploadToS3              *bool            `yaml:"upload_to_s3"`
	DeleteLocalAfterUpload  *bool            `yaml:"delete_local_after_upload"`
	Schemas                 []string         `yaml:"schemas"`
	ExcludeTables           []string         `yaml:"exclude_tables"`
	ExcludeDb               []string         `yaml:"exclude_db"`
	NoOwner                 *bool            `yaml:"no_owner"`
	Clean                   *bool            `yaml:"clean"`
	BackupMethod            *string          `yaml:"backup_method"`
	SSHTunnel               *SSHTunnelConfig `yaml:"ssh_tunnel"`
	S3Prefix                string           `yaml:"s3_prefix"`
	S3Bucket                string           `yaml:"s3_bucket"`
	S3AccessKeyID           string           `yaml:"s3_access_key_id"`
	S3SecretAccessKey       string           `yaml:"s3_secret_access_key"`

	// Reference to parent config defaults and S3 config for effective value resolution
	defaults *BackupDefaults
	s3Config *S3Config
}

// S3Config holds S3/S3-compatible storage configuration
type S3Config struct {
	Enabled                bool   `yaml:"enabled"`
	Bucket                 string `yaml:"bucket"`
	Region                 string `yaml:"region"`
	Prefix                 string `yaml:"prefix"`
	AccessKeyID            string `yaml:"access_key_id"`
	SecretAccessKey         string `yaml:"secret_access_key"`
	Endpoint               string `yaml:"endpoint"`
	DeleteLocalAfterUpload *bool  `yaml:"delete_local_after_upload"`
}

// LoadBackupConfig parses and validates a backup configuration YAML file
func LoadBackupConfig(path string) (*BackupConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config BackupConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	// Link defaults and S3 config to each database target
	for i := range config.Databases {
		config.Databases[i].SetDefaults(&config.Defaults, &config.S3)
	}

	return &config, nil
}

func (c *BackupConfig) validate() error {
	if len(c.Databases) == 0 {
		return fmt.Errorf("at least one database must be configured")
	}

	for i, db := range c.Databases {
		if db.Name == "" {
			return fmt.Errorf("database at index %d: name is required", i)
		}
		if db.ConnectionString == "" && db.Host == "" {
			return fmt.Errorf("database '%s': either connection_string or host must be provided", db.Name)
		}
		// Allow "*" as wildcard, otherwise database name is required
		if db.ConnectionString == "" && db.Database == "" {
			return fmt.Errorf("database '%s': database name is required (use '*' for all databases)", db.Name)
		}
	}

	if c.S3.Enabled && c.S3.Bucket == "" {
		return fmt.Errorf("s3: bucket is required when S3 is enabled")
	}

	return nil
}

// SetDefaults links the default and S3 configuration references for effective value resolution
func (dt *DatabaseTarget) SetDefaults(defaults *BackupDefaults, s3Config *S3Config) {
	dt.defaults = defaults
	dt.s3Config = s3Config
}

// GetEffectiveOutputDir returns the output directory for this database, falling back to defaults
func (dt *DatabaseTarget) GetEffectiveOutputDir() string {
	if dt.OutputDir != "" {
		return dt.OutputDir
	}
	if dt.defaults != nil && dt.defaults.OutputDir != "" {
		return dt.defaults.OutputDir
	}
	return "./backups"
}

// GetEffectiveCompress returns whether compression is enabled for this database
func (dt *DatabaseTarget) GetEffectiveCompress() bool {
	if dt.Compress != nil {
		return *dt.Compress
	}
	if dt.defaults != nil && dt.defaults.Compress != nil {
		return *dt.defaults.Compress
	}
	return true // default: compress enabled
}

// GetEffectiveUploadToS3 returns whether S3 upload is enabled for this database
func (dt *DatabaseTarget) GetEffectiveUploadToS3() bool {
	if dt.UploadToS3 != nil {
		return *dt.UploadToS3
	}
	if dt.defaults != nil && dt.defaults.UploadToS3 != nil {
		return *dt.defaults.UploadToS3
	}
	return false
}

// GetEffectiveDeleteLocalAfterUpload returns whether to delete local file after S3 upload
func (dt *DatabaseTarget) GetEffectiveDeleteLocalAfterUpload() bool {
	if dt.DeleteLocalAfterUpload != nil {
		return *dt.DeleteLocalAfterUpload
	}
	if dt.s3Config != nil && dt.s3Config.DeleteLocalAfterUpload != nil {
		return *dt.s3Config.DeleteLocalAfterUpload
	}
	return false
}

// GetEffectiveNoOwner returns whether to exclude ownership statements
func (dt *DatabaseTarget) GetEffectiveNoOwner() bool {
	if dt.NoOwner != nil {
		return *dt.NoOwner
	}
	if dt.defaults != nil && dt.defaults.NoOwner != nil {
		return *dt.defaults.NoOwner
	}
	return true
}

// GetEffectiveClean returns whether to include DROP IF EXISTS statements
func (dt *DatabaseTarget) GetEffectiveClean() bool {
	if dt.Clean != nil {
		return *dt.Clean
	}
	if dt.defaults != nil && dt.defaults.Clean != nil {
		return *dt.defaults.Clean
	}
	return true
}

// GetEffectiveBackupMethod returns the backup method for this database
func (dt *DatabaseTarget) GetEffectiveBackupMethod() string {
	if dt.BackupMethod != nil {
		return *dt.BackupMethod
	}
	if dt.defaults != nil && dt.defaults.BackupMethod != nil {
		return *dt.defaults.BackupMethod
	}
	return "go"
}

// GetEffectiveSchemas returns the schemas to back up, nil/empty means all non-system schemas
func (dt *DatabaseTarget) GetEffectiveSchemas() []string {
	if len(dt.Schemas) > 0 {
		return dt.Schemas
	}
	if dt.defaults != nil && len(dt.defaults.Schemas) > 0 {
		return dt.defaults.Schemas
	}
	return nil
}

// GetEffectiveSSHTunnel returns the SSH tunnel config for this database, falling back to defaults
func (dt *DatabaseTarget) GetEffectiveSSHTunnel() *SSHTunnelConfig {
	if dt.SSHTunnel != nil {
		return dt.SSHTunnel
	}
	if dt.defaults != nil && dt.defaults.SSHTunnel != nil {
		return dt.defaults.SSHTunnel
	}
	return nil
}

// GetEffectiveS3Prefix returns the S3 prefix for this database, falling back to global S3 config
func (dt *DatabaseTarget) GetEffectiveS3Prefix() string {
	if dt.S3Prefix != "" {
		return dt.S3Prefix
	}
	if dt.s3Config != nil {
		return dt.s3Config.Prefix
	}
	return ""
}

// GetEffectiveS3Bucket returns the S3 bucket for this database, falling back to global S3 config
func (dt *DatabaseTarget) GetEffectiveS3Bucket() string {
	if dt.S3Bucket != "" {
		return dt.S3Bucket
	}
	if dt.s3Config != nil {
		return dt.s3Config.Bucket
	}
	return ""
}

// GetEffectiveS3AccessKeyID returns the S3 access key ID for this database, falling back to global S3 config
func (dt *DatabaseTarget) GetEffectiveS3AccessKeyID() string {
	if dt.S3AccessKeyID != "" {
		return dt.S3AccessKeyID
	}
	if dt.s3Config != nil {
		return dt.s3Config.AccessKeyID
	}
	return ""
}

// GetEffectiveS3SecretAccessKey returns the S3 secret access key for this database, falling back to global S3 config
func (dt *DatabaseTarget) GetEffectiveS3SecretAccessKey() string {
	if dt.S3SecretAccessKey != "" {
		return dt.S3SecretAccessKey
	}
	if dt.s3Config != nil {
		return dt.s3Config.SecretAccessKey
	}
	return ""
}

// ToConnectionString builds a lib/pq connection string from individual parameters
func (dt *DatabaseTarget) ToConnectionString() string {
	if dt.ConnectionString != "" {
		// Normalize connection string to lib/pq format
		normalized, err := normalizeConnectionString(dt.ConnectionString)
		if err != nil {
			// If normalization fails, return original string
			// (error will be caught during actual database connection)
			return dt.ConnectionString
		}
		return normalized
	}

	parts := []string{
		fmt.Sprintf("host=%s", dt.Host),
	}

	port := dt.Port
	if port == 0 {
		port = 5432
	}
	parts = append(parts, fmt.Sprintf("port=%d", port))

	parts = append(parts, fmt.Sprintf("dbname=%s", dt.Database))

	if dt.User != "" {
		parts = append(parts, fmt.Sprintf("user=%s", dt.User))
	}
	if dt.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", dt.Password))
	}

	sslMode := dt.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	parts = append(parts, fmt.Sprintf("sslmode=%s", sslMode))

	return strings.Join(parts, " ")
}

// QueryDatabases connects to PostgreSQL and retrieves list of all databases
// excluding templates and the postgres database
func QueryDatabases(target *DatabaseTarget, baseExclusions []string) ([]string, error) {
	// Build connection string to postgres database (for querying system tables)
	connTarget := *target
	connTarget.Database = "postgres" // Connect to postgres database to query pg_database

	connStr := connTarget.ToConnectionString()

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	// Build exclusion list
	exclusions := []string{"postgres"}
	exclusions = append(exclusions, baseExclusions...)
	exclusions = append(exclusions, target.ExcludeDb...)

	// Build query with exclusions
	placeholders := make([]interface{}, len(exclusions))
	placeholderStrings := make([]string, len(exclusions))
	for i, excl := range exclusions {
		placeholders[i] = excl
		placeholderStrings[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT datname
		FROM pg_database
		WHERE datistemplate = false
		AND datname NOT IN (%s)
		ORDER BY datname`,
		strings.Join(placeholderStrings, ", "))

	rows, err := db.Query(query, placeholders...)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %w", err)
		}
		databases = append(databases, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating database rows: %w", err)
	}

	return databases, nil
}

// ExpandWildcardDatabases takes a DatabaseTarget with database: "*" and expands it
// into multiple DatabaseTarget entries, one per discovered database
func ExpandWildcardDatabases(template DatabaseTarget, verbose bool) ([]DatabaseTarget, error) {
	if template.Database != "*" {
		return []DatabaseTarget{template}, nil // Not a wildcard, return as-is
	}

	if verbose {
		log.Printf("Expanding wildcard for '%s'...", template.Name)
	}

	// Query databases using QueryDatabases
	databases, err := QueryDatabases(&template, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate databases for wildcard: %w", err)
	}

	exclusions := append([]string{"postgres"}, template.ExcludeDb...)

	if verbose {
		log.Printf("  Discovered %d database(s) after exclusions %v", len(databases), exclusions)
		if len(databases) > 0 {
			log.Printf("  Will backup: %v", databases)
		}
	}

	if len(databases) == 0 {
		return nil, fmt.Errorf("wildcard expansion found 0 databases after exclusions %v", exclusions)
	}

	// Create new DatabaseTarget for each discovered database
	var expanded []DatabaseTarget
	for _, dbName := range databases {
		// Clone template target
		target := template
		target.Database = dbName
		target.Name = template.Name + "_" + dbName // Modify name to include database
		target.ExcludeDb = nil // Clear exclude_db from expanded entries
		expanded = append(expanded, target)
	}

	return expanded, nil
}

// ExpandWildcards processes all DatabaseTarget entries and expands any with database: "*"
func (c *BackupConfig) ExpandWildcards(verbose bool) error {
	var expandedDatabases []DatabaseTarget

	for _, dbTarget := range c.Databases {
		expanded, err := ExpandWildcardDatabases(dbTarget, verbose)
		if err != nil {
			return fmt.Errorf("failed to expand wildcard for '%s': %w", dbTarget.Name, err)
		}
		expandedDatabases = append(expandedDatabases, expanded...)
	}

	c.Databases = expandedDatabases

	// Re-link defaults after expansion
	for i := range c.Databases {
		c.Databases[i].SetDefaults(&c.Defaults, &c.S3)
	}

	return nil
}
