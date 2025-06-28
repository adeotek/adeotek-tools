using SqlMigration.Repositories;

namespace SqlMigration.Factories;

public class MigrationHistoryRepositoryFactory : IMigrationHistoryRepositoryFactory
{
    public IMigrationHistoryRepository Create(string connectionString)
    {
        return new MigrationHistoryRepository(connectionString);
    }
}
