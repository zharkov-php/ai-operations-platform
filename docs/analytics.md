# Analytics query behavior

Analytics use half-open UTC intervals (`from <= request_timestamp < to`). Calendar-date parameters are interpreted at midnight in the requested IANA timezone before conversion to UTC, so daylight-saving transitions remain correct. Ranges are limited to 366 days.

Cost aggregation always filters one ISO currency. When the caller omits currency, the organization's default currency is used; the platform does not silently perform foreign-exchange conversion. Monthly projection is an estimate calculated from the selected range's observed daily average multiplied by 30.4375 days.

Queries are bounded: cost breakdowns return at most 100 rows and daily trends at most 367 rows. The ingestion migration supplies organization/time, project/time, and workload/time indexes. Queries begin with organization and time boundaries, then apply optional project, workload, provider, and model filters. Production operators should review `EXPLAIN (ANALYZE, BUFFERS)` against their own distribution before changing indexes.
