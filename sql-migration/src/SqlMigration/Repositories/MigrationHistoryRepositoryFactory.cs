using SqlMigration.Models;
using SqlMigration.Services;

namespace SqlMigration.Repositories;

public class MigrationHistoryRepositoryFactory : IMigrationHistoryRepositoryFactory
{
    public IMigrationHistoryRepository Create(ConnectionParameters connectionParameters)
    {
        return new MigrationHistoryRepository(new SqlRepository(new DbConnectionFactory(connectionParameters)));
    }
}
