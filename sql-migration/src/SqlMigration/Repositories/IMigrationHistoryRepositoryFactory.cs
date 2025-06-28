namespace SqlMigration.Repositories;

public interface IMigrationHistoryRepositoryFactory
{
    IMigrationHistoryRepository Create(string connectionString);
}
