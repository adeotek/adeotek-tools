using SqlMigration.Models;

namespace SqlMigration.Repositories;

public interface IMigrationHistoryRepository
{
    Task<bool> IsHistoryTableCreatedAsync(CancellationToken ct = default);
    Task CreateHistoryTableAsync();
    Task<IEnumerable<ScriptExecutionHistory>> GetExecutedScriptsAsync(CancellationToken ct = default);
    Task UpsertExecutedScriptAsync(ScriptExecutionHistory scriptExecutionHistory);
}
