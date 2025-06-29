using System.Collections;
using System.CommandLine;

namespace SqlMigration.CommandLine;

public static class CommandLineManager
{
    public const string EnvironmentVariablesPrefix = "CLI_SQL_MIGRATION_";

    public static async Task<int> ExecuteCommandAsync(string[] args)
    {
        try
        {
            return await BuildCommandLineConfiguration()
                .InvokeAsync(ProcessArgs(args));
        }
        catch (Exception e)
        {
            ConsoleLogger.WriteException(e);
            return 1;
        }
    }

    private static CommandLineConfiguration BuildCommandLineConfiguration()
    {
        RootCommand rootCommand = new()
        {
            Description = SqlMigrationCommand.Description,
            TreatUnmatchedTokensAsErrors = false
        };
        SqlMigrationCommand.CommandOptions.ForEach(rootCommand.Options.Add);
        rootCommand.Options.Add(DryRunOption);
        rootCommand.Options.Add(VerboseOption);
        rootCommand.SetAction(SqlMigrationCommand.ExecuteAsync);

        return new CommandLineConfiguration(rootCommand)
        {
            EnableDefaultExceptionHandler = false,
            ProcessTerminationTimeout = TimeSpan.FromSeconds(1800) // 30 minutes
        };
    }

    public static readonly Option<bool> DryRunOption =
        new("--dry-run", "-d")
        {
            Description = "Perform a dry run without making any changes",
            Required = false
        };

    public static readonly Option<bool> VerboseOption =
        new("--verbose", "-v")
        {
            Description = "Enable verbose logging",
            Required = false
        };

    private static string[] ProcessArgs(string[] args)
    {
        List<string> processedArgs = new(args);
        foreach (DictionaryEntry envVar in Environment.GetEnvironmentVariables())
        {
            var key = envVar.Key.ToString();
            if (key is null || key.StartsWith(EnvironmentVariablesPrefix) != true)
            {
                continue;
            }

            var varKey = "--" + key
                .Replace(EnvironmentVariablesPrefix, string.Empty, StringComparison.OrdinalIgnoreCase)
                .Replace('_', '-')
                .ToLowerInvariant();

            if (varKey == "--" || processedArgs.Any(arg => arg.Equals(varKey, StringComparison.OrdinalIgnoreCase))) {
                continue;
            }

            processedArgs.Add(varKey);
            processedArgs.Add(envVar.Value?.ToString() ?? string.Empty);
        }
        return processedArgs.ToArray();
    }
}
