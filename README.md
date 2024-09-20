# OpenWeatherMap Store
Service to request and cache weather data using OpenWeatherMap API 3.0. The intention is to abstract the requests made
to the OpenWeatherMap API and provide controlled means of accessing cached weather data which is no longer bound to the
API rate limitation in a form of GRPC API defined in
[proto definitions](https://github.com/giedrius-slegeris/proto-definitions/blob/main/protos/openweathermap-store.proto)
for the project.

## Prerequisites
- latest Docker installed and running

## Environment

Create an .env file in the root directory of the project with below arguments:

- OPEN_WEATHER_MAP_API_KEY (API 3.0 enabled access key)
- OPEN_WEATHER_MAP_BASE_URL (currently tailored to [one call endpoint](https://api.openweathermap.org/data/3.0/onecall))
- OPEN_WEATHER_MAP_LATITUDE
- OPEN_WEATHER_MAP_LONGITUDE
- OPEN_WEATHER_MAP_UNITS (either metric or imperial)
- TIMEZONE
- SCHEDULER_CRON (cron format for the scheduler to run API calls e.g. "*/5 * * * *" to run every 5 minutes)
- GRPC_LISTEN_PORT

example below:

> OPEN_WEATHER_MAP_API_KEY=yourProvisionedApiKey\
> OPEN_WEATHER_MAP_BASE_URL=https://api.openweathermap.org/data/3.0/onecall \
> OPEN_WEATHER_MAP_LATITUDE=51.500937\
> OPEN_WEATHER_MAP_LONGITUDE=-0.124602\
> OPEN_WEATHER_MAP_UNITS=metric\
> TIMEZONE="Europe/London"\
> SCHEDULER_CRON="*/5 * * * *"\
> GRPC_LISTEN_PORT=10060

## RUN OpenWeatherMap Store

Launch docker container by:

``docker compose up``