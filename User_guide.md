# GoFOFA User Guide

## Table of Contents
### Configuration
### Data Query Module

- [Basic Queries](#Basic-Queries)

- [Utility Features for Queries](#Utility-Features-for-Queries)
	- [Batch Search (supports bulk queries via txt file upload)](#Batch-Search)
	- [Specify URL Concatenation](#URL-Concatenation)
	- [Random Data Generation from FOFA](#Random-Data-Generation)
	- [Certificate Line Expansion to Obtain Domains](#Certificate-Line-Expansion)
	- [Favicon Icon Queries](#Favicon-Icon-Queries)
	- [Large Dataset Downloads](#Large-Dataset-Downloads)
- [Statistical Aggregation API](#Statistical-Aggregation-API)
- [HOST Aggregation API](#HOST-Aggregation-API)

### Data Processing Module

- [IP Deduplication](#IP-Deduplication)
- [URL Deduplication](#URL-Deduplication)
- [Wildcard Deduplication](#Wildcard-Deduplication)
- [Liveness Detection (supports bulk input via pipeline)](#Liveness-Detection)
- [JS Rendering Recognition (supports bulk input via pipeline)](#JS-Rendering-Recognition)
- [Data Classification](#Data-Classification)

Note: Certain data processing functions require specific fields; ensure the required fields are included during data retrieval.

### Other Features

- [Gofofa Version](#Other-Features)
- [GoFOFA Parameter list](#GoFOFA-Parameter-list)


## Configuration

- Download GoFOFA:

```shell
$ go install github.com/FofaInfo/GoFOFA/cmd/fofa@latest
```

- A successful installation will display the following:

```shell
$ fofa
NAME:
   fofa - fofa client on Go v0.2.26, commit none, built at unknown

USAGE:
   fofa [global options] command [command options] [arguments...]

VERSION:
   v0.2.26

AUTHOR:
   LubyRuffy <lubyruffy@gmail.com>
   Y13ze <y13ze@outlook.com>

COMMANDS:
   search   fofa host search
   account  fofa account information
   count    fofa query results count
   stats    fofa stats
   icon     fofa icon search
   random   fofa random data generator
   host     fofa host
   dump     fofa dump data
   domains  extend domains from a domain
   active   website active
   dedup    remove duplicate tool
   category  classify data according to config
   jsRender  website js render
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --fofaURL value, -u value  format: <url>/?email=&key=<key>&version=<v2> (default: "https://fofa.info/?key=your_key&version=v1")
   --verbose                  print more information (default: false)
   --accountDebug             print account in error log (default: false)
   --help, -h                 show help (default: false)
   --version, -v              print the version (default: false)
```

- Configure environment variables:

```shell
$ export FOFA_CLIENT_URL='https://fofa.info/?key=your_key'
```

#### MacOS/Linux Version Download

You can extract the GoFOFA archive and place it in the `/usr/local/bin/` directory, enabling usage of the command from any terminal location.

```shell
tar -zxvf ~/Downloads/fofa_0.2.26_darwin_amd64.tar.gz -C /usr/local/bin/
```

Note: If you encounter a "permission denied" error, prepend `sudo` to the command. Ensure you adjust the filename to match the downloaded version.

#### Windows Version Download

Extract the archive and directly run `fofa.exe`.

---

## Features

### Data Query Module

#### Basic Queries

**FOFA Syntax Search:** You can perform queries using single or combined syntax. If no return fields are specified, the default fields are `ip` and `port`. Use the `search` command; if no command is provided, it defaults to query mode.

```shell
fofa search 'port=80 && protocol=ftp'
2024/08/23 11:52:00 query fofa of: port=80 && protocol=ftp
139.196.102.155,80
59.82.133.71,80
69.80.101.32,80
69.80.101.68,80
......
```

**Field Selection:** Use the `-fields` parameter to specify the fields to return. This will output only the specified fields. Both `-fields` and its shorthand `-f` can be used.

```shell
$ fofa search -fields host,ip,port,protocol,lastupdatetime 'port=6379'
2024/08/23 12:09:08 query fofa of: port=6379
168.119.197.62:6379,168.119.197.62,6379,redis,2024-08-23 12:00:00
119.45.170.222:6379,119.45.170.222,6379,redis,2024-08-23 12:00:00
112.126.87.29:6379,112.126.87.29,6379,unknown,2024-08-23 12:00:00
121.43.116.245:6379,121.43.116.245,6379,unknown,2024-08-23 12:00:00
```

**Result Count:** Use the `-size` parameter to set the number of results per query. The default is 100. Both `-size` and its shorthand `-s` can be used.

```shell
$ fofa search -size 5 'port=6379'
2024/08/23 14:07:18 query fofa of: port=6379
47.99.89.216,6379
112.124.14.11,6379
107.154.224.11,6379
39.101.36.243,6379
139.196.136.107,6379
```

**Output Format:** To specify a different output format, use the `-format` parameter. The default is `csv`. Supported formats include `json`, `xml`, and `txt`.

```shell
$ fofa search -format json 'port=6379'
2024/08/23 14:05:49 query fofa of: port=6379
{"ip":"39.101.36.243","port":"6379"}
{"ip":"139.196.136.107","port":"6379"}
{"ip":"47.97.53.84","port":"6379"}
{"ip":"39.104.71.245","port":"6379"}
```

**Data Export:** Use the `-outFile` parameter to export results to a specified file. If not set, results are printed in the terminal. Both `-outFile` and its shorthand `-o` can be used.

```shell
$ fofa search -outFile a.txt 'port=6379'
```

**Account Information Query:** Use the `account` command to retrieve account details.

```shell
$ fofa account
{
  "error": false,
  "fcoin": 0,
  "fofa_point": 99982,
  "isvip": true,
  "vip_level": 5,
  "remain_api_query": 4999635, 
  "remain_api_data": 49949766
}
```

**FOFA Query Result Count:** Use the `count` module to check the number of results for a query.

```shell
$ fofa count port=80
587055296
```

### Utility Features for Queries

#### Batch Search

Use the `-batchType` parameter in the `dump` module to enable batch search. Currently supported fields for batch generation include `ip` and `domain`.

For `ip` batch search, the IPs in the file are grouped into sets of 100 and used to generate batch query statements. Similarly, for `domain`, the domains are grouped into sets of 50.

For example, if the file contains 1,000 IPs, the tool will automatically generate 10 sets of query statements to complete the batch search.

Both `-batchType` and its shorthand `-bt` can be used.

```shell
$ cat ip.txt
106.75.95.206
123.58.224.8
$ fofa dump -i ip.txt --batchType ip
2024/11/25 14:51:10 dump data of query: ip=106.75.95.206 || ip=123.58.224.8
123.58.224.8,40544
123.58.224.8,31497
106.75.95.206,80
......
```

Typically, results are output to a file, and the progress of the export is displayed:

```shell
$ fofa dump -i ip.txt -bt ip -o dump.csv
2024/11/25 14:51:10 dump data of query: ip=106.75.95.206 || ip=123.58.224.8
2024/11/25 14:52:03 size: 188/188, 100.00%
......
```

Use the `-batchSize` parameter to set the number of records fetched per batch. The default is 1,000. For instance, if a batch query result contains 20,000 records and the default fetch size is 1,000, the tool will perform 20 fetches for that batch before proceeding to the next query statement.

Both `-batchSize` and its shorthand `-bs` can be used.

Use the `-size` parameter to set the total number of records to fetch per batch. The default is `-1`, which fetches all available data.

```shell
$ fofa dump -i ip.txt -bt ip -o dump.csv
2024/11/25 17:39:05 dump data of query: ip=112.25.151.122 || ... || ip=58.213.160.221
2024/11/25 17:39:06 size: 115/115, 100.00%
2024/11/25 17:39:06 dump data of query: ip=221.226.119.3 || ... || ip=221.226.6.2
2024/11/25 17:39:12 size: 153/153, 100.00%
```

#### URL Concatenation

1. If you want to retrieve fully concatenated URLs, use the `fixUrl` parameter:

```shell
$ fofa --size 2 --fields "host" title=Gitblit
2024/08/23 14:23:02 query fofa of: title=Gitblit
pmsningbo.veritrans.cn:20202
platform.starpost.cn:8080

$ fofa --size 2 --fields "host" --fixUrl title=Gitblit
2024/08/23 14:23:34 query fofa of: title=Gitblit
http://pmsningbo.veritrans.cn:20202
http://platform.starpost.cn:8080
```

2. If you want to use a custom prefix, use the `urlPrefix` parameter to set the prefix:

```shell
$ fofa --size 1 --fields "host" --fixUrl --urlPrefix "redis://" protocol=redis
2024/08/23 14:29:26 query fofa of: protocol=redis
redis://139.9.222.14:7000
```

#### Random Data Generation from FOFA

Use the `random` command to generate random data from FOFA:

```shell
$ fofa random -f host,ip,port,lastupdatetime,title,header,body --format json
{"body":"","header":"HTTP/1.1 401 Unauthorized\r\nWww-Authenticate: Digest realm=\"IgdAuthentication\", domain=\"/\", nonce=\"ZjVhNGY2YzI6MTUyNDM2N2Y6MzRiMGZjZjQ=\", qop=\"auth\", algorithm=MD5\r\nContent-Length: 0\r\n","host":"95.22.200.127:7547","ip":"95.22.200.127","lastupdatetime":"2024-08-14 13:00:00","port":"7547","title":""}
```

You can set the interval to 500ms between data generations using the `sleep` parameter:

```shell
$ fofa random -s -1 -sleep 500
```

#### Certificate Line Expansion to Obtain Domains

The `domains` submodule is used for simple line expansions through certificates. Use the `withCount` parameter to count the number of results for a query and retrieve additional data.

Required FOFA fields: `certs_domains`, `certs_subject_org`.

```shell
$ fofa domains -s 1000 -withCount baidu.com
baidu.com       660
dwz.cn          620
dlnel.com       614
bcehost.com     614
bdstatic.com    614
......
```

#### Favicon Icon Queries

Favicon icon queries and hash value calculations can be performed in the following ways:

1. Query by reading a local `.ico` file. The `open` parameter automatically redirects to the FOFA search page:

```shell
$ fofa icon --open ./data/favicon.ico
```

2. Query via a website's `.ico` file:

```shell
$ fofa icon --open https://fofa.info/favicon.ico
```

3. Query directly via a URL:

```shell
$ fofa icon --open http://www.baidu.com
```

To calculate icon hash values:

1. Retrieve the hash value of a local `.ico` file:

```shell
$ fofa icon ./data/favicon.ico
-247388890
```

2. Retrieve the hash value of a website's `.ico` file:

```shell
$ fofa icon https://fofa.info/favicon.ico
-247388890
```

3. Retrieve the hash value of a URL's `.ico`:

```shell
$ fofa icon http://www.baidu.com
-1588080585
```

#### Large Dataset Downloads

Use the `--batchSize` parameter to set the number of records per download batch and automatically store data in a specified file:

```shell
$ fofa dump --format json -fixUrl -outFile a.json -batchSize 500 'title=phpinfo'
```

Download and store large datasets using FOFA query statements stored in a file, where each line represents a query:

```shell
cat queries.txt
port=13344
port=23455

# csv
$ fofa dump -outFile out.csv -inFile queries.txt

# json
$ fofa dump -inFile queries.txt -outFile out.json -j
2023/08/09 10:05:33 dump data of query: port=13344
2023/08/09 10:05:35 size: 11/11, 100.00%
2023/08/09 10:05:35 dump data of query: port=23455
2023/08/09 10:05:37 size: 499/499, 100.00%
```


### Statistical Aggregation API

The `stats` module allows for data aggregation and statistical analysis.

```shell
$ fofa stats --fields title,country title="hacked by"
===  title
Hacked By Ashiyane Digital Security Team        706
Hacked By MR.GREEN      465
Hacked by Kn1gh7        259
Hacked By MR.GREEN &#8211; Just another WordPress site  163
HackeD By Desert Warriors       108
===  country
United States    3182
Germany          259
Poland           225
United Kingdom   223
Singapore        205
```

---

### HOST Aggregation API

The `host` module retrieves asset information from a host perspective by inputting a domain:

```shell
$ fofa host demo.cpanel.net
Host:            demo.cpanel.net
IP:              208.74.120.133
ASN:             33522
ORG:             CPANEL-INC
Country:         United States of America
CountryCode:     US
Ports:           [2078 3306 2079 2082 143 993 2086 2095 2083 2087 110 2080 80 995 2096 2077 443]
Protocols:       imaps,mysql,https,imap,pop3s,http,pop3
Categories:      Server Management
Products:        cPanel-MGMT-Products
UpdateTime:      2022-05-30 17:00:00
```

---

### Data Processing Module

#### IP Deduplication

By default, the tool retains one record per IP. Use the `-uniqByIP` command to remove duplicate entries for the same IP.

```shell
fofa search -uniqByIP -s 30 port=80
2024/12/06 17:03:09 query fofa of: port=80
161.156.173.134,80
104.21.0.253,80
31.214.178.70,80
104.21.44.49,80
...
...
```

#### URL Deduplication

Data deduplication can be performed in two ways:

1. **Inline Deduplication During Querying:**  
   This deduplication occurs as data is retrieved from FOFA. Use the `-dedupHost` parameter to retain unique `host` entries. In FOFA, `subdomain` represents web data and `service` represents protocol data. If the `host` field matches, `subdomain` data is given priority.

```shell
$ fofa search -s 3 -f host,type --dedupHost "ip=106.75.95.206"
2024/08/28 19:52:30 query fofa of: ip=106.75.95.206
https://106.75.95.206,subdomain
106.75.95.206:443,service
106.75.95.206,subdomain
```

2. **File-Based Deduplication:**  
   This method uses an input file to deduplicate any field in existing data. The `dedup` command deduplicates a specified field in a CSV file. Use the `input` parameter to specify the file, the `dedup` parameter to choose the field(s) for deduplication, and the `output` parameter to specify the output filename (defaults to `duplicate.csv`).

```shell
$ fofa dedup -output data.csv -dedup ip -output dedup.csv
$ fofa dedup -output data.csv -dedup ip,host,domain -output dedup.csv
```

---

#### Wildcard Deduplication

Required FOFA fields: `ip`, `port`, `domain`, `title`, `fid`.

To reduce the number of wildcard domain entries, use the `--deWildcard` parameter to set the number of wildcard domains to retain. Use the `-f` parameter to specify other fields, such as `link`:

```shell
$ fofa search -s 3 -f link --deWildcard 1 domain=huashunxinan.net
2024/08/27 17:26:42 query fofa of: domain=huashunxinan.net
http://h8huumr2zdmwgy5.huashunxinan.net
https://fwtn2k7oigaiyla.huashunxinan.net
http://huashunxinan.net
```

#### Liveness Detection (Supports Bulk Input via Pipeline)

Liveness detection can be performed in two ways:

1. **Inline Liveness Detection:**  
   This mode requires the `link` field. If the user input does not include the necessary fields, they will be automatically added. The returned data will only include the user-specified fields, with an additional `isActive` field indicating liveness.  
   Use the `--checkActive 3` parameter, where `3` represents the number of retries in case of timeout (this parameter also refreshes the `status_code` data).

```shell
$ fofa search -s 3 --checkActive 3 --format=json port=80
2024/08/26 18:53:33 query fofa of: port=80
{"ip":"54.78.179.223","isActive":"false","port":"80"}
{"ip":"18.155.202.65","isActive":"true","port":"80"}
{"ip":"198.144.179.122","isActive":"true","port":"80"}
```

2. **Pipeline or File-Based Input:**  
   Use the `--url` parameter to input URLs for liveness detection. The output will indicate `true` for live URLs and `false` for inactive ones.

```shell
$ fofa active --url baidu.com
baidu.com,true
```

Liveness detection can also be performed on URLs listed in a file:

```shell
$ cat target.txt
baidu.com
fofa.info
asdsadsasdas.com
$ fofa active -i target.txt  
baidu.com,true
fofa.info,true
asdsadsasdas.com,false
```

Alternatively, URLs from pipeline input can be checked (one URL per line):

```shell
$ fofa search -f link -s 3 port=80 | fofa active
2024/08/23 15:50:11 query fofa of: port=80
http://og823.hb-yj.com,true
http://rw823.tcxzgh.org,true
http://sb823.tcxzgh.org,true
```

---

#### JS Rendering Recognition (Supports Bulk Input via Pipeline)

Required FOFA fields: `link`.  
After completing a `search` task, you can perform standalone JS rendering recognition.

The `jsRender` module processes URLs for JS rendering and allows retrieval of rendered HTML tags. Currently supported tags include `title` and `body`. Use the `-url` parameter to specify a single target and the `-tags` parameter to select the tags to retrieve.

```shell
$ fofa jsRender -url http://baidu.com -tags title
http://baidu.com,Baidu Search
```

This functionality also supports files containing a list of URLs (one URL per line):

```shell
$ cat url.txt
http://baidu.com
https://fofa.info
$ fofa jsRender -i url.txt -t title 
http://baidu.com,Baidu Search
https://fofa.info,Cyberspace Mapping, Security Search Engine - FOFA
```

It can also process pipeline input (one URL per line):

```shell
$ fofa search -f link -s 3 port=80 | fofa jsRender -t title
2024/08/23 15:50:11 query fofa of: port=80
http://project5.abioyibo.com,Just another WordPress site
http://www.valuegoodsbazaar.shop,srv258.sellvir.com — Coming Soon
http://forecasting-preprod.pcasys.co.uk,- Sales Forecasting Tool (preprod)


```


#### Data Classification

The `category` module supports classifying assets in a CSV file based on predefined rules in a `config.yaml` file (the configuration file must be located in the current directory). Below is an example configuration:

```yaml
categories:
  - name: "Baidu Tieba"
    filters:
      - "(protocol == 'http' || protocol == 'https') && CONTAIN(title, 'Baidu Tieba')"
      - "domain=='baidu.com' && CONTAIN(title, 'Baidu Tieba')"

  - name: "Baidu 3xx Pages"
    filters:
      - "domain=='baidu.com' && status_code >= '300' && status_code < '400'"

  - name: "Others"
    filters:
      - "CONTAIN(title, 'Baidu')"
```

You can set filtering rules in the `config.yaml` file using the `filter` option. A built-in `CONTAIN` method is available, which checks if a specific field contains a specified value. If the `-output` parameter is not set, the tool generates a default file named `category.csv`.

Usage:

```shell
$ fofa category -input input.csv [-output category.csv]
```

---

### Other Features

#### GoFOFA Version

To display the current version of GoFOFA:

```shell
$ fofa --version
```

---

## GoFOFA All Parameter list

### Search

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| fields       | f            | ip,port       | FOFA fields to retrieve. [Learn More](https://en.fofa.info/vip) |                             
| format       |              | csv           | Output format: csv/json/xml                              |
| outFile      | o            |               | Output file. If not set, prints to terminal              |
| size         | s            | 100           | Query size. Maximum is 10,000, limited by `deductMode`   |
| deductMode   |              |               | Determines consumption of f-points. Uses free quota by default |
| fixUrl       |              | false         | Concatenates URLs (e.g., `1.1.1.1,80` → `http://1.1.1.1`) |
| urlPrefix    |              | http://       | URL prefix                                               |
| full         |              | false         | Retrieves full data                                      |
| uniqByIP     |              | false         | Removes duplicates by IP                                |
| workers      |              | 10            | Number of threads                                        |
| rate         |              | 2             | Queries per second                                       |
| template     |              | ip={}         | Replaces `{}` with content from pipeline input          |
| inFile       | i            |               | Input file. If not set, reads from pipeline input        |
| checkActive  |              | -1            | Number of retries for liveness detection (-1 disables)  |
| deWildcard   |              | -1            | Number of wildcard entries to retain (-1 disables)      |
| filter       |              |               | Data filtering rules (e.g., `port<100 || host=="baidu.com"`) |
| dedupHost    |              | false         | Removes duplicates for subdomains                       |
| headline     |              | false         | Outputs CSV headers (only applicable for CSV format)    |
| help         | h            | false         | Displays usage instructions                              |

### Dump

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| fields       | f            | ip,port       | FOFA fields to retrieve. [Learn More](https://en.fofa.info/vip) |
| format       |              | csv           | Output format: csv/json/xml                              |
| outFile      | o            |               | Output file. If not set, prints to terminal              |
| inFile       | i            |               | Input file. If not set, reads from pipeline input        |
| size         | s            | 100           | Query size. No upper limit but consumes f-points or free quota |
| fixUrl       |              | false         | Concatenates URLs (e.g., `1.1.1.1,80` → `http://1.1.1.1`) |
| urlPrefix    |              | http://       | URL prefix                                               |
| full         |              | false         | Retrieves full data                                      |
| batchSize    | bs           | 1000          | Number of records fetched per batch                     |
| batchType    | bt           |               | Batch query type: ip/domain                             |
| help         | h            | false         | Displays usage instructions                              |

### jsRender

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| url          | u            |               | Single URL for rendering                                 |
| tags         | t            |               | Tags to retrieve (options: title/body)                  |
| format       |              | csv           | Output format: csv/json/xml                              |
| outFile      | o            |               | Output file. If not set, prints to terminal              |
| inFile       | i            |               | Input file. If not set, reads from pipeline input        |
| workers      |              | 2             | Number of threads                                        |
| retry        |              | 3             | Number of timeout retries                                |
| help         | h            | false         | Displays usage instructions                              |

### Domains

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| outFile      | o            |               | Output file. If not set, prints to terminal              |
| size         | s            | 100           | Query size. Maximum is 10,000, limited by `deductMode`   |
| deductMode   |              |               | Determines consumption of f-points. Uses free quota by default |
| full         |              | false         | Retrieves full data                                      |
| withCount    |              | false         | Outputs domain count                                     |
| clue         |              | false         | Outputs clue statements                                  |
| help         | h            | false         | Displays usage instructions                              |

### Active

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| url          | u            |               | Single URL liveness detection                           |
| format       |              | csv           | Output format: csv/json/xml                              |
| outFile      | o            |               | Output file. If not set, prints to terminal              |
| inFile       | i            |               | Input file. If not set, reads from pipeline input        |
| workers      |              | 2             | Number of threads                                        |
| retry        |              | 3             | Number of timeout retries                                |
| help         | h            | false         | Displays usage instructions                              |

### Category

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| inFile       | i            |               | Input classification file (CSV format)                  |
| unique       |              |               | Ensures unique classification data                      |
| help         | h            | false         | Displays usage instructions                              |

### Dedup

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| dedup        | d            |               | Field(s) to deduplicate                                  |
| inFile       | i            |               | Input file for deduplication (CSV format)               |
| outFile      | o            | duplicate.csv | Output file                                              |
| help         | h            | false         | Displays usage instructions                              |

### Host

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| help         | h            | false         | Displays usage instructions                              |

### Icon

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| open         |              | false         | Opens FOFA search page based on icon results            |
| help         | h            | false         | Displays usage instructions                              |

### Stats

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| fields       | f            | title,country | FOFA fields to retrieve. [Learn More](https://fofa.info/vip) |
| size         | s            | 5             | Number of queries. `-1` for infinite queries             |
| help         | h            | false         | Displays usage instructions                              |

### Random

| Parameter    | Abbreviation | Default Value                        | Description                                              |
|--------------|--------------|--------------------------------------|----------------------------------------------------------|
| fields       | f            | ip,port,host,header,title,server,lastupdatetime | FOFA fields to retrieve. [Learn More](https://fofa.info/vip) |
| format       |              | json                                | Output format: csv/json/xml                              |
| size         | s            | 1                                   | Number of queries. `-1` for infinite queries             |
| sleep        |              | 1000                                | Interval between queries in milliseconds                |
| fixUrl       |              | false                               | Concatenates URLs (e.g., `1.1.1.1,80` → `http://1.1.1.1`) |
| urlPrefix    |              | http://                             | URL prefix                                               |
| full         |              | false                               | Retrieves full data                                      |
| help         | h            | false                               | Displays usage instructions                              |

### Count

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| help         | h            | false         | Displays usage instructions                              |

### Account

| Parameter    | Abbreviation | Default Value | Description                                              |
|--------------|--------------|---------------|----------------------------------------------------------|
| help         | h            | false         | Displays usage instructions                              |

