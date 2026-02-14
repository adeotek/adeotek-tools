package models

import (
	"fmt"
	"os"
	"strings"

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
	NoOwner                 *bool            `yaml:"no_owner"`
	Clean                   *bool            `yaml:"clean"`
	BackupMethod            *string          `yaml:"backup_method"`
	SSHTunnel               *SSHTunnelConfig `yaml:"ssh_tunnel"`
	S3Prefix                string           `yaml:"s3_prefix"`

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
		if db.ConnectionString == "" && db.Database == "" {
			return fmt.Errorf("database '%s': database name is required when not using connection_string", db.Name)
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
