using System.Security.Cryptography;

namespace SqlMigration.Services;

public class HashCalculator : IHashCalculator
{
    public string CalculateHash(string filePath)
    {
        using var sha256 = SHA256.Create();
        using var stream = File.OpenRead(filePath);
        var hash = sha256.ComputeHash(stream);
        return BitConverter.ToString(hash).Replace("-", "").ToLowerInvariant();
    }
}
