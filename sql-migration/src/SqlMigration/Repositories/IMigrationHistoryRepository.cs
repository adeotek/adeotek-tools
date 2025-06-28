using SqlMigration.Models;

namespace SqlMigration.Repositories;

public interface IMigrationHistoryRepository
{
    Task<bool> IsHistoryTableCreated();
    Task CreateHistoryTable();
    Task<IEnumerable<MigrationHistory>> GetExecutedScripts();
    Task AddExecutedScript(MigrationHistory migrationHistory);
}
