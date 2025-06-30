
using System.CommandLine;
using System.CommandLine.Parsing;
using Microsoft.Extensions.Logging;
using SqlMigration.Models;
using SqlMigration.Repositories;
using SqlMigration.Services;

namespace SqlMigration.CommandLine;

public static class SqlMigrationCommand
{
    public const string Description = "Executes SQL migration scripts against a database.";

    public static async Task<int> ExecuteAsync(ParseResult parseResult, CancellationToken ct)
    {
        var isVerbose = parseResult.GetValueForOption(CommandLineManager.VerboseOption);
        var isDryRun = parseResult.GetValueForOption(CommandLineManager.DryRunOption);
        ConsoleLogger.WriteSuccess(isDryRun
            ? "Executing SQL migration command in [DryRun] mode..."
            : "Executing SQL migration command...");
        if (isVerbose)
        {
            ConsoleLogger.WriteLine(parseResult.ToString(), ConsoleColor.DarkCyan);
        }

        try
        {
            // Extract and validate the target path
            var result = ValidateTargetPath(parseResult, out var targetPath);
            if (result != 0) return result;
            // Extract and validate connection parameters
            result = ValidateConnectionParameters(parseResult, out var connectionParameters);
            if (result != 0) return result;

            var logger = new ConsoleLogger<MigrationService>(LogLevel.Debug);
            var sqlScriptsHelpers = new SqlScriptsHelpers();
            var migrationHistoryRepositoryFactory = new MigrationHistoryRepositoryFactory();
            var migrationService = new MigrationService(sqlScriptsHelpers, migrationHistoryRepositoryFactory, logger, isDryRun);
            await migrationService.RunAsync(targetPath, connectionParameters, ct).ConfigureAwait(false);

            ConsoleLogger.WriteSuccess("DONE!!!");
            return 0;
        }
        catch (Exception e)
        {
            ConsoleLogger.WriteException(e, "SqlMigrationCommand.ExecuteAsync");
            return 1;
        }
    }

    public static List<Option> CommandOptions =>
    [
        TargetPathOption,
        ProviderOption,
        ConnectionStringOption,
        HostOption,
        PortOption,
        NameOption,
        UserOption,
        PasswordOption
    ];

    private static int ValidateConnectionParameters(ParseResult parseResult, out ConnectionParameters connectionParameters)
    {
        connectionParameters = new ConnectionParameters(
            parseResult.GetValueForOption(ProviderOption) ?? nameof(DatabaseProvider.PostgreSql),
            parseResult.GetValueForOption(ConnectionStringOption),
            parseResult.GetValueForOption(HostOption),
            parseResult.GetValueForOption(PortOption),
            parseResult.GetValueForOption(NameOption),
            parseResult.GetValueForOption(UserOption),
            parseResult.GetValueForOption(PasswordOption)
        );
        if (connectionParameters.IsValid(out var errors))
        {
            return 0;
        }

        foreach (var error in errors)
        {
            ConsoleLogger.WriteError(error);
        }

        return 20;
    }

    private static int ValidateTargetPath(ParseResult parseResult, out string targetPath)
    {
        targetPath = parseResult.GetValueForOption(TargetPathOption) ?? string.Empty;
        if (string.IsNullOrWhiteSpace(targetPath) || !Directory.Exists(targetPath))
        {
            ConsoleLogger.WriteError("`--target-path` must be a valid directory path");
            return 10;
        }
        if (Directory.GetFiles(targetPath).Length == 0 && Directory.GetDirectories(targetPath).Length == 0)
        {
            ConsoleLogger.WriteError("`--target-path` directory is empty");
            return 11;
        }
        return 0;
    }

    private static readonly Option<string> TargetPathOption = new("--target-path", "-t")
    {
        IsRequired = true,
        Description = "Target path (path to the SQL scripts directory)"
    };

    private static readonly Option<string> ProviderOption = new("--provider", "-r")
    {
        Description = "Database provider (PostgreSQL/SQLite)" +
                      $" or set `{CommandLineManager.EnvironmentVariablesPrefix}PROVIDER` environment variable"
    };

    private static readonly Option<string> ConnectionStringOption = new("--connection-string", "-c")
    {
        Description = "Database connection string" +
                      $" or set `{CommandLineManager.EnvironmentVariablesPrefix}CONNECTION_STRING` environment variable"
    };

    private static readonly Option<string> HostOption = new("--host", "-o")
    {
        Description = "Database host" +
                      $" or set `{CommandLineManager.EnvironmentVariablesPrefix}HOST` environment variable"
    };

    private static readonly Option<int> PortOption = new("--port", "-p")
    {
        Description = "Database port" +
                      $" or set `{CommandLineManager.EnvironmentVariablesPrefix}PORT` environment variable"
    };

    private static readonly Option<string> NameOption = new("--name", "-n")
    {
        Description = "Database name" +
                      $" or set `{CommandLineManager.EnvironmentVariablesPrefix}DATABASE_NAME` environment variable"
    };

    private static readonly Option<string> UserOption = new("--user", "-u")
    {
        Description = "Database user" +
                      $" or set `{CommandLineManager.EnvironmentVariablesPrefix}USER` environment variable"
    };

    private static readonly Option<string> PasswordOption = new("--password", "-s")
    {
        Description = "Database password" +
                      $" or set `{CommandLineManager.EnvironmentVariablesPrefix}PASSWORD` environment variable"
    };
}
