using SqlMigration.Repositories;

namespace SqlMigration.Factories;

public interface IMigrationHistoryRepositoryFactory
{
    IMigrationHistoryRepository Create(string connectionString);
}
