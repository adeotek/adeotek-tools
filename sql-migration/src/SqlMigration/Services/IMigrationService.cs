namespace SqlMigration.Services;

public interface IMigrationService
{
    Task<int> RunAsync(string scriptsPath, string connectionString);
}
