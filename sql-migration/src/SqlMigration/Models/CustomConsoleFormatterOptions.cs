using Microsoft.Extensions.Logging.Console;

namespace SqlMigration.Models;

public sealed class CustomConsoleFormatterOptions : ConsoleFormatterOptions
{
    public bool IncludeTimestamp { get; set; }
    public string? CustomPrefix { get; set; }
}
