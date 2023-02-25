# Shoe price checker
Sends an email, via Gmail, if a shoe meets a price/query threshold.

## Setup
Setup involves:
- Visiting Running Wearhouse to get the Solr query (see below)
- Configuring a `.env` file where the settings are stored

### Find the Solr query URL
1. Visit Running Warehouse, bring up the Network Monitor (`ctrl+shift+e` in Firefox)
1. Sort the `File` column alphabetically and look for `solr_query.php?search=<whatever>`
1. Copy the query URL for use in the `QUERY_URL` setting (e.g. `https://static.runningwarehouse.com/solr/solr_query.php?search=products&brand_str%5B%5D=Altra&facet_base_MSTFIL%5B%5D=facet_value_MTRAILCOND&filter_cat=SALEMS&filter_set=MSFILTER`)

### Configuring .env
Create a `.env` file in the application's working directory that looks something like:
```sh
QUERY_URL=your_query
THRESHOLD_PRICE=some_dollar_value
RECIPIENT_EMAIL=to_whatever_email_address
FROM_GMAIL=from_gmail_address
FROM_GMAIL_APP_PASSWORD=whatever_app_password
```

For example:
```sh
QUERY_URL=https://static.runningwarehouse.com/solr/solr_query.php?search=products&brand_str%5B%5D=Altra&facet_base_MSTFIL%5B%5D=facet_value_MTRAILCOND&filter_cat=SALEMS&filter_set=MSFILTER
THRESHOLD_PRICE=75
RECIPIENT_EMAIL=my_address@gmail.com
FROM_GMAIL=perhaps_the_same_email@gmail.com
FROM_GMAIL_APP_PASSWORD=whatever_app_password
```

## Installation
Get the source and compile it:
```
$ git clone git@github.com:scottbarnes/shoe-price-checker`  # or `git clone https://github.com/scottbarnes/shoe-price-checker`
$ cd shoe-price-checker
$ go build -buildvcs=false
```

Then fill in `.env` with some settings (see above).

Then run the script from cron somehow. E.g.:
`crontab -e`
```
0 7 * * * cd ~/code/shoe-price-checker && ./shoe-price-checker
```

Note, don't forget to `cd` to the directory with the binary and `.env`.

## Misc
- See [Sign in with App Passwords](https://support.google.com/accounts/answer/185833?hl=en) for how to create an app password for a Gmail account.
