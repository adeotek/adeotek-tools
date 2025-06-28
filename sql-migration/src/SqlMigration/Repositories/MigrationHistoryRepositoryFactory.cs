namespace SqlMigration.Repositories;

public class MigrationHistoryRepositoryFactory : IMigrationHistoryRepositoryFactory
{
    public IMigrationHistoryRepository Create(string connectionString)
    {
        return new MigrationHistoryRepository(connectionString);
    }
}
