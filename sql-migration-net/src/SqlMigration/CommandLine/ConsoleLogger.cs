using System.Diagnostics.CodeAnalysis;
using System.Text;
using Microsoft.Extensions.Logging;

namespace SqlMigration.CommandLine;

[ExcludeFromCodeCoverage]
public class ConsoleLogger<TCategoryName>(
    LogLevel minLogLevel = LogLevel.Information,
    bool includeTimestamp = false)
    : ConsoleLogger(minLogLevel, includeTimestamp, typeof(TCategoryName).Name), ILogger<TCategoryName>;

[ExcludeFromCodeCoverage]
public class ConsoleLogger(
    LogLevel minLogLevel = LogLevel.Information,
    bool includeTimestamp = false,
    string? categoryName = null)
    : ILogger
{
    public static void WriteError(string message)
    {
        WriteLine(message, ConsoleColor.Red);
    }

    public static void WriteSuccess(string message)
    {
        WriteLine(message, ConsoleColor.Green);
    }

    public static void WriteDebug(string message)
    {
        WriteLine(message, ConsoleColor.Cyan);
    }

    public static void WriteLine(string message, ConsoleColor color)
    {
        if (!Console.IsOutputRedirected) Console.ForegroundColor = color;
        Console.WriteLine(message);
        if (!Console.IsOutputRedirected) Console.ResetColor();
    }

    public static void WriteErrorLine(string message, ConsoleColor color)
    {
        if (!Console.IsOutputRedirected) Console.ForegroundColor = color;
        Console.Error.WriteLine(message);
        if (!Console.IsOutputRedirected) Console.ResetColor();
    }

    public static void WriteException(Exception exception, string? message = null, LogLevel logLevel = LogLevel.Error)
    {
        if (logLevel == LogLevel.Error)
        {
            WriteErrorLine(FormatLogMessage(logLevel, message, exception), GetLogLevelColor(logLevel));
        }
        else
        {
            WriteLine(FormatLogMessage(logLevel, message, exception), GetLogLevelColor(logLevel));
        }
    }

    public void Log<TState>(LogLevel logLevel, EventId eventId, TState state, Exception? exception, Func<TState, Exception?, string?> formatter)
    {
        if (!IsEnabled(logLevel))
        {
            return;
        }

        var message = formatter(state, exception);
        if (message is null)
        {
            return;
        }

        WriteLine(FormatLogMessage(logLevel, message, exception, includeTimestamp, categoryName), logLevel);
    }

    public bool IsEnabled(LogLevel logLevel) => logLevel >= minLogLevel;

    public IDisposable? BeginScope<TState>(TState state) where TState : notnull => null;

    private static string FormatLogMessage(LogLevel logLevel, string? message, Exception? exception,
        bool includeTimestamp = false, string? categoryName = null)
    {
        StringBuilder sb = new();
        if (includeTimestamp)
        {
            sb.Append($"{DateTime.Now:HH:mm:ss.fff}] - ");
        }

        sb.Append($"[{logLevel switch
        {
            LogLevel.Trace => "TRACE",
            LogLevel.Debug => "DEBUG",
            LogLevel.Information => "INFO",
            LogLevel.Warning => "WARN",
            LogLevel.Error => "ERROR",
            LogLevel.Critical => "CRIT",
            _ => "UNKNOWN"
        }}] ");

        if (!string.IsNullOrEmpty(categoryName))
        {
            sb.Append($"{categoryName} > ");
        }

        if (message is not null)
        {
            sb.Append(message);
        }

        if (exception == null)
        {
            return sb.ToString();
        }

        if (message is not null)
        {
            sb.AppendLine();
        }

        sb.Append($"Exception: {exception.GetType().Name} - {exception.Message}");
        if (exception.StackTrace is null)
        {
            return sb.ToString();
        }

        sb.AppendLine();
        sb.Append(exception.StackTrace);

        return sb.ToString();
    }

    private static void WriteLine(string message, LogLevel logLevel) =>
        Console.WriteLine(message, GetLogLevelColor(logLevel));

    private static ConsoleColor GetLogLevelColor(LogLevel logLevel) =>
        logLevel switch
        {
            LogLevel.Trace => ConsoleColor.Gray,
            LogLevel.Debug => ConsoleColor.Cyan,
            LogLevel.Information => ConsoleColor.White,
            LogLevel.Warning => ConsoleColor.Yellow,
            LogLevel.Error => ConsoleColor.Red,
            LogLevel.Critical => ConsoleColor.DarkRed,
            _ => ConsoleColor.White
        };
}
