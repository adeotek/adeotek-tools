using System.Security.Cryptography;
using Dapper;
using SqlMigration.Models;

namespace SqlMigration.Services;

public class SqlScriptsHelpers : ISqlScriptsHelpers
{
    public List<string> ScanForSqlFiles(string directory)
    {
        if (string.IsNullOrEmpty(directory) || !Directory.Exists(directory))
        {
            throw new DirectoryNotFoundException($"The directory '{directory}' does not exist.");
        }

        var unorderedScripts = Directory
            .EnumerateFiles(directory, "*.sql", SearchOption.AllDirectories)
            .Select(file => Path.GetRelativePath(directory, file));

        List<string> orderedScripts = [];
        orderedScripts.AddRange(unorderedScripts
            .Where(file => file.StartsWith("tables/"))
            .OrderBy(file => file));
        orderedScripts.AddRange(unorderedScripts
            .Where(file => file.StartsWith("views/"))
            .OrderBy(file => file));
        orderedScripts.AddRange(unorderedScripts
            .Where(file => file.StartsWith("stored_procedures/"))
            .OrderBy(file => file));
        orderedScripts.AddRange(unorderedScripts
            .Where(file => file.StartsWith("data/"))
            .OrderBy(file => file));

        return orderedScripts;
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
        using var db = DbConnectionFactory.Create(connectionParameters);
        await db.ExecuteAsync(scriptContent);
    }
}
