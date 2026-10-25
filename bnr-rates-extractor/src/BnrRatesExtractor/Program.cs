using System.Globalization;
using BnrRatesExtractor;

if (args.Length != 2)
{
    PrintUsage();
    return 1;
}

if (!int.TryParse(args[0], NumberStyles.Integer, CultureInfo.InvariantCulture, out var year))
{
    Console.Error.WriteLine($"Error: Invalid year '{args[0]}'. Year must be a number.");
    return 1;
}

var currencyCode = args[1].Trim().ToUpperInvariant();

if (year < BnrRatesParser.FirstSupportedYear || year > DateTime.UtcNow.Year)
{
    Console.Error.WriteLine(
        $"Error: Invalid year '{year}'. Must be between {BnrRatesParser.FirstSupportedYear} and {DateTime.UtcNow.Year}.");
    return 1;
}

if (!BnrRatesParser.IsValidCurrencyCode(currencyCode))
{
    Console.Error.WriteLine($"Error: Invalid currency code '{args[1]}'. Expected a 3-letter code, e.g. USD.");
    return 1;
}

try
{
    return await ExtractRatesAsync(year, currencyCode);
}
catch (Exception ex) when (ex is HttpRequestException or InvalidDataException or TaskCanceledException)
{
    Console.Error.WriteLine($"Error: {ex.Message}");
    return 1;
}

static async Task<int> ExtractRatesAsync(int year, string currencyCode)
{
    Console.WriteLine($"Fetching rates for {currencyCode} in {year}...");

    using var httpClient = new HttpClient { Timeout = TimeSpan.FromSeconds(30) };
    var version = typeof(BnrRatesParser).Assembly.GetName().Version?.ToString(3) ?? "0.0.0";
    httpClient.DefaultRequestHeaders.UserAgent.ParseAdd($"bnr-rates-extractor/{version}");

    var xmlContent = await BnrRatesParser.FetchYearXmlAsync(httpClient, year);
    var rates = BnrRatesParser.Parse(xmlContent, currencyCode, year);

    if (rates.Count == 0)
    {
        Console.Error.WriteLine(
            $"Error: No rates found for {currencyCode} in {year}. "
            + "Check the currency code (BNR publishes e.g. EUR, USD, GBP, XAU) and the year.");
        return 1;
    }

    var fileName = $"{currencyCode}_{year}_rates.csv";
    var filePath = Path.Combine(Directory.GetCurrentDirectory(), fileName);

    Console.WriteLine($"Writing {rates.Count} rates to {fileName}...");
    await File.WriteAllTextAsync(filePath, BnrRatesParser.ToCsv(rates));

    Console.WriteLine($"Successfully created {fileName} with {rates.Count} rates");
    return 0;
}

static void PrintUsage()
{
    Console.WriteLine("BNR Rates Extractor - Extracts exchange rates from National Bank of Romania");
    Console.WriteLine();
    Console.WriteLine("Usage: bnr-rates-extractor <year> <currency-code>");
    Console.WriteLine();
    Console.WriteLine("Arguments:");
    Console.WriteLine($"  <year>          The year for which to extract rates, {BnrRatesParser.FirstSupportedYear}-{DateTime.UtcNow.Year} (e.g., 2024)");
    Console.WriteLine("  <currency-code> The currency code to extract (e.g., USD, EUR, GBP)");
    Console.WriteLine();
    Console.WriteLine("Example:");
    Console.WriteLine("  bnr-rates-extractor 2024 USD");
}
