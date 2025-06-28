using System.Security.Cryptography;

namespace SqlMigration.Services;

public interface IHashCalculator
{
    string CalculateHash(string filePath);
}
