using System.Data;
using System.Diagnostics.CodeAnalysis;
using Dapper;
using SqlMigration.Services;

namespace SqlMigration.Repositories;

[ExcludeFromCodeCoverage]
public class SqlRepository(IDbConnectionFactory dbConnectionFactory)
    : ISqlRepository
{
    public IDbConnectionFactory DbConnectionFactory => dbConnectionFactory;

    public async Task<int> ExecuteAsync(string sql, object? parameters = null, CommandType? commandType = null)
    {
        using var connection = dbConnectionFactory.Create();
        return await connection.ExecuteAsync(sql, parameters, commandType: commandType);
    }

    public async Task<T?> ExecuteScalarAsync<T>(string sql, object? parameters = null, CommandType? commandType = null)
    {
        using var connection = dbConnectionFactory.Create();
        return await connection.ExecuteScalarAsync<T>(sql, parameters, commandType: commandType);
    }

    public async Task<IEnumerable<T>> QueryAsync<T>(string sql, object? parameters = null, CommandType? commandType = null)
    {
        using var connection = dbConnectionFactory.Create();
        return await connection.QueryAsync<T>(sql, parameters, commandType: commandType);
    }

    public async Task<T?> QueryFirstOrDefaultAsync<T>(string sql, object? parameters = null, CommandType? commandType = null)
    {
        using var connection = dbConnectionFactory.Create();
        return await connection.QueryFirstOrDefaultAsync<T>(sql, parameters, commandType: commandType);
    }
}
