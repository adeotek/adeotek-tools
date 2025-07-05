using System.Data;
using System.Data.SQLite;
using Npgsql;
using SqlMigration.Models;

namespace SqlMigration.Services;

public class DbConnectionFactory(ConnectionParameters connectionParameters)
    : IDbConnectionFactory
{
    public DatabaseProvider Provider => connectionParameters.DbProvider;
    public string ConnectionString => connectionParameters.GetConnectionString();

    public IDbConnection Create()
    {
        ArgumentNullException.ThrowIfNull(connectionParameters);
        if (!connectionParameters.IsValid(out var errors))
        {
            throw new Exception($"Invalid ConnectionParameters: {string.Join("; ", errors)}");
        }
        return Create(connectionParameters.DbProvider, connectionParameters.GetConnectionString());
    }

    public static IDbConnection Create(ConnectionParameters connectionParameters)
    {
        ArgumentNullException.ThrowIfNull(connectionParameters);
        if (!connectionParameters.IsValid(out var errors))
        {
            throw new Exception($"Invalid ConnectionParameters: {string.Join("; ", errors)}");
        }
        return Create(connectionParameters.DbProvider, connectionParameters.GetConnectionString());
    }

    public static IDbConnection Create(DatabaseProvider provider, string connectionString)
    {
        return provider switch
        {
            DatabaseProvider.PostgreSql => new NpgsqlConnection(connectionString),
            DatabaseProvider.SqLite => new SQLiteConnection(connectionString),
            DatabaseProvider.Unknown => throw new NotSupportedException($"Database provider '{provider}' is not supported."),
            _ => throw new ArgumentOutOfRangeException(nameof(provider), provider, "Unsupported database provider")
        };
    }
}
