namespace SqlMigration.Services;

public interface IFileScanner
{
    IEnumerable<string> ScanForSqlFiles(string directory);
}
