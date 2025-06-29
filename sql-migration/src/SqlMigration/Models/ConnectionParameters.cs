namespace SqlMigration.Models;

public enum DatabaseProvider
{
    Unknown,
    PostgreSql,
    SqLite
}

public record ConnectionParameters(
    string? RawConnectionString,
    string? Provider,
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

        if (!string.IsNullOrWhiteSpace(RawConnectionString))
        {
            errors = [];
            return true;
        }

        if (Provider is null && Host is null && Port is null && DatabaseName is null)
        {
            errors = ["Either the Database connection string or the individual parameters must be provided."];
            return true;
        }

        DatabaseProvider? databaseProvider = null;
        if (string.IsNullOrWhiteSpace(Provider)
            || !Enum.TryParse<DatabaseProvider>(Provider, true, out var provider)
            || provider == DatabaseProvider.Unknown)
        {
            errorList.Add("Invalid or missing Database provider.");
        }
        else
        {
            databaseProvider = provider;
        }

        if (string.IsNullOrWhiteSpace(DatabaseName))
        {
            errorList.Add("Database name must be provided.");
        }

        if (databaseProvider == DatabaseProvider.PostgreSql && string.IsNullOrWhiteSpace(Host))
        {
            errorList.Add("Database host must be provided for PostgreSQL databases.");
        }

        if (databaseProvider == DatabaseProvider.PostgreSql && Port is null)
        {
            errorList.Add("Database port must be provided for PostgreSQL databases.");
        }

        if (databaseProvider == DatabaseProvider.PostgreSql && Port is <= 0 or > 65535)
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

        if (string.IsNullOrEmpty(Provider) || !Enum.TryParse<DatabaseProvider>(Provider, true, out var provider))
        {
            throw new ArgumentException("Invalid or missing Database provider.");
        }

        return provider switch
        {
            DatabaseProvider.PostgreSql =>
                $"Host={Host};Port={Port};Database={DatabaseName};Username={User};Password={Password};",
            DatabaseProvider.SqLite => $"Data Source={DatabaseName};",
            _ => throw new ArgumentException($"Unknown Database provider {provider}")
        };
    }
}
