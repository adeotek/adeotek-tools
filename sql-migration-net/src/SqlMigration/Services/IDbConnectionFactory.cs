using System.Data;
using SqlMigration.Models;

namespace SqlMigration.Services;

public interface IDbConnectionFactory
{
    DatabaseProvider Provider { get; }
    string ConnectionString { get; }
    IDbConnection Create();
}
