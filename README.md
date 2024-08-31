# OpenWeatherMap Store
Service to request and cache weather data using OpenWeatherMap API 3.0. The intention is to abstract the requests made
to the OpenWeatherMap API and provide easy means to access cached weather data which is no longer bound to the API rate
limitation.

## Environment
The following environment variables should be set either in the working VM, container or .env file placed in the root of
the project:

- OPEN_WEATHER_MAP_API_KEY (API 3.0 enabled access key)
- OPEN_WEATHER_MAP_BASE_URL (currently tailored to one call endpoint https://api.openweathermap.org/data/3.0/onecall)
- OPEN_WEATHER_MAP_LATITUDE
- OPEN_WEATHER_MAP_LONGITUDE
- OPEN_WEATHER_MAP_UNITS (either metric or imperial)
- TIMEZONE (Europe/London)
- SCHEDULER_CRON (cron format for the scheduler to run API calls e.g. "*/5 * * * *" to run every 5 minutes)
