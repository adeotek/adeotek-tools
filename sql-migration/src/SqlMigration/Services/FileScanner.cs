namespace SqlMigration.Services;

public class FileScanner : IFileScanner
{
    public IEnumerable<string> ScanForSqlFiles(string directory)
    {
        return Directory.EnumerateFiles(directory, "*.sql", SearchOption.AllDirectories);
    }
}
