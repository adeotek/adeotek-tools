using SqlMigration.Models;

namespace SqlMigration.Repositories;

public class MigrationHistoryRepository(ISqlRepository sqlRepository)
    : IMigrationHistoryRepository
{
    private const string TableName = "__migrations_history";

    public async Task<bool> IsHistoryTableCreatedAsync(CancellationToken ct = default)
    {
        var sql = sqlRepository.DbConnectionFactory.Provider == DatabaseProvider.PostgreSql
            ? PostgreSqlIsHistoryTableCreated
            : SqLiteIsHistoryTableCreated;
        return await sqlRepository.ExecuteScalarAsync<int>(sql) > 0;
    }

    public async Task CreateHistoryTableAsync()
    {
        var sql = sqlRepository.DbConnectionFactory.Provider == DatabaseProvider.PostgreSql
            ? PostgreSqlHistoryTableCreate
            : SqLiteHistoryTableCreate;
        await sqlRepository.ExecuteAsync(sql);
    }

    public async Task<IEnumerable<ScriptExecutionHistory>> GetExecutedScriptsAsync(CancellationToken ct = default)
    {
        return await sqlRepository.QueryAsync<ScriptExecutionHistory>(SqlHistoryTableSelect);
    }

    public async Task UpsertExecutedScriptAsync(ScriptExecutionHistory scriptExecutionHistory)
    {
        var existingScript = await sqlRepository.QueryFirstOrDefaultAsync<ScriptExecutionHistory>(
            SqlHistoryTableSelectByPk, new { scriptExecutionHistory.ScriptFile });
        if (existingScript is null)
        {
            await sqlRepository.ExecuteAsync(SqlHistoryTableInsert, scriptExecutionHistory);
        }
        else
        {
            scriptExecutionHistory.ExecutedAt = DateTime.UtcNow;
            await sqlRepository.ExecuteAsync(SqlHistoryTableUpdate, scriptExecutionHistory);
        }
    }

    private const string PostgreSqlIsHistoryTableCreated =
        $"SELECT count(*) FROM information_schema.tables WHERE table_name = '{TableName}';";
    private const string SqLiteIsHistoryTableCreated =
        $"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='{TableName}';";
    private const string SqlHistoryTableSelect =
        $"SELECT * FROM \"{TableName}\";";
    private const string SqlHistoryTableSelectByPk =
        $"SELECT * FROM \"{TableName}\" WHERE \"ScriptFile\" = @ScriptFile;";
    private const string SqlHistoryTableInsert =
        $"INSERT INTO \"{TableName}\" (\"ScriptFile\", \"ScriptHash\", \"ExecutedAt\") VALUES (@ScriptFile, @ScriptHash, @ExecutedAt);";
    private const string SqlHistoryTableUpdate =
        $"UPDATE \"{TableName}\" SET \"ScriptHash\" = @ScriptHash, \"ExecutedAt\" = @ExecutedAt WHERE \"ScriptFile\" = @ScriptFile;";

    private const string PostgreSqlHistoryTableCreate =
        $"""
        CREATE TABLE "{TableName}" (
            "ScriptFile" varchar(255) NOT NULL,
            "ScriptHash" varchar(150) NOT NULL,
            "ExecutedAt" timestamptz DEFAULT now() NOT NULL,
            CONSTRAINT "{TableName}_pk" PRIMARY KEY ("ScriptFile")
        )
        """;

    private const string SqLiteHistoryTableCreate =
        $"""
          CREATE TABLE "{TableName}" (
              "ScriptFile" TEXT NOT NULL PRIMARY KEY,
              "ScriptHash" TEXT NOT NULL,
              "ExecutedAt" DATETIME NOT NULL
          )
          """;
}
