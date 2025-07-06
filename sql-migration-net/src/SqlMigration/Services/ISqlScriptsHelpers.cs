using SqlMigration.Models;

namespace SqlMigration.Services;

public interface ISqlScriptsHelpers
{
    List<string> ScanForSqlFiles(string directory);
    string CalculateHash(string filePath);
    Task ExecuteScriptAsync(string scriptContent, ConnectionParameters connectionParameters);
}
