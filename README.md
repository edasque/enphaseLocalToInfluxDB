# NG Envoy local data extractor

This is very much a word in progress. Right now it sends data to influxdb. Looking to do prometheus metrics, some REST endpoint, etc. It doesn't use the stream API endpoint which I can't seem to auth for.

## Authorization flow

The authentication flow for accessing local APIs on a newer Enphase Envoy is counter-intuitive. Thou

When trying to access the local API via https://envoy.lan/home#auth for example, through the browser, one will be redirected to https://envoy.lan/home#auth which asks for a JWT token. This token is obtained by login-in interactively to the https://entrez.enphaseenergy.com with your Enlighten credentials.

Note that you have to disable all certificate verifications on local calls

### Obtaining a long-term JWT Token

In order to obtain the token in a back-end process, one has to POST to https://enlighten.enphaseenergy.com//login/login the *user[email]* and *user[password]*.

Armed with the cookies returned by the endpoint you will then hit https://enlighten.enphaseenergy.com/entrez-auth-token?serial_num=SERIAL_NUMBER. SERIAL_NUMBER is the serial number of your envoy.

This will return a JSON object that looks like so:
```JSON
{
    "generation_time":1657053332,
    "token":"DE34wMDAasdADS1ZC03_LONG_LONG_TRUNCATED_LONG_LONG_bT6enusdwr23DQ",
    "expires_at":1672605332
}
```
*generation_time* and *expires_at* are Unix Timestamps in seconds. *token* is your JWT token

Typically once you have a JWT Token, you can use it by passing it in the header of the calls to the local API as a Bearer Token. Not so here. Calls to the Local API do not use the token but rely on a set of cookies to be there for authentication.

However to obtain that cookie jar, you have to authorize the JWT token with a local call

### Obtaining the cookies

By making a call to https://envoy.lan/auth/check_jwt with a *"Authorization"* token set to "Bearer REPLACE_WITH_JWT_TOKEN" you will receive a cookie jar. You will re-use that cookie jar to make all of your local API calls.

### Making local API calls

Once you've obtained the cookie jar, you can make calls to https://envoy.lan/production.json?details=1 or others providing your HTTP client is configured to include the cookie jar.

## Stream endpoint

Though others have been successful hitting the stream endpoint (https://envoy.lan/stream/meter), it always returns a 401 for me on release D7.0.107 (00f3a9). I tied using the cookies, the JWT token and basic auth with various installer credentials.
## Interesting local API endpoints

* ENDPOINT_URL_PRODUCTION_JSON = "https://envoy.lan/production.json"
* ENDPOINT_URL_PRODUCTION_V1 = "https://envoy.lan/api/v1/production"
* ENDPOINT_URL_PRODUCTION_INVERTERS = "https://envoy.lan/api/v1/production/inverters"
* ENDPOINT_URL_PRODUCTION = "https://envoy.lan/production"
* ENDPOINT_URL_CHECK_JWT = "https://envoy.lan/auth/check_jwt"
* ENDPOINT_URL_ENSEMBLE_INVENTORY = "https://envoy.lan/ivp/ensemble/inventory"
