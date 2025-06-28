using System.Data;
using Dapper;
using Microsoft.Data.SqlClient;
using SqlMigration.Models;

namespace SqlMigration.Repositories;

public class MigrationHistoryRepository : IMigrationHistoryRepository
{
    private readonly string _connectionString;

    public MigrationHistoryRepository(string connectionString)
    {
        _connectionString = connectionString;
    }

    public async Task<bool> IsHistoryTableCreated()
    {
        using IDbConnection db = new SqlConnection(_connectionString);
        const string sql = @"
            IF OBJECT_ID(N'dbo.__SchemaMigrations', N'U') IS NOT NULL
                SELECT 1
            ELSE
                SELECT 0";
        return await db.ExecuteScalarAsync<bool>(sql);
    }

    public async Task CreateHistoryTable()
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

    public async Task<IEnumerable<MigrationHistory>> GetExecutedScripts()
    {
        using IDbConnection db = new SqlConnection(_connectionString);
        const string sql = "SELECT * FROM dbo.__SchemaMigrations";
        return await db.QueryAsync<MigrationHistory>(sql);
    }

    public async Task AddExecutedScript(MigrationHistory migrationHistory)
    {
        using IDbConnection db = new SqlConnection(_connectionString);
        const string sql = @"
            INSERT INTO dbo.__SchemaMigrations (ScriptName, Hash, ExecutedAt)
            VALUES (@ScriptName, @Hash, @ExecutedAt)";
        await db.ExecuteAsync(sql, migrationHistory);
    }
}
