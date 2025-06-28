using System.Data;
using Dapper;
using Microsoft.Data.SqlClient;

namespace SqlMigration.Services;

public class ScriptExecutor : IScriptExecutor
{
    public async Task ExecuteScript(string connectionString, string scriptContent)
    {
        using IDbConnection db = new SqlConnection(connectionString);
        await db.ExecuteAsync(scriptContent);
    }
}
