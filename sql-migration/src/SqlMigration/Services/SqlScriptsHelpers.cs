using System.Data;
using System.Security.Cryptography;
using Dapper;
using Microsoft.Data.SqlClient;
using SqlMigration.Models;

namespace SqlMigration.Services;

public class SqlScriptsHelpers : ISqlScriptsHelpers
{
    public IEnumerable<string> ScanForSqlFiles(string directory)
    {
        if (string.IsNullOrEmpty(directory) || !Directory.Exists(directory))
        {
            throw new DirectoryNotFoundException($"The directory '{directory}' does not exist.");
        }

        return Directory.EnumerateFiles(directory, "*.sql", SearchOption.AllDirectories);
    }

    public string CalculateHash(string filePath)
    {
        using var sha256 = SHA256.Create();
        using var stream = File.OpenRead(filePath);
        var hash = sha256.ComputeHash(stream);
        return Convert.ToHexStringLower(hash);
    }

    public async Task ExecuteScriptAsync(string scriptContent, ConnectionParameters connectionParameters)
    {
        using IDbConnection db = new SqlConnection(connectionParameters.GetConnectionString());
        await db.ExecuteAsync(scriptContent);
    }
}
