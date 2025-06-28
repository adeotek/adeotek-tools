using System.Security.Cryptography;

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
        return BitConverter.ToString(hash).Replace("-", "").ToLowerInvariant();
    }
}
