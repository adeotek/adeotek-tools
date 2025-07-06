namespace SqlMigration.Models;

public record ConnectionParameters(
    string Provider,
    string? RawConnectionString,
    string? Host,
    int? Port,
    string? DatabaseName,
    string? User,
    string? Password)
{
    public DatabaseProvider DbProvider =>
        string.IsNullOrWhiteSpace(Provider) || !Enum.TryParse<DatabaseProvider>(Provider, true, out var provider)
            ? DatabaseProvider.Unknown
            : provider;

    public bool IsValid(out string[] errors)
    {
        var errorList = new List<string>();
        var provider = DbProvider;

        if (provider == DatabaseProvider.Unknown)
        {
            errorList.Add("Invalid or missing Database provider.");
        }

        if (!string.IsNullOrWhiteSpace(RawConnectionString))
        {
            errors = errorList.ToArray();
            return errorList.Count == 0;
        }

        if (Host is null && Port is null && DatabaseName is null)
        {
            errors = ["Either the Database connection string or the individual parameters must be provided."];
            return true;
        }

        if (string.IsNullOrWhiteSpace(DatabaseName))
        {
            errorList.Add("Database name must be provided.");
        }

        if (provider == DatabaseProvider.PostgreSql && string.IsNullOrWhiteSpace(Host))
        {
            errorList.Add("Database host must be provided for PostgreSQL databases.");
        }

        if (provider == DatabaseProvider.PostgreSql && Port is null)
        {
            errorList.Add("Database port must be provided for PostgreSQL databases.");
        }

        if (provider == DatabaseProvider.PostgreSql && Port is <= 0 or > 65535)
        {
            errorList.Add("Database port must be a valid integer between 1 and 65535.");
        }

        if (errorList.Count > 0)
        {
            errors = errorList.ToArray();
            return false;
        }

        errors = [];
        return true;
    }

    public string GetConnectionString()
    {
        if (!string.IsNullOrWhiteSpace(RawConnectionString))
        {
            return RawConnectionString;
        }

        return DbProvider switch
        {
            DatabaseProvider.PostgreSql =>
                $"Host={Host};Port={Port};Database={DatabaseName};Username={User};Password={Password};",
            DatabaseProvider.SqLite => $"Data Source={DatabaseName};",
            DatabaseProvider.Unknown => throw new ArgumentException("Unknown Database provider"),
            _ => throw new ArgumentOutOfRangeException(nameof(DbProvider), DbProvider, "Unsupported database provider")
        };
    }
}
