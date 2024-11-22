# GoFOFA

[![Latest release](https://img.shields.io/github/v/release/FofaInfo/GoFOFA)](https://github.com/FofaInfo/GoFOFA/releases/latest)![GitHub Release Date](https://img.shields.io/github/release-date/FofaInfo/GoFOFA)![GitHub All Releases](https://img.shields.io/github/downloads/FofaInfo/GoFOFA/total)[![GitHub issues](https://img.shields.io/github/issues/FofaInfo/GoFOFA)](https://github.com/FofaInfo/GoFOFA/issues)

[:blue_book: 中文 README](https://github.com/FofaInfo/GoFOFA/blob/370e4ff07e2aaaacdbca868585870893a358a2aa/README_ZH.md)   |   [:floppy_disk: Download]()   |   [:orange_book: FOFA API Documentation](https://en.fofa.info/api)

## Background

GoFOFA is a command-line FOFA query tool written in Go. Besides having basic FOFA API interface calling capabilities, it can also directly process data further. Through modular invocation, it allows for the transformation of data from metadata to value data.

We have integrated many small features for querying and data processing. If you have more ideas and needs, feel free to submit them in the issues.

For any questions about GoFOFA, welcome to join our FOFA community [WeChat Group](https://github.com/FofaInfo/GoFOFA/blob/74544c05a4fdd2267da35d73a7833a03f875b75e/Resource/wechat%20QRScan.jpg) or [Telegram](https://t.me/+-5xC1wYcwollYWQ1) for technical exchanges.

## Content

### Configuration

### Query Module

- [Basic Query](#basic-query)

- [Query Function Module](#Query-Function-Module)
  - [Batch Search (supports uploading a txt file for batch queries and input from a pipe)](#batch-search)
  - [URL Concatenation](#URL-Concatenation)
  - [Randomly Generate Data from FOFA](#randomly-generate-data-from-fofa)
  - [Discovery Domains via Certificates](#Discovery-Domains-via-Certificates)
  - [Favicon Discovery](#Favicon-Discovery)
  - [Download Large Data](#Download-Large-Data)

- [Statistic Aggregation](#Statistic-Aggregation)
- [HOST Aggregation](#HOST-Aggregation)

### Data Processing Module

- [URL Deduplication](#url-deduplication)
- [Wildcard DNS Deduplication](#Wildcard-DNS-Deduplication)
- [Web Liveness Detection (supports input from a pipe in bulk)](#alive-check-supports-input-from-a-pipe-in-bulk)
- [JS Rendering (supports input from a pipe in bulk)](#JS-Rendering)
- [Data Classification](#Data-Classification)

Note: Some data processing functions have necessary field requirements. Please ensure that the fields are included when retrieving data.

### Others

- [GoFOFA Version](#gofofa-version)
- [GoFOFA Parameters List](#GoFOFA-Parameters-List)

## Configuration

- Download GoFOFA:

```
$ go install github.com/LubyRuffy/gofofa/cmd/fofa@latest
```

- If the following is displayed, it indicates a successful installation:

```
$ fofa
NAME:
   fofa - fofa client on Go v0.2.25, commit none, built at unknown

USAGE:
   fofa [global options] command [command options] [arguments...]

VERSION:
   v0.2.25

AUTHOR:
   LubyRuffy <lubyruffy@gmail.com>

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
   --fofaURL value, -u value  format: <url>/?key=<key>&version=<v2> (default: "https://fofa.info/?key=your_key&version=v1")
   --verbose                  print more information (default: false)
   --accountDebug             print account in error log (default: false)
   --help, -h                 show help (default: false)
   --version, -v              print the version (default: false)
```

- Configure environment variables:

```
$ export FOFA_CLIENT_URL='https://fofa.info/?key=your_key'
```
Or:
```
$ export FOFA_KEY='your_key'
```
Note: FOFA_CLIENT_URL has the highest priority.

#### MacOS/Linux Version

You can unzip the downloaded GoFOFA compressed package and place it in the `/usr/local/bin/` directory, allowing you to use the command from any location in the terminal.

```text
tar -zxvf ~/Downloads/fofa_0.2.25_darwin_amd64.tar.gz -C /usr/local/bin/
```

#### Windows Version

Extract the compressed package and directly run fofa.exe.

## Function Introduction

### Data Query Module

#### Basic Query

Use FOFA syntax for queries. You can enter a single query statement, and the default return fields are IP and port.

```shell
$ fofa search port=80
2024/08/23 11:51:19 query fofa of: port=80
69.10.146.92,80
20.193.138.22,80
194.182.72.64,80
......
......
```

You can also enter combined syntax for queries.

```shell
$ fofa search 'port=80 && protocol=ftp'
2024/08/23 11:52:00 query fofa of: port=80 && protocol=ftp
139.196.102.155,80
59.82.133.71,80
69.80.101.32,80
69.80.101.68,80
......
......
```

If you do not select a submodule for the query, it will default to using the search module:

```shell
$ fofa domain=qq.com
2024/08/23 11:53:00 query fofa of: domain=qq.com
14.22.33.13,443
183.47.126.116,443
14.22.33.13,443
14.22.33.13,443
......
......
```

Use `--fields` to select the output fields, which will be returned based on the selected fields. The following example selects the `host, ip, port, protocol, lastupdatetime` fields.

```shell
$ fofa search --fields host,ip,port,protocol,lastupdatetime 'port=6379'
2024/08/23 12:09:08 query fofa of: port=6379
168.119.197.62:6379,168.119.197.62,6379,redis,2024-08-23 12:00:00
119.45.170.222:6379,119.45.170.222,6379,redis,2024-08-23 12:00:00
112.126.87.29:6379,112.126.87.29,6379,unknown,2024-08-23 12:00:00
121.43.116.245:6379,121.43.116.245,6379,unknown,2024-08-23 12:00:00
......
......
```

Or more concisely, use `-f` to represent the fields.

```shell
$ fofa search -f host,ip,port,protocol,lastupdatetime 'port=6379'
2024/08/23 12:09:08 query fofa of: port=6379
168.119.197.62:6379,168.119.197.62,6379,redis,2024-08-23 12:00:00
119.45.170.222:6379,119.45.170.222,6379,redis,2024-08-23 12:00:00
112.126.87.29:6379,112.126.87.29,6379,unknown,2024-08-23 12:00:00
121.43.116.245:6379,121.43.116.245,6379,unknown,2024-08-23 12:00:00
......
......
```

Use `--size` to select the number of data entries returned per query, with a default size of 100:

```shell
$ fofa search --size 5 'port=6379'
2024/08/23 14:07:18 query fofa of: port=6379
47.99.89.216,6379
112.124.14.11,6379
107.154.224.11,6379
39.101.36.243,6379
139.196.136.107,6379
```

Or more concisely, use `-s` to represent the size.

```shell
$ fofa search -s 5 'port=6379'
2024/08/23 14:07:18 query fofa of: port=6379
47.99.89.216,6379
112.124.14.11,6379
107.154.224.11,6379
39.101.36.243,6379
139.196.136.107,6379
```

If you need to output data in a different format, you can set it using `--format`, with the default being `csv`. It also supports `json` and `xml` formats:

```shell
$ fofa search --format=json 'port=6379'
2024/08/23 14:05:49 query fofa of: port=6379
{"ip":"39.101.36.243","port":"6379"}
{"ip":"139.196.136.107","port":"6379"}
{"ip":"47.97.53.84","port":"6379"}
{"ip":"39.104.71.245","port":"6379"}
......
......
$ fofa search --format=xml 'port=6379'
2024/08/23 14:08:19 query fofa of: port=6379
<result><port>6379</port><ip>39.101.36.96</ip></result>
<result><ip>47.99.89.216</ip><port>6379</port></result>
<result><ip>112.124.14.11</ip><port>6379</port></result>
<result><ip>23.224.60.162</ip><port>6379</port></result>
......
......
```

Alternatively, you can use `--format json` or `--format xml` directly for the switch.

```shell
$ fofa search --format json 'port=6379'
2024/08/23 14:05:49 query fofa of: port=6379
{"ip":"39.101.36.243","port":"6379"}
{"ip":"139.196.136.107","port":"6379"}
{"ip":"47.97.53.84","port":"6379"}
{"ip":"39.104.71.245","port":"6379"}
......
......
$ fofa search --format xml 'port=6379'
2024/08/23 14:08:19 query fofa of: port=6379
<result><port>6379</port><ip>39.101.36.96</ip></result>
<result><ip>47.99.89.216</ip><port>6379</port></result>
<result><ip>112.124.14.11</ip><port>6379</port></result>
<result><ip>23.224.60.162</ip><port>6379</port></result>
......
......
```

Using `--outFile` allows you to output results to a specified file. If this parameter is not set, the output will default to the command line:

```shell
$ fofa search --outFile a.txt 'port=6379'
```

Or you can use `-o` to specify the output file directly.

```shell
$ fofa search -o a.txt 'port=6379'
```

The account information query interface allows you to retrieve account details by using `account`.

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

To count the number of FOFA query results, you can use the `count` module to statistically analyze the data quantity.

```shell
$ fofa count port=80
587055296
```
### Query Function Module
#### Batch Search

#### URL Concatenation

1. If you want to obtain a complete URL assembly, you can use the `--fixUrl` parameter:

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

2. If you want to use a different prefix, you can set the prefix using the `--urlPrefix` parameter:

```shell
$ fofa --size 1 --fields "host" --fixUrl --urlPrefix "redis://" protocol=redis
2024/08/23 14:29:26 query fofa of: protocol=redis
redis://139.9.222.14:7000
```

#### Randomly Generate Data from FOFA

Use the `random` command to randomly generate data from FOFA:

```shell
$ fofa random -f host,ip,port,lastupdatetime,title,header,body --format json
{"body":"","header":"HTTP/1.1 401 Unauthorized\r\nWww-Authenticate: Digest realm=\"IgdAuthentication\", domain=\"/\", nonce=\"ZjVhNGY2YzI6MTUyNDM2N2Y6MzRiMGZjZjQ=\", qop=\"auth\", algorithm=MD5\r\nContent-Length: 0\r\n","host":"95.22.200.127:7547","ip":"95.22.200.127","lastupdatetime":"2024-08-14 13:00:00","port":"7547","title":""}
```

You can set the interval to 500ms and generate data every 500ms using the `--sleep` parameter:

```shell
$ fofa random -s -1 --sleep 500
```

#### Discovery Domains via Certificates

The `domains` submodule is primarily used for the simplest form of expansion, leveraging certificates. You can use `--withCount` to count and obtain more data.

Required FOFA fields for this function: `certs_domains, certs_subject_org`

```shell
$ fofa domains -s 1000 --withCount baidu.com
baidu.com       660
dwz.cn  620
dlnel.com       614
bcehost.com     614
bdstatic.com    614
......
......
```

You can also use `--uniqByIP` to remove duplicate IPs:

```shell
$ fofa domains -s 1000 --withCount --uniqByIP baidu.com 
baidu.com       448
dwz.cn  410
aipage.cn       406
```

#### Favicon Discovery

Favicon icon query and hash value calculation.

Three direct query methods:
1. You can query data by reading a local ico file, and the `--open` parameter will automatically redirect you to FOFA:

```shell
$ fofa icon --open ./data/favicon.ico
```

2. You can also query through the ico file of a webpage:

```shell
$ fofa icon --open https://fofa.info/favicon.ico
```

3. You can directly query through a URL:

```shell
$ fofa icon --open http://www.baidu.com
```

Three methods to calculate the icon hash value:

1. Obtain the hash value of a local ico file:

```shell
$ fofa icon ./data/favicon.ico
-247388890
```

2. Obtain the hash value of a webpage's ico file:

```shell
$ fofa icon https://fofa.info/favicon.ico
-247388890
```

3. Directly obtain the `ico_hash` value of a URL:

```shell
$ fofa icon http://www.baidu.com
-1588080585
```

#### Download Large Data

For downloading large amounts of data, use the `--batchSize` parameter to set the download quantity and complete data download and storage into a specified file with one click:

```shell
$ fofa dump --format json --fixUrl --outFile a.json --batchSize 10000 'title=phpinfo'
```

Use a FOFA query file to download and store large data (one query per line):

```shell
cat queries.txt
port=13344
port=23455

# csv
$ fofa dump --outFile out.csv --inFile queries.txt

# json
$ fofa dump --inFile queries.txt --outFile out.json -j
2023/08/09 10:05:33 dump data of query: port=13344
2023/08/09 10:05:35 size: 11/11, 100.00%
2023/08/09 10:05:35 dump data of query: port=23455
2023/08/09 10:05:37 size: 499/499, 100.00%
```

### Statistic Aggregation

Invoke the data statistics interface. The `stats` module can perform data statistics and other operations.

```shell
$ fofa stats --fields title,country title="hacked by"
===  title
Hacked By Ashiyane Digital Security Team        706
Hacked By MR.GREEN      465
Hacked by Kn1gh7        259
Hacked By MR.GREEN &#8211; Just another WordPress site  163
HackeD By Desert Warriors       108
===  country
United States of America    3182
Germany    259
Poland    225
United Kingdom    223
Singapore  205
```

### HOST Aggregation

Invoke the HOST aggregation interface. By using the Host module and inputting a domain name, you can obtain asset information from the host perspective:

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
### Data Processing Module

#### URL Deduplication

Data deduplication can be conducted in two ways.

1. The first method performs deduplication while querying, meaning the data pulled from FOFA is already deduplicated. Use `--dedupHost`, where in FOFA, `subdomain` represents webpage data and `service` represents protocol data. If the `host` is the same, `subdomain` data is prioritized for retention:

```shell
$ fofa search -s 3 -f host,type "ip=106.75.95.206"
2024/08/28 19:49:23 query fofa of: ip=106.75.95.206
106.75.95.206,subdomain
106.75.95.206:443,service
106.75.95.206,service
$ fofa search -s 3 -f host,type --dedupHost "ip=106.75.95.206"
2024/08/28 19:52:30 query fofa of: ip=106.75.95.206
https://106.75.95.206,subdomain
106.75.95.206:443,service
106.75.95.206,subdomain
```

2. The second method supports uploading existing data in the form of a file to perform deduplication on any specified field. The `dedup` command allows deduplication of a specific field in a CSV file. Use the `input` parameter to upload the file, the `dedup` parameter to select the field for deduplication (deduplication is based on the order of fields), and the `output` parameter to set the output file name (default is `duplicate.csv`):

```shell
$ fofa dedup -output data.csv -dedup ip -output dedup.csv
$ fofa dedup -output data.csv -dedup ip,host,domain -output dedup.csv
```
Or more concisely:

```shell
$ fofa dedup -o data.csv -d ip -o dedup.csv
$ fofa dedup -o data.csv -d ip,host,domain -o dedup.csv
```

#### Wildcard DNS Deduplication

Required FOFA fields: ip, port, domain, title, fid

To reduce the number of wildcard domains, you can use `--deWildcard` to set the number of wildcard domains to retain. `-f` supports other fields, and `link` is used here as an example:

```shell
$ fofa search -s 3 -f link domain=huashunxinan.net
2024/08/27 17:19:04 query fofa of: domain=huashunxinan.net
http://h8huumr2zdmwgy5.huashunxinan.net
http://keygatjexlvsznh.huashunxinan.net
http://jobs.huashunxinan.net
$ fofa search -s 3 -f link --deWildcard 1 domain=huashunxinan.net
2024/08/27 17:26:42 query fofa of: domain=huashunxinan.net
http://h8huumr2zdmwgy5.huashunxinan.net
https://fwtn2k7oigaiyla.huashunxinan.net
http://huashunxinan.net
```

#### Liveness Detection (Supports Batch Input from Pipelines)

Liveness detection can be conducted in two ways.

1. The first mode requires the `link` field. It checks if the user's input includes the required field, adds it if not, and finally returns only the user's input fields with an additional `isActive` field.

Use `--checkActive 3`, where `3` is the number of retry attempts before timing out (this parameter also retrieves the `status_code` data):

```shell
$ fofa search -s 3 --checkActive port=80 
2024/08/26 18:52:00 query fofa of: port=80
216.92.244.44,80,true
104.21.31.50,80,true
182.247.239.68,80,true
$ fofa search -s 3 --checkActive 3 --format=json port=80
2024/08/26 18:53:33 query fofa of: port=80
{"ip":"54.78.179.223","isActive":"false","port":"80"}
{"ip":"18.155.202.65","isActive":"true","port":"80"}
{"ip":"198.144.179.122","isActive":"true","port":"80"}
$ fofa search -s 3 --checkActive 3 --format=xml port=80
2024/08/26 18:54:38 query fofa of: port=80
<result><ip>50.16.35.210</ip><port>80</port><isActive>true</isActive></result>
<result><ip>104.21.77.62</ip><port>80</port><isActive>true</isActive></result>
<result><ip>189.193.236.170</ip><port>80</port><isActive>false</isActive></result>
```

2. The second mode supports inputting URLs from a pipeline or a file. Use `--url` to obtain liveness information, with `true` indicating alive and `false` indicating dead.

```shell
$ fofa active --url baidu.com,fofa.info,asdsadsasdas.com
baidu.com,true
fofa.info,true
asdsadsasdas.com,false
```

Or more concisely:

```shell
$ fofa active -u baidu.com,fofa.info,asdsadsasdas.com
baidu.com,true
fofa.info,true
asdsadsasdas.com,false
```

You can also probe a file with each line containing a URL:

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

It also supports probing URLs from a pipeline (each line in the pipeline should contain one URL):

```shell
$ fofa search -f link -s 3 port=80 | fofa active
2024/08/23 15:50:11 query fofa of: port=80
http://og823.hb-yj.com,true
http://rw823.tcxzgh.org,true
http://sb823.tcxzgh.org,true
```

#### JS Rendering

Required FOFA field: link. After completing the search task, separate rendering recognition is required.

The `jsRender` module is used to perform JS rendering on URLs, supporting the selection of rendered HTML tags. Currently, it supports retrieving the `title` and `body` tags. Use `-url` to select a single target and `-tag` to select the rendered tag to retrieve:

```shell
$ fofa jsRender -url http://baidu.com -tag title
http://baidu.com,百度一下，你就知道
```

Or more concisely:

```shell
$ fofa jsRender -u http://baidu.com -t title
http://baidu.com,百度一下，你就知道
```

You can also probe a file with each line containing a URL:

```shell
$ cat url.txt
http://baidu.com
https://fofa.info
$ fofa jsRender -i url.txt -t title 
http://baidu.com,百度一下，你就知道
https://fofa.info,网络空间测绘，网络空间安全搜索引擎，网络空间搜索引擎，安全态势感知 - FOFA网络空间测绘系统
```

It also supports probing URLs from a pipeline (each line in the pipeline should contain one URL):

```shell
$ fofa search -f link -s 3 port=80 | fofa jsRender -t title
2024/08/23 15:50:11 query fofa of: port=80
http://project5.abioyibo.com,Just another WordPress site
http://www.valuegoodsbazaar.shop,srv258.sellvir.com — Coming Soon
http://forecasting-preprod.pcasys.co.uk,- Sales Forecasting Tool (preprod)
```

### Data Classification

The `fofa category` command allows you to process CSV file data using classification rules defined in the `config.yaml` file. You can set multiple classification rules based on protocol, title, domain, etc. For example, the following `config.yaml` file defines several classification rules:

```shell
categories:
  - name: "百度贴吧"
    filters:
      - "(protocol == 'http' || protocol == 'http') && CONTAIN(title, '百度贴吧')"
      - "domain=='baidu.com' && CONTAIN(title, '百度贴吧')"

  - name: "百度3xx页面"
    filters:
      - "domain=='baidu.com' && status_code >= '300' && status_code < '400'"

  - name: "其他"
    filters:
      - "CONTAIN(title, '百度')"

```

You can use the following command to classify the input CSV file based on these rules:

```shell
$ fofa category -input input.csv [-output category.csv]
```

Or use the more concise command:

```shell
$ fofa category -i input.csv [-o category.csv]
```

### Others

Get GoFOFA Version

```shell
$ fofa --version
```

### GoFOFA Parameters List

| Parameter   | Abbreviation | Default | Description                                           |
| ----------- | ------------ | ------- | ----------------------------------------------------- |
| fields      | f            | ip,port | Retrieve fofa fields                                  |
| format      |              | csv     | Output format, can be csv/json/xml                     |
| outFile     | o            |         | Output file, if not set, it will print to terminal     |
| size        | s            | 100     | Number of results, maximum of 10000, limited by deductMode |
| deductMode  |              |         | Consume f points, if not set, it reads the maximum free quota |
| fixUrl      |              | false   | Whether to combine the URL, e.g., 1.1.1.1,80 becomes http://1.1.1.1 |
| urlPrefix   |              | http:// | URL prefix                                            |
| full        |              | false   | Whether to retrieve full data                          |
| uniqByIP    |              | false   | Whether to deduplicate by IP                          |
| workers     |              | 10      | Number of threads                                      |
| rate        |              | 2       | Queries per second                                    |
| template    |              | ip={}   | Input from the pipeline, input content will replace `{}` |
| inFile      | i            |         | Input file, if not set, it reads from the pipeline    |
| checkActive |              | -1      | Probe retry times, -1 means no probing                |
| deWildcard  |              | -1      | Wildcard deduplication, -1 means no wildcard deduplication |
| filter      |              |         | Data filtering rule, e.g., `port<100 || host=="baidu.com"` |
| dedupHost   |              | false   | Subdomain deduplication                                |
| headline    |              | false   | Whether to output CSV header, only valid when format is CSV |
| help        | h            | false   | Usage help                                             |

