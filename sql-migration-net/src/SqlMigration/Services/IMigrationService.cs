using SqlMigration.Models;

namespace SqlMigration.Services;

public interface IMigrationService
{
    Task<int> RunAsync(string scriptsPath, ConnectionParameters connectionParameters, CancellationToken ct = default);
}
