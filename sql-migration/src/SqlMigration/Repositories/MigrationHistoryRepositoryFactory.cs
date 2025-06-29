using SqlMigration.Models;

namespace SqlMigration.Repositories;

public class MigrationHistoryRepositoryFactory : IMigrationHistoryRepositoryFactory
{
    public IMigrationHistoryRepository Create(ConnectionParameters connectionParameters)
    {
        return new MigrationHistoryRepository(connectionParameters);
    }
}
