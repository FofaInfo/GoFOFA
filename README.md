# gofofa用户指南

fofa client in Go

[![Test status](https://github.com/lubyruffy/gofofa/workflows/Go/badge.svg)](https://github.com/lubyruffy/gofofa/actions?query=workflow%3A%22Go%22)
[![codecov](https://codecov.io/gh/lubyruffy/gofofa/branch/main/graph/badge.svg)](https://codecov.io/gh/lubyruffy/gofofa)
[![License: MIT](https://img.shields.io/github/license/lubyruffy/gofofa)](https://github.com/LubyRuffy/gofofa/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/lubyruffy/gofofa)](https://goreportcard.com/report/github.com/lubyruffy/gofofa)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/3eadab4e412e4c3494bbc5f188d441e8)](https://www.codacy.com/gh/LubyRuffy/gofofa/dashboard?utm_source=github.com&utm_medium=referral&utm_content=LubyRuffy/gofofa&utm_campaign=Badge_Grade)
[![Github Release](https://img.shields.io/github/release/lubyruffy/gofofa/all.svg)](https://github.com/lubyruffy/gofofa/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/LubyRuffy/gofofa.svg)](https://pkg.go.dev/github.com/LubyRuffy/gofofa)

## Background

The official library doesn't has unittests,  之前官方的库功能不全，代码质量差，完全没有社区活跃度，不符合开源项目的基本要求。因此，想就fofa的客户端作为练手，解决上述问题。

## Usage

### Build and run

> 安装配置

- 下载gofofa:

```
$ go install github.com/LubyRuffy/gofofa/cmd/fofa@latest
```

- 显示如下表示安装成功:

```
$ fofa
NAME:
   fofa - fofa client on Go v0.2.23, commit none, built at unknown

USAGE:
   fofa [global options] command [command options] [arguments...]

VERSION:
   v0.2.23

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
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --fofaURL value, -u value  format: <url>/?email=<email>&key=<key>&version=<v2> (default: "https://fofa.info/?email=your_email&key=your_key&version=v1")
   --verbose                  print more information (default: false)
   --accountDebug             print account in error log (default: false)
   --help, -h                 show help (default: false)
   --version, -v              print the version (default: false)
```

- 配置环境变量:

```
$ FOFA_CLIENT_URL='https://fofa.info/?email=your_email&key=your_key'
```

### Search

> 搜索

-   fofa语法查询，可以输入单个查询语句，默认会输出ip,端口:

```shell
$ fofa search port=80
2024/08/23 11:51:19 query fofa of: port=80
69.10.146.92,80
20.193.138.22,80
194.182.72.64,80
......
......
```

或输入多个查询语句:

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

-   不选择子模块查询的话，会默认使用search模块进行查询:

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

-   使用fields来选择输出的字段，默认会输出ip，端口:

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

或者更简洁一些:

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



-   使用size来选择输出数量， 默认大小100:

```shell
$ fofa search --size 5 'port=6379'
2024/08/23 14:07:18 query fofa of: port=6379
47.99.89.216,6379
112.124.14.11,6379
107.154.224.11,6379
39.101.36.243,6379
139.196.136.107,6379
```

或者更简洁一些:

```shell
$ fofa search -s 5 'port=6379'
2024/08/23 14:07:18 query fofa of: port=6379
47.99.89.216,6379
112.124.14.11,6379
107.154.224.11,6379
39.101.36.243,6379
139.196.136.107,6379
```

如果size大于您的帐户免费限制，您可以设置 `-deductMode` 来决定是否自动扣除f点

-   如果需要输出不同的数据格式，可以通过format来设置，默认是csv格式，还支持json和xml格式:

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

或者:

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

-   使用outFile可以将结果输出到指定文件中，若不设置次参数则默认输出在命令行中:

```shell
$ fofa search --outFile a.txt 'port=6379'
```

或者更简洁一些:

```shell
$ fofa search -o a.txt 'port=6379'
```

-   如果你想获取完整的url，可以使用fixUrl参数:

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
- 如果你想要使用其他的前缀，可以使用`urlPrefix`:

```shell
$ fofa --size 1 --fields "host" --fixUrl --urlPrefix "redis://" protocol=redis
2024/08/23 14:29:26 query fofa of: protocol=redis
redis://139.9.222.14:7000
```

- 如果你想要进行web存活探测，可以使用```--checkActive 3```，`3`是超时重复次数（使用这个参数之后也会重新获取status_code数据）:

```shell
$ fofa -s 3 --checkActive port=80 
2024/08/26 18:52:00 query fofa of: port=80
216.92.244.44,80,true
104.21.31.50,80,true
182.247.239.68,80,true
$ fofa -s 3 --checkActive 3 --format=json port=80
2024/08/26 18:53:33 query fofa of: port=80
{"ip":"54.78.179.223","isActive":"false","port":"80"}
{"ip":"18.155.202.65","isActive":"true","port":"80"}
{"ip":"198.144.179.122","isActive":"true","port":"80"}
$ fofa -s 3 --checkActive 3 --format=xml port=80
2024/08/26 18:54:38 query fofa of: port=80
<result><ip>50.16.35.210</ip><port>80</port><isActive>true</isActive></result>
<result><ip>104.21.77.62</ip><port>80</port><isActive>true</isActive></result>
<result><ip>189.193.236.170</ip><port>80</port><isActive>false</isActive></result>
```

- 如果你想要减少泛域名数量，可以使用```--deWildcard```设置保留泛域名数量，```-f```可以支持其他字段选用link做为演示（该参数只有企业账号以上可用）:

```shell
$ fofa -s 3 -f link domain=huashunxinan.net
2024/08/27 17:19:04 query fofa of: domain=huashunxinan.net
http://h8huumr2zdmwgy5.huashunxinan.net
http://keygatjexlvsznh.huashunxinan.net
http://jobs.huashunxinan.net
$ fofa -s 3 -f link --deWildcard 1 domain=huashunxinan.net
2024/08/27 17:26:42 query fofa of: domain=huashunxinan.net
http://h8huumr2zdmwgy5.huashunxinan.net
https://fwtn2k7oigaiyla.huashunxinan.net
http://huashunxinan.net
```

- ```--dedupHost```，在fofa中subdomain代表网页数据，service代表协议数据，如果host相同，优先保留subdomain数据:

```shell
$ fofa -s 3 -f host,type "ip=106.75.95.206"
2024/08/28 19:49:23 query fofa of: ip=106.75.95.206
106.75.95.206,subdomain
106.75.95.206:443,service
106.75.95.206,service
$ fofa -s 3 -f host,type --dedupHost "ip=106.75.95.206"
2024/08/28 19:52:30 query fofa of: ip=106.75.95.206
https://106.75.95.206,subdomain
106.75.95.206:443,service
106.75.95.206,subdomain
```

- 如果你想要针对结果进行过滤，你可以使用```-filter```过滤器，它的值是一个布尔表达式，保留符合filter表达式结果的数据:

```shell
$ fofa -s 3 -f host,title,status_code domain=huashunxinan.net
2024/08/28 19:56:47 query fofa of: domain=huashunxinan.net
eedwwsqpoq1yjrf.huashunxinan.net,301 Moved Permanently,301
https://keygatjexlvsznh.huashunxinan.net,华顺信安-网络空间测绘的先行者,200
https://i.huashunxinan.net,,301
$ fofa -s 3 -f host,title,status_code -filter "status_code=='200'&&title!=''" domain=huashunxinan.net
2024/08/27 17:26:42 query fofa of: domain=huashunxinan.net
https://www.huashunxinan.net,华顺信安-网络空间测绘的先行者,200
huashunxinan.net,华顺信安-网络空间测绘的先行者,200
https://huashunxinan.net,华顺信安-网络空间测绘的先行者,200
```

-   如果你想在输出csv文件的时候添加表头，可以使用参数```--headline```（只有在format为csv的情况下才可以使用）:

```shell
$ fofa search -f host,port --headline -o output.csv port=80
```

-   如果你想查看更多的debug信息，可以使用全局参数```--verbose```:

```shell
$ fofa --verbose search port=80
```

-   支持管道:

```shell
$ fofa -fields "host" -fixUrl 'app="Aspera-Faspex"' | nuclei -t http/cves/2022/CVE-2022-47986.yaml
```

-   如果你想要根据ip进行去重，可以使用```--uniqByIP```:

```shell
$ fofa --fixUrl --size 5 --fields host ip=123.58.224.8
2024/08/23 17:58:24 query fofa of: ip=123.58.224.8
http://123.58.224.8:8008
http://123.58.224.8
https://123.58.224.8:63739
http://123.58.224.8:22937
https://123.58.224.8:14272
$ fofa --fixUrl --size 5 --fields host --uniqByIP ip=123.58.224.8
2024/08/23 17:58:49 query fofa of: ip=123.58.224.8
http://123.58.224.8:8008
```

-   如果你想要更高级的使用方法，可以使用`{}`做为占位符来达到批量获取数据的效果:

```shell
$ fofa -f ip "is_ipv6=false && port=22" | fofa -f ip -uniqByIP -template "port=8443 && ip={}" 
```
你也可以通过 `-rate 3` 来设置速率, 默认是 2

### Stats

> 数据统计

-   stats模块可以做数据统计等操作

```shell
$ fofa stats --fields title,country title="hacked by"
```
![fofa stats](./data/fofa_stats.png)

### Icon

>Icon查询（商业版及以上）

- 你可以通过读取本地的ico文件来查询数据，open参数会自动帮你跳转到fofa:

```shell
$ fofa icon --open ./data/favicon.ico
```

也可以通过网页的ico文件来查询:

```shell
$ fofa icon --open https://fofa.info/favicon.ico
```

还可以直接通过url来查询:

```shell
$ fofa icon --open http://www.baidu.com
```

- 获取本地ico文件的hash值:

```shell
$ fofa icon ./data/favicon.ico
-247388890
```

也可以获取网页ico文件的hash值:

```shell
$ fofa icon https://fofa.info/favicon.ico
-247388890
```

还可以直接获取url的ico_hash值:

```shell
$ fofa icon http://www.baidu.com
-1588080585
```

### Host

> 获取Host信息

-   Host模块，输入域名即可获取host信息:

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

### Dump

> 数据存储

-   存储超大数据，使用`-batchSize`设置数量:

```shell
$ fofa dump --format json -fixUrl -outFile a.json -batchSize 10000 'title=phpinfo'
```

-   通过fofa语句文件，来存储超大数据（每条数据一行）:

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

### Domains

> 简单域名拓线

-   domains子模块主要用于最简单的拓线，通过证书进行拓线，可以使用withCount来统计数量:

```shell
$ fofa domains -s 1000 -withCount baidu.com
baidu.com       660
dwz.cn  620
dlnel.com       614
bcehost.com     614
bdstatic.com    614
......
......
```

你还可以使用 `-uniqByIP` 来去除相同的ip:
```shell
$ fofa domains -s 1000 -withCount -uniqByIP baidu.com 
baidu.com       448
dwz.cn  410
aipage.cn       406

```

### Active

> 存活探测

- active模块用来对url进行web存活探测，可以使用target来获取存活信息，true为存活，false为不存活:


```shell
$ fofa active -target baidu.com,fofa.info,asdsadsasdas.com
baidu.com,true
fofa.info,true
asdsadsasdas.com,false
```

或者可以更简洁一些:

```shell
$ fofa active -t baidu.com,fofa.info,asdsadsasdas.com
baidu.com,true
fofa.info,true
asdsadsasdas.com,false
```

- 还可以通过对一个每行为一个url的文件进行探测:


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

- 还支持对管道中的url进行探测（管道中的数据需为每行一条url）:


```shell
$ fofa search -f link -s 3 port=80 | fofa active
2024/08/23 15:50:11 query fofa of: port=80
http://og823.hb-yj.com,true
http://rw823.tcxzgh.org,true
http://sb823.tcxzgh.org,true
```

### Dedup

> 去重

- dedup支持对一个csv文件中的某一个字段进行去重，通过input参数上传文件，通过dedup参数选择去重字段（会根据字段顺序进行去重），通过output设置输出文件名（默认duplicate.csv）:

```shell
$ fofa dedup -output data.csv -dedup ip -output dedup.csv
$ fofa dedup -output data.csv -dedup ip,host,domain -output dedup.csv
```
或者可以更简洁一些:

```shell
$ fofa dedup -o data.csv -d ip -o dedup.csv
$ fofa dedup -o data.csv -d ip,host,domain -o dedup.csv
```

### Category

> 分类

- Category支持对一个csv文件进行分类，通过config.yaml配置文件来进行分类（配置文件必须在当前目录下），配置文件如下格式:

```shell
categories:
  - name: "hard"
    filters:
      - "CONTAIN(category, '数据证书')"

  - name: "soft"
    filters:
      - "category == '其他支撑系统'"

  - name: "buss"
    filters:
      - "category == '电子邮件系统' || CONTAIN(category, '其他企业应用')"

```
- 可以在config.yaml文件中设置好过滤规则`filter`，内置了一个`CONTAIN`方法，意思是某一个字段是否含有什么值`-output`不设置会默认生成`category.csv`文件:

```shell
$ fofa category -input input.csv [-output category.csv]
```

或者更简洁一些:

```shell
$ fofa category -i input.csv [-o category.csv]
```


### JsRender

> 存活探测

- jsRender模块用来对url进行js渲染，可以使用`-url`来选择单个目标，`-tag`选择获取渲染后的标签:


```shell
$ fofa jsRender -url http://baidu.com -tag title
http://baidu.com,百度一下，你就知道
```

或者可以更简洁一些:

```shell
$ fofa jsRender -u http://baidu.com -t title
http://baidu.com,百度一下，你就知道
```

- 还可以通过对一个每行为一个url的文件进行探测:

```shell
$ cat url.txt
http://baidu.com
https://fofa.info
$ fofa jsRender -i url.txt -t title 
http://baidu.com,百度一下，你就知道
https://fofa.info,网络空间测绘，网络空间安全搜索引擎，网络空间搜索引擎，安全态势感知 - FOFA网络空间测绘系统
```

- 还支持对管道中的url进行探测（管道中的数据需为每行一条url）:


```shell
$ fofa search -f link -s 3 port=80 | fofa jsRender -t title
2024/08/23 15:50:11 query fofa of: port=80
http://project5.abioyibo.com,Just another WordPress site
http://www.valuegoodsbazaar.shop,srv258.sellvir.com — Coming Soon
http://forecasting-preprod.pcasys.co.uk,- Sales Forecasting Tool (preprod)


### Utils

> 其他

-   random 模块

随机从fofa生成数据:
```shell
$ fofa random -f host,ip,port,lastupdatetime,title,header,body --format json
{"body":"","header":"HTTP/1.1 401 Unauthorized\r\nWww-Authenticate: Digest realm=\"IgdAuthentication\", domain=\"/\", nonce=\"ZjVhNGY2YzI6MTUyNDM2N2Y6MzRiMGZjZjQ=\", qop=\"auth\", algorithm=MD5\r\nContent-Length: 0\r\n","host":"95.22.200.127:7547","ip":"95.22.200.127","lastupdatetime":"2024-08-14 13:00:00","port":"7547","title":""}
```

可以通过sleep参数设置时间500ms，按照时间每500ms生成一次数据:

```shell
$ fofa random -s -1 -sleep 500
```

-   count 模块

可以通过count模块统计数据数量:

```shell
$ fofa count port=80
```

-   account 模块

可以获取账户信息:

```shell
$ fofa account
```

-   version

获取gofofa版本号

```shell
$ fofa --version
```

## Features

-   ☑ Cross-platform
    -   ☑ Windows
    -   ☑ Linux
    -   ☑ Mac
-   ☑ Code coverage > 90%
-   ☑ As SDK
    -   ☑ Client: NewClient
        -   ☑ HostSearch
        -   ☑ HostSize
        -   ☑ AccountInfo
        -   ☑ IconHash
        -   ☑ support cancel through SetContext
-   ☑ As Client
    -   ☑ Sub Commands
        -   ☑ account
        -   ☑ search
            -   ☑ query
            -   ☑ fields/f
            -   ☑ size/s
            - group/g 根据字段聚合：group by ip 根据ip合并，比如查询一个app会有很多域名，其中多个域名对应一个ip，这时只测试其中一个就好了
            -   ☑ fixUrl build valid url，默认的字段如果是http的话前面没有http://前缀，导致命令行一些工具不能使用，通过这个参数进行修复
              -   ☑ can use with urlPrefix, such as use `app://` instead of `http://`
              -   ☑ support socks5
              -   ☑ support redis
            -   ☑ full 匹配所有，而不只是一年内的
            -   ☑ format
                -   ☑ csv
                -   ☑ json
                -   ☑ xml
                -   ☐ table
                -   ☐ excel
            -   ☑ outFile/o
        -   ☑ stats
        -   ☑ icon
        -   ☐ web
        -   ☑ dump https://en.fofa.info/api/batches_pages large-scale data retrieval
        -   ☑ domains
        -   ☑ active
        -   ☑ duplicate
    -   ☑ Terminal color 
    -   ☑ Global Config
        -   ☑ fofaURL
        -   ☑ deductMode
    -   ☑ Envirement
        -   ☑ FOFA_CLIENT_URL format: <url>/?email=\<email\>&key=\<key\>&version=\<v1\>
        -   ☑ FOFA_SERVER
        -   ☑ FOFA_EMAIL
        -   ☑ FOFA_KEY
-   ☐ Publish
    -   ☑ github
    -   ☐ brew
    -   ☐ apt
    -   ☐ yum


## Scenes

### How to dump all domains that cert is valid and contains google?

```shell
./fofa stats -f domain -s 100 'cert.is_valid=true && (cert="google")'
```
