using System.Data;
using Dapper;
using Microsoft.Data.SqlClient;
using SqlMigration.Models;

namespace SqlMigration.Repositories;

public class MigrationHistoryRepository(ConnectionParameters connectionParameters)
    : IMigrationHistoryRepository
{
    private const string TableName = "__migrations_history";

    private readonly DatabaseProvider _dbProvider = connectionParameters.DbProvider;
    private readonly string _connectionString = connectionParameters.GetConnectionString();

    public async Task<bool> IsHistoryTableCreatedAsync(CancellationToken ct = default)
    {
        using IDbConnection db = new SqlConnection(_connectionString);
        const string sql = $"""
                           IF OBJECT_ID('{TableName}', N'U') IS NOT NULL
                               SELECT 1
                           ELSE
                               SELECT 0
                           """;
        return await db.ExecuteScalarAsync<bool>(sql);
    }

    public async Task CreateHistoryTableAsync()
    {
        using IDbConnection db = new SqlConnection(_connectionString);
        const string sql = @"
            CREATE TABLE dbo.__SchemaMigrations (
                Id INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
                ScriptName NVARCHAR(255) NOT NULL,
                Hash NVARCHAR(255) NOT NULL,
                ExecutedAt DATETIME NOT NULL
            )";
        await db.ExecuteAsync(sql);
    }

    public async Task<IEnumerable<ScriptExecutionHistory>> GetExecutedScriptsAsync(CancellationToken ct = default)
    {
        using IDbConnection db = new SqlConnection(_connectionString);
        const string sql = "SELECT * FROM dbo.__SchemaMigrations";
        return await db.QueryAsync<ScriptExecutionHistory>(sql);
    }

    public async Task UpsertExecutedScriptAsync(ScriptExecutionHistory scriptExecutionHistory)
    {
        using IDbConnection db = new SqlConnection(_connectionString);
        const string sql = @"
            INSERT INTO dbo.__SchemaMigrations (ScriptName, Hash, ExecutedAt)
            VALUES (@ScriptName, @Hash, @ExecutedAt)";
        await db.ExecuteAsync(sql, scriptExecutionHistory);
    }
}
