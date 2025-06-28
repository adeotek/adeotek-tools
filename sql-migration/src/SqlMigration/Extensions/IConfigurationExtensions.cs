using SqlMigration.Models;
using Microsoft.Extensions.Configuration;

namespace SqlMigration.Extensions;

public static class IConfigurationExtensions
{
    public static string GetConnectionString(this IConfiguration configuration)
    {
        var connectionString = configuration[Constants.ConnectionStringArgumentName];
        if (!string.IsNullOrEmpty(connectionString))
        {
            return connectionString;
        }

        var host = configuration["host"];
        var port = configuration["port"];
        var database = configuration["database"];
        var user = configuration["user"];
        var password = configuration["password"];

        return $"Server={host},{port};Database={database};User Id={user};Password={password};";
    }
}
