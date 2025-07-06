using System.Collections;
using System.CommandLine;

namespace SqlMigration.CommandLine;

public static class CommandLineManager
{
    public const string EnvironmentVariablesPrefix = "CLI_SQL_MIGRATION_";

    public static async Task<int> ExecuteCommandAsync(string[] args)
    {
        var rootCommand = new RootCommand(SqlMigrationCommand.Description);
        SqlMigrationCommand.CommandOptions.ForEach(rootCommand.Options.Add);
        rootCommand.Options.Add(DryRunOption);
        rootCommand.Options.Add(VerboseOption);
        rootCommand.SetAction(SqlMigrationCommand.ExecuteAsync);
        return await rootCommand.Parse(ProcessArgs(args)).InvokeAsync();
    }

    public static readonly Option<bool> DryRunOption =
        new("--dry-run", "-d")
        {
            Description = "Run the command in dry-run mode, which simulates the execution without making any changes."
        };

    public static readonly Option<bool> VerboseOption =
        new("--verbose", "-v")
        {
            Description = "Enable verbose output, providing detailed information about the command execution."
        };

    private static string[] ProcessArgs(string[] args)
    {
        List<string> processedArgs = new(args);
        foreach (DictionaryEntry envVar in Environment.GetEnvironmentVariables())
        {
            var key = envVar.Key.ToString();
            if (key is null || key.StartsWith(EnvironmentVariablesPrefix) != true || string.IsNullOrEmpty(envVar.Value?.ToString()))
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
