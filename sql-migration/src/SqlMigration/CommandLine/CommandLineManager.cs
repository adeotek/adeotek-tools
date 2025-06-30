using System.Collections;
using System.CommandLine;
using System.CommandLine.Builder;
using System.CommandLine.Parsing;

namespace SqlMigration.CommandLine;

public static class CommandLineManager
{
    public const string EnvironmentVariablesPrefix = "CLI_SQL_MIGRATION_";

    public static async Task<int> ExecuteCommandAsync(string[] args)
    {
        var rootCommand = new RootCommand(SqlMigrationCommand.Description);
        SqlMigrationCommand.CommandOptions.ForEach(rootCommand.AddOption);
        rootCommand.AddOption(DryRunOption);
        rootCommand.AddOption(VerboseOption);

        rootCommand.SetHandler(async context =>
        {
            var parseResult = context.ParseResult;
            var cancellationToken = context.GetCancellationToken();
            await SqlMigrationCommand.ExecuteAsync(parseResult, cancellationToken);
        });

        var commandLineBuilder = new CommandLineBuilder(rootCommand)
            .UseDefaults();

        var parser = commandLineBuilder.Build();
        
        return await parser.InvokeAsync(ProcessArgs(args));
    }

    public static readonly Option<bool> DryRunOption =
        new("--dry-run", "-d");

    public static readonly Option<bool> VerboseOption =
        new("--verbose", "-v");

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