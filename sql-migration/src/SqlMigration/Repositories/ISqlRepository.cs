using System.Data;
using SqlMigration.Services;

namespace SqlMigration.Repositories;

public interface ISqlRepository
{
    IDbConnectionFactory DbConnectionFactory { get; }
    Task<int> ExecuteAsync(string sql, object? parameters = null, CommandType? commandType = null);
    Task<T?> ExecuteScalarAsync<T>(string sql, object? parameters = null, CommandType? commandType = null);
    Task<IEnumerable<T>> QueryAsync<T>(string sql, object? parameters = null, CommandType? commandType = null);
    Task<T?> QueryFirstOrDefaultAsync<T>(string sql, object? parameters = null, CommandType? commandType = null);
}
