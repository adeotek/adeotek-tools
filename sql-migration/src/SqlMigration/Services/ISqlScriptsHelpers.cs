namespace SqlMigration.Services;

public interface ISqlScriptsHelpers
{
    IEnumerable<string> ScanForSqlFiles(string directory);
    string CalculateHash(string filePath);
}
