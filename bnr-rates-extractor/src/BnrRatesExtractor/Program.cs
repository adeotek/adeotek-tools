using System.Globalization;
using System.Text;
using System.Xml.Linq;

if (args.Length != 2)
{
    Console.WriteLine("BNR Rates Extractor - Extracts exchange rates from National Bank of Romania");
    Console.WriteLine();
    Console.WriteLine("Usage: bnr-rates-extractor <year> <currency-code>");
    Console.WriteLine();
    Console.WriteLine("Arguments:");
    Console.WriteLine("  <year>          The year for which to extract rates (e.g., 2024)");
    Console.WriteLine("  <currency-code> The currency code to extract (e.g., USD, EUR, GBP)");
    Console.WriteLine();
    Console.WriteLine("Example:");
    Console.WriteLine("  bnr-rates-extractor 2024 USD");
    return 1;
}

if (!int.TryParse(args[0], out var year))
{
    Console.Error.WriteLine($"Error: Invalid year '{args[0]}'. Year must be a number.");
    return 1;
}

var currencyCode = args[1];

try
{
    await ExtractRatesAsync(year, currencyCode);
    return 0;
}
catch (Exception ex)
{
    Console.Error.WriteLine($"Error: {ex.Message}");
    return 1;
}

static async Task ExtractRatesAsync(int year, string currencyCode)
{
    // Validate inputs
    if (year < 2000 || year > DateTime.Now.Year)
    {
        throw new ArgumentException($"Invalid year. Must be between 2000 and {DateTime.Now.Year}");
    }

    if (string.IsNullOrWhiteSpace(currencyCode))
    {
        throw new ArgumentException("Currency code cannot be empty");
    }

    currencyCode = currencyCode.ToUpperInvariant();

    // Fetch XML data
    Console.WriteLine($"Fetching rates for {currencyCode} in {year}...");
    var xmlUrl = $"https://www.bnr.ro/files/xml/years/nbrfxrates{year}.xml";

    using var httpClient = new HttpClient();
    httpClient.Timeout = TimeSpan.FromSeconds(30);

    var xmlContent = await httpClient.GetStringAsync(xmlUrl);

    // Parse XML
    var doc = XDocument.Parse(xmlContent);
    var ns = doc.Root?.GetDefaultNamespace() ?? XNamespace.None;

    var rates = new List<RateData>();

    // Navigate through the XML structure: DataSet -> Body -> Cube (multiple with dates) -> Rate (multiple)
    var cubes = doc.Descendants(ns + "Cube")
        .Where(c => c.Attribute("date") != null);

    foreach (var cube in cubes)
    {
        var dateStr = cube.Attribute("date")?.Value;
        if (string.IsNullOrEmpty(dateStr))
            continue;

        if (!DateTime.TryParse(dateStr, CultureInfo.InvariantCulture, DateTimeStyles.None, out var date))
            continue;

        var rateElements = cube.Elements(ns + "Rate")
            .Where(r => r.Attribute("currency")?.Value?.Equals(currencyCode, StringComparison.OrdinalIgnoreCase) == true);

        foreach (var rateElement in rateElements)
        {
            var rateValue = rateElement.Value;
            if (decimal.TryParse(rateValue, NumberStyles.Any, CultureInfo.InvariantCulture, out var rate))
            {
                rates.Add(new RateData(currencyCode, date, rate));
            }
        }
    }

    if (rates.Count == 0)
    {
        Console.WriteLine($"No rates found for {currencyCode} in {year}");
        return;
    }

    // Sort by date
    rates.Sort((a, b) => a.Date.CompareTo(b.Date));

    // Generate CSV
    var fileName = $"{currencyCode}_{year}_rates.csv";
    var filePath = Path.Combine(Directory.GetCurrentDirectory(), fileName);

    Console.WriteLine($"Writing {rates.Count} rates to {fileName}...");

    var csv = new StringBuilder();
    csv.AppendLine("Currency,Date,Rate");

    foreach (var rate in rates)
    {
        csv.AppendLine($"{rate.Currency},{rate.Date:yyyy-MM-dd},{rate.Rate:F4}");
    }

    await File.WriteAllTextAsync(filePath, csv.ToString());

    Console.WriteLine($"Successfully created {fileName} with {rates.Count} rates");
}

readonly record struct RateData(string Currency, DateTime Date, decimal Rate);
