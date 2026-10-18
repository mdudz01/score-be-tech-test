# Affordability Outgoings API

A small Go service that estimates a household's monthly outgoings for a mortgage affordability check.

You give it a **Case ID**. It looks up the client's postcode, works out which ONS region that is in, and returns estimated monthly spending for each outgoings category (`outgoings_food`, `outgoings_water`, etc.) using ONS household spending data.

## How to run it

You need Go 1.22 or later.

```bash
go mod download
go run ./cmd/server
```

The server starts on port 8080. Try it:

```bash
curl http://localhost:8080/cases/56763bf7-d046-4a54-a4a1-0728c724a092/outgoings
```

(On Windows PowerShell, use `Invoke-RestMethod <url>` instead of `curl`.)

### Settings

Settings are environment variables. All have defaults.

- `PORT`: port to listen on. Default `8080`.
- `REGIONAL_FILE`: path to the ONS regional spreadsheet. Default `regional.xlsx`. Set this to the real filename in the repo.
- `CASE_FILE`: path to the case JSON file. Default `case.json`.

Example (PowerShell):

```powershell
$env:REGIONAL_FILE="your-file.xlsx"
go run ./cmd/server
```

### Run the tests

```bash
go test ./...
```

## The API

### `GET /cases/{id}/outgoings`

Returns estimated **monthly** amounts in pounds, rounded to pence. All 21 fields are always returned, for example:

```json
{
  "outgoings_food": 265.20,
  "outgoings_water": 12.40,
  "outgoings_mortgage_rent": 0
}
```

(Shortened here. The real response has every field from the brief.)

If something goes wrong, you get a JSON error such as `{"error": "case not found"}` with one of these status codes:

- **404**: no case with that ID.
- **422**: the case has no postcode, or the postcode is invalid or not recognised.
- **500**: something unexpected went wrong. The details are logged on the server but not sent back.

### Other endpoints

- `GET /regions/{region}/outgoings`: the same estimates for a named region, e.g. `East%20Midlands`. Useful for checking the numbers. Returns 404 for an unknown region.
- `GET /healthz`: returns `{"status":"ok"}`. Used to check the service is alive.

## How it works

The request goes through four steps:

1. **Find the case.** The case is loaded from `case.json` when the server starts, and looked up by its `case_id`.
2. **Find the region.** The first client postcode is used. The leading letters (the "postcode area", such as `NG` from `NG1 2ET`) are looked up in a table that gives the ONS region, such as East Midlands.
3. **Find the spending.** The ONS regional spreadsheet is read once at startup into memory. Each row is identified by its ONS category code (like `4.3` or `7.2.2`), not its label, because labels contain footnote numbers and sometimes wrap onto two rows.
4. **Build the outgoings.** Each outgoings field is made from one or more ONS categories. They are added together, converted from weekly to monthly (multiply by 52 and divide by 12), and rounded to 2 decimal places.

### Reading the spreadsheet

- The sheet is found by locating the header cell `UK`. The region columns start there and run to the right.
- Where a long label wraps onto a second spreadsheet row, the code is on the first row and the numbers are on the second. The loader remembers the code and attaches the numbers from the next row. A blank row resets this.
- Cells that are not numbers (blank, `-`, `[x]`) are skipped, so that value counts as 0.

## Which ONS categories make up each field

The mapping is in one place, `Mapping` in `internal/ons/outgoings.go`, so it is easy to read and change. No category is used twice, and no parent category is used together with its own sub-category, so nothing is double counted. A test checks this.

- `outgoings_food`: 1 (food and non-alcoholic drinks)
- `outgoings_clothing`: 3 (clothing and footwear)
- `outgoings_water`: 4.3 (water supply and related services)
- `outgoings_household_repairs`: 4.2 (maintenance and repair of dwelling)
- `outgoings_fuel`: 4.4 (household electricity, gas and other fuels)
- `outgoings_communications`: 8
- `outgoings_insurance`: 12.4
- `outgoings_investments`: 14.6 (savings and investments)
- `outgoings_pension`: 14.1 (life assurance and pension contributions)
- `outgoings_council_tax`: 13.1 (see the council tax note below)
- `outgoings_car_costs`: 7.1, 7.2.1, 7.2.2, 7.2.3, 7.2.4 (buying and running a vehicle, including petrol and diesel)
- `outgoings_other_transport_costs`: 7.3 (trains, buses and so on)
- `outgoings_television_license`: 9.4.3
- `outgoings_sports`: 9.3.3, 9.4.1 (sports equipment, memberships and classes)
- `outgoings_holidays`: 9.6, 11.2 (package holidays and accommodation)
- `outgoings_entertainment`: 9.1, 9.2, 9.3.1, 9.3.2, 9.3.4, 9.3.5, 9.4.2, 9.4.4, 9.4.6, 9.5, 11.1 (electronics, hobbies, pets, cinema, gambling, books, restaurants and takeaways)
- `outgoings_child_care`: 12.3 (the closest ONS category, "social protection")
- `outgoings_other_living_costs`: 2, 5, 6, 10, 12.1, 12.2, 12.5 (alcohol and tobacco, household goods, health, education, personal care, personal effects, other services)

These three fields return **0** because there is no good ONS equivalent:

- `outgoings_mortgage_rent`
- `outgoings_ground_rent_service_charge_shared_equity_rent`
- `outgoings_additional_details`

## Assumptions and known limitations

### About the numbers

- ONS figures are **weekly**. The API returns **monthly** values (weekly x 52 / 12). To use a different period, change `weeksPerMonth` in `internal/ons/outgoings.go`.
- The figures are **averages for all households in a region**, not for this particular household. The extension section below describes how to improve this.
- The figures come from an ONS survey from a past year and are not adjusted for inflation.

### About the mapping

- Matching ONS categories to outgoings fields is a judgment call, and it is the biggest source of inaccuracy.
- **Council tax:** ONS category 13.1 combines council tax with mortgage interest, so `outgoings_council_tax` will be too high. Treat it as an upper estimate.
- **Fuel:** `outgoings_fuel` is treated as household energy (gas and electricity). Petrol and diesel are included in `outgoings_car_costs`.
- **Child care:** the ONS category used is the nearest match, but it is not child care specifically.
- **Mortgage and rent:** returned as 0 because the mortgage is the thing being assessed.

### About postcodes and regions

- The postcode lookup uses only the leading letters, so it is approximate. A few postcode areas cross regional borders (for example CH, SY, DN and TD). Each is assigned to the region that contains most of it, so some addresses will get the wrong region.
- A more accurate option would be the ONS Postcode Directory or a service such as postcodes.io. The lookup sits behind an interface, so it can be swapped without changing the rest of the code.

### About cases and clients

- The brief does not say where cases are stored, so `case.json` is loaded at startup and treated as the only case. The case storage sits behind an interface so a real database or service could replace it.
- In the sample `case.json`, both clients have the **same `client_id`**. This looks like a typo. Nothing in this service depends on client IDs being unique.
- If a case has several clients, the postcode of the **first client that has one** is used. In the sample, both clients share an address.

## Quality and maintainability

- **Speed:** the ONS data is read once at startup and kept in memory, so each request is just a few lookups.
- **Safe for many users:** the data is never changed after loading, so many requests can read it at the same time without locks.
- **Clear errors:** every error is JSON, with a sensible status code. Unexpected errors are logged and a generic message is returned, so internal details are not exposed.
- **Fails early:** if the spreadsheet or case file cannot be loaded, the server stops at startup instead of running with bad data.
- **Configurable:** file paths and the port come from environment variables.
- **Logging:** structured logs (`log/slog`) record the case ID and region. Personal details such as names and dates of birth are not logged.
- **Timeouts:** the HTTP server has read and write timeouts so slow clients cannot hold connections open.
- **Health check:** `/healthz` shows the service is up.
- **Easy to change:** the HTTP code depends on small interfaces (case storage, region lookup, ONS data), not on concrete types.
- **Privacy:** the case contains personal data. Only the postcode is used. A real deployment would also need authentication, HTTPS and rate limiting.

## Tests

Run `go test ./...` (add `-cover` to see coverage).

- **Postcode and region:** valid, lowercase, padded, single-letter areas, invalid and unknown postcodes. Also checks every region name in the lookup is a real ONS region name.
- **Outgoings calculation:** checks the weekly to monthly maths with a small hand-made data set. Also checks missing categories, the zero fields, an unknown region, that all 21 fields are always present, and that the mapping has no overlaps.
- **HTTP handler:** uses fake versions of the case storage, region lookup and ONS data to test success, case not found, storage failure, no clients, invalid postcode, unknown postcode area and a data mismatch. It checks the status code, the content type and the shape of the JSON.

## Project layout

```
cmd/server/       starts the server (settings and wiring only)
cmd/probe/        small tool that prints outgoings for one region
internal/api/     HTTP handlers and error responses (with tests)
internal/cases/   loads and finds cases
internal/domain/  Case, Client, Details and Address types
internal/ons/     reads the spreadsheet, field mapping, calculation (with tests)
internal/region/  postcode to region lookup (with tests)
case.json         the sample case
*.xlsx            the ONS datasets (region, age, composition)
```

## Extension: hybrid values

*This section describes a proposed design. It is not built yet.*

### The problem

The regional numbers are averages for every household in a region. A family of four spends differently from a single person, and a 25-year-old spends differently from a 60-year-old. The other two ONS datasets (expenditure by household composition, and by age) can be used to adjust the regional numbers for the household in the case.

### The idea

Use the regional figure as the starting point, then scale it up or down depending on how households like this one compare with the average household.

For each ONS category:

```
factor = spend for this type of household / spend for all households
estimate = regional spend x factor
```

Both numbers in the factor come from the same dataset, so we are only using it to say "households like this spend 37% more than average on food". The regional dataset still supplies the actual pound amount.

Example (made-up numbers, for illustration):

- Food in the East Midlands: £60 a week
- Food for "2 adults, 2 children" in the composition dataset: £85 a week
- Food for all households in the composition dataset: £62 a week
- Factor: 85 / 62 = 1.37
- Hybrid estimate: 60 x 1.37 = about £82 a week

The result is then converted to monthly and rounded like in the normal version.

### Working out the household type

- **Adults:** the number of clients in the case.
- **Children:** the `dependents` value. Dependents are not always children, so this is approximate.
- Pick the matching category in the composition dataset, such as "2 adults, 2 children". If there is no exact match, use the closest broader one. If nothing fits, just return the regional value.
- Both clients in the sample have `dependents: 2`. Take the value once, not added together, or the same children will be counted twice.

### Using age as well

The age dataset can give a second factor in the same way, using the date of birth of the main applicant (the rule for joint applications should be written down).

Age and household type overlap (families with children tend to be in certain age bands), so multiplying both factors would count part of the effect twice. Options:

1. **Composition only.** Simplest, and the one I would start with.
2. **Weighted mix.** For example `composition^0.7 x age^0.3`, with weights tuned by testing.
3. **Choose per category.** Use composition for things driven by household size (food, clothing, child care) and age for things driven by life stage (pension, holidays, transport).

### Safety checks

- Limit factors to a range such as 0.5 to 2.0, so one odd survey figure cannot cause an extreme result.
- If a figure is missing, use a factor of 1, so the regional value is unchanged.
- The zero fields (mortgage, rent, ground rent, additional details) stay at 0.
- Apply the factors to the same ONS categories as before, then add them up using the same mapping, so nothing is double counted.

### What would need to change in the code

- **API:** add `?mode=hybrid` to `GET /cases/{id}/outgoings`. The default stays as the regional version, so nothing changes for existing users. An unknown mode returns 400. The response shape stays the same.
- **Loader:** a new `ons.LoadComposition` that reads the composition spreadsheet the same way the regional loader does (header detection, matching rows by code, handling wrapped rows).
- **Calculation:** a new method that takes the region and the household details (adults, children, optionally age) and applies the factors.
- **Tests:** check the factor maths by hand, the limits, the fallback to regional values, and the handler with `mode=hybrid`.

### Limits of this approach

- It assumes the effect of household type is the same in every region. In reality, having children might change food spending differently in London than in the North East.
- The composition and age data are national, and some household types have small samples.
- It is still an average for a type of household. Income, tenure and lifestyle are not captured. An adviser should always be able to overwrite the estimates with the customer's real figures.

## Possible future work

- Build the hybrid mode above.
- Graceful shutdown and request logging.
- A Dockerfile and a Makefile.
- A test that loads the real spreadsheet and checks that rows with wrapped labels (such as water, 4.3) are not zero.
- An OpenAPI description of the endpoints.
- A more accurate postcode to region lookup.
- A better source for council tax.
- Replace the file-based case storage with a real service or database.
- Authentication, HTTPS and rate limiting.
