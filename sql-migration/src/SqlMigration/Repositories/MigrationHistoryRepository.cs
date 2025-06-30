using Dapper;
using SqlMigration.Models;
using SqlMigration.Services;

namespace SqlMigration.Repositories;

public class MigrationHistoryRepository(ConnectionParameters connectionParameters)
    : IMigrationHistoryRepository
{
    private const string TableName = "__migrations_history";

    private readonly IDbConnectionFactory _dbConnectionFactory = new DbConnectionFactory(connectionParameters);

    public async Task<bool> IsHistoryTableCreatedAsync(CancellationToken ct = default)
    {
        var sql = _dbConnectionFactory.Provider == DatabaseProvider.PostgreSql
            ? PostgreSqlIsHistoryTableCreated
            : SqLiteIsHistoryTableCreated;
        using var connection = _dbConnectionFactory.Create();
        return await connection.ExecuteScalarAsync<int>(sql) > 0;
    }

    public async Task CreateHistoryTableAsync()
    {
        var sql = _dbConnectionFactory.Provider == DatabaseProvider.PostgreSql
            ? PostgreSqlHistoryTableCreate
            : SqLiteHistoryTableCreate;
        using var connection = _dbConnectionFactory.Create();
        await connection.ExecuteAsync(sql);
    }

    public async Task<IEnumerable<ScriptExecutionHistory>> GetExecutedScriptsAsync(CancellationToken ct = default)
    {
        using var connection = _dbConnectionFactory.Create();
        return await connection.QueryAsync<ScriptExecutionHistory>(SqlHistoryTableSelect);
    }

    public async Task UpsertExecutedScriptAsync(ScriptExecutionHistory scriptExecutionHistory)
    {
        using var connection = _dbConnectionFactory.Create();
        var existingScript = await connection.QueryFirstOrDefaultAsync<ScriptExecutionHistory>(
            SqlHistoryTableSelectByPk, new { scriptExecutionHistory.ScriptFile });
        if (existingScript is null)
        {
            await connection.ExecuteAsync(SqlHistoryTableInsert, scriptExecutionHistory);
        }
        else
        {
            scriptExecutionHistory.ExecutedAt = DateTime.UtcNow;
            await connection.ExecuteAsync(SqlHistoryTableUpdate, scriptExecutionHistory);
        }
    }

    private const string PostgreSqlIsHistoryTableCreated =
        $"SELECT count(*) FROM information_schema.tables WHERE table_name = '{TableName}';";
    private const string SqLiteIsHistoryTableCreated =
        $"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='{TableName}';";
    private const string SqlHistoryTableSelect =
        $"SELECT * FROM {TableName}";
    private const string SqlHistoryTableSelectByPk =
        $"SELECT * FROM {TableName} WHERE ScriptFile = @ScriptFile;";
    private const string SqlHistoryTableInsert =
        $"INSERT INTO {TableName} (ScriptFile, ScriptHash, ExecutedAt) VALUES (@ScriptFile, @ScriptHash, @ExecutedAt);";
    private const string SqlHistoryTableUpdate =
        $"UPDATE {TableName} SET ScriptHash = @ScriptHash, ExecutedAt = @ExecutedAt WHERE ScriptFile = @ScriptFile;";

    private const string PostgreSqlHistoryTableCreate =
        $"""
        CREATE TABLE {TableName} (
            ScriptFile varchar(255) NOT NULL,
            ScriptHash varchar(150) NOT NULL,
            ExecutedAt timestamptz DEFAULT now() NOT NULL,
            CONSTRAINT {TableName}_pk PRIMARY KEY (ScriptFile)
        )
        """;

    private const string SqLiteHistoryTableCreate =
        $"""
          CREATE TABLE {TableName} (
              ScriptFile TEXT NOT NULL PRIMARY KEY,
              ScriptHash TEXT NOT NULL,
              ExecutedAt DATETIME NOT NULL
          )
          """;
}
