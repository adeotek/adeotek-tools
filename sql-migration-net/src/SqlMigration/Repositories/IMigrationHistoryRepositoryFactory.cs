using SqlMigration.Models;

namespace SqlMigration.Repositories;

public interface IMigrationHistoryRepositoryFactory
{
    IMigrationHistoryRepository Create(ConnectionParameters connectionParameters);
}
