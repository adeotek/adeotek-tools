namespace SqlMigration.Models;

public static class Constants {
    public const string EnvironmentVariablesPrefix = "SQL_MIGRATION_";
    public const string ConnectionStringArgumentName = "connection-string";
    public const string ScriptsPathArgumentName = "target-path";
    public const string MigrationHistoryTableName = "_migration_history";
}
