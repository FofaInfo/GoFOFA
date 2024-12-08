# GoFOFA

[![Latest release](https://img.shields.io/github/v/release/FofaInfo/GoFOFA)](https://github.com/FofaInfo/GoFOFA/releases/latest)![GitHub Release Date](https://img.shields.io/github/release-date/FofaInfo/GoFOFA)![GitHub All Releases](https://img.shields.io/github/downloads/FofaInfo/GoFOFA/total)[![GitHub issues](https://img.shields.io/github/issues/FofaInfo/GoFOFA)](https://github.com/FofaInfo/GoFOFA/issues)

[中文 README](https://github.com/FofaInfo/GoFOFA/blob/main/README_ZH.md)   |  [User Guide]()   |  [Download](https://github.com/FofaInfo/GoFOFA/releases) 


## Project Background

GoFOFA is a command-line FOFA query tool written in Go. In addition to basic FOFA API calls, it can also process data further, enabling the transformation from raw data to business data through a modular invocation mechanism.

For any questions about GoFOFA, feel free to join our FOFA community via [WeChat Group](https://github.com/FofaInfo/GoFOFA/blob/74544c05a4fdd2267da35d73a7833a03f875b75e/Resource/wechat%20QRScan.jpg) or [Telegram](https://t.me/+-5xC1wYcwollYWQ1) for technical discussions.

## Installation
Run the following command in the terminal to install the latest version of GoFOFA:

```
go install github.com/FofaInfo/GoFOFA/cmd/fofa@latest
```
Configure the environment variable with your FOFA API key:

```
export FOFA_KEY='your_key'
```
Execute the test command `fofa search -s 1 ip=1.1.1.1`. If results are returned, the installation and configuration are successful.

Output example:
```
fofa search -s 1 ip=1.1.1.1
query fofa of: ip=1.1.1.1
1.1.1.1,8880
```

GoFOFA offers rich functionalities. For a complete user guide and installation instructions, [click here]().


## Key Features

### Batch Search and Download Module

In addition to basic API calls, the query module introduces many practical features, such as batch search, URL concatenation, obtaining domains from certificate extension, diverse icon-based queries, and large-scale data downloads.

Batch search addresses issues with querying large volumes of homogeneous data. Currently, it supports batch queries for IPs or domains.

The program can automatically process IPs in a file, concatenate statements for batch queries, and complete data downloads. For example, if a task involves 1,000 IPs, the program will concatenate them into groups of 100 queries (10 groups in total) and perform automated downloads, significantly reducing data retrieval time.

Example case:
1. A task file `ip.txt` containing 1,000 IPs;
2. Query results from the last year;
3. Retrieved fields include `ip`, `port`, `host`, `link`, `title`, and `domain`;
4. Automatically saved as a `data.csv` file.

```
fofa dump -i ip.txt -bt ip -bs 50000 -f ip,port,host,protocol,lastupdatetime -o data.csv
```

### Web Liveness Detection Module

Liveness detection is a common need after retrieving data. This module supports simultaneous data retrieval and liveness detection, returning results with an added `isActive` field and updating the `status_code` field.

Command parameter: `checkActive`, with a numerical value representing the number of retries in case of timeout.

Example case:
1. Requesting 100 results;
2. Selecting `json` format for output;
3. Setting 3 retries for timeouts.

```
fofa --checkActive 3 -s 100 --format=json port=80 && type=subdomain
```

### JS Rendering Module

The JS Rendering Module processes URLs in the retrieved data for JavaScript rendering, supporting the extraction of rendered HTML tags such as `title` and `body`.

Command parameters:  
- `jsRender` to activate the rendering module;  
- `-url` to specify a single target;  
- `-tags` to select the tags to retrieve after rendering.

:red_circle: **Note:** The `jsRender` module is resource-intensive; it's not recommended to arbitrarily modify the `workers` parameter.

Example case:
1. Requesting 10 results;
2. Rendering and recognizing URLs from the pipeline;
3. Adding the newly extracted `title` field.

```
fofa jsRender -tags title --format=json -i link.txt -o data.txt
```

### Data Classification Module

In the result output phase, GoFOFA supports categorizing CSV files based on key characteristics. This operation can be configured through the `config.yaml` file.

Users can set filtering rules in the `config.yaml` file beforehand. Using a built-in `contain` method, data can be processed and categorized.

Usage example:

```
fofa category -input input.csv [-o category.csv]
```

Configuration example:
1. `sheet1`: Websites with asset titles containing keywords such as "backend", "system", "login" , "management";
2. `sheet2`: Assets with titles containing "nginx," "tomcat," "IIS," or "Welcome to OpenResty";
3. `sheet3`: Assets with protocols other than HTTP/HTTPS;
4. `sheet4`: Assets whose `category` field includes labels like "router," "video surveillance," "network storage," or "firewall."

```yaml
categories:  
  - name: "sheet1"    
    filters:    
      - "CONTAIN(title, 'backend') && CONTAIN(title, 'login')"
      - "CONTAIN(title, 'management') && CONTAIN(title, 'system')"

  - name: "sheet2"    
    filters:      
      - "CONTAIN(title, 'nginx') && CONTAIN(title, 'tomcat')"
      - "CONTAIN(title, 'IIS') && CONTAIN(title, 'Welcome to OpenResty')"

  - name: "sheet3"
    filters:
      - "protocol!='http' || protocol!='https'"

  - name: "sheet4"
    filters:
      - "CONTAIN(product_category, 'router') && CONTAIN(product_category, 'video surveillance')"
      - "CONTAIN(product_category, 'network storage') && CONTAIN(product_category, 'firewall')"
```


## GoFOFA All Parameters list

### `search`

| Parameter   | Abbreviation | Default Value | Description                                               |
|-------------|--------------|---------------|-----------------------------------------------------------|
| fields      | f            | ip,port       | Fields returned by FOFA. [Learn More](https://fofa.info/vip) |
| format      |              | csv           | Output format: csv/json/xml                               |
| outFile     | o            |               | Output file. If not set, prints to terminal               |
| size        | s            | 100           | Query size. Maximum is 10,000, subject to `deductMode`    |
| deductMode  |              |               | Consumption of f-points. If not set, uses max free query limit |
| fixUrl      |              | false         | Combines URLs (e.g., 1.1.1.1,80 becomes http://1.1.1.1)   |
| urlPrefix   |              | http://       | URL prefix                                                |
| full        |              | false         | Retrieves full data                                       |
| uniqByIP    |              | false         | Removes duplicates based on IP                           |
| workers     |              | 10            | Number of threads                                         |
| rate        |              | 2             | Query rate per second                                     |
| template    |              | ip={}         | Replaces `{}` with content from pipeline input           |
| inFile      | i            |               | Input file. If not set, reads from pipeline input         |
| checkActive |              | -1            | Number of retries for liveness checks. `-1` disables it   |
| deWildcard  |              | -1            | Removes wildcard domains. `-1` disables this feature      |
| filter      |              |               | Data filtering rules (e.g., `port<100 || host=="baidu.com"`) |
| dedupHost   |              | false         | Removes duplicates for subdomains                        |
| headline    |              | false         | Outputs CSV headers. Available only when format is CSV    |
| help        | h            | false         | Displays usage information                                |

### `dump`

| Parameter   | Abbreviation | Default Value | Description                                               |
|-------------|--------------|---------------|-----------------------------------------------------------|
| fields      | f            | ip,port       | Fields returned by FOFA. [Learn More](https://fofa.info/vip) |
| format      |              | csv           | Output format: csv/json/xml                               |
| outFile     | o            |               | Output file. If not set, prints to terminal               |
| inFile      | i            |               | Input file. If not set, reads from pipeline input         |
| size        | s            | 100           | Query size. No upper limit but consumes f-points or free query quota |
| fixUrl      |              | false         | Combines URLs (e.g., 1.1.1.1,80 becomes http://1.1.1.1)   |
| urlPrefix   |              | http://       | URL prefix                                                |
| full        |              | false         | Retrieves full data                                       |
| batchSize   | bs           | 1000          | Number of records to fetch per batch                     |
| batchType   | bt           |               | Batch query type: ip/domain                              |
| help        | h            | false         | Displays usage information                                |

### `jsRender`

| Parameter   | Abbreviation | Default Value | Description                                               |
|-------------|--------------|---------------|-----------------------------------------------------------|
| url         | u            |               | Single URL rendering                                      |
| tags        | t            |               | Tags to extract. Options: title/body                     |
| format      |              | csv           | Output format: csv/json/xml                               |
| outFile     | o            |               | Output file. If not set, prints to terminal               |
| inFile      | i            |               | Input file. If not set, reads from pipeline input         |
| workers     |              | 2             | Number of threads                                         |
| retry       |              | 3             | Number of timeout retries                                |
| help        | h            | false         | Displays usage information                                |

### `domains`

| Parameter   | Abbreviation | Default Value | Description                                               |
|-------------|--------------|---------------|-----------------------------------------------------------|
| outFile     | o            |               | Output file. If not set, prints to terminal               |
| size        | s            | 100           | Query size. Maximum is 10,000, subject to `deductMode`    |
| deductMode  |              |               | Consumption of f-points. If not set, uses max free query limit |
| full        |              | false         | Retrieves full data                                       |
| withCount   |              | false         | Outputs domain count                                      |
| clue        |              | false         | Outputs clue statements                                   |
| help        | h            | false         | Displays usage information                                |

### `active`

| Parameter   | Abbreviation | Default Value | Description                                               |
|-------------|--------------|---------------|-----------------------------------------------------------|
| url         | u            |               | Single URL liveness check                                 |
| format      |              | csv           | Output format: csv/json/xml                               |
| outFile     | o            |               | Output file. If not set, prints to terminal               |
| inFile      | i            |               | Input file. If not set, reads from pipeline input         |
| workers     |              | 2             | Number of threads                                         |
| retry       |              | 3             | Number of timeout retries                                |
| help        | h            | false         | Displays usage information                                |


### `category`

| Parameter   | Abbreviation | Default Value | Description                        |
|-------------|--------------|---------------|------------------------------------|
| inFile      | i            |               | Input classification file (CSV)   |
| unique      |              |               | Ensures unique classification data |
| help        | h            | false         | Displays usage information         |

### `dedup`

| Parameter   | Abbreviation | Default Value     | Description                              |
|-------------|--------------|-------------------|------------------------------------------|
| dedup       | d            |                   | Field to deduplicate                     |
| inFile      | i            |                   | Input file for deduplication (CSV)       |
| outFile     | o            | duplicate.csv     | Output file                              |
| help        | h            | false             | Displays usage information               |

### `host`

| Parameter   | Abbreviation | Default Value | Description                        |
|-------------|--------------|---------------|------------------------------------|
| help        | h            | false         | Displays usage information         |

### `icon`

| Parameter   | Abbreviation | Default Value | Description                                   |
|-------------|--------------|---------------|-----------------------------------------------|
| open        |              | false         | Opens FOFA search page based on icon results |
| help        | h            | false         | Displays usage information                   |

### `stats`

| Parameter   | Abbreviation | Default Value   | Description                              |
|-------------|--------------|-----------------|------------------------------------------|
| fields      | f            | title,country   | Fields returned by FOFA. [Learn More](https://en.fofa.info/vip) |
| size        | s            | 5               | Query count. `-1` for infinite queries   |
| help        | h            | false           | Displays usage information               |

### `random`

| Parameter   | Abbreviation | Default Value                        | Description                              |
|-------------|--------------|--------------------------------------|------------------------------------------|
| fields      | f            | ip,port,host,header,title,server,lastupdatetime | Fields returned by FOFA. [Learn More](https://en.fofa.info/vip) |
| format      |              | json                                | Output format: csv/json/xml              |
| size        | s            | 1                                   | Query count. `-1` for infinite queries   |
| sleep       |              | 1000                                | Interval between queries in milliseconds |
| fixUrl      |              | false                               | Combines URLs (e.g., 1.1.1.1,80 becomes http://1.1.1.1) |
| urlPrefix   |              | http://                             | URL prefix                              |
| full        |              | false                               | Retrieves full data                     |
| help        | h            | false                               | Displays usage information               |

### `count`

| Parameter   | Abbreviation | Default Value | Description                        |
|-------------|--------------|---------------|------------------------------------|
| help        | h            | false         | Displays usage information         |

### `account`

| Parameter   | Abbreviation | Default Value | Description                        |
|-------------|--------------|---------------|------------------------------------|
| help        | h            | false         | Displays usage information         |

---

## Final Thoughts

Before deciding to open-source GoFOFA, we were well aware that the community already had several excellent FOFA API tools. Examples include [fofa_viewer](https://github.com/wgpsec/fofa_viewer) by WgpSec, [fofax](https://github.com/xiecat/fofax) by Xiecat, and [FofaMap](https://github.com/asaotomo/FofaMap) by the HxO team. These tools are powerful and beloved by white-hat researchers. Initially, we believed that with so many outstanding FOFA API tools already available, creating another similar tool as an official project might seem unnecessary.

However, through a deeper understanding of user needs, we discovered that merely retrieving API data is far from sufficient. The real challenge lies in efficient post-processing of the data to meet various practical scenarios. This pain point resonated with us deeply. “There’s nothing beyond the capabilities of our users,” we thought, as the diverse requirements for data processing continued to emerge. Even within our team, while handling data, we often found ourselves developing various solutions for secondary processing.

This led us to a thought: why not consolidate common needs into an open-source tool? A tool that not only retrieves data efficiently but also provides powerful secondary processing capabilities, while serving as a platform for community co-creation. Users can experience FOFA's features, propose additional requirements, and even co-create new functionalities with us.

We hope that by open-sourcing GoFOFA, more users can enjoy this powerful and flexible tool. We also look forward to collaborating with everyone to foster innovation and co-creation within the FOFA community.

If you think this tool is useful, give us a ⭐ on GitHub!


